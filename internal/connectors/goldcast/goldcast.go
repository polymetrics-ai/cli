// Package goldcast implements the native pm Goldcast connector. It follows the
// declarative-HTTP shape of the stripe reference connector: a thin package that
// composes the connsdk toolkit (Requester + APIKeyHeader auth + RecordsAt
// extraction) with Goldcast-specific stream definitions and endpoints.
//
// The Goldcast customapi is a Django REST Framework API at
// https://customapi.goldcast.io. Authentication is an "Authorization: Token
// <access_key>" header. List endpoints return either a raw top-level JSON array
// or a DRF envelope ({count, next, results}); this connector handles both, and
// follows the absolute "next" link for paginated envelopes. The API exposes only
// full-refresh syncs (no cursor field), so the connector is read-only with no
// incremental state.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package goldcast

import (
	"bytes"
	"context"
	"encoding/json"
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
	goldcastDefaultBaseURL = "https://customapi.goldcast.io"
	goldcastUserAgent      = "polymetrics-go-cli"
	// goldcastMaxPages bounds how many pages a single read will follow as a
	// safety valve against a server that never stops returning a next link.
	goldcastMaxPages = 10000
)

func init() {
	connectors.RegisterFactory("goldcast", New)
}

// New returns the Goldcast connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Goldcast connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "goldcast" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "goldcast",
		DisplayName:     "Goldcast",
		IntegrationType: "api",
		Description:     "Reads Goldcast organizations, events, agenda items, discussion groups, and tracks through the Goldcast customapi REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Goldcast. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := goldcastBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(goldcastSecret(cfg)) == "" {
		return errors.New("goldcast connector requires secret access_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the events list confirms auth and connectivity without
	// mutating anything.
	if _, err := r.Do(ctx, http.MethodGet, "event/", nil, nil); err != nil {
		return fmt.Errorf("check goldcast: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: goldcastStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "events"
	}
	endpoint, ok := goldcastStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("goldcast stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, emit)
}

// Write is unsupported: Goldcast is a read-only source connector (no reverse
// ETL). It satisfies the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Goldcast's read. The first request hits the stream resource;
// each response is either a raw top-level array (single page, documented common
// case) or a DRF envelope {count, next, results}. When an envelope carries a
// non-null absolute "next" URL, the loop follows it until exhausted. connsdk's
// Requester treats an absolute http(s) path as-is, so next links are followed
// directly.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	path := endpoint.resource
	for page := 0; page < goldcastMaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read goldcast %s: %w", endpoint.resource, err)
		}
		records, next, err := decodePage(resp.Body)
		if err != nil {
			return fmt.Errorf("decode goldcast %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
	}
	return nil
}

// decodePage extracts records and the next-page URL from a Goldcast list
// response. It accepts a raw top-level JSON array (records, no next) or a DRF
// envelope object with "results" (records) and "next" (absolute next URL or
// null). For any other object shape it falls back to RecordsAt's single-object
// behavior so unexpected payloads still surface a record rather than vanish.
func decodePage(body []byte) ([]map[string]any, string, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		records, err := connsdk.RecordsAt(body, "")
		if err != nil {
			return nil, "", err
		}
		return toMaps(records), "", nil
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var envelope map[string]any
	if err := dec.Decode(&envelope); err != nil {
		return nil, "", err
	}
	if results, ok := envelope["results"]; ok {
		return mapsFromAny(results), stringFromAny(envelope["next"]), nil
	}
	// Not a DRF envelope: treat the object itself as a single record.
	return []map[string]any{envelope}, "", nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise goldcast credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"title":        fmt.Sprintf("Fixture %s %d", stream, i),
			"slug":         fmt.Sprintf("fixture-%d", i),
			"domain":       "example.goldcast.io",
			"organization": "org_fixture_1",
			"event":        "event_fixture_1",
			"status":       "live",
			"description":  "Fixture record.",
			"start_time":   "2026-01-01T00:00:00Z",
			"end_time":     "2026-01-01T01:00:00Z",
			"timezone":     "UTC",
			"capacity":     int64(10 * i),
			"color":        "#112233",
			"created_at":   "2026-01-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Token-header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := goldcastBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := goldcastSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("goldcast connector requires secret access_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "Token "),
		UserAgent: goldcastUserAgent,
	}, nil
}

func goldcastSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_key"]
}

// goldcastBaseURL resolves and validates the base URL. The default is
// customapi.goldcast.io; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func goldcastBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return goldcastDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("goldcast config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("goldcast config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("goldcast config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// toMaps unwraps connsdk.Record (a map[string]any alias) values into plain maps.
func toMaps(records []connsdk.Record) []map[string]any {
	out := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		out = append(out, map[string]any(rec))
	}
	return out
}

// mapsFromAny extracts the object elements of a decoded JSON array value.
func mapsFromAny(v any) []map[string]any {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if obj, ok := item.(map[string]any); ok {
			out = append(out, obj)
		}
	}
	return out
}

// stringFromAny renders a decoded JSON scalar as a string ("" for null).
func stringFromAny(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
