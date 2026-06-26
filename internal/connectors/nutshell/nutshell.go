// Package nutshell implements the native pm Nutshell CRM connector. It is a thin
// declarative-HTTP package (modeled on the stripe reference connector) that
// composes the connsdk toolkit (Requester + Basic auth + RecordsAt extraction)
// with Nutshell-specific stream definitions, endpoints, and pagination.
//
// Nutshell's REST API (https://app.nutshell.com/rest/) authenticates with HTTP
// Basic auth using a username and an API token (the "password" secret) and
// paginates list endpoints with page[page] (0-indexed) + page[limit]. Records
// live under a per-stream top-level envelope key (e.g. {"contacts":[...]}).
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. The connector is read-only: Nutshell writes are
// not exposed as reverse-ETL actions here.
package nutshell

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
	nutshellDefaultBaseURL  = "https://app.nutshell.com/rest"
	nutshellDefaultPageSize = 500
	nutshellMaxPageSize     = 500
	nutshellUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("nutshell", New)
}

// New returns the Nutshell connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Nutshell connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nutshell" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nutshell",
		DisplayName:     "Nutshell",
		IntegrationType: "api",
		Description:     "Reads Nutshell CRM accounts, contacts, leads, activities, and users through the Nutshell REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Nutshell. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nutshellBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(nutshellUsername(cfg)) == "" {
		return errors.New("nutshell connector requires config username")
	}
	if strings.TrimSpace(nutshellSecret(cfg)) == "" {
		return errors.New("nutshell connector requires secret password (API token)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"page[limit]": []string{"1"}, "page[page]": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "users", q, nil, nil); err != nil {
		return fmt.Errorf("check nutshell: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nutshellStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := nutshellStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("nutshell stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := nutshellPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := nutshellMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Nutshell's page[page] pagination. List endpoints return
// {<recordsKey>:[...]}; the next page is requested by incrementing page[page]
// (0-indexed). A page shorter than the requested page[limit] terminates the loop.
// Reference endpoints flagged paginated=false are read once.
//
// connsdk has no body-key paginator for this page[]-bracketed param shape, so the
// loop lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("page[page]", strconv.Itoa(page))
			query.Set("page[limit]", strconv.Itoa(pageSize))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read nutshell %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode nutshell %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Single-page reference endpoints, an empty page, or a short page all stop.
		if !endpoint.paginated || len(records) == 0 || len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise nutshell credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              int64(i),
			"entityType":      "Fixture",
			"name":            fmt.Sprintf("%s fixture %d", stream, i),
			"description":     fmt.Sprintf("fixture %d", i),
			"url":             "https://example.com",
			"htmlUrl":         "https://example.com",
			"value":           "$1,000",
			"status":          int64(0),
			"confidence":      int64(50),
			"industryId":      int64(1),
			"accountTypeId":   int64(1),
			"activityTypeId":  int64(1),
			"isHotLead":       false,
			"isOverdue":       false,
			"isFlagged":       false,
			"isEnabled":       true,
			"isAdministrator": false,
			"emails":          fmt.Sprintf("fixture+%d@example.com", i),
			"logNote":         "",
			"createdTime":     "2026-01-01T00:00:00+0000",
			"modifiedTime":    fmt.Sprintf("2026-01-0%dT00:00:00+0000", i),
			"closedTime":      nil,
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

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The secret only ever flows into connsdk.Basic; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := nutshellBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(nutshellUsername(cfg))
	if username == "" {
		return nil, errors.New("nutshell connector requires config username")
	}
	secret := nutshellSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("nutshell connector requires secret password (API token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: nutshellUserAgent,
	}, nil
}

func nutshellUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func nutshellSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// nutshellBaseURL resolves and validates the base URL. The default is
// app.nutshell.com/rest; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func nutshellBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return nutshellDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("nutshell config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nutshell config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nutshell config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func nutshellPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return nutshellDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nutshell config page_size must be an integer: %w", err)
	}
	if value < 1 || value > nutshellMaxPageSize {
		return 0, fmt.Errorf("nutshell config page_size must be between 1 and %d", nutshellMaxPageSize)
	}
	return value, nil
}

func nutshellMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nutshell config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("nutshell config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Nutshell is exposed
// read-only, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
