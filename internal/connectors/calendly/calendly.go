// Package calendly implements the native pm Calendly source connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester
// + Bearer auth + collection extraction + cursor-URL pagination) wired to
// Calendly v2 stream definitions and endpoints. It mirrors the stripe connector
// template.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Calendly is read-only here: the public API surface that is useful to pm is
// list/read of scheduling data, so the connector advertises Write=false.
package calendly

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
	calendlyDefaultBaseURL  = "https://api.calendly.com"
	calendlyDefaultPageSize = 100
	calendlyMaxPageSize     = 100
	calendlyUserAgent       = "polymetrics-go-cli"
	// calendlyFixtureTime is the deterministic timestamp used by fixture records.
	calendlyFixtureTime = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("calendly", New)
}

// New returns the Calendly connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Calendly source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "calendly" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "calendly",
		DisplayName:     "Calendly",
		IntegrationType: "api",
		Description:     "Reads Calendly scheduled events, event types, organization memberships, groups, and the current user through the Calendly v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Calendly. In
// fixture mode it short-circuits without a network call; otherwise it confirms
// auth/connectivity by resolving the current user via /users/me.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := calendlyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(calendlySecret(cfg)) == "" {
		return errors.New("calendly connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if _, err := c.currentUser(ctx, r); err != nil {
		return fmt.Errorf("check calendly: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: calendlyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Calendly stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "scheduled_events"
	}
	endpoint, ok := calendlyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("calendly stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	// Single-resource streams (users/me) return one object, not a collection.
	if endpoint.single {
		item, err := c.currentUser(ctx, r)
		if err != nil {
			return err
		}
		return emit(endpoint.mapRecord(item))
	}

	user, err := c.currentUser(ctx, r)
	if err != nil {
		return err
	}
	base, err := scopeQuery(endpoint.scope, user)
	if err != nil {
		return err
	}
	pageSize, err := calendlyPageSize(req.Config)
	if err != nil {
		return err
	}
	base.Set("count", strconv.Itoa(pageSize))

	maxPages, err := calendlyMaxPages(req.Config)
	if err != nil {
		return err
	}
	if lower := incrementalLowerBound(req, endpoint); lower != "" {
		// scheduled_events supports min_start_time; other streams ignore an
		// unknown filter param, so only apply it where meaningful.
		if endpoint.resource == "scheduled_events" {
			base.Set("min_start_time", lower)
		}
	}

	return c.harvest(ctx, r, endpoint, base, maxPages, emit)
}

// Write is unsupported: the Calendly connector is read-only, so it satisfies the
// Connector interface by returning ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Calendly's cursor pagination. List responses are shaped
// {collection:[...], pagination:{next_page:"<absolute url>", next_page_token}}.
// The next page is fetched by following the absolute next_page URL, so the loop
// lives here built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := base
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read calendly %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "collection")
		if err != nil {
			return fmt.Errorf("decode calendly %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next_page")
		if err != nil {
			return fmt.Errorf("decode calendly %s next_page: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// The next_page is an absolute URL carrying the cursor; the Requester
		// treats an http(s)-prefixed path as absolute, so no extra query merge
		// is needed.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise calendly credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	count := 2
	if endpoint.single {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		uri := fmt.Sprintf("https://api.calendly.com/%s/%s_fixture_%d", endpoint.resource, strings.TrimSuffix(stream, "s"), i)
		item := map[string]any{
			"uri":                  uri,
			"name":                 fmt.Sprintf("Fixture %d", i),
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"slug":                 fmt.Sprintf("fixture-%d", i),
			"status":               "active",
			"active":               true,
			"role":                 "user",
			"kind":                 "solo",
			"duration":             int64(30 * i),
			"timezone":             "UTC",
			"scheduling_url":       uri,
			"start_time":           calendlyFixtureTime,
			"end_time":             calendlyFixtureTime,
			"created_at":           calendlyFixtureTime,
			"updated_at":           calendlyFixtureTime,
			"current_organization": "https://api.calendly.com/organizations/org_fixture",
			"organization":         "https://api.calendly.com/organizations/org_fixture",
			"user": map[string]any{
				"uri":   uri,
				"name":  fmt.Sprintf("Fixture %d", i),
				"email": fmt.Sprintf("fixture+%d@example.com", i),
			},
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

// currentUser resolves /users/me, returning the resource object. Its uri and
// current_organization are used to scope subsequent list requests.
func (c Connector) currentUser(ctx context.Context, r *connsdk.Requester) (map[string]any, error) {
	resp, err := r.Do(ctx, http.MethodGet, "users/me", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("read calendly current user: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "resource")
	if err != nil {
		return nil, fmt.Errorf("decode calendly current user: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("calendly current user response missing resource")
	}
	return map[string]any(records[0]), nil
}

// scopeQuery builds the org/user binding query params for a list endpoint from
// the resolved current user object.
func scopeQuery(s scope, user map[string]any) (url.Values, error) {
	q := url.Values{}
	switch s {
	case scopeNone:
		return q, nil
	case scopeOrg:
		org := asString(user["current_organization"])
		if org == "" {
			return nil, errors.New("calendly current user missing current_organization")
		}
		q.Set("organization", org)
	case scopeUser:
		uri := asString(user["uri"])
		if uri == "" {
			return nil, errors.New("calendly current user missing uri")
		}
		q.Set("user", uri)
	}
	return q, nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := calendlyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := calendlySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("calendly connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: calendlyUserAgent,
	}, nil
}

// incrementalLowerBound returns the RFC3339 lower bound derived from the
// incremental cursor (if any) or else the start_date config. Empty means no
// lower bound (full sync). Only streams with a cursor field use it.
func incrementalLowerBound(req connectors.ReadRequest, endpoint streamEndpoint) string {
	if endpoint.cursor == "" {
		return ""
	}
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func calendlySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// calendlyBaseURL resolves and validates the base URL. The default is
// api.calendly.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func calendlyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return calendlyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("calendly config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("calendly config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("calendly config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func calendlyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return calendlyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("calendly config page_size must be an integer: %w", err)
	}
	if value < 1 || value > calendlyMaxPageSize {
		return 0, fmt.Errorf("calendly config page_size must be between 1 and %d", calendlyMaxPageSize)
	}
	return value, nil
}

func calendlyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("calendly config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("calendly config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// idFromURI reduces a Calendly resource uri to its trailing id segment, used as
// the stable primary key. A nil/empty uri yields "".
func idFromURI(v any) string {
	s := asString(v)
	if s == "" {
		return ""
	}
	s = strings.TrimRight(s, "/")
	if idx := strings.LastIndex(s, "/"); idx >= 0 && idx+1 < len(s) {
		return s[idx+1:]
	}
	return s
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

func asObject(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
