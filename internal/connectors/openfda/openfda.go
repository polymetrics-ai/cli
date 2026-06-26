// Package openfda implements the native pm openFDA connector. It follows the
// declarative-HTTP reference shape established by the stripe connector: a thin
// package that composes the connsdk toolkit (Requester + optional api_key query
// auth + RecordsAt extraction) with openFDA-specific stream definitions,
// endpoints, and offset (skip/limit) pagination.
//
// openFDA (https://open.fda.gov/apis/) is a public, read-only API over FDA
// datasets (drug, device, food). An API key is optional and only raises rate
// limits, so the connector works credential-free; fixture mode lets the
// conformance harness exercise it without any network access.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package openfda

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
	openfdaDefaultBaseURL  = "https://api.fda.gov"
	openfdaDefaultPageSize = 100
	// openfdaMaxPageSize is the per-request `limit` ceiling enforced by openFDA.
	openfdaMaxPageSize = 1000
	// openfdaMaxSkip is openFDA's hard cap on the `skip` (offset) parameter.
	// Reads stop once the next page would exceed it.
	openfdaMaxSkip   = 25000
	openfdaUserAgent = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("openfda", New)
}

// New returns the openFDA connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm openFDA connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "openfda" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "openfda",
		DisplayName:     "OpenFDA",
		IntegrationType: "api",
		Description:     "Reads FDA drug, device, and food datasets (adverse events, labels, and recall enforcement reports) from the public openFDA REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to openFDA. In
// fixture mode it short-circuits without a network call. openFDA is public, so
// no secret is required; Check performs a bounded read to confirm connectivity.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := openfdaBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A single-record read of the drug event endpoint confirms connectivity
	// (and the api_key, if supplied) without mutating anything.
	query := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "drug/event.json", query, nil, nil); err != nil {
		return fmt.Errorf("check openfda: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: openfdaStreams()}, nil
}

// Write satisfies connectors.Connector. openFDA is a public read-only API with
// no reverse-ETL surface, so writes are unsupported (Capabilities.Write=false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "drug_event"
	}
	endpoint, ok := openfdaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("openfda stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := openfdaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := openfdaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, strings.TrimSpace(req.Config.Config["search"]), emit)
}

// harvest drives openFDA's offset pagination. List responses look like
// {meta:{results:{skip,limit,total}}, results:[...]}; the next page is requested
// with an increasing `skip`. There is no connsdk paginator for openFDA's skip
// ceiling, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt
// + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, search string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if search != "" {
		base.Set("search", search)
	}

	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if skip > 0 {
			query.Set("skip", strconv.Itoa(skip))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read openfda %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode openfda %s page: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than requested) means the dataset is exhausted.
		if len(records) < pageSize {
			return nil
		}
		// Respect openFDA's total count when present.
		if total, err := connsdk.StringAt(resp.Body, "meta.results.total"); err == nil && total != "" {
			if t, perr := strconv.Atoi(total); perr == nil && skip+pageSize >= t {
				return nil
			}
		}
		skip += pageSize
		// openFDA caps skip at openfdaMaxSkip; stop before exceeding it.
		if skip > openfdaMaxSkip {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise openfda credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			endpoint.primaryKey:    fmt.Sprintf("%s_fixture_%d", stream, i),
			"safetyreportid":       fmt.Sprintf("%s_fixture_%d", stream, i),
			"id":                   fmt.Sprintf("%s_fixture_%d", stream, i),
			"recall_number":        fmt.Sprintf("%s_fixture_%d", stream, i),
			"mdr_report_key":       fmt.Sprintf("%s_fixture_%d", stream, i),
			"receivedate":          "20260101",
			"report_date":          "20260101",
			"date_received":        "20260101",
			"effective_time":       "20260101",
			"status":               "Ongoing",
			"classification":       "Class II",
			"product_type":         "Drugs",
			"recalling_firm":       "Fixture Pharma",
			"reason_for_recall":    "fixture recall reason",
			"serious":              "1",
			"event_type":           "Malfunction",
			"manufacturer_name":    "Fixture Devices Inc",
			"primarysourcecountry": "US",
			"country":              "United States",
			"connector":            "openfda",
			"fixture":              true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the resolved base URL and, when
// an api_key secret is configured, the optional openFDA api_key query
// authenticator. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := openfdaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	r := &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: openfdaUserAgent,
	}
	if key := openfdaAPIKey(cfg); strings.TrimSpace(key) != "" {
		r.Auth = connsdk.APIKeyQuery("api_key", key)
	}
	return r, nil
}

func openfdaAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// openfdaBaseURL resolves and validates the base URL. The default is
// api.fda.gov; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func openfdaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return openfdaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("openfda config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("openfda config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("openfda config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func openfdaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return openfdaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("openfda config page_size must be an integer: %w", err)
	}
	if value < 1 || value > openfdaMaxPageSize {
		return 0, fmt.Errorf("openfda config page_size must be between 1 and %d", openfdaMaxPageSize)
	}
	return value, nil
}

func openfdaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("openfda config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("openfda config max_pages must be 0 for unlimited or a positive integer")
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
