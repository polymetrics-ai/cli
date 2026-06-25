// Package calcom implements the native pm Cal.com connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction) wired to Cal.com's v2 REST
// API. It mirrors the stripe reference connector's shape.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The directory and registry key are "cal-com"; the Go package identifier is
// "calcom" (hyphens are not valid in Go identifiers).
package calcom

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
	registryName          = "cal-com"
	displayName           = "Cal.com"
	defaultBaseURL        = "https://api.cal.com"
	defaultAPIVersion     = "2024-08-13"
	defaultPageSize       = 100
	maxPageSize           = 100
	userAgent             = "polymetrics-go-cli"
	secretField           = "api_key"
	fixtureRecordsPerPage = 2
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Cal.com connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Cal.com connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     displayName,
		IntegrationType: "api",
		Description:     "Reads Cal.com bookings, event types, schedules, and profile through the Cal.com v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Cal.com. In
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
	if strings.TrimSpace(secret(cfg)) == "" {
		return fmt.Errorf("%s connector requires secret %s", registryName, secretField)
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of /v2/me confirms auth and connectivity without mutating.
	if err := r.DoJSON(ctx, http.MethodGet, "v2/me", nil, nil, nil); err != nil {
		return fmt.Errorf("check %s: %w", registryName, err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: calcomStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "bookings"
	}
	endpoint, ok := calcomStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", registryName, stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSizeFromConfig(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesFromConfig(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Cal.com's offset (skip/take) pagination. A page shorter than
// take signals the end of the stream. Non-paginated streams (my_profile) and the
// nested event-types stream make a single request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "v2/" + endpoint.resource

	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read %s %s: %w", registryName, endpoint.resource, err)
		}
		_, err = c.emitResponse(ctx, endpoint, resp.Body, emit)
		return err
	}

	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("take", strconv.Itoa(pageSize))
		query.Set("skip", strconv.Itoa(skip))
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read %s %s: %w", registryName, endpoint.resource, err)
		}
		count, err := c.emitResponse(ctx, endpoint, resp.Body, emit)
		if err != nil {
			return err
		}
		// A short page (fewer than the requested take) ends pagination.
		if count < pageSize {
			return nil
		}
		skip += pageSize
	}
	return nil
}

// emitResponse extracts records from a response body for the given endpoint and
// emits each mapped record. It returns the number of records emitted so the
// pagination loop can detect a short page. The nested event-types stream is
// flattened from data.eventTypeGroups[].eventTypes[].
func (c Connector) emitResponse(ctx context.Context, endpoint streamEndpoint, body []byte, emit func(connectors.Record) error) (int, error) {
	if endpoint.nested {
		return emitNested(ctx, endpoint, body, emit)
	}
	records, err := connsdk.RecordsAt(body, endpoint.recordsPath)
	if err != nil {
		return 0, fmt.Errorf("decode %s %s page: %w", registryName, endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

// emitNested flattens the event-types envelope: data.eventTypeGroups is an array
// of groups, each holding an eventTypes array. Every nested event type becomes a
// record.
func emitNested(ctx context.Context, endpoint streamEndpoint, body []byte, emit func(connectors.Record) error) (int, error) {
	groups, err := connsdk.RecordsAt(body, endpoint.recordsPath)
	if err != nil {
		return 0, fmt.Errorf("decode %s %s groups: %w", registryName, endpoint.resource, err)
	}
	count := 0
	for _, group := range groups {
		raw, ok := group["eventTypes"].([]any)
		if !ok {
			continue
		}
		for _, item := range raw {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if err := ctx.Err(); err != nil {
				return count, err
			}
			if err := emit(endpoint.mapRecord(obj)); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= fixtureRecordsPerPage; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          int64(i),
			"uid":         fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"slug":        fmt.Sprintf("%s-%d", stream, i),
			"title":       fmt.Sprintf("Fixture %s %d", stream, i),
			"name":        fmt.Sprintf("Fixture %s %d", stream, i),
			"description": "fixture record",
			"status":      "accepted",
			"start":       "2026-01-01T00:00:00Z",
			"end":         "2026-01-01T00:30:00Z",
			"eventTypeId": int64(100 + i),
			"createdAt":   "2026-01-01T00:00:00Z",
			"updatedAt":   "2026-01-01T00:00:00Z",
			"length":      int64(30),
			"hidden":      false,
			"position":    int64(i),
			"timeZone":    "UTC",
			"isDefault":   i == 1,
			"ownerId":     int64(1),
			"username":    fmt.Sprintf("fixture%d", i),
			"email":       fmt.Sprintf("fixture+%d@example.com", i),
			"timeFormat":  int64(24),
			"weekStart":   "Monday",
			"connector":   registryName,
			"fixture":     true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the cal-api-version header. The secret only ever flows into
// connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("%s connector requires secret %s", registryName, secretField)
	}
	headers := map[string]string{
		"cal-api-version": apiVersion(cfg),
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.Bearer(token),
		UserAgent:      userAgent,
		DefaultHeaders: headers,
	}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[secretField]
}

// apiVersion resolves the cal-api-version header value, allowing a config
// override for forward compatibility.
func apiVersion(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["api_version"]); v != "" {
			return v
		}
	}
	return defaultAPIVersion
}

// baseURL resolves and validates the base URL. The default is api.cal.com; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", registryName, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", registryName, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New(registryName + " config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSizeFromConfig(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config page_size must be an integer: %w", registryName, err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("%s config page_size must be between 1 and %d", registryName, maxPageSize)
	}
	return value, nil
}

func maxPagesFromConfig(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", registryName, err)
	}
	if value < 0 {
		return 0, errors.New(registryName + " config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is not supported: Cal.com is exposed read-only for now. The method
// exists to satisfy the connectors.Connector interface.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
