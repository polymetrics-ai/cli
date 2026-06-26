// Package gologin implements the native pm GoLogin source connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction + page-number pagination) with GoLogin-specific
// stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// GoLogin API: https://api.gologin.com, Bearer (api_key) auth, page-number
// pagination on the profiles list. The connector is read-only (the GoLogin API
// has no obvious safe reverse-ETL writes).
package gologin

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
	gologinDefaultBaseURL  = "https://api.gologin.com"
	gologinDefaultPageSize = 30
	gologinMaxPageSize     = 100
	gologinUserAgent       = "polymetrics-go-cli"
	gologinDefaultStream   = "profiles"
)

func init() {
	connectors.RegisterFactory("gologin", New)
}

// New returns the GoLogin connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm GoLogin source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gologin" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gologin",
		DisplayName:     "GoLogin",
		IntegrationType: "api",
		Description:     "Reads GoLogin browser profiles, folders, tags, and account information through the GoLogin REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to GoLogin. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gologinBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gologinSecret(cfg)) == "" {
		return errors.New("gologin connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the profiles list confirms auth and connectivity.
	q := url.Values{"page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "browser/v2", q, nil, nil); err != nil {
		return fmt.Errorf("check gologin: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gologinStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a GoLogin stream starts with
// an empty incremental cursor (full sync).
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
		stream = gologinDefaultStream
	}
	endpoint, ok := gologinStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gologin stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := gologinPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gologinMaxPages(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.paginated {
		return c.readSinglePage(ctx, r, endpoint, emit)
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// readSinglePage reads a non-paginated endpoint (folders, user, tags) in one
// request and emits the records found at the endpoint's record selector path.
func (c Connector) readSinglePage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read gologin %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode gologin %s: %w", endpoint.resource, err)
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

// harvest drives GoLogin's page-number pagination over the profiles list. The
// list returns {"profiles":[...]}; pages advance via ?page=N (1-based) until a
// short page (fewer than pageSize records) is returned. A page-number paginator
// fits this shape, so the loop is built on the connsdk PageNumberPaginator
// semantics inline (the records live under a selector path, so we read them with
// RecordsAt and stop on a short page).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read gologin %s page %d: %w", endpoint.resource, page, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode gologin %s page %d: %w", endpoint.resource, page, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short (or empty) page means we have reached the end.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gologin credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"_id":           fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":          fmt.Sprintf("Fixture %s %d", stream, i),
			"title":         fmt.Sprintf("Fixture tag %d", i),
			"notes":         "fixture record",
			"role":          "owner",
			"os":            "win",
			"browserType":   "chrome",
			"folderName":    "Fixtures",
			"color":         "#00aaff",
			"field":         "custom",
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":     "Fixture",
			"lastName":      fmt.Sprintf("User %d", i),
			"plan":          "free",
			"profilesCount": int64(i),
			"createdAt":     "2026-01-01T00:00:00Z",
			"updatedAt":     fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
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
	base, err := gologinBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := gologinSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("gologin connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: gologinUserAgent,
	}, nil
}

func gologinSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// gologinBaseURL resolves and validates the base URL. The default is
// api.gologin.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func gologinBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gologinDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gologin config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gologin config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gologin config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gologinPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gologinDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gologin config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gologinMaxPageSize {
		return 0, fmt.Errorf("gologin config page_size must be between 1 and %d", gologinMaxPageSize)
	}
	return value, nil
}

func gologinMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gologin config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gologin config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. GoLogin is read-only in pm,
// so writes are rejected as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
