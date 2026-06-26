// Package guru implements the native pm Guru connector. It is a declarative-HTTP
// per-system connector built on the connsdk toolkit: a connsdk.Requester wired
// with HTTP Basic auth (email + API token) reads Guru's REST list endpoints,
// which return top-level JSON arrays and advertise their next page via an RFC
// 5988 Link header. The connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package guru

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
	guruDefaultBaseURL  = "https://api.getguru.com/api/v1"
	guruDefaultPageSize = 50
	guruMaxPageSize     = 250
	guruUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("guru", New)
}

// New returns the Guru connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Guru connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "guru" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "guru",
		DisplayName:     "Guru",
		IntegrationType: "api",
		Description:     "Reads Guru collections, groups, members, and teams through the Guru REST API using HTTP Basic authentication (email + API token).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Guru. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := guruBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(guruUsername(cfg)) == "" {
		return errors.New("guru connector requires config username (email)")
	}
	if strings.TrimSpace(guruSecret(cfg)) == "" {
		return errors.New("guru connector requires secret password (API token)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the collections list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "collections", url.Values{"pageSize": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check guru: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Guru is a read-only
// source connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: guruStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "collections"
	}
	endpoint, ok := guruStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("guru stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := guruPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := guruMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("pageSize", strconv.Itoa(pageSize))
	// Guru pages results with an RFC 5988 Link header (rel="next"), which the
	// LinkHeaderPaginator follows. Records live at the top-level array root ("").
	paginator := &connsdk.LinkHeaderPaginator{}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	})
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise guru credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":               fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"slug":               fmt.Sprintf("fixture/%s-%d", endpoint.resource, i),
			"description":        "Deterministic fixture record (no network).",
			"color":              "#00b1a0",
			"publicCardsEnabled": false,
			"collectionType":     "INTERNAL",
			"dateCreated":        "2026-01-01T00:00:00.000Z",
			"modifiable":         true,
			"groupType":          "STANDARD",
			"memberCount":        int64(i),
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":          "Fixture",
			"lastName":           fmt.Sprintf("User%d", i),
			"status":             "ACTIVE",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The secret (API token) only ever flows into connsdk.Basic;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := guruBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(guruUsername(cfg))
	if username == "" {
		return nil, errors.New("guru connector requires config username (email)")
	}
	secret := guruSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("guru connector requires secret password (API token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: guruUserAgent,
	}, nil
}

func guruUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func guruSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// guruBaseURL resolves and validates the base URL. The default is
// api.getguru.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func guruBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return guruDefaultBaseURL, nil
	}
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return guruDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("guru config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("guru config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("guru config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func guruPageSize(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return guruDefaultPageSize, nil
	}
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return guruDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("guru config page_size must be an integer: %w", err)
	}
	if value < 1 || value > guruMaxPageSize {
		return 0, fmt.Errorf("guru config page_size must be between 1 and %d", guruMaxPageSize)
	}
	return value, nil
}

func guruMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("guru config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("guru config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
