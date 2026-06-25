// Package breezometer implements the native pm BreezoMeter connector. It follows
// the stripe declarative-HTTP template: a thin package that composes the connsdk
// toolkit (Requester + APIKeyQuery auth + RecordsAt extraction) with
// BreezoMeter-specific stream definitions, endpoints, and record mappers.
//
// BreezoMeter (now part of Google Maps Platform's Environment APIs) exposes
// point-in-time environmental data (air quality, pollen, weather) for a single
// lat/lon location. Authentication is an API key supplied as the `key` query
// parameter. Forecast/history endpoints return a `data` array of time-series
// points; current-conditions endpoints return a single `data` object. This
// connector is read-only — there is no safe reverse-ETL write surface — so
// Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package breezometer

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
	breezometerDefaultBaseURL = "https://api.breezometer.com"
	breezometerUserAgent      = "polymetrics-go-cli"
	// breezometerMaxPages bounds an unconfigured paginated read so a misbehaving
	// next_page_token cannot loop forever.
	breezometerDefaultMaxPages = 50
)

func init() {
	connectors.RegisterFactory("breezometer", New)
}

// New returns the BreezoMeter connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm BreezoMeter connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "breezometer" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "breezometer",
		DisplayName:     "Breezometer",
		IntegrationType: "api",
		Description:     "Reads BreezoMeter (Google Environment) air quality, pollen, and weather conditions and forecasts for a configured location via the BreezoMeter REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to BreezoMeter.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := breezometerBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(breezometerSecret(cfg)) == "" {
		return errors.New("breezometer connector requires secret api_key")
	}
	lat, lng, err := breezometerLocation(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded current-conditions read confirms auth, the location, and
	// connectivity without mutating anything.
	q := url.Values{"lat": []string{lat}, "lon": []string{lng}, "features": []string{"local_aqi"}}
	if err := r.DoJSON(ctx, http.MethodGet, "air-quality/v2/current-conditions", q, nil, nil); err != nil {
		return fmt.Errorf("check breezometer: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: breezometerStreams()}, nil
}

// Write is unsupported: BreezoMeter is a read-only environmental data source
// with no safe reverse-ETL surface. It satisfies the connectors.Connector
// interface by returning ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a BreezoMeter stream starts
// with an empty incremental cursor (full refresh), which read time can raise.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "air_quality_current"
	}
	endpoint, ok := breezometerStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("breezometer stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	lat, lng, err := breezometerLocation(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := breezometerMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	base.Set("lat", lat)
	base.Set("lon", lng)
	if hours := strings.TrimSpace(req.Config.Config["hours_to_forecast"]); hours != "" {
		base.Set("hours", hours)
	}
	if hours := strings.TrimSpace(req.Config.Config["historic_hours"]); hours != "" {
		base.Set("hours", hours)
	}
	if days := strings.TrimSpace(req.Config.Config["days_to_forecast"]); days != "" {
		base.Set("days", days)
	}
	return c.harvest(ctx, r, endpoint, base, lat, lng, maxPages, emit)
}

// harvest drives BreezoMeter's optional next_page_token pagination. A response is
// {data:[...]|{...}, next_page_token?:string}; the next page is requested with
// page_token=<token>. Current-conditions endpoints return a single object and no
// token, so the loop emits one record and stops. There is no body-token
// paginator in connsdk for this exact field name, so the loop lives here, built
// on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, lat, lng string, maxPages int, emit func(connectors.Record) error) error {
	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if pageToken != "" {
			query.Set("page_token", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read breezometer %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode breezometer %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			record := endpoint.mapRecord(item)
			record["latitude"] = lat
			record["longitude"] = lng
			if err := emit(record); err != nil {
				return err
			}
		}
		nextToken, err := connsdk.StringAt(resp.Body, "next_page_token")
		if err != nil {
			return fmt.Errorf("decode breezometer %s next_page_token: %w", endpoint.resource, err)
		}
		// Single-object streams never list; stop after the first response.
		if !endpoint.list || strings.TrimSpace(nextToken) == "" {
			return nil
		}
		pageToken = nextToken
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise breezometer credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	count := 1
	if endpoint.list {
		count = 2
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"datetime": fmt.Sprintf("2026-01-01T%02d:00:00Z", i-1),
			"date":     "2026-01-01",
			"indexes": map[string]any{
				"baqi": map[string]any{"aqi": 40 + i, "category": "Good air quality"},
			},
			"pollutants": map[string]any{"co": map[string]any{"concentration": map[string]any{"value": 100 + i}}},
			"types":      map[string]any{"grass": map[string]any{"index": map[string]any{"value": i}}},
			"plants":     map[string]any{},
			"index":      map[string]any{"value": i, "category": "Low"},
			"temperature": map[string]any{
				"value": 20 + i, "units": "celsius",
			},
			"relative_humidity": 50 + i,
			"wind":              map[string]any{"speed": map[string]any{"value": i}},
			"precipitation":     map[string]any{},
			"weather_condition": map[string]any{"text": "Clear"},
			"data_available":    true,
		}
		record := endpoint.mapRecord(item)
		record["latitude"] = "54.675003"
		record["longitude"] = "-113.550282"
		record["connector"] = "breezometer"
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with API-key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := breezometerBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := breezometerSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("breezometer connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("key", secret),
		UserAgent: breezometerUserAgent,
	}, nil
}

func breezometerSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// breezometerLocation resolves and validates the required lat/lon config.
func breezometerLocation(cfg connectors.RuntimeConfig) (string, string, error) {
	lat := strings.TrimSpace(cfg.Config["latitude"])
	lng := strings.TrimSpace(cfg.Config["longitude"])
	if lat == "" || lng == "" {
		return "", "", errors.New("breezometer connector requires config latitude and longitude")
	}
	if _, err := strconv.ParseFloat(lat, 64); err != nil {
		return "", "", fmt.Errorf("breezometer config latitude must be numeric: %w", err)
	}
	if _, err := strconv.ParseFloat(lng, 64); err != nil {
		return "", "", fmt.Errorf("breezometer config longitude must be numeric: %w", err)
	}
	return lat, lng, nil
}

// breezometerBaseURL resolves and validates the base URL. The default is
// api.breezometer.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func breezometerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return breezometerDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("breezometer config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("breezometer config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("breezometer config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func breezometerMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return breezometerDefaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("breezometer config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("breezometer config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
