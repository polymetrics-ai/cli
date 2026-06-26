// Package mailerlite implements the native pm MailerLite connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package composing the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + cursor state) with MailerLite-specific stream definitions and
// endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The MailerLite v2 API (https://connect.mailerlite.com/api) wraps every list
// response in {data:[...], meta:{next_cursor:...}} and authenticates with a
// bearer token, so the read path is uniform across streams.
package mailerlite

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
	mailerliteDefaultBaseURL  = "https://connect.mailerlite.com/api"
	mailerliteDefaultPageSize = 100
	mailerliteMaxPageSize     = 1000
	mailerliteUserAgent       = "polymetrics-go-cli"
	// mailerliteFixtureCreated is the deterministic created_at timestamp used by
	// fixture-mode records.
	mailerliteFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("mailerlite", New)
}

// New returns the MailerLite connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm MailerLite connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailerlite" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailerlite",
		DisplayName:     "MailerLite",
		IntegrationType: "api",
		Description:     "Reads MailerLite subscribers, campaigns, groups, segments, and automations through the MailerLite v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to MailerLite.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mailerliteBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mailerliteSecret(cfg)) == "" {
		return errors.New("mailerlite connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the subscribers list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "subscribers", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check mailerlite: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. MailerLite is wired as a
// read-only source (Capabilities.Write=false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mailerliteStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a MailerLite stream starts
// with an empty incremental cursor (full sync).
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
		stream = "subscribers"
	}
	endpoint, ok := mailerliteStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailerlite stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mailerlitePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mailerliteMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives MailerLite's cursor pagination. List responses are
// {data:[...], meta:{next_cursor:...}}; the next page is requested with
// cursor=<meta.next_cursor> until next_cursor is null/empty. The loop is built
// on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
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
			return fmt.Errorf("read mailerlite %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode mailerlite %s page: %w", endpoint.resource, err)
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
			return fmt.Errorf("decode mailerlite %s next_cursor: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		// A null next_cursor stringifies to "" and signals the last page. A
		// non-advancing cursor is also treated as terminal to avoid loops.
		if next == "" || next == cursor || len(records) == 0 {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailerlite credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"status":             "active",
			"source":             "api",
			"name":               fmt.Sprintf("Fixture %s %d", stream, i),
			"type":               "regular",
			"enabled":            true,
			"is_stopped":         false,
			"sent":               i,
			"opens_count":        i,
			"clicks_count":       i,
			"active_count":       i,
			"sent_count":         i,
			"unsubscribed_count": 0,
			"total":              i,
			"created_at":         mailerliteFixtureCreated,
			"updated_at":         mailerliteFixtureCreated,
			"subscribed_at":      mailerliteFixtureCreated,
			"connector":          "mailerlite",
			"fixture":            true,
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
	base, err := mailerliteBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mailerliteSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mailerlite connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: mailerliteUserAgent,
	}, nil
}

func mailerliteSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// mailerliteBaseURL resolves and validates the base URL. The default is
// connect.mailerlite.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func mailerliteBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mailerliteDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mailerlite config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mailerlite config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mailerlite config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mailerlitePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mailerliteDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailerlite config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mailerliteMaxPageSize {
		return 0, fmt.Errorf("mailerlite config page_size must be between 1 and %d", mailerliteMaxPageSize)
	}
	return value, nil
}

func mailerliteMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailerlite config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mailerlite config max_pages must be 0 for unlimited or a positive integer")
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
