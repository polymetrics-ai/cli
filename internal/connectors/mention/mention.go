// Package mention implements the native pm Mention connector. Mention is a social
// listening and media monitoring tool; this connector reads accounts, alerts,
// mentions, and alert tags from the Mention REST API.
//
// It follows the declarative-HTTP template established by the stripe connector: a
// thin package that composes the connsdk toolkit (Requester + API-key header auth
// + RecordsAt extraction + a body-cursor pagination loop) with Mention-specific
// stream definitions and endpoints. It self-registers with the connectors
// registry via RegisterFactory in init().
package mention

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
	mentionDefaultBaseURL  = "https://api.mention.net/api"
	mentionDefaultPageSize = 100
	mentionMaxPageSize     = 1000
	mentionUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("mention", New)
}

// New returns the Mention connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mention connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mention" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mention",
		DisplayName:     "Mention",
		IntegrationType: "api",
		Description:     "Reads Mention accounts, alerts, mentions, and alert tags from the Mention social listening REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mention. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mentionBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mentionSecret(cfg)) == "" {
		return errors.New("mention connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// /accounts/me confirms auth and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "accounts/me", nil, nil, nil); err != nil {
		return fmt.Errorf("check mention: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Mention is read-only for
// reverse ETL purposes, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mentionStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "account_me"
	}
	endpoint, ok := mentionStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mention stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := c.resolvePath(ctx, r, endpoint.scope, req.Config)
	if err != nil {
		return err
	}
	if !endpoint.paginated {
		return c.readSingle(ctx, r, path, endpoint, emit)
	}
	pageSize, err := mentionPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mentionMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, endpoint, pageSize, maxPages, emit)
}

// resolvePath builds the API path for a scope, resolving the account id (via
// /accounts/me when not configured) and alert id (from config) as needed.
func (c Connector) resolvePath(ctx context.Context, r *connsdk.Requester, s scope, cfg connectors.RuntimeConfig) (string, error) {
	switch s {
	case scopeAccountMe:
		return "accounts/me", nil
	case scopeAccount:
		account, err := c.resolveAccountID(ctx, r, cfg)
		if err != nil {
			return "", err
		}
		return "accounts/" + url.PathEscape(account), nil
	case scopeAccountAlerts:
		account, err := c.resolveAccountID(ctx, r, cfg)
		if err != nil {
			return "", err
		}
		return "accounts/" + url.PathEscape(account) + "/alerts", nil
	case scopeAlertMentions, scopeAlertTags:
		account, err := c.resolveAccountID(ctx, r, cfg)
		if err != nil {
			return "", err
		}
		alert := strings.TrimSpace(cfg.Config["alert_id"])
		if alert == "" {
			return "", errors.New("mention config alert_id is required for the mention and alert_tag streams")
		}
		suffix := "/mentions"
		if s == scopeAlertTags {
			suffix = "/tags"
		}
		return "accounts/" + url.PathEscape(account) + "/alerts/" + url.PathEscape(alert) + suffix, nil
	default:
		return "", fmt.Errorf("mention: unknown stream scope %d", s)
	}
}

// resolveAccountID returns the configured account_id, or discovers it from
// /accounts/me when not configured.
func (c Connector) resolveAccountID(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig) (string, error) {
	if id := strings.TrimSpace(cfg.Config["account_id"]); id != "" {
		return id, nil
	}
	resp, err := r.Do(ctx, http.MethodGet, "accounts/me", nil, nil)
	if err != nil {
		return "", fmt.Errorf("mention discover account: %w", err)
	}
	id, err := connsdk.StringAt(resp.Body, "account.id")
	if err != nil {
		return "", fmt.Errorf("mention decode /accounts/me: %w", err)
	}
	if strings.TrimSpace(id) == "" {
		return "", errors.New("mention: /accounts/me returned no account id; set config account_id")
	}
	return id, nil
}

// readSingle reads an unpaginated endpoint and emits its records.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read mention %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.fieldPath)
	if err != nil {
		return fmt.Errorf("decode mention %s: %w", path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives Mention's cursor pagination. List responses carry the next
// cursor at _links.more.params.cursor; it is absent on the final page. The cursor
// is supplied to the next request as the `cursor` query param. connsdk has no
// paginator for this exact nested-body shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read mention %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.fieldPath)
		if err != nil {
			return fmt.Errorf("decode mention %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "_links.more.params.cursor")
		if err != nil {
			return fmt.Errorf("decode mention %s cursor: %w", path, err)
		}
		if strings.TrimSpace(next) == "" || next == cursor {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mention credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"title":        fmt.Sprintf("Fixture mention %d", i),
			"description":  "fixture record",
			"url":          fmt.Sprintf("https://example.com/m/%d", i),
			"language":     "en",
			"timezone":     "UTC",
			"permission":   "owner",
			"source_name":  "example.com",
			"source_type":  "web",
			"tone":         0,
			"favorite":     false,
			"color":        "#ff0000",
			"languages":    []any{"en"},
			"countries":    []any{"US"},
			"sources":      []any{"web"},
			"published_at": "2026-01-01T00:00:00Z",
			"created_at":   "2026-01-01T00:00:00Z",
			"updated_at":   "2026-01-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Mention's API-key header auth
// and the resolved base URL. Mention expects the raw api_key in the Authorization
// header (no Bearer prefix). The secret only ever flows into connsdk auth; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mentionBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mentionSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mention connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: mentionUserAgent,
	}, nil
}

func mentionSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// mentionBaseURL resolves and validates the base URL. The default is
// api.mention.net; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mentionBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mentionDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mention config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mention config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mention config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mentionPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mentionDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mention config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mentionMaxPageSize {
		return 0, fmt.Errorf("mention config page_size must be between 1 and %d", mentionMaxPageSize)
	}
	return value, nil
}

func mentionMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mention config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mention config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
