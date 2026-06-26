// Package devinai implements the native pm Devin AI connector. It is a
// declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction + cursor state) with
// Devin-specific stream definitions and the Devin v3 REST API.
//
// It follows the stripe reference connector shape and self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
//
// The Devin v3 API exposes organization-scoped list endpoints under
// /v3/organizations/{org_id}/... that return {items:[...], has_next_page,
// end_cursor} and paginate with an `after` cursor. Auth is a Bearer token
// (service-user API keys prefixed cog_). This connector is read-only: Devin has
// no obviously-safe reverse-ETL write surface, so Capabilities.Write is false.
package devinai

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
	devinDefaultBaseURL  = "https://api.devin.ai"
	devinDefaultPageSize = 100
	devinMaxPageSize     = 200
	devinUserAgent       = "polymetrics-go-cli"
	// devinFixtureCreated is the deterministic created_at used by fixture records.
	devinFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("devin-ai", New)
}

// New returns the Devin AI connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Devin AI connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "devin-ai" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "devin-ai",
		DisplayName:     "Devin AI",
		IntegrationType: "api",
		Description:     "Reads Devin AI sessions, session insights, session messages, playbooks, and secret metadata through the Devin v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Devin. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := devinBaseURL(cfg); err != nil {
		return err
	}
	orgID, err := devinOrgID(cfg)
	if err != nil {
		return err
	}
	if strings.TrimSpace(devinSecret(cfg)) == "" {
		return errors.New("devin-ai connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the sessions list confirms auth and connectivity.
	path := devinPath(orgID, "sessions")
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"first": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check devin-ai: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: devinStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Devin connector is
// read-only (Capabilities.Write is false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Devin stream starts with an
// empty incremental cursor (full sync); the start_date config can raise it.
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
		stream = "sessions"
	}
	endpoint, ok := devinStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("devin-ai stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	if _, err := devinBaseURL(req.Config); err != nil {
		return err
	}
	orgID, err := devinOrgID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := devinPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := devinMaxPages(req.Config)
	if err != nil {
		return err
	}
	createdAfter := incrementalLowerBound(req)
	return c.harvest(ctx, r, devinPath(orgID, endpoint.resource), endpoint, pageSize, maxPages, createdAfter, emit)
}

// harvest drives Devin's cursor pagination. Devin lists return
// {items:[...], has_next_page:bool, end_cursor:string}; the next page is
// requested with after=<end_cursor>. The loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, createdAfter string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("first", strconv.Itoa(pageSize))
	if createdAfter != "" {
		base.Set("created_after", createdAfter)
	}

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read devin-ai %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode devin-ai %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasNext, err := connsdk.StringAt(resp.Body, "has_next_page")
		if err != nil {
			return fmt.Errorf("decode devin-ai %s has_next_page: %w", endpoint.resource, err)
		}
		endCursor, err := connsdk.StringAt(resp.Body, "end_cursor")
		if err != nil {
			return fmt.Errorf("decode devin-ai %s end_cursor: %w", endpoint.resource, err)
		}
		if hasNext != "true" || strings.TrimSpace(endCursor) == "" {
			return nil
		}
		after = endCursor
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise devin-ai credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idField := endpoint.idField
		if idField == "" {
			idField = "id"
		}
		item := map[string]any{
			idField:         fmt.Sprintf("%s_fixture_%d", strings.ReplaceAll(endpoint.resource, "/", "_"), i),
			"session_id":    fmt.Sprintf("devin_session_fixture_%d", i),
			"title":         fmt.Sprintf("Fixture session %d", i),
			"status":        "finished",
			"category":      "default",
			"origin":        "api",
			"user_id":       "usr_fixture_1",
			"playbook_id":   "playbook_fixture_1",
			"is_archived":   false,
			"message_count": i,
			"summary":       fmt.Sprintf("Fixture summary %d", i),
			"role":          "assistant",
			"type":          "message",
			"content":       fmt.Sprintf("Fixture content %d", i),
			"name":          fmt.Sprintf("Fixture %d", i),
			"description":   "Fixture description",
			"created_at":    devinFixtureCreated,
			"updated_at":    devinFixtureCreated,
			"connector":     "devin-ai",
			"fixture":       true,
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := devinBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := devinSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("devin-ai connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: devinUserAgent,
	}, nil
}

// devinPath builds the org-scoped resource path for an endpoint, escaping the
// org id segment.
func devinPath(orgID, resource string) string {
	return "v3/organizations/" + url.PathEscape(orgID) + "/" + resource
}

// incrementalLowerBound returns the created_after lower bound, derived from the
// incremental cursor (if any) or else the start_date config. Empty means no
// lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func devinSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// devinOrgID resolves the required org_id config field.
func devinOrgID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("devin-ai connector requires config org_id")
	}
	org := strings.TrimSpace(cfg.Config["org_id"])
	if org == "" {
		return "", errors.New("devin-ai connector requires config org_id")
	}
	return org, nil
}

// devinBaseURL resolves and validates the base URL. The default is api.devin.ai;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func devinBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return devinDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("devin-ai config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("devin-ai config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("devin-ai config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func devinPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return devinDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("devin-ai config page_size must be an integer: %w", err)
	}
	if value < 1 || value > devinMaxPageSize {
		return 0, fmt.Errorf("devin-ai config page_size must be between 1 and %d", devinMaxPageSize)
	}
	return value, nil
}

func devinMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("devin-ai config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("devin-ai config max_pages must be 0 for unlimited or a positive integer")
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
