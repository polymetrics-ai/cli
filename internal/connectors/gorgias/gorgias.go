// Package gorgias implements the native pm Gorgias connector. It follows the
// declarative-HTTP per-system connector shape (see internal/connectors/stripe):
// a thin package that composes the connsdk toolkit (Requester + Basic auth +
// RecordsAt extraction + cursor pagination) with Gorgias-specific stream
// definitions, endpoints, and record mappers.
//
// Gorgias is a customer-support helpdesk. The source is read-only: it reads
// tickets, customers, messages, and satisfaction surveys from the Gorgias REST
// API at https://<domain>.gorgias.com/api, authenticating with HTTP Basic auth
// (username + API key). It self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package gorgias

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
	gorgiasDefaultPageSize = 100
	gorgiasMaxPageSize     = 100
	gorgiasUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("gorgias", New)
}

// New returns the Gorgias connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Gorgias connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gorgias" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gorgias",
		DisplayName:     "Gorgias",
		IntegrationType: "api",
		Description:     "Reads Gorgias helpdesk tickets, customers, messages, and satisfaction surveys through the Gorgias REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Gorgias. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gorgiasBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gorgiasUsername(cfg)) == "" {
		return errors.New("gorgias connector requires config username")
	}
	if strings.TrimSpace(gorgiasSecret(cfg)) == "" {
		return errors.New("gorgias connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the tickets list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "api/tickets", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check gorgias: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gorgiasStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Gorgias is read-only for
// pm reverse-ETL purposes, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Gorgias stream starts with
// an empty incremental cursor (full sync).
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
		stream = "tickets"
	}
	endpoint, ok := gorgiasStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gorgias stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := gorgiasPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gorgiasMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Gorgias cursor pagination. List endpoints return
// {data:[...], meta:{next_cursor:"..."}}; the next page is requested with
// cursor=<next_cursor>. The loop is built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt; it stops when next_cursor is absent or
// null.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))

	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read gorgias %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode gorgias %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "meta.next_cursor")
		if err != nil {
			return fmt.Errorf("decode gorgias %s next_cursor: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || len(records) == 0 {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gorgias credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               int64(i),
			"ticket_id":        int64(100 + i),
			"customer_id":      int64(200 + i),
			"subject":          fmt.Sprintf("Fixture %s %d", stream, i),
			"status":           "open",
			"channel":          "email",
			"via":              "email",
			"priority":         "normal",
			"language":         "en",
			"is_unread":        false,
			"spam":             false,
			"from_agent":       false,
			"public":           true,
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"name":             fmt.Sprintf("Fixture %d", i),
			"firstname":        "Fixture",
			"lastname":         strconv.Itoa(i),
			"body_text":        fmt.Sprintf("fixture message %d", i),
			"stripped_text":    fmt.Sprintf("fixture message %d", i),
			"score":            int64(5),
			"scale_range":      int64(5),
			"created_datetime": "2026-01-01T00:00:0" + strconv.Itoa(i) + "Z",
			"updated_datetime": "2026-01-02T00:00:0" + strconv.Itoa(i) + "Z",
			"sent_datetime":    "2026-01-01T00:00:0" + strconv.Itoa(i) + "Z",
			"scored_datetime":  "2026-01-03T00:00:0" + strconv.Itoa(i) + "Z",
			"connector":        "gorgias",
			"fixture":          true,
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

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gorgiasBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(gorgiasUsername(cfg))
	if username == "" {
		return nil, errors.New("gorgias connector requires config username")
	}
	secret := gorgiasSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("gorgias connector requires secret password")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: gorgiasUserAgent,
	}, nil
}

func gorgiasUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func gorgiasSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// gorgiasBaseURL resolves and validates the base URL. The default is derived
// from the domain_name config as https://<domain>.gorgias.com/api. An explicit
// base_url override (used by tests and self-hosted proxies) must be an absolute
// https (or http for local test servers) URL with a host to bound SSRF risk.
func gorgiasBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("gorgias config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("gorgias config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("gorgias config base_url must include a host")
		}
		return strings.TrimRight(base, "/"), nil
	}

	domain := gorgiasDomain(cfg)
	if domain == "" {
		return "", errors.New("gorgias connector requires config domain_name (or base_url)")
	}
	if strings.ContainsAny(domain, "/:@ ") {
		return "", fmt.Errorf("gorgias config domain_name %q must be a bare subdomain", domain)
	}
	return "https://" + domain + ".gorgias.com/api", nil
}

// gorgiasDomain extracts the bare Gorgias subdomain from domain_name. Callers may
// supply just the subdomain (e.g. "acme") or the full host (e.g.
// "acme.gorgias.com"); both normalize to "acme".
func gorgiasDomain(cfg connectors.RuntimeConfig) string {
	raw := strings.TrimSpace(cfg.Config["domain_name"])
	if raw == "" {
		return ""
	}
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimSuffix(raw, "/")
	raw = strings.TrimSuffix(raw, ".gorgias.com")
	return strings.TrimSpace(raw)
}

func gorgiasPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gorgiasDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gorgias config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gorgiasMaxPageSize {
		return 0, fmt.Errorf("gorgias config page_size must be between 1 and %d", gorgiasMaxPageSize)
	}
	return value, nil
}

func gorgiasMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gorgias config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gorgias config max_pages must be 0 for unlimited or a positive integer")
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
