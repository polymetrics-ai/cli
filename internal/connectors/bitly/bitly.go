// Package bitly implements the native pm Bitly connector. It follows the
// declarative-HTTP per-system connector shape established by the stripe package:
// a thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction) with Bitly-specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Bitly's API (https://api-ssl.bitly.com/v4) is authenticated with a Bearer
// access token and is read-only here: the core list endpoints (organizations,
// groups, campaigns, bitlinks) have no obviously-safe reverse-ETL writes, so the
// connector advertises Write=false.
package bitly

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
	bitlyDefaultBaseURL  = "https://api-ssl.bitly.com/v4"
	bitlyDefaultPageSize = 50
	bitlyMaxPageSize     = 100
	bitlyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("bitly", New)
}

// New returns the Bitly connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Bitly connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "bitly" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "bitly",
		DisplayName:     "Bitly",
		IntegrationType: "api",
		Description:     "Reads Bitly organizations, groups, campaigns, and bitlinks through the Bitly v4 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Bitly. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := bitlyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(bitlySecret(cfg)) == "" {
		return errors.New("bitly connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the groups list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "groups", nil, nil, nil); err != nil {
		return fmt.Errorf("check bitly: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bitlyStreams()}, nil
}

// Write is unsupported: the Bitly connector is read-only (the core list
// endpoints have no obviously-safe reverse-ETL writes). It satisfies the
// connectors.Connector interface by reporting the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "groups"
	}
	endpoint, ok := bitlyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("bitly stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := bitlyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := bitlyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, resource, pageSize, maxPages, emit)
}

// harvest drives a Bitly list read. Non-paginated endpoints return a single
// page; paginated endpoints (bitlinks) carry a pagination.next absolute URL that
// is followed until empty. Both shapes are decoded with connsdk.RecordsAt at the
// endpoint's recordsKey.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, resource string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	query := url.Values{}
	if endpoint.paginated {
		query.Set("size", strconv.Itoa(pageSize))
	}

	path := resource
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read bitly %s: %w", endpoint.recordsKey, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode bitly %s page: %w", endpoint.recordsKey, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if !endpoint.paginated {
			return nil
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next")
		if err != nil {
			return fmt.Errorf("decode bitly %s pagination: %w", endpoint.recordsKey, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// pagination.next is an absolute URL with its own query (search_after);
		// the Requester treats an absolute path as-is, so clear the base query.
		path = next
		query = url.Values{}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bitly credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"guid":              fmt.Sprintf("%s_fixture_%d", stream, i),
			"id":                fmt.Sprintf("bit.ly/fixture_%d", i),
			"name":              fmt.Sprintf("Fixture %s %d", stream, i),
			"description":       "fixture record",
			"organization_guid": "org_fixture_1",
			"group_guid":        "grp_fixture_1",
			"is_active":         true,
			"link":              fmt.Sprintf("https://bit.ly/fixture_%d", i),
			"long_url":          fmt.Sprintf("https://example.com/%d", i),
			"title":             fmt.Sprintf("Fixture link %d", i),
			"archived":          false,
			"created":           "2026-01-01T00:00:00+0000",
			"modified":          "2026-01-01T00:00:00+0000",
			"created_at":        "2026-01-01T00:00:00+0000",
			"role":              "admin",
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "bitly"
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := bitlyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := bitlySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("bitly connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: bitlyUserAgent,
	}, nil
}

// resolveResource substitutes the {group_guid} placeholder for group-scoped
// endpoints (bitlinks). The group guid comes from config; it is required for
// those streams.
func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.groupScoped {
		return endpoint.resource, nil
	}
	guid := strings.TrimSpace(cfg.Config["group_guid"])
	if guid == "" {
		return "", errors.New("bitly bitlinks stream requires config group_guid")
	}
	return strings.ReplaceAll(endpoint.resource, "{group_guid}", url.PathEscape(guid)), nil
}

func bitlySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// bitlyBaseURL resolves and validates the base URL. The default is
// api-ssl.bitly.com/v4; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func bitlyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return bitlyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("bitly config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("bitly config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("bitly config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func bitlyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return bitlyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bitly config page_size must be an integer: %w", err)
	}
	if value < 1 || value > bitlyMaxPageSize {
		return 0, fmt.Errorf("bitly config page_size must be between 1 and %d", bitlyMaxPageSize)
	}
	return value, nil
}

func bitlyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bitly config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("bitly config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
