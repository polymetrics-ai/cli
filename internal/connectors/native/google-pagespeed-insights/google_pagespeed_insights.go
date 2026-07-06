// Package googlepagespeedinsights implements the native pm Google PageSpeed
// Insights connector. It is a declarative-HTTP per-system connector built on the
// same shape as the stripe reference and the alpha-vantage connector (single
// endpoint, no pagination, query-param auth): a thin package that composes the
// connsdk toolkit (Requester + api-key query auth) with PageSpeed-specific stream
// definitions.
//
// PageSpeed Insights exposes one operation: GET /runPagespeed on
// www.googleapis.com/pagespeedonline/v5. Each request analyzes a single page for
// a single strategy (desktop or mobile) and one or more Lighthouse categories,
// returning ONE report object — there is no list endpoint, no record array, and
// no pagination. The connector therefore synthesizes a single report stream by
// iterating the cartesian product of the configured `urls` and `strategies`,
// issuing one request per (url, strategy) pair and flattening each report into a
// record. Auth is the optional `key` query param (the API works keyless but is
// heavily throttled; the secret is api_key).
//
// The connector is read-only: PageSpeed Insights exposes no reverse-ETL writes,
// so Capabilities.Write is false.
package googlepagespeedinsights

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	pagespeedConnectorName  = "google-pagespeed-insights"
	pagespeedDefaultBaseURL = "https://www.googleapis.com/pagespeedonline/v5"
	pagespeedRunPath        = "runPagespeed"
	pagespeedUserAgent      = "polymetrics-go-cli"

	pagespeedDefaultStrategy = "desktop"
	pagespeedDefaultCategory = "performance"
	pagespeedDefaultURL      = "https://example.com"
)

// New returns the PageSpeed Insights connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google PageSpeed Insights connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return pagespeedConnectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            pagespeedConnectorName,
		DisplayName:     "Google PageSpeed Insights",
		IntegrationType: "api",
		Description:     "Reads Lighthouse PageSpeed Insights reports (performance, accessibility, best-practices, SEO, PWA scores) for the configured URLs and strategies via the PageSpeed Insights v5 API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to PageSpeed
// Insights. In fixture mode it short-circuits without a network call. Outside
// fixture mode it runs one bounded report against the first configured URL,
// which confirms connectivity (and the key, if supplied) without a list scan.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := pagespeedBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	urls := pagespeedURLs(cfg)
	strategies := pagespeedStrategies(cfg)
	categories := pagespeedCategories(cfg)
	query := url.Values{}
	query.Set("url", urls[0])
	query.Set("strategy", strategies[0])
	query.Set("category", categories[0])
	if err := r.DoJSON(ctx, http.MethodGet, pagespeedRunPath, query, nil, nil); err != nil {
		return fmt.Errorf("check google-pagespeed-insights: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: pagespeedStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = streamName
	}
	if stream != streamName {
		return fmt.Errorf("google-pagespeed-insights stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	urls := pagespeedURLs(req.Config)
	strategies := pagespeedStrategies(req.Config)
	categories := pagespeedCategories(req.Config)

	// One request per (url, strategy) pair; categories are passed together so a
	// single report carries every requested Lighthouse category.
	for _, analyzedURL := range urls {
		for _, strategy := range strategies {
			if err := ctx.Err(); err != nil {
				return err
			}
			query := url.Values{}
			query.Set("url", analyzedURL)
			query.Set("strategy", strategy)
			for _, cat := range categories {
				query.Add("category", cat)
			}
			resp, err := r.Do(ctx, http.MethodGet, pagespeedRunPath, query, nil)
			if err != nil {
				return fmt.Errorf("read google-pagespeed-insights %s %s: %w", analyzedURL, strategy, err)
			}
			body, err := decodeReport(resp.Body)
			if err != nil {
				return fmt.Errorf("decode google-pagespeed-insights report for %s: %w", analyzedURL, err)
			}
			if err := emit(pagespeedRecord(analyzedURL, strategy, body)); err != nil {
				return err
			}
		}
	}
	return nil
}

// decodeReport decodes a runPagespeed response into a generic map, preserving
// numbers as json.Number so Lighthouse scores keep their fidelity.
func decodeReport(body []byte) (map[string]any, error) {
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free. It walks the
// same (url, strategy) iteration the live path does, using configured values
// when present and otherwise canned defaults.
func (c Connector) readFixture(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	urls := pagespeedURLs(req.Config)
	strategies := pagespeedStrategies(req.Config)
	for i, analyzedURL := range urls {
		for j, strategy := range strategies {
			if err := ctx.Err(); err != nil {
				return err
			}
			body := map[string]any{
				"kind":                 "pagespeedonline#result",
				"id":                   analyzedURL,
				"analysisUTCTimestamp": "2026-01-01T00:00:00.000Z",
				"loadingExperience":    map[string]any{"overall_category": "FAST"},
				"lighthouseResult": map[string]any{
					"requestedUrl":      analyzedURL,
					"finalUrl":          analyzedURL,
					"lighthouseVersion": "11.0.0",
					"fetchTime":         "2026-01-01T00:00:00.000Z",
					"categories": map[string]any{
						"performance":    map[string]any{"id": "performance", "score": json.Number(fmt.Sprintf("0.%02d", 90+i+j))},
						"accessibility":  map[string]any{"id": "accessibility", "score": json.Number("0.88")},
						"best-practices": map[string]any{"id": "best-practices", "score": json.Number("0.92")},
						"seo":            map[string]any{"id": "seo", "score": json.Number("1")},
					},
				},
			}
			if err := emit(pagespeedRecord(analyzedURL, strategy, body)); err != nil {
				return err
			}
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api-key query auth and the
// resolved base URL. The api_key secret is optional (PageSpeed works keyless but
// throttled); when absent the requester sends no key param. The secret only ever
// flows into connsdk.APIKeyQuery; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := pagespeedBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	r := &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: pagespeedUserAgent,
	}
	if secret := strings.TrimSpace(pagespeedSecret(cfg)); secret != "" {
		r.Auth = connsdk.APIKeyQuery("key", secret)
	}
	return r, nil
}

func pagespeedSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// pagespeedURLs resolves the configured urls list (comma-separated), falling
// back to a single default url so a credential-free read still produces output.
func pagespeedURLs(cfg connectors.RuntimeConfig) []string {
	if urls := splitList(configValue(cfg, "urls")); len(urls) > 0 {
		return urls
	}
	return []string{pagespeedDefaultURL}
}

// pagespeedStrategies resolves the configured strategies list, defaulting to
// desktop.
func pagespeedStrategies(cfg connectors.RuntimeConfig) []string {
	if strategies := splitList(configValue(cfg, "strategies")); len(strategies) > 0 {
		return strategies
	}
	return []string{pagespeedDefaultStrategy}
}

// pagespeedCategories resolves the configured Lighthouse categories list,
// defaulting to performance.
func pagespeedCategories(cfg connectors.RuntimeConfig) []string {
	if categories := splitList(configValue(cfg, "categories")); len(categories) > 0 {
		return categories
	}
	return []string{pagespeedDefaultCategory}
}

func configValue(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config[key]
}

// splitList splits a comma-separated config value into trimmed, non-empty
// entries. PageSpeed's array config fields (urls/strategies/categories) are
// carried as comma-joined strings in the flat RuntimeConfig.Config map.
func splitList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// pagespeedBaseURL resolves and validates the base URL. The default is
// www.googleapis.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func pagespeedBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(configValue(cfg, "base_url"))
	if base == "" {
		return pagespeedDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("google-pagespeed-insights config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-pagespeed-insights config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("google-pagespeed-insights config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: PageSpeed Insights is a read-only analysis API.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
