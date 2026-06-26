// Package googleads implements a conservative read-only Google Ads REST connector.
// Google Ads reads are deliberately allow-listed to customer discovery and a
// small set of GAQL search resources; arbitrary GAQL is not accepted here.
package googleads

import (
	"bytes"
	"context"
	"encoding/json"
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
	googleAdsName            = "google-ads"
	googleAdsDefaultBaseURL  = "https://googleads.googleapis.com/v24"
	googleAdsDefaultPageSize = 1000
	googleAdsMaxPageSize     = 10000
	googleAdsUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("google-ads", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return googleAdsName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            googleAdsName,
		DisplayName:     "Google Ads",
		IntegrationType: "api",
		Description:     "Reads accessible customers and allow-listed Google Ads resources through the Google Ads REST API. Read-only.",
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
	if err := googleAdsValidateSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "customers:listAccessibleCustomers", nil, nil, nil); err != nil {
		return fmt.Errorf("check google-ads: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: googleAdsStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accessible_customers"
	}
	endpoint, ok := googleAdsStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("google-ads stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if stream == "accessible_customers" {
		return c.readAccessibleCustomers(ctx, r, emit)
	}
	cid := strings.TrimSpace(req.Config.Config["customer_id"])
	if cid == "" {
		return errors.New("google-ads connector requires config customer_id for GAQL streams")
	}
	pageSize, err := googleAdsPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := googleAdsMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.search(ctx, r, cid, endpoint, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) readAccessibleCustomers(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "customers:listAccessibleCustomers", nil, nil)
	if err != nil {
		return fmt.Errorf("read google-ads accessible customers: %w", err)
	}
	var out struct {
		ResourceNames []string `json:"resourceNames"`
	}
	if err := json.NewDecoder(bytes.NewReader(resp.Body)).Decode(&out); err != nil {
		return fmt.Errorf("decode google-ads accessible customers: %w", err)
	}
	for _, rn := range out.ResourceNames {
		parts := strings.Split(strings.TrimSpace(rn), "/")
		customerID := rn
		if len(parts) > 0 {
			customerID = parts[len(parts)-1]
		}
		if err := emit(connectors.Record{"resource_name": rn, "customer_id": customerID}); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) search(ctx context.Context, r *connsdk.Requester, customerID string, endpoint googleAdsStreamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "customers/" + customerID + "/googleAds:search"
	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		body := googleAdsSearchRequest{Query: endpoint.query, PageSize: pageSize, PageToken: pageToken}
		resp, err := r.Do(ctx, http.MethodPost, path, nil, body)
		if err != nil {
			return fmt.Errorf("read google-ads %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode google-ads %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode google-ads nextPageToken: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint googleAdsStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"resourceName": fmt.Sprintf("customers/123/%s/%d", endpoint.resource, i),
			"id":           strconv.Itoa(100 + i),
			"name":         fmt.Sprintf("Fixture %s %d", endpoint.resource, i),
			"status":       "ENABLED",
		}
		if endpoint.resource == "customer" {
			item["customer_id"] = strconv.Itoa(1000000000 + i)
			item["resource_name"] = fmt.Sprintf("customers/%d", 1000000000+i)
		}
		if err := emit(endpoint.fixtureRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	if err := googleAdsValidateSecrets(cfg); err != nil {
		return nil, err
	}
	base, err := googleAdsBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      googleAdsAuth{accessToken: cfg.Secrets["access_token"], developerToken: cfg.Secrets["developer_token"], loginCustomerID: cfg.Config["login_customer_id"]},
		UserAgent: googleAdsUserAgent,
	}, nil
}

type googleAdsAuth struct {
	accessToken     string
	developerToken  string
	loginCustomerID string
}

func (a googleAdsAuth) Apply(_ context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(a.accessToken))
	req.Header.Set("developer-token", strings.TrimSpace(a.developerToken))
	if strings.TrimSpace(a.loginCustomerID) != "" {
		req.Header.Set("login-customer-id", strings.TrimSpace(a.loginCustomerID))
	}
	return nil
}

type googleAdsSearchRequest struct {
	Query     string `json:"query"`
	PageSize  int    `json:"pageSize,omitempty"`
	PageToken string `json:"pageToken,omitempty"`
}

type googleAdsStreamEndpoint struct {
	resource      string
	query         string
	mapRecord     func(map[string]any) connectors.Record
	fixtureRecord func(map[string]any) connectors.Record
}

var googleAdsStreamEndpoints = map[string]googleAdsStreamEndpoint{
	"accessible_customers": {resource: "customer", fixtureRecord: googleAdsAccessibleCustomerRecord},
	"campaigns":            {resource: "campaign", query: "SELECT campaign.id, campaign.name, campaign.status, campaign.resource_name FROM campaign", mapRecord: googleAdsCampaignResultRecord, fixtureRecord: googleAdsGenericRecord},
	"ad_groups":            {resource: "ad_group", query: "SELECT ad_group.id, ad_group.name, ad_group.status, ad_group.resource_name FROM ad_group", mapRecord: googleAdsAdGroupResultRecord, fixtureRecord: googleAdsGenericRecord},
}

func googleAdsStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "accessible_customers", Description: "Google Ads customers accessible to the OAuth token.", PrimaryKey: []string{"customer_id"}, Fields: []connectors.Field{{Name: "customer_id", Type: "string"}, {Name: "resource_name", Type: "string"}}},
		{Name: "campaigns", Description: "Campaigns for the configured customer_id.", PrimaryKey: []string{"id"}, Fields: googleAdsStandardFields()},
		{Name: "ad_groups", Description: "Ad groups for the configured customer_id.", PrimaryKey: []string{"id"}, Fields: googleAdsStandardFields()},
	}
}

func googleAdsStandardFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "resource_name", Type: "string"}}
}

func googleAdsCampaignResultRecord(item map[string]any) connectors.Record {
	return googleAdsNestedRecord(item, "campaign")
}

func googleAdsAdGroupResultRecord(item map[string]any) connectors.Record {
	return googleAdsNestedRecord(item, "adGroup")
}

func googleAdsNestedRecord(item map[string]any, key string) connectors.Record {
	nested, _ := item[key].(map[string]any)
	if nested == nil {
		nested, _ = item[strings.ToLower(key)].(map[string]any)
	}
	return connectors.Record{"id": nested["id"], "name": nested["name"], "status": nested["status"], "resource_name": first(nested["resourceName"], nested["resource_name"])}
}

func googleAdsGenericRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "status": item["status"], "resource_name": first(item["resourceName"], item["resource_name"])}
}

func googleAdsAccessibleCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"customer_id": item["customer_id"], "resource_name": item["resource_name"]}
}

func googleAdsValidateSecrets(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(cfg.Secrets["access_token"]) == "" || strings.TrimSpace(cfg.Secrets["developer_token"]) == "" {
		return errors.New("google-ads connector requires secrets access_token and developer_token")
	}
	return nil
}

func googleAdsBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validBaseURL(googleAdsName, cfg.Config["base_url"], googleAdsDefaultBaseURL)
}

func googleAdsPageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(googleAdsName, cfg.Config["page_size"], googleAdsDefaultPageSize, 1, googleAdsMaxPageSize)
}

func googleAdsMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(googleAdsName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func first(values ...any) any {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		if value != nil {
			return value
		}
	}
	return nil
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
