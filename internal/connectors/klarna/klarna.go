// Package klarna implements the native pm Klarna connector. It is a declarative
// HTTP per-system connector modeled on the stripe reference: a thin package that
// composes the connsdk toolkit (Requester + HTTP Basic auth + RecordsAt
// extraction + offset pagination) with Klarna Settlements API stream
// definitions and endpoints.
//
// Klarna's Settlements API is read-only for reverse-ETL purposes, so the
// connector exposes Read/Check/Catalog and declares Write=false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package klarna

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
	klarnaDefaultPageSize = 100
	klarnaMaxPageSize     = 500
	klarnaUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("klarna", New)
}

// New returns the Klarna connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Klarna connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "klarna" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "klarna",
		DisplayName:     "Klarna",
		IntegrationType: "api",
		Description:     "Reads Klarna settlement payouts and transactions through the Klarna Settlements API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Klarna. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := klarnaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(klarnaSecret(cfg)) == "" {
		return errors.New("klarna connector requires secret password")
	}
	if strings.TrimSpace(klarnaUsername(cfg)) == "" {
		return errors.New("klarna connector requires config username")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the payouts list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "settlements/v1/payouts", url.Values{"size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check klarna: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: klarnaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "payouts"
	}
	endpoint, ok := klarnaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("klarna stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := klarnaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := klarnaMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Klarna Settlements paginates with offset/size; a short page (count < size)
	// signals the end. connsdk's OffsetPaginator implements exactly this.
	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "size",
		OffsetParam: "offset",
		PageSize:    pageSize,
	}
	base := url.Values{}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write is unsupported: the Klarna Settlements API is read-only for reverse-ETL
// purposes, so the connector declares Write=false in Metadata.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise klarna credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"payout_reference":    fmt.Sprintf("payout_fixture_%d", i),
			"transaction_id":      fmt.Sprintf("txn_fixture_%d", i),
			"currency_code":       "EUR",
			"payment_reference":   fmt.Sprintf("pmt_fixture_%d", i),
			"type":                "Sale",
			"amount":              int64(1000 * i),
			"capture_id":          fmt.Sprintf("cap_fixture_%d", i),
			"order_id":            fmt.Sprintf("ord_fixture_%d", i),
			"short_order_id":      fmt.Sprintf("so_%d", i),
			"merchant_reference1": fmt.Sprintf("ref1_%d", i),
			"merchant_reference2": fmt.Sprintf("ref2_%d", i),
			"sale_date":           "2026-01-01T00:00:00Z",
			"capture_date":        "2026-01-02T00:00:00Z",
			"totals": map[string]any{
				"settlement_amount": int64(1000 * i),
				"sale_amount":       int64(1200 * i),
				"return_amount":     int64(0),
				"fee_amount":        int64(200 * i),
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The password only ever flows into connsdk.Basic; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := klarnaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	password := klarnaSecret(cfg)
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("klarna connector requires secret password")
	}
	username := klarnaUsername(cfg)
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("klarna connector requires config username")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, password),
		UserAgent: klarnaUserAgent,
	}, nil
}

func klarnaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// klarnaUsername resolves the merchant UID. It may live in Secrets or Config;
// it is not itself a secret value but the catalog models it alongside password.
func klarnaUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["username"]); v != "" {
			return v
		}
	}
	if cfg.Config != nil {
		return strings.TrimSpace(cfg.Config["username"])
	}
	return ""
}

// klarnaRegionHosts maps a region code to its production API host. The
// playground host inserts ".playground" before ".klarna.com".
var klarnaRegionHosts = map[string]string{
	"eu": "https://api.klarna.com",
	"na": "https://api-na.klarna.com",
	"oc": "https://api-oc.klarna.com",
}

// klarnaBaseURL resolves and validates the base URL. An explicit base_url
// override wins (validated for scheme+host to bound SSRF risk); otherwise the
// region + playground config select an official Klarna host. The default region
// is eu, production.
func klarnaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config != nil {
		if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
			parsed, err := url.Parse(base)
			if err != nil {
				return "", fmt.Errorf("klarna config base_url is invalid: %w", err)
			}
			if parsed.Scheme != "https" && parsed.Scheme != "http" {
				return "", fmt.Errorf("klarna config base_url must use http or https, got %q", parsed.Scheme)
			}
			if parsed.Host == "" {
				return "", errors.New("klarna config base_url must include a host")
			}
			return strings.TrimRight(base, "/"), nil
		}
	}

	region := "eu"
	if cfg.Config != nil {
		if r := strings.TrimSpace(strings.ToLower(cfg.Config["region"])); r != "" {
			region = r
		}
	}
	host, ok := klarnaRegionHosts[region]
	if !ok {
		return "", fmt.Errorf("klarna config region must be one of eu, na, oc, got %q", region)
	}
	if klarnaPlayground(cfg) {
		// https://api.klarna.com -> https://api.playground.klarna.com
		host = strings.Replace(host, ".klarna.com", ".playground.klarna.com", 1)
	}
	return host, nil
}

func klarnaPlayground(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["playground"]))
	switch raw {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

func klarnaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return klarnaDefaultPageSize, nil
	}
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return klarnaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("klarna config page_size must be an integer: %w", err)
	}
	if value < 1 || value > klarnaMaxPageSize {
		return 0, fmt.Errorf("klarna config page_size must be between 1 and %d", klarnaMaxPageSize)
	}
	return value, nil
}

func klarnaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("klarna config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("klarna config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
