// Package opendatadc implements the native pm Open Data DC connector. It reads
// the District of Columbia Master Address Repository (MAR 2) API, a read-only
// address/place/block lookup service.
//
// It follows the declarative-HTTP template established by the stripe connector:
// a thin package composing the connsdk toolkit (Requester + APIKeyQuery auth +
// RecordsAt extraction) with MAR-specific stream definitions and record
// mappers. Like the other per-system connectors it self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
//
// The directory name is open-data-dc (matching the catalog slug and registry
// key); the Go package identifier is sanitized to opendatadc.
package opendatadc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://datagate.dc.gov/mar/open/api/v2.2"
	userAgent      = "polymetrics-go-cli"
	// apiKeyParam is the query parameter the MAR API expects the key in.
	apiKeyParam = "apikey"
)

func init() {
	connectors.RegisterFactory("open-data-dc", New)
}

// New returns the Open Data DC connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Open Data DC connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "open-data-dc" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "open-data-dc",
		DisplayName:     "Open Data DC",
		IntegrationType: "api",
		Description:     "Reads District of Columbia Master Address Repository (MAR 2) locations, units, and SSL parcel records via the Open Data DC API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to the MAR API.
// In fixture mode it short-circuits without a network call.
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
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("open-data-dc connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded ssls lookup confirms auth and connectivity. marid may be empty;
	// the API still answers (with an empty/short result) which is enough to
	// validate the credential without mutating anything.
	query := url.Values{}
	if marid := strings.TrimSpace(cfg.Config["marid"]); marid != "" {
		query.Set("marid", marid)
	}
	if err := r.DoJSON(ctx, http.MethodGet, "ssls", query, nil, nil); err != nil {
		return fmt.Errorf("check open-data-dc: %w", err)
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
		stream = "locations"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("open-data-dc stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, query, err := requestFor(stream, req.Config)
	if err != nil {
		return err
	}

	// The MAR API is not paginated: each stream returns its full result set in a
	// single response under endpoint.recordsPath.
	resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return fmt.Errorf("read open-data-dc %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode open-data-dc %s: %w", stream, err)
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

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requestFor builds the path and query for a stream. locations embeds the
// location search term in the path; units embeds the marid; ssls passes marid
// as a query parameter.
func requestFor(stream string, cfg connectors.RuntimeConfig) (string, url.Values, error) {
	query := url.Values{}
	location := strings.TrimSpace(cfg.Config["location"])
	marid := strings.TrimSpace(cfg.Config["marid"])
	switch stream {
	case "locations":
		if location == "" {
			return "", nil, errors.New("open-data-dc locations stream requires config location (address, place, or block)")
		}
		return "locations/" + location, query, nil
	case "units":
		if marid == "" {
			return "", nil, errors.New("open-data-dc units stream requires config marid")
		}
		return "units/" + marid, query, nil
	case "ssls":
		if marid != "" {
			query.Set("marid", marid)
		}
		return "ssls", query, nil
	default:
		return "", nil, fmt.Errorf("open-data-dc stream %q not found", stream)
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise open-data-dc credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureItem(stream, i)
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// fixtureItem builds a deterministic raw item in the same shape the live API
// returns, so the same mappers exercise it.
func fixtureItem(stream string, i int) map[string]any {
	marID := fmt.Sprintf("fixture_%d", i)
	full := fmt.Sprintf("%d FIXTURE ST NW", 100*i)
	ssl := fmt.Sprintf("0%03d    080%d", 100+i, i)
	switch stream {
	case "locations":
		return map[string]any{
			"address": map[string]any{
				"properties": map[string]any{
					"MarId":         marID,
					"FullAddress":   full,
					"SSL":           ssl,
					"StName":        "FIXTURE ST",
					"AddrNum":       fmt.Sprintf("%d", 100*i),
					"Quadrant":      "NW",
					"Ward":          "Ward 2",
					"Anc":           "2A",
					"Zipcode":       "20004",
					"Status":        "ACTIVE",
					"ResidenceType": "RESIDENTIAL",
					"CensusTract":   "004701",
					"Latitude":      38.9 + float64(i)/1000,
					"Longitude":     -77.03 - float64(i)/1000,
					"Xcoord":        float64(396000 + i),
					"Ycoord":        float64(137000 + i),
				},
			},
			"distance": float64(i),
		}
	case "units":
		return map[string]any{
			"UnitNum":     fmt.Sprintf("%d", i),
			"MarId":       marID,
			"FullAddress": full,
			"UnitType":    "CONDO",
			"UnitSSL":     ssl,
			"Status":      "ACTIVE",
		}
	case "ssls":
		return map[string]any{
			"SSL":         ssl,
			"MarId":       marID,
			"FullAddress": full,
			"Square":      fmt.Sprintf("0%03d", 100+i),
			"Lot":         fmt.Sprintf("080%d", i),
			"Col":         "",
			"Lot_type":    "RECORD",
		}
	default:
		return map[string]any{"MarId": marID}
	}
}

// requester builds a connsdk.Requester wired with the api key in the apikey
// query parameter and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyQuery; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("open-data-dc connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery(apiKeyParam, key),
		UserAgent: userAgent,
	}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// baseURL resolves and validates the base URL. The default is datagate.dc.gov;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("open-data-dc config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("open-data-dc config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("open-data-dc config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
