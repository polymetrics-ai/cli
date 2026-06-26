// Package hibob implements the native pm HiBob connector. It follows the
// declarative-HTTP template established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction) with HiBob-specific stream definitions and endpoints.
//
// HiBob is a read-only HR data source. It authenticates with HTTP Basic auth
// using a service-user id (config "username") and a token (secret "password").
// The default base URL is https://api.hibob.com/v1; setting config is_sandbox
// switches to the sandbox host. Like stripe, it self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
package hibob

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
	hibobProdBaseURL     = "https://api.hibob.com/v1"
	hibobSandboxBaseURL  = "https://api.sandbox.hibob.com/v1"
	hibobDefaultPageSize = 50
	hibobMaxPageSize     = 250
	hibobUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("hibob", New)
}

// New returns the HiBob connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm HiBob connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "hibob" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "hibob",
		DisplayName:     "HiBob",
		IntegrationType: "api",
		Description:     "Reads HiBob HR data: employee profiles, company named lists, and people field definitions via the HiBob REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to HiBob. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := hibobBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(hibobUsername(cfg)) == "" {
		return errors.New("hibob connector requires config username")
	}
	if strings.TrimSpace(hibobSecret(cfg)) == "" {
		return errors.New("hibob connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of profiles confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "profiles", nil, nil, nil); err != nil {
		return fmt.Errorf("check hibob: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: hibobStreams()}, nil
}

// Write satisfies the connectors.Connector interface but HiBob is a read-only
// source: there are no approved reverse-ETL actions, so writes are rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "profiles"
	}
	endpoint, ok := hibobStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("hibob stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := hibobPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := hibobMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest reads an endpoint's records. Paginated endpoints (profiles) advance by
// offset/limit until a short page is returned; non-paginated metadata endpoints
// are read in a single request. RecordsAt selects the records array at the
// stream's recordsPath.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
		if err != nil {
			return fmt.Errorf("read hibob %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode hibob %s: %w", endpoint.resource, err)
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

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read hibob %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode hibob %s page: %w", endpoint.resource, err)
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
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise hibob credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"email":       fmt.Sprintf("fixture+%d@example.com", i),
			"displayName": fmt.Sprintf("Fixture %d", i),
			"firstName":   "Fixture",
			"surname":     fmt.Sprintf("%d", i),
			"fullName":    fmt.Sprintf("Fixture %d", i),
			"name":        fmt.Sprintf("Fixture List %d", i),
			"value":       fmt.Sprintf("value_%d", i),
			"category":    "about",
			"type":        "text",
			"description": "fixture field",
			"work": map[string]any{
				"title":      "Engineer",
				"department": "R&D",
				"site":       "Remote",
				"startDate":  "2026-01-01",
				"isManager":  false,
			},
			"personal": map[string]any{"pronouns": "they/them"},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "hibob"
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := hibobBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := hibobUsername(cfg)
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("hibob connector requires config username")
	}
	secret := hibobSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("hibob connector requires secret password")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: hibobUserAgent,
	}, nil
}

func hibobUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func hibobSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// hibobBaseURL resolves and validates the base URL. The default is
// api.hibob.com; setting is_sandbox=true switches to the sandbox host. Any
// explicit base_url override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func hibobBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if isSandbox(cfg) {
			return hibobSandboxBaseURL, nil
		}
		return hibobProdBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("hibob config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("hibob config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("hibob config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func isSandbox(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["is_sandbox"]), "true")
}

func hibobPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return hibobDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hibob config page_size must be an integer: %w", err)
	}
	if value < 1 || value > hibobMaxPageSize {
		return 0, fmt.Errorf("hibob config page_size must be between 1 and %d", hibobMaxPageSize)
	}
	return value, nil
}

func hibobMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hibob config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("hibob config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// nestedString returns item[key] as a string when it is a string, otherwise "".
// HiBob ids arrive as strings; this guards against numeric/null variants.
func nestedString(item map[string]any, key string) string {
	if item == nil {
		return ""
	}
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// mapAt returns item[key] as a map, or an empty map when absent/wrong type.
func mapAt(item map[string]any, key string) map[string]any {
	if item == nil {
		return map[string]any{}
	}
	if m, ok := item[key].(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
