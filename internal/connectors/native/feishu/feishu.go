// Package feishu implements the native pm Feishu/Lark connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference connector: a thin package that composes connsdk.Requester
// with Feishu-specific auth (a tenant_access_token exchange), Bitable endpoints,
// stream definitions, and cursor pagination.
//
// Feishu's "Bitable" (Base) is a hosted spreadsheet/database. This connector
// reads the rows (records), the tables, and the field schema of a configured
// Base, so it is read-only (Write=false).
package feishu

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
	feishuDefaultBaseURL  = "https://open.feishu.cn"
	feishuBitablePrefix   = "open-apis/bitable/v1/apps"
	feishuTokenPath       = "open-apis/auth/v3/tenant_access_token/internal"
	feishuDefaultPageSize = 100
	feishuMaxPageSize     = 500
	feishuUserAgent       = "polymetrics-go-cli"
)

// New returns the Feishu connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Feishu/Lark connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and by the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "feishu" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "feishu",
		DisplayName:     "Feishu / Lark",
		IntegrationType: "api",
		Description:     "Reads Feishu/Lark Bitable (Base) records, tables, and field schemas via the Open Platform REST API using a tenant_access_token exchange.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Feishu. In
// fixture mode it short-circuits without a network call. Otherwise it confirms
// the credentials by performing the tenant_access_token exchange (which proves
// app_id/app_secret are valid) without reading any user data.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := feishuBaseURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	auth, err := c.authenticator(cfg)
	if err != nil {
		return err
	}
	// A token exchange validates app_id/app_secret without touching user data.
	if _, err := auth.token(ctx); err != nil {
		return fmt.Errorf("check feishu: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: feishuStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "records"
	}
	endpoint, ok := feishuStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("feishu stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	if _, err := feishuBaseURL(req.Config); err != nil {
		return err
	}
	if err := requireSecrets(req.Config); err != nil {
		return err
	}

	tableID := strings.TrimSpace(req.Config.Config["table_id"])
	if endpoint.needsTable && tableID == "" {
		return fmt.Errorf("feishu stream %q requires config table_id", stream)
	}
	appToken := strings.TrimSpace(feishuSecret(req.Config, "app_token"))
	if appToken == "" {
		return errors.New("feishu connector requires secret app_token")
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := feishuPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := feishuMaxPages(req.Config)
	if err != nil {
		return err
	}
	resource := feishuBitablePrefix + "/" + url.PathEscape(appToken) + "/" + endpoint.resourceFor(url.PathEscape(tableID))
	return c.harvest(ctx, r, resource, endpoint.mapRecord, pageSize, maxPages, emit)
}

// Write is unsupported: the Feishu connector is read-only. It satisfies the
// connectors.Connector interface and returns ErrUnsupportedOperation so the
// reverse-ETL path rejects writes cleanly.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Feishu's page_token/has_more cursor pagination. Bitable list
// endpoints return {code, msg, data:{items:[...], has_more, page_token}}. The
// next page is requested with page_token=<data.page_token>. The loop lives here
// because the body-token shape (token nested under data) plus the page-size
// param is Feishu-specific; it is built on connsdk.Requester + RecordsAt +
// StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page_size", strconv.Itoa(pageSize))
		if pageToken != "" {
			query.Set("page_token", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read feishu %s: %w", resource, err)
		}
		// Feishu signals API-level errors via a non-zero code even on HTTP 200.
		if code, _ := connsdk.StringAt(resp.Body, "code"); code != "" && code != "0" {
			msg, _ := connsdk.StringAt(resp.Body, "msg")
			return fmt.Errorf("read feishu %s: api error code %s: %s", resource, code, msg)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data.items")
		if err != nil {
			return fmt.Errorf("decode feishu %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "data.has_more")
		if err != nil {
			return fmt.Errorf("decode feishu %s has_more: %w", resource, err)
		}
		next, err := connsdk.StringAt(resp.Body, "data.page_token")
		if err != nil {
			return fmt.Errorf("decode feishu %s page_token: %w", resource, err)
		}
		if hasMore != "true" || strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise feishu credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := feishuStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"record_id":  fmt.Sprintf("rec_fixture_%d", i),
			"table_id":   fmt.Sprintf("tbl_fixture_%d", i),
			"field_id":   fmt.Sprintf("fld_fixture_%d", i),
			"field_name": fmt.Sprintf("Column %d", i),
			"name":       fmt.Sprintf("Fixture Table %d", i),
			"revision":   int64(i),
			"type":       int64(1),
			"ui_type":    "Text",
			"is_primary": i == 1,
			"is_hidden":  false,
			"property":   map[string]any{},
			"fields": map[string]any{
				"Name":   fmt.Sprintf("Fixture %d", i),
				"Status": "active",
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester for Bitable calls, wired with the
// tenant_access_token authenticator and the resolved base URL. Secrets only ever
// flow into the authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := feishuBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := c.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: feishuUserAgent,
	}, nil
}

// authenticator builds the tenant_access_token exchanger from app_id/app_secret.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (*tenantTokenAuth, error) {
	base, err := feishuBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	appID := strings.TrimSpace(feishuSecret(cfg, "app_id"))
	appSecret := strings.TrimSpace(feishuSecret(cfg, "app_secret"))
	if appID == "" || appSecret == "" {
		return nil, errors.New("feishu connector requires secrets app_id and app_secret")
	}
	return &tenantTokenAuth{
		client:    c.Client,
		baseURL:   base,
		appID:     appID,
		appSecret: appSecret,
	}, nil
}

func feishuSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// requireSecrets confirms the three credential secrets are present (without
// validating them against the remote service).
func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, key := range []string{"app_id", "app_secret", "app_token"} {
		if strings.TrimSpace(feishuSecret(cfg, key)) == "" {
			return fmt.Errorf("feishu connector requires secret %s", key)
		}
	}
	return nil
}

// feishuBaseURL resolves and validates the base URL. The default is
// open.feishu.cn; the catalog's lark_host config (or a base_url override for
// tests) may point at open.larksuite.com. Any override must be an absolute
// http(s) URL with a host to bound SSRF risk.
func feishuBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["lark_host"])
	}
	if base == "" {
		return feishuDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("feishu config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("feishu config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("feishu config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func feishuPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return feishuDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("feishu config page_size must be an integer: %w", err)
	}
	if value < 1 || value > feishuMaxPageSize {
		return 0, fmt.Errorf("feishu config page_size must be between 1 and %d", feishuMaxPageSize)
	}
	return value, nil
}

func feishuMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("feishu config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("feishu config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
