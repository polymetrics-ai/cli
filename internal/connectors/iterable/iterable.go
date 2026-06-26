// Package iterable implements a read-only Iterable API connector. Iterable list
// endpoints are not uniformly paginated, so this connector follows nextPageToken
// only when the response supplies it and otherwise treats the first page as full.
package iterable

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
	iterableName            = "iterable"
	iterableDefaultBaseURL  = "https://api.iterable.com/api"
	iterableDefaultPageSize = 100
	iterableMaxPageSize     = 1000
	iterableUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("iterable", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return iterableName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: iterableName, DisplayName: "Iterable", IntegrationType: "api", Description: "Reads Iterable lists, campaigns, and templates through the Iterable REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(iterableAPIKey(cfg)) == "" {
		return errors.New("iterable connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "lists", url.Values{"pageSize": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check iterable: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: iterableStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "lists"
	}
	endpoint, ok := iterableStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("iterable stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := iterablePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := iterableMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint iterableStreamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{}
		query.Set("pageSize", strconv.Itoa(pageSize))
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read iterable %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode iterable %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode iterable nextPageToken: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint iterableStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "name": fmt.Sprintf("Fixture %s %d", endpoint.recordsPath, i), "listType": "Standard", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z"}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := iterableBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := iterableAPIKey(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("iterable connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Api-Key", key, ""), UserAgent: iterableUserAgent}, nil
}

type iterableStreamEndpoint struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var iterableStreamEndpoints = map[string]iterableStreamEndpoint{
	"lists":     {resource: "lists", recordsPath: "lists", mapRecord: iterableListRecord},
	"campaigns": {resource: "campaigns", recordsPath: "campaigns", mapRecord: iterableCampaignRecord},
	"templates": {resource: "templates", recordsPath: "templates", mapRecord: iterableTemplateRecord},
}

func iterableStreams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "createdAt", Type: "string"}, {Name: "updatedAt", Type: "string"}}
	return []connectors.Stream{
		{Name: "lists", Description: "Iterable static and dynamic lists.", PrimaryKey: []string{"id"}, Fields: append(fields, connectors.Field{Name: "listType", Type: "string"})},
		{Name: "campaigns", Description: "Iterable campaigns.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "templates", Description: "Iterable templates.", PrimaryKey: []string{"id"}, Fields: fields},
	}
}

func iterableListRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "listType": item["listType"], "createdAt": item["createdAt"], "updatedAt": item["updatedAt"]}
}

func iterableCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "createdAt": item["createdAt"], "updatedAt": item["updatedAt"]}
}

func iterableTemplateRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "createdAt": item["createdAt"], "updatedAt": item["updatedAt"]}
}

func iterableAPIKey(cfg connectors.RuntimeConfig) string { return cfg.Secrets["api_key"] }

func iterableBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validBaseURL(iterableName, cfg.Config["base_url"], iterableDefaultBaseURL)
}

func iterablePageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(iterableName, cfg.Config["page_size"], iterableDefaultPageSize, 1, iterableMaxPageSize)
}

func iterableMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(iterableName, cfg.Config["max_pages"])
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
