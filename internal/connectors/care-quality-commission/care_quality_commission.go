// Package carequalitycommission implements the native pm Care Quality Commission
// (CQC) connector. It reads the public CQC Syndication API
// (https://api.service.cqc.org.uk/public/v1) for the core top-level streams
// (locations, providers, inspection_areas).
//
// It follows the stripe declarative-HTTP template: a thin package composing the
// connsdk toolkit (Requester + APIKeyHeader auth + RecordsAt extraction) with
// CQC-specific stream definitions and a page-increment read loop. The CQC API
// is read-only (full-refresh only), so the connector exposes no write actions.
//
// The package name is carequalitycommission (Go identifiers cannot contain
// hyphens) but the registry key and Name() are the bare system name
// "care-quality-commission". It self-registers via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package carequalitycommission

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
	// connectorName is the bare system name used as the registry key.
	connectorName = "care-quality-commission"
	// cqcDefaultBaseURL is the public CQC Syndication API root.
	cqcDefaultBaseURL = "https://api.service.cqc.org.uk/public/v1"
	// cqcDefaultPageSize matches the upstream connector's perPage of 1000.
	cqcDefaultPageSize = 1000
	cqcMaxPageSize     = 1000
	// cqcSubscriptionHeader is the Azure API Management header the CQC API uses
	// to carry the primary subscription key.
	cqcSubscriptionHeader = "Ocp-Apim-Subscription-Key"
	cqcUserAgent          = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the CQC connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Care Quality Commission connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Care Quality Commission",
		IntegrationType: "api",
		Description:     "Reads Care Quality Commission (CQC) registered locations, providers, and inspection areas from the public CQC Syndication API. Read-only (full-refresh).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the CQC API.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := cqcBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cqcSecret(cfg)) == "" {
		return errors.New("care-quality-commission connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the inspection-areas endpoint confirms auth and
	// connectivity without listing the large locations/providers tables.
	if err := r.DoJSON(ctx, http.MethodGet, "inspection-areas", nil, nil, nil); err != nil {
		return fmt.Errorf("check care-quality-commission: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: cqcStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "locations"
	}
	endpoint, ok := cqcStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("care-quality-commission stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := cqcPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := cqcMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives the CQC page-increment pagination. The list endpoints accept
// page (1-based) + perPage and return {page, perPage, totalPages, total,
// <records>:[...]}. The loop stops when totalPages is reached, when a short page
// (< perPage records) is returned, or when an empty page is returned. Unpaginated
// endpoints (inspection_areas) are read in a single request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}

		query := url.Values{}
		if endpoint.paginated {
			query.Set("page", strconv.Itoa(page))
			query.Set("perPage", strconv.Itoa(pageSize))
		}

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read care-quality-commission %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode care-quality-commission %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Non-paginated endpoints are exhausted after the first request.
		if !endpoint.paginated {
			return nil
		}
		// Stop on a short or empty page (fewer than the requested page size).
		if len(records) < pageSize {
			return nil
		}
		// Honour the reported totalPages when present (string-compared to be
		// resilient to json.Number vs float encoding).
		if total := strings.TrimSpace(stringAtTotalPages(resp.Body)); total != "" {
			if tp, perr := strconv.Atoi(total); perr == nil && page >= tp {
				return nil
			}
		}
	}
}

// stringAtTotalPages reads the totalPages field from the response envelope, or
// "" when absent. Errors are swallowed: the short-page heuristic remains a safe
// fallback stop condition.
func stringAtTotalPages(body []byte) string {
	v, err := connsdk.StringAt(body, "totalPages")
	if err != nil {
		return ""
	}
	return v
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"locationId":         fmt.Sprintf("1-fixture-loc-%d", i),
			"locationName":       fmt.Sprintf("Fixture Care Home %d", i),
			"postalCode":         fmt.Sprintf("AB%d 2CD", i),
			"providerId":         fmt.Sprintf("1-fixture-prov-%d", i),
			"providerName":       fmt.Sprintf("Fixture Healthcare %d Ltd", i),
			"inspectionAreaId":   fmt.Sprintf("IA-fixture-%d", i),
			"inspectionAreaName": fmt.Sprintf("Fixture Key Question %d", i),
			"inspectionAreaType": "key_question",
			"status":             "active",
			"connector":          connectorName,
			"fixture":            true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the CQC subscription-key header
// auth and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := cqcBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := cqcSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("care-quality-commission connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(cqcSubscriptionHeader, secret, ""),
		UserAgent: cqcUserAgent,
	}, nil
}

func cqcSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// cqcBaseURL resolves and validates the base URL. The default is
// api.service.cqc.org.uk; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func cqcBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return cqcDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("care-quality-commission config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("care-quality-commission config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("care-quality-commission config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cqcPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return cqcDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("care-quality-commission config page_size must be an integer: %w", err)
	}
	if value < 1 || value > cqcMaxPageSize {
		return 0, fmt.Errorf("care-quality-commission config page_size must be between 1 and %d", cqcMaxPageSize)
	}
	return value, nil
}

func cqcMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("care-quality-commission config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("care-quality-commission config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. The CQC API is read-only,
// so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
