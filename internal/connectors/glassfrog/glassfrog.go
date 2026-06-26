// Package glassfrog implements the native pm GlassFrog connector. It is a
// declarative-HTTP per-system connector built in the shape of the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + API-key header auth + nested record extraction) with
// GlassFrog-specific stream definitions and endpoints.
//
// The GlassFrog API v3 (https://api.glassfrog.com/api/v3) authenticates with an
// X-Auth-Token header, supports only full-refresh reads, and nests each list
// under a key named after the resource (e.g. {"circles":[...]}). It is a
// read-only source: Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package glassfrog

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
	glassfrogDefaultBaseURL  = "https://api.glassfrog.com/api/v3"
	glassfrogDefaultPageSize = 100
	glassfrogMaxPageSize     = 100
	glassfrogUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("glassfrog", New)
}

// New returns the GlassFrog connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm GlassFrog connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "glassfrog" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "glassfrog",
		DisplayName:     "GlassFrog",
		IntegrationType: "api",
		Description:     "Reads GlassFrog circles, roles, people, projects, and assignments through the GlassFrog API v3 (read-only full-refresh source).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to GlassFrog. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := glassfrogBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(glassfrogSecret(cfg)) == "" {
		return errors.New("glassfrog connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the circles list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "circles", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check glassfrog: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: glassfrogStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "circles"
	}
	endpoint, ok := glassfrogStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("glassfrog stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := glassfrogPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := glassfrogMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives page-number pagination. GlassFrog v3 returns each list nested
// under a resource-named key (e.g. {"circles":[...]}); it accepts page/per_page
// query params. We advance the page number until a short (or empty) page is
// returned. The loop lives here, built on connsdk.Requester + connsdk.RecordsAt,
// because GlassFrog's nested-key shape is not a one-line connsdk paginator.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("per_page", strconv.Itoa(pageSize))
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read glassfrog %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode glassfrog %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page means we have reached the last page. An empty page also
		// terminates. This avoids an extra request when the API does not paginate.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise glassfrog credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                              i,
			"name":                            fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"short_name":                      fmt.Sprintf("FX%d", i),
			"organization_id":                 7,
			"strategy":                        "fixture strategy",
			"email":                           fmt.Sprintf("fixture+%d@example.com", i),
			"external_id":                     fmt.Sprintf("ext_%d", i),
			"tag_names":                       []any{"fixture"},
			"description":                     "fixture project",
			"status":                          "Current",
			"effort":                          "1",
			"value":                           "high",
			"roi":                             "5",
			"private_to_circle":               false,
			"link":                            "https://app.glassfrog.com/fixture",
			"created_at":                      "2026-01-01T00:00:00Z",
			"archived_at":                     nil,
			"waiting_on_who":                  "",
			"waiting_on_what":                 "",
			"purpose":                         "fixture purpose",
			"is_core":                         false,
			"elected_until":                   nil,
			"name_with_circle_for_core_roles": fmt.Sprintf("Role %d", i),
			"election":                        nil,
			"exclude_from_meetings":           false,
			"focus":                           "fixture focus",
			"links": map[string]any{
				"person":         100 + i,
				"role":           200 + i,
				"supported_role": 300 + i,
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-Auth-Token API-key header
// auth and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := glassfrogBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := glassfrogSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("glassfrog connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Auth-Token", secret, ""),
		UserAgent: glassfrogUserAgent,
	}, nil
}

func glassfrogSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// glassfrogBaseURL resolves and validates the base URL. The default is
// api.glassfrog.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func glassfrogBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return glassfrogDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("glassfrog config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("glassfrog config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("glassfrog config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func glassfrogPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return glassfrogDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("glassfrog config page_size must be an integer: %w", err)
	}
	if value < 1 || value > glassfrogMaxPageSize {
		return 0, fmt.Errorf("glassfrog config page_size must be between 1 and %d", glassfrogMaxPageSize)
	}
	return value, nil
}

func glassfrogMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("glassfrog config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("glassfrog config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. GlassFrog is a read-only
// source, so write is unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
