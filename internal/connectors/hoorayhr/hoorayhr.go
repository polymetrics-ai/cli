// Package hoorayhr implements the native pm HoorayHR connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, following
// the stripe reference shape: a thin package that composes a connsdk.Requester
// with HoorayHR-specific stream definitions, endpoints, and a session-token
// authenticator.
//
// HoorayHR (https://hoorayhr.io) uses a session-token auth flow: the connector
// POSTs the configured username/password to /authentication, reads accessToken
// from the JSON response, and injects that token raw into the Authorization
// header of every data request. Stream responses are top-level JSON arrays with
// no pagination, and the source is read-only (full refresh).
//
// Like the other per-system connectors it self-registers with the connectors
// registry via RegisterFactory in init(); the registryset package blank-imports
// this package in the production binary to run that side effect.
package hoorayhr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	hoorayhrDefaultBaseURL = "https://api.hooray.nl"
	hoorayhrLoginPath      = "authentication"
	hoorayhrUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("hoorayhr", New)
}

// New returns the HoorayHR connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm HoorayHR connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "hoorayhr" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "hoorayhr",
		DisplayName:     "HoorayHR",
		IntegrationType: "api",
		Description:     "Reads HoorayHR users, time-off, leave-types, and sick-leave records through the HoorayHR REST API using session-token authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to HoorayHR. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := hoorayhrBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(hoorayhrUsername(cfg)) == "" {
		return errors.New("hoorayhr connector requires config hoorayhrusername")
	}
	if strings.TrimSpace(hoorayhrPassword(cfg)) == "" {
		return errors.New("hoorayhr connector requires secret hoorayhrpassword")
	}
	// Authenticating (which performs the login round-trip) confirms credentials
	// and connectivity without reading any data.
	auth, err := c.authenticator(cfg)
	if err != nil {
		return err
	}
	if _, err := auth.token(ctx); err != nil {
		return fmt.Errorf("check hoorayhr: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: hoorayhrStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	endpoint, ok := hoorayhrStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("hoorayhr stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	// HoorayHR responses are top-level JSON arrays and there is no pagination, so
	// a single GET yields the full stream.
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read hoorayhr %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode hoorayhr %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// Write is unsupported: HoorayHR is a read-only source connector. It satisfies
// the connectors.Connector interface but always returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise hoorayhr credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               i,
			"userId":           100 + i,
			"companyId":        1,
			"leaveTypeId":      i,
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":        fmt.Sprintf("Fixture%d", i),
			"lastName":         "User",
			"jobTitle":         "Engineer",
			"status":           "active",
			"isAdmin":          false,
			"companyStartDate": "2026-01-01",
			"start":            "2026-02-01",
			"end":              "2026-02-05",
			"timeOffType":      "vacation",
			"leaveUnit":        "days",
			"name":             fmt.Sprintf("Leave Type %d", i),
			"icon":             "beach",
			"color":            "#00aa00",
			"budget":           25,
			"default":          false,
			"unpaidLeave":      false,
			"leaveInDays":      true,
			"percentage":       100,
			"actualStart":      "2026-03-01",
			"actualReturn":     "2026-03-03",
			"reportedStart":    "2026-03-01",
			"reportedReturn":   "2026-03-03",
			"notes":            fmt.Sprintf("fixture note %d", i),
			"connector":        "hoorayhr",
			"fixture":          true,
			"createdAt":        "2026-01-01T00:00:00.000Z",
			"updatedAt":        "2026-01-02T00:00:00.000Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the session-token
// authenticator and the resolved base URL.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := hoorayhrBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := c.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: hoorayhrUserAgent,
	}, nil
}

// authenticator builds the session-token authenticator from config/secrets,
// validating that both username and password are present.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (*sessionTokenAuth, error) {
	base, err := hoorayhrBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(hoorayhrUsername(cfg))
	if username == "" {
		return nil, errors.New("hoorayhr connector requires config hoorayhrusername")
	}
	password := hoorayhrPassword(cfg)
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("hoorayhr connector requires secret hoorayhrpassword")
	}
	return &sessionTokenAuth{
		login: &connsdk.Requester{
			Client:    c.Client,
			BaseURL:   base,
			UserAgent: hoorayhrUserAgent,
		},
		username: username,
		password: password,
	}, nil
}

// sessionTokenAuth implements connsdk.Authenticator for HoorayHR's
// SessionTokenAuthenticator flow: it POSTs credentials to /authentication once,
// caches accessToken, and injects it raw into the Authorization header. The
// password and token are never logged.
type sessionTokenAuth struct {
	login    *connsdk.Requester
	username string
	password string

	mu     sync.Mutex
	cached string
}

// Apply ensures a session token has been fetched and sets it on the request.
func (a *sessionTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	return nil
}

// token returns the cached access token, fetching one via /authentication on
// first use.
func (a *sessionTokenAuth) token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cached != "" {
		return a.cached, nil
	}
	body := map[string]any{
		"email":    a.username,
		"password": a.password,
		"strategy": "local",
	}
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	if err := a.login.DoJSON(ctx, http.MethodPost, hoorayhrLoginPath, nil, body, &out); err != nil {
		return "", fmt.Errorf("hoorayhr authentication: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("hoorayhr authentication response missing accessToken")
	}
	a.cached = out.AccessToken
	return a.cached, nil
}

func hoorayhrUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["hoorayhrusername"]
}

func hoorayhrPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["hoorayhrpassword"]
}

// hoorayhrBaseURL resolves and validates the base URL. The default is
// api.hooray.nl; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func hoorayhrBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return hoorayhrDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("hoorayhr config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("hoorayhr config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("hoorayhr config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
