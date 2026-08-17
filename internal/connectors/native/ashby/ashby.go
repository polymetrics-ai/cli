// Package ashby implements the native pm Ashby connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + Basic auth + RecordsAt extraction) with Ashby-specific stream
// definitions, endpoints, and its POST cursor-in-body pagination.
//
// Ashby is an applicant-tracking system. Its REST API lives at
// https://api.ashbyhq.com; every list endpoint is a POST to "<resource>/list"
// authenticated with HTTP Basic auth (the API key is the username, the password
// is blank) and returns {success, results:[...], moreDataAvailable, nextCursor}.
// The Ashby upstream source is read-only (full_refresh), so this connector is
// read-only too.
package ashby

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
	ashbyDefaultBaseURL  = "https://api.ashbyhq.com"
	ashbyDefaultPageSize = 100
	ashbyMaxPageSize     = 100
	ashbyUserAgent       = "polymetrics-go-cli"
	// ashbyAPIVersion is sent in the Accept header per Ashby's docs
	// (Accept: application/json; version=1).
	ashbyAccept = "application/json; version=1"
)

// New returns the Ashby connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Ashby connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "ashby" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "ashby",
		DisplayName:     "Ashby",
		IntegrationType: "api",
		Description:     "Reads Ashby applicant-tracking data — candidates, jobs, applications, and users — through the Ashby REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Ashby. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := ashbyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(ashbySecret(cfg)) == "" {
		return errors.New("ashby connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if _, err := r.Do(ctx, http.MethodPost, "user/list", nil, map[string]any{"limit": 1}); err != nil {
		return fmt.Errorf("check ashby: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: ashbyStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Ashby source is
// read-only (its upstream source supports full_refresh only and exposes no safe
// reverse-ETL writes), so writes are rejected. Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "candidates"
	}
	endpoint, ok := ashbyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("ashby stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := ashbyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := ashbyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Ashby's POST cursor-in-body pagination. List endpoints accept
// a JSON body {limit, cursor} and return {results:[...], moreDataAvailable,
// nextCursor}. The next page is requested by passing nextCursor as cursor until
// moreDataAvailable is false. connsdk's paginators are query-param based, so this
// body-cursor loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := map[string]any{"limit": pageSize}
		if cursor != "" {
			body["cursor"] = cursor
		}
		resp, err := r.Do(ctx, http.MethodPost, endpoint.resource, nil, body)
		if err != nil {
			return fmt.Errorf("read ashby %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode ashby %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		more, err := connsdk.StringAt(resp.Body, "moreDataAvailable")
		if err != nil {
			return fmt.Errorf("decode ashby %s moreDataAvailable: %w", endpoint.resource, err)
		}
		next, err := connsdk.StringAt(resp.Body, "nextCursor")
		if err != nil {
			return fmt.Errorf("decode ashby %s nextCursor: %w", endpoint.resource, err)
		}
		if more != "true" || strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise ashby credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(stream, "s"), i),
			"name":           fmt.Sprintf("Fixture %d", i),
			"title":          fmt.Sprintf("Fixture Title %d", i),
			"status":         "Active",
			"firstName":      fmt.Sprintf("Fixture%d", i),
			"lastName":       "Example",
			"email":          fmt.Sprintf("fixture+%d@example.com", i),
			"globalRole":     "Member",
			"isEnabled":      true,
			"candidateId":    "candidate_fixture_1",
			"jobId":          "job_fixture_1",
			"employmentType": "FullTime",
			"company":        "Example Inc",
			"createdAt":      "2026-01-01T00:00:00Z",
			"updatedAt":      fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "ashby"
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

// requester builds a connsdk.Requester wired with Basic auth (api key as
// username, blank password), the resolved base URL, and the Ashby Accept header.
// The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := ashbyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := ashbySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("ashby connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: ashbyUserAgent,
		Accept:    ashbyAccept,
	}, nil
}

func ashbySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// ashbyBaseURL resolves and validates the base URL. The default is
// api.ashbyhq.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func ashbyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return ashbyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("ashby config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("ashby config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("ashby config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func ashbyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return ashbyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ashby config page_size must be an integer: %w", err)
	}
	if value < 1 || value > ashbyMaxPageSize {
		return 0, fmt.Errorf("ashby config page_size must be between 1 and %d", ashbyMaxPageSize)
	}
	return value, nil
}

func ashbyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ashby config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("ashby config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
