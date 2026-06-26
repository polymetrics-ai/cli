// Package finnworlds implements the native pm Finnworlds connector. It follows
// the declarative-HTTP template established by the stripe connector: a thin
// package that composes the connsdk toolkit (Requester + APIKeyQuery auth +
// RecordsAt extraction) with Finnworlds-specific stream definitions, endpoints,
// and a partition fan-out read.
//
// Finnworlds (https://api.finnworlds.com/api/v1/) returns global financial data
// (dividends, stock splits, candlesticks, commodities). Authentication is an API
// key passed as the `key` query parameter. Responses are wrapped as
// {"result":{"output": ...}}; there is no page-token pagination, so a read fans
// out across the configured tickers/commodities list, one request per value.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. The connector is read-only.
package finnworlds

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
	finnworldsDefaultBaseURL = "https://api.finnworlds.com/api/v1"
	finnworldsUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("finnworlds", New)
}

// New returns the Finnworlds connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Finnworlds connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "finnworlds" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "finnworlds",
		DisplayName:     "Finnworlds",
		IntegrationType: "api",
		Description:     "Reads global financial data (dividends, stock splits, historical candlesticks, and commodity prices) from the Finnworlds REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Finnworlds.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := finnworldsBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(finnworldsSecret(cfg)) == "" {
		return errors.New("finnworlds connector requires secret key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the commodities endpoint confirms the key and
	// connectivity without depending on ticker configuration.
	if err := r.DoJSON(ctx, http.MethodGet, "commodities", nil, nil, nil); err != nil {
		return fmt.Errorf("check finnworlds: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: finnworldsStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "commodities"
	}
	endpoint, ok := finnworldsStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("finnworlds stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req.Config, emit)
}

// harvest fans out across the partition list for the stream (tickers or
// commodities). Finnworlds has no page-token pagination: each request returns the
// full dataset for one partition value, so the loop issues one request per value
// and aggregates. Streams with no configured partition (or partitionNone) issue a
// single request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	values := partitionValues(endpoint, cfg)
	for _, value := range values {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		stitch := ""
		if value != "" && endpoint.partitionKey != "" {
			query.Set(endpoint.partitionKey, value)
			stitch = value
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read finnworlds %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode finnworlds %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			record := endpoint.mapRecord(item)
			// Mirror the upstream AddFields transformation: stitch the partition
			// value onto each record so the primary key is complete even when the
			// API payload omits it.
			if stitch != "" && endpoint.stitchField != "" {
				if existing, ok := record[endpoint.stitchField]; !ok || existing == nil || existing == "" {
					record[endpoint.stitchField] = stitch
				}
			}
			if err := emit(record); err != nil {
				return err
			}
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise finnworlds credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"ticker":           "FIX",
			"commodity_name":   "fixture_commodity",
			"date":             fmt.Sprintf("2026-01-0%d", i),
			"datetime":         fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"dividend_rate":    "0.10",
			"stock_split":      "2:1",
			"open":             "100.00",
			"high":             "110.00",
			"low":              "95.00",
			"close":            "105.00",
			"adjusted_close":   "105.00",
			"trade_volume":     "1000000",
			"opentime":         int64(1767225600 + i),
			"closetime":        int64(1767312000 + i),
			"commodity_price":  "75.50",
			"commodity_unit":   "USD/Bbl",
			"price_change_day": "0.50",
			"percentage_day":   "0.66",
			"percentage_week":  "1.20",
			"percentage_month": "3.40",
			"percentage_year":  "12.00",
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with API-key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := finnworldsBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := finnworldsSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("finnworlds connector requires secret key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("key", secret),
		UserAgent: finnworldsUserAgent,
	}, nil
}

// partitionValues resolves the list of partition values a stream fans out over,
// from the matching config list. An empty list yields a single unpartitioned
// request so the stream still produces whatever the API returns by default.
func partitionValues(endpoint streamEndpoint, cfg connectors.RuntimeConfig) []string {
	var raw string
	switch endpoint.partition {
	case partitionTickers:
		raw = cfg.Config["tickers"]
	case partitionCommodities:
		raw = cfg.Config["commodities"]
	default:
		return []string{""}
	}
	values := splitList(raw)
	if len(values) == 0 {
		return []string{""}
	}
	return values
}

// splitList splits a comma-separated config value into trimmed, non-empty items.
func splitList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func finnworldsSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["key"]
}

// finnworldsBaseURL resolves and validates the base URL. The default is
// api.finnworlds.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func finnworldsBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return finnworldsDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("finnworlds config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("finnworlds config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("finnworlds config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Finnworlds is a read-only
// market-data source, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
