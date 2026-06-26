package teamtailor

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
	connectorName  = "teamtailor"
	defaultBaseURL = "https://api.teamtailor.com/v1"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Teamtailor", IntegrationType: "api", Description: "Reads Teamtailor jobs through the Teamtailor API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(apiKey(cfg)) == "" {
		return errors.New("teamtailor connector requires config api or secret api_key")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "jobs", Description: "Teamtailor jobs.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "created_at", Type: "timestamp"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "jobs"
	}
	if stream != "jobs" {
		return fmt.Errorf("teamtailor stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size := pageSize(req.Config, 100)
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page[number]", strconv.Itoa(page))
		q.Set("page[size]", strconv.Itoa(size))
		resp, err := r.Do(ctx, http.MethodGet, "jobs", q, nil)
		if err != nil {
			return fmt.Errorf("read teamtailor jobs: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return err
		}
		for _, item := range records {
			attrs := object(item["attributes"])
			if err := emit(connectors.Record{"id": item["id"], "title": attrs["title"], "created_at": attrs["created-at"]}); err != nil {
				return err
			}
		}
		if len(records) < size {
			return nil
		}
	}
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := strings.TrimSpace(apiKey(cfg))
	if key == "" {
		return nil, errors.New("teamtailor connector requires config api or secret api_key")
	}
	headers := map[string]string{}
	if version := strings.TrimSpace(cfg.Config["x_api_version"]); version != "" {
		headers["X-Api-Version"] = version
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", key, "Token token="), DefaultHeaders: headers, UserAgent: userAgent}, nil
}

func apiKey(cfg connectors.RuntimeConfig) string {
	if strings.TrimSpace(cfg.Config["api"]) != "" {
		return cfg.Config["api"]
	}
	return cfg.Secrets["api_key"]
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	row := connectors.Record{"id": "job_fixture_1", "title": "Fixture Job", "created_at": "2026-01-01T00:00:00Z", "fixture": true}
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(row)
}

func object(v any) map[string]any {
	if out, ok := v.(map[string]any); ok {
		return out
	}
	return map[string]any{}
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func pageSize(cfg connectors.RuntimeConfig, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(cfg.Config["page_size"]))
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("teamtailor config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("teamtailor config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
