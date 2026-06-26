// Package mixmax implements the native pm Mixmax source connector. Like stripe,
// it is a thin declarative-HTTP per-system connector: it composes the connsdk
// toolkit (Requester + APIKeyHeader auth + RecordsAt extraction + cursor state)
// with Mixmax-specific stream definitions, endpoints, and the Mixmax
// {results,next,hasNext} cursor pagination.
//
// Mixmax exposes a read-only REST API (https://api.mixmax.com/v1) authenticated
// with an X-API-Token header. There is no safe reverse-ETL write surface for the
// streams modeled here, so the connector is read-only (Capabilities.Write=false).
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package mixmax

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
	mixmaxDefaultBaseURL  = "https://api.mixmax.com/v1"
	mixmaxDefaultPageSize = 50
	mixmaxMaxPageSize     = 100
	mixmaxAuthHeader      = "X-API-Token"
	mixmaxUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("mixmax", New)
}

// New returns the Mixmax connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mixmax source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mixmax" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mixmax",
		DisplayName:     "Mixmax",
		IntegrationType: "api",
		Description:     "Reads Mixmax code snippets, messages, rules, sequences, and meeting types through the Mixmax REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mixmax. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mixmaxBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mixmaxSecret(cfg)) == "" {
		return errors.New("mixmax connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the code snippets list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "codesnippets", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check mixmax: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mixmaxStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Mixmax stream starts with
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
		stream = "codesnippets"
	}
	endpoint, ok := mixmaxStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mixmax stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mixmaxPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mixmaxMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Mixmax's body-cursor pagination. Mixmax lists return
// {results:[...], next:<token>, hasNext:<bool>}; the next page is requested with
// next=<token>. The loop lives here, built on connsdk.Requester + RecordsAt +
// StringAt, because the stop condition (hasNext==false OR an empty next token)
// does not map onto a single connsdk paginator.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	next := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if next != "" {
			query.Set("next", next)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read mixmax %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode mixmax %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasNext, err := connsdk.StringAt(resp.Body, "hasNext")
		if err != nil {
			return fmt.Errorf("decode mixmax %s hasNext: %w", endpoint.resource, err)
		}
		token, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode mixmax %s next: %w", endpoint.resource, err)
		}
		token = strings.TrimSpace(token)
		// Stop when the API signals no more pages or fails to provide a token.
		if hasNext == "false" || token == "" {
			return nil
		}
		next = token
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mixmax credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"_id":                  fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"title":                fmt.Sprintf("Fixture %d", i),
			"name":                 fmt.Sprintf("Fixture %d", i),
			"subject":              fmt.Sprintf("Fixture subject %d", i),
			"language":             "javascript",
			"theme":                "default",
			"trigger":              "manual",
			"type":                 "round-robin",
			"timezone":             "UTC",
			"userId":               "usr_fixture_1",
			"sequence":             "seq_fixture_1",
			"link":                 "https://example.com/meet",
			"durationMin":          30,
			"trackingEnabled":      true,
			"linkTrackingEnabled":  true,
			"fileTrackingEnabled":  false,
			"notificationsEnabled": true,
			"syncToOrg":            false,
			"isPaused":             false,
			"createdAt":            fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updatedAt":            fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"modifiedAt":           fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
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

// requester builds a connsdk.Requester wired with X-API-Token header auth and
// the resolved base URL. The secret only ever flows into connsdk.APIKeyHeader;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mixmaxBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mixmaxSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mixmax connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(mixmaxAuthHeader, secret, ""),
		UserAgent: mixmaxUserAgent,
	}, nil
}

func mixmaxSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// mixmaxBaseURL resolves and validates the base URL. The default is
// api.mixmax.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func mixmaxBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mixmaxDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mixmax config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mixmax config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mixmax config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mixmaxPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mixmaxDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mixmax config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mixmaxMaxPageSize {
		return 0, fmt.Errorf("mixmax config page_size must be between 1 and %d", mixmaxMaxPageSize)
	}
	return value, nil
}

func mixmaxMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mixmax config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mixmax config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Mixmax is modeled as a
// read-only source: there is no approved reverse-ETL write surface for these
// streams, so every write is rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{RecordsFailed: len(records)}, fmt.Errorf("mixmax is a read-only source: %w", connectors.ErrUnsupportedOperation)
}
