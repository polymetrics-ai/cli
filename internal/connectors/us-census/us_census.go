package uscensus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://api.census.gov"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("us-census", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "us-census" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "us-census",
		DisplayName:     "US Census",
		IntegrationType: "api",
		Description:     "Reads configured datasets from the US Census API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	_, err := c.requester(cfg)
	return err
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{queryStream()}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "query"
	}
	if stream != "query" {
		return fmt.Errorf("us-census stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path := strings.TrimSpace(req.Config.Config["query_path"])
	if path == "" {
		return errors.New("us-census connector requires config query_path")
	}
	query, err := url.ParseQuery(strings.TrimSpace(req.Config.Config["query_params"]))
	if err != nil {
		return fmt.Errorf("us-census config query_params is invalid: %w", err)
	}
	resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return fmt.Errorf("read us-census query: %w", err)
	}
	records, err := censusRows(resp.Body)
	if err != nil {
		return err
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("key", secret(cfg, "api_key")), UserAgent: userAgent}, nil
}

func queryStream() connectors.Stream {
	return connectors.Stream{
		Name:        "query",
		Description: "Rows returned by the configured Census query.",
		Fields: []connectors.Field{
			{Name: "name", Type: "string"},
			{Name: "estab", Type: "string"},
		},
	}
}

func censusRows(body []byte) ([]connectors.Record, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var rows [][]any
	if err := dec.Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode us-census query: %w", err)
	}
	if len(rows) < 2 {
		return nil, nil
	}
	headers := make([]string, len(rows[0]))
	for i, header := range rows[0] {
		headers[i] = strings.ToLower(stringValue(header))
	}
	out := make([]connectors.Record, 0, len(rows)-1)
	for _, row := range rows[1:] {
		rec := connectors.Record{}
		for i, value := range row {
			if i >= len(headers) || headers[i] == "" {
				continue
			}
			rec[headers[i]] = stringValue(value)
		}
		out = append(out, rec)
	}
	return out, nil
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"name": "United States", "estab": "1"})
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("us-census config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("us-census config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("us-census config base_url requires a host")
	}
	return strings.TrimRight(base, "/"), nil
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

func stringValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}
