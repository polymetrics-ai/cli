// Package beamer implements the native pm Beamer source connector. It follows the
// declarative-HTTP template established by the stripe package: a thin package that
// composes the connsdk toolkit (Requester + Bearer auth + RecordsAt extraction +
// cursor state) with Beamer-specific stream definitions, endpoints, and the
// Beamer page-increment pagination over root-level JSON arrays.
//
// The contract is taken from the upstream upstream source-beamer manifest:
//   - base URL https://api.getbeamer.com/v0/
//   - BearerAuthenticator with the api_key as the token (Authorization: Bearer)
//   - PageIncrement pagination: page=0,1,... with maxResults as the page size
//   - record selector at the root (responses are bare JSON arrays)
//   - the nps stream is incremental on the "date" field, filtered via dateFrom
//
// Beamer's REST API is a read-only feedback/changelog surface (the upstream source
// supports full_refresh only and there is no safe reverse-ETL write target), so
// Capabilities.Write is false and there is no write.go.
//
// Like stripe/github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package in
// the production binary to run that side effect.
package beamer

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
	beamerDefaultBaseURL  = "https://api.getbeamer.com"
	beamerAPIPrefix       = "v0"
	beamerDefaultPageSize = 100
	beamerMaxPageSize     = 100
	beamerUserAgent       = "polymetrics-go-cli"
	// beamerFixtureDate is the deterministic ISO-8601 timestamp used by the
	// fixture-mode records.
	beamerFixtureDate = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("beamer", New)
}

// New returns the Beamer connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Beamer connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "beamer" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "beamer",
		DisplayName:     "Beamer",
		IntegrationType: "api",
		Description:     "Reads Beamer NPS survey responses, announcement posts, feature requests, and comments through the Beamer REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Beamer. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := beamerBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(beamerSecret(cfg)) == "" {
		return errors.New("beamer connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the nps list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, beamerAPIPrefix+"/nps", url.Values{"page": []string{"0"}, "maxResults": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check beamer: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: beamerStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Beamer is a read-only source
// (Capabilities.Write is false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Beamer stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time.
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
		stream = "nps"
	}
	endpoint, ok := beamerStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("beamer stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := beamerPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := beamerMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Beamer's page-increment pagination. Beamer list endpoints return
// a bare JSON array per page; the next page is requested with page=page+1
// (starting from 0) and maxResults carrying the page size. There is no pagination
// envelope, so we stop when a page returns fewer than pageSize records (or no
// records at all). The loop lives here on connsdk.Requester + connsdk.RecordsAt.
//
// When the stream is incremental and a lower bound is known (from the cursor or
// start_date config), it is injected via the stream's cursorParam (dateFrom for
// nps) so the server filters to records at or after that timestamp.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, lower string, emit func(connectors.Record) error) error {
	path := beamerAPIPrefix + "/" + endpoint.resource
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("maxResults", strconv.Itoa(pageSize))
		if lower != "" && endpoint.cursorParam != "" {
			query.Set(endpoint.cursorParam, lower)
		}

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read beamer %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode beamer %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short or empty page is the last page (Beamer exposes no total-count
		// envelope on these list endpoints).
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise beamer credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"date":             beamerFixtureDate,
			"score":            float64(7 + i),
			"feedback":         fmt.Sprintf("Fixture feedback %d", i),
			"userId":           fmt.Sprintf("user_%d", i),
			"userEmail":        fmt.Sprintf("fixture+%d@example.com", i),
			"userFirstName":    fmt.Sprintf("First%d", i),
			"userLastName":     fmt.Sprintf("Last%d", i),
			"country":          "US",
			"city":             "San Francisco",
			"language":         "en",
			"os":               "macOS",
			"browser":          "Chrome",
			"origin":           "web",
			"url":              fmt.Sprintf("https://app.example.com/%s/%d", endpoint.resource, i),
			"refUrl":           "https://app.example.com",
			"filter":           "",
			"category":         "new",
			"title":            fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"content":          fmt.Sprintf("Fixture content %d", i),
			"feedbackEnabled":  true,
			"reactionsEnabled": true,
			"published":        true,
			"clicks":           10 * i,
			"views":            100 * i,
			"status":           "open",
			"votesCount":       i,
			"commentsCount":    i,
			"postId":           "post_fixture_1",
			"featureRequestId": "fr_fixture_1",
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth (Beamer API keys are
// passed as the bearer token) and the resolved base URL. The secret only ever
// flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := beamerBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := beamerSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("beamer connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: beamerUserAgent,
	}, nil
}

// incrementalLowerBound returns the lower-bound timestamp for the stream's
// cursor parameter, derived from the incremental cursor (if any) or else the
// start_date config. An empty result means no lower bound (full sync). Beamer
// expects RFC3339 timestamps, which is exactly what the cursor and start_date
// carry, so no reformatting is required.
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func beamerSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// beamerBaseURL resolves and validates the base URL. The default is
// api.getbeamer.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func beamerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return beamerDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("beamer config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("beamer config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("beamer config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func beamerPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["max_results"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return beamerDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("beamer config page_size must be an integer: %w", err)
	}
	if value < 1 || value > beamerMaxPageSize {
		return 0, fmt.Errorf("beamer config page_size must be between 1 and %d", beamerMaxPageSize)
	}
	return value, nil
}

func beamerMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("beamer config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("beamer config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
