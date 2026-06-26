// Package ebayfulfillment implements the native pm eBay Fulfillment connector.
// It is a declarative-HTTP per-system connector built on the connsdk toolkit,
// following the stripe reference shape: a thin package that composes a Requester
// with eBay-specific stream definitions, endpoints, OAuth2 refresh-token auth,
// and offset/next-link pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The connector is read-only: the eBay Fulfillment API surfaces seller orders,
// line items, shipping fulfillments, and payment disputes, none of which is a
// safe reverse-ETL write target, so Capabilities.Write is false.
package ebayfulfillment

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
	registryName        = "ebay-fulfillment"
	defaultAPIHost      = "https://api.ebay.com"
	defaultTokenURL     = "https://api.ebay.com/identity/v1/oauth2/token"
	defaultPageSize     = 50
	maxPageSize         = 1000
	userAgent           = "polymetrics-go-cli"
	fixtureCreationDate = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the eBay Fulfillment connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm eBay Fulfillment connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "eBay Fulfillment",
		IntegrationType: "api",
		Description:     "Reads eBay seller orders, line items, shipping fulfillments, and payment disputes through the eBay Sell Fulfillment REST API.",
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
	if _, err := apiHost(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(refreshToken(cfg)) == "" {
		return errors.New("ebay-fulfillment connector requires secret refresh_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the orders list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "sell/fulfillment/v1/order", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check ebay-fulfillment: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
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

// Write is unsupported: this connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "orders"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("ebay-fulfillment stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := resolvePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := resolveMaxPages(req.Config)
	if err != nil {
		return err
	}
	filter := dateFilter(req)
	return c.harvest(ctx, r, stream, endpoint, pageSize, maxPages, filter, emit)
}

// harvest drives eBay's offset/next-link pagination. getOrders returns
// {orders:[...], total, limit, offset, next:"<absolute url>"}; the next page is
// requested by following the `next` URL when present, falling back to advancing
// offset when it is absent but a full page was returned.
//
// For order_line_items and shipping_fulfillments the same order payload is
// re-projected: each order's lineItems array is exploded, or the order is mapped
// to a shipment row, respectively.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, stream string, endpoint streamEndpoint, pageSize, maxPages int, filter string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if filter != "" {
		base.Set("filter", filter)
	}

	path := endpoint.resource
	query := cloneValues(base)
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read ebay-fulfillment %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode ebay-fulfillment %s page: %w", stream, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emitProjected(stream, endpoint, item, emit); err != nil {
				return err
			}
		}

		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode ebay-fulfillment %s next: %w", stream, err)
		}
		if strings.TrimSpace(next) != "" {
			// Follow the absolute next URL eBay supplies; clear per-request query
			// since the full URL already carries limit/offset/filter.
			path = next
			query = url.Values{}
			continue
		}
		// No next link: stop once a short page is returned, else advance offset.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
		path = endpoint.resource
		query = cloneValues(base)
		query.Set("offset", strconv.Itoa(offset))
	}
	return nil
}

// emitProjected maps a raw order/dispute item into one or more records depending
// on the requested stream and emits them.
func emitProjected(stream string, endpoint streamEndpoint, item map[string]any, emit func(connectors.Record) error) error {
	switch stream {
	case "order_line_items":
		lineItems, _ := item["lineItems"].([]any)
		for _, li := range lineItems {
			lineItem, ok := li.(map[string]any)
			if !ok {
				continue
			}
			if err := emit(lineItemRecord(item, lineItem)); err != nil {
				return err
			}
		}
		return nil
	case "shipping_fulfillments":
		return emit(shippingFulfillmentRecord(item))
	default:
		return emit(endpoint.mapRecord(item))
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		order := map[string]any{
			"orderId":                fmt.Sprintf("03-%05d", i),
			"legacyOrderId":          fmt.Sprintf("1000000%d", i),
			"creationDate":           fixtureCreationDate,
			"lastModifiedDate":       fixtureCreationDate,
			"orderFulfillmentStatus": "FULFILLED",
			"orderPaymentStatus":     "PAID",
			"sellerId":               "fixture_seller",
			"salesRecordReference":   fmt.Sprintf("%d", 100+i),
			"buyer":                  map[string]any{"username": fmt.Sprintf("buyer_%d", i)},
			"pricingSummary":         map[string]any{"total": map[string]any{"value": fmt.Sprintf("%d.00", 10*i), "currency": "USD"}},
			"lineItems": []any{
				map[string]any{
					"lineItemId":                fmt.Sprintf("li_%d", i),
					"legacyItemId":              fmt.Sprintf("item_%d", i),
					"sku":                       fmt.Sprintf("SKU-%d", i),
					"title":                     fmt.Sprintf("Fixture Product %d", i),
					"quantity":                  int64(i),
					"lineItemFulfillmentStatus": "FULFILLED",
					"total":                     map[string]any{"value": fmt.Sprintf("%d.00", 10*i), "currency": "USD"},
				},
			},
			"fulfillmentStartInstructions": []any{
				map[string]any{
					"fulfillmentInstructionsType": "SHIP_TO",
					"shippingStep": map[string]any{
						"shipTo": map[string]any{
							"fullName":       fmt.Sprintf("Buyer %d", i),
							"contactAddress": map[string]any{"city": "San Jose", "stateOrProvince": "CA", "postalCode": "95125", "countryCode": "US"},
						},
					},
				},
			},
		}

		var record connectors.Record
		switch stream {
		case "order_line_items":
			li := order["lineItems"].([]any)[0].(map[string]any)
			record = lineItemRecord(order, li)
		case "shipping_fulfillments":
			record = shippingFulfillmentRecord(order)
		case "payment_disputes":
			record = paymentDisputeRecord(map[string]any{
				"paymentDisputeId":     fmt.Sprintf("dispute_%d", i),
				"orderId":              fmt.Sprintf("03-%05d", i),
				"paymentDisputeStatus": "OPEN",
				"reason":               "ITEM_NOT_RECEIVED",
				"openDate":             fixtureCreationDate,
				"amount":               map[string]any{"value": fmt.Sprintf("%d.00", 10*i), "currency": "USD"},
				"buyerUsername":        fmt.Sprintf("buyer_%d", i),
			})
		default:
			record = orderRecord(order)
		}
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the OAuth2 refresh-token
// authenticator and the resolved API host. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	host, err := apiHost(cfg)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(refreshToken(cfg))
	if token == "" {
		return nil, errors.New("ebay-fulfillment connector requires secret refresh_token")
	}
	tokenURL, err := tokenEndpoint(cfg)
	if err != nil {
		return nil, err
	}
	auth := &refreshTokenAuth{
		tokenURL:     tokenURL,
		clientID:     strings.TrimSpace(cfg.Config["username"]),
		clientSecret: clientSecret(cfg),
		refreshToken: token,
		scopes:       strings.TrimSpace(cfg.Config["scope"]),
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   host,
		Auth:      auth,
		UserAgent: userAgent,
	}, nil
}

// dateFilter builds the eBay creationdate filter from the incremental cursor (if
// any) or else the start_date config. An empty result means no filter.
func dateFilter(req connectors.ReadRequest) string {
	lower := connsdk.Cursor(req.State)
	if lower == "" {
		lower = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if lower == "" {
		return ""
	}
	return "creationdate:[" + lower + "..]"
}

func refreshToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["refresh_token"]
}

func clientSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// apiHost resolves and validates the API host. The default is api.ebay.com; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func apiHost(cfg connectors.RuntimeConfig) (string, error) {
	host := strings.TrimSpace(cfg.Config["api_host"])
	if host == "" {
		host = strings.TrimSpace(cfg.Config["base_url"])
	}
	if host == "" {
		return defaultAPIHost, nil
	}
	return validateURL("api_host", host)
}

// tokenEndpoint resolves and validates the OAuth2 token endpoint.
func tokenEndpoint(cfg connectors.RuntimeConfig) (string, error) {
	endpoint := strings.TrimSpace(cfg.Config["refresh_token_endpoint"])
	if endpoint == "" {
		return defaultTokenURL, nil
	}
	return validateURL("refresh_token_endpoint", endpoint)
}

func validateURL(field, raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("ebay-fulfillment config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("ebay-fulfillment config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("ebay-fulfillment config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func resolvePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ebay-fulfillment config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("ebay-fulfillment config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func resolveMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ebay-fulfillment config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("ebay-fulfillment config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
