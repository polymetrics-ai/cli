// Package flexport implements the native pm Flexport connector. It follows the
// stripe declarative-HTTP template: a thin package composing the connsdk toolkit
// (Requester + Bearer auth + nested-envelope extraction + cursor-URL pagination)
// with Flexport-specific stream definitions and endpoints.
//
// Flexport is a freight/logistics platform; this connector is read-only (there
// are no safe reverse-ETL writes), exposing companies, locations, products,
// invoices, and shipments.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package flexport

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
	flexportDefaultBaseURL  = "https://api.flexport.com"
	flexportDefaultPageSize = 100
	flexportMaxPageSize     = 250
	flexportUserAgent       = "polymetrics-go-cli"
	// flexportRecordsPath is the dotted path to the records array inside the
	// Flexport list envelope: {"data":{"data":[...],"next":"<url>"}}.
	flexportRecordsPath = "data.data"
	// flexportNextPath is the dotted path to the absolute next-page URL.
	flexportNextPath = "data.next"
	// flexportPageParam is the query param controlling page size.
	flexportPageSizeParam = "per"
)

func init() {
	connectors.RegisterFactory("flexport", New)
}

// New returns the Flexport connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Flexport connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "flexport" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "flexport",
		DisplayName:     "Flexport",
		IntegrationType: "api",
		Description:     "Reads Flexport companies, locations, products, invoices, and shipments through the Flexport REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Flexport. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := flexportBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(flexportSecret(cfg)) == "" {
		return errors.New("flexport connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the products list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "products", url.Values{flexportPageSizeParam: []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check flexport: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: flexportStreams()}, nil
}

// Write is unsupported: Flexport is a read-only source connector. It satisfies
// the connectors.Connector interface but always reports the operation as
// unsupported rather than mutating the Flexport account.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Flexport stream starts with
// an empty incremental cursor (full sync).
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
		stream = "products"
	}
	endpoint, ok := flexportStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("flexport stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := flexportPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := flexportMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Flexport's cursor pagination. List responses wrap records in a
// nested envelope: {"data":{"data":[...],"next":"<absolute url>"}}. The next page
// is fetched directly from data.next (an absolute URL the Requester uses as-is);
// when null/empty the read is complete. connsdk has no paginator for this exact
// shape, so the loop lives here built on Requester + RecordsAt + StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := url.Values{}
	query.Set(flexportPageSizeParam, strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read flexport %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, flexportRecordsPath)
		if err != nil {
			return fmt.Errorf("decode flexport %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, flexportNextPath)
		if err != nil {
			return fmt.Errorf("decode flexport %s next: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// Subsequent pages use the absolute next URL verbatim; query params are
		// already encoded into it, so clear our base query.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise flexport credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                       fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"_object":                  strings.TrimSuffix(stream, "s"),
			"name":                     fmt.Sprintf("Fixture %s %d", stream, i),
			"legal_name":               fmt.Sprintf("Fixture %s %d LLC", stream, i),
			"dba_name":                 fmt.Sprintf("Fixture %d", i),
			"sku":                      fmt.Sprintf("SKU-%d", i),
			"description":              "fixture record",
			"hts_code":                 "0000.00.0000",
			"country_of_origin":        "US",
			"street_address":           fmt.Sprintf("%d Market St", i),
			"city":                     "San Francisco",
			"state":                    "CA",
			"country_code":             "US",
			"zip":                      "94105",
			"invoice_number":           fmt.Sprintf("INV-%d", i),
			"status":                   "open",
			"currency":                 "USD",
			"total":                    fmt.Sprintf("%d.00", 1000*i),
			"issued_date":              "2026-01-01",
			"due_date":                 "2026-02-01",
			"transportation_mode":      "ocean",
			"freight_type":             "fcl",
			"origin_port":              "CNSHA",
			"destination_port":         "USLAX",
			"estimated_departure_date": "2026-01-05",
			"estimated_arrival_date":   "2026-01-25",
			"created_at":               "2026-01-01T00:00:00Z",
			"updated_at":               fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := flexportBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := flexportSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("flexport connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: flexportUserAgent,
	}, nil
}

func flexportSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// flexportBaseURL resolves and validates the base URL. The default is
// api.flexport.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func flexportBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return flexportDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("flexport config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("flexport config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("flexport config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func flexportPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return flexportDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("flexport config page_size must be an integer: %w", err)
	}
	if value < 1 || value > flexportMaxPageSize {
		return 0, fmt.Errorf("flexport config page_size must be between 1 and %d", flexportMaxPageSize)
	}
	return value, nil
}

func flexportMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("flexport config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("flexport config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
