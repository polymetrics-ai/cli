package convex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("convex", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "convex" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "convex", DisplayName: "Convex", IntegrationType: "api", Description: "Reads Convex tables and documents through the deployment HTTP API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "api/tables", nil, nil, nil); err != nil {
		return fmt.Errorf("check convex: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{Name: "tables", Description: "Convex table metadata.", Fields: []connectors.Field{{Name: "name", Type: "string"}}, PrimaryKey: []string{"name"}},
		{Name: "documents", Description: "Convex table documents.", Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "_id", Type: "string"}}, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "documents"
	}
	if stream != "documents" && stream != "tables" {
		return fmt.Errorf("convex stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if stream == "tables" {
		return readTables(ctx, r, emit)
	}
	return readDocuments(ctx, r, tableName(req.Config), defaultMaxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readTables(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "api/tables", nil, nil)
	if err != nil {
		return fmt.Errorf("read convex tables: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "tables")
	if err != nil {
		return fmt.Errorf("decode convex tables: %w", err)
	}
	for _, rec := range records {
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func readDocuments(ctx context.Context, r *connsdk.Requester, table string, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for i := 0; i < maxPages; i++ {
		query := url.Values{}
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, "api/tables/"+url.PathEscape(table)+"/documents", query, nil)
		if err != nil {
			return fmt.Errorf("read convex documents: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "documents")
		if err != nil {
			return fmt.Errorf("decode convex documents: %w", err)
		}
		for _, rec := range records {
			out := connectors.Record(rec)
			if out["id"] == nil && out["_id"] != nil {
				out["id"] = out["_id"]
			}
			if err := emit(out); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "cursor")
		if err != nil || strings.TrimSpace(next) == "" {
			return err
		}
		cursor = next
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"fixture": true}
		if stream == "tables" {
			rec["name"] = fmt.Sprintf("fixture_table_%d", i)
		} else {
			rec["id"] = fmt.Sprintf("doc_fixture_%d", i)
			rec["_id"] = rec["id"]
			rec["text"] = fmt.Sprintf("fixture document %d", i)
		}
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
	key := secret(cfg, "access_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("convex connector requires secret access_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["deployment_url"])
	}
	if base == "" {
		return "", errors.New("convex connector requires config deployment_url or base_url")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("convex config deployment_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("convex config deployment_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("convex config deployment_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func tableName(cfg connectors.RuntimeConfig) string {
	if table := strings.TrimSpace(cfg.Config["table"]); table != "" {
		return table
	}
	return "data"
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
