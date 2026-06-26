// Package jamfpro implements the native pm Jamf Pro connector. It is a
// declarative-HTTP per-system connector following the stripe reference shape: a
// thin package that composes the connsdk toolkit (Requester + Basic-auth token
// exchange + Bearer auth + RecordsAt extraction) with Jamf-Pro-specific stream
// definitions, endpoints, and page-based pagination.
//
// Jamf Pro's modern API authenticates in two steps: POST Basic credentials to
// /v1/auth/token to obtain a short-lived bearer token, then send that token as
// Authorization: Bearer on every data request. The token is fetched once per
// Read and reused across pages.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The registry key, directory,
// and Name() are all the bare system name "jamf-pro"; the Go package identifier
// is jamfpro because identifiers cannot contain hyphens.
package jamfpro

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
	jamfTokenPath       = "v1/auth/token"
	jamfDefaultPageSize = 100
	jamfMaxPageSize     = 2000
	jamfUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("jamf-pro", New)
}

// New returns the Jamf Pro connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Jamf Pro connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "jamf-pro" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "jamf-pro",
		DisplayName:     "Jamf Pro",
		IntegrationType: "api",
		Description:     "Reads Jamf Pro buildings, departments, categories, and scripts through the Jamf Pro REST API using token-based authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Jamf Pro. In
// fixture mode it short-circuits without a network call. Otherwise it performs
// the token exchange, which confirms credentials and connectivity without
// reading or mutating any data.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := jamfBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(jamfUsername(cfg)) == "" {
		return errors.New("jamf-pro connector requires config username")
	}
	if strings.TrimSpace(jamfPassword(cfg)) == "" {
		return errors.New("jamf-pro connector requires secret password")
	}
	if _, err := c.fetchToken(ctx, cfg); err != nil {
		return fmt.Errorf("check jamf-pro: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: jamfStreams()}, nil
}

// Write is unsupported: jamf-pro is a read-only source connector. It satisfies
// the connectors.Connector interface while declaring no write capability.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "buildings"
	}
	endpoint, ok := jamfStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("jamf-pro stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	pageSize, err := jamfPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := jamfMaxPages(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(ctx, req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Jamf Pro's page-based pagination. List endpoints return
// {totalCount:int, results:[...]}; pages are requested with page=0,1,2,...
// and page-size=<n>. The loop stops when a short page is returned, when the
// running count reaches totalCount, or when maxPages is hit. connsdk has no
// paginator for this exact body shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	seen := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("page-size", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read jamf-pro %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode jamf-pro %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		seen += len(records)

		// Stop on a short page (fewer than requested) — the canonical end-of-list
		// signal for page-based APIs.
		if len(records) < pageSize {
			return nil
		}
		// Also stop if totalCount is known and we have collected all of it.
		if total, ok := jamfTotalCount(resp.Body); ok && seen >= total {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise jamf-pro credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             strconv.Itoa(i),
			"name":           fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"streetAddress1": fmt.Sprintf("%d Example St", i),
			"streetAddress2": "",
			"city":           "Minneapolis",
			"stateProvince":  "MN",
			"zipPostalCode":  "55401",
			"country":        "United States",
			"priority":       int64(i),
			"info":           "fixture info",
			"notes":          "fixture notes",
			"categoryId":     "1",
			"categoryName":   "Fixtures",
			"osRequirements": "13",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth using a freshly
// fetched token, the resolved base URL, and the standard user agent. The token
// and password only ever flow into connsdk auth; they are never logged.
func (c Connector) requester(ctx context.Context, cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := jamfBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token, err := c.fetchToken(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(token),
		UserAgent: jamfUserAgent,
	}, nil
}

// fetchToken performs the Jamf Pro token exchange: POST Basic(username,password)
// to /v1/auth/token, returning the bearer token from the JSON response. The
// password is passed only to connsdk.Basic and is never logged.
func (c Connector) fetchToken(ctx context.Context, cfg connectors.RuntimeConfig) (string, error) {
	base, err := jamfBaseURL(cfg)
	if err != nil {
		return "", err
	}
	username := strings.TrimSpace(jamfUsername(cfg))
	if username == "" {
		return "", errors.New("jamf-pro connector requires config username")
	}
	password := jamfPassword(cfg)
	if strings.TrimSpace(password) == "" {
		return "", errors.New("jamf-pro connector requires secret password")
	}

	tokenReq := &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, password),
		UserAgent: jamfUserAgent,
	}
	var out struct {
		Token   string `json:"token"`
		Expires string `json:"expires"`
	}
	if err := tokenReq.DoJSON(ctx, http.MethodPost, jamfTokenPath, nil, nil, &out); err != nil {
		return "", fmt.Errorf("jamf-pro token exchange: %w", err)
	}
	if strings.TrimSpace(out.Token) == "" {
		return "", errors.New("jamf-pro token exchange returned an empty token")
	}
	return out.Token, nil
}

// jamfTotalCount reads the totalCount field from a paginated response body.
func jamfTotalCount(body []byte) (int, bool) {
	raw, err := connsdk.StringAt(body, "totalCount")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, false
	}
	return n, true
}

func jamfUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func jamfPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// jamfBaseURL resolves and validates the base URL. The default is derived from
// the required `subdomain` config as https://<subdomain>.jamfcloud.com/api. A
// base_url override must be an absolute http/https URL with a host to bound SSRF
// risk.
func jamfBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["base_url"]); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("jamf-pro config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("jamf-pro config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("jamf-pro config base_url must include a host")
		}
		return strings.TrimRight(override, "/"), nil
	}

	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("jamf-pro connector requires config subdomain (or base_url)")
	}
	if strings.ContainsAny(subdomain, "/:@ ") {
		return "", fmt.Errorf("jamf-pro config subdomain %q must be a bare subdomain", subdomain)
	}
	return fmt.Sprintf("https://%s.jamfcloud.com/api", subdomain), nil
}

func jamfPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return jamfDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jamf-pro config page_size must be an integer: %w", err)
	}
	if value < 1 || value > jamfMaxPageSize {
		return 0, fmt.Errorf("jamf-pro config page_size must be between 1 and %d", jamfMaxPageSize)
	}
	return value, nil
}

func jamfMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jamf-pro config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("jamf-pro config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
