// Package stigg implements the native pm Stigg connector.
package stigg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName  = "stigg"
	defaultBaseURL = "https://api.stigg.io"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	query       string
	recordsPath string
	fields      []connectors.Field
	mapRecord   func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Stigg", IntegrationType: "api", Description: "Reads Stigg products, plans, customers, and subscriptions through the Stigg GraphQL API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	body := map[string]any{"query": "query PolymetricsCheck { products { id } }"}
	if err := r.DoJSON(ctx, http.MethodPost, "graphql", nil, body, nil); err != nil {
		return fmt.Errorf("check stigg: %w", err)
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
		stream = "products"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("stigg stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodPost, "graphql", nil, map[string]any{"query": endpoint.query})
	if err != nil {
		return fmt.Errorf("read stigg %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode stigg %s: %w", stream, err)
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(rec)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("stigg connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("stigg config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("stigg config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("stigg config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
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
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "refId": fmt.Sprintf("fixture-%d", i), "displayName": fmt.Sprintf("Fixture %s %d", stream, i), "status": "ACTIVE", "customerId": "customer_fixture_1"}
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
		{Name: "products", Description: "Stigg products.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["products"].fields},
		{Name: "plans", Description: "Stigg plans.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["plans"].fields},
		{Name: "customers", Description: "Stigg customers.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["customers"].fields},
		{Name: "subscriptions", Description: "Stigg subscriptions.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["subscriptions"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"products":      {query: "query PolymetricsProducts { products { id refId displayName status } }", recordsPath: "data.products", fields: commonFields(), mapRecord: copyRecord("id", "refId", "displayName", "status")},
	"plans":         {query: "query PolymetricsPlans { plans { id refId displayName status } }", recordsPath: "data.plans", fields: commonFields(), mapRecord: copyRecord("id", "refId", "displayName", "status")},
	"customers":     {query: "query PolymetricsCustomers { customers { id refId displayName status } }", recordsPath: "data.customers", fields: commonFields(), mapRecord: copyRecord("id", "refId", "displayName", "status")},
	"subscriptions": {query: "query PolymetricsSubscriptions { subscriptions { id refId customerId status } }", recordsPath: "data.subscriptions", fields: subscriptionFields(), mapRecord: copyRecord("id", "refId", "customerId", "status")},
}

func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "refId", Type: "string"}, {Name: "displayName", Type: "string"}, {Name: "status", Type: "string"}}
}

func subscriptionFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "refId", Type: "string"}, {Name: "customerId", Type: "string"}, {Name: "status", Type: "string"}}
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
