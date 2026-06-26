// Package nebiusai implements the native pm Nebius AI Studio connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + cursor state) with Nebius-specific stream definitions
// and endpoints.
//
// Nebius AI Studio exposes an OpenAI-compatible REST API at
// https://api.studio.nebius.com. List endpoints return
// {"object":"list","data":[...]} and paginate with an after=<last id> cursor
// plus a has_more flag, mirroring the OpenAI/Stripe list shape.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// The directory is named "nebius-ai" (with a hyphen) to match the bare system
// name; the Go package identifier is the sanitized "nebiusai".
package nebiusai

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
	nebiusDefaultBaseURL  = "https://api.studio.nebius.com"
	nebiusDefaultPageSize = 20
	nebiusMaxPageSize     = 100
	nebiusUserAgent       = "polymetrics-go-cli"
	// nebiusFixtureCreated is the deterministic created timestamp used by the
	// fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	nebiusFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("nebius-ai", New)
}

// New returns the Nebius AI connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Nebius AI Studio connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nebius-ai" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nebius-ai",
		DisplayName:     "Nebius AI",
		IntegrationType: "api",
		Description:     "Reads Nebius AI Studio models, files, and batch jobs through the OpenAI-compatible Nebius AI Studio REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Nebius. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nebiusBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(nebiusSecret(cfg)) == "" {
		return errors.New("nebius-ai connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing models confirms auth and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v1/models", nil, nil, nil); err != nil {
		return fmt.Errorf("check nebius-ai: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nebiusStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Nebius stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "models"
	}
	endpoint, ok := nebiusStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("nebius-ai stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := nebiusPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := nebiusMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write is unsupported: the Nebius AI Studio read connector is read-only. It is
// present to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives the OpenAI-compatible list pagination. List endpoints return
// {object:"list", data:[...], has_more:bool}; the next page is requested with
// after=<last object id>. There is no body token paginator in connsdk for this
// exact shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt. Endpoints that return everything in one
// page (e.g. models) simply omit has_more and the loop stops after one page.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	if pageSize > 0 {
		base.Set("limit", strconv.Itoa(pageSize))
	}

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read nebius-ai %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode nebius-ai %s page: %w", endpoint.resource, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, "id")
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode nebius-ai %s has_more: %w", endpoint.resource, err)
		}
		if hasMore != "true" || lastID == "" {
			return nil
		}
		after = lastID
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise nebius-ai credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(stream, "s"), i),
			"object":         strings.TrimSuffix(stream, "s"),
			"created":        nebiusFixtureCreated + int64(i),
			"created_at":     nebiusFixtureCreated + int64(i),
			"completed_at":   nebiusFixtureCreated + int64(i) + 60,
			"owned_by":       "nebius",
			"filename":       fmt.Sprintf("fixture-%d.jsonl", i),
			"bytes":          int64(1024 * i),
			"purpose":        "batch",
			"status":         "completed",
			"endpoint":       "/v1/chat/completions",
			"input_file_id":  "file_fixture_1",
			"output_file_id": "file_fixture_out_1",
			"error_file_id":  "",
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
	base, err := nebiusBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := nebiusSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("nebius-ai connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: nebiusUserAgent,
	}, nil
}

func nebiusSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// nebiusBaseURL resolves and validates the base URL. The default is
// api.studio.nebius.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func nebiusBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return nebiusDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("nebius-ai config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nebius-ai config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nebius-ai config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func nebiusPageSize(cfg connectors.RuntimeConfig) (int, error) {
	// The catalog field is named "limit"; accept page_size as an alias.
	raw := strings.TrimSpace(cfg.Config["limit"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return nebiusDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nebius-ai config limit must be an integer: %w", err)
	}
	if value < 1 || value > nebiusMaxPageSize {
		return 0, fmt.Errorf("nebius-ai config limit must be between 1 and %d", nebiusMaxPageSize)
	}
	return value, nil
}

func nebiusMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nebius-ai config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("nebius-ai config max_pages must be 0 for unlimited or a positive integer")
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
