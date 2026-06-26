// Package height implements the native pm Height connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a thin
// package that composes a Requester + api-key header auth + RecordsAt extraction
// with Height-specific stream definitions and endpoints. It follows the stripe
// reference connector's shape.
//
// Height is read-only here (the catalog only declares full_refresh syncs). Like
// the other per-system connectors it self-registers with the connectors registry
// via RegisterFactory in init(); the registryset package blank-imports this
// package in the production binary to run that side effect.
package height

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
	heightDefaultBaseURL = "https://api.height.app"
	heightUserAgent      = "polymetrics-go-cli"
	// heightFixtureCreated is the deterministic createdAt used by fixture records.
	heightFixtureCreated = "2026-01-01T00:00:00.000Z"
	// heightMaxPages bounds the pagination loop as a safety stop. 0 means use the
	// default; an explicit max_pages config can raise or lower it.
	heightDefaultMaxPages = 1000
)

func init() {
	connectors.RegisterFactory("height", New)
}

// New returns the Height connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Height connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "height" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "height",
		DisplayName:     "Height",
		IntegrationType: "api",
		Description:     "Reads Height tasks, lists, field templates, users, and workspace through the Height REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Height. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := heightBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(heightSecret(cfg)) == "" {
		return errors.New("height connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// The workspace endpoint is a cheap, read-only confirmation of auth and
	// connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "workspace", nil, nil, nil); err != nil {
		return fmt.Errorf("check height: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: heightStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Height stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time.
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
		stream = "tasks"
	}
	endpoint, ok := heightStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("height stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := heightMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, maxPages, emit)
}

// harvest drives Height's list reads. Most endpoints return a {"list":[...]}
// envelope; the workspace endpoint returns a single object at the root. Paginated
// endpoints (tasks) carry a nextPageToken in the body that is supplied as the
// `after` query param on the following request. This exact body-token shape has no
// dedicated connsdk paginator, so the bounded loop lives here on top of
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	after := ""
	for page := 0; page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("usePagination", "true")
			if after != "" {
				query.Set("after", after)
			}
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read height %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode height %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if !endpoint.paginated {
			return nil
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode height %s nextPageToken: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "null" || len(records) == 0 {
			return nil
		}
		after = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise height credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	count := 2
	if stream == "workspace" {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"model":         stream,
			"name":          fmt.Sprintf("Fixture %s %d", stream, i),
			"key":           fmt.Sprintf("KEY-%d", i),
			"createdAt":     heightFixtureCreated,
			"updatedAt":     heightFixtureCreated,
			"createdUserId": "user_fixture_1",
			"userId":        "user_fixture_1",
			"status":        "backlog",
			"type":          "list",
			"completed":     false,
			"deleted":       false,
			"archived":      false,
			"hidden":        false,
			"required":      false,
			"frozen":        false,
			"admin":         false,
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"username":      fmt.Sprintf("fixture%d", i),
			"firstname":     "Fixture",
			"lastname":      strconv.Itoa(i),
			"state":         "enabled",
			"standardType":  "custom",
			"url":           "https://height.app/fixture",
			"urlType":       "default",
			"index":         i,
			"defaultList":   false,
			"visualization": "list",
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

// requester builds a connsdk.Requester wired with api-key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := heightBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := heightSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("height connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "api-key "),
		UserAgent: heightUserAgent,
	}, nil
}

func heightSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// heightBaseURL resolves and validates the base URL. The default is
// api.height.app; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func heightBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return heightDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("height config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("height config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("height config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func heightMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return heightDefaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		// Still bound with a generous cap to avoid an unbounded loop on a
		// misbehaving nextPageToken.
		return heightDefaultMaxPages, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("height config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 1 {
		return 0, errors.New("height config max_pages must be a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Height is read-only in this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
