// Package easypromos implements the native pm Easypromos source connector. It is
// a declarative-HTTP per-system connector built on the stripe template: a thin
// package composing the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Easypromos-specific stream definitions and endpoints.
//
// Easypromos (https://api.easypromosapp.com/v2) is a platform for running
// contests, giveaways, and promotions. This connector is read-only: the upstream
// source supports only full_refresh syncs and exposes no safe reverse-ETL writes.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package easypromos

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
	easypromosDefaultBaseURL = "https://api.easypromosapp.com/v2"
	easypromosUserAgent      = "polymetrics-go-cli"
	// easypromosMaxPagesGuard bounds the cursor loop so a misbehaving API that
	// never returns a null next_cursor cannot loop forever.
	easypromosMaxPagesGuard = 10000
)

func init() {
	connectors.RegisterFactory("easypromos", New)
}

// New returns the Easypromos connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Easypromos source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "easypromos" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "easypromos",
		DisplayName:     "Easypromos",
		IntegrationType: "api",
		Description:     "Reads Easypromos promotions, organizing brands, stages, users, participations, and prizes through the Easypromos REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Easypromos.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := easypromosBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(easypromosSecret(cfg)) == "" {
		return errors.New("easypromos connector requires secret bearer_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the promotions list confirms auth and connectivity
	// without mutating anything. This mirrors the upstream connector's
	// CheckStream(promotions).
	if err := r.DoJSON(ctx, http.MethodGet, "promotions", nil, nil, nil); err != nil {
		return fmt.Errorf("check easypromos: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: easypromosStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "promotions"
	}
	endpoint, ok := easypromosStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("easypromos stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	path := endpoint.resource
	if endpoint.perPromotion {
		promotionID := strings.TrimSpace(req.Config.Config["promotion_id"])
		if promotionID == "" {
			return fmt.Errorf("easypromos stream %q requires config promotion_id", stream)
		}
		path = endpoint.resource + "/" + url.PathEscape(promotionID)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, endpoint.mapRecord, emit)
}

// harvest drives Easypromos cursor pagination. List responses have the shape
// {items:[...], paging:{next_cursor:"..."|null}}; the next page is requested with
// next_cursor=<token> until paging.next_cursor is null. connsdk has no body-token
// paginator that also reads from items[], so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	nextCursor := ""
	for page := 0; page < easypromosMaxPagesGuard; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if nextCursor != "" {
			query.Set("next_cursor", nextCursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read easypromos %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode easypromos %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		token, err := connsdk.StringAt(resp.Body, "paging.next_cursor")
		if err != nil {
			return fmt.Errorf("decode easypromos %s next_cursor: %w", path, err)
		}
		if strings.TrimSpace(token) == "" {
			return nil
		}
		nextCursor = token
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise easypromos credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, _ connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"title":            fmt.Sprintf("Fixture Promotion %d", i),
			"name":             fmt.Sprintf("Fixture %d", i),
			"description":      "Deterministic fixture record.",
			"promotion_type":   "giveaway",
			"status":           "active",
			"default_language": "en",
			"timezone":         "UTC",
			"url":              fmt.Sprintf("https://example.com/promo/%d", i),
			"start_date":       "2026-01-01T00:00:00Z",
			"end_date":         "2026-02-01T00:00:00Z",
			"created":          "2026-01-01T00:00:00Z",
			"type":             "form",
			"visible":          true,
			"external_id":      fmt.Sprintf("ext_%d", i),
			"first_name":       "Fixture",
			"last_name":        fmt.Sprintf("User %d", i),
			"nickname":         fmt.Sprintf("fixture%d", i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"country":          "US",
			"language":         "en",
			"login_type":       "form",
			"promotion_id":     "promo_fixture_1",
			"user_id":          "users_fixture_1",
			"stage_id":         "stages_fixture_1",
			"ip":               "203.0.113.1",
			"user_agent":       "polymetrics-fixture",
			"participation_id": "participations_fixture_1",
			"code":             fmt.Sprintf("CODE-%d", i),
			"redeem_url":       fmt.Sprintf("https://example.com/redeem/%d", i),
			"download_url":     fmt.Sprintf("https://example.com/download/%d", i),
			"organizing_brand": map[string]any{
				"id":   "brand_fixture_1",
				"name": "Fixture Brand",
			},
			"prize_type": map[string]any{
				"id":   "prizetype_fixture_1",
				"name": "Fixture Prize",
			},
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
	base, err := easypromosBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := easypromosSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("easypromos connector requires secret bearer_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: easypromosUserAgent,
	}, nil
}

func easypromosSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["bearer_token"]
}

// easypromosBaseURL resolves and validates the base URL. The default is
// api.easypromosapp.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func easypromosBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return easypromosDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("easypromos config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("easypromos config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("easypromos config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: easypromos is a read-only source connector.
func (c Connector) Write(ctx context.Context, _ connectors.WriteRequest, _ []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
