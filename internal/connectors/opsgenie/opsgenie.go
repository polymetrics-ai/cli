// Package opsgenie implements the native pm Opsgenie connector. It is a
// read-only connector that composes the connsdk toolkit (Requester + GenieKey
// auth + data[] extraction) with Opsgenie-specific streams and paging.next
// pagination.
package opsgenie

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
	opsgenieDefaultBaseURL  = "https://api.opsgenie.com/v2"
	opsgenieDefaultPageSize = 100
	opsgenieMaxPageSize     = 100
	opsgenieUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("opsgenie", New)
}

// New returns the Opsgenie connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Opsgenie connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "opsgenie" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "opsgenie",
		DisplayName:     "Opsgenie",
		IntegrationType: "api",
		Description:     "Reads Opsgenie alerts, incidents, users, teams, and services through the Opsgenie REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Opsgenie. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := opsgenieBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(opsgenieSecret(cfg)) == "" {
		return errors.New("opsgenie connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of alerts confirms auth and connectivity without mutating
	// anything.
	if err := r.DoJSON(ctx, http.MethodGet, "alerts", url.Values{"limit": []string{"1"}, "offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check opsgenie: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: opsgenieStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader with an empty cursor for full
// refresh reads.
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
		stream = "alerts"
	}
	endpoint, ok := opsgenieStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("opsgenie stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := opsgeniePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := opsgenieMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Opsgenie's paging.next pagination. List responses are shaped
// {data:[...], paging:{next:"<full url>"}}; the next page is fetched verbatim
// and carries its own offset/limit query parameters.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := url.Values{}
	query.Set("limit", strconv.Itoa(pageSize))
	query.Set("offset", "0")

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read opsgenie %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode opsgenie %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return fmt.Errorf("decode opsgenie %s paging: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "null" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise opsgenie credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"tinyId":           strconv.Itoa(i),
			"alias":            fmt.Sprintf("fixture-%s-%d", stream, i),
			"message":          fmt.Sprintf("Fixture %s %d", stream, i),
			"description":      "Deterministic fixture record.",
			"status":           "open",
			"priority":         "P3",
			"createdAt":        fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updatedAt":        fmt.Sprintf("2026-01-0%dT01:00:00Z", i),
			"lastOccurredAt":   fmt.Sprintf("2026-01-0%dT02:00:00Z", i),
			"username":         fmt.Sprintf("fixture+%d@example.com", i),
			"fullName":         fmt.Sprintf("Fixture User %d", i),
			"role":             map[string]any{"name": "User"},
			"name":             fmt.Sprintf("Fixture %s %d", stream, i),
			"teamId":           "team_fixture_1",
			"ownerTeam":        map[string]any{"id": "team_fixture_1", "name": "Fixture Team"},
			"impactedServices": []any{"service_fixture_1"},
			"tags":             []any{"fixture"},
			"responders":       []any{},
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

// requester builds a connsdk.Requester wired with Opsgenie's GenieKey auth. The
// api_token secret only ever flows into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := opsgenieBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := opsgenieSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("opsgenie connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "GenieKey "),
		UserAgent: opsgenieUserAgent,
	}, nil
}

func opsgenieSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// opsgenieBaseURL resolves and validates the base URL. Overrides must be
// absolute http(s) URLs with a host to bound SSRF risk.
func opsgenieBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return opsgenieDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("opsgenie config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("opsgenie config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("opsgenie config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func opsgeniePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return opsgenieDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("opsgenie config page_size must be an integer: %w", err)
	}
	if value < 1 || value > opsgenieMaxPageSize {
		return 0, fmt.Errorf("opsgenie config page_size must be between 1 and %d", opsgenieMaxPageSize)
	}
	return value, nil
}

func opsgenieMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("opsgenie config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("opsgenie config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. Opsgenie is exposed as a
// read-only source connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
