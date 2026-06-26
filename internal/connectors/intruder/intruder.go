// Package intruder implements the native pm Intruder connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package composing the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Intruder-specific stream definitions, endpoints, and
// offset/limit pagination.
//
// The Intruder API (https://api.intruder.io/v1) is full-refresh only: list
// endpoints return {"results":[...]} and are walked with offset/limit query
// parameters. Auth is a bearer access_token. The connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package intruder

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
	intruderDefaultBaseURL  = "https://api.intruder.io/v1"
	intruderDefaultPageSize = 100
	intruderMaxPageSize     = 100
	intruderUserAgent       = "polymetrics-go-cli"
	intruderResultsPath     = "results"
)

func init() {
	connectors.RegisterFactory("intruder", New)
}

// New returns the Intruder connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Intruder connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "intruder" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "intruder",
		DisplayName:     "Intruder",
		IntegrationType: "api",
		Description:     "Reads Intruder issues, issue occurrences, scans, and targets through the Intruder REST API (read-only, full refresh).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Intruder. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := intruderBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(intruderSecret(cfg)) == "" {
		return errors.New("intruder connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the targets list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "targets", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check intruder: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: intruderStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Intruder is read-only
// (full-refresh source), so writes are explicitly unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader. Intruder is full-refresh
// only, so streams start (and stay) with an empty incremental cursor.
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
		stream = "issues"
	}
	endpoint, ok := intruderStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("intruder stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := intruderPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := intruderMaxPages(req.Config)
	if err != nil {
		return err
	}

	if endpoint.substream {
		return c.readOccurrences(ctx, r, endpoint, pageSize, maxPages, emit)
	}
	return c.harvest(ctx, r, endpoint.resource, pageSize, maxPages, endpoint.mapRecord, emit)
}

// harvest drives Intruder's offset/limit pagination. List endpoints return
// {"results":[...]}; the next page is requested with offset += limit, stopping
// when a short page (fewer than pageSize records) is returned. connsdk's
// OffsetPaginator matches this shape, but the small in-package loop keeps the
// per-record mapper and the substream variant on one code path.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, pageSize, maxPages int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read intruder %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, intruderResultsPath)
		if err != nil {
			return fmt.Errorf("decode intruder %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
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

// readOccurrences reads the occurrences substream: it first lists every issue
// id, then reads /issues/{id}/occurrences for each, tagging every occurrence
// with its parent issue_id. This mirrors the Airbyte SubstreamPartitionRouter.
func (c Connector) readOccurrences(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	issueIDs, err := c.collectIssueIDs(ctx, r, pageSize, maxPages)
	if err != nil {
		return err
	}
	for _, issueID := range issueIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		resource := "issues/" + url.PathEscape(issueID) + "/occurrences"
		wrap := func(item map[string]any) connectors.Record {
			rec := endpoint.mapRecord(item)
			if rec["issue_id"] == nil {
				rec["issue_id"] = issueID
			}
			return rec
		}
		if err := c.harvest(ctx, r, resource, pageSize, maxPages, wrap, emit); err != nil {
			return err
		}
	}
	return nil
}

// collectIssueIDs lists the issue ids that seed the occurrences substream.
func (c Connector) collectIssueIDs(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int) ([]string, error) {
	var ids []string
	err := c.harvest(ctx, r, "issues", pageSize, maxPages, intruderIssueRecord, func(rec connectors.Record) error {
		if id := stringField(rec, "id"); id != "" {
			ids = append(ids, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise intruder credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            int64(i),
			"issue_id":      int64(1),
			"title":         fmt.Sprintf("Fixture issue %d", i),
			"description":   "Fixture description",
			"severity":      "high",
			"remediation":   "Fixture remediation",
			"occurrences":   "1",
			"status":        "completed",
			"created_at":    "2026-01-01T00:00:00Z",
			"address":       fmt.Sprintf("host-%d.example.com", i),
			"tags":          []any{"web"},
			"target":        fmt.Sprintf("host-%d.example.com", i),
			"port":          int64(443),
			"age":           "1 day",
			"extra_info":    map[string]any{"fixture": true},
			"snoozed":       false,
			"snooze_reason": "",
			"snooze_until":  "",
			"connector":     "intruder",
			"fixture":       true,
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
	base, err := intruderBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := intruderSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("intruder connector requires secret access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: intruderUserAgent,
	}, nil
}

func intruderSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// intruderBaseURL resolves and validates the base URL. The default is
// api.intruder.io/v1; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func intruderBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return intruderDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("intruder config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("intruder config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("intruder config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intruderPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return intruderDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("intruder config page_size must be an integer: %w", err)
	}
	if value < 1 || value > intruderMaxPageSize {
		return 0, fmt.Errorf("intruder config page_size must be between 1 and %d", intruderMaxPageSize)
	}
	return value, nil
}

func intruderMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("intruder config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("intruder config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
