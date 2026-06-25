// Package bamboohr implements the native pm BambooHR connector. It follows the
// declarative-HTTP per-system template established by the stripe connector: a
// thin package composing the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction) with BambooHR-specific stream definitions and endpoints.
//
// BambooHR is a read-only HR data source for this connector. It authenticates
// with HTTP Basic auth using the API key as the username and any string ("x") as
// the password, and serves JSON from https://<subdomain>.bamboohr.com/api/v1.
//
// The package self-registers with the connectors registry via RegisterFactory in
// init() under the key "bamboo-hr"; the registryset package blank-imports this
// package in the production binary to run that side effect.
package bamboohr

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
	bambooConnectorName   = "bamboo-hr"
	bambooDefaultPageSize = 100
	bambooMaxPageSize     = 1000
	bambooUserAgent       = "polymetrics-go-cli"
	// bambooBasicPassword is the conventional throwaway password BambooHR expects
	// when the API key is supplied as the Basic-auth username.
	bambooBasicPassword = "x"
)

func init() {
	connectors.RegisterFactory(bambooConnectorName, New)
}

// New returns the BambooHR connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm BambooHR connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return bambooConnectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            bambooConnectorName,
		DisplayName:     "BambooHR",
		IntegrationType: "api",
		Description:     "Reads BambooHR employees, field metadata, list metadata, and time off types through the BambooHR REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to BambooHR. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := bambooBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(bambooSecret(cfg)) == "" {
		return errors.New("bamboo-hr connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the field metadata confirms auth and connectivity
	// without mutating anything; meta/fields is a small, always-available list.
	if err := r.DoJSON(ctx, http.MethodGet, "meta/fields", nil, nil, nil); err != nil {
		return fmt.Errorf("check bamboo-hr: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bambooStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "employees"
	}
	endpoint, ok := bambooStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("bamboo-hr stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := bambooPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := bambooMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest reads an endpoint and emits its mapped records. Flat BambooHR meta
// endpoints return their entire list in a single response, so they are read with
// one request. The employees directory is read with a page/limit loop that stops
// when a short page (fewer than pageSize records) is returned — the standard
// "advance until a partial page" pattern, kept in-package because BambooHR has no
// next-token or Link-header cursor for connsdk's paginators to follow.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	if !endpoint.paginated {
		return c.readPage(ctx, r, endpoint, nil, emit)
	}

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("limit", strconv.Itoa(pageSize))
		count, err := c.readPageCount(ctx, r, endpoint, query, emit)
		if err != nil {
			return err
		}
		if count < pageSize {
			return nil
		}
	}
	return nil
}

// readPage reads a single response and emits its mapped records.
func (c Connector) readPage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query url.Values, emit func(connectors.Record) error) error {
	_, err := c.readPageCount(ctx, r, endpoint, query, emit)
	return err
}

// readPageCount reads a single response, emits its mapped records, and returns the
// number of raw records seen (used by the pagination loop to detect a short page).
func (c Connector) readPageCount(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query url.Values, emit func(connectors.Record) error) (int, error) {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return 0, fmt.Errorf("read bamboo-hr %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return 0, fmt.Errorf("decode bamboo-hr %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bamboo-hr credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%d", i),
			"fieldId":       fmt.Sprintf("%d", i),
			"displayName":   fmt.Sprintf("Fixture Employee %d", i),
			"firstName":     fmt.Sprintf("Fixture%d", i),
			"lastName":      "Employee",
			"preferredName": fmt.Sprintf("Fix%d", i),
			"jobTitle":      "Engineer",
			"department":    "Engineering",
			"division":      "Product",
			"location":      "Remote",
			"workEmail":     fmt.Sprintf("fixture+%d@example.com", i),
			"workPhone":     "555-0100",
			"mobilePhone":   "555-0200",
			"supervisor":    "Fixture Manager",
			"photoUrl":      "https://example.com/photo.png",
			"name":          fmt.Sprintf("Fixture %s %d", stream, i),
			"type":          "text",
			"alias":         fmt.Sprintf("alias%d", i),
			"deprecated":    false,
			"manageable":    true,
			"multiple":      false,
			"options":       []any{},
			"units":         "days",
			"color":         "#00aa00",
			"icon":          "time-off",
		}
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The api_key only ever flows into connsdk.Basic; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := bambooBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := bambooSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("bamboo-hr connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, bambooBasicPassword),
		UserAgent: bambooUserAgent,
		Accept:    "application/json",
	}, nil
}

func bambooSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// bambooBaseURL resolves and validates the base URL. The default is derived from
// the required subdomain config (https://<subdomain>.bamboohr.com/api/v1). Any
// explicit base_url override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func bambooBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["base_url"]); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("bamboo-hr config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("bamboo-hr config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("bamboo-hr config base_url must include a host")
		}
		return strings.TrimRight(override, "/"), nil
	}

	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("bamboo-hr connector requires config subdomain (or base_url override)")
	}
	if !validSubdomain(subdomain) {
		return "", fmt.Errorf("bamboo-hr config subdomain %q is invalid", subdomain)
	}
	return fmt.Sprintf("https://%s.bamboohr.com/api/v1", subdomain), nil
}

// validSubdomain restricts the subdomain to DNS-label characters so it cannot be
// abused to redirect requests to an arbitrary host.
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

func bambooPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return bambooDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bamboo-hr config page_size must be an integer: %w", err)
	}
	if value < 1 || value > bambooMaxPageSize {
		return 0, fmt.Errorf("bamboo-hr config page_size must be between 1 and %d", bambooMaxPageSize)
	}
	return value, nil
}

func bambooMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bamboo-hr config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("bamboo-hr config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: BambooHR is exposed as a read-only HR source. It satisfies
// the connectors.Connector interface while signalling that reverse-ETL is not
// available for this connector.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
