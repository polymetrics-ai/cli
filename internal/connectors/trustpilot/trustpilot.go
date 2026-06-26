// Package trustpilot implements a read-only native connector for Trustpilot APIs.
package trustpilot

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
	defaultBaseURL  = "https://api.trustpilot.com"
	defaultPageSize = 100
	maxPageSize     = 100
	defaultMaxPages = 1
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("trustpilot", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "trustpilot" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "trustpilot", DisplayName: "Trustpilot", IntegrationType: "api", Description: "Reads Trustpilot business-unit reviews, invitations, and business-unit profile metadata.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	path, err := resourcePath(cfg, streamSpecs["reviews"])
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"perPage": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check trustpilot: %w", err)
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
		stream = "reviews"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("trustpilot stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := resourcePath(req.Config, spec)
	if err != nil {
		return err
	}
	return harvest(ctx, r, req.Config, path, spec, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func harvest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, path string, spec streamSpec, emit func(connectors.Record) error) error {
	pageSize, err := boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "trustpilot config page_size")
	if err != nil {
		return err
	}
	maxPages, err := configuredMaxPages(cfg)
	if err != nil {
		return err
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		query := url.Values{"perPage": []string{strconv.Itoa(pageSize)}, "page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read trustpilot %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode trustpilot %s: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "displayName": "Fixture Business", "stars": 5, "title": fmt.Sprintf("Fixture review %d", i), "createdAt": "2026-01-01T00:00:00Z"}
		rec := spec.mapRecord(item)
		rec["connector"] = "trustpilot"
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
	key := strings.TrimSpace(cfg.Secrets["api_key"])
	if key == "" {
		return nil, errors.New("trustpilot connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("apikey", key), UserAgent: userAgent}, nil
}

func resourcePath(cfg connectors.RuntimeConfig, spec streamSpec) (string, error) {
	if !strings.Contains(spec.resource, "%s") {
		return spec.resource, nil
	}
	businessUnitID := strings.TrimSpace(cfg.Config["business_unit_id"])
	if businessUnitID == "" {
		return "", errors.New("trustpilot config business_unit_id is required")
	}
	return fmt.Sprintf(spec.resource, url.PathEscape(businessUnitID)), nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("trustpilot config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("trustpilot config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("trustpilot config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func configuredMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return defaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("trustpilot config max_pages must be a non-negative integer: %w", err)
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

type streamSpec struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamSpecs = map[string]streamSpec{
	"reviews":        {resource: "v1/business-units/%s/reviews", recordsPath: "reviews", mapRecord: reviewRecord},
	"invitations":    {resource: "v1/private/business-units/%s/invitations", recordsPath: "invitations", mapRecord: invitationRecord},
	"business_units": {resource: "v1/business-units/%s", recordsPath: ".", mapRecord: businessUnitRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "reviews", Description: "Trustpilot business-unit reviews.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "stars", Type: "integer"}, {Name: "title", Type: "string"}, {Name: "created_at", Type: "string"}}},
		{Name: "invitations", Description: "Trustpilot review invitations.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "string"}}},
		{Name: "business_units", Description: "Trustpilot business-unit profile metadata.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "display_name", Type: "string"}}},
	}
}

func reviewRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "stars": item["stars"], "title": item["title"], "created_at": first(item, "createdAt", "created_at")}
}
func invitationRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "status": item["status"], "created_at": first(item, "createdAt", "created_at")}
}
func businessUnitRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "display_name": first(item, "displayName", "display_name")}
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
