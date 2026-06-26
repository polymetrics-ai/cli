// Package ninjaonermm implements the native pm NinjaOne RMM connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package composing the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with NinjaOne-specific stream definitions and endpoints.
//
// The directory is internal/connectors/ninjaone-rmm; the Go package identifier is
// ninjaonermm and the registry/factory key is the bare hyphenated name
// "ninjaone-rmm". It self-registers via RegisterFactory in init().
//
// NinjaOne's v2 list endpoints return a bare top-level JSON array and page forward
// with pageSize + after=<last entity id>. Auth is a Bearer token (the api_key
// secret). The connector is read-only.
package ninjaonermm

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
	defaultBaseURL  = "https://app.ninjarmm.com"
	apiPathPrefix   = "v2"
	defaultPageSize = 100
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("ninjaone-rmm", New)
}

// New returns the NinjaOne RMM connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm NinjaOne RMM connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "ninjaone-rmm" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "ninjaone-rmm",
		DisplayName:     "NinjaOne RMM",
		IntegrationType: "api",
		Description:     "Reads NinjaOne RMM organizations, devices, locations, activities, and policies through the NinjaOne v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to NinjaOne. In
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
	if strings.TrimSpace(apiKey(cfg)) == "" {
		return errors.New("ninjaone-rmm connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organizations list confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"pageSize": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, apiPathPrefix+"/organizations", q, nil, nil); err != nil {
		return fmt.Errorf("check ninjaone-rmm: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "organizations"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("ninjaone-rmm stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
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
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives NinjaOne's after-cursor pagination. v2 list endpoints return a
// bare JSON array; the next page is requested with after=<last entity id> until a
// page shorter than pageSize comes back. Non-paginated endpoints return their
// full set in one request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := apiPathPrefix + "/" + endpoint.resource
	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("pageSize", strconv.Itoa(pageSize))
			if after != "" {
				query.Set("after", after)
			}
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read ninjaone-rmm %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode ninjaone-rmm %s page: %w", endpoint.resource, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, "id")
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Non-paginated endpoints and short pages terminate the loop.
		if !endpoint.paginated || lastID == "" || len(records) < pageSize {
			return nil
		}
		after = lastID
	}
	return nil
}

// Write is unsupported: NinjaOne RMM is exposed read-only (no safe reverse-ETL
// write actions in this stream set). It satisfies the connectors.Connector
// interface by returning ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               i,
			"name":             fmt.Sprintf("%s fixture %d", endpoint.resource, i),
			"description":      fmt.Sprintf("fixture %s record %d", stream, i),
			"nodeApprovalMode": "AUTOMATIC",
			"organizationId":   1,
			"locationId":       1,
			"systemName":       fmt.Sprintf("host-%d", i),
			"dnsName":          fmt.Sprintf("host-%d.example.com", i),
			"nodeClass":        "WINDOWS_WORKSTATION",
			"offline":          false,
			"approvalStatus":   "APPROVED",
			"address":          fmt.Sprintf("%d Example St", i),
			"activityTime":     1767225600 + i,
			"deviceId":         i,
			"activityType":     "ACTION",
			"status":           "COMPLETED",
			"message":          fmt.Sprintf("fixture activity %d", i),
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
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := apiKey(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("ninjaone-rmm connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: userAgent,
	}, nil
}

func apiKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// baseURL resolves and validates the base URL. The default is app.ninjarmm.com;
// any override must be an absolute https (or http for local test servers) URL with
// a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("ninjaone-rmm config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("ninjaone-rmm config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("ninjaone-rmm config base_url must include a host")
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
		return 0, fmt.Errorf("ninjaone-rmm config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("ninjaone-rmm config page_size must be between 1 and %d", maxPageSize)
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
		return 0, fmt.Errorf("ninjaone-rmm config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("ninjaone-rmm config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
