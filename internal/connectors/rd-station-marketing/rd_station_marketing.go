// Package rdstationmarketing implements a read-only RD Station Marketing
// connector. It uses bearer auth against the platform REST endpoints and keeps
// the read surface to allow-listed list resources.
package rdstationmarketing

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
	rdName            = "rd-station-marketing"
	rdDefaultBaseURL  = "https://api.rd.services/platform"
	rdDefaultPageSize = 125
	rdMaxPageSize     = 125
	rdUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("rd-station-marketing", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return rdName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: rdName, DisplayName: "RD Station Marketing", IntegrationType: "api", Description: "Reads RD Station Marketing contacts, segmentations, events, landing pages, and email templates. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", url.Values{"page": []string{"1"}, "page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check rd-station-marketing: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: rdStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := rdEndpoints[stream]
	if !ok {
		return fmt.Errorf("rd-station-marketing stream %q not found", stream)
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
	page := 1
	for fetched := 0; maxPages == 0 || fetched < maxPages; fetched++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("page_size", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read rd-station-marketing %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode rd-station-marketing %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next_page")
		if err != nil {
			return fmt.Errorf("decode rd-station-marketing %s next_page: %w", endpoint.path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		nextPage, err := strconv.Atoi(next)
		if err != nil || nextPage <= page {
			return nil
		}
		page = nextPage
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"uuid": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "id": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "name": fmt.Sprintf("Fixture %d", i), "title": fmt.Sprintf("Fixture %d", i), "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z"}
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
	token := strings.TrimSpace(secret(cfg, "access_token"))
	if token == "" {
		return nil, errors.New("rd-station-marketing connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: rdUserAgent}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var rdEndpoints = map[string]streamEndpoint{
	"contacts":        {path: "contacts", recordsPath: "contacts", mapRecord: contactRecord},
	"segmentations":   {path: "segmentations", recordsPath: "segmentations", mapRecord: namedRecord},
	"events":          {path: "events", recordsPath: "events", mapRecord: eventRecord},
	"landing_pages":   {path: "landing_pages", recordsPath: "landing_pages", mapRecord: namedRecord},
	"email_templates": {path: "email_templates", recordsPath: "email_templates", mapRecord: namedRecord},
}

func rdStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "contacts", Description: "RD Station contacts.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "email", "name", "created_at", "updated_at")},
		{Name: "segmentations", Description: "RD Station segmentations.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "created_at")},
		{Name: "events", Description: "RD Station events.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "event_type", "email", "created_at")},
		{Name: "landing_pages", Description: "RD Station landing pages.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "created_at")},
		{Name: "email_templates", Description: "RD Station email templates.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "created_at")},
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "uuid", "id"), "email": first(item, "email"), "name": first(item, "name"), "created_at": first(item, "created_at"), "updated_at": first(item, "updated_at")}
}
func namedRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "uuid", "id"), "name": first(item, "name", "title"), "created_at": first(item, "created_at")}
}
func eventRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "uuid", "id", "event_identifier"), "event_type": first(item, "event_type", "event"), "email": first(item, "email"), "created_at": first(item, "created_at")}
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
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
		return rdDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("rd-station-marketing config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("rd-station-marketing config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("rd-station-marketing config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return rdDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > rdMaxPageSize {
		return 0, fmt.Errorf("rd-station-marketing config page_size must be between 1 and %d", rdMaxPageSize)
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
		return 0, errors.New("rd-station-marketing config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
