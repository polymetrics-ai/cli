// Package apptivo implements the native pm Apptivo connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + APIKeyQuery auth +
// RecordsAt extraction) with Apptivo-specific stream definitions, DAO
// endpoints, and offset pagination.
//
// Apptivo's REST API (https://app.apptivo.com/app/dao/v6/<object>) is a CRM
// read source: customers, contacts, leads, and opportunities are returned as
// full-refresh lists under the JSON "data" key, paged with startIndex/numRecords
// offset parameters and authenticated with apiKey + accessKey query parameters.
//
// Like stripe it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the
// production binary to run that side effect.
package apptivo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	apptivoDefaultBaseURL  = "https://app.apptivo.com"
	apptivoDAOPrefix       = "app/dao/v6"
	apptivoListAction      = "getAll"
	apptivoDefaultPageSize = 100
	apptivoMaxPageSize     = 500
	apptivoUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("apptivo", New)
}

// New returns the Apptivo connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Apptivo connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "apptivo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "apptivo",
		DisplayName:     "Apptivo",
		IntegrationType: "api",
		Description:     "Reads Apptivo CRM customers, contacts, leads, and opportunities through the Apptivo REST DAO API (full refresh, read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Apptivo. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := apptivoBaseURL(cfg); err != nil {
		return err
	}
	apiKey, accessKey := apptivoSecrets(cfg)
	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(accessKey) == "" {
		return errors.New("apptivo connector requires secrets api_key and access_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded single-record read of the customers list confirms auth and
	// connectivity without mutating anything.
	query := url.Values{"a": []string{apptivoListAction}, "numRecords": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, apptivoPath("customers"), query, nil, nil); err != nil {
		return fmt.Errorf("check apptivo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: apptivoStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := apptivoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("apptivo stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := apptivoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := apptivoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Apptivo's offset pagination. Each list call carries the
// a=getAll action; the first request omits startIndex (the API treats a missing
// startIndex as 0 and rejecting an explicit 0 is avoided by mirroring the
// upstream inject_on_first_request:false behaviour). Subsequent pages advance
// startIndex by the page size, supplied alongside numRecords. A page shorter
// than the requested size terminates the loop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("a", apptivoListAction)
		query.Set("numRecords", strconv.Itoa(pageSize))
		if offset > 0 {
			query.Set("startIndex", strconv.Itoa(offset))
		}

		resp, err := r.Do(ctx, http.MethodGet, apptivoPath(endpoint.resource), query, nil)
		if err != nil {
			return fmt.Errorf("read apptivo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode apptivo %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short (or empty) page means we have reached the end.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise apptivo credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idValue := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			endpoint.primaryKey: idValue,
			"id":                idValue,
			"customerName":      fmt.Sprintf("Fixture Customer %d", i),
			"customerNumber":    fmt.Sprintf("CUST-%04d", i),
			"firstName":         fmt.Sprintf("Fixture%d", i),
			"lastName":          "Example",
			"fullName":          fmt.Sprintf("Fixture%d Example", i),
			"companyName":       "Fixture Co",
			"opportunityName":   fmt.Sprintf("Fixture Deal %d", i),
			"emailAddress":      fmt.Sprintf("fixture+%d@example.com", i),
			"phoneNumber":       fmt.Sprintf("555-010%d", i),
			"website":           "https://example.com",
			"currencyCode":      "USD",
			"statusName":        "Active",
			"salesStageName":    "Prospecting",
			"opportunityAmount": fmt.Sprintf("%d000", i),
			"leadSource":        "Web",
			"closingDate":       "2026-12-31",
			"creationDate":      fmt.Sprintf("2026-01-0%d", i),
			"lastUpdateDate":    fmt.Sprintf("2026-02-0%d", i),
			"fixture":           true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with APIKeyQuery auth for apiKey,
// the resolved base URL, and the accessKey carried as a default header so both
// credentials reach the API. Apptivo expects both apiKey and accessKey on every
// request; APIKeyQuery handles apiKey, and accessKey is injected via the
// per-request query in harvest/Check. Secrets only ever flow into the query and
// are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := apptivoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey, accessKey := apptivoSecrets(cfg)
	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(accessKey) == "" {
		return nil, errors.New("apptivo connector requires secrets api_key and access_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      apptivoAuth(apiKey, accessKey),
		UserAgent: apptivoUserAgent,
	}, nil
}

// apptivoAuth applies both apiKey and accessKey as query parameters on every
// request, matching Apptivo's ApiKeyAuthenticator scheme.
func apptivoAuth(apiKey, accessKey string) connsdk.Authenticator {
	return connsdk.AuthFunc(func(_ context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("apiKey", strings.TrimSpace(apiKey))
		q.Set("accessKey", strings.TrimSpace(accessKey))
		req.URL.RawQuery = q.Encode()
		return nil
	})
}

// apptivoPath joins the shared DAO prefix with a stream resource.
func apptivoPath(resource string) string {
	return apptivoDAOPrefix + "/" + resource
}

func apptivoSecrets(cfg connectors.RuntimeConfig) (apiKey, accessKey string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["api_key"], cfg.Secrets["access_key"]
}

// apptivoBaseURL resolves and validates the base URL. The default is
// app.apptivo.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func apptivoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return apptivoDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("apptivo config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("apptivo config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("apptivo config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func apptivoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return apptivoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apptivo config page_size must be an integer: %w", err)
	}
	if value < 1 || value > apptivoMaxPageSize {
		return 0, fmt.Errorf("apptivo config page_size must be between 1 and %d", apptivoMaxPageSize)
	}
	return value, nil
}

func apptivoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apptivo config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("apptivo config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: the Apptivo source is full-refresh read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
