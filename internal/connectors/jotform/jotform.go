// Package jotform implements the native pm Jotform connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// APIKEY header auth + content[] extraction + resultSet offset/limit pagination)
// with Jotform-specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The Jotform API (https://api.jotform.com/docs/) authenticates with an APIKEY
// request header, wraps list payloads as {responseCode, content:[...],
// resultSet:{offset,limit,count}}, and paginates with offset/limit query params.
// This connector is read-only: the Jotform API's writes (creating/deleting
// forms, submissions) are not safe reverse-ETL targets.
package jotform

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
	jotformDefaultBaseURL  = "https://api.jotform.com"
	jotformEUBaseURL       = "https://eu-api.jotform.com"
	jotformHIPAABaseURL    = "https://hipaa-api.jotform.com"
	jotformDefaultPageSize = 100
	jotformMaxPageSize     = 1000
	jotformUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("jotform", New)
}

// New returns the Jotform connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Jotform connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "jotform" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "jotform",
		DisplayName:     "Jotform",
		IntegrationType: "api",
		Description:     "Reads Jotform forms, submissions, reports, folders, and the account profile through the Jotform REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Jotform. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := jotformBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(jotformSecret(cfg)) == "" {
		return errors.New("jotform connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the user profile confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "user", nil, nil, nil); err != nil {
		return fmt.Errorf("check jotform: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: jotformStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Jotform stream starts with
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
		stream = "forms"
	}
	endpoint, ok := jotformStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("jotform stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := jotformPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := jotformMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Jotform's resultSet offset/limit pagination. List endpoints
// return {content:[...], resultSet:{offset,limit,count}}; the next page is
// requested with offset+=limit, stopping when a page returns fewer records than
// the requested limit (or zero). Non-paginated endpoints (reports, folders,
// user) read a single bounded response.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("limit", strconv.Itoa(pageSize))
			query.Set("offset", strconv.Itoa(offset))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read jotform %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "content")
		if err != nil {
			return fmt.Errorf("decode jotform %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Non-paginated endpoints, an empty page, or a short page all end the
		// loop. A short page means fewer than the requested limit came back.
		if !endpoint.paginated || len(records) == 0 || len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise jotform credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idBase := strings.ReplaceAll(endpoint.resource, "/", "_")
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", idBase, i),
			"username":        "fixture_user",
			"title":           fmt.Sprintf("Fixture %s %d", stream, i),
			"name":            fmt.Sprintf("Fixture %d", i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"status":          "ENABLED",
			"created_at":      fmt.Sprintf("2026-01-0%d 00:00:00", i),
			"updated_at":      fmt.Sprintf("2026-01-0%d 12:00:00", i),
			"last_submission": fmt.Sprintf("2026-01-0%d 06:00:00", i),
			"new":             "1",
			"count":           strconv.Itoa(i),
			"type":            "LEGACY",
			"url":             fmt.Sprintf("https://www.jotform.com/%s_%d", idBase, i),
			"form_id":         "form_fixture_1",
			"flag":            "0",
			"notes":           "",
			"ip":              "203.0.113.1",
			"answers":         map[string]any{"1": map[string]any{"name": "field", "answer": fmt.Sprintf("value %d", i)}},
			"account_type":    "PREMIUM",
			"usage":           "0",
			"time_zone":       "UTC",
			"name_folder":     fmt.Sprintf("Fixture Folder %d", i),
			"parent":          "",
			"owner":           "fixture_user",
			"color":           "#ffffff",
			"fields":          "",
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

// requester builds a connsdk.Requester wired with APIKEY header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := jotformBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := jotformSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("jotform connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("APIKEY", secret, ""),
		UserAgent: jotformUserAgent,
	}, nil
}

func jotformSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// jotformBaseURL resolves and validates the base URL. The default is
// api.jotform.com; the url_prefix config selects the EU or HIPAA host. Any
// explicit base_url override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func jotformBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		switch strings.ToLower(strings.TrimSpace(cfg.Config["url_prefix"])) {
		case "eu":
			return jotformEUBaseURL, nil
		case "hipaa":
			return jotformHIPAABaseURL, nil
		default:
			return jotformDefaultBaseURL, nil
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("jotform config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("jotform config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("jotform config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func jotformPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return jotformDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jotform config page_size must be an integer: %w", err)
	}
	if value < 1 || value > jotformMaxPageSize {
		return 0, fmt.Errorf("jotform config page_size must be between 1 and %d", jotformMaxPageSize)
	}
	return value, nil
}

func jotformMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jotform config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("jotform config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Jotform is a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
