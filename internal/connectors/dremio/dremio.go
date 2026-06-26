// Package dremio implements the native pm Dremio connector. It is a declarative-
// HTTP per-system connector built on the connsdk toolkit (Requester + Bearer PAT
// auth + RecordsAt extraction) with Dremio-specific stream definitions and
// endpoints, copying the shape of the stripe reference connector.
//
// Dremio's REST API authenticates with a Personal Access Token presented as
// `Authorization: Bearer <api_key>`, returns list payloads as {"data":[...]} and
// paginates via a top-level "nextPageToken" echoed back as the "pageToken" query
// parameter. The connector is read-only.
//
// Like the other native connectors it self-registers with the connectors
// registry via RegisterFactory in init(); the registryset package blank-imports
// this package in the production binary to run that side effect.
package dremio

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
	// dremioDefaultBaseURL is the Dremio Cloud (US) REST API root. Software /
	// self-hosted and EU users override this via the base_url config field.
	dremioDefaultBaseURL  = "https://api.dremio.cloud/v0"
	dremioDefaultPageSize = 100
	dremioMaxPageSize     = 500
	dremioUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("dremio", New)
}

// New returns the Dremio connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Dremio connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "dremio" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "dremio",
		DisplayName:     "Dremio",
		IntegrationType: "api",
		Description:     "Reads Dremio catalog entries, reflections, sources, and users through the Dremio REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Dremio. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dremioBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(dremioSecret(cfg)) == "" {
		return errors.New("dremio connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the catalog confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "catalog", url.Values{"maxResults": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check dremio: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Dremio is read-only, so
// any write request is rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dremioStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "catalog"
	}
	endpoint, ok := dremioStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("dremio stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := dremioPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := dremioMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Dremio's body-cursor pagination. List endpoints return
// {"data":[...], "nextPageToken":"..."}; the next page is requested with
// pageToken=<token>. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt, because connsdk's CursorPaginator
// terminates on an empty page rather than a missing token and this shape needs
// the explicit token check.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("maxResults", strconv.Itoa(pageSize))

	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read dremio %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode dremio %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode dremio %s nextPageToken: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise dremio credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type":          "CONTAINER",
			"containerType": "SOURCE",
			"datasetType":   "PROMOTED",
			"path":          []any{"fixture", stream, strconv.Itoa(i)},
			"tag":           fmt.Sprintf("tag_%d", i),
			"createdAt":     "2026-01-01T00:00:00.000Z",
			"name":          fmt.Sprintf("fixture-%s-%d", stream, i),
			"firstName":     "Fixture",
			"lastName":      strconv.Itoa(i),
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"active":        true,
			"enabled":       true,
			"datasetId":     "ds_fixture_1",
			"status":        map[string]any{"config": "OK"},
			"updatedAt":     "2026-01-02T00:00:00.000Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer (PAT) auth and the
// resolved base URL. The secret only ever flows into connsdk.Bearer; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dremioBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := dremioSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("dremio connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: dremioUserAgent,
	}, nil
}

func dremioSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// dremioBaseURL resolves and validates the base URL. The default is the Dremio
// Cloud REST API root; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func dremioBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return dremioDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dremio config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("dremio config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("dremio config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func dremioPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dremioDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dremio config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dremioMaxPageSize {
		return 0, fmt.Errorf("dremio config page_size must be between 1 and %d", dremioMaxPageSize)
	}
	return value, nil
}

func dremioMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dremio config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("dremio config max_pages must be 0 for unlimited or a positive integer")
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
