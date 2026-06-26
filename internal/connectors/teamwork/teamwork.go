package teamwork

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
	connectorName  = "teamwork"
	defaultBaseURL = "https://api.teamwork.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Teamwork", IntegrationType: "api", Description: "Reads Teamwork projects through the Teamwork API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Config["username"]) == "" || strings.TrimSpace(cfg.Secrets["password"]) == "" {
		return errors.New("teamwork connector requires config username and secret password")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "projects", Description: "Teamwork projects.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	if stream != "projects" {
		return fmt.Errorf("teamwork stream %q not found", stream)
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
		q.Set("page", strconv.Itoa(page))
		q.Set("pageSize", strconv.Itoa(size))
		resp, err := r.Do(ctx, http.MethodGet, "projects.json", q, nil)
		if err != nil {
			return fmt.Errorf("read teamwork projects: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "projects")
		if err != nil {
			return err
		}
		for _, item := range records {
			if err := emit(connectors.Record{"id": item["id"], "name": item["name"], "created_at": item["created-on"]}); err != nil {
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
	username := strings.TrimSpace(cfg.Config["username"])
	password := strings.TrimSpace(cfg.Secrets["password"])
	if username == "" || password == "" {
		return nil, errors.New("teamwork connector requires config username and secret password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(username, password), UserAgent: userAgent}, nil
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": "project_fixture_1", "name": "Fixture Project", "created_at": "2026-01-01T00:00:00Z", "fixture": true})
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
		return "", fmt.Errorf("teamwork config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("teamwork config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
