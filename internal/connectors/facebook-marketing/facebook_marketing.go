// Package facebookmarketing implements a read-only Facebook Marketing API connector.
package facebookmarketing

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
	facebookName            = "facebook-marketing"
	facebookDefaultBaseURL  = "https://graph.facebook.com/v20.0"
	facebookDefaultPageSize = 100
	facebookMaxPageSize     = 500
	facebookUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("facebook-marketing", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return facebookName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            facebookName,
		DisplayName:     "Facebook Marketing",
		IntegrationType: "api",
		Description:     "Reads Facebook Marketing ad accounts, campaigns, and ads through the Graph API. Read-only.",
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
	if strings.TrimSpace(facebookAccessToken(cfg)) == "" {
		return errors.New("facebook-marketing connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	q := url.Values{"limit": []string{"1"}, "fields": []string{"id,account_id,name"}}
	if err := r.DoJSON(ctx, http.MethodGet, "me/adaccounts", q, nil, nil); err != nil {
		return fmt.Errorf("check facebook-marketing: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: facebookStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "ad_accounts"
	}
	endpoint, ok := facebookStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("facebook-marketing stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := facebookPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := facebookMaxPages(req.Config)
	if err != nil {
		return err
	}
	resource, err := facebookResource(req.Config, endpoint)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, endpoint facebookStreamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(pageSize))
	query.Set("fields", endpoint.fields)
	path := resource
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read facebook-marketing %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode facebook-marketing %s: %w", resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return fmt.Errorf("decode facebook-marketing next page: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint facebookStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_%d", endpoint.resource, i),
			"account_id":       strconv.Itoa(1000 + i),
			"name":             fmt.Sprintf("Fixture %s %d", endpoint.resource, i),
			"status":           "ACTIVE",
			"effective_status": "ACTIVE",
			"objective":        "OUTCOME_TRAFFIC",
			"created_time":     "2026-01-01T00:00:00+0000",
			"updated_time":     "2026-01-02T00:00:00+0000",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := facebookBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := facebookAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("facebook-marketing connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: facebookUserAgent}, nil
}

type facebookStreamEndpoint struct {
	resource      string
	fields        string
	accountScoped bool
	mapRecord     func(map[string]any) connectors.Record
}

var facebookStreamEndpoints = map[string]facebookStreamEndpoint{
	"ad_accounts": {resource: "me/adaccounts", fields: "id,account_id,name,account_status,currency,timezone_name", mapRecord: facebookAdAccountRecord},
	"campaigns":   {resource: "campaigns", fields: "id,name,status,effective_status,objective,created_time,updated_time", accountScoped: true, mapRecord: facebookCampaignRecord},
	"ads":         {resource: "ads", fields: "id,name,status,effective_status,created_time,updated_time", accountScoped: true, mapRecord: facebookAdRecord},
}

func facebookStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "ad_accounts", Description: "Facebook ad accounts visible to the token.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "account_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "account_status", Type: "string"}, {Name: "currency", Type: "string"}, {Name: "timezone_name", Type: "string"}}},
		{Name: "campaigns", Description: "Campaigns for the configured ad account.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "effective_status", Type: "string"}, {Name: "objective", Type: "string"}, {Name: "created_time", Type: "string"}, {Name: "updated_time", Type: "string"}}},
		{Name: "ads", Description: "Ads for the configured ad account.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "effective_status", Type: "string"}, {Name: "created_time", Type: "string"}, {Name: "updated_time", Type: "string"}}},
	}
}

func facebookAdAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "account_id": item["account_id"], "name": item["name"], "account_status": item["account_status"], "currency": item["currency"], "timezone_name": item["timezone_name"]}
}

func facebookCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "status": item["status"], "effective_status": item["effective_status"], "objective": item["objective"], "created_time": item["created_time"], "updated_time": item["updated_time"]}
}

func facebookAdRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "status": item["status"], "effective_status": item["effective_status"], "created_time": item["created_time"], "updated_time": item["updated_time"]}
}

func facebookResource(cfg connectors.RuntimeConfig, endpoint facebookStreamEndpoint) (string, error) {
	if !endpoint.accountScoped {
		return endpoint.resource, nil
	}
	accountID := strings.TrimSpace(cfg.Config["ad_account_id"])
	if accountID == "" {
		return "", errors.New("facebook-marketing connector requires config ad_account_id for this stream")
	}
	if !strings.HasPrefix(accountID, "act_") {
		accountID = "act_" + accountID
	}
	return accountID + "/" + endpoint.resource, nil
}

func facebookAccessToken(cfg connectors.RuntimeConfig) string {
	return cfg.Secrets["access_token"]
}

func facebookBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validBaseURL(facebookName, cfg.Config["base_url"], facebookDefaultBaseURL)
}

func facebookPageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(facebookName, cfg.Config["page_size"], facebookDefaultPageSize, 1, facebookMaxPageSize)
}

func facebookMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(facebookName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func validBaseURL(name, raw, fallback string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(name, raw string, fallback, min, max int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s config value must be an integer: %w", name, err)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("%s config value must be between %d and %d", name, min, max)
	}
	return value, nil
}

func maxPagesConfig(name, raw string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", name, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be 0 for unlimited or a positive integer", name)
	}
	return value, nil
}
