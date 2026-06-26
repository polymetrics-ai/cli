// Package lemlist implements the native pm lemlist connector. lemlist is a cold
// outreach / sales engagement platform; this connector reads its team,
// campaigns, activities, and unsubscribes via the lemlist REST API.
//
// It follows the stripe reference shape: a thin package that composes the connsdk
// toolkit (Requester + the lemlist access_token query auth + RecordsAt extraction
// over a root-level array) with lemlist-specific stream definitions and offset
// pagination. It self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// lemlist supports full-refresh reads only and exposes no safe reverse-ETL write
// surface here, so the connector is read-only (Capabilities.Write=false).
package lemlist

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
	lemlistDefaultBaseURL  = "https://api.lemlist.com/api"
	lemlistDefaultPageSize = 100
	lemlistMaxPageSize     = 100
	lemlistTokenParam      = "access_token"
	lemlistUserAgent       = "polymetrics-go-cli"
	// lemlistFixtureCreated is the deterministic createdAt used by fixture records.
	lemlistFixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("lemlist", New)
}

// New returns the lemlist connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm lemlist connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "lemlist" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "lemlist",
		DisplayName:     "Lemlist",
		IntegrationType: "api",
		Description:     "Reads lemlist team, campaigns, activities, and unsubscribes through the lemlist REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to lemlist. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := lemlistBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(lemlistSecret(cfg)) == "" {
		return errors.New("lemlist connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A read of the team endpoint confirms auth and connectivity without
	// mutating anything; it is the canonical CheckStream in the upstream spec.
	if err := r.DoJSON(ctx, http.MethodGet, "team", nil, nil, nil); err != nil {
		return fmt.Errorf("check lemlist: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: lemlistStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "campaigns"
	}
	endpoint, ok := lemlistStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("lemlist stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := lemlistPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := lemlistMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives lemlist's offset/limit pagination. List endpoints return a
// root-level JSON array (records at path ""), advance via offset, and end when a
// short page is returned. The team endpoint returns a single object and is read
// in one request. The secret never appears in this loop (it rides on the
// requester's auth).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
		if err != nil {
			return fmt.Errorf("read lemlist %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode lemlist %s: %w", endpoint.resource, err)
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

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read lemlist %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode lemlist %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than pageSize records) means we are done.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise lemlist credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	count := 2
	if !endpoint.paginated {
		// team is a single workspace object.
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"_id":          fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"createdAt":    lemlistFixtureCreated,
			"_updatedAt":   lemlistFixtureCreated,
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"type":         "emailsSent",
			"campaignId":   "cmp_fixture_1",
			"campaignName": "Fixture Campaign",
			"isFirst":      i == 1,
			"sequenceStep": i,
			"labels":       []any{"fixture"},
			"userIds":      []any{"usr_fixture_1"},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the lemlist access_token query
// authenticator and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyQuery; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := lemlistBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := lemlistSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("lemlist connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery(lemlistTokenParam, secret),
		UserAgent: lemlistUserAgent,
	}, nil
}

func lemlistSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// lemlistBaseURL resolves and validates the base URL. The default is
// api.lemlist.com/api; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func lemlistBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return lemlistDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("lemlist config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("lemlist config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("lemlist config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func lemlistPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return lemlistDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lemlist config page_size must be an integer: %w", err)
	}
	if value < 1 || value > lemlistMaxPageSize {
		return 0, fmt.Errorf("lemlist config page_size must be between 1 and %d", lemlistMaxPageSize)
	}
	return value, nil
}

func lemlistMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lemlist config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("lemlist config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: lemlist is a read-only source connector here.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
