// Package nylas implements the native pm Nylas connector. It is a declarative-
// HTTP per-system connector following the stripe reference shape: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + body-cursor pagination) with Nylas-specific stream definitions
// and grant-scoped endpoints.
//
// Nylas v3 lists are read from /v3/grants/{grant_id}/<resource>, authenticate
// with the api_key as a Bearer token, return records under data[], and paginate
// with a next_cursor body token supplied back as the page_token query param.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. The connector is read-only.
package nylas

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
	nylasDefaultBaseURL  = "https://api.us.nylas.com"
	nylasEUBaseURL       = "https://api.eu.nylas.com"
	nylasDefaultGrantID  = "me"
	nylasDefaultPageSize = 50
	nylasMaxPageSize     = 200
	nylasUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("nylas", New)
}

// New returns the Nylas connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Nylas connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nylas" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nylas",
		DisplayName:     "Nylas",
		IntegrationType: "api",
		Description:     "Reads Nylas calendars, contacts, messages, and events for a connected grant through the Nylas v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Nylas. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nylasBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(nylasSecret(cfg)) == "" {
		return errors.New("nylas connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the calendars list confirms auth and connectivity
	// without mutating anything.
	path := grantPath(cfg, "calendars")
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check nylas: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Nylas is a read-only
// source connector (no reverse-ETL writes), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nylasStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "calendars"
	}
	endpoint, ok := nylasStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("nylas stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := nylasPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := nylasMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if endpoint.requiresCalendarID {
		calendarID := strings.TrimSpace(req.Config.Config["calendar_id"])
		if calendarID == "" {
			return errors.New("nylas events stream requires config calendar_id")
		}
		base.Set("calendar_id", calendarID)
	}
	path := grantPath(req.Config, endpoint.resource)
	return c.harvest(ctx, r, path, endpoint, base, maxPages, emit)
}

// harvest drives Nylas's cursor pagination. v3 lists return
// {data:[...], next_cursor:"..."}; the next page is requested with
// page_token=<next_cursor>. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt, because the body-token-to-query-param
// shape is connector-specific.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if pageToken != "" {
			query.Set("page_token", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read nylas %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode nylas %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextCursor, err := connsdk.StringAt(resp.Body, "next_cursor")
		if err != nil {
			return fmt.Errorf("decode nylas %s next_cursor: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(nextCursor) == "" {
			return nil
		}
		pageToken = nextCursor
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise nylas credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"grant_id":     nylasDefaultGrantID,
			"object":       strings.TrimSuffix(stream, "s"),
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"timezone":     "UTC",
			"is_primary":   i == 1,
			"read_only":    false,
			"given_name":   fmt.Sprintf("Fixture%d", i),
			"surname":      "Example",
			"company_name": "Example Inc",
			"subject":      fmt.Sprintf("Fixture message %d", i),
			"snippet":      "fixture body",
			"thread_id":    "thread_fixture_1",
			"date":         int64(1767225600 + i),
			"unread":       i == 1,
			"starred":      false,
			"calendar_id":  "primary",
			"title":        fmt.Sprintf("Fixture event %d", i),
			"status":       "confirmed",
			"busy":         true,
			"updated_at":   int64(1767225600 + i),
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
	base, err := nylasBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := nylasSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("nylas connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: nylasUserAgent,
	}, nil
}

func nylasSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// grantPath builds the grant-scoped resource path, e.g.
// /v3/grants/me/calendars. The grant id is taken from config (default "me").
func grantPath(cfg connectors.RuntimeConfig, resource string) string {
	grant := strings.TrimSpace(cfg.Config["grant_id"])
	if grant == "" {
		grant = nylasDefaultGrantID
	}
	return "v3/grants/" + url.PathEscape(grant) + "/" + resource
}

// nylasBaseURL resolves and validates the base URL. The default is
// api.us.nylas.com; an api_server of "eu" selects the EU host. Any explicit
// base_url override must be an absolute https (or http for local test servers)
// URL with a host to bound SSRF risk.
func nylasBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		switch strings.ToLower(strings.TrimSpace(cfg.Config["api_server"])) {
		case "eu":
			return nylasEUBaseURL, nil
		default:
			return nylasDefaultBaseURL, nil
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("nylas config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nylas config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nylas config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func nylasPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return nylasDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nylas config page_size must be an integer: %w", err)
	}
	if value < 1 || value > nylasMaxPageSize {
		return 0, fmt.Errorf("nylas config page_size must be between 1 and %d", nylasMaxPageSize)
	}
	return value, nil
}

func nylasMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nylas config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("nylas config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
