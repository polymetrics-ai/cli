// Package mailersend implements the native pm MailerSend connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + page/limit pagination) with MailerSend-specific stream
// definitions and endpoints.
//
// MailerSend is read-only here: its mutating endpoints send transactional email
// rather than perform safe reverse-ETL upserts, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package mailersend

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
	mailersendDefaultBaseURL  = "https://api.mailersend.com/v1"
	mailersendDefaultPageSize = 25
	mailersendMinPageSize     = 10
	mailersendMaxPageSize     = 100
	mailersendUserAgent       = "polymetrics-go-cli"
	// mailersendFixtureTime is the deterministic created_at used by fixture records.
	mailersendFixtureTime = "2026-01-01T00:00:00.000000Z"
)

func init() {
	connectors.RegisterFactory("mailersend", New)
}

// New returns the MailerSend connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm MailerSend connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailersend" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailersend",
		DisplayName:     "MailerSend",
		IntegrationType: "api",
		Description:     "Reads MailerSend email activity, domains, messages, and recipients through the MailerSend REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to MailerSend.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mailersendBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mailersendSecret(cfg)) == "" {
		return errors.New("mailersend connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the domains list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "domains", url.Values{"limit": []string{strconv.Itoa(mailersendMinPageSize)}}, nil, nil); err != nil {
		return fmt.Errorf("check mailersend: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mailersendStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a MailerSend stream starts
// with an empty cursor (full sync). MailerSend supports only full_refresh syncs,
// so the cursor is informational.
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
		stream = "domains"
	}
	endpoint, ok := mailersendStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailersend stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	base, err := baseQuery(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mailersendPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mailersendMaxPages(req.Config)
	if err != nil {
		return err
	}

	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "limit",
		StartPage: 1,
		PageSize:  pageSize,
	}
	mapRecord := endpoint.mapRecord
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, base, paginator, "data", maxPages, func(rec connsdk.Record) error {
		return emit(mapRecord(rec))
	})
}

// Write is required by the connectors.Connector interface. MailerSend is exposed
// read-only (its mutating endpoints send transactional email, not safe
// reverse-ETL upserts), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailersend credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type":          "delivered",
			"name":          fmt.Sprintf("fixture-%d.example.com", i),
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"dkim":          true,
			"spf":           true,
			"tracking":      true,
			"is_verified":   true,
			"is_dns_active": true,
			"created_at":    mailersendFixtureTime,
			"updated_at":    mailersendFixtureTime,
			"deleted_at":    nil,
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
	base, err := mailersendBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mailersendSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mailersend connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: mailersendUserAgent,
	}, nil
}

// resolveResource templates the endpoint path with domain_id for streams that
// need it (activity/{domain_id}).
func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.requiresDomain {
		return endpoint.resource, nil
	}
	domain := strings.TrimSpace(cfg.Config["domain_id"])
	if domain == "" {
		return "", errors.New("mailersend activity stream requires config domain_id")
	}
	return endpoint.resource + "/" + url.PathEscape(domain), nil
}

// baseQuery builds the per-stream query params common to every page. For the
// activity stream this includes the required date_from/date_to window. The
// optional domain_id config also filters the messages stream when present.
func baseQuery(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (url.Values, error) {
	q := url.Values{}
	if endpoint.requiresDateRange {
		from := firstNonEmpty(cfg.Config["date_from"], cfg.Config["start_date"])
		to := strings.TrimSpace(cfg.Config["date_to"])
		if from == "" || to == "" {
			return nil, errors.New("mailersend activity stream requires config date_from and date_to (unix seconds)")
		}
		q.Set("date_from", from)
		q.Set("date_to", to)
		for _, event := range splitCSV(cfg.Config["event"]) {
			q.Add("event[]", event)
		}
	}
	if endpoint.resource == "messages" {
		if domain := strings.TrimSpace(cfg.Config["domain_id"]); domain != "" {
			q.Set("domain_id", domain)
		}
	}
	return q, nil
}

func mailersendSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// mailersendBaseURL resolves and validates the base URL. The default is
// api.mailersend.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mailersendBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mailersendDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mailersend config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mailersend config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mailersend config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mailersendPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mailersendDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailersend config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mailersendMaxPageSize {
		return 0, fmt.Errorf("mailersend config page_size must be between 1 and %d", mailersendMaxPageSize)
	}
	return value, nil
}

func mailersendMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailersend config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mailersend config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
