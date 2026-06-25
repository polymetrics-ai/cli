// Package strava implements the native pm Strava connector. It is a
// declarative-HTTP per-system connector that copies the stripe template: a thin
// package composing the connsdk toolkit (Requester + a refresh-token OAuth2
// authenticator + RecordsAt extraction) with Strava-specific stream definitions,
// endpoints, and page/per_page pagination.
//
// Strava authenticates with a short-lived bearer access token obtained by
// exchanging a long-lived refresh_token (plus client_id/client_secret) at
// https://www.strava.com/oauth/token. The connector is read-only: the Strava
// API has no obvious safe reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package strava

import (
	"context"
	"encoding/json"
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
	stravaDefaultBaseURL  = "https://www.strava.com/api/v3"
	stravaDefaultTokenURL = "https://www.strava.com/oauth/token"
	stravaDefaultPageSize = 100
	stravaMaxPageSize     = 200
	stravaUserAgent       = "polymetrics-go-cli"
	// stravaFixtureStart is the deterministic ISO8601 start_date used by the
	// fixture-mode activity records.
	stravaFixtureStart = "2026-01-01T07:00:00Z"
)

func init() {
	connectors.RegisterFactory("strava", New)
}

// New returns the Strava connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Strava connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and by the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "strava" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "strava",
		DisplayName:     "Strava",
		IntegrationType: "api",
		Description:     "Reads the authenticated Strava athlete's profile, activities, lifetime stats, and clubs through the Strava v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Strava. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := stravaBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the athlete profile confirms the refresh-token exchange,
	// auth, and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "athlete", nil, nil, nil); err != nil {
		return fmt.Errorf("check strava: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Strava is read-only: the
// API exposes no safe reverse-ETL writes, so this always reports the operation
// as unsupported and Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: stravaStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an activities sync starts
// from an empty incremental cursor (the start_date config can raise it).
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
		stream = "activities"
	}
	endpoint, ok := stravaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("strava stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	// Validate the base URL (SSRF bound) before any network access.
	if _, err := stravaBaseURL(req.Config); err != nil {
		return err
	}
	resource, err := resolveResource(endpoint.resource, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.list {
		return c.readSingleton(ctx, r, resource, endpoint, req, emit)
	}
	pageSize, err := stravaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := stravaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint, pageSize, maxPages, req, emit)
}

// readSingleton reads a single-object endpoint (athlete, athlete_stats) and
// emits exactly one mapped record.
func (c Connector) readSingleton(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read strava %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode strava %s: %w", resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(injectAthleteID(item, req.Config))); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives Strava's page/per_page pagination over a top-level JSON array.
// A page shorter than per_page (or empty) terminates the loop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, pageSize, maxPages int, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read strava %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode strava %s page: %w", resource, err)
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
// conformance harness can exercise strava credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	athleteID := fixtureAthleteID(req.Config)
	var items []map[string]any
	switch stream {
	case "activities":
		items = []map[string]any{
			{
				"id": int64(1001), "name": "Fixture Morning Run", "type": "Run", "sport_type": "Run",
				"distance": json.Number("5000"), "moving_time": int64(1500), "elapsed_time": int64(1600),
				"total_elevation_gain": json.Number("42"), "start_date": stravaFixtureStart,
				"start_date_local": stravaFixtureStart, "timezone": "(GMT+00:00) UTC",
				"average_speed": json.Number("3.3"), "max_speed": json.Number("4.1"),
				"kudos_count": int64(7), "achievement_count": int64(1),
			},
			{
				"id": int64(1002), "name": "Fixture Evening Ride", "type": "Ride", "sport_type": "Ride",
				"distance": json.Number("20000"), "moving_time": int64(3000), "elapsed_time": int64(3200),
				"total_elevation_gain": json.Number("180"), "start_date": "2026-01-02T18:00:00Z",
				"start_date_local": "2026-01-02T18:00:00Z", "timezone": "(GMT+00:00) UTC",
				"average_speed": json.Number("6.6"), "max_speed": json.Number("12.0"),
				"kudos_count": int64(3), "achievement_count": int64(0),
			},
		}
	case "athlete":
		items = []map[string]any{{
			"id": athleteID, "username": "fixture_runner", "firstname": "Ada", "lastname": "Lovelace",
			"city": "London", "state": "England", "country": "United Kingdom", "sex": "F",
			"weight": json.Number("61.0"), "created_at": stravaFixtureStart, "updated_at": stravaFixtureStart,
		}}
	case "athlete_stats":
		items = []map[string]any{{
			"id":                           athleteID,
			"biggest_ride_distance":        json.Number("120000"),
			"biggest_climb_elevation_gain": json.Number("1500"),
			"recent_ride_totals":           map[string]any{"count": int64(4), "distance": json.Number("80000")},
			"recent_run_totals":            map[string]any{"count": int64(6), "distance": json.Number("42195")},
			"recent_swim_totals":           map[string]any{"count": int64(0), "distance": json.Number("0")},
			"ytd_ride_totals":              map[string]any{"count": int64(40), "distance": json.Number("800000")},
			"ytd_run_totals":               map[string]any{"count": int64(60), "distance": json.Number("421950")},
			"ytd_swim_totals":              map[string]any{"count": int64(2), "distance": json.Number("4000")},
			"all_ride_totals":              map[string]any{"count": int64(400), "distance": json.Number("8000000")},
			"all_run_totals":               map[string]any{"count": int64(600), "distance": json.Number("4219500")},
			"all_swim_totals":              map[string]any{"count": int64(20), "distance": json.Number("40000")},
		}}
	case "clubs":
		items = []map[string]any{
			{
				"id": int64(9001), "name": "Fixture Runners Club", "sport_type": "running",
				"city": "London", "state": "England", "country": "United Kingdom",
				"member_count": int64(128), "private": false, "membership": "member",
				"url": "fixture-runners",
			},
			{
				"id": int64(9002), "name": "Fixture Cyclists Club", "sport_type": "cycling",
				"city": "Cambridge", "state": "England", "country": "United Kingdom",
				"member_count": int64(56), "private": true, "membership": "member",
				"url": "fixture-cyclists",
			},
		}
	default:
		return fmt.Errorf("strava fixture stream %q not found", stream)
	}

	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return err
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

// requester builds a connsdk.Requester wired with the refresh-token OAuth2
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := stravaBaseURL(cfg)
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
		UserAgent: stravaUserAgent,
	}, nil
}

// authenticator builds the refresh-token OAuth2 authenticator from config +
// secrets, validating that the required credentials are present.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	clientID := strings.TrimSpace(cfg.Config["client_id"])
	if clientID == "" {
		return nil, errors.New("strava connector requires config client_id")
	}
	clientSecret := strings.TrimSpace(secret(cfg, "client_secret"))
	if clientSecret == "" {
		return nil, errors.New("strava connector requires secret client_secret")
	}
	refreshToken := strings.TrimSpace(secret(cfg, "refresh_token"))
	if refreshToken == "" {
		return nil, errors.New("strava connector requires secret refresh_token")
	}
	tokenURL, err := stravaTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	return &refreshTokenAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		client:       c.Client,
	}, nil
}

// refreshTokenAuth implements connsdk.Authenticator for Strava's OAuth2
// refresh-token grant. It exchanges client_id/client_secret/refresh_token for a
// short-lived bearer access token at tokenURL, caches it until shortly before
// expiry, and sets Authorization on each request. Secrets are never logged.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && time.Now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	form := url.Values{}
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("strava token: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("strava token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("strava token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
		ExpiresAt   json.Number `json:"expires_at"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("strava token: decode response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("strava token response missing access_token")
	}

	a.token = out.AccessToken
	a.expires = time.Now().Add(6 * time.Hour)
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		a.expires = time.Now().Add(time.Duration(secs) * time.Second)
	} else if at, err := out.ExpiresAt.Int64(); err == nil && at > 0 {
		a.expires = time.Unix(at, 0)
	}
	return a.token, nil
}

// resolveResource substitutes the {athlete_id} placeholder in an endpoint path
// from the athlete_id config, validating it is present and numeric.
func resolveResource(resource string, cfg connectors.RuntimeConfig) (string, error) {
	if !strings.Contains(resource, "{athlete_id}") {
		return resource, nil
	}
	athleteID := strings.TrimSpace(cfg.Config["athlete_id"])
	if athleteID == "" {
		return "", errors.New("strava connector requires config athlete_id for this stream")
	}
	if _, err := strconv.ParseInt(athleteID, 10, 64); err != nil {
		return "", fmt.Errorf("strava config athlete_id must be an integer: %w", err)
	}
	return strings.ReplaceAll(resource, "{athlete_id}", athleteID), nil
}

// injectAthleteID ensures single-object stats records carry an id (the stats
// endpoint response has no id of its own; the primary key is the athlete id).
func injectAthleteID(item map[string]any, cfg connectors.RuntimeConfig) map[string]any {
	if item == nil {
		return item
	}
	if _, ok := item["id"]; ok {
		return item
	}
	if athleteID := strings.TrimSpace(cfg.Config["athlete_id"]); athleteID != "" {
		if n, err := strconv.ParseInt(athleteID, 10, 64); err == nil {
			item["id"] = n
		} else {
			item["id"] = athleteID
		}
	}
	return item
}

func fixtureAthleteID(cfg connectors.RuntimeConfig) int64 {
	if cfg.Config != nil {
		if athleteID := strings.TrimSpace(cfg.Config["athlete_id"]); athleteID != "" {
			if n, err := strconv.ParseInt(athleteID, 10, 64); err == nil {
				return n
			}
		}
	}
	return 17831421
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// stravaBaseURL resolves and validates the base URL. The default is
// www.strava.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func stravaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveHTTPURL(cfg.Config["base_url"], stravaDefaultBaseURL, "base_url")
}

// stravaTokenURL resolves and validates the OAuth token URL with the same SSRF
// bounds as the base URL.
func stravaTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveHTTPURL(cfg.Config["token_url"], stravaDefaultTokenURL, "token_url")
}

func resolveHTTPURL(raw, fallback, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("strava config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("strava config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("strava config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func stravaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return stravaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("strava config page_size must be an integer: %w", err)
	}
	if value < 1 || value > stravaMaxPageSize {
		return 0, fmt.Errorf("strava config page_size must be between 1 and %d", stravaMaxPageSize)
	}
	return value, nil
}

func stravaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("strava config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("strava config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
