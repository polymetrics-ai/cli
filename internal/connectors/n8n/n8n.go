// Package n8n implements the native pm n8n connector. It is a declarative-HTTP
// per-system connector built on the connsdk toolkit (Requester + X-N8N-API-KEY
// header auth + RecordsAt extraction + nextCursor pagination) with n8n-specific
// stream definitions and endpoints. It follows the stripe reference shape.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// n8n exposes a read-only public REST API at {host}/api/v1, authenticated with an
// X-N8N-API-KEY header, paginated with ?limit=&cursor= where the response carries
// {data:[...], nextCursor:"..."}.
package n8n

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
	n8nAPIVersionPath  = "api/v1"
	n8nDefaultPageSize = 100
	n8nMaxPageSize     = 250
	n8nUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("n8n", New)
}

// New returns the n8n connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm n8n connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "n8n" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "n8n",
		DisplayName:     "n8n",
		IntegrationType: "api",
		Description:     "Reads n8n workflows, executions, tags, and users from a self-hosted or cloud n8n instance via its public REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to n8n. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := n8nBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(n8nSecret(cfg)) == "" {
		return errors.New("n8n connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the workflows list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "workflows", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check n8n: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: n8nStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an n8n stream starts with an
// empty incremental cursor (full sync).
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
		stream = "workflows"
	}
	endpoint, ok := n8nStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("n8n stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := n8nPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := n8nMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives n8n's nextCursor pagination. n8n list responses are
// {data:[...], nextCursor:"..."}; the next page is requested with
// cursor=<nextCursor>. nextCursor is null/absent on the final page. The loop
// lives here (the connsdk CursorPaginator stops on an empty token, which matches,
// but keeping it explicit lets us bound pages and reuse RecordsAt/StringAt).
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
			return fmt.Errorf("read n8n %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode n8n %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextCursor")
		if err != nil {
			return fmt.Errorf("decode n8n %s nextCursor: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == cursor {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise n8n credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		created := fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i)
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":         fmt.Sprintf("Fixture %d", i),
			"active":       i%2 == 0,
			"isArchived":   false,
			"createdAt":    created,
			"updatedAt":    created,
			"triggerCount": int64(i),
			"versionId":    fmt.Sprintf("v_%d", i),
			"workflowId":   "wf_fixture_1",
			"finished":     true,
			"mode":         "trigger",
			"status":       "success",
			"startedAt":    created,
			"stoppedAt":    created,
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":    fmt.Sprintf("First%d", i),
			"lastName":     "Fixture",
			"isPending":    false,
			"role":         "global:member",
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

// requester builds a connsdk.Requester wired with X-N8N-API-KEY header auth and
// the resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := n8nBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := n8nSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("n8n connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-N8N-API-KEY", secret, ""),
		UserAgent: n8nUserAgent,
	}, nil
}

func n8nSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets["api_key"]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Secrets["client_secret"])
}

// n8nBaseURL resolves and validates the base URL against which API paths are
// requested. Precedence: explicit base_url config, then the n8n-native host
// config (an instance URL/hostname). Either is validated for SSRF (absolute
// http/https with a host) and the /api/v1 version path is appended when absent.
func n8nBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["host"])
	}
	if raw == "" {
		return "", errors.New("n8n connector requires config base_url or host")
	}
	// Allow a bare hostname (no scheme) by defaulting to https.
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("n8n config base_url/host is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("n8n config base_url/host must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("n8n config base_url/host must include a host")
	}
	trimmed := strings.TrimRight(raw, "/")
	// Append the API version path unless the caller already supplied it.
	if strings.HasSuffix(trimmed, "/"+n8nAPIVersionPath) || strings.Contains(trimmed, "/"+n8nAPIVersionPath+"/") {
		return trimmed, nil
	}
	return trimmed + "/" + n8nAPIVersionPath, nil
}

func n8nPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return n8nDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("n8n config page_size must be an integer: %w", err)
	}
	if value < 1 || value > n8nMaxPageSize {
		return 0, fmt.Errorf("n8n config page_size must be between 1 and %d", n8nMaxPageSize)
	}
	return value, nil
}

func n8nMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("n8n config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("n8n config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: n8n is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
