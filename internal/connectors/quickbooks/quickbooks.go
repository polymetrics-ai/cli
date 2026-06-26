// Package quickbooks implements a read-only QuickBooks Online connector using
// the v3 Query API. It uses SELECT-list queries over allow-listed entities only;
// arbitrary SQL is never accepted from configuration.
package quickbooks

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
	quickBooksDefaultBaseURL  = "https://quickbooks.api.intuit.com"
	quickBooksDefaultPageSize = 1000
	quickBooksMaxPageSize     = 1000
	quickBooksUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("quickbooks", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "quickbooks" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "quickbooks", DisplayName: "QuickBooks", IntegrationType: "api", Description: "Reads QuickBooks Online customers, invoices, payments, accounts, and vendors through the v3 Query API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, realmID, err := c.requester(cfg)
	if err != nil {
		return err
	}
	query := url.Values{}
	query.Set("query", "SELECT * FROM Customer STARTPOSITION 1 MAXRESULTS 1")
	if err := r.DoJSON(ctx, http.MethodGet, "v3/company/"+realmID+"/query", query, nil, nil); err != nil {
		return fmt.Errorf("check quickbooks: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: qbStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := qbEndpoints[stream]
	if !ok {
		return fmt.Errorf("quickbooks stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, realmID, err := c.requester(req.Config)
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
	return c.harvest(ctx, r, realmID, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, realmID string, endpoint qbEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	start := 1
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("query", fmt.Sprintf("SELECT * FROM %s STARTPOSITION %d MAXRESULTS %d", endpoint.entity, start, pageSize))
		resp, err := r.Do(ctx, http.MethodGet, "v3/company/"+realmID+"/query", query, nil)
		if err != nil {
			return fmt.Errorf("read quickbooks %s: %w", endpoint.entity, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "QueryResponse."+endpoint.entity)
		if err != nil {
			return fmt.Errorf("decode quickbooks %s: %w", endpoint.entity, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		start += pageSize
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint qbEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"Id": strconv.Itoa(i), "DisplayName": fmt.Sprintf("Fixture %s %d", endpoint.entity, i), "Name": fmt.Sprintf("Fixture %d", i), "Active": true, "Balance": int64(100 * i), "TotalAmt": int64(100 * i), "DocNumber": fmt.Sprintf("DOC-%d", i)}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, string, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, "", err
	}
	realmID := cleanSegment(strings.TrimSpace(cfg.Config["realm_id"]))
	if realmID == "" {
		return nil, "", errors.New("quickbooks connector requires config realm_id")
	}
	token := strings.TrimSpace(secret(cfg, "access_token"))
	if token == "" {
		return nil, "", errors.New("quickbooks connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: quickBooksUserAgent, Accept: "application/json"}, realmID, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type qbEndpoint struct {
	entity    string
	mapRecord func(map[string]any) connectors.Record
}

var qbEndpoints = map[string]qbEndpoint{
	"customers": {entity: "Customer", mapRecord: qbCustomer},
	"invoices":  {entity: "Invoice", mapRecord: qbInvoice},
	"payments":  {entity: "Payment", mapRecord: qbPayment},
	"accounts":  {entity: "Account", mapRecord: qbAccount},
	"vendors":   {entity: "Vendor", mapRecord: qbVendor},
}

func qbStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "customers", Description: "QuickBooks customers.", PrimaryKey: []string{"id"}, Fields: fields("id", "display_name", "active", "balance")},
		{Name: "invoices", Description: "QuickBooks invoices.", PrimaryKey: []string{"id"}, Fields: fields("id", "doc_number", "customer_ref", "total_amt", "balance")},
		{Name: "payments", Description: "QuickBooks payments.", PrimaryKey: []string{"id"}, Fields: fields("id", "customer_ref", "total_amt", "txn_date")},
		{Name: "accounts", Description: "QuickBooks chart-of-accounts records.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "classification", "account_type")},
		{Name: "vendors", Description: "QuickBooks vendors.", PrimaryKey: []string{"id"}, Fields: fields("id", "display_name", "active", "balance")},
	}
}

func qbCustomer(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "display_name": item["DisplayName"], "active": item["Active"], "balance": item["Balance"]}
}
func qbInvoice(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "doc_number": item["DocNumber"], "customer_ref": refValue(item["CustomerRef"]), "total_amt": item["TotalAmt"], "balance": item["Balance"]}
}
func qbPayment(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "customer_ref": refValue(item["CustomerRef"]), "total_amt": item["TotalAmt"], "txn_date": item["TxnDate"]}
}
func qbAccount(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "name": item["Name"], "classification": item["Classification"], "account_type": item["AccountType"]}
}
func qbVendor(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "display_name": item["DisplayName"], "active": item["Active"], "balance": item["Balance"]}
}

func refValue(v any) any {
	if m, ok := v.(map[string]any); ok {
		return m["value"]
	}
	return v
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
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

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return quickBooksDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("quickbooks config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("quickbooks config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("quickbooks config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cleanSegment(value string) string {
	if value == "" || strings.ContainsAny(value, "/?#") || strings.Contains(value, "..") {
		return ""
	}
	return value
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return quickBooksDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > quickBooksMaxPageSize {
		return 0, fmt.Errorf("quickbooks config page_size must be between 1 and %d", quickBooksMaxPageSize)
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
		return 0, errors.New("quickbooks config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
