// Package youneedabudgetynab implements a read-only native YNAB API connector.
package youneedabudgetynab

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
	connectorName  = "you-need-a-budget-ynab"
	defaultBaseURL = "https://api.ynab.com/v1"
	userAgent      = "polymetrics-go-cli"
	fixtureTime    = "2026-01-01T00:00:00Z"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "You Need A Budget (YNAB)", IntegrationType: "api", Description: "Reads YNAB budgets, accounts, and transactions through the YNAB REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

type streamEndpoint struct {
	path        func(connectors.RuntimeConfig) string
	recordsPath string
	desc        string
	idKeys      []string
	nameKeys    []string
	cursorKeys  []string
}

var streamOrder = []string{"budgets", "accounts", "transactions"}

var streamEndpoints = map[string]streamEndpoint{
	"budgets":      {path: func(connectors.RuntimeConfig) string { return "budgets" }, recordsPath: "data.budgets", desc: "YNAB budgets.", idKeys: []string{"id"}, nameKeys: []string{"name"}, cursorKeys: []string{"last_modified_on"}},
	"accounts":     {path: budgetPath("accounts"), recordsPath: "data.accounts", desc: "YNAB budget accounts.", idKeys: []string{"id"}, nameKeys: []string{"name"}, cursorKeys: []string{"last_modified_on", "updated_at"}},
	"transactions": {path: budgetPath("transactions"), recordsPath: "data.transactions", desc: "YNAB budget transactions.", idKeys: []string{"id"}, nameKeys: []string{"payee_name", "memo"}, cursorKeys: []string{"date", "last_modified_on"}},
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
	if _, err := r.Do(ctx, http.MethodGet, "budgets", nil, nil); err != nil {
		return fmt.Errorf("check %s: %w", connectorName, err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(streamOrder))
	for _, name := range streamOrder {
		ep := streamEndpoints[name]
		streams = append(streams, connectors.Stream{Name: name, Description: ep.desc, Fields: catalogFields(), PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}})
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = streamOrder[0]
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", connectorName, stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, ep, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, ep.path(req.Config), baseQuery(req.Config), nil)
	if err != nil {
		return fmt.Errorf("read %s %s: %w", connectorName, stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
	if err != nil {
		return fmt.Errorf("decode %s %s: %w", connectorName, stream, err)
	}
	for _, rec := range records {
		if err := emit(mapRecord(ep, rec)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "api_key")
	if token == "" {
		return nil, errors.New("you-need-a-budget-ynab connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := mapRecord(ep, map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i), "last_modified_on": fixtureTime, "date": "2026-01-01"})
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func mapRecord(ep streamEndpoint, in map[string]any) connectors.Record {
	out := connectors.Record{}
	for k, v := range in {
		out[k] = v
	}
	if out["id"] == nil {
		out["id"] = firstValue(in, ep.idKeys)
	}
	if out["name"] == nil {
		out["name"] = firstValue(in, ep.nameKeys)
	}
	if out["updated_at"] == nil {
		out["updated_at"] = firstValue(in, ep.cursorKeys)
	}
	return out
}

func firstValue(in map[string]any, keys []string) any {
	for _, key := range keys {
		if value, ok := in[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func catalogFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
}

func budgetPath(resource string) func(connectors.RuntimeConfig) string {
	return func(cfg connectors.RuntimeConfig) string {
		budgetID := configValue(cfg, "budget_id")
		if budgetID == "" {
			budgetID = "last-used"
		}
		return "budgets/" + url.PathEscape(budgetID) + "/" + resource
	}
}

func baseQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if since := configValue(cfg, "since_date"); since != "" {
		q.Set("since_date", since)
	}
	if limit := configValue(cfg, "limit"); limit != "" {
		q.Set("limit", limit)
	}
	return q
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(configValue(cfg, "mode"), "fixture")
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Secrets[name])
}

func configValue(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config[name])
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := configValue(cfg, "base_url")
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url is invalid", connectorName)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https", connectorName)
	}
	return strings.TrimRight(base, "/"), nil
}

func boundedInt(raw string, def, max int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > max {
		return 0, fmt.Errorf("%s integer config must be between 1 and %d", connectorName, max)
	}
	return value, nil
}
