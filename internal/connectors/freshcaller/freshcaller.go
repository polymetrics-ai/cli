// Package freshcaller implements the native pm Freshcaller connector. It is a
// small read-only HTTP connector built on connsdk with token-header auth,
// page-number pagination, deterministic fixture mode, and no write surface.
package freshcaller

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
	defaultBaseURL  = "https://api.freshcaller.com/api/v1"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("freshcaller", New)
}

// New returns the Freshcaller connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Freshcaller connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	Client *http.Client
}

func (Connector) Name() string { return "freshcaller" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "freshcaller",
		DisplayName:     "Freshcaller",
		IntegrationType: "api",
		Description:     "Reads Freshcaller calls, agents, teams, and phone numbers through the Freshcaller REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("freshcaller connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "calls", url.Values{"page": []string{"1"}, "page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check freshcaller: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "calls"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("freshcaller stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"page": []string{strconv.Itoa(page)}, "page_size": []string{strconv.Itoa(pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read freshcaller %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode freshcaller %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           i,
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"direction":    "inbound",
			"status":       "completed",
			"call_time":    fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"duration":     60 * i,
			"agent_id":     1000 + i,
			"phone_number": "+15550000000",
		}
		rec := endpoint.mapRecord(item)
		rec["connector"] = "freshcaller"
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("freshcaller connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", token, "Token token="),
		UserAgent: userAgent,
	}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("freshcaller config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("freshcaller config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("freshcaller config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "freshcaller config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("freshcaller config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}

func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
