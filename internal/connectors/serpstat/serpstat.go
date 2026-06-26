package serpstat

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
	connectorName    = "serpstat"
	defaultBaseURL   = "https://api.serpstat.com/v4"
	defaultPageSize  = 10
	defaultMaxPages  = 1
	defaultDomain    = "serpstat.com"
	fixtureUpdatedAt = "2026-01-01T00:00:00Z"
	userAgent        = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Serpstat",
		IntegrationType: "api",
		Description:     "Reads Serpstat SEO domain keyword and competitor data through the Serpstat API.",
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
	req := connectors.ReadRequest{Stream: "domain_keywords", Config: cfg}
	return c.Read(ctx, req, func(connectors.Record) error { return nil })
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "domain_keywords"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("serpstat stream %q not found", req.Stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, req.State, emit)
	}
	r, token, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := positiveInt(req.Config.Config["page_size"], defaultPageSize, 1, 1000, "page_size")
	if err != nil {
		return err
	}
	pages, err := parsePages(req.Config.Config["pages_to_fetch"])
	if err != nil {
		return err
	}
	domain := strings.TrimSpace(req.Config.Config["domain"])
	if domain == "" {
		domain = defaultDomain
	}
	region := strings.TrimSpace(req.Config.Config["region_id"])
	if region == "" {
		region = "g_us"
	}
	query := url.Values{"token": []string{token}}
	for page := 1; pages == 0 || page <= pages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := map[string]any{
			"id":     page,
			"method": ep.method,
			"params": map[string]any{"domain": domain, "se": region, "page": page, "size": pageSize},
		}
		resp, err := r.Do(ctx, http.MethodPost, strings.TrimRight(r.BaseURL, "/"), query, body)
		if err != nil {
			return err
		}
		records, err := connsdk.RecordsAt(resp.Body, "result.data")
		if err != nil {
			return err
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, string, error) {
	key := strings.TrimSpace(secret(cfg, "api_key"))
	if key == "" {
		return nil, "", errors.New("serpstat connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: baseURL(cfg, defaultBaseURL), UserAgent: userAgent}, key, nil
}

type streamEndpoint struct{ method string }

var streamEndpoints = map[string]streamEndpoint{
	"domain_keywords":    {method: "SerpstatDomainProcedure.getKeywords"},
	"domain_competitors": {method: "SerpstatDomainProcedure.getCompetitors"},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "keyword", Type: "string"}, {Name: "position", Type: "integer"}, {Name: "url", Type: "string"}, {Name: "updated_at", Type: "string"}}
	return []connectors.Stream{
		{Name: "domain_keywords", Description: "Serpstat domain keywords.", PrimaryKey: []string{"keyword", "url"}, Fields: fields},
		{Name: "domain_competitors", Description: "Serpstat domain competitors.", PrimaryKey: []string{"domain"}, Fields: []connectors.Field{{Name: "domain", Type: "string"}, {Name: "visibility", Type: "number"}}},
	}
}

func readFixture(ctx context.Context, stream string, state map[string]string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"keyword": fmt.Sprintf("fixture-%s-%d", stream, i), "position": i, "url": "https://example.com", "updated_at": fixtureUpdatedAt}
		if cursor := connsdk.Cursor(state); cursor != "" {
			rec["previous_cursor"] = cursor
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
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

func baseURL(cfg connectors.RuntimeConfig, fallback string) string {
	if v := strings.TrimSpace(cfg.Config["base_url"]); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fallback
}

func positiveInt(raw string, def, min, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return n, nil
}

func parsePages(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultMaxPages, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("pages_to_fetch must be a non-negative integer")
	}
	return n, nil
}
