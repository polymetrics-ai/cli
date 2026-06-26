// Package opinionstage implements the native pm Opinion Stage connector.
//
// Opinion Stage is a read-only JSON:API source. It authenticates with HTTP Basic
// auth using the API key as username and an empty password, exposes full-refresh
// items plus per-item responses and questions substreams, and paginates with
// page[number] / page[size].
package opinionstage

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
	opinionStageConnectorName   = "opinion-stage"
	opinionStageDefaultBaseURL  = "https://api.opinionstage.com"
	opinionStageDefaultPageSize = 50
	opinionStageMaxPageSize     = 1000
	opinionStageUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(opinionStageConnectorName, New)
}

// New returns the Opinion Stage connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Opinion Stage connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return opinionStageConnectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            opinionStageConnectorName,
		DisplayName:     "Opinion Stage",
		IntegrationType: "api",
		Description:     "Reads Opinion Stage items, responses, and questions through the Opinion Stage API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector can authenticate and reach the items stream. In
// fixture mode it short-circuits without credentials or network access.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := opinionStageBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(opinionStageSecret(cfg)) == "" {
		return errors.New("opinion-stage connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	query := url.Values{}
	query.Set("page[number]", "1")
	query.Set("page[size]", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "/api/v2/items", query, nil, nil); err != nil {
		return fmt.Errorf("check opinion-stage: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: opinionStageStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "items"
	}
	endpoint, ok := opinionStageStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("opinion-stage stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := opinionStagePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := opinionStageMaxPages(req.Config)
	if err != nil {
		return err
	}
	if endpoint.isSubstream() {
		return c.readSubstream(ctx, r, endpoint, pageSize, maxPages, emit)
	}
	return c.harvest(ctx, r, endpoint.path, pageSize, maxPages, func(item map[string]any) connectors.Record {
		return endpoint.mapRecord(item, "")
	}, emit)
}

// harvest drives Opinion Stage page-number pagination over a single endpoint.
// Pagination stops when a page returns fewer records than page[size] or when the
// optional max_pages guard is reached.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, pageSize, maxPages int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page[number]", strconv.Itoa(page))
		query.Set("page[size]", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read opinion-stage %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode opinion-stage %s page: %w", path, err)
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
	}
	return nil
}

// readSubstream first lists item IDs, then fans out to each per-item child
// endpoint. Emitted records are annotated with item_id, mirroring the upstream
// SubstreamPartitionRouter partition field.
func (c Connector) readSubstream(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	ids, err := c.itemIDs(ctx, r, pageSize, maxPages)
	if err != nil {
		return err
	}
	for _, itemID := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v2/items/%s/%s", url.PathEscape(itemID), endpoint.subresource)
		if err := c.harvest(ctx, r, path, pageSize, maxPages, func(item map[string]any) connectors.Record {
			return endpoint.mapRecord(item, itemID)
		}, emit); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) itemIDs(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int) ([]string, error) {
	var ids []string
	err := c.harvest(ctx, r, "/api/v2/items", pageSize, maxPages, opinionStageItemRecordForParent, func(rec connectors.Record) error {
		if id, ok := rec["id"].(string); ok && id != "" {
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
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		itemID := "item_fixture_1"
		item := map[string]any{
			"id":   fmt.Sprintf("%s_fixture_%d", stream, i),
			"type": strings.TrimSuffix(stream, "s"),
			"attributes": map[string]any{
				"title":  fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
				"status": "published",
				"kind":   "poll",
				"lead":   i == 1,
				"answers": []any{
					map[string]any{"question": "Fixture question", "selection": []any{"Fixture answer"}},
				},
				"result":  map[string]any{"title": "Fixture result", "text": "Fixture result text"},
				"timings": map[string]any{"duration": i},
				"utm":     map[string]any{},
				"timestamps": map[string]any{
					"created":  fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
					"modified": fmt.Sprintf("2026-01-0%dT01:00:00Z", i),
				},
			},
		}
		if endpoint.isSubstream() {
			itemID = "item_fixture_1"
		}
		if err := emit(endpoint.mapRecord(item, itemID)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth. The API key
// is the username and the password is intentionally empty.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := opinionStageBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := opinionStageSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("opinion-stage connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(strings.TrimSpace(secret), ""),
		UserAgent: opinionStageUserAgent,
	}, nil
}

func opinionStageSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func opinionStageBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return opinionStageDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("opinion-stage config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("opinion-stage config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("opinion-stage config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func opinionStagePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return opinionStageDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("opinion-stage config page_size must be an integer: %w", err)
	}
	if value < 1 || value > opinionStageMaxPageSize {
		return 0, fmt.Errorf("opinion-stage config page_size must be between 1 and %d", opinionStageMaxPageSize)
	}
	return value, nil
}

func opinionStageMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("opinion-stage config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("opinion-stage config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies connectors.Connector. Opinion Stage is a read-only source.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
