// Package ezofficeinventory implements the native pm EZOfficeInventory connector.
// It is a declarative-HTTP per-system connector built on the connsdk toolkit,
// modeled on the stripe reference connector: a thin package that composes a
// connsdk Requester with an APIKeyHeader authenticator (EZOfficeInventory's
// `token` header), page-number pagination, and EZOfficeInventory-specific stream
// definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package ezofficeinventory

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
	ezoDefaultPageSize = 25
	ezoMaxPageSize     = 100
	ezoUserAgent       = "polymetrics-go-cli"
	// ezoAuthHeader is the header EZOfficeInventory expects the API access token
	// in (per the upstream upstream manifest's ApiKeyAuthenticator).
	ezoAuthHeader = "token"
)

func init() {
	connectors.RegisterFactory("ezofficeinventory", New)
}

// New returns the EZOfficeInventory connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm EZOfficeInventory connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "ezofficeinventory" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "ezofficeinventory",
		DisplayName:     "EZOfficeInventory",
		IntegrationType: "api",
		Description:     "Reads EZOfficeInventory assets, inventory items, stock assets, members, and locations through the EZOfficeInventory REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to
// EZOfficeInventory. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := ezoBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(ezoSecret(cfg)) == "" {
		return errors.New("ezofficeinventory connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the members list (the upstream check stream) confirms
	// auth and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "members.api", url.Values{"page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check ezofficeinventory: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. EZOfficeInventory is a
// read-only source connector (Capabilities.Write=false), so writes are rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: ezoStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "assets"
	}
	endpoint, ok := ezoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("ezofficeinventory stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := ezoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := ezoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives EZOfficeInventory's page-number pagination. List responses look
// like {<recordsPath>:[...], total_pages:N, page:M}; the next page is requested
// with page=M+1. The loop stops when an empty page is returned, when the current
// page reaches total_pages, or when maxPages is hit. There is no body-token
// paginator in connsdk for this exact total_pages shape, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))
		for k, v := range endpoint.staticParams {
			query.Set(k, v)
		}

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read ezofficeinventory %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode ezofficeinventory %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop on an empty page (defensive) or once we've consumed total_pages.
		if len(records) == 0 {
			return nil
		}
		totalPages, ok := ezoTotalPages(resp.Body)
		if ok && page >= totalPages {
			return nil
		}
		// If total_pages is absent, fall back to stopping on a short page.
		if !ok && len(records) < pageSize {
			return nil
		}
	}
}

// ezoTotalPages reads the total_pages field from a list response body. It returns
// (0, false) when the field is missing or not an integer.
func ezoTotalPages(body []byte) (int, bool) {
	raw, err := connsdk.StringAt(body, "total_pages")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, false
	}
	return n, true
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	idKey := "identifier"
	if stream == "members" || stream == "locations" {
		idKey = "id"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			idKey:                    int64(i),
			"name":                   fmt.Sprintf("%s Fixture %d", stream, i),
			"description":            "fixture record",
			"asset_type":             "Fixed Asset",
			"group_id":               int64(10),
			"location_id":            int64(100),
			"location_name":          "HQ",
			"assigned_to_user_email": fmt.Sprintf("fixture+%d@example.com", i),
			"assigned_to_user_name":  fmt.Sprintf("Fixture %d", i),
			"price":                  "100.00",
			"net_quantity":           int64(i),
			"purchased_on":           "2026-01-01",
			"created_at":             "2026-01-01T00:00:00Z",
			"updated_at":             "2026-01-02T00:00:00Z",
			"first_name":             "Fixture",
			"last_name":              fmt.Sprintf("User%d", i),
			"full_name":              fmt.Sprintf("Fixture User%d", i),
			"email":                  fmt.Sprintf("fixture+%d@example.com", i),
			"role_id":                int64(1),
			"role_name":              "Administrator",
			"status":                 "active",
			"contact_type":           "member",
			"country":                "US",
			"parent_id":              int64(0),
			"street1":                "1 Fixture St",
			"street2":                "",
			"city":                   "Testville",
			"state":                  "CA",
			"zipcode":                "90210",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the `token` header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := ezoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := ezoSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("ezofficeinventory connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(ezoAuthHeader, secret, ""),
		UserAgent: ezoUserAgent,
	}, nil
}

func ezoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// ezoBaseURL resolves and validates the base URL. When base_url is set (e.g. for
// tests) it is validated and used directly. Otherwise it is constructed from the
// required subdomain config: https://<subdomain>.ezofficeinventory.com. Any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func ezoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("ezofficeinventory config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("ezofficeinventory config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("ezofficeinventory config base_url must include a host")
		}
		return strings.TrimRight(base, "/"), nil
	}
	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("ezofficeinventory connector requires config subdomain (or base_url)")
	}
	if !validSubdomain(subdomain) {
		return "", fmt.Errorf("ezofficeinventory config subdomain %q is invalid", subdomain)
	}
	return "https://" + subdomain + ".ezofficeinventory.com", nil
}

// validSubdomain bounds the subdomain to DNS-label characters so it cannot inject
// a different host or path into the constructed base URL.
func validSubdomain(s string) bool {
	if len(s) > 63 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func ezoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return ezoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ezofficeinventory config page_size must be an integer: %w", err)
	}
	if value < 1 || value > ezoMaxPageSize {
		return 0, fmt.Errorf("ezofficeinventory config page_size must be between 1 and %d", ezoMaxPageSize)
	}
	return value, nil
}

func ezoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ezofficeinventory config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("ezofficeinventory config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
