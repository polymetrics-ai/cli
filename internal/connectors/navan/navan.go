// Package navan implements the native pm Navan connector. It follows the
// declarative-HTTP per-system connector shape established by the stripe package:
// a thin package that composes the connsdk toolkit (Requester +
// OAuth2ClientCredentials auth + RecordsAt extraction + cursor state) with
// Navan-specific stream definitions, endpoints, and pagination.
//
// Navan authenticates with the OAuth2 client-credentials grant: a client_id and
// client_secret are exchanged at /ta-auth/oauth/token for a bearer token that is
// applied to the data requests. The travel API exposes bookings (filtered by
// bookingType) with page-increment pagination over a `data` array.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package navan

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
	navanDefaultBaseURL  = "https://api.navan.com"
	navanTokenPath       = "ta-auth/oauth/token"
	navanDefaultPageSize = 50
	navanMaxPageSize     = 200
	navanUserAgent       = "polymetrics-go-cli"
	// navanFixtureModified is the deterministic lastModified timestamp used by
	// the fixture-mode records.
	navanFixtureModified = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("navan", New)
}

// New returns the Navan connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Navan connector.
type Connector struct {
	// Client overrides the HTTP client used by both the connsdk Requester and
	// the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "navan" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "navan",
		DisplayName:     "Navan",
		IntegrationType: "api",
		Description:     "Reads Navan travel bookings (flight, hotel, car, and rail) through the Navan REST API using OAuth2 client-credentials authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Navan. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := navanBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(navanClientID(cfg)) == "" || strings.TrimSpace(navanClientSecret(cfg)) == "" {
		return errors.New("navan connector requires secrets client_id and client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the first bookings page confirms the token exchange,
	// auth, and connectivity without mutating anything.
	query := url.Values{}
	query.Set("bookingType", "FLIGHT")
	query.Set("page", "0")
	query.Set("size", "1")
	if _, err := r.Do(ctx, http.MethodGet, "v1/bookings", query, nil); err != nil {
		return fmt.Errorf("check navan: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: navanStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Navan stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "bookings"
	}
	endpoint, ok := navanStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("navan stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := navanPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := navanMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Navan's page-increment pagination. Bookings lists return
// {data:[...]}; the next page is requested with page=<n+1> until a short
// (under-full) page is returned. The loop is built on connsdk.Requester +
// connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, createdFrom string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("size", strconv.Itoa(pageSize))
	for k, v := range endpoint.params {
		base.Set(k, v)
	}
	if createdFrom != "" {
		base.Set("createdFrom", createdFrom)
	}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read navan %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode navan %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested size) signals the end of the
		// result set; an empty page also stops the loop.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise navan credential-free.
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	bookingType := endpoint.params["bookingType"]
	if bookingType == "" {
		bookingType = "FLIGHT"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"uuid":               fmt.Sprintf("bk_fixture_%d", i),
			"bookingId":          fmt.Sprintf("BK-%04d", i),
			"bookingType":        bookingType,
			"bookingStatus":      "CONFIRMED",
			"bookingMethod":      "ONLINE",
			"approvalStatus":     "APPROVED",
			"confirmationNumber": fmt.Sprintf("CONF%04d", i),
			"currency":           "USD",
			"grandTotal":         float64(100 * i),
			"basePrice":          float64(90 * i),
			"bookingFee":         float64(10),
			"destination":        "SFO",
			"startDate":          navanFixtureModified,
			"endDate":            navanFixtureModified,
			"created":            navanFixtureModified,
			"lastModified":       navanFixtureModified,
			"domestic":           true,
			"expensed":           false,
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

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved base URL. The token endpoint is derived from the same base
// URL so test servers and EU/alternate hosts are honoured. The secrets only ever
// flow into the authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := navanBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID := navanClientID(cfg)
	clientSecret := navanClientSecret(cfg)
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New("navan connector requires secrets client_id and client_secret")
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     base + "/" + navanTokenPath,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: navanUserAgent,
	}, nil
}

// incrementalLowerBound returns the createdFrom lower bound, derived from the
// incremental cursor (if any) or else the start_date config. An empty result
// means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func navanClientID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_id"]
}

func navanClientSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

// navanBaseURL resolves and validates the base URL. The default is
// api.navan.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func navanBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return navanDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("navan config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("navan config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("navan config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func navanPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return navanDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("navan config page_size must be an integer: %w", err)
	}
	if value < 1 || value > navanMaxPageSize {
		return 0, fmt.Errorf("navan config page_size must be between 1 and %d", navanMaxPageSize)
	}
	return value, nil
}

func navanMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("navan config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("navan config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: Navan is a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
