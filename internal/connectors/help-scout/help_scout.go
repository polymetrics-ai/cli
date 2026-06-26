// Package helpscout implements the native pm Help Scout connector. It follows the
// stripe declarative-HTTP template: a thin package that composes the connsdk
// toolkit (Requester + OAuth2 client-credentials auth + RecordsAt extraction +
// cursor state) with Help Scout-specific stream definitions, endpoints, and HAL+
// JSON page pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. The registry key is the bare system name "help-scout".
package helpscout

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
	helpScoutDefaultBaseURL  = "https://api.helpscout.net/v2"
	helpScoutDefaultTokenURL = "https://api.helpscout.net/v2/oauth2/token"
	helpScoutDefaultPageSize = 50
	helpScoutMaxPageSize     = 50
	helpScoutUserAgent       = "polymetrics-go-cli"
	// helpScoutFixtureTime is the deterministic timestamp used by fixture-mode
	// records (2026-01-01T00:00:00Z).
	helpScoutFixtureTime = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("help-scout", New)
}

// New returns the Help Scout connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Help Scout connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "help-scout" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "help-scout",
		DisplayName:     "Help Scout",
		IntegrationType: "api",
		Description:     "Reads Help Scout conversations, customers, mailboxes, and users through the Mailbox API using OAuth2 client-credentials authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Help Scout.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := helpScoutBaseURL(cfg); err != nil {
		return err
	}
	id, secret := helpScoutCreds(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("help-scout connector requires secrets client_id and client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the mailboxes list confirms the OAuth2 exchange and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "mailboxes", url.Values{"page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check help-scout: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: helpScoutStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Help Scout stream starts
// with an empty incremental cursor (full sync). Help Scout only supports
// full_refresh upstream, but an empty cursor keeps the interface uniform.
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
		stream = "conversations"
	}
	endpoint, ok := helpScoutStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("help-scout stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := helpScoutPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := helpScoutMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, req, emit)
}

// harvest drives Help Scout's HAL+JSON page pagination. List responses return
// {_embedded:{<resource>:[...]}, page:{number, totalPages, ...}}; the next page
// is requested with page=<n+1> until page.number reaches page.totalPages. The
// page count lives in the body rather than being signalled by a short page, so
// this loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("size", strconv.Itoa(pageSize))
	base.Set("sortField", "modifiedAt")
	base.Set("sortOrder", "asc")
	if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
		// modifiedSince scopes conversations/customers to records changed at or
		// after start_date; ignored by endpoints that don't support it.
		base.Set("modifiedSince", start)
	}

	recordsPath := "_embedded." + endpoint.embeddedKey
	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read help-scout %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode help-scout %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		totalPages := pageInt(resp.Body, "page.totalPages")
		if totalPages <= 0 {
			// Defensive: if the page envelope is missing, stop when a page returns
			// fewer records than the page size.
			if len(records) < pageSize {
				return nil
			}
		} else if page >= totalPages {
			return nil
		}
		if len(records) == 0 {
			return nil
		}
		page++
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise help-scout credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            int64(i),
			"number":        int64(100 + i),
			"type":          "email",
			"status":        "active",
			"state":         "published",
			"subject":       fmt.Sprintf("Fixture conversation %d", i),
			"mailboxId":     int64(1),
			"folderId":      int64(1),
			"assigneeId":    int64(1),
			"preview":       "fixture preview",
			"threads":       int64(2),
			"firstName":     fmt.Sprintf("Fixture%d", i),
			"lastName":      "Example",
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"gender":        "unknown",
			"jobTitle":      "Tester",
			"organization":  "Polymetrics",
			"photoUrl":      "",
			"age":           "30",
			"name":          fmt.Sprintf("Fixture Mailbox %d", i),
			"slug":          fmt.Sprintf("fixture-%d", i),
			"role":          "user",
			"timezone":      "UTC",
			"createdAt":     helpScoutFixtureTime,
			"closedAt":      nil,
			"updatedAt":     helpScoutFixtureTime,
			"userUpdatedAt": helpScoutFixtureTime,
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

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved base URL. The client_id/client_secret only ever flow into the
// OAuth2 token exchange; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := helpScoutBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	id, secret := helpScoutCreds(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("help-scout connector requires secrets client_id and client_secret")
	}
	tokenURL, err := helpScoutTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     tokenURL,
		ClientID:     id,
		ClientSecret: secret,
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: helpScoutUserAgent,
	}, nil
}

func helpScoutCreds(cfg connectors.RuntimeConfig) (string, string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["client_id"], cfg.Secrets["client_secret"]
}

// helpScoutBaseURL resolves and validates the base URL. The default is
// api.helpscout.net; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func helpScoutBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validateURL(cfg.Config["base_url"], helpScoutDefaultBaseURL, "base_url")
}

// helpScoutTokenURL resolves and validates the OAuth2 token endpoint, applying
// the same scheme/host SSRF guard as the base URL.
func helpScoutTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	return validateURL(cfg.Config["token_url"], helpScoutDefaultTokenURL, "token_url")
}

func validateURL(raw, fallback, field string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback, nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("help-scout config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("help-scout config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("help-scout config %s must include a host", field)
	}
	return strings.TrimRight(value, "/"), nil
}

func helpScoutPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return helpScoutDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("help-scout config page_size must be an integer: %w", err)
	}
	if value < 1 || value > helpScoutMaxPageSize {
		return 0, fmt.Errorf("help-scout config page_size must be between 1 and %d", helpScoutMaxPageSize)
	}
	return value, nil
}

func helpScoutMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("help-scout config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("help-scout config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// pageInt reads an integer value out of the HAL page envelope at the dotted path
// (e.g. "page.totalPages"). Missing or non-numeric values return 0.
func pageInt(body []byte, path string) int {
	value, err := connsdk.StringAt(body, path)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return n
}

// Write is unsupported: Help Scout is read-only for this connector. It satisfies
// the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
