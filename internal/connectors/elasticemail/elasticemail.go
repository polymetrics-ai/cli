// Package elasticemail implements the native pm Elastic Email connector. It is a
// declarative-HTTP per-system connector built on the same template as stripe: a
// thin package that composes the connsdk toolkit (Requester + APIKeyHeader auth +
// RecordsAt extraction + offset/limit pagination) with Elastic Email-specific
// stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The Elastic Email v4 REST API is read-only here (no reverse-ETL writes are
// exposed) and uses the X-ElasticEmail-ApiKey header for auth. List endpoints
// return a top-level JSON array and page with limit/offset query parameters.
package elasticemail

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
	elasticEmailDefaultBaseURL  = "https://api.elasticemail.com/v4"
	elasticEmailDefaultPageSize = 100
	elasticEmailMaxPageSize     = 1000
	elasticEmailUserAgent       = "polymetrics-go-cli"
	elasticEmailAPIKeyHeader    = "X-ElasticEmail-ApiKey"
	// elasticEmailFixtureDate is the deterministic timestamp used by fixture-mode
	// records (2026-01-01T00:00:00Z).
	elasticEmailFixtureDate = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("elasticemail", New)
}

// New returns the Elastic Email connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Elastic Email connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "elasticemail" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "elasticemail",
		DisplayName:     "Elastic Email",
		IntegrationType: "api",
		Description:     "Reads Elastic Email contacts, campaigns, lists, segments, and templates through the Elastic Email v4 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Elastic
// Email. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := elasticEmailBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(elasticEmailSecret(cfg)) == "" {
		return errors.New("elasticemail connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the lists endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "lists", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check elasticemail: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: elasticEmailStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (full sync). Elastic Email's v4 list endpoints are
// full-refresh, so the cursor is informational only.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := elasticEmailStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("elasticemail stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := elasticEmailPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := elasticEmailMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Elastic Email v4 list endpoints return a top-level JSON array and page via
	// limit/offset. connsdk.OffsetPaginator advances offset until a short page is
	// returned; RecordsAt with an empty path selects the root array.
	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    pageSize,
	}
	mapped := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, url.Values{}, paginator, "", maxPages, mapped); err != nil {
		return fmt.Errorf("read elasticemail %s: %w", endpoint.resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise elasticemail credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"Email":            fmt.Sprintf("fixture+%d@example.com", i),
			"FirstName":        fmt.Sprintf("Fixture%d", i),
			"LastName":         "User",
			"Status":           "Active",
			"Source":           "Fixture",
			"DateAdded":        elasticEmailFixtureDate,
			"DateUpdated":      elasticEmailFixtureDate,
			"StatusChangeDate": elasticEmailFixtureDate,
			"Name":             fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"Status_":          "Active",
			"ListName":         fmt.Sprintf("List %d", i),
			"PublicListID":     fmt.Sprintf("pub_%d", i),
			"AllowUnsubscribe": true,
			"Rule":             "Status = Active",
			"Subject":          fmt.Sprintf("Fixture subject %d", i),
			"TemplateScope":    "Personal",
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

// requester builds a connsdk.Requester wired with API-key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := elasticEmailBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := elasticEmailSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("elasticemail connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(elasticEmailAPIKeyHeader, secret, ""),
		UserAgent: elasticEmailUserAgent,
	}, nil
}

func elasticEmailSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// elasticEmailBaseURL resolves and validates the base URL. The default is
// api.elasticemail.com/v4; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func elasticEmailBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return elasticEmailDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("elasticemail config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("elasticemail config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("elasticemail config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func elasticEmailPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return elasticEmailDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("elasticemail config page_size must be an integer: %w", err)
	}
	if value < 1 || value > elasticEmailMaxPageSize {
		return 0, fmt.Errorf("elasticemail config page_size must be between 1 and %d", elasticEmailMaxPageSize)
	}
	return value, nil
}

func elasticEmailMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("elasticemail config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("elasticemail config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. Elastic Email is exposed
// read-only here: no reverse-ETL write actions are allow-listed, so every write
// is rejected as an unsupported operation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
