// Package eventee implements the native pm Eventee connector. It follows the
// declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Eventee-specific stream definitions and endpoints.
//
// Eventee's public API (https://api.eventee.co/public/v1) is read-only for our
// purposes: most streams are served from a single /content endpoint whose JSON
// body holds nested arrays (lectures, speakers, days, halls, tracks, ...), while
// partners/participants have dedicated endpoints returning top-level arrays.
// There is no pagination and no incremental cursor, so this connector is a
// straightforward full-refresh source.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package eventee

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
	eventeeDefaultBaseURL = "https://api.eventee.co/public/v1"
	eventeeUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("eventee", New)
}

// New returns the Eventee connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Eventee connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "eventee" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "eventee",
		DisplayName:     "Eventee",
		IntegrationType: "api",
		Description:     "Reads Eventee event agenda data (lectures, speakers, days, halls, tracks, workshops, partners) through the Eventee public REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Eventee. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := eventeeBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(eventeeSecret(cfg)) == "" {
		return errors.New("eventee connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of /content confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "content", nil, nil, nil); err != nil {
		return fmt.Errorf("check eventee: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: eventeeStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "lectures"
	}
	endpoint, ok := eventeeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("eventee stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	// Eventee has no pagination: a single GET returns the full collection.
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, url.Values{}, nil)
	if err != nil {
		return fmt.Errorf("read eventee %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode eventee %s: %w", stream, err)
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
// conformance harness can exercise eventee credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           i,
			"name":         fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"description":  fmt.Sprintf("Fixture %s record %d.", stream, i),
			"type":         "lecture",
			"code":         fmt.Sprintf("%s_%d", stream, i),
			"company":      fmt.Sprintf("Fixture Co %d", i),
			"position":     "Speaker",
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"phone":        "",
			"country":      "US",
			"language":     "en",
			"bio":          "Fixture bio.",
			"web":          "https://example.com",
			"address":      "1 Fixture St",
			"color":        "#112233",
			"date":         "2026-06-01",
			"content_url":  "https://example.com/content",
			"start":        "2026-06-01T09:00:00Z",
			"end":          "2026-06-01T10:00:00Z",
			"capacity":     100,
			"available":    true,
			"booked":       0,
			"sponsor":      true,
			"exhibitor":    false,
			"event_id":     1,
			"event_day_id": 1,
			"hall_id":      1,
			"order":        i,
			"created_at":   "2026-06-01T00:00:00Z",
			"updated_at":   "2026-06-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := eventeeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := eventeeSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("eventee connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: eventeeUserAgent,
	}, nil
}

func eventeeSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// eventeeBaseURL resolves and validates the base URL. The default is
// api.eventee.co; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func eventeeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return eventeeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("eventee config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("eventee config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("eventee config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Eventee is read-only, so
// writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
