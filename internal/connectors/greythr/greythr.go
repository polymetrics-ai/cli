// Package greythr implements the native pm greytHR connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package that composes the connsdk toolkit (Requester + RecordsAt extraction)
// with greytHR-specific stream definitions, endpoints, and pagination.
//
// greytHR authenticates with a two-step session-token flow that has no off-the-
// shelf connsdk authenticator: a Basic-auth POST to <domain>/uas/v1/oauth2/
// client-token returns an access_token, which is then sent on every data
// request as an ACCESS-TOKEN header (alongside an x-greythr-domain header). That
// flow lives in sessionTokenAuth below.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package greythr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	greythrDefaultBaseURL  = "https://api.greythr.com"
	greythrDefaultPageSize = 100
	greythrMaxPageSize     = 200
	greythrUserAgent       = "polymetrics-go-cli"
	greythrTokenPath       = "uas/v1/oauth2/client-token"
	greythrTokenHeader     = "ACCESS-TOKEN"
	greythrDomainHeader    = "x-greythr-domain"
)

func init() {
	connectors.RegisterFactory("greythr", New)
}

// New returns the greytHR connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm greytHR connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the login requester. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "greythr" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "greythr",
		DisplayName:     "greytHR",
		IntegrationType: "api",
		Description:     "Reads greytHR employees, profiles, work details, bank details, and users via the greytHR REST API using session-token authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to greytHR. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := greythrBaseURL(cfg); err != nil {
		return err
	}
	if _, err := greythrDomain(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Config["username"]) == "" {
		return errors.New("greythr connector requires config username")
	}
	if strings.TrimSpace(greythrPassword(cfg)) == "" {
		return errors.New("greythr connector requires secret password")
	}
	r, auth, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the first employees page confirms the login handshake
	// and connectivity without mutating anything.
	endpoint := greythrStreamEndpoints["employees"]
	query := url.Values{"page": []string{strconv.Itoa(endpoint.startPage)}, "size": []string{"1"}}
	if _, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil); err != nil {
		auth.discardToken()
		return fmt.Errorf("check greythr: %w", err)
	}
	return nil
}

// Write is unsupported: greytHR is a read-only source. The method exists only
// to satisfy connectors.Connector; Metadata reports Write=false.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: greythrStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "employees"
	}
	endpoint, ok := greythrStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("greythr stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, _, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := greythrPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := greythrMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives greytHR's PageIncrement pagination. Pages are requested with
// page=<n>&size=<pageSize>; the loop stops when a page returns fewer than
// pageSize records (the conventional short-page terminator). The per-stream
// startPage handles greytHR's inconsistent 0- vs 1-based paging.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page, fetched := endpoint.startPage, 0; maxPages == 0 || fetched < maxPages; page, fetched = page+1, fetched+1 {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{
			"page": []string{strconv.Itoa(page)},
			"size": []string{strconv.Itoa(pageSize)},
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read greythr %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode greythr %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise greythr credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"employeeId":       int64(i),
			"id":               int64(i),
			"employeeNo":       fmt.Sprintf("E%03d", i),
			"name":             fmt.Sprintf("Fixture %d", i),
			"firstName":        "Fixture",
			"lastName":         strconv.Itoa(i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"userName":         fmt.Sprintf("fixture%d", i),
			"status":           int64(1),
			"leftorg":          false,
			"admin":            false,
			"deleted":          false,
			"confirmDate":      "2026-01-01",
			"bankName":         int64(1),
			"nickname":         fmt.Sprintf("fix%d", i),
			"onboardingStatus": "completed",
			"connector":        "greythr",
			"fixture":          true,
			"stream":           stream,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the session-token
// authenticator, the resolved base URL, and the x-greythr-domain header. The
// password only ever flows into the Basic-auth login request inside
// sessionTokenAuth; it is never logged. The returned *sessionTokenAuth is also
// handed back so callers (Check) can invalidate a cached token on failure.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, *sessionTokenAuth, error) {
	base, err := greythrBaseURL(cfg)
	if err != nil {
		return nil, nil, err
	}
	domain, err := greythrDomain(cfg)
	if err != nil {
		return nil, nil, err
	}
	username := strings.TrimSpace(cfg.Config["username"])
	if username == "" {
		return nil, nil, errors.New("greythr connector requires config username")
	}
	password := greythrPassword(cfg)
	if strings.TrimSpace(password) == "" {
		return nil, nil, errors.New("greythr connector requires secret password")
	}

	auth := &sessionTokenAuth{
		client:   c.Client,
		loginURL: domainTokenURL(domain),
		username: username,
		password: password,
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           auth,
		UserAgent:      greythrUserAgent,
		DefaultHeaders: map[string]string{greythrDomainHeader: domain},
	}, auth, nil
}

// sessionTokenAuth implements connsdk.Authenticator for greytHR's session-token
// flow: it performs a Basic-auth POST to loginURL once, caches the access_token,
// and applies it as the ACCESS-TOKEN header on every subsequent request. The
// token is fetched lazily and reused for the lifetime of the requester.
type sessionTokenAuth struct {
	client   *http.Client
	loginURL string
	username string
	password string

	mu    sync.Mutex
	token string
}

// Apply ensures a cached token and sets the ACCESS-TOKEN header. It never logs
// the password or token.
func (a *sessionTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set(greythrTokenHeader, token)
	return nil
}

func (a *sessionTokenAuth) discardToken() {
	a.mu.Lock()
	a.token = ""
	a.mu.Unlock()
}

func (a *sessionTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" {
		return a.token, nil
	}
	// The login requester carries its own Basic auth and never retries auth
	// failures, mirroring the upstream SessionTokenAuthenticator login_requester.
	login := &connsdk.Requester{
		Client:    a.client,
		Auth:      connsdk.Basic(a.username, a.password),
		UserAgent: greythrUserAgent,
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := login.DoJSON(ctx, http.MethodPost, a.loginURL, nil, nil, &out); err != nil {
		return "", fmt.Errorf("greythr login: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("greythr login response missing access_token")
	}
	a.token = out.AccessToken
	return a.token, nil
}

func greythrPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// greythrBaseURL resolves and validates the API base URL. The default is
// api.greythr.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func greythrBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return greythrDefaultBaseURL, nil
	}
	if err := validateHTTPURL("base_url", base); err != nil {
		return "", err
	}
	return strings.TrimRight(base, "/"), nil
}

// greythrDomain resolves and validates the host URL used to build the login
// endpoint. It accepts either a bare host (api.example.greythr.com) or a full
// http(s) URL; the scheme/host are validated to bound SSRF risk.
func greythrDomain(cfg connectors.RuntimeConfig) (string, error) {
	domain := strings.TrimSpace(cfg.Config["domain"])
	if domain == "" {
		return "", errors.New("greythr connector requires config domain")
	}
	probe := domain
	if !strings.Contains(domain, "://") {
		probe = "https://" + domain
	}
	if err := validateHTTPURL("domain", probe); err != nil {
		return "", err
	}
	return strings.TrimRight(domain, "/"), nil
}

// domainTokenURL builds the absolute login URL from the domain, which may be a
// bare host or a full http(s) URL.
func domainTokenURL(domain string) string {
	domain = strings.TrimRight(domain, "/")
	if strings.Contains(domain, "://") {
		return domain + "/" + greythrTokenPath
	}
	return "https://" + domain + "/" + greythrTokenPath
}

func validateHTTPURL(field, raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("greythr config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("greythr config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("greythr config %s must include a host", field)
	}
	return nil
}

func greythrPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return greythrDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("greythr config page_size must be an integer: %w", err)
	}
	if value < 1 || value > greythrMaxPageSize {
		return 0, fmt.Errorf("greythr config page_size must be between 1 and %d", greythrMaxPageSize)
	}
	return value, nil
}

func greythrMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("greythr config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("greythr config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
