// Package imagga implements the native pm Imagga connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, following
// the stripe reference template: a thin package that composes a connsdk Requester
// (HTTP Basic auth with the api_key/api_secret pair) with Imagga-specific stream
// definitions and per-image detection endpoints.
//
// Imagga is an image-recognition REST API. Its detection endpoints (tags,
// categories, colors, faces) analyze one image per request and return a result
// object; this connector issues one request per configured image URL and fans
// the nested result arrays out into records. The usage endpoint is account
// scoped. All streams are full-refresh (no incremental cursor).
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package imagga

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
	imaggaDefaultBaseURL = "https://api.imagga.com/v2"
	imaggaUserAgent      = "polymetrics-go-cli"
	// imaggaDefaultImage is the sample image Imagga ships in its docs/config; used
	// when no image_urls are configured so a best-effort read still works.
	imaggaDefaultImage = "https://imagga.com/static/images/categorization/child-476506_640.jpg"
)

func init() {
	connectors.RegisterFactory("imagga", New)
}

// New returns the Imagga connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Imagga connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "imagga" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "imagga",
		DisplayName:     "Imagga",
		IntegrationType: "api",
		Description:     "Reads Imagga image-recognition results (tags, categories, colors, face detections) and account usage via the Imagga REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector can authenticate against Imagga. In fixture mode
// it short-circuits without a network call; otherwise it performs a bounded read
// of the account usage endpoint, which confirms credentials without analyzing an
// image (and without spending detection quota).
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := imaggaBaseURL(cfg); err != nil {
		return err
	}
	key, secret := imaggaCredentials(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("imagga connector requires secrets api_key and api_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "usage", nil, nil, nil); err != nil {
		return fmt.Errorf("check imagga: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: imaggaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "tags"
	}
	endpoint, ok := imaggaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("imagga stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.perImage {
		// Account-scoped endpoint (usage): a single request, no image.
		return c.readOne(ctx, r, endpoint, "", emit)
	}

	images := imaggaImages(req.Config)
	for _, image := range images {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := c.readOne(ctx, r, endpoint, image, emit); err != nil {
			return err
		}
	}
	return nil
}

// readOne issues a single request to an Imagga endpoint and emits the mapped
// records. For per-image endpoints the image is passed as the image_url query
// parameter.
func (c Connector) readOne(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, image string, emit func(connectors.Record) error) error {
	query := url.Values{}
	if endpoint.perImage && image != "" {
		query.Set("image_url", image)
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read imagga %s: %w", endpoint.resource, err)
	}
	body, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode imagga %s: %w", endpoint.resource, err)
	}
	// RecordsAt on a single top-level object returns a one-element set; reuse its
	// decoded map for the stream-specific mapper.
	if len(body) == 0 {
		return nil
	}
	for _, rec := range endpoint.mapRecords(map[string]any(body[0]), image) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise imagga credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	image := imaggaDefaultImage
	var body map[string]any
	switch stream {
	case "tags":
		body = map[string]any{"result": map[string]any{"tags": []any{
			map[string]any{"confidence": 99.0, "tag": map[string]any{"en": "fixture_cat"}},
			map[string]any{"confidence": 80.5, "tag": map[string]any{"en": "fixture_pet"}},
		}}}
	case "categories":
		body = map[string]any{"result": map[string]any{"categories": []any{
			map[string]any{"confidence": 70.0, "name": map[string]any{"en": "pets animals"}},
		}}}
	case "colors":
		body = map[string]any{"result": map[string]any{"colors": map[string]any{
			"overall_colors": []any{
				map[string]any{"html_code": "#1f2a44", "closest_palette_color": "navy blue", "percent": 35.0, "r": 31, "g": 42, "b": 68},
			},
		}}}
	case "faces_detections":
		body = map[string]any{"result": map[string]any{"faces": []any{
			map[string]any{"confidence": 99.9, "coordinates": map[string]any{"xmin": 10, "ymin": 12, "xmax": 110, "ymax": 132}},
		}}}
	case "usage":
		image = ""
		body = map[string]any{"result": map[string]any{"total": 42, "monthly_processed": 100, "monthly_limit": 2000, "daily_processed": 7}}
	default:
		return fmt.Errorf("imagga fixture stream %q not found", stream)
	}
	for _, rec := range endpoint.mapRecords(body, image) {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (api_key as
// username, api_secret as password) and the resolved base URL. Secrets only ever
// flow into connsdk.Basic; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := imaggaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	key, secret := imaggaCredentials(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("imagga connector requires secrets api_key and api_secret")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(key, secret),
		UserAgent: imaggaUserAgent,
	}, nil
}

func imaggaCredentials(cfg connectors.RuntimeConfig) (key, secret string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["api_key"], cfg.Secrets["api_secret"]
}

// imaggaImages returns the list of image URLs to analyze, from the image_urls
// (comma-separated) or img_for_detection config, falling back to Imagga's sample
// image so a best-effort read always issues at least one request.
func imaggaImages(cfg connectors.RuntimeConfig) []string {
	if cfg.Config != nil {
		if raw := strings.TrimSpace(cfg.Config["image_urls"]); raw != "" {
			var out []string
			for _, part := range strings.Split(raw, ",") {
				if v := strings.TrimSpace(part); v != "" {
					out = append(out, v)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
		if single := strings.TrimSpace(cfg.Config["img_for_detection"]); single != "" {
			return []string{single}
		}
	}
	return []string{imaggaDefaultImage}
}

// imaggaBaseURL resolves and validates the base URL. The default is
// api.imagga.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func imaggaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return imaggaDefaultBaseURL, nil
	}
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return imaggaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("imagga config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("imagga config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("imagga config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is not supported: Imagga is a read-only analysis API with no reverse-ETL
// surface. It satisfies the Connector interface and reports unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
