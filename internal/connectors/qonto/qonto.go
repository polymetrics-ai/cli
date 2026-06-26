// Package qonto implements a read-only native connector for Qonto's REST API.
package qonto

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
	defaultBaseURL  = "https://thirdparty.qonto.com/v2"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("qonto", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

type streamSpec struct {
	path         string
	recordsPath  string
	requiresIBAN bool
}

var streamSpecs = map[string]streamSpec{
	"transactions": {path: "transactions", recordsPath: "transactions", requiresIBAN: true},
	"memberships":  {path: "memberships", recordsPath: "memberships"},
	"accounts":     {path: "accounts", recordsPath: "accounts"},
}

func (Connector) Name() string { return "qonto" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "qonto",
		DisplayName:     "Qonto",
		IntegrationType: "api",
		Description:     "Reads Qonto transactions, memberships, and accounts through the Qonto API. Read-only.",
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
	if _, err := iban(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "transactions", url.Values{"page": []string{"1"}, "per_page": []string{"1"}, "iban": []string{cfg.Config["iban"]}}, nil, nil); err != nil {
		return fmt.Errorf("check qonto: %w", err)
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
		stream = "transactions"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("qonto stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
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
	query := url.Values{"page": []string{"1"}, "per_page": []string{strconv.Itoa(pageSize)}}
	if spec.requiresIBAN {
		iban, err := iban(req.Config)
		if err != nil {
			return err
		}
		query.Set("iban", iban)
	}
	if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
		query.Set("start_date", start)
	}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, spec.path, query, nil)
		if err != nil {
			return fmt.Errorf("read qonto %s: %w", spec.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode qonto %s: %w", spec.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		nextPage, err := connsdk.StringAt(resp.Body, "meta.next_page")
		if err != nil {
			return fmt.Errorf("decode qonto next_page: %w", err)
		}
		if strings.TrimSpace(nextPage) == "" {
			return nil
		}
		query.Set("page", nextPage)
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "api_key")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("qonto connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", token, ""), UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "amount", Type: "string"}, {Name: "side", Type: "string"}, {Name: "settled_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "transactions", Description: "Qonto bank transactions for a configured IBAN.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"settled_at"}},
		{Name: "memberships", Description: "Qonto memberships.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "accounts", Description: "Qonto accounts.", Fields: fields, PrimaryKey: []string{"id"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         first(item, "transaction_id", "id", "slug"),
		"amount":     item["amount"],
		"side":       item["side"],
		"label":      first(item, "label", "name"),
		"settled_at": item["settled_at"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"raw":        item,
	}
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "amount": fmt.Sprintf("%d.00", i), "side": "credit", "settled_at": fmt.Sprintf("2026-01-0%d", i)}); err != nil {
			return err
		}
	}
	return nil
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func iban(cfg connectors.RuntimeConfig) (string, error) {
	value := strings.TrimSpace(cfg.Config["iban"])
	if value == "" {
		return "", errors.New("qonto connector requires config iban for transactions")
	}
	if strings.ContainsAny(value, "/?#") {
		return "", fmt.Errorf("qonto config iban %q is invalid", cfg.Config["iban"])
	}
	return value, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("qonto config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("qonto config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("qonto config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 500 {
		return 0, errors.New("qonto config page_size must be between 1 and 500")
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
		return 0, errors.New("qonto config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
