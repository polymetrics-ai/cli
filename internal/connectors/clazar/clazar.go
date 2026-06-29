// Package clazar implements the native pm Clazar connector. Clazar is a cloud
// GTM platform (AWS/Azure/GCP marketplace co-sell, listings, private offers,
// contracts). This is a declarative-HTTP per-system connector built on the
// connsdk toolkit, following the stripe reference shape: a thin package that
// composes a connsdk Requester + OAuth2 client-credentials auth + RecordsAt
// extraction + page-increment pagination with Clazar-specific stream definitions.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// API shape (from the upstream source-clazar manifest + Clazar developer docs):
//   - base URL https://api.clazar.io
//   - OAuth2 client_credentials, token endpoint https://api.clazar.io/authenticate/
//   - list endpoints return {"results":[...], "next":..., "count":...}
//   - page-increment pagination via the `page` query param (page_size default 100)
//   - incremental cursor field last_modified_at, filtered with last_modified_at_after
//
// Clazar is read-only here: the upstream source connector only supports
// full_refresh reads and exposes no obviously-safe reverse-ETL writes, so
// Capabilities.Write is false.
package clazar

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
	clazarDefaultBaseURL  = "https://api.clazar.io"
	clazarTokenPath       = "/authenticate/"
	clazarDefaultPageSize = 100
	clazarMaxPageSize     = 100
	clazarRecordsPath     = "results"
	clazarUserAgent       = "polymetrics-go-cli"
	// clazarFixtureModified is the deterministic last_modified_at base used by
	// fixture-mode records.
	clazarFixtureModified = "2026-01-01T00:00:00.000000Z"
)

func init() {
	connectors.RegisterFactory("clazar", New)
}

// New returns the Clazar connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Clazar connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token fetch. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "clazar" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "clazar",
		DisplayName:     "Clazar",
		IntegrationType: "api",
		Description:     "Reads Clazar cloud GTM data (buyers, listings, contracts, opportunities, and private offers) from the Clazar REST API using OAuth2 client credentials.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Clazar. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := clazarBaseURL(cfg); err != nil {
		return err
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the listings endpoint confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"page": []string{"1"}, "page_size": []string{"1"}, "response_format": []string{"common"}}
	if err := r.DoJSON(ctx, http.MethodGet, "listings", q, nil, nil); err != nil {
		return fmt.Errorf("check clazar: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: clazarStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Clazar stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "buyers"
	}
	endpoint, ok := clazarStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("clazar stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := clazarPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := clazarMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("response_format", "common")
	if lower := incrementalLowerBound(req); lower != "" {
		base.Set("last_modified_at_after", lower)
	}

	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "page_size",
		StartPage: 1,
		PageSize:  pageSize,
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, clazarRecordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}); err != nil {
		return fmt.Errorf("read clazar %s: %w", endpoint.resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise clazar credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":               fmt.Sprintf("Fixture %d", i),
			"cloud":              "aws",
			"domain":             fmt.Sprintf("fixture%d.example.com", i),
			"status":             "active",
			"title":              fmt.Sprintf("Fixture Listing %d", i),
			"stage":              "launched",
			"offer_type":         "private_offer",
			"buyer_id":           "buyers_fixture_1",
			"listing_id":         "listings_fixture_1",
			"cloud_id":           fmt.Sprintf("cloud_%d", i),
			"auto_renew":         false,
			"archived":           "false",
			"last_modified_at":   clazarFixtureModified,
			"start_at":           clazarFixtureModified,
			"created_at":         clazarFixtureModified,
			"accepted_at":        clazarFixtureModified,
			"published_at":       clazarFixtureModified,
			"latest_contract_id": "contracts_fixture_1",
			"latest_offer_id":    "private_offers_fixture_1",
		}
		record := endpoint.mapRecord(item)
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved base URL. The client_id/client_secret only ever flow into the
// connsdk OAuth2ClientCredentials authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := clazarBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     base + clazarTokenPath,
		ClientID:     clazarClientID(cfg),
		ClientSecret: clazarClientSecret(cfg),
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: clazarUserAgent,
	}, nil
}

// incrementalLowerBound returns the last_modified_at lower bound for the
// incremental filter, derived from the incremental cursor (if any) or else the
// start_date config. An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(clazarClientID(cfg)) == "" {
		return errors.New("clazar connector requires secret client_id")
	}
	if strings.TrimSpace(clazarClientSecret(cfg)) == "" {
		return errors.New("clazar connector requires secret client_secret")
	}
	return nil
}

func clazarClientID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_id"]
}

func clazarClientSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

// clazarBaseURL resolves and validates the base URL. The default is
// api.clazar.io; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func clazarBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return clazarDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("clazar config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("clazar config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("clazar config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func clazarPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return clazarDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("clazar config page_size must be an integer: %w", err)
	}
	if value < 1 || value > clazarMaxPageSize {
		return 0, fmt.Errorf("clazar config page_size must be between 1 and %d", clazarMaxPageSize)
	}
	return value, nil
}

func clazarMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("clazar config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("clazar config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Clazar is read-only in pm. The method exists to satisfy
// the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
