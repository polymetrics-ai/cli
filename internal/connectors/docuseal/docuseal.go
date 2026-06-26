// Package docuseal implements the native pm DocuSeal source connector. It follows
// the declarative-HTTP template (see internal/connectors/stripe): a thin package
// that composes the connsdk toolkit (Requester + X-Auth-Token API-key auth +
// RecordsAt extraction + cursor state) with DocuSeal-specific stream
// definitions, endpoints, and cursor pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// DocuSeal is a read-only source: it exposes templates, submissions, and
// submitters over the DocuSeal REST API. There are no safe generic reverse-ETL
// writes for this source, so Capabilities.Write is false.
package docuseal

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
	docusealDefaultBaseURL  = "https://api.docuseal.com"
	docusealDefaultPageSize = 10
	docusealMaxPageSize     = 100
	docusealUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("docuseal", New)
}

// New returns the DocuSeal connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm DocuSeal source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "docuseal" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "docuseal",
		DisplayName:     "DocuSeal",
		IntegrationType: "api",
		Description:     "Reads DocuSeal templates, submissions, and submitters through the DocuSeal REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to DocuSeal. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := docusealBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(docusealSecret(cfg)) == "" {
		return errors.New("docuseal connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the templates list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "templates", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check docuseal: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: docusealStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a DocuSeal stream starts with
// an empty incremental cursor (full sync). DocuSeal's list endpoints do not
// filter by timestamp, so the cursor is informational and used only to record
// progress.
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
		stream = "templates"
	}
	endpoint, ok := docusealStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("docuseal stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := docusealPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := docusealMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives DocuSeal's cursor pagination. List endpoints return
// {data:[...], pagination:{count,next,prev}}; the next page is requested with
// after=<pagination.next>. When pagination.next is null/absent, iteration stops.
// There is no connsdk paginator for this exact "next id in body -> after param"
// shape, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read docuseal %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode docuseal %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next")
		if err != nil {
			return fmt.Errorf("decode docuseal %s pagination: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		// A null/absent next, an empty page, or a non-advancing cursor ends it.
		if next == "" || next == "null" || len(records) == 0 || next == after {
			return nil
		}
		after = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise docuseal credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    int64(i),
			"slug":                  fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":                  fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"source":                "api",
			"status":                "completed",
			"email":                 fmt.Sprintf("fixture+%d@example.com", i),
			"role":                  "First Party",
			"submission_id":         int64(i),
			"uuid":                  fmt.Sprintf("uuid-fixture-%d", i),
			"folder_name":           "Default",
			"external_id":           fmt.Sprintf("ext_%d", i),
			"author_id":             int64(1),
			"audit_log_url":         "",
			"combined_document_url": "",
			"template": map[string]any{
				"id":   int64(99),
				"name": "Fixture Template",
			},
			"created_at": "2026-01-01T00:00:00Z",
			"updated_at": "2026-01-02T00:00:00Z",
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

// requester builds a connsdk.Requester wired with X-Auth-Token API-key auth and
// the resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := docusealBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := docusealSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("docuseal connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Auth-Token", secret, ""),
		UserAgent: docusealUserAgent,
	}, nil
}

func docusealSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// docusealBaseURL resolves and validates the base URL. The default is
// api.docuseal.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func docusealBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return docusealDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("docuseal config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("docuseal config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("docuseal config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func docusealPageSize(cfg connectors.RuntimeConfig) (int, error) {
	// DocuSeal calls the page size "limit"; accept either page_size or limit.
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["limit"])
	}
	if raw == "" {
		return docusealDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("docuseal config limit must be an integer: %w", err)
	}
	if value < 1 || value > docusealMaxPageSize {
		return 0, fmt.Errorf("docuseal config limit must be between 1 and %d", docusealMaxPageSize)
	}
	return value, nil
}

func docusealMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("docuseal config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("docuseal config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: DocuSeal is exposed as a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
