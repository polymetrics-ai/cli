// Package merge implements the native pm Merge connector. Merge
// (https://merge.dev) is a unified API that normalizes data across many
// third-party platforms into common-model objects; this connector reads the
// Merge ATS (Applicant Tracking System) category over its REST API.
//
// It is a declarative-HTTP per-system connector built on the stripe template:
// a thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + cursor pagination) with Merge-specific stream
// definitions and endpoints. Merge uses dual-token auth: an account-wide API
// token in Authorization: Bearer, plus a per-linked-account X-Account-Token
// header selecting which connected end-customer account to read.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The connector is read-only:
// Merge's write surface is narrow and not a general reverse-ETL target, so
// Capabilities.Write is false.
package merge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	// mergeDefaultBaseURL is the Merge ATS Common Model v1 base. Override via the
	// base_url config for other Merge categories (hris, accounting, ...) or test
	// servers.
	mergeDefaultBaseURL  = "https://api.merge.dev/api/ats/v1"
	mergeDefaultPageSize = 100
	mergeMaxPageSize     = 100
	mergeUserAgent       = "polymetrics-go-cli"
	// mergeFixtureModified is the deterministic modified_at timestamp used by the
	// fixture-mode records.
	mergeFixtureModified = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("merge", New)
}

// New returns the Merge connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Merge connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "merge" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "merge",
		DisplayName:     "Merge",
		IntegrationType: "api",
		Description:     "Reads Merge ATS common-model objects (candidates, applications, jobs, offers, departments, users) through the Merge unified REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Merge. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mergeBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mergeAPIToken(cfg)) == "" {
		return errors.New("merge connector requires secret api_token")
	}
	if strings.TrimSpace(mergeAccountToken(cfg)) == "" {
		return errors.New("merge connector requires secret account_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the candidates list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "candidates", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check merge: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mergeStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Merge stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "candidates"
	}
	endpoint, ok := mergeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("merge stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mergePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mergeMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Base query: page size plus an optional modified_after lower bound derived
	// from the incremental cursor or start_date config.
	base := url.Values{}
	base.Set("page_size", strconv.Itoa(pageSize))
	lower, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	if lower != "" {
		base.Set("modified_after", lower)
	}

	// Merge cursor pagination: each list response is
	// {next:<cursor>|null, previous:..., results:[...]}; the next page is
	// requested with ?cursor=<next>. connsdk's CursorPaginator reads the body
	// token at "next" and stops when it is null/empty.
	paginator := &connsdk.CursorPaginator{
		CursorParam: "cursor",
		TokenPath:   "next",
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "results", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	}); err != nil {
		return fmt.Errorf("read merge %s: %w", endpoint.resource, err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. The Merge connector is
// read-only (Capabilities.Write is false), so any write request is rejected
// rather than silently no-op'd.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, fmt.Errorf("merge connector is read-only: %w", connectors.ErrUnsupportedOperation)
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise merge credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"remote_id":           fmt.Sprintf("remote_%d", i),
			"first_name":          fmt.Sprintf("Fixture%d", i),
			"last_name":           "Example",
			"name":                fmt.Sprintf("Fixture %s %d", endpoint.resource, i),
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"company":             "Example Inc",
			"title":               "Engineer",
			"status":              "OPEN",
			"type":                "POSTING",
			"source":              "LinkedIn",
			"candidate":           "candidates_fixture_1",
			"job":                 "jobs_fixture_1",
			"application":         "applications_fixture_1",
			"access_role":         "ADMIN",
			"is_private":          false,
			"can_email":           true,
			"confidential":        false,
			"disabled":            false,
			"remote_created_at":   mergeFixtureModified,
			"remote_updated_at":   mergeFixtureModified,
			"applied_at":          mergeFixtureModified,
			"start_date":          mergeFixtureModified,
			"modified_at":         mergeFixtureModified,
			"remote_was_deleted":  false,
			"last_interaction_at": mergeFixtureModified,
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

// requester builds a connsdk.Requester wired with Bearer auth (the account-wide
// API token), the resolved base URL, and the per-linked-account X-Account-Token
// header. The secrets only ever flow into connsdk.Bearer / a static header; they
// are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mergeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiToken := mergeAPIToken(cfg)
	if strings.TrimSpace(apiToken) == "" {
		return nil, errors.New("merge connector requires secret api_token")
	}
	accountToken := mergeAccountToken(cfg)
	if strings.TrimSpace(accountToken) == "" {
		return nil, errors.New("merge connector requires secret account_token")
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.Bearer(apiToken),
		UserAgent:      mergeUserAgent,
		DefaultHeaders: map[string]string{"X-Account-Token": strings.TrimSpace(accountToken)},
	}, nil
}

// incrementalLowerBound returns the ISO-8601 lower bound for modified_after,
// derived from the incremental cursor (if any) or else the start_date config.
// An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, startDate); err != nil {
		return "", fmt.Errorf("merge config start_date must be RFC3339: %w", err)
	}
	return startDate, nil
}

func mergeAPIToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

func mergeAccountToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["account_token"]
}

// mergeBaseURL resolves and validates the base URL. The default is
// api.merge.dev; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func mergeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mergeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("merge config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("merge config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("merge config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mergePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mergeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("merge config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mergeMaxPageSize {
		return 0, fmt.Errorf("merge config page_size must be between 1 and %d", mergeMaxPageSize)
	}
	return value, nil
}

func mergeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("merge config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("merge config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
