// Package convertkit implements the native pm ConvertKit (Kit) connector. It is a
// declarative-HTTP per-system connector built on the same template as stripe: a
// thin package that composes the connsdk toolkit (Requester + api_secret query
// auth + RecordsAt extraction) with ConvertKit-specific stream definitions and
// endpoints.
//
// ConvertKit's v3 REST API authenticates with an api_secret query parameter, lists
// each resource under a JSON key named after the resource (subscribers[],
// forms[], ...), and paginates the large collections (subscribers, broadcasts)
// with a "page" query param plus a "total_pages" field in the response body. The
// connector is read-only: the upstream Airbyte source supports full-refresh only
// and there are no safe reverse-ETL writes to expose.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package convertkit

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
	convertkitDefaultBaseURL  = "https://api.convertkit.com/v3"
	convertkitDefaultPageSize = 50
	convertkitMaxPageSize     = 50
	convertkitUserAgent       = "polymetrics-go-cli"
	// convertkitFixtureCreated is the deterministic created_at used by
	// fixture-mode records.
	convertkitFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("convertkit", New)
}

// New returns the ConvertKit connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm ConvertKit connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "convertkit" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "convertkit",
		DisplayName:     "ConvertKit",
		IntegrationType: "api",
		Description:     "Reads ConvertKit (Kit) subscribers, forms, sequences, tags, and broadcasts through the ConvertKit v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to ConvertKit.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := convertkitBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(convertkitSecret(cfg)) == "" {
		return errors.New("convertkit connector requires an api secret (credentials.api_key or credentials.access_token)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the forms list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "forms", nil, nil, nil); err != nil {
		return fmt.Errorf("check convertkit: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: convertkitStreams()}, nil
}

// Write satisfies the connectors.Connector interface. ConvertKit is exposed
// read-only (the upstream source is full-refresh only and there are no safe
// reverse-ETL writes to allow-list), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "subscribers"
	}
	endpoint, ok := convertkitStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("convertkit stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := convertkitPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := convertkitMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives ConvertKit's page-based pagination. Paginated list responses
// look like {<resource>:[...], page:N, total_pages:M}; the next page is requested
// with page=N+1 until page reaches total_pages. Non-paginated resources (forms,
// tags, sequences) return the whole collection in a single array, so the loop
// stops after page 1. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("page", strconv.Itoa(page))
			query.Set("per_page", strconv.Itoa(pageSize))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read convertkit %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayKey)
		if err != nil {
			return fmt.Errorf("decode convertkit %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if !endpoint.paginated {
			return nil
		}
		totalPages, err := connsdk.StringAt(resp.Body, "total_pages")
		if err != nil {
			return fmt.Errorf("decode convertkit %s total_pages: %w", endpoint.resource, err)
		}
		total, perr := strconv.Atoi(strings.TrimSpace(totalPages))
		if perr != nil || total <= page || len(records) == 0 {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise convertkit credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            int64(i),
			"first_name":    fmt.Sprintf("Fixture %d", i),
			"email_address": fmt.Sprintf("fixture+%d@example.com", i),
			"state":         "active",
			"created_at":    convertkitFixtureCreated,
			"name":          fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"type":          "embed",
			"format":        "inline",
			"archived":      false,
			"hold":          false,
			"repeat":        false,
			"subject":       fmt.Sprintf("Broadcast %d", i),
			"description":   "fixture broadcast",
			"public":        true,
			"published_at":  convertkitFixtureCreated,
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "convertkit"
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api_secret query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := convertkitBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := convertkitSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("convertkit connector requires an api secret (credentials.api_key or credentials.access_token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("api_secret", secret),
		UserAgent: convertkitUserAgent,
	}, nil
}

// convertkitSecret resolves the api secret from the runtime secrets. The catalog
// nests credentials under credentials.{api_key,access_token,...}; the secret
// loader flattens those to the leaf keys, with a legacy "api_secret" fallback.
func convertkitSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{"api_key", "access_token", "api_secret"} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// convertkitBaseURL resolves and validates the base URL. The default is
// api.convertkit.com/v3; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func convertkitBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return convertkitDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("convertkit config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("convertkit config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("convertkit config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func convertkitPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return convertkitDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("convertkit config page_size must be an integer: %w", err)
	}
	if value < 1 || value > convertkitMaxPageSize {
		return 0, fmt.Errorf("convertkit config page_size must be between 1 and %d", convertkitMaxPageSize)
	}
	return value, nil
}

func convertkitMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("convertkit config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("convertkit config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
