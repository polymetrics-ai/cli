// Package dropboxsign implements the native pm Dropbox Sign (formerly HelloSign)
// connector. It is a declarative-HTTP per-system connector following the stripe
// template: a thin package that composes the connsdk toolkit (Requester + HTTP
// Basic auth + RecordsAt extraction) with Dropbox Sign-specific stream
// definitions, endpoints, and page-number pagination.
//
// The package self-registers with the connectors registry via RegisterFactory in
// init() under the key "dropbox-sign"; the registryset package blank-imports this
// package in the production binary to run that side effect.
//
// Dropbox Sign authenticates with HTTP Basic auth using the API key as the
// username and an empty password. List endpoints return a `list_info` object
// ({page, num_pages, num_results, page_size}) alongside a named records array;
// pagination advances by incrementing the `page` query param until num_pages is
// reached. This connector is read-only.
package dropboxsign

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
	dropboxSignConnectorName   = "dropbox-sign"
	dropboxSignDefaultBaseURL  = "https://api.hellosign.com/v3"
	dropboxSignDefaultPageSize = 100
	dropboxSignMaxPageSize     = 100
	dropboxSignUserAgent       = "polymetrics-go-cli"
	// dropboxSignFixtureCreated is the deterministic timestamp used by
	// fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	dropboxSignFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory(dropboxSignConnectorName, New)
}

// New returns the Dropbox Sign connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Dropbox Sign connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return dropboxSignConnectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            dropboxSignConnectorName,
		DisplayName:     "Dropbox Sign",
		IntegrationType: "api",
		Description:     "Reads Dropbox Sign (HelloSign) signature requests, templates, team members, and account details through the Dropbox Sign REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Dropbox Sign.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dropboxSignBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(dropboxSignSecret(cfg)) == "" {
		return errors.New("dropbox-sign connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the account confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "account", nil, nil, nil); err != nil {
		return fmt.Errorf("check dropbox-sign: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dropboxSignStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "signature_requests"
	}
	endpoint, ok := dropboxSignStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("dropbox-sign stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := dropboxSignPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := dropboxSignMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Dropbox Sign's list_info page-number pagination. List endpoints
// return {<records_path>:[...], list_info:{page,num_pages,...}}; the next page is
// requested with page=<n+1> until page >= num_pages. The single-object account
// endpoint has no list_info and is emitted as one record. The loop lives here
// (rather than connsdk.Harvest) because the stop condition reads num_pages from
// the body, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for emitted := 0; maxPages == 0 || emitted < maxPages; emitted++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		// The account endpoint takes no pagination params.
		paginated := endpoint.recordsPath != "account"
		if paginated {
			query.Set("page", strconv.Itoa(page))
			query.Set("page_size", strconv.Itoa(pageSize))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read dropbox-sign %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode dropbox-sign %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if !paginated {
			return nil
		}
		numPages, err := pageCount(resp.Body)
		if err != nil {
			return fmt.Errorf("decode dropbox-sign %s list_info: %w", endpoint.resource, err)
		}
		if len(records) == 0 || page >= numPages {
			return nil
		}
		page++
	}
	return nil
}

// pageCount reads list_info.num_pages from the body, defaulting to 1 when absent.
func pageCount(body []byte) (int, error) {
	raw, err := connsdk.StringAt(body, "list_info.num_pages")
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(raw) == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 1, nil
	}
	if n < 1 {
		return 1, nil
	}
	return n, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise dropbox-sign credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"signature_request_id":    fmt.Sprintf("sr_fixture_%d", i),
			"template_id":             fmt.Sprintf("tpl_fixture_%d", i),
			"account_id":              fmt.Sprintf("acct_fixture_%d", i),
			"title":                   fmt.Sprintf("Fixture Document %d", i),
			"subject":                 fmt.Sprintf("Please sign %d", i),
			"message":                 "Fixture message",
			"is_complete":             i%2 == 0,
			"is_declined":             false,
			"has_error":               false,
			"is_creator":              true,
			"is_embedded":             false,
			"is_locked":               false,
			"test_mode":               true,
			"requester_email_address": fmt.Sprintf("fixture+%d@example.com", i),
			"email_address":           fmt.Sprintf("fixture+%d@example.com", i),
			"role":                    "Member",
			"role_code":               "a",
			"locale":                  "en-US",
			"is_paid_hs":              true,
			"is_paid_hf":              false,
			"created_at":              dropboxSignFixtureCreated + int64(i),
			"updated_at":              dropboxSignFixtureCreated + int64(i),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
		// The account stream represents a single authenticated account.
		if endpoint.recordsPath == "account" {
			return nil
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (API key as the
// username, empty password) and the resolved base URL. The secret only ever flows
// into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dropboxSignBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := dropboxSignSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("dropbox-sign connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: dropboxSignUserAgent,
	}, nil
}

func dropboxSignSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// dropboxSignBaseURL resolves and validates the base URL. The default is
// api.hellosign.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func dropboxSignBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return dropboxSignDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dropbox-sign config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("dropbox-sign config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("dropbox-sign config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func dropboxSignPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dropboxSignDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dropbox-sign config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dropboxSignMaxPageSize {
		return 0, fmt.Errorf("dropbox-sign config page_size must be between 1 and %d", dropboxSignMaxPageSize)
	}
	return value, nil
}

func dropboxSignMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dropbox-sign config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("dropbox-sign config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
