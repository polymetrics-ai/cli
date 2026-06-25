// Package tempo implements the native pm Tempo connector. It follows the stripe
// reference shape for declarative-HTTP per-system connectors: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + cursor state) with Tempo-specific stream definitions, endpoints,
// and pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Tempo Cloud REST API v4 reference: https://apidocs.tempo.io/. Auth is the
// Tempo API token presented as a Bearer credential. List endpoints return
// {"results":[...], "metadata":{"count":N,"offset":N,"limit":N,"next":"<url>"}}
// and are walked by following metadata.next (an absolute URL) until it is
// absent. Read-only: Tempo writes are not exposed for reverse ETL here.
package tempo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	tempoDefaultBaseURL  = "https://api.tempo.io/4"
	tempoDefaultPageSize = 50
	tempoMaxPageSize     = 1000
	tempoUserAgent       = "polymetrics-go-cli"
	// tempoFixtureDate is the deterministic date used by fixture-mode records.
	tempoFixtureDate = "2026-01-01"
)

func init() {
	connectors.RegisterFactory("tempo", New)
}

// New returns the Tempo connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Tempo connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "tempo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "tempo",
		DisplayName:     "Tempo",
		IntegrationType: "api",
		Description:     "Reads Tempo accounts, customers, worklogs, and workload schemes through the Tempo Cloud REST API v4.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Tempo. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := tempoBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(tempoSecret(cfg)) == "" {
		return errors.New("tempo connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the accounts list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "accounts", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check tempo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: tempoStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "worklogs"
	}
	endpoint, ok := tempoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("tempo stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := tempoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := tempoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Tempo v4 pagination. List responses carry their records under
// "results" plus a "metadata" object whose "next" field is an absolute URL for
// the following page (absent when exhausted). The connsdk.Requester treats an
// absolute http(s) path as-is, so each next URL is requested directly. The loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt,
// because this exact metadata.next shape has no off-the-shelf connsdk paginator.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// The first request hits the relative resource path with limit/offset; every
	// subsequent request follows the absolute metadata.next URL verbatim.
	path := endpoint.resource
	query := url.Values{}
	query.Set("limit", strconv.Itoa(pageSize))
	query.Set("offset", "0")

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read tempo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode tempo %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "metadata.next")
		if err != nil {
			return fmt.Errorf("decode tempo %s metadata.next: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// Follow the absolute next URL; it already carries limit/offset, so the
		// merged query is cleared to avoid clobbering it.
		path = next
		query = url.Values{}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise tempo credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               int64(i),
			"tempoWorklogId":   int64(i),
			"jiraWorklogId":    int64(1000 + i),
			"key":              fmt.Sprintf("%s-%d", strings.ToUpper(endpoint.resource), i),
			"name":             fmt.Sprintf("Fixture %s %d", stream, i),
			"status":           "OPEN",
			"global":           false,
			"monthlyBudget":    float64(1000 * i),
			"timeSpentSeconds": int64(3600 * i),
			"billableSeconds":  int64(3600 * i),
			"startDate":        tempoFixtureDate,
			"startTime":        "09:00:00",
			"description":      fmt.Sprintf("fixture worklog %d", i),
			"createdAt":        tempoFixtureDate + "T00:00:00Z",
			"updatedAt":        tempoFixtureDate + "T00:00:00Z",
			"defaultScheme":    false,
			"issue":            map[string]any{"id": int64(5000 + i)},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "tempo"
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := tempoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := tempoSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("tempo connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: tempoUserAgent,
	}, nil
}

func tempoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// tempoBaseURL resolves and validates the base URL. The default is
// api.tempo.io/4; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func tempoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return tempoDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("tempo config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("tempo config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("tempo config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func tempoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return tempoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("tempo config page_size must be an integer: %w", err)
	}
	if value < 1 || value > tempoMaxPageSize {
		return 0, fmt.Errorf("tempo config page_size must be between 1 and %d", tempoMaxPageSize)
	}
	return value, nil
}

func tempoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("tempo config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("tempo config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: the Tempo connector is read-only for ETL.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
