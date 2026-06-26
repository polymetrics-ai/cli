// Package flexmail implements the native pm Flexmail connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package composing the connsdk toolkit (Requester + Basic
// auth + RecordsAt extraction + offset pagination) with Flexmail-specific stream
// definitions and endpoints.
//
// Flexmail is an email-marketing platform. Its public API
// (https://api.flexmail.eu) authenticates with HTTP Basic auth where the
// username is the account id and the password is a personal access token, and it
// returns HAL-style collections under {"_embedded":{"item":[...]}}. The
// connector is read-only (full refresh); Flexmail exposes no reverse-ETL writes
// that make sense here, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package flexmail

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
	flexmailDefaultBaseURL  = "https://api.flexmail.eu"
	flexmailDefaultPageSize = 500
	flexmailMaxPageSize     = 500
	flexmailUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("flexmail", New)
}

// New returns the Flexmail connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Flexmail connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "flexmail" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "flexmail",
		DisplayName:     "Flexmail",
		IntegrationType: "api",
		Description:     "Reads Flexmail contacts, custom fields, interests, segments, and sources through the Flexmail REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Flexmail. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := flexmailBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(flexmailSecret(cfg)) == "" {
		return errors.New("flexmail connector requires secret personal_access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contacts list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"limit": []string{"1"}, "offset": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", query, nil, nil); err != nil {
		return fmt.Errorf("check flexmail: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: flexmailStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := flexmailStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("flexmail stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := flexmailPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := flexmailMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write is unsupported: Flexmail is a read-only source in pm.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Flexmail's offset pagination. Collections come back as
// {"_embedded":{"item":[...]}}. Paginated endpoints (contacts, sources) advance
// offset by page size until a short page is returned; non-paginated endpoints
// fetch a single page. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt, mirroring stripe.harvest.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("limit", strconv.Itoa(pageSize))
			query.Set("offset", strconv.Itoa(offset))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read flexmail %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode flexmail %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Non-paginated endpoints return everything in one response.
		if !endpoint.paginated {
			return nil
		}
		// A short page (fewer than pageSize records) signals the end.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise flexmail credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := flexmailStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fixtureID(stream, i),
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"name":               fmt.Sprintf("Fixture %d", i),
			"first_name":         fmt.Sprintf("Fixture%d", i),
			"language":           "en",
			"custom_fields":      map[string]any{"tier": "gold"},
			"type":               "text",
			"placeholder":        "value",
			"label":              fmt.Sprintf("Label %d", i),
			"description":        "fixture interest",
			"visibility":         "public",
			"number_of_contacts": int64(10 * i),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// fixtureID returns a deterministic primary key whose type matches each stream's
// real id type: contacts and sources use numeric ids; the rest use string ids.
func fixtureID(stream string, i int) any {
	switch stream {
	case "contacts", "sources":
		return int64(i)
	default:
		return fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(stream, "s"), i)
	}
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (account_id as
// username, personal access token as password) and the resolved base URL. The
// secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := flexmailBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := flexmailSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("flexmail connector requires secret personal_access_token")
	}
	accountID := strings.TrimSpace(cfg.Config["account_id"])
	if accountID == "" {
		return nil, errors.New("flexmail connector requires config account_id")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(accountID, secret),
		UserAgent: flexmailUserAgent,
	}, nil
}

func flexmailSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["personal_access_token"]
}

// flexmailBaseURL resolves and validates the base URL. The default is
// api.flexmail.eu; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func flexmailBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return flexmailDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("flexmail config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("flexmail config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("flexmail config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func flexmailPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return flexmailDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("flexmail config page_size must be an integer: %w", err)
	}
	if value < 1 || value > flexmailMaxPageSize {
		return 0, fmt.Errorf("flexmail config page_size must be between 1 and %d", flexmailMaxPageSize)
	}
	return value, nil
}

func flexmailMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("flexmail config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("flexmail config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
