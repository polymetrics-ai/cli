// Package salesforce implements a read-only Salesforce REST API connector.
package salesforce

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
	salesforceName       = "salesforce"
	salesforceAPIVersion = "v60.0"
	salesforceUserAgent  = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("salesforce", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return salesforceName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: salesforceName, DisplayName: "Salesforce", IntegrationType: "api", Description: "Reads Salesforce object metadata and allow-listed Account, Contact, and Lead SOQL queries through the REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(salesforceAccessToken(cfg)) == "" {
		return errors.New("salesforce connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "services/data/"+salesforceVersion(cfg)+"/", nil, nil, nil); err != nil {
		return fmt.Errorf("check salesforce: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: salesforceStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	endpoint, ok := salesforceStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("salesforce stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := salesforceMaxPages(req.Config)
	if err != nil {
		return err
	}
	if endpoint.soql == "" {
		return c.readSObjects(ctx, r, req.Config, endpoint, emit)
	}
	return c.query(ctx, r, req.Config, endpoint, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) readSObjects(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, endpoint salesforceStreamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "services/data/"+salesforceVersion(cfg)+"/sobjects", nil, nil)
	if err != nil {
		return fmt.Errorf("read salesforce sobjects: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "sobjects")
	if err != nil {
		return fmt.Errorf("decode salesforce sobjects: %w", err)
	}
	for _, item := range records {
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) query(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, endpoint salesforceStreamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	path := "services/data/" + salesforceVersion(cfg) + "/query"
	query := url.Values{"q": []string{endpoint.soql}}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read salesforce %s: %w", endpoint.name, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "records")
		if err != nil {
			return fmt.Errorf("decode salesforce %s: %w", endpoint.name, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextRecordsUrl")
		if err != nil {
			return fmt.Errorf("decode salesforce nextRecordsUrl: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint salesforceStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"Id": fmt.Sprintf("00%d", i), "Name": fmt.Sprintf("Fixture %s %d", endpoint.name, i), "Email": fmt.Sprintf("fixture%d@example.com", i), "LastModifiedDate": "2026-01-01T00:00:00Z", "QualifiedApiName": endpoint.name, "Label": endpoint.name}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := salesforceBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := salesforceAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("salesforce connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: salesforceUserAgent}, nil
}

type salesforceStreamEndpoint struct {
	name      string
	soql      string
	mapRecord func(map[string]any) connectors.Record
}

var salesforceStreamEndpoints = map[string]salesforceStreamEndpoint{
	"sobjects": {name: "sobjects", mapRecord: salesforceSObjectRecord},
	"accounts": {name: "accounts", soql: "SELECT Id, Name, LastModifiedDate FROM Account ORDER BY LastModifiedDate ASC", mapRecord: salesforceNamedObjectRecord},
	"contacts": {name: "contacts", soql: "SELECT Id, Name, Email, LastModifiedDate FROM Contact ORDER BY LastModifiedDate ASC", mapRecord: salesforceNamedObjectRecord},
	"leads":    {name: "leads", soql: "SELECT Id, Name, Email, LastModifiedDate FROM Lead ORDER BY LastModifiedDate ASC", mapRecord: salesforceNamedObjectRecord},
}

func salesforceStreams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "last_modified_date", Type: "string"}}
	return []connectors.Stream{
		{Name: "sobjects", Description: "Salesforce object metadata.", PrimaryKey: []string{"qualified_api_name"}, Fields: []connectors.Field{{Name: "qualified_api_name", Type: "string"}, {Name: "label", Type: "string"}}},
		{Name: "accounts", Description: "Salesforce Account records.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "contacts", Description: "Salesforce Contact records.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "leads", Description: "Salesforce Lead records.", PrimaryKey: []string{"id"}, Fields: fields},
	}
}

func salesforceNamedObjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "name": item["Name"], "email": item["Email"], "last_modified_date": item["LastModifiedDate"]}
}

func salesforceSObjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{"qualified_api_name": first(item["qualifiedApiName"], item["QualifiedApiName"], item["name"]), "label": first(item["label"], item["Label"])}
}

func salesforceAccessToken(cfg connectors.RuntimeConfig) string { return cfg.Secrets["access_token"] }

func salesforceBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(firstString(cfg.Config["instance_url"], cfg.Config["base_url"]))
	if base == "" {
		return "", errors.New("salesforce connector requires config instance_url")
	}
	return validBaseURL(salesforceName, base, "")
}

func salesforceVersion(cfg connectors.RuntimeConfig) string {
	version := strings.Trim(strings.TrimSpace(cfg.Config["api_version"]), "/")
	if version == "" {
		return salesforceAPIVersion
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}

func salesforceMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(salesforceName, cfg.Config["max_pages"])
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

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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
