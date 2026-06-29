// Package cloudbeds implements the native pm Cloudbeds connector. It is a
// declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction at "data") with
// Cloudbeds-specific stream definitions and endpoints.
//
// Cloudbeds is a hospitality platform; the API (v1.2) is read-only here:
// Bearer-token auth, page-increment pagination via pageNumber/pageSize, and
// records returned under the top-level "data" array. The connector is
// full-refresh only, matching the upstream upstream source.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package cloudbeds

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
	cloudbedsDefaultBaseURL  = "https://api.cloudbeds.com/api/v1.2"
	cloudbedsDefaultPageSize = 100
	cloudbedsMaxPageSize     = 100
	cloudbedsUserAgent       = "polymetrics-go-cli"
	cloudbedsStartPage       = 1
)

func init() {
	connectors.RegisterFactory("cloudbeds", New)
}

// New returns the Cloudbeds connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Cloudbeds connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "cloudbeds" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "cloudbeds",
		DisplayName:     "Cloudbeds",
		IntegrationType: "api",
		Description:     "Reads Cloudbeds guests, hotels, rooms, reservations, and transactions through the Cloudbeds v1.2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Cloudbeds. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := cloudbedsBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cloudbedsSecret(cfg)) == "" {
		return errors.New("cloudbeds connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the hotels list confirms auth and connectivity without
	// mutating anything.
	query := url.Values{"pageNumber": []string{"1"}, "pageSize": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "getHotels", query, nil, nil); err != nil {
		return fmt.Errorf("check cloudbeds: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: cloudbedsStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "reservations"
	}
	endpoint, ok := cloudbedsStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("cloudbeds stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := cloudbedsPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := cloudbedsMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Cloudbeds page-increment pagination. Lists return
// {success:true, data:[...], count:N}; pages are requested with an incrementing
// pageNumber (starting at 1) and a fixed pageSize. A page with fewer than
// pageSize records is the last page. This shape is a connsdk.PageNumberPaginator,
// but Cloudbeds keys its records at "data", so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := cloudbedsStartPage; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page-cloudbedsStartPage >= maxPages {
			return nil
		}
		query := url.Values{}
		query.Set("pageNumber", strconv.Itoa(page))
		query.Set("pageSize", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read cloudbeds %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode cloudbeds %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested pageSize) means there are no more
		// pages. An empty page also terminates.
		if len(records) < pageSize {
			return nil
		}
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise cloudbeds credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idBase := fmt.Sprintf("%s_fixture_%d", stream, i)
		item := map[string]any{
			"guestID":                idBase,
			"propertyID":             "prop_fixture_1",
			"organizationID":         "org_fixture_1",
			"reservationID":          fmt.Sprintf("res_fixture_%d", i),
			"transactionID":          fmt.Sprintf("txn_fixture_%d", i),
			"guestName":              fmt.Sprintf("Fixture Guest %d", i),
			"guestEmail":             fmt.Sprintf("fixture+%d@example.com", i),
			"propertyName":           "Fixture Hotel",
			"propertyDescription":    "Deterministic fixture property.",
			"propertyCurrency":       "USD",
			"propertyTimezone":       "America/New_York",
			"propertyImage":          "https://example.com/fixture.png",
			"status":                 "confirmed",
			"startDate":              "2026-01-01",
			"endDate":                "2026-01-03",
			"adults":                 int64(2),
			"children":               int64(0),
			"balance":                int64(100 * i),
			"sourceName":             "Direct",
			"origin":                 "website",
			"amount":                 int64(50 * i),
			"currency":               "USD",
			"category":               "room",
			"transactionCategory":    "charge",
			"transactionType":        "charge",
			"transactionCode":        "ROOM",
			"description":            fmt.Sprintf("Fixture transaction %d", i),
			"transactionDateTime":    "2026-01-01T12:00:00",
			"transactionDateTimeUTC": "2026-01-01T17:00:00Z",
			"isMainGuest":            i == 1,
			"isAnonymized":           false,
			"dateCreated":            "2026-01-01 12:00:00",
			"dateModified":           "2026-01-02 09:00:00",
			"roomBlocks":             []any{},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := cloudbedsBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := cloudbedsSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("cloudbeds connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: cloudbedsUserAgent,
	}, nil
}

func cloudbedsSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// cloudbedsBaseURL resolves and validates the base URL. The default is
// api.cloudbeds.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func cloudbedsBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return cloudbedsDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("cloudbeds config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("cloudbeds config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("cloudbeds config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cloudbedsPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return cloudbedsDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("cloudbeds config page_size must be an integer: %w", err)
	}
	if value < 1 || value > cloudbedsMaxPageSize {
		return 0, fmt.Errorf("cloudbeds config page_size must be between 1 and %d", cloudbedsMaxPageSize)
	}
	return value, nil
}

func cloudbedsMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("cloudbeds config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("cloudbeds config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is not supported: Cloudbeds is read-only in this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
