// Package amplitude implements the native pm Amplitude connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction) with Amplitude-specific stream definitions and
// endpoints.
//
// Amplitude's Analytics REST API authenticates with HTTP Basic auth using the
// project API key as the username and the secret key as the password. The
// connector reads a core set of full-refresh list endpoints (cohorts,
// annotations, event types). It is read-only: this is an analytics API with no
// safe reverse-ETL write surface.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package amplitude

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
	// amplitudeStandardBaseURL is the default Analytics REST API host.
	amplitudeStandardBaseURL = "https://amplitude.com"
	// amplitudeEUBaseURL serves EU-residency projects.
	amplitudeEUBaseURL = "https://analytics.eu.amplitude.com"
	amplitudeUserAgent = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("amplitude", New)
}

// New returns the Amplitude connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Amplitude connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "amplitude" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "amplitude",
		DisplayName:     "Amplitude",
		IntegrationType: "api",
		Description:     "Reads Amplitude behavioral cohorts, chart annotations, and event types through the Amplitude Analytics REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Amplitude.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := amplitudeBaseURL(cfg); err != nil {
		return err
	}
	apiKey, secretKey := amplitudeSecrets(cfg)
	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(secretKey) == "" {
		return errors.New("amplitude connector requires secrets api_key and secret_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of cohorts confirms auth and connectivity without mutating
	// anything (Amplitude has no read-only ping endpoint; cohorts is cheap).
	if err := r.DoJSON(ctx, http.MethodGet, "api/3/cohorts", nil, nil, nil); err != nil {
		return fmt.Errorf("check amplitude: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: amplitudeStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "cohorts"
	}
	endpoint, ok := amplitudeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("amplitude stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.read(ctx, r, endpoint, emit)
}

// Write is unsupported: Amplitude is exposed read-only here (analytics API with
// no safe reverse-ETL write surface). It satisfies the Connector interface and
// always returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// read performs a single GET against the stream's list endpoint and emits each
// mapped record. Amplitude's list endpoints (cohorts, annotations, events list)
// return the full collection in one response, so there is no pagination loop.
func (c Connector) read(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read amplitude %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode amplitude %s: %w", endpoint.resource, err)
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

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise amplitude credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"value":       fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":        fmt.Sprintf("Fixture %d", i),
			"display":     fmt.Sprintf("Fixture %d", i),
			"label":       fmt.Sprintf("Fixture %d", i),
			"description": "fixture record",
			"details":     "fixture record",
			"date":        "2026-01-01",
			"size":        int64(10 * i),
			"totals":      int64(100 * i),
			"archived":    false,
			"published":   true,
			"hidden":      false,
			"non_active":  false,
			"deleted":     false,
			"flow_hidden": false,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth
// (api_key:secret_key) and the resolved base URL. The secrets only ever flow
// into connsdk.Basic; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := amplitudeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey, secretKey := amplitudeSecrets(cfg)
	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(secretKey) == "" {
		return nil, errors.New("amplitude connector requires secrets api_key and secret_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(apiKey, secretKey),
		UserAgent: amplitudeUserAgent,
	}, nil
}

func amplitudeSecrets(cfg connectors.RuntimeConfig) (apiKey, secretKey string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["api_key"], cfg.Secrets["secret_key"]
}

// amplitudeBaseURL resolves and validates the base URL. The default is
// amplitude.com; setting data_region to "EU Residency Server" selects the EU
// host. An explicit base_url override wins and must be an absolute http(s) URL
// with a host to bound SSRF risk.
func amplitudeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if isEURegion(cfg.Config["data_region"]) {
			return amplitudeEUBaseURL, nil
		}
		return amplitudeStandardBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("amplitude config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("amplitude config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("amplitude config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func isEURegion(region string) bool {
	region = strings.ToLower(strings.TrimSpace(region))
	return strings.Contains(region, "eu")
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
