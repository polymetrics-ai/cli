// Package sendgrid implements the native pm SendGrid connector. It is a
// declarative-HTTP per-system connector: a thin package that composes the
// connsdk toolkit (Requester + Bearer auth + RecordsAt extraction) with
// SendGrid-specific stream definitions, endpoints, and pagination styles. It
// follows the stripe reference connector shape.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. SendGrid is read-only.
package sendgrid

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
	sendgridDefaultBaseURL  = "https://api.sendgrid.com/v3"
	sendgridDefaultPageSize = 100
	sendgridMaxPageSize     = 1000
	sendgridUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("sendgrid", New)
}

// New returns the SendGrid connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm SendGrid connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "sendgrid" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "sendgrid",
		DisplayName:     "Sendgrid",
		IntegrationType: "api",
		Description:     "Reads SendGrid Marketing Campaigns lists, segments, and contacts, plus suppression bounces, through the SendGrid v3 REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to SendGrid. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := sendgridBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(sendgridSecret(cfg)) == "" {
		return errors.New("sendgrid connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the marketing lists endpoint confirms auth and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "marketing/lists", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check sendgrid: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. SendGrid is read-only, so
// writes are unsupported and the connector advertises Capabilities.Write=false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: sendgridStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "lists"
	}
	endpoint, ok := sendgridStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("sendgrid stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := sendgridPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := sendgridMaxPages(req.Config)
	if err != nil {
		return err
	}

	switch endpoint.pagination {
	case offsetLimit:
		return c.harvestOffset(ctx, r, endpoint, pageSize, maxPages, emit)
	default:
		return c.harvestMetadataNext(ctx, r, endpoint, pageSize, maxPages, emit)
	}
}

// harvestMetadataNext drives SendGrid's marketing-API cursor pagination. Marketing
// list endpoints return {<recordsPath>:[...], _metadata:{next:"<full url>"}}; the
// next page is fetched by following the _metadata.next absolute URL until it is
// absent. connsdk.Requester treats an http(s) path as absolute, so the next URL
// is passed straight through.
func (c Connector) harvestMetadataNext(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := url.Values{}
	query.Set("page_size", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read sendgrid %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode sendgrid %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "_metadata.next")
		if err != nil {
			return fmt.Errorf("decode sendgrid %s next: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// The next link already carries page_token/page_size in its query; pass
		// it through as an absolute URL with no extra params.
		path = next
		query = nil
	}
	return nil
}

// harvestOffset drives limit/offset pagination over endpoints that return a
// top-level JSON array (e.g. suppression/bounces). It stops when a page returns
// fewer records than the page size.
func (c Connector) harvestOffset(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read sendgrid %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode sendgrid %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise sendgrid credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":              fmt.Sprintf("Fixture %s %d", stream, i),
			"contact_count":     int64(i),
			"contacts_count":    int64(i),
			"query_version":     "2",
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":        "Fixture",
			"last_name":         strconv.Itoa(i),
			"phone_number":      "",
			"created":           int64(1767225600 + i),
			"created_at":        "2026-01-01T00:00:00Z",
			"updated_at":        "2026-01-01T00:00:00Z",
			"sample_updated_at": "2026-01-01T00:00:00Z",
			"reason":            "550 5.1.1 user unknown",
			"status":            "5.1.1",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := sendgridBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := sendgridSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("sendgrid connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: sendgridUserAgent,
	}, nil
}

func sendgridSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// sendgridBaseURL resolves and validates the base URL. The default is
// api.sendgrid.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func sendgridBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return sendgridDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("sendgrid config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("sendgrid config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("sendgrid config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func sendgridPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return sendgridDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("sendgrid config page_size must be an integer: %w", err)
	}
	if value < 1 || value > sendgridMaxPageSize {
		return 0, fmt.Errorf("sendgrid config page_size must be between 1 and %d", sendgridMaxPageSize)
	}
	return value, nil
}

func sendgridMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("sendgrid config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("sendgrid config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
