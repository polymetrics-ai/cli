// Package justsift implements the native pm JustSift (Sift people directory)
// connector. It follows the declarative-HTTP per-system template established by
// the stripe connector: a thin package that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction) with JustSift-specific stream
// definitions, endpoints, and pagination.
//
// The JustSift API (https://api.justsift.com/v1) is read-only for this
// connector: it exposes people directory profiles and person field definitions.
// Auth is a Bearer api_token. The connector self-registers with the connectors
// registry via RegisterFactory in init(); the registryset package blank-imports
// this package in the production binary to run that side effect.
package justsift

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
	registryName    = "just-sift"
	defaultBaseURL  = "https://api.justsift.com/v1"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
	recordsPath     = "data"
	nextLinkPath    = "links.next"
	pageParam       = "page"
	pageSizeParam   = "limit"
	cursorParam     = "cursor"
	hardPageCeiling = 10000
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the JustSift connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm JustSift connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "JustSift",
		IntegrationType: "api",
		Description:     "Reads JustSift people directory profiles and person field definitions through the Sift REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to JustSift. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("just-sift connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the person field definitions confirms auth and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "fields/person", nil, nil, nil); err != nil {
		return fmt.Errorf("check just-sift: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// Write satisfies connectors.Connector. The JustSift API surface used here is
// read-only, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "peoples"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("just-sift stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSizeOf(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesOf(req.Config)
	if err != nil {
		return err
	}

	switch endpoint.pagination {
	case linkCursor:
		return c.harvestCursor(ctx, r, endpoint, pageSize, maxPages, emit)
	default:
		return c.harvestPageIncrement(ctx, r, endpoint, pageSize, maxPages, emit)
	}
}

// harvestPageIncrement drives JustSift's 1-based page-number pagination
// (/search/people). It requests pages until a page returns fewer than pageSize
// records (or zero), which signals the last page.
func (c Connector) harvestPageIncrement(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}
		if page > hardPageCeiling {
			return nil
		}
		query := url.Values{}
		query.Set(pageParam, strconv.Itoa(page))
		query.Set(pageSizeParam, strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read just-sift %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode just-sift %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
}

// harvestCursor drives JustSift's links.next cursor pagination (/fields/person).
// The next-link token is read from the response body at links.next and supplied
// as the cursor query parameter on the following request; an absent token ends
// the walk.
func (c Connector) harvestCursor(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set(pageSizeParam, strconv.Itoa(pageSize))

	next := ""
	for page := 0; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page >= maxPages {
			return nil
		}
		if page > hardPageCeiling {
			return nil
		}
		query := cloneValues(base)
		applyNextCursor(query, next)

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read just-sift %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode just-sift %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		token, err := connsdk.StringAt(resp.Body, nextLinkPath)
		if err != nil {
			return fmt.Errorf("decode just-sift %s next link: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(token)
		if next == "" {
			return nil
		}
	}
}

// applyNextCursor merges a links.next token into the request query. JustSift
// returns the token either as a bare cursor value or as a "key=value" query
// fragment (e.g. "cursor=abc"); both are normalized to the cursor param so the
// loop never follows an off-host absolute URL (SSRF guard).
func applyNextCursor(query url.Values, next string) {
	if next == "" {
		return
	}
	if i := strings.Index(next, "?"); i >= 0 {
		next = next[i+1:]
	}
	if parsed, err := url.ParseQuery(next); err == nil && len(parsed) > 0 {
		for k, vs := range parsed {
			query.Del(k)
			for _, v := range vs {
				query.Add(k, v)
			}
		}
		return
	}
	query.Set(cursorParam, next)
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise just-sift credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		switch stream {
		case "fields":
			item = map[string]any{
				"id":          fmt.Sprintf("field_fixture_%d", i),
				"type":        "text",
				"displayName": fmt.Sprintf("Fixture Field %d", i),
				"objectKey":   fmt.Sprintf("fixtureField%d", i),
				"filterable":  true,
				"searchable":  true,
			}
		default:
			item = map[string]any{
				"id":                fmt.Sprintf("person_fixture_%d", i),
				"displayName":       fmt.Sprintf("Fixture Person %d", i),
				"firstName":         "Fixture",
				"lastName":          fmt.Sprintf("Person %d", i),
				"email":             fmt.Sprintf("fixture+%d@example.com", i),
				"phone":             "+1-555-0100",
				"title":             "Engineer",
				"department":        "Engineering",
				"companyName":       "Example Co",
				"officeCity":        "Boston",
				"officeState":       "MA",
				"directoryId":       "dir_fixture",
				"isTeamLeader":      false,
				"directReportCount": int64(i),
				"pictureUrl":        "https://example.com/fixture.png",
			}
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
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("just-sift connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(token),
		UserAgent: userAgent,
	}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// baseURL resolves and validates the base URL. The default is api.justsift.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("just-sift config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("just-sift config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("just-sift config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSizeOf(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("just-sift config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("just-sift config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPagesOf(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("just-sift config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("just-sift config max_pages must be 0 for unlimited or a positive integer")
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
