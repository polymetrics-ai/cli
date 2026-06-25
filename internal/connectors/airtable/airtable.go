// Package airtable implements the native pm Airtable connector. It follows the
// declarative-HTTP template established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Airtable-specific stream definitions, endpoints, and
// body-offset pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Airtable Web API surface used:
//   - GET /v0/meta/bases                  (bases stream)
//   - GET /v0/meta/bases/{baseId}/tables  (tables stream)
//   - GET /v0/{baseId}/{tableId}          (records stream)
//
// Auth is a Bearer token: a Personal Access Token (credentials.api_key) or an
// OAuth2 access token (credentials.access_token). The connector is read-only;
// Airtable record writes exist but are intentionally out of scope here.
package airtable

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
	airtableDefaultBaseURL  = "https://api.airtable.com/v0"
	airtableDefaultPageSize = 100
	airtableMaxPageSize     = 100
	airtableUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("airtable", New)
}

// New returns the Airtable connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Airtable connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "airtable" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "airtable",
		DisplayName:     "Airtable",
		IntegrationType: "api",
		Description:     "Reads Airtable bases, tables, and records through the Airtable Web API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Airtable. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := airtableBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(airtableSecret(cfg)) == "" {
		return errors.New("airtable connector requires a secret: credentials.api_key or credentials.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing bases confirms auth and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "meta/bases", nil, nil, nil); err != nil {
		return fmt.Errorf("check airtable: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: airtableStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Airtable connector is
// read-only (Capabilities.Write=false); reverse-ETL writes are out of scope.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "bases"
	}
	endpoint, ok := airtableStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("airtable stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	resource, err := streamResource(stream, endpoint, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := airtablePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := airtableMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint, pageSize, maxPages, emit)
}

// harvest drives Airtable's body-offset pagination. List responses include an
// optional top-level "offset" string; when present it is passed back as the
// ?offset= query param to fetch the next page, until it is absent.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	// pageSize only applies to the records endpoint; the metadata endpoints
	// ignore it harmlessly.
	if endpoint.needsTable {
		base.Set("pageSize", strconv.Itoa(pageSize))
	}

	offset := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if offset != "" {
			query.Set("offset", offset)
		}
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read airtable %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode airtable %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "offset")
		if err != nil {
			return fmt.Errorf("decode airtable %s offset: %w", resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		offset = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise airtable credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":            fmt.Sprintf("Fixture %s %d", stream, i),
			"permissionLevel": "create",
			"primaryFieldId":  "fld_fixture_1",
			"description":     "fixture",
			"createdTime":     "2026-01-01T00:00:00.000Z",
			"fields": map[string]any{
				"Name":      fmt.Sprintf("Fixture %d", i),
				"Score":     i * 10,
				"connector": "airtable",
				"fixture":   true,
			},
			"views": []any{},
		}
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := airtableBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := airtableSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("airtable connector requires a secret: credentials.api_key or credentials.access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: airtableUserAgent,
	}, nil
}

// streamResource builds the API path for a stream from the required config
// inputs, validating base_id / table_id presence and shape.
func streamResource(stream string, endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	switch {
	case !endpoint.needsBase && !endpoint.needsTable:
		return "meta/bases", nil
	case endpoint.needsBase && !endpoint.needsTable:
		baseID, err := configID(cfg, "base_id")
		if err != nil {
			return "", fmt.Errorf("airtable stream %q: %w", stream, err)
		}
		return "meta/bases/" + url.PathEscape(baseID) + "/tables", nil
	default:
		baseID, err := configID(cfg, "base_id")
		if err != nil {
			return "", fmt.Errorf("airtable stream %q: %w", stream, err)
		}
		tableID, err := configID(cfg, "table_id")
		if err != nil {
			return "", fmt.Errorf("airtable stream %q: %w", stream, err)
		}
		return url.PathEscape(baseID) + "/" + url.PathEscape(tableID), nil
	}
}

// configID reads and validates a required config identifier (base_id/table_id).
func configID(cfg connectors.RuntimeConfig, key string) (string, error) {
	value := strings.TrimSpace(cfg.Config[key])
	if value == "" {
		return "", fmt.Errorf("config %s is required", key)
	}
	return value, nil
}

// airtableSecret resolves the Bearer token, preferring the Personal Access Token
// (api_key) then the OAuth2 access token. Both the dotted catalog key and a bare
// key are accepted for ergonomics.
func airtableSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{
		"credentials.api_key",
		"credentials.access_token",
		"api_key",
		"access_token",
	} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// airtableBaseURL resolves and validates the base URL. The default is
// api.airtable.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func airtableBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return airtableDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("airtable config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("airtable config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("airtable config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func airtablePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return airtableDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("airtable config page_size must be an integer: %w", err)
	}
	if value < 1 || value > airtableMaxPageSize {
		return 0, fmt.Errorf("airtable config page_size must be between 1 and %d", airtableMaxPageSize)
	}
	return value, nil
}

func airtableMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("airtable config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("airtable config max_pages must be 0 for unlimited or a positive integer")
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
