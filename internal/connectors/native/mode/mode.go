// Package mode implements the native pm Mode connector. Like stripe it is a thin
// declarative-HTTP package that composes the connsdk toolkit (Requester + Basic
// auth + RecordsAt extraction + cursor state) with Mode-specific stream
// definitions, endpoints, and HAL+JSON pagination.
//
// Mode's REST API is HAL+JSON: list responses look like
// {"_embedded":{"<resource>":[...]}, "_links":{"next":{"href":"..."}}}. Records
// are read from _embedded.<resource> and the next page is followed via
// _links.next.href. Authentication is HTTP Basic with the workspace API token as
// the username and the API secret as the password.
package mode

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
	modeDefaultBaseURL = "https://app.mode.com/api"
	modeUserAgent      = "polymetrics-go-cli"
	// modeAcceptHAL is the content type Mode's API returns for its hypermedia
	// list responses.
	modeAcceptHAL = "application/hal+json"
	// modeMaxPagesGuard bounds the in-package next-link follow loop so a
	// misbehaving server cannot loop forever; 0 in config means use this guard.
	modeDefaultMaxPages = 1000
)

// New returns the Mode connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mode connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mode" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mode",
		DisplayName:     "Mode",
		IntegrationType: "api",
		Description:     "Reads Mode collections (spaces), reports, data sources, groups, and memberships through the Mode REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mode. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := modeBaseURL(cfg); err != nil {
		return err
	}
	workspace, err := modeWorkspace(cfg)
	if err != nil {
		return err
	}
	token, secret := modeCredentials(cfg)
	if strings.TrimSpace(token) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("mode connector requires secrets api_token and api_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the spaces list confirms auth, the workspace, and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, workspace+"/spaces", nil, nil, nil); err != nil {
		return fmt.Errorf("check mode: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Mode is a read-only
// analytics source, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: modeStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Mode stream starts with an
// empty incremental cursor (full sync). Mode's list endpoints are full-refresh,
// so the cursor is informational.
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
		stream = "spaces"
	}
	endpoint, ok := modeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mode stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	workspace, err := modeWorkspace(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := modeMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, workspace+"/"+endpoint.resource, endpoint, maxPages, emit)
}

// harvest drives Mode's HAL+JSON link-following pagination. Each list response
// carries its records under _embedded.<resource> and (when more pages remain) a
// _links.next.href pointing at the following page. The href may be a relative
// path-with-query or an absolute URL; either is fed straight back to the
// Requester, which treats absolute URLs as-is and resolves relative ones against
// base_url.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, firstPath string, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	path := firstPath
	embeddedPath := "_embedded." + endpoint.embedded
	for page := 0; page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read mode %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, embeddedPath)
		if err != nil {
			return fmt.Errorf("decode mode %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "_links.next.href")
		if err != nil {
			return fmt.Errorf("decode mode %s next link: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == path {
			return nil
		}
		path = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mode credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"token":            fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"id":               int64(i),
			"name":             fmt.Sprintf("Fixture %s %d", stream, i),
			"description":      "deterministic fixture record",
			"space_type":       "custom",
			"space_token":      "spaces_fixture_1",
			"account_username": "acme",
			"adapter":          "postgres",
			"database":         "analytics",
			"host":             "db.example.com",
			"username":         fmt.Sprintf("user%d", i),
			"email":            fmt.Sprintf("user%d@example.com", i),
			"state":            "active",
			"archived":         false,
			"public":           false,
			"restricted":       false,
			"queryable":        true,
			"asleep":           false,
			"admin":            i == 1,
			"created_at":       "2026-01-01T00:00:00Z",
			"updated_at":       fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"last_run_at":      "2026-01-02T00:00:00Z",
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

// requester builds a connsdk.Requester wired with Basic auth, the resolved base
// URL, and the HAL+JSON Accept header. The token/secret only ever flow into
// connsdk.Basic; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := modeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token, secret := modeCredentials(cfg)
	if strings.TrimSpace(token) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("mode connector requires secrets api_token and api_secret")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(token, secret),
		UserAgent: modeUserAgent,
		Accept:    modeAcceptHAL,
	}, nil
}

func modeCredentials(cfg connectors.RuntimeConfig) (token, secret string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["api_token"], cfg.Secrets["api_secret"]
}

// modeWorkspace resolves the required workspace (Mode account username) that
// scopes every list endpoint.
func modeWorkspace(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("mode connector requires config workspace")
	}
	workspace := strings.Trim(strings.TrimSpace(cfg.Config["workspace"]), "/")
	if workspace == "" {
		return "", errors.New("mode connector requires config workspace")
	}
	return workspace, nil
}

// modeBaseURL resolves and validates the base URL. The default is app.mode.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func modeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return modeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mode config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mode config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mode config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func modeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return modeDefaultMaxPages, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mode config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 1 {
		return 0, errors.New("mode config max_pages must be a positive integer, all, or unlimited")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
