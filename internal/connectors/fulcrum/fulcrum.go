// Package fulcrum implements the native pm Fulcrum connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + X-ApiToken header auth
// + RecordsAt extraction) with Fulcrum-specific stream definitions, endpoints,
// and page-number pagination.
//
// Fulcrum (https://docs.fulcrumapp.com/reference) is a mobile data-collection
// platform. Its REST API v2 is read-oriented here: forms, records, projects,
// choice_lists, and classification_sets. Like stripe, it self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
package fulcrum

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
	fulcrumDefaultBaseURL  = "https://api.fulcrumapp.com/api/v2"
	fulcrumDefaultPageSize = 100
	fulcrumMaxPageSize     = 20000
	fulcrumUserAgent       = "polymetrics-go-cli"
	// fulcrumAuthHeader is the Fulcrum REST API token header. The api_key secret
	// is the value; it is never logged.
	fulcrumAuthHeader = "X-ApiToken"
	// fulcrumFixtureUpdated is the deterministic updated_at timestamp used by the
	// fixture-mode records.
	fulcrumFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("fulcrum", New)
}

// New returns the Fulcrum connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Fulcrum connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "fulcrum" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "fulcrum",
		DisplayName:     "Fulcrum",
		IntegrationType: "api",
		Description:     "Reads Fulcrum forms, records, projects, choice lists, and classification sets through the Fulcrum REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Fulcrum. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := fulcrumBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(fulcrumSecret(cfg)) == "" {
		return errors.New("fulcrum connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the forms list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "forms.json", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check fulcrum: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Fulcrum is exposed
// read-only here (Capabilities.Write=false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: fulcrumStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Fulcrum stream starts with
// an empty incremental cursor (full sync). Fulcrum supports only full_refresh in
// the upstream catalog, so the cursor is informational.
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
		stream = "forms"
	}
	endpoint, ok := fulcrumStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("fulcrum stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := fulcrumPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := fulcrumMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Fulcrum's page-number pagination. List responses carry
// current_page/total_pages alongside the records array; the loop requests
// page=1,2,... until current_page reaches total_pages (or a page comes back
// empty). connsdk has a PageNumberPaginator, but it stops on a short page, which
// is unreliable when the final page happens to be full; reading the reported
// total_pages is the authoritative stop condition for this API.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource + ".json"
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read fulcrum %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.resource)
		if err != nil {
			return fmt.Errorf("decode fulcrum %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) == 0 {
			return nil
		}
		current := pageNumber(resp.Body, "current_page", page)
		total := pageNumber(resp.Body, "total_pages", page)
		if current >= total {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise fulcrum credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":         fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"description":  "Deterministic fixture record.",
			"status":       "active",
			"record_count": int64(i),
			"form_id":      "form_fixture_1",
			"project_id":   "project_fixture_1",
			"latitude":     45.0 + float64(i),
			"longitude":    -122.0 - float64(i),
			"created_by":   "fixture-user",
			"updated_by":   "fixture-user",
			"auto_assign":  false,
			"created_at":   fulcrumFixtureUpdated,
			"updated_at":   fulcrumFixtureUpdated,
			"connector":    "fulcrum",
			"fixture":      true,
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

// requester builds a connsdk.Requester wired with X-ApiToken header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := fulcrumBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := fulcrumSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("fulcrum connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(fulcrumAuthHeader, secret, ""),
		UserAgent: fulcrumUserAgent,
	}, nil
}

func fulcrumSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// fulcrumBaseURL resolves and validates the base URL. The default is
// api.fulcrumapp.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func fulcrumBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return fulcrumDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("fulcrum config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("fulcrum config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("fulcrum config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fulcrumPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return fulcrumDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fulcrum config page_size must be an integer: %w", err)
	}
	if value < 1 || value > fulcrumMaxPageSize {
		return 0, fmt.Errorf("fulcrum config page_size must be between 1 and %d", fulcrumMaxPageSize)
	}
	return value, nil
}

func fulcrumMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fulcrum config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("fulcrum config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// pageNumber reads an integer pagination field (e.g. total_pages) from the
// response body, falling back to def when the field is missing or unparsable.
func pageNumber(body []byte, key string, def int) int {
	raw, err := connsdk.StringAt(body, key)
	if err != nil {
		return def
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return value
}
