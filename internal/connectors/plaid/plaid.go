// Package plaid implements a conservative read-only Plaid HTTP connector.
// It covers Plaid catalog endpoints that do not require an end-user access_token.
package plaid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL  = "https://production.plaid.com"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("plaid", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "plaid" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "plaid",
		DisplayName:     "Plaid",
		IntegrationType: "api",
		Description:     "Reads Plaid institutions and category metadata through read-only POST endpoints.",
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
	body, err := authBody(cfg)
	if err != nil {
		return err
	}
	if _, err := r.Do(ctx, http.MethodPost, "categories/get", nil, body); err != nil {
		return fmt.Errorf("check plaid: %w", err)
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
		stream = "institutions"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("plaid stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
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
	body, err := authBody(req.Config)
	if err != nil {
		return err
	}
	if stream == "categories" {
		_, err := c.readPage(ctx, r, endpoint, body, emit)
		return err
	}
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		body["count"] = pageSize
		body["offset"] = offset
		body["country_codes"] = countryCodes(req.Config)
		count, err := c.readPage(ctx, r, endpoint, body, emit)
		if err != nil {
			return err
		}
		if count < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

func (c Connector) readPage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, body map[string]any, emit func(connectors.Record) error) (int, error) {
	resp, err := r.Do(ctx, http.MethodPost, endpoint.path, nil, body)
	if err != nil {
		return 0, fmt.Errorf("read plaid %s: %w", endpoint.path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return 0, fmt.Errorf("decode plaid %s: %w", endpoint.path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"institutions": {path: "institutions/get", recordsPath: "institutions", mapRecord: institutionRecord},
	"categories":   {path: "categories/get", recordsPath: "categories", mapRecord: categoryRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "institutions", Description: "Plaid financial institutions.", PrimaryKey: []string{"institution_id"}, Fields: []connectors.Field{{Name: "institution_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "country_codes", Type: "string"}}},
		{Name: "categories", Description: "Plaid transaction category metadata.", PrimaryKey: []string{"category_id"}, Fields: []connectors.Field{{Name: "category_id", Type: "string"}, {Name: "group", Type: "string"}, {Name: "hierarchy", Type: "string"}}},
	}
}

func institutionRecord(item map[string]any) connectors.Record {
	return connectors.Record{"institution_id": item["institution_id"], "name": item["name"], "country_codes": joinAny(item["country_codes"])}
}

func categoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{"category_id": item["category_id"], "group": item["group"], "hierarchy": joinAny(item["hierarchy"])}
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"institution_id": fmt.Sprintf("ins_fixture_%d", i), "name": fmt.Sprintf("Fixture Bank %d", i), "country_codes": []any{"US"}, "category_id": fmt.Sprintf("%d000", i), "group": "transfer", "hierarchy": []any{"Transfer", "Debit"}}
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
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func authBody(cfg connectors.RuntimeConfig) (map[string]any, error) {
	clientID := strings.TrimSpace(cfg.Secrets["client_id"])
	secret := strings.TrimSpace(cfg.Secrets["secret"])
	if clientID == "" || secret == "" {
		return nil, errors.New("plaid connector requires secrets client_id and secret")
	}
	return map[string]any{"client_id": clientID, "secret": secret}, nil
}

func countryCodes(cfg connectors.RuntimeConfig) []string {
	values := splitCSV(cfg.Config["country_codes"])
	if len(values) == 0 {
		return []string{"US"}
	}
	return values
}

func splitCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func joinAny(v any) string {
	list, ok := v.([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(list))
	for _, item := range list {
		parts = append(parts, fmt.Sprintf("%v", item))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("plaid config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("plaid config base_url must be an absolute http or https URL")
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
		return 0, fmt.Errorf("plaid config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
