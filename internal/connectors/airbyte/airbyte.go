// Package airbyte implements the native pm Airbyte connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference connector: a thin package that composes a connsdk
// Requester with OAuth2 client-credentials auth, offset/limit pagination, and
// data[] extraction over the Airbyte public API.
//
// The Airbyte public API authenticates with an OAuth2 client-credentials grant
// (client_id + client_secret -> short-lived bearer token) and serves list
// endpoints (workspaces, connections, sources, destinations, jobs) paginated by
// limit/offset with records under "data".
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package airbyte

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
	airbyteDefaultBaseURL  = "https://api.airbyte.com/v1"
	airbyteDefaultTokenURL = "https://api.airbyte.com/v1/applications/token"
	airbyteDefaultPageSize = 100
	airbyteMaxPageSize     = 100
	airbyteUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("airbyte", New)
}

// New returns the Airbyte connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Airbyte connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// (and the OAuth2 token exchange). Left nil in production; injectable for
	// tests.
	Client *http.Client
}

func (Connector) Name() string { return "airbyte" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "airbyte",
		DisplayName:     "Airbyte",
		IntegrationType: "api",
		Description:     "Reads Airbyte workspaces, connections, sources, destinations, and jobs from the Airbyte public API using an OAuth2 client-credentials grant.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the Airbyte
// API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := airbyteBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(airbyteClientID(cfg)) == "" {
		return errors.New("airbyte connector requires config client_id")
	}
	if strings.TrimSpace(airbyteSecret(cfg)) == "" {
		return errors.New("airbyte connector requires secret client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the workspaces list confirms the token exchange and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "workspaces", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check airbyte: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: airbyteStreams()}, nil
}

// Write is unsupported: the Airbyte public API source is read-only for the
// reverse-ETL pipeline. The method exists to satisfy connectors.Connector;
// Metadata reports Write=false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "connections"
	}
	endpoint, ok := airbyteStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("airbyte stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := airbytePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := airbyteMaxPages(req.Config)
	if err != nil {
		return err
	}

	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    pageSize,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, paginator, "data", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise airbyte credential-free (mirrors the stripe
// connector's fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			"workspaceId":     "ws_fixture_1",
			"connectionId":    id,
			"sourceId":        "src_fixture_1",
			"destinationId":   "dst_fixture_1",
			"definitionId":    "def_fixture_1",
			"name":            fmt.Sprintf("Fixture %d", i),
			"status":          "active",
			"jobType":         "sync",
			"sourceType":      "postgres",
			"destinationType": "snowflake",
			"dataResidency":   "auto",
			"namespaceFormat": "${SOURCE_NAMESPACE}",
			"prefix":          "",
			"jobId":           int64(1000 + i),
			"startTime":       "2026-01-01T00:00:00Z",
			"lastUpdatedAt":   "2026-01-01T00:00:00Z",
			"duration":        "PT1M",
			"bytesSynced":     int64(1024 * i),
			"rowsSynced":      int64(10 * i),
			"schedule":        map[string]any{"scheduleType": "manual"},
		}
		// Override the per-resource id so the primary key is populated correctly.
		switch stream {
		case "workspaces":
			item["workspaceId"] = id
		case "sources":
			item["sourceId"] = id
		case "destinations":
			item["destinationId"] = id
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

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved base URL. The client_secret only ever flows into the
// OAuth2ClientCredentials authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := airbyteBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID := strings.TrimSpace(airbyteClientID(cfg))
	if clientID == "" {
		return nil, errors.New("airbyte connector requires config client_id")
	}
	secret := strings.TrimSpace(airbyteSecret(cfg))
	if secret == "" {
		return nil, errors.New("airbyte connector requires secret client_secret")
	}
	tokenURL, err := airbyteTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: secret,
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: airbyteUserAgent,
	}, nil
}

func airbyteSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

func airbyteClientID(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config["client_id"])
}

// airbyteBaseURL resolves and validates the base URL. The default is
// api.airbyte.com; a self-managed deployment can set host or a full base_url
// override. Any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func airbyteBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config != nil {
		if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
			return validateURL(base, "base_url")
		}
		// host is the documented self-managed deployment field (e.g.
		// airbyte.mydomain.com); we build the public API path under it.
		if host := strings.TrimSpace(cfg.Config["host"]); host != "" {
			candidate := host
			if !strings.HasPrefix(candidate, "http://") && !strings.HasPrefix(candidate, "https://") {
				candidate = "https://" + candidate
			}
			validated, err := validateURL(candidate, "host")
			if err != nil {
				return "", err
			}
			return strings.TrimRight(validated, "/") + "/api/public/v1", nil
		}
	}
	return airbyteDefaultBaseURL, nil
}

// airbyteTokenURL resolves the OAuth2 token endpoint. It defaults to the public
// API token endpoint, can be overridden directly, or is derived from a
// self-managed host.
func airbyteTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config != nil {
		if tok := strings.TrimSpace(cfg.Config["token_url"]); tok != "" {
			return validateURL(tok, "token_url")
		}
	}
	base, err := airbyteBaseURL(cfg)
	if err != nil {
		return "", err
	}
	if base == airbyteDefaultBaseURL {
		return airbyteDefaultTokenURL, nil
	}
	return strings.TrimRight(base, "/") + "/applications/token", nil
}

// validateURL bounds SSRF risk: an override must be an absolute http/https URL
// with a host.
func validateURL(raw, field string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("airbyte config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("airbyte config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("airbyte config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func airbytePageSize(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return airbyteDefaultPageSize, nil
	}
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return airbyteDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("airbyte config page_size must be an integer: %w", err)
	}
	if value < 1 || value > airbyteMaxPageSize {
		return 0, fmt.Errorf("airbyte config page_size must be between 1 and %d", airbyteMaxPageSize)
	}
	return value, nil
}

func airbyteMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("airbyte config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("airbyte config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
