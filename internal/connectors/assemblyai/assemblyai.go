// Package assemblyai implements the native pm AssemblyAI connector. It is a
// declarative-HTTP per-system connector following the stripe reference shape: a
// thin package that composes the connsdk toolkit (Requester + raw API-key header
// auth + RecordsAt extraction + cursor state) with AssemblyAI-specific stream
// definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// AssemblyAI's API is read-only here (no safe reverse-ETL writes), so the
// connector exposes Read/Check/Catalog only.
package assemblyai

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
	assemblyaiDefaultBaseURL  = "https://api.assemblyai.com"
	assemblyaiDefaultPageSize = 100
	assemblyaiMaxPageSize     = 200
	assemblyaiUserAgent       = "polymetrics-go-cli"
	// assemblyaiFixtureCreated is the deterministic `created` timestamp used by
	// fixture-mode records.
	assemblyaiFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("assemblyai", New)
}

// New returns the AssemblyAI connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm AssemblyAI connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "assemblyai" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "assemblyai",
		DisplayName:     "AssemblyAI",
		IntegrationType: "api",
		Description:     "Reads AssemblyAI transcripts and per-transcript sentence, paragraph, and subtitle references through the AssemblyAI REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to AssemblyAI.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := assemblyaiBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(assemblyaiSecret(cfg)) == "" {
		return errors.New("assemblyai connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the transcript list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v2/transcript", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check assemblyai: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: assemblyaiStreams()}, nil
}

// Write satisfies the connectors.Connector interface. AssemblyAI has no safe
// reverse-ETL write surface, so the connector is read-only and Write always
// reports the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time.
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
		stream = "transcript"
	}
	endpoint, ok := assemblyaiStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("assemblyai stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := assemblyaiPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := assemblyaiMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives AssemblyAI's next_url pagination. Transcript lists return
// {transcripts:[...], page_details:{next_url:"<absolute url>"|null}}; the next
// page is requested by following the absolute next_url. There is no body token
// paginator in connsdk for an absolute-URL next link of this shape, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First request: relative resource path with the configured page size.
	path := endpoint.resource
	var query url.Values = url.Values{"limit": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read assemblyai %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode assemblyai %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "page_details.next_url")
		if err != nil {
			return fmt.Errorf("decode assemblyai %s next_url: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// next_url is an absolute URL already carrying its own query params; the
		// Requester treats an http(s)-prefixed path as absolute and uses it as-is.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise assemblyai credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", stream, i)
		item := map[string]any{
			"id":           id,
			"status":       "completed",
			"created":      assemblyaiFixtureCreated,
			"completed":    assemblyaiFixtureCreated,
			"audio_url":    fmt.Sprintf("https://example.com/audio/%d.mp3", i),
			"resource_url": fmt.Sprintf("https://api.assemblyai.com/v2/transcript/%s", id),
			"error":        nil,
			"connector":    "assemblyai",
			"fixture":      true,
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

// requester builds a connsdk.Requester wired with raw API-key Authorization auth
// and the resolved base URL. AssemblyAI authenticates with the raw key in the
// Authorization header (no "Bearer" prefix). The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := assemblyaiBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := assemblyaiSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("assemblyai connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: assemblyaiUserAgent,
	}, nil
}

func assemblyaiSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// assemblyaiBaseURL resolves and validates the base URL. The default is
// api.assemblyai.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func assemblyaiBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return assemblyaiDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("assemblyai config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("assemblyai config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("assemblyai config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func assemblyaiPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return assemblyaiDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("assemblyai config page_size must be an integer: %w", err)
	}
	if value < 1 || value > assemblyaiMaxPageSize {
		return 0, fmt.Errorf("assemblyai config page_size must be between 1 and %d", assemblyaiMaxPageSize)
	}
	return value, nil
}

func assemblyaiMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("assemblyai config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("assemblyai config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
