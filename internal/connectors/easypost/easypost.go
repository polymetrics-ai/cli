// Package easypost implements the native pm EasyPost connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester + HTTP
// Basic auth + RecordsAt extraction + cursor state) with EasyPost-specific
// stream definitions, endpoints, and pagination.
//
// EasyPost authenticates with HTTP Basic auth using the API key as the username
// and an empty password. List endpoints return {"<resource>":[...],
// "has_more":bool} ordered newest-first; the next (older) page is requested with
// before_id=<last id on the current page>.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package easypost

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
	easypostDefaultBaseURL  = "https://api.easypost.com/v2"
	easypostDefaultPageSize = 100
	easypostMaxPageSize     = 100
	easypostUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("easypost", New)
}

// New returns the EasyPost connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm EasyPost connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "easypost" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "easypost",
		DisplayName:     "EasyPost",
		IntegrationType: "api",
		Description:     "Reads EasyPost shipments, trackers, addresses, parcels, and insurances through the EasyPost REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to EasyPost. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := easypostBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(easypostSecret(cfg)) == "" {
		return errors.New("easypost connector requires secret username (API key)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the shipments list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "shipments", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check easypost: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: easypostStreams()}, nil
}

// Write satisfies the connectors.Connector interface. EasyPost is read-only in
// this connector (reverse-ETL writes would mutate live shipping/insurance
// objects), so writes are explicitly unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: an EasyPost stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time.
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
		stream = "shipments"
	}
	endpoint, ok := easypostStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("easypost stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := easypostPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := easypostMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, startDateBound(req), emit)
}

// harvest drives EasyPost's before_id cursor pagination. Lists return
// {"<resource>":[...], "has_more":bool} ordered newest-first; the next (older)
// page is requested with before_id=<last id on the current page>. There is no
// body-token paginator in connsdk for this exact shape, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, startDatetime string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("page_size", strconv.Itoa(pageSize))
	if startDatetime != "" {
		base.Set("start_datetime", startDatetime)
	}

	beforeID := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if beforeID != "" {
			query.Set("before_id", beforeID)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read easypost %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayKey)
		if err != nil {
			return fmt.Errorf("decode easypost %s page: %w", endpoint.resource, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, "id")
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode easypost %s has_more: %w", endpoint.resource, err)
		}
		if hasMore != "true" || lastID == "" {
			return nil
		}
		beforeID = lastID
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise easypost credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"object":             titleCase(strings.TrimSuffix(stream, "s")),
			"mode":               "test",
			"created_at":         fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at":         fmt.Sprintf("2026-01-0%dT01:00:00Z", i),
			"status":             "delivered",
			"status_detail":      "arrived_at_destination",
			"tracking_code":      fmt.Sprintf("TRACK%d", i),
			"reference":          fmt.Sprintf("ref-%d", i),
			"carrier":            "USPS",
			"name":               fmt.Sprintf("Fixture %d", i),
			"company":            "Polymetrics",
			"street1":            "100 Main St",
			"city":               "San Francisco",
			"state":              "CA",
			"zip":                "94105",
			"country":            "US",
			"residential":        false,
			"length":             10.0,
			"width":              8.0,
			"height":             4.0,
			"weight":             16.0,
			"predefined_package": "",
			"amount":             "100.00",
			"provider":           "XCover",
			"shipment_id":        "shp_fixture_1",
			"batch_id":           "",
			"batch_status":       "",
			"is_return":          false,
			"signed_by":          "",
			"est_delivery_date":  fmt.Sprintf("2026-01-0%dT00:00:00Z", i+2),
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

// requester builds a connsdk.Requester wired with HTTP Basic auth (API key as
// username, empty password) and the resolved base URL. The secret only ever
// flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := easypostBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := easypostSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("easypost connector requires secret username (API key)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: easypostUserAgent,
	}, nil
}

// startDateBound returns the RFC3339 start_datetime lower bound derived from the
// incremental cursor (if any) or else the start_date config. An empty result
// means no lower bound (full sync).
func startDateBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func easypostSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["username"]
}

// easypostBaseURL resolves and validates the base URL. The default is
// api.easypost.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func easypostBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return easypostDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("easypost config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("easypost config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("easypost config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func easypostPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return easypostDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("easypost config page_size must be an integer: %w", err)
	}
	if value < 1 || value > easypostMaxPageSize {
		return 0, fmt.Errorf("easypost config page_size must be between 1 and %d", easypostMaxPageSize)
	}
	return value, nil
}

func easypostMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("easypost config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("easypost config max_pages must be 0 for unlimited or a positive integer")
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

// titleCase upper-cases the first rune of s. Used only for the deterministic
// fixture "object" label; ASCII stream names make this sufficient.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
