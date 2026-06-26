// Package ebayfinance implements the native pm eBay Finance connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + offset pagination) with eBay Sell Finances stream
// definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The eBay Sell Finances API (https://developer.ebay.com/api-docs/sell/finances)
// is read-only for a seller's monetary records, so this connector exposes no
// write actions.
package ebayfinance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	registryName          = "ebay-finance"
	ebayDefaultBaseURL    = "https://apiz.ebay.com/sell/finances/v1"
	ebayDefaultPageSize   = 200
	ebayMaxPageSize       = 1000
	ebayUserAgent         = "polymetrics-go-cli"
	ebayFixtureDatePrefix = "2026-01-0"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the eBay Finance connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm eBay Finance connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "eBay Finance",
		IntegrationType: "api",
		Description:     "Reads eBay seller financial data — transactions, payouts, transfers, and the seller funds summary — through the eBay Sell Finances REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to eBay. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := ebayBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(ebaySecret(cfg)) == "" {
		return errors.New("ebay-finance connector requires secret client_access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the seller funds summary confirms auth and connectivity
	// without enumerating any large stream.
	if err := r.DoJSON(ctx, http.MethodGet, "seller_funds_summary", nil, nil, nil); err != nil {
		return fmt.Errorf("check ebay-finance: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: ebayStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (full sync), which the start_date config can raise at read
// time.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "transactions"
	}
	endpoint, ok := ebayStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("ebay-finance stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	if endpoint.singleObject {
		return c.readSingle(ctx, r, endpoint, emit)
	}

	pageSize, err := ebayPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := ebayMaxPages(req.Config)
	if err != nil {
		return err
	}
	filter, err := ebayDateFilter(stream, req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, filter, emit)
}

// harvest drives eBay's limit/offset pagination. List endpoints return
// {total, limit, offset, <recordsField>:[...]}. The loop stops when a short page
// is returned or when offset has advanced past the reported total. The loop lives
// in-package because the offset advances by the records returned, and a filter
// param must be preserved across pages.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, filter string, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		if filter != "" {
			query.Set("filter", filter)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read ebay-finance %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsField)
		if err != nil {
			return fmt.Errorf("decode ebay-finance %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A page shorter than the requested limit means we have reached the end.
		if len(records) < pageSize {
			return nil
		}
		total, _ := ebayInt(resp.Body, "total")
		offset += len(records)
		if total > 0 && offset >= total {
			return nil
		}
	}
	return nil
}

// readSingle reads a single-object endpoint (seller_funds_summary) and emits the
// mapped record.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read ebay-finance %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode ebay-finance %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	count := 2
	if endpoint.singleObject {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureItem(stream, i)
		record := endpoint.mapRecord(item)
		record["connector"] = registryName
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// fixtureItem builds a deterministic raw object shaped like the eBay API
// response for the given stream.
func fixtureItem(stream string, i int) map[string]any {
	date := fmt.Sprintf("%s%dT00:00:00.000Z", ebayFixtureDatePrefix, i)
	money := map[string]any{"value": fmt.Sprintf("%d.00", 10*i), "currency": "USD"}
	switch stream {
	case "payouts":
		return map[string]any{
			"payoutId":         fmt.Sprintf("payout_fixture_%d", i),
			"payoutStatus":     "SUCCEEDED",
			"payoutDate":       date,
			"transactionCount": i,
			"amount":           money,
			"payoutInstrument": map[string]any{"nickname": "Checking", "accountLastFourDigits": "6789"},
		}
	case "transfers":
		return map[string]any{
			"transferId":     fmt.Sprintf("transfer_fixture_%d", i),
			"transferStatus": "COMPLETED",
			"transferType":   "WITHDRAWAL",
			"transferDate":   date,
			"reason":         "SHIPPING_LABEL",
			"amount":         money,
		}
	case "seller_funds_summary":
		return map[string]any{
			"totalFunds":      money,
			"availableFunds":  money,
			"fundsOnHold":     map[string]any{"value": "0.00", "currency": "USD"},
			"processingFunds": map[string]any{"value": "0.00", "currency": "USD"},
		}
	default: // transactions
		return map[string]any{
			"transactionId":     fmt.Sprintf("transaction_fixture_%d", i),
			"transactionType":   "SALE",
			"transactionStatus": "FUNDS_AVAILABLE_FOR_PAYOUT",
			"transactionDate":   date,
			"bookingEntry":      "CREDIT",
			"orderId":           fmt.Sprintf("order_%d", i),
			"amount":            money,
		}
	}
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := ebayBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := ebaySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("ebay-finance connector requires secret client_access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: ebayUserAgent,
	}, nil
}

// ebayDateFilter builds the transactionDate/payoutDate/transferDate range filter
// from the incremental cursor or the start_date config. eBay expects an
// RFC3339-style range filter, e.g.
// transactionDate:[2026-01-01T00:00:00.000Z..]. An empty result means no filter.
func ebayDateFilter(stream string, req connectors.ReadRequest) (string, error) {
	field := dateFilterField(stream)
	if field == "" {
		return "", nil
	}
	lower := connsdk.Cursor(req.State)
	if lower == "" {
		lower = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if lower == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, lower); err != nil {
		return "", fmt.Errorf("ebay-finance config start_date/cursor must be RFC3339: %w", err)
	}
	return fmt.Sprintf("%s:[%s..]", field, lower), nil
}

func dateFilterField(stream string) string {
	switch stream {
	case "transactions":
		return "transactionDate"
	case "payouts":
		return "payoutDate"
	case "transfers":
		return "transactionDate"
	default:
		return ""
	}
}

func ebaySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_access_token"]
}

// ebayBaseURL resolves and validates the base URL. The default is
// apiz.ebay.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func ebayBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		// api_host config mirrors the upstream Airbyte connector's prod/sandbox
		// switch; honour it when base_url is not explicitly overridden.
		if host := strings.TrimSpace(cfg.Config["api_host"]); host != "" {
			parsed, err := url.Parse(host)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return "", fmt.Errorf("ebay-finance config api_host is invalid: %q", host)
			}
			if parsed.Scheme != "https" && parsed.Scheme != "http" {
				return "", fmt.Errorf("ebay-finance config api_host must use http or https, got %q", parsed.Scheme)
			}
			return strings.TrimRight(host, "/") + "/sell/finances/v1", nil
		}
		return ebayDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("ebay-finance config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("ebay-finance config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("ebay-finance config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func ebayPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return ebayDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ebay-finance config page_size must be an integer: %w", err)
	}
	if value < 1 || value > ebayMaxPageSize {
		return 0, fmt.Errorf("ebay-finance config page_size must be between 1 and %d", ebayMaxPageSize)
	}
	return value, nil
}

func ebayMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ebay-finance config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("ebay-finance config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// ebayInt reads an integer field (e.g. "total") out of a response body.
func ebayInt(body []byte, path string) (int, error) {
	s, err := connsdk.StringAt(body, path)
	if err != nil || strings.TrimSpace(s) == "" {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(s))
}

// Write is unsupported: the eBay Sell Finances API is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
