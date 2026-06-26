// Package spotifyads implements the native pm Spotify Ads connector.
package spotifyads

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
	connectorName   = "spotify-ads"
	defaultBaseURL  = "https://api-partner.spotify.com/ads/v2"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	resource       string
	recordsPath    string
	needsAdAccount bool
	fields         []connectors.Field
	mapRecord      func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Spotify Ads", IntegrationType: "api", Description: "Reads Spotify Ads ad accounts, campaigns, ad sets, and ads through the Spotify Ads API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "ad_accounts", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check spotify-ads: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "ad_accounts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("spotify-ads stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
	}
	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
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
	paginator := &connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "access_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("spotify-ads connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("spotify-ads config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("spotify-ads config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("spotify-ads config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.needsAdAccount {
		return endpoint.resource, nil
	}
	accountID := strings.TrimSpace(cfg.Config["ad_account_id"])
	if accountID == "" {
		return "", fmt.Errorf("spotify-ads stream requires config ad_account_id for path %q", endpoint.resource)
	}
	return strings.ReplaceAll(endpoint.resource, "{ad_account_id}", url.PathEscape(accountID)), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "spotify-ads config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("spotify-ads config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}

func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
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

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "status": "ACTIVE", "country": "US", "objective": "AWARENESS"}
		rec := endpoint.mapRecord(item)
		rec["connector"] = connectorName
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "ad_accounts", Description: "Spotify Ads ad accounts.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["ad_accounts"].fields},
		{Name: "campaigns", Description: "Spotify Ads campaigns for an ad_account_id.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["campaigns"].fields},
		{Name: "ad_sets", Description: "Spotify Ads ad sets for an ad_account_id.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["ad_sets"].fields},
		{Name: "ads", Description: "Spotify Ads ads for an ad_account_id.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["ads"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"ad_accounts": {resource: "ad_accounts", recordsPath: "ad_accounts", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "country", Type: "string"}}, mapRecord: copyRecord("id", "name", "country")},
	"campaigns":   {resource: "ad_accounts/{ad_account_id}/campaigns", recordsPath: "campaigns", needsAdAccount: true, fields: campaignFields(), mapRecord: copyRecord("id", "name", "status", "objective")},
	"ad_sets":     {resource: "ad_accounts/{ad_account_id}/ad_sets", recordsPath: "ad_sets", needsAdAccount: true, fields: campaignFields(), mapRecord: copyRecord("id", "name", "status", "objective")},
	"ads":         {resource: "ad_accounts/{ad_account_id}/ads", recordsPath: "ads", needsAdAccount: true, fields: campaignFields(), mapRecord: copyRecord("id", "name", "status", "objective")},
}

func campaignFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "objective", Type: "string"}}
}

func copyRecord(keys ...string) func(map[string]any) connectors.Record {
	return func(item map[string]any) connectors.Record {
		rec := connectors.Record{}
		for _, key := range keys {
			rec[key] = item[key]
		}
		return rec
	}
}
