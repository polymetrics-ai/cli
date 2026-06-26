// Package nytimes implements the native pm New York Times connector. It is a
// declarative-HTTP per-system connector built in the shape of the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// api-key query auth + RecordsAt extraction) with NYTimes-specific stream
// definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The NYTimes APIs are read-only (no reverse-ETL writes make sense), so the
// connector exposes Read/Catalog/Check only.
package nytimes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	nytimesDefaultBaseURL = "https://api.nytimes.com/svc"
	nytimesUserAgent      = "polymetrics-go-cli"
	// nytimesMaxMonths bounds the archive month iteration to avoid unbounded
	// fan-out when a wide date range is configured.
	nytimesMaxMonths = 1200
)

func init() {
	connectors.RegisterFactory("nytimes", New)
}

// New returns the NYTimes connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm New York Times connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nytimes" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nytimes",
		DisplayName:     "New York Times",
		IntegrationType: "api",
		Description:     "Reads New York Times Most Popular (viewed, emailed, shared) articles and the article Archive via the NYTimes Developer APIs.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the NYTimes
// API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nytimesBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(nytimesSecret(cfg)) == "" {
		return errors.New("nytimes connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	period := nytimesPeriod(cfg)
	// A bounded read of the most-viewed feed confirms auth and connectivity
	// without mutating anything.
	path := fmt.Sprintf("mostpopular/v2/viewed/%s.json", period)
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check nytimes: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nytimesStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "most_popular_viewed"
	}
	def, ok := nytimesStreamDefs[stream]
	if !ok {
		return fmt.Errorf("nytimes stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, def, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	switch def.kind {
	case kindMostPopular:
		return c.readMostPopular(ctx, r, def, req.Config, emit)
	case kindArchive:
		return c.readArchive(ctx, r, def, req.Config, emit)
	default:
		return fmt.Errorf("nytimes stream %q has unknown kind", stream)
	}
}

// Write is unsupported: the NYTimes APIs are read-only, so the connector
// declares Capabilities.Write=false and rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readMostPopular issues a single request to the Most Popular API for the
// configured period (and share_type for the shared metric). The response carries
// every result inline; there is no pagination.
func (c Connector) readMostPopular(ctx context.Context, r *connsdk.Requester, def streamDef, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	period := nytimesPeriod(cfg)
	var path string
	if def.metric == "shared" {
		if shareType := strings.TrimSpace(cfg.Config["share_type"]); shareType != "" {
			path = fmt.Sprintf("mostpopular/v2/shared/%s/%s.json", period, shareType)
		} else {
			path = fmt.Sprintf("mostpopular/v2/shared/%s.json", period)
		}
	} else {
		path = fmt.Sprintf("mostpopular/v2/%s/%s.json", def.metric, period)
	}

	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read nytimes most_popular_%s: %w", def.metric, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
	if err != nil {
		return fmt.Errorf("decode nytimes most_popular_%s: %w", def.metric, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readArchive iterates the Archive API one calendar month at a time from
// start_date to end_date (inclusive). Each month is a separate request; this is
// the connector's natural pagination across the configured window.
func (c Connector) readArchive(ctx context.Context, r *connsdk.Requester, def streamDef, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	start, err := parseYearMonth(strings.TrimSpace(cfg.Config["start_date"]))
	if err != nil {
		return fmt.Errorf("nytimes config start_date: %w", err)
	}
	end := start
	if raw := strings.TrimSpace(cfg.Config["end_date"]); raw != "" {
		end, err = parseYearMonth(raw)
		if err != nil {
			return fmt.Errorf("nytimes config end_date: %w", err)
		}
	}
	if end.Before(start) {
		return errors.New("nytimes config end_date must not be before start_date")
	}

	cur := start
	for months := 0; !cur.After(end); months++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if months >= nytimesMaxMonths {
			return fmt.Errorf("nytimes archive window exceeds %d months", nytimesMaxMonths)
		}
		path := fmt.Sprintf("archive/v1/%d/%d.json", cur.Year(), int(cur.Month()))
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read nytimes archive %d/%02d: %w", cur.Year(), int(cur.Month()), err)
		}
		records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
		if err != nil {
			return fmt.Errorf("decode nytimes archive %d/%02d: %w", cur.Year(), int(cur.Month()), err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(def.mapRecord(item)); err != nil {
				return err
			}
		}
		cur = cur.AddDate(0, 1, 0)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise nytimes credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if def.kind == kindArchive {
			item = map[string]any{
				"_id":              fmt.Sprintf("nyt://article/fixture-%d", i),
				"web_url":          fmt.Sprintf("https://www.nytimes.com/fixture/%d", i),
				"snippet":          fmt.Sprintf("Fixture archive snippet %d", i),
				"pub_date":         fmt.Sprintf("2022-01-0%dT00:00:00Z", i),
				"document_type":    "article",
				"news_desk":        "Business",
				"section_name":     "Technology",
				"type_of_material": "News",
				"word_count":       int64(500 * i),
				"headline":         map[string]any{"main": fmt.Sprintf("Fixture Headline %d", i)},
			}
		} else {
			item = map[string]any{
				"id":             int64(1000 + i),
				"url":            fmt.Sprintf("https://www.nytimes.com/fixture/%d", i),
				"title":          fmt.Sprintf("Fixture %s %d", def.metric, i),
				"abstract":       fmt.Sprintf("Fixture abstract %d", i),
				"byline":         "By Fixture Author",
				"section":        "Technology",
				"source":         "New York Times",
				"type":           "Article",
				"published_date": fmt.Sprintf("2022-01-0%d", i),
				"updated":        fmt.Sprintf("2022-01-0%d 12:00:00", i),
				"uri":            fmt.Sprintf("nyt://article/fixture-%d", i),
			}
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api-key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := nytimesBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := nytimesSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("nytimes connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("api-key", secret),
		UserAgent: nytimesUserAgent,
	}, nil
}

func nytimesSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// nytimesPeriod resolves the Most Popular period (1, 7, or 30 days). Invalid or
// missing values fall back to 7, which the NYTimes API accepts.
func nytimesPeriod(cfg connectors.RuntimeConfig) string {
	raw := strings.TrimSpace(cfg.Config["period"])
	switch raw {
	case "1", "7", "30":
		return raw
	default:
		return "7"
	}
}

// nytimesBaseURL resolves and validates the base URL. The default is
// api.nytimes.com/svc; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func nytimesBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return nytimesDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("nytimes config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nytimes config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nytimes config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// parseYearMonth parses a YYYY-MM string into the first day of that month (UTC).
func parseYearMonth(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("must be provided in YYYY-MM format")
	}
	t, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("must be YYYY-MM: %w", err)
	}
	return t.UTC(), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
