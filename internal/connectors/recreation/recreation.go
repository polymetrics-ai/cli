// Package recreation implements a read-only Recreation.gov RIDB connector. The
// API uses an apikey header and returns paged records under RECDATA.
package recreation

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
	recreationDefaultBaseURL  = "https://ridb.recreation.gov/api/v1"
	recreationDefaultPageSize = 50
	recreationMaxPageSize     = 500
	recreationUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("recreation", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "recreation" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "recreation", DisplayName: "Recreation.gov", IntegrationType: "api", Description: "Reads Recreation.gov RIDB facilities, campsites, activities, organizations, and recreation areas. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "facilities", url.Values{"limit": []string{"1"}, "offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check recreation: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: recreationStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "facilities"
	}
	endpoint, ok := recreationEndpoints[stream]
	if !ok {
		return fmt.Errorf("recreation stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
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
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read recreation %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "RECDATA")
		if err != nil {
			return fmt.Errorf("decode recreation %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		totalRaw, _ := connsdk.StringAt(resp.Body, "METADATA.RESULTS.TOTAL_COUNT")
		total, _ := strconv.Atoi(totalRaw)
		offset += len(records)
		if len(records) < pageSize || (total > 0 && offset >= total) {
			return nil
		}
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"FacilityID": strconv.Itoa(i), "FacilityName": fmt.Sprintf("Fixture Facility %d", i), "FacilityTypeDescription": "Campground", "CampsiteID": strconv.Itoa(i), "CampsiteName": fmt.Sprintf("Fixture Campsite %d", i), "ActivityID": strconv.Itoa(i), "ActivityName": fmt.Sprintf("Fixture Activity %d", i), "OrgID": strconv.Itoa(i), "OrgName": fmt.Sprintf("Fixture Org %d", i), "RecAreaID": strconv.Itoa(i), "RecAreaName": fmt.Sprintf("Fixture Area %d", i), "LastUpdatedDate": "2026-01-01"}
		if err := emit(endpoint.mapRecord(item)); err != nil {
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
	key := strings.TrimSpace(secret(cfg, "api_key"))
	if key == "" {
		return nil, errors.New("recreation connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("apikey", key, ""), UserAgent: recreationUserAgent}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var recreationEndpoints = map[string]streamEndpoint{
	"facilities":    {path: "facilities", mapRecord: facilityRecord},
	"campsites":     {path: "campsites", mapRecord: campsiteRecord},
	"activities":    {path: "activities", mapRecord: activityRecord},
	"organizations": {path: "organizations", mapRecord: organizationRecord},
	"recareas":      {path: "recareas", mapRecord: recAreaRecord},
}

func recreationStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "facilities", Description: "RIDB facilities.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "name", "type", "updated_at")},
		{Name: "campsites", Description: "RIDB campsites.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "name", "type", "updated_at")},
		{Name: "activities", Description: "RIDB activities.", PrimaryKey: []string{"id"}, Fields: fields("id", "name")},
		{Name: "organizations", Description: "RIDB organizations.", PrimaryKey: []string{"id"}, Fields: fields("id", "name")},
		{Name: "recareas", Description: "RIDB recreation areas.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "updated_at")},
	}
}

func facilityRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["FacilityID"], "name": item["FacilityName"], "type": item["FacilityTypeDescription"], "updated_at": item["LastUpdatedDate"]}
}
func campsiteRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["CampsiteID"], "name": item["CampsiteName"], "type": item["CampsiteType"], "updated_at": item["LastUpdatedDate"]}
}
func activityRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["ActivityID"], "name": item["ActivityName"]}
}
func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["OrgID"], "name": item["OrgName"]}
}
func recAreaRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["RecAreaID"], "name": item["RecAreaName"], "updated_at": item["LastUpdatedDate"]}
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return recreationDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("recreation config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("recreation config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("recreation config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return recreationDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > recreationMaxPageSize {
		return 0, fmt.Errorf("recreation config page_size must be between 1 and %d", recreationMaxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("recreation config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
