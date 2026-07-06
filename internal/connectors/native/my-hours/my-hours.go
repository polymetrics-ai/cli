// Package myhours implements the native pm My Hours connector. It follows the
// declarative-HTTP shape of the stripe reference connector: a thin package that
// composes the connsdk toolkit (Requester + RecordsAt extraction) with My
// Hours-specific stream definitions and a small login-based authenticator.
//
// My Hours authenticates by exchanging an email/password for a short-lived
// bearer token (POST tokens/login), then sends Authorization: Bearer <token>
// plus an api-version header on every data request. List endpoints return a
// top-level JSON array (no envelope, no pagination); the time_logs report is
// fetched in DateFrom/DateTo windows sized by logs_batch_size days.
package myhours

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL    = "https://api2.myhours.com/api"
	apiVersionHeader  = "api-version"
	apiVersionValue   = "1.0"
	userAgent         = "polymetrics-go-cli"
	defaultLogsBatch  = 30
	maxLogsBatch      = 365
	dateLayout        = "2006-01-02"
	maxTimeLogWindows = 600 // hard ceiling so a bad date range can't loop forever
	fixtureDate       = "2026-01-15"
)

// New returns the My Hours connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm My Hours connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and by the login token exchange. Left nil in production; injectable for
	// tests.
	Client *http.Client
}

func (Connector) Name() string { return "my-hours" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "my-hours",
		DisplayName:     "My Hours",
		IntegrationType: "api",
		Description:     "Reads My Hours clients, projects, team members, tags, and time log activity through the My Hours REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to My Hours. In
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
	if strings.TrimSpace(cfg.Config["email"]) == "" {
		return errors.New("my-hours connector requires config email")
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("my-hours connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of clients confirms login + connectivity without mutating
	// anything.
	if _, err := r.Do(ctx, http.MethodGet, "Clients", nil, nil); err != nil {
		return fmt.Errorf("check my-hours: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: myHoursStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "clients"
	}
	endpoint, ok := myHoursStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("my-hours stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if endpoint.timeRanged {
		return c.readTimeRanged(ctx, r, endpoint, req.Config, emit)
	}
	return c.readList(ctx, r, endpoint, nil, emit)
}

// readList fetches a single top-level-array endpoint and emits each mapped
// record.
func (c Connector) readList(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read my-hours %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode my-hours %s: %w", endpoint.resource, err)
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

// readTimeRanged drives the time_logs report across DateFrom/DateTo windows of
// logs_batch_size days each. This is My Hours' de facto pagination: the API has
// no page tokens, so a long date range is split into bounded windows.
func (c Connector) readTimeRanged(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	start, end, err := timeRange(cfg)
	if err != nil {
		return err
	}
	batch, err := logsBatchSize(cfg)
	if err != nil {
		return err
	}

	windowStart := start
	for i := 0; i < maxTimeLogWindows; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if windowStart.After(end) {
			return nil
		}
		windowEnd := windowStart.AddDate(0, 0, batch-1)
		if windowEnd.After(end) {
			windowEnd = end
		}
		query := url.Values{}
		query.Set("DateFrom", windowStart.Format(dateLayout))
		query.Set("DateTo", windowEnd.Format(dateLayout))
		if err := c.readList(ctx, r, endpoint, query, emit); err != nil {
			return err
		}
		windowStart = windowEnd.AddDate(0, 0, 1)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise my-hours credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          int64(i),
			"logId":       int64(i),
			"name":        fmt.Sprintf("%s fixture %d", stream, i),
			"email":       fmt.Sprintf("fixture+%d@example.com", i),
			"archived":    false,
			"active":      true,
			"billable":    true,
			"date":        fixtureDate,
			"projectName": fmt.Sprintf("Project %d", i),
			"clientName":  fmt.Sprintf("Client %d", i),
			"logDuration": float64(3600 * i),
			"rate":        float64(50),
			"connector":   "my-hours",
			"fixture":     true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the login authenticator, the
// resolved base URL, and the api-version header. The password only ever flows
// into the authenticator's login exchange; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(cfg.Config["email"])
	if email == "" {
		return nil, errors.New("my-hours connector requires config email")
	}
	password := secret(cfg)
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("my-hours connector requires secret password")
	}
	auth := &loginAuthenticator{
		baseURL:  base,
		email:    email,
		password: password,
		client:   c.Client,
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           auth,
		UserAgent:      userAgent,
		DefaultHeaders: map[string]string{apiVersionHeader: apiVersionValue},
	}, nil
}

// loginAuthenticator exchanges email/password for a bearer token via
// POST tokens/login, caches it, and applies it as Authorization: Bearer. My
// Hours tokens are short-lived; the cache is best-effort and refetches on first
// use per requester instance.
type loginAuthenticator struct {
	baseURL  string
	email    string
	password string
	client   *http.Client

	mu    sync.Mutex
	token string
}

func (a *loginAuthenticator) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *loginAuthenticator) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" {
		return a.token, nil
	}
	login := &connsdk.Requester{
		Client:  a.client,
		BaseURL: a.baseURL,
		DefaultHeaders: map[string]string{
			apiVersionHeader: apiVersionValue,
		},
	}
	body := map[string]string{
		"email":     a.email,
		"password":  a.password,
		"grantType": "password",
		"clientId":  "api",
	}
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	if err := login.DoJSON(ctx, http.MethodPost, "tokens/login", nil, body, &out); err != nil {
		return "", fmt.Errorf("my-hours login: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("my-hours login response missing accessToken")
	}
	a.token = out.AccessToken
	return a.token, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// baseURL resolves and validates the base URL. The default is api2.myhours.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("my-hours config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("my-hours config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("my-hours config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// timeRange resolves the [start, end] dates for the time_logs report. start_date
// is required for live reads; end_date defaults to today (UTC).
func timeRange(cfg connectors.RuntimeConfig) (time.Time, time.Time, error) {
	rawStart := strings.TrimSpace(cfg.Config["start_date"])
	if rawStart == "" {
		return time.Time{}, time.Time{}, errors.New("my-hours config start_date is required for time_logs")
	}
	start, err := time.Parse(dateLayout, rawStart)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("my-hours config start_date must be YYYY-MM-DD: %w", err)
	}
	end := time.Now().UTC().Truncate(24 * time.Hour)
	if rawEnd := strings.TrimSpace(cfg.Config["end_date"]); rawEnd != "" {
		end, err = time.Parse(dateLayout, rawEnd)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("my-hours config end_date must be YYYY-MM-DD: %w", err)
		}
	}
	if end.Before(start) {
		end = start
	}
	return start, end, nil
}

func logsBatchSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["logs_batch_size"])
	if raw == "" {
		return defaultLogsBatch, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("my-hours config logs_batch_size must be an integer: %w", err)
	}
	if value < 1 || value > maxLogsBatch {
		return 0, fmt.Errorf("my-hours config logs_batch_size must be between 1 and %d", maxLogsBatch)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: My Hours is read-only for reverse ETL in this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
