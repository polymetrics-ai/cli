// Package appcues implements the native pm Appcues connector. It follows the
// stripe declarative-HTTP template: a thin package that composes the connsdk
// toolkit (Requester + Basic auth + RecordsAt extraction) with Appcues-specific
// stream definitions and endpoints.
//
// Appcues exposes a read-only REST API at https://api.appcues.com/v2. Resources
// are listed under accounts/{account_id}/{resource} and each list endpoint
// returns a top-level JSON array. Authentication is HTTP Basic, where the
// username is the API key (config) and the password is the API secret (secret).
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package appcues

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
	appcuesDefaultBaseURL  = "https://api.appcues.com/v2"
	appcuesDefaultPageSize = 100
	appcuesMaxPageSize     = 1000
	appcuesUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("appcues", New)
}

// New returns the Appcues connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Appcues connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "appcues" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "appcues",
		DisplayName:     "Appcues",
		IntegrationType: "api",
		Description:     "Reads Appcues flows, segments, tags, checklists, and banners through the Appcues REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Appcues. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := appcuesBaseURL(cfg); err != nil {
		return err
	}
	accountID := strings.TrimSpace(cfg.Config["account_id"])
	if accountID == "" {
		return errors.New("appcues connector requires config account_id")
	}
	if strings.TrimSpace(cfg.Config["username"]) == "" {
		return errors.New("appcues connector requires config username (API key)")
	}
	if strings.TrimSpace(appcuesSecret(cfg)) == "" {
		return errors.New("appcues connector requires secret password (API secret)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the flows list confirms auth and connectivity without
	// mutating anything.
	path := appcuesResourcePath(accountID, "flows")
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check appcues: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: appcuesStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "flows"
	}
	endpoint, ok := appcuesStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("appcues stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	accountID := strings.TrimSpace(req.Config.Config["account_id"])
	if accountID == "" {
		return errors.New("appcues connector requires config account_id")
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := appcuesPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := appcuesMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, accountID, endpoint, pageSize, maxPages, emit)
}

// harvest reads every page of an Appcues list endpoint. The Appcues v2 list
// endpoints return a top-level JSON array; large accounts page with the `page`
// query param (1-based) and a `limit`. We request pages until a short page (fewer
// than pageSize records) is returned, which mirrors connsdk.PageNumberPaginator
// semantics but lets us read the array at the response root via RecordsAt("").
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, accountID string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := appcuesResourcePath(accountID, endpoint.resource)
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("limit", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read appcues %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode appcues %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page means there are no more records to fetch.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise appcues credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":        fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"description": fmt.Sprintf("Fixture %s record %d", stream, i),
			"state":       "PUBLISHED",
			"published":   true,
			"createdAt":   "2026-01-01T00:00:00.000Z",
			"updatedAt":   fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i),
			"createdBy":   "user_fixture",
			"updatedBy":   "user_fixture",
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

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := appcuesBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(cfg.Config["username"])
	if username == "" {
		return nil, errors.New("appcues connector requires config username (API key)")
	}
	secret := appcuesSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("appcues connector requires secret password (API secret)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: appcuesUserAgent,
	}, nil
}

func appcuesResourcePath(accountID, resource string) string {
	return "accounts/" + url.PathEscape(accountID) + "/" + resource
}

func appcuesSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// appcuesBaseURL resolves and validates the base URL. The default is
// api.appcues.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func appcuesBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return appcuesDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("appcues config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("appcues config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("appcues config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func appcuesPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return appcuesDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("appcues config page_size must be an integer: %w", err)
	}
	if value < 1 || value > appcuesMaxPageSize {
		return 0, fmt.Errorf("appcues config page_size must be between 1 and %d", appcuesMaxPageSize)
	}
	return value, nil
}

func appcuesMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("appcues config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("appcues config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Appcues is a read-only
// source for pm, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
