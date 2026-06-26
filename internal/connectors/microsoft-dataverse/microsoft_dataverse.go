// Package microsoftdataverse implements a read-only Microsoft Dataverse
// connector using OAuth2 client credentials and OData @odata.nextLink pagination.
package microsoftdataverse

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
	connectorName       = "microsoft-dataverse"
	defaultLoginBaseURL = "https://login.microsoftonline.com"
	defaultPageSize     = 100
	maxPageSize         = 5000
	userAgent           = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Microsoft Dataverse", IntegrationType: "api", Description: "Reads Microsoft Dataverse accounts, contacts, leads, opportunities, and users through the Web API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "accounts", url.Values{"$top": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check microsoft-dataverse: %w", err)
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
		stream = "accounts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("microsoft-dataverse stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
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
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := url.Values{"$top": []string{strconv.Itoa(pageSize)}}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read microsoft-dataverse %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return fmt.Errorf("decode microsoft-dataverse %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := nextLink(resp.Body)
		if err != nil {
			return fmt.Errorf("decode microsoft-dataverse nextLink: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func nextLink(body []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var root map[string]any
	if err := dec.Decode(&root); err != nil {
		return "", err
	}
	next, _ := root["@odata.nextLink"].(string)
	return next, nil
}

func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"accountid": fmt.Sprintf("account-%d", i), "contactid": fmt.Sprintf("contact-%d", i), "leadid": fmt.Sprintf("lead-%d", i), "opportunityid": fmt.Sprintf("opportunity-%d", i), "systemuserid": fmt.Sprintf("user-%d", i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "fullname": fmt.Sprintf("Fixture User %d", i), "emailaddress1": fmt.Sprintf("fixture+%d@example.com", i)}
		rec := endpoint.mapRecord(item)
		rec["connector"] = connectorName
		rec["fixture"] = true
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
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{TokenURL: tokenURL(cfg), ClientID: cfg.Secrets["client_id"], ClientSecret: cfg.Secrets["client_secret"], Scopes: []string{scope(cfg, base)}, Client: c.Client}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent, DefaultHeaders: map[string]string{"OData-MaxVersion": "4.0", "OData-Version": "4.0"}}, nil
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	if cfg.Secrets == nil || strings.TrimSpace(cfg.Secrets["client_id"]) == "" || strings.TrimSpace(cfg.Secrets["client_secret"]) == "" {
		return errors.New("microsoft-dataverse connector requires secrets client_id and client_secret")
	}
	if strings.TrimSpace(cfg.Config["token_url"]) == "" && strings.TrimSpace(cfg.Secrets["tenant_id"]) == "" {
		return errors.New("microsoft-dataverse connector requires secret tenant_id unless config token_url is set")
	}
	return nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		org := strings.TrimRight(strings.TrimSpace(cfg.Config["org_url"]), "/")
		if org == "" {
			return "", errors.New("microsoft-dataverse connector requires config base_url or org_url")
		}
		base = org + "/api/data/v9.2"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("microsoft-dataverse config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("microsoft-dataverse config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("microsoft-dataverse config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func tokenURL(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["token_url"]); v != "" {
		return v
	}
	loginBase := strings.TrimSpace(cfg.Config["login_base_url"])
	if loginBase == "" {
		loginBase = defaultLoginBaseURL
	}
	return strings.TrimRight(loginBase, "/") + "/" + cfg.Secrets["tenant_id"] + "/oauth2/v2.0/token"
}
func scope(cfg connectors.RuntimeConfig, base string) string {
	if v := strings.TrimSpace(cfg.Config["scope"]); v != "" {
		return v
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return base + "/.default"
	}
	return parsed.Scheme + "://" + parsed.Host + "/.default"
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "microsoft-dataverse config page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("microsoft-dataverse config max_pages must be a non-negative integer: %w", err)
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
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
