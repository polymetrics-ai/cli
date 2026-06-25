// Package coda implements the native pm Coda connector. It is a declarative-HTTP
// per-system connector built on the connsdk toolkit (Requester + Bearer auth +
// items[] extraction + nextPageToken cursor pagination), mirroring the stripe
// reference connector.
//
// Coda is a read-only source: it reads docs and doc-scoped objects (tables,
// pages, formulas, controls) from the Coda REST API v1. It self-registers with
// the connectors registry via RegisterFactory in init(); the registryset
// package blank-imports this package in the production binary to run that side
// effect.
package coda

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
	codaDefaultBaseURL  = "https://coda.io/apis/v1"
	codaDefaultPageSize = 25
	codaMaxPageSize     = 100
	codaUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("coda", New)
}

// New returns the Coda connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Coda connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "coda" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "coda",
		DisplayName:     "Coda",
		IntegrationType: "api",
		Description:     "Reads Coda docs and doc-scoped tables, pages, formulas, and controls through the Coda REST API v1. Read-only source.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Coda. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := codaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(codaSecret(cfg)) == "" {
		return errors.New("coda connector requires secret auth_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the docs list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "docs", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check coda: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: codaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "docs"
	}
	endpoint, ok := codaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("coda stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	path, docID, err := resolvePath(endpoint, req.Config)
	if err != nil {
		return err
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := codaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := codaMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	paginator := &connsdk.CursorPaginator{
		CursorParam: "pageToken",
		TokenPath:   "nextPageToken",
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, path, base, paginator, "items", maxPages, func(rec connsdk.Record) error {
		out := endpoint.mapRecord(rec)
		// Stamp the doc id on doc-scoped records so downstream joins have it.
		if endpoint.docScoped && docID != "" {
			out["doc_id"] = docID
		}
		return emit(out)
	})
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise coda credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type":        strings.TrimSuffix(stream, "s"),
			"name":        fmt.Sprintf("Fixture %s %d", stream, i),
			"href":        fmt.Sprintf("https://coda.io/apis/v1/%s/fixture_%d", endpoint.resource, i),
			"browserLink": fmt.Sprintf("https://coda.io/d/fixture_%d", i),
			"tableType":   "grid",
			"controlType": "button",
			"contentType": "canvas",
			"subtitle":    "fixture subtitle",
			"rowCount":    int64(10 * i),
			"owner":       "fixture@example.com",
			"ownerName":   "Fixture Owner",
			"createdAt":   "2026-01-01T00:00:00.000Z",
			"updatedAt":   "2026-01-02T00:00:00.000Z",
			"workspaceId": "ws-fixture",
			"folderId":    "fl-fixture",
		}
		record := endpoint.mapRecord(item)
		if endpoint.docScoped {
			record["doc_id"] = "fixture_doc"
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
	base, err := codaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := codaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("coda connector requires secret auth_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: codaUserAgent,
	}, nil
}

// resolvePath builds the request path for a stream. Workspace lists use the
// resource directly; doc-scoped lists need a doc_id config and resolve to
// docs/<doc_id>/<resource>. The returned doc id is empty for workspace lists.
func resolvePath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, string, error) {
	if !endpoint.docScoped {
		return endpoint.resource, "", nil
	}
	docID := strings.TrimSpace(cfg.Config["doc_id"])
	if docID == "" {
		return "", "", fmt.Errorf("coda stream %q requires config doc_id", endpoint.resource)
	}
	return "docs/" + url.PathEscape(docID) + "/" + endpoint.resource, docID, nil
}

func codaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["auth_token"]
}

// codaBaseURL resolves and validates the base URL. The default is coda.io; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func codaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return codaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("coda config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("coda config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("coda config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func codaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return codaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("coda config page_size must be an integer: %w", err)
	}
	if value < 1 || value > codaMaxPageSize {
		return 0, fmt.Errorf("coda config page_size must be between 1 and %d", codaMaxPageSize)
	}
	return value, nil
}

func codaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("coda config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("coda config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Coda is a read-only source for pm.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
