// Package openweather implements the native pm OpenWeather connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + APIKeyQuery auth +
// RecordsAt extraction) with OpenWeather-specific stream definitions.
//
// It reads the OpenWeather One Call API 3.0, whose single /onecall endpoint
// returns one JSON document with current/hourly/daily/alerts sections. Because
// the API is not paginated, the connector instead iterates over one or more
// configured geographic locations, issuing one request per location; each
// location's array sections (hourly/daily/alerts) yield multiple records.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package openweather

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
	openweatherDefaultBaseURL = "https://api.openweathermap.org/data/3.0"
	openweatherEndpoint       = "onecall"
	openweatherUserAgent      = "polymetrics-go-cli"
	// openweatherFixtureDT is the deterministic unix `dt` used by fixture records
	// (2026-01-01T00:00:00Z).
	openweatherFixtureDT int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("openweather", New)
}

// New returns the OpenWeather connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm OpenWeather connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "openweather" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "openweather",
		DisplayName:     "OpenWeather",
		IntegrationType: "api",
		Description:     "Reads current weather, hourly and daily forecasts, and government alerts for one or more geographic locations from the OpenWeather One Call API 3.0.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to OpenWeather.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := openweatherBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(openweatherSecret(cfg)) == "" {
		return errors.New("openweather connector requires secret appid")
	}
	locations, err := openweatherLocations(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A single bounded read confirms auth and connectivity without mutating
	// anything (the API is read-only regardless).
	query := locationQuery(locations[0], cfg)
	query.Set("exclude", "minutely,hourly,daily,alerts")
	if err := r.DoJSON(ctx, http.MethodGet, openweatherEndpoint, query, nil, nil); err != nil {
		return fmt.Errorf("check openweather: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: openweatherStreams()}, nil
}

// Write satisfies the connectors.Connector interface. OpenWeather is read-only,
// so reverse-ETL writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "current"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("openweather stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, spec, emit)
	}

	locations, err := openweatherLocations(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, spec, locations, req.Config, emit)
}

// harvest iterates over the configured locations, issuing one /onecall request
// per location. The One Call API is not paginated, so "pages" here are
// locations: each response's stream section (a single object for current, an
// array otherwise) is mapped and emitted, annotated with the source lat/lon and
// timezone so downstream rows are self-identifying.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, spec streamSpec, locations []location, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	for _, loc := range locations {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := locationQuery(loc, cfg)
		resp, err := r.Do(ctx, http.MethodGet, openweatherEndpoint, query, nil)
		if err != nil {
			return fmt.Errorf("read openweather %s: %w", spec.jsonKey, err)
		}
		// timezone is a sibling top-level field, useful context per row.
		timezone, _ := connsdk.StringAt(resp.Body, "timezone")
		records, err := connsdk.RecordsAt(resp.Body, spec.jsonKey)
		if err != nil {
			return fmt.Errorf("decode openweather %s: %w", spec.jsonKey, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			out := spec.mapRecord(item)
			annotate(out, loc, timezone)
			if err := emit(out); err != nil {
				return err
			}
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise openweather credential-free.
func (c Connector) readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	loc := location{lat: "33.44", lon: "-94.04"}
	count := 2
	if spec.single {
		count = 1
	}
	for i := 0; i < count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureItem(spec.jsonKey, i)
		out := spec.mapRecord(item)
		annotate(out, loc, "America/Chicago")
		out["fixture"] = true
		if err := emit(out); err != nil {
			return err
		}
	}
	return nil
}

// fixtureItem builds a deterministic raw One Call object for the given section.
func fixtureItem(jsonKey string, i int) map[string]any {
	weather := []any{map[string]any{"id": 803, "main": "Clouds", "description": "broken clouds", "icon": "04d"}}
	dt := openweatherFixtureDT + int64(i*3600)
	base := map[string]any{
		"dt":         dt,
		"temp":       292.55 + float64(i),
		"feels_like": 292.87,
		"pressure":   1014,
		"humidity":   89,
		"dew_point":  290.9,
		"uvi":        0.0,
		"clouds":     75,
		"visibility": 10000,
		"wind_speed": 3.13,
		"wind_deg":   93,
		"wind_gust":  5.2,
		"pop":        0.1,
		"weather":    weather,
	}
	switch jsonKey {
	case "daily":
		base["summary"] = "Expect a day of partly cloudy weather"
		base["temp"] = map[string]any{"day": 299.03, "min": 290.69, "max": 300.35}
		base["sunrise"] = dt
		base["sunset"] = dt + 43200
	case "alerts":
		return map[string]any{
			"sender_name": "NWS",
			"event":       "Heat Advisory",
			"start":       dt,
			"end":         dt + 7200,
			"description": "Stay hydrated and avoid the midday sun.",
			"tags":        []any{"Extreme temperature value"},
		}
	case "current":
		base["sunrise"] = dt
		base["sunset"] = dt + 43200
	}
	return base
}

// requester builds a connsdk.Requester wired with APIKeyQuery auth (appid=...)
// and the resolved base URL. The secret only ever flows into connsdk.APIKeyQuery;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := openweatherBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := openweatherSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("openweather connector requires secret appid")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("appid", secret),
		UserAgent: openweatherUserAgent,
	}, nil
}

// locationQuery builds the per-request query (lat, lon, optional units/lang)
// for one location. The appid is added by the APIKeyQuery authenticator.
func locationQuery(loc location, cfg connectors.RuntimeConfig) url.Values {
	query := url.Values{}
	query.Set("lat", loc.lat)
	query.Set("lon", loc.lon)
	if units := strings.TrimSpace(cfg.Config["units"]); units != "" {
		query.Set("units", units)
	}
	if lang := strings.TrimSpace(cfg.Config["lang"]); lang != "" {
		query.Set("lang", lang)
	}
	return query
}

// annotate stamps the source location and timezone onto an emitted record so
// rows from multiple locations stay distinguishable.
func annotate(rec connectors.Record, loc location, timezone string) {
	rec["lat"] = loc.lat
	rec["lon"] = loc.lon
	if timezone != "" {
		rec["timezone"] = timezone
	}
}

// location is a single geographic coordinate pair the connector reads.
type location struct {
	lat string
	lon string
}

func openweatherSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["appid"]
}

// openweatherLocations resolves the set of coordinates to read. It accepts
// either a single lat/lon pair (config lat + lon) or a semicolon-separated
// "locations" list of "lat,lon" pairs, enabling the multi-request read loop.
func openweatherLocations(cfg connectors.RuntimeConfig) ([]location, error) {
	if raw := strings.TrimSpace(cfg.Config["locations"]); raw != "" {
		var locs []location
		for _, part := range strings.Split(raw, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			lat, lon, ok := strings.Cut(part, ",")
			lat = strings.TrimSpace(lat)
			lon = strings.TrimSpace(lon)
			if !ok || lat == "" || lon == "" {
				return nil, fmt.Errorf("openweather config locations entry %q must be \"lat,lon\"", part)
			}
			locs = append(locs, location{lat: lat, lon: lon})
		}
		if len(locs) == 0 {
			return nil, errors.New("openweather config locations is empty")
		}
		return locs, nil
	}

	lat := strings.TrimSpace(cfg.Config["lat"])
	lon := strings.TrimSpace(cfg.Config["lon"])
	if lat == "" || lon == "" {
		return nil, errors.New("openweather requires config lat and lon, or a locations list")
	}
	return []location{{lat: lat, lon: lon}}, nil
}

// openweatherBaseURL resolves and validates the base URL. The default is
// api.openweathermap.org; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func openweatherBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return openweatherDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("openweather config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("openweather config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("openweather config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
