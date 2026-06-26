// Package netsuite implements a conservative read-only NetSuite REST Record API
// connector. NetSuite's full object model and SuiteQL surface are broad; this
// native port intentionally covers a small allow-listed set of REST record list
// endpoints and maps only fields returned by those endpoints.
package netsuite

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	netsuiteDefaultPageSize = 100
	netsuiteMaxPageSize     = 1000
	netsuiteUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("netsuite", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "netsuite" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "netsuite",
		DisplayName:     "NetSuite",
		IntegrationType: "api",
		Description:     "Reads selected NetSuite REST Record API resources such as customers, vendors, items, and sales orders.",
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
	return r.DoJSON(ctx, http.MethodGet, "customer", url.Values{"limit": []string{"1"}}, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "netsuite", Streams: streams()}, nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("netsuite stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, ep, req, emit)
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
	return harvest(ctx, r, ep, pageSize, maxPages, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if offset > 0 {
			query.Set("offset", strconv.Itoa(offset))
		}
		resp, err := r.Do(ctx, http.MethodGet, ep.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read netsuite %s: %w", ep.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode netsuite %s page: %w", ep.resource, err)
		}
		for _, item := range records {
			if err := emit(ep.mapRecord(item)); err != nil {
				return err
			}
		}
		if !hasMore(resp.Body) || len(records) == 0 {
			return nil
		}
		offset += len(records)
	}
	return nil
}

func readFixture(ctx context.Context, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", ep.resource, i),
			"entityId":         fmt.Sprintf("Fixture %d", i),
			"name":             fmt.Sprintf("Fixture %d", i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"status":           "active",
			"lastModifiedDate": "2026-01-01T00:00:00Z",
		}
		rec := ep.mapRecord(item)
		if cursor := req.State[connsdk.CursorStateKey]; cursor != "" {
			rec["previous_cursor"] = cursor
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
	auth, err := oauth1Auth(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: netsuiteUserAgent}, nil
}

type oauth1 struct {
	realm, consumerKey, consumerSecret, tokenKey, tokenSecret string
}

func (a oauth1) Apply(_ context.Context, req *http.Request) error {
	params := map[string]string{
		"oauth_consumer_key":     a.consumerKey,
		"oauth_token":            a.tokenKey,
		"oauth_signature_method": "HMAC-SHA256",
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_nonce":            strconv.FormatInt(time.Now().UnixNano(), 36),
		"oauth_version":          "1.0",
	}
	for key, values := range req.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	params["oauth_signature"] = oauthSignature(req.Method, req.URL, params, a.consumerSecret, a.tokenSecret)
	headerParams := []string{fmt.Sprintf("realm=%q", a.realm)}
	for _, key := range []string{"oauth_consumer_key", "oauth_token", "oauth_signature_method", "oauth_timestamp", "oauth_nonce", "oauth_version", "oauth_signature"} {
		headerParams = append(headerParams, fmt.Sprintf("%s=%q", key, percent(params[key])))
	}
	req.Header.Set("Authorization", "OAuth "+strings.Join(headerParams, ","))
	return nil
}

func oauthSignature(method string, u *url.URL, params map[string]string, consumerSecret, tokenSecret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key != "oauth_signature" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, percent(key)+"="+percent(params[key]))
	}
	baseURL := *u
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	base := strings.ToUpper(method) + "&" + percent(baseURL.String()) + "&" + percent(strings.Join(parts, "&"))
	mac := hmac.New(sha256.New, []byte(percent(consumerSecret)+"&"+percent(tokenSecret)))
	_, _ = mac.Write([]byte(base))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func percent(s string) string { return strings.ReplaceAll(url.QueryEscape(s), "+", "%20") }

func oauth1Auth(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	realm := configOrSecret(cfg, "realm")
	ck := configOrSecret(cfg, "consumer_key")
	cs := configOrSecret(cfg, "consumer_secret")
	tk := configOrSecret(cfg, "token_key")
	ts := configOrSecret(cfg, "token_secret")
	if realm == "" || ck == "" || cs == "" || tk == "" || ts == "" {
		return nil, errors.New("netsuite connector requires realm, consumer_key, consumer_secret, token_key, and token_secret")
	}
	return oauth1{realm: realm, consumerKey: ck, consumerSecret: cs, tokenKey: tk, tokenSecret: ts}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		return validateBaseURL("netsuite", base)
	}
	realm := configOrSecret(cfg, "realm")
	if realm == "" {
		return "", errors.New("netsuite connector requires config base_url or realm")
	}
	host := strings.ToLower(strings.ReplaceAll(realm, "_", "-"))
	return "https://" + host + ".suitetalk.api.netsuite.com/services/rest/record/v1", nil
}

func configOrSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Config != nil {
		if value := strings.TrimSpace(cfg.Config[key]); value != "" {
			return value
		}
	}
	if cfg.Secrets != nil {
		return strings.TrimSpace(cfg.Secrets[key])
	}
	return ""
}

func validateBaseURL(name, raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(raw, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return netsuiteDefaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > netsuiteMaxPageSize {
		return 0, fmt.Errorf("netsuite config page_size must be between 1 and %d", netsuiteMaxPageSize)
	}
	return n, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("netsuite config max_pages must be a non-negative integer")
	}
	return n, nil
}

func hasMore(body []byte) bool {
	var env struct {
		HasMore bool `json:"hasMore"`
	}
	_ = json.Unmarshal(body, &env)
	return env.HasMore
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"customers":    {resource: "customer", mapRecord: record},
	"vendors":      {resource: "vendor", mapRecord: record},
	"items":        {resource: "inventoryItem", mapRecord: record},
	"sales_orders": {resource: "salesOrder", mapRecord: record},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "entity_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "status", Type: "string"}, {Name: "last_modified_date", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "customers", Description: "NetSuite customer records.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_modified_date"}, Fields: fields},
		{Name: "vendors", Description: "NetSuite vendor records.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_modified_date"}, Fields: fields},
		{Name: "items", Description: "NetSuite inventory item records.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_modified_date"}, Fields: fields},
		{Name: "sales_orders", Description: "NetSuite sales order records.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_modified_date"}, Fields: fields},
	}
}

func record(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"entity_id":          first(item, "entityId", "tranId"),
		"name":               first(item, "companyName", "name", "title"),
		"email":              item["email"],
		"status":             first(item, "status", "entityStatus"),
		"last_modified_date": item["lastModifiedDate"],
	}
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
