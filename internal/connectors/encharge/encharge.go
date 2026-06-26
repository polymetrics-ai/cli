// Package encharge implements the native pm Encharge connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction + offset pagination),
// modeled on the stripe reference connector.
//
// Encharge (https://encharge.io) is a marketing automation platform. The API
// (https://api.encharge.io/v1) authenticates with an X-Encharge-Token header and
// pages list endpoints with limit/offset. This connector is read-only: the
// upstream Airbyte source supports full-refresh extraction only, and there is no
// obviously safe reverse-ETL write surface, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package encharge

import (
	"context"
	"encoding/json"
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
	enchargeDefaultBaseURL  = "https://api.encharge.io/v1"
	enchargeDefaultPageSize = 100
	enchargeMaxPageSize     = 100
	enchargeTokenHeader     = "X-Encharge-Token"
	enchargeUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("encharge", New)
}

// New returns the Encharge connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Encharge connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "encharge" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "encharge",
		DisplayName:     "Encharge",
		IntegrationType: "api",
		Description:     "Reads Encharge people, segments, fields, account tags, and schemas through the Encharge REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Encharge. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := enchargeBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(enchargeSecret(cfg)) == "" {
		return errors.New("encharge connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the people list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"limit": []string{"1"}, "offset": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "people/all", q, nil, nil); err != nil {
		return fmt.Errorf("check encharge: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Encharge is read-only
// (Capabilities.Write is false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: enchargeStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "peoples"
	}
	endpoint, ok := enchargeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("encharge stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := enchargePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := enchargeMaxPages(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.paginated {
		return c.readSinglePage(ctx, r, endpoint, emit)
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// readSinglePage reads a non-paginated Encharge endpoint in one request.
func (c Connector) readSinglePage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read encharge %s: %w", endpoint.resource, err)
	}
	records, err := decodeRecords(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode encharge %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives Encharge's offset pagination. List endpoints return
// {<recordsPath>:[...]}; the next page is requested with offset += limit. A page
// shorter than the limit signals the end. There is no body token to follow, so
// the loop lives here, built on connsdk.Requester + decodeRecords.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{
			"limit":  []string{strconv.Itoa(pageSize)},
			"offset": []string{strconv.Itoa(offset)},
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read encharge %s: %w", endpoint.resource, err)
		}
		records, err := decodeRecords(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode encharge %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (or empty page) means we have read everything.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise encharge credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":        fmt.Sprintf("%s_fixture_%d", stream, i),
			"email":     fmt.Sprintf("fixture+%d@example.com", i),
			"name":      fmt.Sprintf("Fixture %d", i),
			"firstName": "Fixture",
			"lastName":  strconv.Itoa(i),
			"phone":     "",
			"title":     "Contact",
			"company":   "Example Inc",
			"country":   "US",
			"createdAt": "2026-01-01T00:00:00Z",
			"updatedAt": "2026-01-02T00:00:00Z",
			"userId":    fmt.Sprintf("u_%d", i),
			"type":      "people",
			"format":    "text",
			"tag":       fmt.Sprintf("tag_%d", i),
		}
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-Encharge-Token API-key auth
// and the resolved base URL. The secret only ever flows into the authenticator;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := enchargeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := enchargeSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("encharge connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(enchargeTokenHeader, secret, ""),
		UserAgent: enchargeUserAgent,
	}, nil
}

func enchargeSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// enchargeBaseURL resolves and validates the base URL. The default is
// api.encharge.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func enchargeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return enchargeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("encharge config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("encharge config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("encharge config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func enchargePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return enchargeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("encharge config page_size must be an integer: %w", err)
	}
	if value < 1 || value > enchargeMaxPageSize {
		return 0, fmt.Errorf("encharge config page_size must be between 1 and %d", enchargeMaxPageSize)
	}
	return value, nil
}

func enchargeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("encharge config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("encharge config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// decodeRecords extracts the records array at path from an Encharge response. It
// mirrors connsdk.RecordsAt for object arrays, but additionally wraps scalar
// (string) array elements into {"tag": <value>} records so the account_tags
// stream — whose /tags-management payload may be a plain array of tag strings —
// yields usable records instead of being dropped.
func decodeRecords(body []byte, path string) ([]map[string]any, error) {
	var root any
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	node := selectPath(root, path)
	if node == nil {
		return nil, nil
	}
	switch v := node.(type) {
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			switch obj := item.(type) {
			case map[string]any:
				out = append(out, obj)
			case string:
				out = append(out, map[string]any{"tag": obj})
			}
		}
		return out, nil
	case map[string]any:
		return []map[string]any{v}, nil
	default:
		return nil, nil
	}
}

// selectPath walks a decoded JSON value along a dotted path (e.g. "people",
// "result.items"). It returns nil when any segment is missing.
func selectPath(root any, path string) any {
	path = strings.TrimSpace(path)
	if path == "" || path == "." {
		return root
	}
	cur := root
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			continue
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = obj[seg]
		if !ok {
			return nil
		}
	}
	return cur
}
