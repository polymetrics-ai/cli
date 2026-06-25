// Package braze implements the native pm Braze connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction) with Braze-specific stream
// definitions and endpoints. It mirrors the stripe reference connector's shape.
//
// Braze's REST API is Bearer authenticated (the REST API key is passed as
// "Authorization: Bearer <key>"), and its list-export endpoints paginate with a
// 0-based ?page= parameter returning up to ~100 items per page. The connector is
// read-only: there are no safe, generic reverse-ETL writes to expose, so
// Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package braze

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	brazeDefaultPageSize = 100
	brazeMaxPageSize     = 100
	brazeUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("braze", New)
}

// New returns the Braze connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Braze connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "braze" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "braze",
		DisplayName:     "Braze",
		IntegrationType: "api",
		Description:     "Reads Braze campaigns, Canvases, segments, and custom events through the Braze REST API list-export endpoints.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Braze. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	base, err := brazeBaseURL(cfg)
	if err != nil {
		return err
	}
	if strings.TrimSpace(brazeSecret(cfg)) == "" {
		return errors.New("braze connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the first campaigns page confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "campaigns/list", url.Values{"page": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check braze (%s): %w", base, err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: brazeStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Braze is read-only: the
// REST API has no generic, safe reverse-ETL action to expose, so writes are
// unsupported (Capabilities.Write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "campaigns"
	}
	endpoint, ok := brazeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("braze stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := brazePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := brazeMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Braze's 0-based page-number pagination. Each list response is
// {"<recordsField>":[...]}; pages are requested with ?page=0,1,2,... and the
// loop stops on a short page (fewer than pageSize records) or an empty page.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read braze %s: %w", endpoint.resource, err)
		}
		records, err := decodeRecords(resp.Body, endpoint.recordsField)
		if err != nil {
			return fmt.Errorf("decode braze %s page %d: %w", endpoint.resource, page, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short or empty page means there is no further data to fetch.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// decodeRecords extracts the records array at field from a Braze list response.
// Most endpoints return arrays of objects; the events endpoint returns an array
// of event-name strings, which are wrapped into {"name": <string>} so a single
// mapper shape covers every stream.
func decodeRecords(body []byte, field string) ([]map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var envelope map[string]json.RawMessage
	if err := dec.Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}
	raw, ok := envelope[field]
	if !ok || len(raw) == 0 {
		return nil, nil
	}
	var elems []json.RawMessage
	if err := json.Unmarshal(raw, &elems); err != nil {
		return nil, fmt.Errorf("decode %s array: %w", field, err)
	}
	out := make([]map[string]any, 0, len(elems))
	for _, el := range elems {
		trimmed := bytes.TrimSpace(el)
		if len(trimmed) > 0 && trimmed[0] == '"' {
			// A bare string element (events list): wrap as {"name": <string>}.
			var s string
			if err := json.Unmarshal(trimmed, &s); err != nil {
				return nil, fmt.Errorf("decode %s string element: %w", field, err)
			}
			out = append(out, map[string]any{"name": s})
			continue
		}
		objDec := json.NewDecoder(bytes.NewReader(trimmed))
		objDec.UseNumber()
		var obj map[string]any
		if err := objDec.Decode(&obj); err != nil {
			// Skip non-object, non-string elements rather than failing the page.
			continue
		}
		out = append(out, obj)
	}
	return out, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise braze credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                         fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":                       fmt.Sprintf("Fixture %s %d", stream, i),
			"is_api_campaign":            false,
			"analytics_tracking_enabled": true,
			"tags":                       []any{"fixture"},
			"last_edited":                fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
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
	base, err := brazeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := brazeSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("braze connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: brazeUserAgent,
	}, nil
}

func brazeSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// brazeBaseURL resolves and validates the base URL. Braze has no single global
// host (each customer is on a regional REST endpoint, e.g.
// https://rest.iad-01.braze.com), so base_url is required in live mode. Any
// value must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func brazeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		// Accept the catalog's "url" field as an alias for base_url.
		base = strings.TrimSpace(cfg.Config["url"])
	}
	if base == "" {
		return "", errors.New("braze connector requires config base_url (your regional Braze REST endpoint, e.g. https://rest.iad-01.braze.com)")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("braze config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("braze config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("braze config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func brazePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return brazeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("braze config page_size must be an integer: %w", err)
	}
	if value < 1 || value > brazeMaxPageSize {
		return 0, fmt.Errorf("braze config page_size must be between 1 and %d", brazeMaxPageSize)
	}
	return value, nil
}

func brazeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("braze config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("braze config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
