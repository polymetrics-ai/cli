// Package missive implements the native pm Missive connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester +
// Bearer auth + RecordsAt extraction + offset pagination), following the stripe
// reference connector's shape.
//
// Missive's source is read-only (full-refresh): it lists contacts, contact
// groups, users, teams, and shared labels from the Missive REST API. Like the
// other per-system connectors it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package missive

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
	missiveDefaultBaseURL  = "https://public.missiveapp.com/v1"
	missiveDefaultPageSize = 50
	missiveMaxPageSize     = 200
	missiveUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("missive", New)
}

// New returns the Missive connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Missive connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "missive" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "missive",
		DisplayName:     "Missive",
		IntegrationType: "api",
		Description:     "Reads Missive contacts, contact groups, users, teams, and shared labels through the Missive REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Missive. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := missiveBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(missiveSecret(cfg)) == "" {
		return errors.New("missive connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check missive: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: missiveStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := missiveStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("missive stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := missivePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := missiveMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	if stream == "contact_groups" {
		if kind := strings.TrimSpace(req.Config.Config["kind"]); kind != "" {
			base.Set("kind", kind)
		}
	}
	return c.harvest(ctx, r, endpoint, base, pageSize, maxPages, emit)
}

// harvest drives Missive's offset pagination. List endpoints return
// {"<resource>":[...]}; the next page is requested with offset += limit until a
// short page (fewer than limit records) is returned. The loop lives here, built
// on connsdk.Requester + connsdk.RecordsAt, because the records key varies per
// stream.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read missive %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode missive %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested limit) marks the end.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise missive credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                     fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"first_name":             fmt.Sprintf("Fixture%d", i),
			"last_name":              "Example",
			"modified_at":            int64(1767225600 + i),
			"name":                   fmt.Sprintf("Fixture %d", i),
			"email":                  fmt.Sprintf("fixture+%d@example.com", i),
			"kind":                   "group",
			"organization":           "org_fixture_1",
			"name_with_parent_names": fmt.Sprintf("Parent/Fixture %d", i),
			"color":                  "#1d6fe0",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := missiveBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := missiveSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("missive connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: missiveUserAgent,
	}, nil
}

func missiveSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// missiveBaseURL resolves and validates the base URL. The default is
// public.missiveapp.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func missiveBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return missiveDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("missive config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("missive config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("missive config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// missivePageSize resolves the per-page record limit. Missive's config calls this
// "limit" (default 50). The value is bounded to keep requests sane.
func missivePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["limit"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return missiveDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("missive config limit must be an integer: %w", err)
	}
	if value < 1 || value > missiveMaxPageSize {
		return 0, fmt.Errorf("missive config limit must be between 1 and %d", missiveMaxPageSize)
	}
	return value, nil
}

func missiveMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("missive config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("missive config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// Write satisfies the connectors.Connector interface. The Missive source is
// read-only, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
