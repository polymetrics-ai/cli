// Package keka implements the native pm Keka connector. It is a declarative-HTTP
// per-system connector built on the same shape as the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + a Keka OAuth2 token
// authenticator + RecordsAt extraction + pageNumber/pageSize pagination) with
// Keka-specific stream definitions and endpoints.
//
// Keka is an HRIS/payroll/PSA platform. Its public API authenticates with an
// OAuth2 client-credentials style token exchange (client_id, client_secret,
// api_key, grant_type, scope) and returns paginated list responses shaped as
// {"succeeded":true,"data":[...],"totalPages":N,"pageNumber":M}. The connector is
// read-only.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package keka

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
	kekaDefaultPageSize = 100
	kekaMaxPageSize     = 200
	kekaUserAgent       = "polymetrics-go-cli"
	kekaDefaultScope    = "kekaapi"
	// kekaDefaultGrantType is Keka's custom client-credentials grant value.
	kekaDefaultGrantType = "kekaapi"
)

func init() {
	connectors.RegisterFactory("keka", New)
}

// New returns the Keka connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Keka connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "keka" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "keka",
		DisplayName:     "Keka",
		IntegrationType: "api",
		Description:     "Reads Keka HRIS employees, attendance, leave types, leave requests, clients, and projects through the Keka REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Keka. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := kekaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(kekaSecret(cfg)) == "" {
		return errors.New("keka connector requires secret client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the first page of employees confirms the token exchange
	// and connectivity without mutating anything.
	q := url.Values{"pageNumber": []string{"1"}, "pageSize": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "hris/employees", q, nil, nil); err != nil {
		return fmt.Errorf("check keka: %w", err)
	}
	return nil
}

// Write is unsupported: keka is a read-only source connector. It satisfies the
// connectors.Connector interface while declaring Capabilities.Write=false.
func (Connector) Write(_ context.Context, _ connectors.WriteRequest, _ []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: kekaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "employees"
	}
	endpoint, ok := kekaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("keka stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := kekaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := kekaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Keka's pageNumber/pageSize pagination. List responses are shaped
// {"succeeded":true,"data":[...],"totalPages":N,"pageNumber":M}; the loop advances
// pageNumber until it exceeds totalPages or an empty page is returned. There is no
// body-token paginator in connsdk for this exact shape, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("pageNumber", strconv.Itoa(page))
		query.Set("pageSize", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read keka %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode keka %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop on a short/empty page, or when we have reached the reported last
		// page. totalPages is authoritative when present; the empty-page guard is
		// a fallback for responses that omit it.
		if len(records) == 0 {
			return nil
		}
		totalPages := intAt(resp.Body, "totalPages")
		if totalPages > 0 && page >= totalPages {
			return nil
		}
		if totalPages == 0 && len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise keka credential-free (mirrors stripe's fixture
// intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", stream, i),
			"employeeNumber":   fmt.Sprintf("EMP%03d", i),
			"firstName":        fmt.Sprintf("Fixture%d", i),
			"lastName":         "User",
			"displayName":      fmt.Sprintf("Fixture %d", i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"jobTitle":         "Engineer",
			"department":       "Engineering",
			"employmentStatus": "active",
			"employeeId":       fmt.Sprintf("employees_fixture_%d", i),
			"attendanceDate":   "2026-01-01",
			"shiftStartTime":   "2026-01-01T09:00:00Z",
			"shiftEndTime":     "2026-01-01T17:00:00Z",
			"status":           "Present",
			"totalGrossHours":  8,
			"name":             fmt.Sprintf("Fixture %d", i),
			"identifier":       fmt.Sprintf("LT%d", i),
			"leaveTypeUnit":    "Days",
			"isActive":         true,
			"leaveTypeId":      "leave_types_fixture_1",
			"fromDate":         "2026-01-01",
			"toDate":           "2026-01-02",
			"dayCount":         1,
			"code":             fmt.Sprintf("CODE%d", i),
			"clientId":         "clients_fixture_1",
			"billingType":      "FixedBid",
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

// requester builds a connsdk.Requester wired with the Keka OAuth2 token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := kekaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := kekaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("keka connector requires secret client_secret")
	}
	auth, err := c.authenticator(cfg, secret)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: kekaUserAgent,
	}, nil
}

// authenticator builds the Keka token authenticator from config. It validates the
// token URL for SSRF the same way base_url is validated.
func (c Connector) authenticator(cfg connectors.RuntimeConfig, secret string) (connsdk.Authenticator, error) {
	tokenURL, err := kekaTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	return &kekaTokenAuth{
		TokenURL:     tokenURL,
		ClientID:     strings.TrimSpace(cfg.Config["client_id"]),
		ClientSecret: secret,
		APIKey:       strings.TrimSpace(cfg.Config["api_key"]),
		GrantType:    valueOr(cfg.Config["grant_type"], kekaDefaultGrantType),
		Scope:        valueOr(cfg.Config["scope"], kekaDefaultScope),
		Client:       c.Client,
	}, nil
}

func kekaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

// kekaBaseURL resolves and validates the base URL. Keka base URLs are
// company-specific (e.g. https://<company>.keka.com/api/v1) so there is no global
// default; the override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func kekaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("keka connector requires config base_url (e.g. https://<company>.keka.com/api/v1)")
	}
	return validateURL(base, "base_url")
}

// kekaTokenURL resolves and validates the OAuth2 token endpoint. It defaults to
// Keka's hosted identity endpoint when not overridden.
func kekaTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["token_url"])
	if raw == "" {
		return "https://login.keka.com/connect/token", nil
	}
	return validateURL(raw, "token_url")
}

func validateURL(raw, field string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("keka config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("keka config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("keka config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func kekaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return kekaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("keka config page_size must be an integer: %w", err)
	}
	if value < 1 || value > kekaMaxPageSize {
		return 0, fmt.Errorf("keka config page_size must be between 1 and %d", kekaMaxPageSize)
	}
	return value, nil
}

func kekaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("keka config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("keka config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func valueOr(raw, fallback string) string {
	if v := strings.TrimSpace(raw); v != "" {
		return v
	}
	return fallback
}

// intAt reads a non-negative integer from the JSON body at a dotted path,
// returning 0 when absent or non-numeric.
func intAt(body []byte, path string) int {
	s, err := connsdk.StringAt(body, path)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0
	}
	return n
}
