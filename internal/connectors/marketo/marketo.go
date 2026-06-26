// Package marketo implements a conservative read-only Marketo REST connector.
// Marketo identity hosts are tenant-specific; this connector expects a
// caller-supplied access_token and does not refresh OAuth tokens internally.
package marketo

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
	marketoName            = "marketo"
	marketoDefaultPageSize = 300
	marketoMaxPageSize     = 300
	marketoUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("marketo", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return marketoName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: marketoName, DisplayName: "Marketo", IntegrationType: "api", Description: "Reads Marketo leads, programs, and activities through Marketo REST endpoints. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := marketoBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(marketoAccessToken(cfg)) == "" {
		return errors.New("marketo connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "leads.json", url.Values{"batchSize": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check marketo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: marketoStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "leads"
	}
	endpoint, ok := marketoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("marketo stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := marketoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := marketoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req.Config, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint marketoStreamEndpoint, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	nextToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{}
		query.Set("batchSize", strconv.Itoa(pageSize))
		if nextToken != "" {
			query.Set("nextPageToken", nextToken)
		}
		if endpoint.resource == "activities.json" && strings.TrimSpace(cfg.Config["activity_type_ids"]) != "" {
			query.Set("activityTypeIds", strings.TrimSpace(cfg.Config["activity_type_ids"]))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read marketo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "result")
		if err != nil {
			return fmt.Errorf("decode marketo %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode marketo nextPageToken: %w", err)
		}
		more, err := connsdk.StringAt(resp.Body, "moreResult")
		if err != nil {
			return fmt.Errorf("decode marketo moreResult: %w", err)
		}
		if strings.TrimSpace(next) == "" || !strings.EqualFold(strings.TrimSpace(more), "true") {
			return nil
		}
		nextToken = next
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint marketoStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": 100 + i, "email": fmt.Sprintf("fixture%d@example.com", i), "name": fmt.Sprintf("Fixture %s %d", endpoint.resource, i), "updatedAt": "2026-01-01T00:00:00Z", "createdAt": "2026-01-01T00:00:00Z"}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := marketoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := marketoAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("marketo connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: marketoUserAgent}, nil
}

type marketoStreamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var marketoStreamEndpoints = map[string]marketoStreamEndpoint{
	"leads":      {resource: "leads.json", mapRecord: marketoLeadRecord},
	"programs":   {resource: "programs.json", mapRecord: marketoProgramRecord},
	"activities": {resource: "activities.json", mapRecord: marketoActivityRecord},
}

func marketoStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "leads", Description: "Marketo leads from leads.json.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "email", Type: "string"}, {Name: "updatedAt", Type: "string"}, {Name: "createdAt", Type: "string"}}},
		{Name: "programs", Description: "Marketo programs from programs.json.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "updatedAt", Type: "string"}, {Name: "createdAt", Type: "string"}}},
		{Name: "activities", Description: "Marketo activities from activities.json. Configure activity_type_ids for live reads.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "activityTypeId", Type: "integer"}, {Name: "activityDate", Type: "string"}, {Name: "leadId", Type: "integer"}}},
	}
}

func marketoLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "email": item["email"], "updatedAt": item["updatedAt"], "createdAt": item["createdAt"]}
}

func marketoProgramRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "updatedAt": item["updatedAt"], "createdAt": item["createdAt"]}
}

func marketoActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "activityTypeId": item["activityTypeId"], "activityDate": item["activityDate"], "leadId": item["leadId"]}
}

func marketoAccessToken(cfg connectors.RuntimeConfig) string { return cfg.Secrets["access_token"] }

func marketoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("marketo connector requires config base_url ending in /rest/v1")
	}
	return validBaseURL(marketoName, base, "")
}

func marketoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(marketoName, cfg.Config["page_size"], marketoDefaultPageSize, 1, marketoMaxPageSize)
}

func marketoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(marketoName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func validBaseURL(name, raw, fallback string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(name, raw string, fallback, min, max int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s config value must be an integer: %w", name, err)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("%s config value must be between %d and %d", name, min, max)
	}
	return value, nil
}

func maxPagesConfig(name, raw string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", name, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be 0 for unlimited or a positive integer", name)
	}
	return value, nil
}
