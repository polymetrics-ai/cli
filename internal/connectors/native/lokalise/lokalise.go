// Package lokalise implements the native pm Lokalise connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + APIKeyHeader auth +
// RecordsAt extraction) with Lokalise-specific stream definitions and endpoints.
//
// Lokalise is a localization/translation management platform. Its REST API
// (https://api.lokalise.com/api2) is project-scoped: list endpoints live under
// projects/{project_id}/{resource} and return {"<resource>":[...]} bodies with
// offset pagination (page/limit) reported via X-Pagination-Page-Count and
// X-Pagination-Page response headers. Authentication is via the X-Api-Token
// header carrying a read-access API key.
package lokalise

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
	lokaliseDefaultBaseURL  = "https://api.lokalise.com/api2"
	lokaliseDefaultPageSize = 100
	lokaliseMaxPageSize     = 5000
	lokaliseUserAgent       = "polymetrics-go-cli"
	// lokaliseSafetyPageCap bounds pagination when a server omits the
	// X-Pagination-Page-Count header, preventing an unbounded loop.
	lokaliseSafetyPageCap = 10000
)

// New returns the Lokalise connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Lokalise connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "lokalise" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "lokalise",
		DisplayName:     "Lokalise",
		IntegrationType: "api",
		Description:     "Reads Lokalise project keys, languages, translations, contributors, and comments through the Lokalise REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Lokalise. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := lokaliseBaseURL(cfg); err != nil {
		return err
	}
	projectID, err := lokaliseProjectID(cfg)
	if err != nil {
		return err
	}
	if strings.TrimSpace(lokaliseSecret(cfg)) == "" {
		return errors.New("lokalise connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the languages list confirms auth and connectivity
	// without mutating anything.
	path := "projects/" + url.PathEscape(projectID) + "/languages"
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check lokalise: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: lokaliseStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Lokalise is exposed as a
// read-only source, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "keys"
	}
	endpoint, ok := lokaliseStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("lokalise stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	projectID, err := lokaliseProjectID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := lokalisePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := lokaliseMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, projectID, endpoint, pageSize, maxPages, emit)
}

// harvest drives Lokalise's offset pagination. List responses carry the records
// under endpoint.recordsKey and report X-Pagination-Page-Count /
// X-Pagination-Page headers. The loop requests pages 1..pageCount, stopping when
// the current page reaches the reported page count, when a short/empty page is
// returned, when maxPages is reached, or when the safety cap is hit. Built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, projectID string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "projects/" + url.PathEscape(projectID) + "/" + endpoint.resource

	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}
		if page > lokaliseSafetyPageCap {
			return nil
		}

		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read lokalise %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode lokalise %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Stop when the X-Pagination-Page-Count header says we have read the last
		// page. When the header is absent, fall back to stopping on a short page.
		pageCount := headerInt(resp.Header, "X-Pagination-Page-Count")
		if pageCount > 0 {
			if page >= pageCount {
				return nil
			}
			continue
		}
		if len(records) < pageSize {
			return nil
		}
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise lokalise credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"key_id":                int64(i),
			"translation_id":        int64(i),
			"lang_id":               int64(640 + i),
			"user_id":               int64(1000 + i),
			"comment_id":            int64(i),
			"created_at":            "2026-01-01 00:00:0" + strconv.Itoa(i) + " (Etc/UTC)",
			"created_at_timestamp":  int64(1767225600 + i),
			"modified_at":           "2026-01-02 00:00:0" + strconv.Itoa(i) + " (Etc/UTC)",
			"modified_at_timestamp": int64(1767312000 + i),
			"key_name":              map[string]any{"web": fmt.Sprintf("welcome_%d", i)},
			"description":           fmt.Sprintf("Fixture key %d", i),
			"platforms":             []any{"web"},
			"tags":                  []any{"fixture"},
			"is_plural":             false,
			"is_hidden":             false,
			"is_archived":           false,
			"lang_iso":              fmt.Sprintf("en-%d", i),
			"lang_name":             fmt.Sprintf("English %d", i),
			"is_rtl":                false,
			"plural_forms":          []any{"one", "other"},
			"language_iso":          "en",
			"translation":           fmt.Sprintf("Hello %d", i),
			"modified_by":           int64(1000 + i),
			"modified_by_email":     fmt.Sprintf("fixture+%d@example.com", i),
			"is_reviewed":           false,
			"is_unverified":         false,
			"reviewed_by":           int64(0),
			"email":                 fmt.Sprintf("fixture+%d@example.com", i),
			"fullname":              fmt.Sprintf("Fixture %d", i),
			"is_admin":              false,
			"is_reviewer":           false,
			"languages":             []any{},
			"role_id":               int64(i),
			"comment":               fmt.Sprintf("Fixture comment %d", i),
			"added_by":              int64(1000 + i),
			"added_by_email":        fmt.Sprintf("fixture+%d@example.com", i),
			"added_at":              "2026-01-01 00:00:0" + strconv.Itoa(i) + " (Etc/UTC)",
			"added_at_timestamp":    int64(1767225600 + i),
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "lokalise"
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

// requester builds a connsdk.Requester wired with X-Api-Token auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := lokaliseBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := lokaliseSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("lokalise connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Api-Token", secret, ""),
		UserAgent: lokaliseUserAgent,
	}, nil
}

func lokaliseSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// lokaliseProjectID resolves the required project_id config value.
func lokaliseProjectID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("lokalise connector requires config project_id")
	}
	projectID := strings.TrimSpace(cfg.Config["project_id"])
	if projectID == "" {
		return "", errors.New("lokalise connector requires config project_id")
	}
	return projectID, nil
}

// lokaliseBaseURL resolves and validates the base URL. The default is
// api.lokalise.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func lokaliseBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return lokaliseDefaultBaseURL, nil
	}
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return lokaliseDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("lokalise config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("lokalise config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("lokalise config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func lokalisePageSize(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return lokaliseDefaultPageSize, nil
	}
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return lokaliseDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lokalise config page_size must be an integer: %w", err)
	}
	if value < 1 || value > lokaliseMaxPageSize {
		return 0, fmt.Errorf("lokalise config page_size must be between 1 and %d", lokaliseMaxPageSize)
	}
	return value, nil
}

func lokaliseMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lokalise config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("lokalise config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// headerInt reads an integer-valued response header, returning 0 when absent or
// unparseable.
func headerInt(h http.Header, name string) int {
	raw := strings.TrimSpace(h.Get(name))
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}
