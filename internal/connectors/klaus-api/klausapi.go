// Package klausapi implements the native pm Klaus (Zendesk QA) connector. It is a
// declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction + date-window state) with
// Klaus-specific stream definitions and account/workspace-scoped endpoints.
//
// It mirrors the stripe reference connector: a thin package that self-registers
// with the connectors registry via RegisterFactory in init(); the registryset
// package blank-imports it in the production binary to run that side effect.
//
// Klaus is read-only here (quality-review data has no obvious safe reverse-ETL
// writes), so Capabilities.Write is false.
package klausapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	klausConnectorName  = "klaus-api"
	klausDefaultBaseURL = "https://kibbles.klausapp.com/api/v2"
	klausUserAgent      = "polymetrics-go-cli"
	klausTimeFormat     = "2006-01-02T15:04:05Z"
	// klausWindow is the review date-window step (manifest uses P1W).
	klausWindow = 7 * 24 * time.Hour
	// klausMaxWindows bounds the windowed loop so an unbounded/empty config can
	// never spin forever.
	klausMaxWindows = 520 // ~10 years of weekly windows
)

func init() {
	connectors.RegisterFactory(klausConnectorName, New)
}

// New returns the Klaus connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Klaus connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return klausConnectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            klausConnectorName,
		DisplayName:     "Klaus API",
		IntegrationType: "api",
		Description:     "Reads Klaus (Zendesk QA) users, rating categories, and review conversations through the Klaus public REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Klaus. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := klausBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(klausSecret(cfg)) == "" {
		return errors.New("klaus-api connector requires secret api_key")
	}
	account, err := klausAccount(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the account users list confirms auth and connectivity
	// without mutating anything.
	path := fmt.Sprintf("account/%s/users", account)
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check klaus-api: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: klausStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Klaus stream starts with an
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
		stream = "users"
	}
	def, ok := klausStreamDefs[stream]
	if !ok {
		return fmt.Errorf("klaus-api stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, def, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := klausPath(def, req.Config)
	if err != nil {
		return err
	}

	if def.windowed {
		return c.harvestWindowed(ctx, r, path, def, req, emit)
	}
	return c.harvestSingle(ctx, r, path, def, emit)
}

// harvestSingle reads a single non-paginated list response (users, categories).
func (c Connector) harvestSingle(ctx context.Context, r *connsdk.Requester, path string, def streamDef, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read klaus-api %s: %w", def.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
	if err != nil {
		return fmt.Errorf("decode klaus-api %s: %w", def.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvestWindowed drives the reviews date-window pagination. The Klaus reviews
// endpoint takes fromDate/toDate filters; the manifest walks forward in weekly
// (P1W) windows from start_date to now. There is no body/page token, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvestWindowed(ctx context.Context, r *connsdk.Requester, path string, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	start, err := klausWindowStart(req)
	if err != nil {
		return err
	}
	end, err := klausWindowEnd(req)
	if err != nil {
		return err
	}
	if !end.After(start) {
		end = start.Add(klausWindow)
	}

	for window := 0; start.Before(end) && window < klausMaxWindows; window++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		to := start.Add(klausWindow)
		if to.After(end) {
			to = end
		}
		query := url.Values{}
		query.Set("fromDate", start.UTC().Format(klausTimeFormat))
		query.Set("toDate", to.UTC().Format(klausTimeFormat))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read klaus-api %s window: %w", def.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
		if err != nil {
			return fmt.Errorf("decode klaus-api %s window: %w", def.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(def.mapRecord(item)); err != nil {
				return err
			}
		}
		start = to
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise klaus-api credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             fmt.Sprintf("%s_fixture_%d", def.resource, i),
			"name":           fmt.Sprintf("Fixture %d", i),
			"email":          fmt.Sprintf("fixture+%d@example.com", i),
			"description":    "fixture category",
			"groupId":        "grp_fixture",
			"groupName":      "Fixture Group",
			"groupPosition":  i,
			"position":       i,
			"maxRating":      5,
			"weight":         1.0,
			"critical":       false,
			"archived":       false,
			"externalId":     fmt.Sprintf("ext_%d", i),
			"externalUrl":    "https://example.com/conversation",
			"url":            "https://example.com/conversation",
			"sourceType":     "zendesk",
			"workspaceId":    "7",
			"createdAt":      "1767225600",
			"createdAtISO":   "2026-01-01T00:00:00Z",
			"lastUpdatedISO": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at":     fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"reviews":        []any{},
			"comments":       []any{},
		}
		record := def.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := klausBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := klausSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("klaus-api connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: klausUserAgent,
	}, nil
}

// klausPath templates the account/workspace-scoped request path for a stream.
func klausPath(def streamDef, cfg connectors.RuntimeConfig) (string, error) {
	account, err := klausAccount(cfg)
	if err != nil {
		return "", err
	}
	switch def.scope {
	case scopeAccount:
		return fmt.Sprintf("account/%s/%s", account, def.resource), nil
	case scopeWorkspace:
		workspace, err := klausWorkspace(cfg)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("account/%s/workspace/%s/%s", account, workspace, def.resource), nil
	default:
		return "", fmt.Errorf("klaus-api unknown scope for resource %q", def.resource)
	}
}

func klausSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// klausAccount returns the required account id from config.
func klausAccount(cfg connectors.RuntimeConfig) (string, error) {
	return requireIntConfig(cfg, "account")
}

// klausWorkspace returns the required workspace id from config.
func klausWorkspace(cfg connectors.RuntimeConfig) (string, error) {
	return requireIntConfig(cfg, "workspace")
}

// requireIntConfig reads a required integer-valued config field and returns its
// canonical string form (Klaus path segments are integers).
func requireIntConfig(cfg connectors.RuntimeConfig, key string) (string, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return "", fmt.Errorf("klaus-api connector requires config %s", key)
	}
	if _, err := strconv.Atoi(raw); err != nil {
		return "", fmt.Errorf("klaus-api config %s must be an integer: %w", key, err)
	}
	return raw, nil
}

// klausBaseURL resolves and validates the base URL. The default is
// kibbles.klausapp.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func klausBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return klausDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("klaus-api config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("klaus-api config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("klaus-api config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// klausWindowStart resolves the lower bound for the reviews date window: the
// incremental cursor (if any) else the start_date config else a default of one
// window before end.
func klausWindowStart(req connectors.ReadRequest) (time.Time, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		t, err := parseKlausTime(cursor)
		if err != nil {
			return time.Time{}, fmt.Errorf("klaus-api state cursor invalid: %w", err)
		}
		return t, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		// No bound configured: read just the trailing window up to now.
		return time.Now().UTC().Add(-klausWindow), nil
	}
	t, err := parseKlausTime(startDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("klaus-api config start_date must be RFC3339 (YYYY-MM-DDThh:mm:ssZ): %w", err)
	}
	return t, nil
}

// klausWindowEnd resolves the upper bound for the reviews date window: the
// end_date config (mainly for deterministic tests) else now.
func klausWindowEnd(req connectors.ReadRequest) (time.Time, error) {
	endDate := strings.TrimSpace(req.Config.Config["end_date"])
	if endDate == "" {
		return time.Now().UTC(), nil
	}
	t, err := parseKlausTime(endDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("klaus-api config end_date must be RFC3339 (YYYY-MM-DDThh:mm:ssZ): %w", err)
	}
	return t, nil
}

func parseKlausTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if t, err := time.Parse(klausTimeFormat, value); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Klaus is read-only, so
// writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
