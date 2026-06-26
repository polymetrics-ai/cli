// Package interzoid implements the native pm Interzoid connector. Interzoid is
// an AI-powered data-quality / data-matching API: each "stream" is a single
// lookup endpoint that, given an input value (a company name, person name,
// street address, or organization name), returns one JSON object containing a
// similarity key (SimKey) or a standardized value (Standard) plus a remaining
// credit count.
//
// It follows the declarative-HTTP shape of the stripe reference connector: a
// thin package composing the connsdk toolkit (Requester + APIKeyQuery auth +
// RecordsAt extraction) with Interzoid-specific endpoints and record mappers.
// Unlike most sources, the Interzoid lookups are read-only, single-record, and
// have no pagination or incremental cursor, so the read path is a one-shot GET.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package interzoid

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
	interzoidDefaultBaseURL = "https://api.interzoid.com"
	interzoidUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("interzoid", New)
}

// New returns the Interzoid connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Interzoid connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "interzoid" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "interzoid",
		DisplayName:     "Interzoid",
		IntegrationType: "api",
		Description:     "Reads Interzoid data-matching lookups: company-name, individual-name, and street-address similarity keys, plus organization-name standardization, via the Interzoid REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Interzoid.
// In fixture mode it short-circuits without a network call. Otherwise it
// requires the api_key secret and a valid base URL; it does not spend an API
// credit by performing a live lookup.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := interzoidBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(interzoidSecret(cfg)) == "" {
		return errors.New("interzoid connector requires secret api_key")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: interzoidStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "company_name_matching"
	}
	endpoint, ok := interzoidStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("interzoid stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	query, echo, err := buildInputs(endpoint, req.Config)
	if err != nil {
		return err
	}

	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read interzoid %s: %w", endpoint.resource, err)
	}
	// Interzoid returns a single JSON object at the response root; RecordsAt with
	// an empty path yields that object as a one-element record set.
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode interzoid %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item, echo)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits a deterministic record without any network access so the
// conformance harness can exercise interzoid credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	item := map[string]any{
		"Code":     "Success",
		"SimKey":   "FIXTUREKEY0001",
		"Standard": "FIXTURE STANDARD",
		"Credits":  "9999",
	}
	echo := map[string]string{
		"company":  "Fixture Company Inc",
		"fullname": "Fixture Person",
		"address":  "1 Fixture Way",
		"org":      "Fixture Organization",
	}
	return emit(endpoint.mapRecord(item, echo))
}

// buildInputs assembles the query parameters from the stream's declared inputs
// and returns the echo map (config values keyed by config key) used by record
// mappers to attach the originating input to each record. A missing required
// input is an error; the api_key is injected separately by the authenticator.
func buildInputs(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (url.Values, map[string]string, error) {
	query := url.Values{}
	echo := map[string]string{}
	for _, in := range endpoint.inputs {
		value := ""
		if cfg.Config != nil {
			value = strings.TrimSpace(cfg.Config[in.configKey])
		}
		if value == "" {
			if in.required {
				return nil, nil, fmt.Errorf("interzoid stream requires config %q", in.configKey)
			}
			continue
		}
		query.Set(in.paramName, value)
		echo[in.configKey] = value
	}
	return query, echo, nil
}

// requester builds a connsdk.Requester wired with APIKeyQuery auth (the api_key
// is injected as the `license` query parameter, matching the Interzoid API) and
// the resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := interzoidBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := interzoidSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("interzoid connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("license", secret),
		UserAgent: interzoidUserAgent,
	}, nil
}

func interzoidSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// interzoidBaseURL resolves and validates the base URL. The default is
// api.interzoid.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func interzoidBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return interzoidDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("interzoid config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("interzoid config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("interzoid config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// Write satisfies the connectors.Connector interface. Interzoid is a read-only
// data-matching API with no reverse-ETL surface, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
