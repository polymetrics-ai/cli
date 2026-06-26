// Package nexiopay implements the native pm Nexio Pay connector. It is a
// declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + Basic auth + RecordsAt extraction + offset pagination) with
// Nexio-specific stream definitions and endpoints, mirroring the stripe
// reference connector.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Nexio authenticates with HTTP Basic auth using the API username and API key
// (password). The connector is read-only: Nexio's reporting/management endpoints
// have no safe reverse-ETL write surface, so Capabilities.Write is false.
package nexiopay

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
	nexioDefaultSubdomain = "nexiopay"
	nexioDefaultPageSize  = 10
	nexioMaxPageSize      = 100
	nexioUserAgent        = "polymetrics-go-cli"
	nexioLimitParam       = "limit"
	nexioOffsetParam      = "offset"
)

func init() {
	connectors.RegisterFactory("nexiopay", New)
}

// New returns the Nexio Pay connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Nexio Pay connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nexiopay" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nexiopay",
		DisplayName:     "Nexio Pay",
		IntegrationType: "api",
		Description:     "Reads Nexio Pay card tokens, payout recipients, spendbacks, payment types, terminals, and the API user via the Nexio REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Nexio. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nexioBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(nexioAPIKey(cfg)) == "" {
		return errors.New("nexiopay connector requires secret api_key")
	}
	if strings.TrimSpace(nexioUsername(cfg)) == "" {
		return errors.New("nexiopay connector requires username")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// whoAmI is an unpaginated, side-effect-free identity endpoint: a clean way
	// to confirm auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "user/v3/account/whoAmI", nil, nil, nil); err != nil {
		return fmt.Errorf("check nexiopay: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nexioStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "card_tokens"
	}
	endpoint, ok := nexioStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("nexiopay stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := nexioPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := nexioMaxPages(req.Config)
	if err != nil {
		return err
	}

	// mapAndEmit adapts a raw connsdk record (map[string]any) through the
	// stream's mapper into a connectors.Record before emitting.
	mapAndEmit := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}

	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
		if err != nil {
			return fmt.Errorf("read nexiopay %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode nexiopay %s: %w", stream, err)
		}
		for _, rec := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := mapAndEmit(rec); err != nil {
				return err
			}
		}
		return nil
	}

	paginator := &connsdk.OffsetPaginator{
		LimitParam:  nexioLimitParam,
		OffsetParam: nexioOffsetParam,
		PageSize:    pageSize,
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, paginator, endpoint.recordsPath, maxPages, mapAndEmit); err != nil {
		return fmt.Errorf("read nexiopay %s: %w", stream, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise nexiopay credential-free.
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			endpoint.primaryKey: id,
			"name":              fmt.Sprintf("Fixture %d", i),
			"displayName":       fmt.Sprintf("Fixture %d", i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"status":            "active",
			"enabled":           true,
			"currency":          "USD",
			"cardType":          "visa",
			"lastFour":          "4242",
			"recipientId":       "rec_fixture_1",
			"merchantId":        "merch_fixture_1",
			"username":          "fixture_user",
			"role":              "api",
			"amount":            float64(100 * i),
			"createdDate":       "2026-01-01T00:00:00Z",
			"updatedDate":       "2026-01-01T00:00:00Z",
			"fixture":           true,
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

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The api_key only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := nexioBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := nexioAPIKey(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("nexiopay connector requires secret api_key")
	}
	username := nexioUsername(cfg)
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("nexiopay connector requires username")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, apiKey),
		UserAgent: nexioUserAgent,
	}, nil
}

// nexioAPIKey resolves the API key (Basic auth password) from secrets.
func nexioAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// nexioUsername resolves the API username. It is part of the credential pair, so
// it is read from secrets first (where the runtime injects it alongside api_key)
// and falls back to plain config.
func nexioUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["username"]); v != "" {
			return v
		}
	}
	if cfg.Config != nil {
		return strings.TrimSpace(cfg.Config["username"])
	}
	return ""
}

// nexioBaseURL resolves and validates the base URL. The default is derived from
// the subdomain config (https://api.<subdomain>.com). Any base_url override must
// be an absolute https (or http for local test servers) URL with a host to bound
// SSRF risk.
func nexioBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	override := ""
	if cfg.Config != nil {
		override = strings.TrimSpace(cfg.Config["base_url"])
	}
	if override == "" {
		subdomain := nexioDefaultSubdomain
		if cfg.Config != nil {
			if s := strings.TrimSpace(cfg.Config["subdomain"]); s != "" {
				subdomain = s
			}
		}
		if !validSubdomain(subdomain) {
			return "", fmt.Errorf("nexiopay config subdomain %q is invalid", subdomain)
		}
		return "https://api." + subdomain + ".com", nil
	}
	parsed, err := url.Parse(override)
	if err != nil {
		return "", fmt.Errorf("nexiopay config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nexiopay config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nexiopay config base_url must include a host")
	}
	return strings.TrimRight(override, "/"), nil
}

// validSubdomain bounds the subdomain to a conservative alnum set so it cannot
// be used to inject a different host into the derived base URL.
func validSubdomain(s string) bool {
	if s == "" || len(s) > 63 {
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

func nexioPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return nexioDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nexiopay config page_size must be an integer: %w", err)
	}
	if value < 1 || value > nexioMaxPageSize {
		return 0, fmt.Errorf("nexiopay config page_size must be between 1 and %d", nexioMaxPageSize)
	}
	return value, nil
}

func nexioMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	}
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nexiopay config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("nexiopay config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: the Nexio connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}
