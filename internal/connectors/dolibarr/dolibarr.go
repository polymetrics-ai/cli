// Package dolibarr implements the native pm Dolibarr connector. It follows the
// stripe declarative-HTTP template: a thin package that composes the connsdk
// toolkit (Requester + APIKeyHeader auth + RecordsAt extraction) with
// Dolibarr-specific stream definitions, endpoints, and page-based pagination.
//
// Dolibarr exposes a REST API under {base}/api/index.php authenticated with a
// "DOLAPIKEY: <api_key>" header. List endpoints return a top-level JSON array and
// page with the 0-indexed `page` + `limit` query parameters; the end of data is
// signalled either by a short/empty page or by a 404 "No record found" response.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. The connector is read-only.
package dolibarr

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
	dolibarrDefaultBaseURL  = "https://demo.dolibarr.org/api/index.php"
	dolibarrAPIPathSuffix   = "/api/index.php"
	dolibarrDefaultPageSize = 100
	dolibarrMaxPageSize     = 1000
	dolibarrUserAgent       = "polymetrics-go-cli"
	// dolibarrFixtureModified is the deterministic date_modification value used
	// by fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	dolibarrFixtureModified int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("dolibarr", New)
}

// New returns the Dolibarr connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Dolibarr connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "dolibarr" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "dolibarr",
		DisplayName:     "Dolibarr",
		IntegrationType: "api",
		Description:     "Reads Dolibarr ERP/CRM third parties, contacts, products, customer invoices, and orders through the Dolibarr REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Dolibarr. In
// fixture mode it short-circuits without a network call; otherwise it performs a
// bounded authenticated read of the thirdparties list to confirm auth and
// connectivity without mutating anything.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dolibarrBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(dolibarrSecret(cfg)) == "" {
		return errors.New("dolibarr connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	query := url.Values{"limit": []string{"1"}, "page": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "thirdparties", query, nil, nil); err != nil {
		// Dolibarr returns 404 with "No record found" when the list is empty,
		// which still proves auth + connectivity.
		if isNotFound(err) {
			return nil
		}
		return fmt.Errorf("check dolibarr: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dolibarrStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Dolibarr stream starts with
// an empty incremental cursor (full sync). Dolibarr's catalog only advertises
// full_refresh, but the cursor scaffolding mirrors the stripe template.
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
		stream = "thirdparties"
	}
	endpoint, ok := dolibarrStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("dolibarr stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := dolibarrPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := dolibarrMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Dolibarr's page-based pagination. List endpoints return a
// top-level JSON array; the next page is requested with page=page+1 (page is
// 0-indexed). The walk stops when a page returns fewer records than the page size
// (the final page), an empty page, or a 404 "No record found" past the end.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("page", strconv.Itoa(page))
		query.Set("sortfield", "t.rowid")
		query.Set("sortorder", "ASC")

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			if isNotFound(err) {
				// Dolibarr signals "no more records" with a 404 on an
				// out-of-range page; treat it as a clean end-of-data.
				return nil
			}
			return fmt.Errorf("read dolibarr %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode dolibarr %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short or empty page means we have reached the end of the data set.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise dolibarr credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                strconv.Itoa(i),
			"name":              fmt.Sprintf("Fixture %s %d", endpoint.resource, i),
			"ref":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"label":             fmt.Sprintf("Fixture %s %d", endpoint.resource, i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"socid":             "1",
			"lastname":          fmt.Sprintf("Lastname%d", i),
			"firstname":         fmt.Sprintf("Firstname%d", i),
			"type":              "0",
			"status":            "1",
			"statut":            "1",
			"client":            "1",
			"fournisseur":       "0",
			"price":             "100.00",
			"total_ht":          "100.00",
			"total_tva":         "20.00",
			"total_ttc":         "120.00",
			"paye":              "0",
			"billed":            "0",
			"country_code":      "US",
			"date":              dolibarrFixtureModified,
			"date_creation":     dolibarrFixtureModified,
			"date_modification": dolibarrFixtureModified + int64(i),
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

// requester builds a connsdk.Requester wired with DOLAPIKEY header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dolibarrBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := dolibarrSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("dolibarr connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("DOLAPIKEY", secret, ""),
		UserAgent: dolibarrUserAgent,
	}, nil
}

func dolibarrSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// dolibarrBaseURL resolves and validates the base URL.
//
// Precedence: an explicit base_url config override wins (used by tests and
// self-hosted deployments). Otherwise the catalog's my_dolibarr_domain_url config
// (a bare "domain/path" without scheme, per the catalog field) is turned into
// https://<domain>/api/index.php. The default is the public Dolibarr demo.
//
// Any resolved URL must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func dolibarrBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config != nil {
		if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
			return validateBaseURL(base)
		}
		if domain := strings.TrimSpace(cfg.Config["my_dolibarr_domain_url"]); domain != "" {
			return validateBaseURL(domainToBaseURL(domain))
		}
	}
	return dolibarrDefaultBaseURL, nil
}

// domainToBaseURL turns a bare "mydomain.com/dolibarr" (the catalog config shape)
// into a full API base URL, defaulting to https and appending the REST path
// suffix when the caller has not already included it.
func domainToBaseURL(domain string) string {
	domain = strings.TrimSpace(domain)
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}
	domain = strings.TrimRight(domain, "/")
	if !strings.HasSuffix(domain, dolibarrAPIPathSuffix) {
		domain += dolibarrAPIPathSuffix
	}
	return domain
}

func validateBaseURL(base string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dolibarr config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("dolibarr config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("dolibarr config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func dolibarrPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dolibarrDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dolibarr config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dolibarrMaxPageSize {
		return 0, fmt.Errorf("dolibarr config page_size must be between 1 and %d", dolibarrMaxPageSize)
	}
	return value, nil
}

func dolibarrMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dolibarr config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("dolibarr config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// isNotFound reports whether err is a connsdk HTTP 404, which Dolibarr returns
// for an empty list or an out-of-range page.
func isNotFound(err error) bool {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Status == http.StatusNotFound
	}
	return false
}

// Write is unsupported: the Dolibarr connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
