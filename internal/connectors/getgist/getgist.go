// Package getgist implements the native pm Gist (getgist.com) connector. It is
// a declarative-HTTP per-system connector following the stripe reference shape:
// a thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction) with Gist-specific stream definitions, endpoints, and
// page-number pagination.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// The Gist REST API is read-oriented for the resources exposed here, so the
// connector is read-only (Capabilities.Write = false).
package getgist

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
	getgistDefaultBaseURL  = "https://api.getgist.com"
	getgistDefaultPageSize = 50
	getgistMaxPageSize     = 100
	getgistUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("getgist", New)
}

// New returns the Gist connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Gist connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "getgist" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "getgist",
		DisplayName:     "GetGist",
		IntegrationType: "api",
		Description:     "Reads Gist contacts, tags, segments, campaigns, forms, and teammates through the Gist REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Gist. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := getgistBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(getgistSecret(cfg)) == "" {
		return errors.New("getgist connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contacts list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check getgist: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: getgistStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Gist resources exposed
// here are read-only, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := getgistStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("getgist stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := getgistPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := getgistMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Gist's page-number pagination. Gist list responses look like
// {"<resource>":[...],"pages":{"next":"...url..."}}. The loop advances the page
// query param and stops when either the response carries no pages.next link or a
// short page (fewer than pageSize records) is returned. The short-page guard
// protects against APIs that omit the link, preventing an infinite loop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read getgist %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode getgist %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop on a short page: fewer records than the requested size means the
		// last page was reached.
		if len(records) < pageSize {
			return nil
		}
		// Stop when the API reports no next page.
		next, err := connsdk.StringAt(resp.Body, "pages.next")
		if err != nil {
			return fmt.Errorf("decode getgist %s pages.next: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise getgist credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                       int64(i),
			"type":                     "user",
			"user_id":                  fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"email":                    fmt.Sprintf("fixture+%d@example.com", i),
			"name":                     fmt.Sprintf("Fixture %d", i),
			"phone":                    "",
			"created_at":               int64(1767225600 + i),
			"signed_up_at":             int64(1767225600 + i),
			"updated_at":               int64(1767225600 + i),
			"last_seen_at":             int64(1767225600 + i),
			"last_contacted_at":        int64(1767225600 + i),
			"session_count":            int64(i),
			"unsubscribed_from_emails": false,
			"subject":                  "Fixture subject",
			"status":                   "active",
			"person_type":              "user",
			"count":                    int64(i),
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := getgistBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := getgistSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("getgist connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: getgistUserAgent,
	}, nil
}

func getgistSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// getgistBaseURL resolves and validates the base URL. The default is
// api.getgist.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func getgistBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return getgistDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("getgist config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("getgist config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("getgist config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func getgistPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return getgistDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("getgist config page_size must be an integer: %w", err)
	}
	if value < 1 || value > getgistMaxPageSize {
		return 0, fmt.Errorf("getgist config page_size must be between 1 and %d", getgistMaxPageSize)
	}
	return value, nil
}

func getgistMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("getgist config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("getgist config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
