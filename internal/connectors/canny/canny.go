// Package canny implements the native pm Canny connector. It follows the
// declarative-HTTP per-system connector shape established by the stripe package:
// a thin package that composes the connsdk toolkit (Requester + form-body auth +
// RecordsAt extraction) with Canny-specific stream definitions, endpoints, and
// the skip/limit offset pagination Canny uses.
//
// Canny is read-only here (feedback boards, posts, comments, categories,
// companies); there is no safe reverse-ETL surface, so Capabilities.Write=false.
//
// Like stripe/github it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Canny authenticates by passing the secret apiKey in the POST form body of every
// request. The secret only ever flows into the form payload; it is never placed
// in the URL or logged.
package canny

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
	cannyDefaultBaseURL  = "https://canny.io/api/v1"
	cannyDefaultPageSize = 100
	cannyMaxPageSize     = 100
	cannyUserAgent       = "polymetrics-go-cli"
	// cannyFixtureCreated is the deterministic ISO-8601 `created` timestamp used
	// by the fixture-mode records.
	cannyFixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("canny", New)
}

// New returns the Canny connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Canny connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "canny" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "canny",
		DisplayName:     "Canny",
		IntegrationType: "api",
		Description:     "Reads Canny boards, posts, comments, categories, and companies through the Canny REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Canny. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := cannyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cannySecret(cfg)) == "" {
		return errors.New("canny connector requires secret api_key")
	}
	r, base, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the boards list confirms auth and connectivity without
	// mutating anything. The api key rides in the form body.
	form := url.Values{}
	form.Set("apiKey", base.secret)
	if _, err := r.DoForm(ctx, http.MethodPost, "boards/list", nil, form); err != nil {
		return fmt.Errorf("check canny: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: cannyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Canny stream starts with an
// empty incremental cursor (full sync). Canny only supports full_refresh, so the
// cursor is informational.
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
		stream = "posts"
	}
	endpoint, ok := cannyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("canny stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, base, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := cannyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := cannyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, base, endpoint, pageSize, maxPages, emit)
}

// harvest drives Canny's skip/limit offset pagination. List endpoints are POST
// calls returning {<recordsKey>:[...], hasMore:bool}; the next page advances the
// skip offset by the page size. Boards is unpaginated, so a single page is read.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, base requestBase, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		form := url.Values{}
		form.Set("apiKey", base.secret)
		if endpoint.paginated {
			form.Set("limit", strconv.Itoa(pageSize))
			form.Set("skip", strconv.Itoa(skip))
		}

		resp, err := r.DoForm(ctx, http.MethodPost, endpoint.resource, nil, form)
		if err != nil {
			return fmt.Errorf("read canny %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode canny %s page: %w", endpoint.resource, err)
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
		hasMore, err := connsdk.StringAt(resp.Body, "hasMore")
		if err != nil {
			return fmt.Errorf("decode canny %s hasMore: %w", endpoint.resource, err)
		}
		// Stop on hasMore=false or on a short/empty page to avoid an infinite
		// loop if the server omits hasMore.
		if hasMore != "true" || len(records) == 0 {
			return nil
		}
		skip += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise canny credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":            fmt.Sprintf("Fixture %s %d", stream, i),
			"title":           fmt.Sprintf("Fixture post %d", i),
			"value":           fmt.Sprintf("Fixture comment %d", i),
			"details":         "fixture details",
			"status":          "open",
			"created":         cannyFixtureCreated,
			"score":           int64(10 * i),
			"commentCount":    int64(i),
			"postCount":       int64(i),
			"memberCount":     int64(i),
			"monthlySpend":    float64(i) * 9.99,
			"likeCount":       int64(i),
			"url":             fmt.Sprintf("https://canny.io/fixture/%s/%d", stream, i),
			"eta":             "",
			"statusChangedAt": cannyFixtureCreated,
			"isPrivate":       false,
			"internal":        false,
			"private":         false,
			"parentID":        "",
			"domain":          "example.com",
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

// requestBase carries per-read derived values that are not safe to log (the
// resolved secret).
type requestBase struct {
	secret string
}

// requester builds a connsdk.Requester for Canny. Canny authenticates via an
// apiKey form parameter rather than an Authorization header, so no connsdk
// Authenticator is wired; the secret is returned in requestBase for the caller to
// place into each form body. The secret is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, requestBase, error) {
	base, err := cannyBaseURL(cfg)
	if err != nil {
		return nil, requestBase{}, err
	}
	secret := strings.TrimSpace(cannySecret(cfg))
	if secret == "" {
		return nil, requestBase{}, errors.New("canny connector requires secret api_key")
	}
	r := &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: cannyUserAgent,
	}
	return r, requestBase{secret: secret}, nil
}

func cannySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// cannyBaseURL resolves and validates the base URL. The default is canny.io; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func cannyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return cannyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("canny config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("canny config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("canny config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cannyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return cannyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("canny config page_size must be an integer: %w", err)
	}
	if value < 1 || value > cannyMaxPageSize {
		return 0, fmt.Errorf("canny config page_size must be between 1 and %d", cannyMaxPageSize)
	}
	return value, nil
}

func cannyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("canny config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("canny config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Canny is read-only here, so
// writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
