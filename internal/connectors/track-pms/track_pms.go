// Package trackpms implements a read-only native connector for Track PMS APIs.
package trackpms

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
	defaultBaseURL  = "https://api.trackhs.com"
	defaultPageSize = 100
	maxPageSize     = 500
	defaultMaxPages = 1
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("track-pms", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "track-pms" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "track-pms", DisplayName: "Track PMS", IntegrationType: "api", Description: "Reads Track PMS reservations, guests, units, and owners through API list endpoints.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "reservations", url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check track-pms: %w", err)
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
		stream = "reservations"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("track-pms stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return harvest(ctx, r, req.Config, spec, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func harvest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, spec streamSpec, emit func(connectors.Record) error) error {
	pageSize, err := pageSize(cfg)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(cfg)
	if err != nil {
		return err
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, spec.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read track-pms %s: %w", spec.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode track-pms %s: %w", spec.resource, err)
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
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "confirmation_number": fmt.Sprintf("CN-%d", i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "status": "booked", "arrival_date": "2026-01-02"}
		rec := spec.mapRecord(item)
		rec["connector"] = "track-pms"
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
	token := strings.TrimSpace(cfg.Secrets["access_token"])
	if token == "" {
		return nil, errors.New("track-pms connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("track-pms config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("track-pms config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("track-pms config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "track-pms config page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return defaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("track-pms config max_pages must be a non-negative integer: %w", err)
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
	"reservations": {resource: "reservations", recordsPath: "reservations", mapRecord: reservationRecord},
	"guests":       {resource: "guests", recordsPath: "guests", mapRecord: personRecord},
	"units":        {resource: "units", recordsPath: "units", mapRecord: unitRecord},
	"owners":       {resource: "owners", recordsPath: "owners", mapRecord: personRecord},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}}
	return []connectors.Stream{
		{Name: "reservations", Description: "Track PMS reservations.", PrimaryKey: []string{"id"}, CursorFields: []string{"arrival_date"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "confirmation_number", Type: "string"}, {Name: "status", Type: "string"}, {Name: "arrival_date", Type: "string"}}},
		{Name: "guests", Description: "Track PMS guests.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "units", Description: "Track PMS units/properties.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "owners", Description: "Track PMS owners.", PrimaryKey: []string{"id"}, Fields: fields},
	}
}

func reservationRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "confirmation_number": first(item, "confirmation_number", "confirmationNumber"), "status": item["status"], "arrival_date": first(item, "arrival_date", "arrivalDate")}
}
func personRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": first(item, "name", "full_name"), "status": item["status"]}
}
func unitRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": first(item, "name", "unit_name"), "status": item["status"]}
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
