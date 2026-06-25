// Package ciscomeraki implements the native pm Cisco Meraki connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction + RFC5988 Link-header pagination) with
// Meraki-specific stream definitions and endpoints.
//
// The directory/registry key is the bare system name "cisco-meraki"; the Go
// package identifier drops the hyphen ("ciscomeraki"). It self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
//
// Read-only: the Meraki Dashboard API exposes configuration/state for networks
// and devices and offers no obvious safe reverse-ETL write for these streams, so
// Capabilities.Write is false.
package ciscomeraki

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
	registryName          = "cisco-meraki"
	merakiDefaultBaseURL  = "https://api.meraki.com/api/v1"
	merakiDefaultPageSize = 1000
	merakiMaxPageSize     = 1000
	merakiUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Cisco Meraki connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Cisco Meraki connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Cisco Meraki",
		IntegrationType: "api",
		Description:     "Reads Cisco Meraki organizations, networks, devices, and admins from the Meraki Dashboard API v1.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Meraki. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := merakiBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(merakiSecret(cfg)) == "" {
		return errors.New("cisco-meraki connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing organizations confirms auth and connectivity without mutating
	// anything; it is the lightest authenticated read in the API.
	if err := r.DoJSON(ctx, http.MethodGet, "organizations", url.Values{"perPage": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check cisco-meraki: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: merakiStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "organizations"
	}
	endpoint, ok := merakiStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("cisco-meraki stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := merakiPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := merakiMaxPages(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.orgScoped {
		return c.harvest(ctx, r, endpoint.resource, pageSize, maxPages, "", endpoint.mapRecord, emit)
	}

	// Org-scoped streams fan out: list every accessible organization, then read
	// the per-organization resource for each, stamping organizationId on every
	// record so the rows remain attributable after the fan-in.
	orgIDs, err := c.organizationIDs(ctx, r, pageSize, maxPages)
	if err != nil {
		return err
	}
	for _, orgID := range orgIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := endpoint.pathFor(orgID)
		if err := c.harvest(ctx, r, path, pageSize, maxPages, orgID, endpoint.mapRecord, emit); err != nil {
			return err
		}
	}
	return nil
}

// organizationIDs lists every organization id the API key can access. It is used
// to drive the fan-out for org-scoped streams.
func (c Connector) organizationIDs(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int) ([]string, error) {
	var ids []string
	err := c.harvest(ctx, r, "organizations", pageSize, maxPages, "", merakiOrganizationRecord, func(rec connectors.Record) error {
		if id := stringField(rec, "id"); id != "" {
			ids = append(ids, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// harvest drives Meraki's RFC5988 Link-header pagination for a single endpoint.
// Meraki list responses are top-level JSON arrays; the next page is advertised as
// a rel="next" link in the Link header (an absolute URL carrying startingAfter).
// orgID, when non-empty, is stamped onto each mapped record.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, pageSize, maxPages int, orgID string, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("perPage", strconv.Itoa(pageSize))

	p := &connsdk.LinkHeaderPaginator{FirstQuery: base}
	page := p.Start()
	for pageNum := 0; page != nil; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && pageNum >= maxPages {
			return nil
		}

		reqPath := path
		query := base
		if page.URL != "" {
			// Link-header next URLs are absolute and already carry pagination
			// params, so the path is replaced and no base query is re-applied.
			reqPath = page.URL
			query = url.Values{}
		}

		resp, err := r.Do(ctx, http.MethodGet, reqPath, query, nil)
		if err != nil {
			return fmt.Errorf("read cisco-meraki %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode cisco-meraki %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			record := mapRecord(item)
			if orgID != "" {
				record["organizationId"] = orgID
			}
			if err := emit(record); err != nil {
				return err
			}
		}
		page = p.Next(resp, len(records))
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise cisco-meraki credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := merakiStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                   fmt.Sprintf("%s_fixture_%d", stream, i),
			"organizationId":       "org_fixture_1",
			"name":                 fmt.Sprintf("Fixture %s %d", strings.TrimPrefix(stream, "organization_"), i),
			"url":                  fmt.Sprintf("https://dashboard.meraki.com/o/fixture_%d", i),
			"serial":               fmt.Sprintf("Q2XX-FIX%d-0000", i),
			"model":                "MR46",
			"mac":                  fmt.Sprintf("00:11:22:33:44:0%d", i),
			"networkId":            "N_fixture_1",
			"productType":          "wireless",
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"orgAccess":            "read-only",
			"authenticationMethod": "Email",
			"accountStatus":        "ok",
			"twoFactorAuthEnabled": false,
			"hasApiKey":            false,
			"timeZone":             "America/Los_Angeles",
			"productTypes":         []any{"appliance", "wireless"},
			"tags":                 []any{"fixture"},
		}
		record := endpoint.mapRecord(item)
		if stream != "organizations" {
			record["organizationId"] = "org_fixture_1"
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
	base, err := merakiBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := merakiSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("cisco-meraki connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: merakiUserAgent,
	}, nil
}

func merakiSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// merakiBaseURL resolves and validates the base URL. The default is
// api.meraki.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func merakiBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return merakiDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("cisco-meraki config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("cisco-meraki config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("cisco-meraki config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func merakiPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return merakiDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("cisco-meraki config page_size must be an integer: %w", err)
	}
	if value < 1 || value > merakiMaxPageSize {
		return 0, fmt.Errorf("cisco-meraki config page_size must be between 1 and %d", merakiMaxPageSize)
	}
	return value, nil
}

func merakiMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("cisco-meraki config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("cisco-meraki config max_pages must be 0 for unlimited or a positive integer")
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

// Write is required by the Connector interface but cisco-meraki is read-only.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
