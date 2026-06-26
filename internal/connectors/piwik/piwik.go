// Package piwik implements a read-only Piwik/Matomo Reporting API connector.
package piwik

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
	defaultBaseURL  = "https://matomo.example.com"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("piwik", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "piwik" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "piwik",
		DisplayName:     "Piwik / Matomo",
		IntegrationType: "api",
		Description:     "Reads Piwik/Matomo sites, visits, page actions, and goals through the Reporting API.",
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
	query := apiQuery("SitesManager.getAllSites", cfg, false)
	if _, err := r.Do(ctx, http.MethodGet, "index.php", query, nil); err != nil {
		return fmt.Errorf("check piwik: %w", err)
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
		stream = "sites"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("piwik stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
	}
	if endpoint.requiresSite && strings.TrimSpace(siteID(req.Config)) == "" {
		return errors.New("piwik stream requires config site_id")
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
		query := apiQuery(endpoint.method, req.Config, endpoint.requiresSite)
		query.Set("filter_limit", strconv.Itoa(pageSize))
		query.Set("filter_offset", strconv.Itoa(offset))
		resp, err := r.Do(ctx, http.MethodGet, "index.php", query, nil)
		if err != nil {
			return fmt.Errorf("read piwik %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode piwik %s: %w", stream, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize || !endpoint.paginated {
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
	method       string
	recordsPath  string
	requiresSite bool
	paginated    bool
	mapRecord    func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"sites":   {method: "SitesManager.getAllSites", recordsPath: ".", mapRecord: siteRecord},
	"visits":  {method: "Live.getLastVisitsDetails", recordsPath: ".", requiresSite: true, paginated: true, mapRecord: visitRecord},
	"actions": {method: "Actions.getPageUrls", recordsPath: ".", requiresSite: true, paginated: true, mapRecord: actionRecord},
	"goals":   {method: "Goals.getGoals", recordsPath: ".", requiresSite: true, mapRecord: goalRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "sites", Description: "Sites configured in Piwik/Matomo.", PrimaryKey: []string{"site_id"}, Fields: []connectors.Field{{Name: "site_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "main_url", Type: "string"}}},
		{Name: "visits", Description: "Recent visit details.", PrimaryKey: []string{"visit_id"}, CursorFields: []string{"last_action_at"}, Fields: []connectors.Field{{Name: "visit_id", Type: "string"}, {Name: "visitor_id", Type: "string"}, {Name: "last_action_at", Type: "timestamp"}}},
		{Name: "actions", Description: "Page URL action metrics.", PrimaryKey: []string{"label"}, Fields: []connectors.Field{{Name: "label", Type: "string"}, {Name: "hits", Type: "number"}, {Name: "visits", Type: "number"}}},
		{Name: "goals", Description: "Configured conversion goals.", PrimaryKey: []string{"goal_id"}, Fields: []connectors.Field{{Name: "goal_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "active", Type: "boolean"}}},
	}
}

func siteRecord(item map[string]any) connectors.Record {
	return connectors.Record{"site_id": first(item, "idsite", "idSite", "site_id"), "name": item["name"], "main_url": first(item, "main_url", "mainUrl")}
}

func visitRecord(item map[string]any) connectors.Record {
	return connectors.Record{"visit_id": first(item, "idVisit", "visit_id"), "visitor_id": first(item, "visitorId", "visitor_id"), "last_action_at": first(item, "lastActionDateTime", "last_action_at")}
}

func actionRecord(item map[string]any) connectors.Record {
	return connectors.Record{"label": item["label"], "hits": first(item, "nb_hits", "hits"), "visits": first(item, "nb_visits", "visits")}
}

func goalRecord(item map[string]any) connectors.Record {
	return connectors.Record{"goal_id": first(item, "idgoal", "idGoal"), "name": item["name"], "active": item["active"]}
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"idsite": fmt.Sprintf("%d", i), "idVisit": fmt.Sprintf("v%d", i), "visitorId": fmt.Sprintf("visitor-%d", i), "lastActionDateTime": fmt.Sprintf("2026-01-0%d 00:00:00", i), "name": fmt.Sprintf("Fixture %d", i), "main_url": "https://example.com", "label": "/", "nb_hits": i, "idgoal": i, "active": true}
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
	token := strings.TrimSpace(cfg.Secrets["token_auth"])
	if token == "" {
		return nil, errors.New("piwik connector requires secret token_auth")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("token_auth", token), UserAgent: userAgent}, nil
}

func apiQuery(method string, cfg connectors.RuntimeConfig, requireSite bool) url.Values {
	query := url.Values{"module": {"API"}, "format": {"JSON"}, "method": {method}}
	if requireSite {
		query.Set("idSite", siteID(cfg))
		query.Set("period", valueOrDefault(cfg.Config["period"], "day"))
		query.Set("date", valueOrDefault(cfg.Config["date"], "today"))
	}
	return query
}

func siteID(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["site_id"]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Config["id_site"])
}

func valueOrDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return strings.TrimSpace(v)
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("piwik config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("piwik config base_url must be an absolute http or https URL")
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
		return 0, fmt.Errorf("piwik config %s must be an integer >= %d", key, min)
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
