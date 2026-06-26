package braintree

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
	defaultSandboxBaseURL    = "https://api.sandbox.braintreegateway.com"
	defaultProductionBaseURL = "https://api.braintreegateway.com"
	defaultPageSize          = 100
	defaultMaxPages          = 100
	userAgent                = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("braintree", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type endpoint struct{ resource, recordsPath string }

var endpoints = map[string]endpoint{
	"transactions":  {resource: "transactions", recordsPath: "transactions"},
	"customers":     {resource: "customers", recordsPath: "customers"},
	"subscriptions": {resource: "subscriptions", recordsPath: "subscriptions"},
}

func (Connector) Name() string { return "braintree" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "braintree", DisplayName: "Braintree", IntegrationType: "api", Description: "Reads Braintree transactions, customers, and subscriptions through a read-only HTTP API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	merchant := strings.TrimSpace(cfg.Config["merchant_id"])
	if err := r.DoJSON(ctx, http.MethodGet, "merchants/"+url.PathEscape(merchant)+"/transactions", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check braintree: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "status", Type: "string"}, {Name: "amount", Type: "string"}}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{Name: "transactions", Description: "Braintree transactions.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "customers", Description: "Braintree customers.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "subscriptions", Description: "Braintree subscriptions.", Fields: fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "transactions"
	}
	e, ok := endpoints[stream]
	if !ok {
		return fmt.Errorf("braintree stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	merchant := strings.TrimSpace(req.Config.Config["merchant_id"])
	if merchant == "" {
		return errors.New("braintree connector requires config merchant_id")
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	return readPaged(ctx, r, "merchants/"+url.PathEscape(merchant)+"/"+e.resource, e.recordsPath, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readPaged(ctx context.Context, r *connsdk.Requester, resource, recordsPath string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := "1"
	for i := 0; i < maxPages; i++ {
		query := url.Values{"page": []string{page}, "page_size": []string{strconv.Itoa(pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read braintree %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode braintree %s: %w", resource, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next_page")
		if err != nil || strings.TrimSpace(next) == "" {
			return err
		}
		page = next
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "amount": fmt.Sprintf("%d.00", 10+i), "status": "fixture", "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := braintreeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(cfg.Config["public_key"], secret(cfg, "private_key")), UserAgent: userAgent}, nil
}

func braintreeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if strings.EqualFold(strings.TrimSpace(cfg.Config["environment"]), "production") {
			base = defaultProductionBaseURL
		} else {
			base = defaultSandboxBaseURL
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("braintree config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("braintree config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("braintree config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(cfg.Config["merchant_id"]) == "" {
		return errors.New("braintree connector requires config merchant_id")
	}
	if strings.TrimSpace(cfg.Config["public_key"]) == "" {
		return errors.New("braintree connector requires config public_key")
	}
	if strings.TrimSpace(secret(cfg, "private_key")) == "" {
		return errors.New("braintree connector requires secret private_key")
	}
	return nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("braintree config %s must be a positive integer", key)
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
