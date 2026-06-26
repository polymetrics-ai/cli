// Package economic implements the native pm e-conomic connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package that composes the connsdk toolkit (Requester + dual-token header auth
// + RecordsAt extraction) with e-conomic-specific stream definitions, endpoints,
// and pagination.
//
// The e-conomic REST API authenticates with two private headers, X-AppSecretToken
// (the app's secret) and X-AgreementGrantToken (the per-agreement grant). List
// endpoints under https://restapi.e-conomic.com return a JSON object with the
// records under a "collection" array and a "pagination" object whose nextPage
// field is an absolute URL to the following page (skippages/pagesize style). The
// e-conomic source is read-only (full-refresh); it exposes no reverse-ETL writes,
// so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The directory is hyphenated
// (e-conomic) while the Go package name strips the hyphen (economic); the
// registry key is the bare system name "e-conomic".
package economic

import (
	"context"
	"encoding/json"
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
	registryName       = "e-conomic"
	defaultBaseURL     = "https://restapi.e-conomic.com"
	defaultPageSize    = 100
	maxPageSize        = 1000
	userAgent          = "polymetrics-go-cli"
	appSecretHeader    = "X-AppSecretToken"
	agreementGrantHead = "X-AgreementGrantToken"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the e-conomic connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm e-conomic connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "e-conomic",
		IntegrationType: "api",
		Description:     "Reads e-conomic customers, products, suppliers, accounts, and booked invoices through the e-conomic REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to e-conomic. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	appSecret, grant := secrets(cfg)
	if strings.TrimSpace(appSecret) == "" || strings.TrimSpace(grant) == "" {
		return errors.New("e-conomic connector requires secrets app_secret_token and agreement_grant_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers collection confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"skippages": []string{"0"}, "pagesize": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "customers", q, nil, nil); err != nil {
		return fmt.Errorf("check e-conomic: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: economicStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("e-conomic stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	pages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, size, pages, emit)
}

// harvest drives e-conomic's skippages/pagesize pagination. A collection
// response is {collection:[...], pagination:{..., nextPage:"<absolute url>"}};
// when pagination.nextPage is present we follow that absolute URL directly
// (connsdk.Requester treats http(s)-prefixed paths as absolute). The loop lives
// here rather than in a generic paginator because the next-page token is an
// absolute URL embedded in the body, not a header or query token.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First request goes to the resource path with explicit paging params.
	next := ""
	first := url.Values{}
	first.Set("skippages", "0")
	first.Set("pagesize", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var (
			resp *connsdk.Response
			err  error
		)
		if next == "" {
			resp, err = r.Do(ctx, http.MethodGet, endpoint.resource, first, nil)
		} else {
			// next is an absolute URL already carrying its own query string.
			resp, err = r.Do(ctx, http.MethodGet, next, nil, nil)
		}
		if err != nil {
			return fmt.Errorf("read e-conomic %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "collection")
		if err != nil {
			return fmt.Errorf("decode e-conomic %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextPage, err := connsdk.StringAt(resp.Body, "pagination.nextPage")
		if err != nil {
			return fmt.Errorf("decode e-conomic %s nextPage: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(nextPage) == "" {
			return nil
		}
		next = nextPage
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise e-conomic credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		n := json.Number(strconv.Itoa(i))
		self := fmt.Sprintf("%s/%s/%d", defaultBaseURL, endpoint.resource, i)
		item := map[string]any{
			"customerNumber":      n,
			"productNumber":       strconv.Itoa(i),
			"supplierNumber":      n,
			"accountNumber":       n,
			"bookedInvoiceNumber": n,
			"name":                fmt.Sprintf("Fixture %s %d", stream, i),
			"description":         fmt.Sprintf("fixture %s %d", stream, i),
			"currency":            "DKK",
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"city":                "Copenhagen",
			"zip":                 "1000",
			"country":             "Denmark",
			"address":             fmt.Sprintf("%d Test Street", i),
			"barred":              false,
			"balance":             json.Number(strconv.Itoa(100 * i)),
			"creditLimit":         json.Number("10000"),
			"salesPrice":          json.Number(strconv.Itoa(50 * i)),
			"costPrice":           json.Number(strconv.Itoa(25 * i)),
			"recommendedPrice":    json.Number(strconv.Itoa(60 * i)),
			"accountType":         "profitAndLoss",
			"blockDirectEntries":  false,
			"debitCredit":         "debit",
			"vatCode":             "I25",
			"date":                "2026-01-01",
			"dueDate":             "2026-01-31",
			"netAmount":           json.Number(strconv.Itoa(1000 * i)),
			"grossAmount":         json.Number(strconv.Itoa(1250 * i)),
			"vatAmount":           json.Number(strconv.Itoa(250 * i)),
			"remainder":           json.Number("0"),
			"vatZone":             map[string]any{"vatZoneNumber": json.Number("1")},
			"customerGroup":       map[string]any{"customerGroupNumber": json.Number("1")},
			"supplierGroup":       map[string]any{"supplierGroupNumber": json.Number("1")},
			"productGroup":        map[string]any{"productGroupNumber": json.Number("1")},
			"unit":                map[string]any{"unitNumber": json.Number("1")},
			"paymentTerms":        map[string]any{"paymentTermsNumber": json.Number("1")},
			"customer":            map[string]any{"customerNumber": n},
			"self":                self,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the two e-conomic auth headers,
// the resolved base URL, and a JSON Content-Type. The secrets only ever flow into
// DefaultHeaders; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	appSecret, grant := secrets(cfg)
	if strings.TrimSpace(appSecret) == "" || strings.TrimSpace(grant) == "" {
		return nil, errors.New("e-conomic connector requires secrets app_secret_token and agreement_grant_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: userAgent,
		DefaultHeaders: map[string]string{
			appSecretHeader:    strings.TrimSpace(appSecret),
			agreementGrantHead: strings.TrimSpace(grant),
			"Content-Type":     "application/json",
		},
	}, nil
}

func secrets(cfg connectors.RuntimeConfig) (appSecret, grant string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["app_secret_token"], cfg.Secrets["agreement_grant_token"]
}

// baseURL resolves and validates the base URL. The default is
// restapi.e-conomic.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("e-conomic config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("e-conomic config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("e-conomic config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("e-conomic config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("e-conomic config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("e-conomic config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("e-conomic config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write is unsupported: the e-conomic source is read-only (full-refresh), so it
// exposes no reverse-ETL writes. Capabilities.Write is false to match.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
