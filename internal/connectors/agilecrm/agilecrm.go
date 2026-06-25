// Package agilecrm implements the native pm AgileCRM connector. It follows the
// declarative-HTTP per-system connector template (see internal/connectors/stripe):
// a thin package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction) with AgileCRM-specific stream definitions, endpoints, and
// its last-record cursor pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// AgileCRM auth is HTTP Basic with the account email as username and the API key
// as password, against a per-account subdomain base URL
// (https://{domain}.agilecrm.com/dev/api). The connector is read-only: the
// AgileCRM source exposes full-refresh reads only, and there are no obviously
// safe reverse-ETL write actions to allow-list here.
package agilecrm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	agilecrmDefaultPageSize = 50
	agilecrmMaxPageSize     = 100
	agilecrmUserAgent       = "polymetrics-go-cli"
	// agilecrmFixtureCreated is the deterministic created_time used by fixture
	// records (2026-01-01T00:00:00Z in unix milliseconds, AgileCRM's unit).
	agilecrmFixtureCreated int64 = 1767225600000
)

func init() {
	connectors.RegisterFactory("agilecrm", New)
}

// New returns the AgileCRM connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm AgileCRM connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "agilecrm" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "agilecrm",
		DisplayName:     "AgileCRM",
		IntegrationType: "api",
		Description:     "Reads AgileCRM contacts, deals, tasks, and milestone pipelines through the AgileCRM REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to AgileCRM. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := agilecrmBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(agilecrmEmail(cfg)) == "" {
		return errors.New("agilecrm connector requires config email")
	}
	if strings.TrimSpace(agilecrmSecret(cfg)) == "" {
		return errors.New("agilecrm connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of contacts confirms auth and connectivity without mutating.
	query := url.Values{"page_size": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", query, nil, nil); err != nil {
		return fmt.Errorf("check agilecrm: %w", err)
	}
	return nil
}

// Write satisfies the Connector interface. AgileCRM is read-only here (the
// source exposes full-refresh reads only and there is no allow-listed reverse
// ETL action set), so writes are explicitly unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: agilecrmStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := agilecrmStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("agilecrm stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := agilecrmPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := agilecrmMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives AgileCRM's last-record cursor pagination. List endpoints return
// a top-level JSON array; when a page is full the final element carries a
// "cursor" string used to request the next page (?cursor=...). The absence of a
// cursor on the last element signals the end of the list. Non-paginated streams
// return a single bounded array and stop after one request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("page_size", strconv.Itoa(pageSize))
			if cursor != "" {
				query.Set("cursor", cursor)
			}
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read agilecrm %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode agilecrm %s page: %w", endpoint.resource, err)
		}
		nextCursor := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			// The cursor lives on the last record; capture it then drop it from
			// the emitted record so it does not leak into downstream schemas.
			if cv := stringField(item, "cursor"); cv != "" {
				nextCursor = cv
			}
			delete(item, "cursor")
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// AgileCRM's documented end-of-list signal is the absence of a cursor on
		// the last record; a missing cursor (or a non-paginated stream, or an
		// empty page) terminates the harvest.
		if !endpoint.paginated || nextCursor == "" || len(records) == 0 {
			return nil
		}
		cursor = nextCursor
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise agilecrm credential-free (mirrors the stripe
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               int64(i),
			"type":             "PERSON",
			"name":             fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"subject":          fmt.Sprintf("Fixture task %d", i),
			"created_time":     agilecrmFixtureCreated + int64(i),
			"updated_time":     agilecrmFixtureCreated + int64(i),
			"star_value":       int64(0),
			"lead_score":       int64(10 * i),
			"tags":             []any{"fixture"},
			"properties":       []any{},
			"expected_value":   float64(1000 * i),
			"probability":      int64(50),
			"milestone":        "Prospect",
			"close_date":       agilecrmFixtureCreated,
			"pipeline_id":      int64(1),
			"priority_type":    "HIGH",
			"status":           "YET_TO_START",
			"due":              agilecrmFixtureCreated,
			"is_complete":      false,
			"milestones":       "Prospect,Won,Lost",
			"pipeline_default": i == 1,
			"owner_id":         "fixture-owner",
			"connector":        "agilecrm",
			"fixture":          true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (email +
// api_key) and the resolved base URL. The secret only ever flows into
// connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := agilecrmBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(agilecrmEmail(cfg))
	if email == "" {
		return nil, errors.New("agilecrm connector requires config email")
	}
	secret := agilecrmSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("agilecrm connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(email, secret),
		UserAgent: agilecrmUserAgent,
	}, nil
}

func agilecrmEmail(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["email"]
}

func agilecrmSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// agilecrmBaseURL resolves and validates the base URL. When base_url is set it is
// used directly (after scheme+host validation to bound SSRF risk); otherwise the
// per-account subdomain URL is derived from the required domain config.
func agilecrmBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		domain := strings.TrimSpace(cfg.Config["domain"])
		if domain == "" {
			return "", errors.New("agilecrm connector requires config domain or base_url")
		}
		if !validDomain(domain) {
			return "", fmt.Errorf("agilecrm config domain %q is invalid", domain)
		}
		return fmt.Sprintf("https://%s.agilecrm.com/dev/api", domain), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("agilecrm config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("agilecrm config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("agilecrm config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// validDomain restricts the subdomain to the safe label charset to keep the
// derived URL well-formed and prevent host injection.
func validDomain(domain string) bool {
	if domain == "" || len(domain) > 100 {
		return false
	}
	for _, r := range domain {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func agilecrmPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return agilecrmDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("agilecrm config page_size must be an integer: %w", err)
	}
	if value < 1 || value > agilecrmMaxPageSize {
		return 0, fmt.Errorf("agilecrm config page_size must be between 1 and %d", agilecrmMaxPageSize)
	}
	return value, nil
}

func agilecrmMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("agilecrm config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("agilecrm config max_pages must be 0 for unlimited or a positive integer")
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
