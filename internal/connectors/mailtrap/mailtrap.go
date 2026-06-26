// Package mailtrap implements the native pm Mailtrap source connector. It follows
// the declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Mailtrap-specific stream definitions and endpoints.
//
// Mailtrap exposes a simple account-management REST API at https://mailtrap.io/api.
// Most list endpoints return a bare JSON array (sending_domains is wrapped in a
// {"data":[...]} envelope) and are not formally paginated, so the read loop
// advances a `page` query param and stops on a short or fully-duplicate page,
// which is safe whether or not the upstream honours pagination.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. Mailtrap is read-only (full-refresh source),
// so it does not implement the write capability.
package mailtrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	mailtrapDefaultBaseURL  = "https://mailtrap.io/api"
	mailtrapDefaultPageSize = 100
	mailtrapMaxPageSize     = 1000
	mailtrapMaxPagesGuard   = 1000
	mailtrapUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("mailtrap", New)
}

// New returns the Mailtrap connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mailtrap source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailtrap" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailtrap",
		DisplayName:     "Mailtrap",
		IntegrationType: "api",
		Description:     "Reads Mailtrap accounts, inboxes, projects, and sending domains through the Mailtrap account-management REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mailtrap. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mailtrapBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mailtrapSecret(cfg)) == "" {
		return errors.New("mailtrap connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing accounts confirms auth and connectivity without mutating anything.
	if _, err := r.Do(ctx, http.MethodGet, "accounts", nil, nil); err != nil {
		return fmt.Errorf("check mailtrap: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mailtrapStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	endpoint, ok := mailtrapStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailtrap stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	path, accountID, err := resolvePath(endpoint, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mailtrapPageSize(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, path, accountID, pageSize, emit)
}

// harvest reads each page of a Mailtrap list endpoint. Mailtrap list endpoints
// return a bare array (or a {"data":[...]} envelope) and do not document a
// pagination token, so the loop advances a `page` query param and stops when a
// page is short or yields no records the caller has not already seen (keyed by
// id). The id-dedupe guard means that even when the upstream ignores `page` and
// returns the same array every time, the loop terminates after one effective
// page instead of looping forever.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, path, accountID string, pageSize int, emit func(connectors.Record) error) error {
	seen := map[string]struct{}{}
	for page := 1; page <= mailtrapMaxPagesGuard; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read mailtrap %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode mailtrap %s page: %w", endpoint.resource, err)
		}

		emitted := 0
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			key := stringField(item, "id")
			if key != "" {
				if _, dup := seen[key]; dup {
					continue
				}
				seen[key] = struct{}{}
			}
			if err := emit(stampAccount(endpoint, accountID, item)); err != nil {
				return err
			}
			emitted++
		}

		// Stop on a short page (fewer than the requested size) or when the page
		// contributed no new records (upstream ignores pagination / repeats).
		if len(records) < pageSize || emitted == 0 {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailtrap credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             int64(i),
			"name":           fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"access_levels":  []any{int64(1000)},
			"domain":         "example.mailtrap.io",
			"email_username": fmt.Sprintf("fixture+%d", i),
			"emails_count":   int64(i),
			"status":         "active",
			"max_size":       int64(1000),
			"used_size":      int64(i),
			"domain_name":    "example.com",
			"demo":           false,
		}
		record := stampAccount(endpoint, "fixture_account", item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// Write is unsupported: Mailtrap is a full-refresh source with no safe
// reverse-ETL surface, so the connector is read-only. The method exists only to
// satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mailtrapBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mailtrapSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mailtrap connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: mailtrapUserAgent,
	}, nil
}

// resolvePath builds the request path for a stream, inserting the configured
// account_id for account-scoped streams. It returns the path and the resolved
// account_id (empty for root-scoped streams).
func resolvePath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, string, error) {
	if endpoint.scope == scopeRoot {
		return endpoint.resource, "", nil
	}
	accountID := strings.TrimSpace(cfg.Config["account_id"])
	if accountID == "" {
		return "", "", fmt.Errorf("mailtrap stream requires config account_id for %q", endpoint.resource)
	}
	return "accounts/" + url.PathEscape(accountID) + "/" + endpoint.resource, accountID, nil
}

// stampAccount maps an item to a record and, for account-scoped streams, stamps
// the parent account_id so the record carries its parent context.
func stampAccount(endpoint streamEndpoint, accountID string, item map[string]any) connectors.Record {
	record := endpoint.mapRecord(item)
	if endpoint.scope == scopeAccount {
		record["account_id"] = accountID
	}
	return record
}

func mailtrapSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// mailtrapBaseURL resolves and validates the base URL. The default is
// mailtrap.io/api; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mailtrapBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mailtrapDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mailtrap config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mailtrap config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mailtrap config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mailtrapPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mailtrapDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailtrap config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mailtrapMaxPageSize {
		return 0, fmt.Errorf("mailtrap config page_size must be between 1 and %d", mailtrapMaxPageSize)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
