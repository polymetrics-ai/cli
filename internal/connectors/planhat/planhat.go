// Package planhat implements a read-only Planhat API connector.
package planhat

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
	defaultBaseURL  = "https://api.planhat.com"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("planhat", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "planhat" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "planhat",
		DisplayName:     "Planhat",
		IntegrationType: "api",
		Description:     "Reads Planhat companies, end users, and licenses through the REST API.",
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
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if _, err := r.Do(ctx, http.MethodGet, "companies", url.Values{"limit": {"1"}, "offset": {"0"}}, nil); err != nil {
		return fmt.Errorf("check planhat: %w", err)
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
		stream = "companies"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("planhat stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
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
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{"limit": {strconv.Itoa(pageSize)}, "offset": {strconv.Itoa(offset)}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read planhat %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode planhat %s: %w", stream, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"companies": {path: "companies", recordsPath: ".", mapRecord: companyRecord},
	"endusers":  {path: "endusers", recordsPath: ".", mapRecord: userRecord},
	"licenses":  {path: "licenses", recordsPath: ".", mapRecord: licenseRecord},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "phase", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "companies", Description: "Planhat company records.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields},
		{Name: "endusers", Description: "Planhat end user records.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields},
		{Name: "licenses", Description: "Planhat license records.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields},
	}
}

func companyRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "_id", "id"), "name": item["name"], "email": item["email"], "phase": item["phase"], "updated_at": first(item, "updatedAt", "updated_at")}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "_id", "id"), "name": first(item, "name", "fullName"), "email": item["email"], "phase": item["status"], "updated_at": first(item, "updatedAt", "updated_at")}
}

func licenseRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "_id", "id"), "name": first(item, "name", "product"), "email": nil, "phase": item["status"], "updated_at": first(item, "updatedAt", "updated_at")}
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"_id": fmt.Sprintf("fixture-%d", i), "name": fmt.Sprintf("Fixture %d", i), "email": fmt.Sprintf("fixture%d@example.com", i), "phase": "live", "status": "active", "updatedAt": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}
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
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("planhat connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("planhat config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("planhat config base_url must be an absolute http or https URL")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "page_size", defaultPageSize, 1, 500)
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "max_pages", defaultMaxPages, 0, 10000)
}

func intConfig(cfg connectors.RuntimeConfig, key string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	if raw == "" {
		return def, nil
	}
	if key == "max_pages" && (raw == "all" || raw == "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("planhat config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func first(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
