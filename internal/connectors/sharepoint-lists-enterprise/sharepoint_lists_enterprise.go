package sharepointlistsenterprise

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
	connectorName       = "sharepoint-lists-enterprise"
	defaultBaseURL      = "https://graph.microsoft.com/v1.0"
	defaultLoginBaseURL = "https://login.microsoftonline.com"
	graphScope          = "https://graph.microsoft.com/.default"
	defaultPageSize     = 100
	defaultMaxPages     = 1
	fixtureUpdatedAt    = "2026-01-01T00:00:00Z"
	userAgent           = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "SharePoint Lists Enterprise",
		IntegrationType: "api",
		Description:     "Reads SharePoint lists and list items through Microsoft Graph.",
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
	path, err := resourcePath(cfg, "lists")
	if err != nil {
		return err
	}
	return r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil)
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "lists"
	}
	if stream != "lists" && stream != "list_items" {
		return fmt.Errorf("sharepoint-lists-enterprise stream %q not found", req.Stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, req.State, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := resourcePath(req.Config, stream)
	if err != nil {
		return err
	}
	pageSize, err := positiveInt(req.Config.Config["page_size"], defaultPageSize, 1, 999, "page_size")
	if err != nil {
		return err
	}
	maxPages, err := parseMaxPages(req.Config.Config["max_pages"])
	if err != nil {
		return err
	}
	pager := &connsdk.OffsetPaginator{LimitParam: "$top", OffsetParam: "$skip", PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, path, nil, pager, "value", maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	tenantID := strings.TrimSpace(cfg.Config["tenant_id"])
	if tenantID == "" {
		return nil, errors.New("sharepoint-lists-enterprise connector requires config tenant_id")
	}
	clientID := strings.TrimSpace(secret(cfg, "client_id"))
	clientSecret := strings.TrimSpace(secret(cfg, "client_secret"))
	if clientID == "" {
		return nil, errors.New("sharepoint-lists-enterprise connector requires secret client_id")
	}
	if clientSecret == "" {
		return nil, errors.New("sharepoint-lists-enterprise connector requires secret client_secret")
	}
	loginBase := baseURLValue(cfg.Config["login_base_url"], defaultLoginBaseURL)
	tokenURL := strings.TrimRight(loginBase, "/") + "/" + url.PathEscape(tenantID) + "/oauth2/v2.0/token"
	if override := strings.TrimSpace(cfg.Config["token_url"]); override != "" {
		tokenURL = override
	}
	return &connsdk.Requester{
		Client:  c.Client,
		BaseURL: baseURLValue(cfg.Config["base_url"], defaultBaseURL),
		Auth: &connsdk.OAuth2ClientCredentials{
			TokenURL:     tokenURL,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{graphScope},
			Client:       c.Client,
		},
		UserAgent: userAgent,
	}, nil
}

func resourcePath(cfg connectors.RuntimeConfig, stream string) (string, error) {
	siteID := strings.TrimSpace(cfg.Config["site_id"])
	if siteID == "" {
		return "", errors.New("sharepoint-lists-enterprise connector requires config site_id")
	}
	base := "sites/" + url.PathEscape(siteID) + "/lists"
	if stream == "lists" {
		return base, nil
	}
	listID := strings.TrimSpace(cfg.Config["list_id"])
	if listID == "" {
		return "", errors.New("sharepoint-lists-enterprise list_items stream requires config list_id")
	}
	return base + "/" + url.PathEscape(listID) + "/items", nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "lists", Description: "SharePoint lists in the configured site.", PrimaryKey: []string{"id"}, CursorFields: []string{"lastModifiedDateTime"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "displayName", Type: "string"}, {Name: "lastModifiedDateTime", Type: "string"}}},
		{Name: "list_items", Description: "Items in the configured SharePoint list.", PrimaryKey: []string{"id"}, CursorFields: []string{"lastModifiedDateTime"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "fields", Type: "object"}, {Name: "lastModifiedDateTime", Type: "string"}}},
	}
}

func readFixture(ctx context.Context, stream string, state map[string]string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "displayName": fmt.Sprintf("Fixture %s %d", stream, i), "lastModifiedDateTime": fixtureUpdatedAt}
		if stream == "list_items" {
			rec["fields"] = map[string]any{"Title": rec["displayName"]}
		}
		if cursor := connsdk.Cursor(state); cursor != "" {
			rec["previous_cursor"] = cursor
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
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

func baseURLValue(raw, fallback string) string {
	if v := strings.TrimSpace(raw); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fallback
}

func positiveInt(raw string, def, min, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return n, nil
}

func parseMaxPages(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultMaxPages, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("max_pages must be a non-negative integer")
	}
	return n, nil
}
