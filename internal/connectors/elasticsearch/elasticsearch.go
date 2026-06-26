package elasticsearch

import (
	"context"
	"encoding/base64"
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
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("elasticsearch", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "elasticsearch" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "elasticsearch", DisplayName: "Elasticsearch", IntegrationType: "api", Description: "Reads Elasticsearch index metadata and documents through the REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "", nil, nil, nil); err != nil {
		return fmt.Errorf("check elasticsearch: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{Name: "indices", Description: "Elasticsearch index metadata.", Fields: []connectors.Field{{Name: "index", Type: "string"}, {Name: "docs.count", Type: "integer"}}, PrimaryKey: []string{"index"}},
		{Name: "documents", Description: "Elasticsearch index documents.", Fields: []connectors.Field{{Name: "id", Type: "string"}}, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "documents"
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	switch stream {
	case "indices":
		return readIndices(ctx, r, emit)
	case "documents":
		return readDocuments(ctx, r, req.Config, emit)
	default:
		return fmt.Errorf("elasticsearch stream %q not found", stream)
	}
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readIndices(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "_cat/indices", url.Values{"format": []string{"json"}}, nil)
	if err != nil {
		return fmt.Errorf("read elasticsearch indices: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode elasticsearch indices: %w", err)
	}
	for _, rec := range records {
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func readDocuments(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	index := strings.TrimSpace(cfg.Config["index"])
	if index == "" {
		return errors.New("elasticsearch documents stream requires config index")
	}
	pageSize, err := intConfig(cfg, "page_size", defaultPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(cfg, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	from := 0
	for page := 0; page < maxPages; page++ {
		query := url.Values{"from": []string{strconv.Itoa(from)}, "size": []string{strconv.Itoa(pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, url.PathEscape(index)+"/_search", query, nil)
		if err != nil {
			return fmt.Errorf("read elasticsearch documents: %w", err)
		}
		hits, err := connsdk.RecordsAt(resp.Body, "hits.hits")
		if err != nil {
			return fmt.Errorf("decode elasticsearch documents: %w", err)
		}
		for _, hit := range hits {
			if err := emit(mapHit(hit)); err != nil {
				return err
			}
		}
		if len(hits) < pageSize {
			return nil
		}
		from += pageSize
	}
	return nil
}

func mapHit(hit map[string]any) connectors.Record {
	out := connectors.Record{}
	if src, ok := hit["_source"].(map[string]any); ok {
		for k, v := range src {
			out[k] = v
		}
	}
	if id, ok := hit["_id"].(string); ok {
		out["id"] = id
	}
	return out
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	if stream != "indices" && stream != "documents" {
		return fmt.Errorf("elasticsearch fixture stream %q not found", stream)
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"fixture": true}
		if stream == "indices" {
			rec["index"] = fmt.Sprintf("fixture_%d", i)
			rec["docs.count"] = i
		} else {
			rec["id"] = fmt.Sprintf("doc_fixture_%d", i)
			rec["order_number"] = fmt.Sprintf("F-%04d", i)
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
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth(cfg), UserAgent: userAgent}, nil
}

func auth(cfg connectors.RuntimeConfig) connsdk.Authenticator {
	if apiKeyID := strings.TrimSpace(cfg.Config["apiKeyId"]); apiKeyID != "" {
		secret := strings.TrimSpace(secret(cfg, "apiKeySecret"))
		if secret != "" {
			encoded := base64.StdEncoding.EncodeToString([]byte(apiKeyID + ":" + secret))
			return connsdk.APIKeyHeader("Authorization", encoded, "ApiKey ")
		}
	}
	if username := strings.TrimSpace(cfg.Config["username"]); username != "" {
		return connsdk.Basic(username, secret(cfg, "password"))
	}
	return nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["endpoint"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["base_url"])
	}
	if base == "" {
		return "", errors.New("elasticsearch connector requires config endpoint")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("elasticsearch config endpoint is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("elasticsearch config endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("elasticsearch config endpoint must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("elasticsearch config %s must be a positive integer", key)
	}
	return value, nil
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
