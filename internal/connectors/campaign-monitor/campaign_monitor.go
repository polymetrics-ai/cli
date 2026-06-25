// Package campaignmonitor implements the native pm Campaign Monitor connector.
// It is a declarative-HTTP per-system connector following the stripe template: a
// thin package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction) with Campaign Monitor stream definitions, endpoints, and
// pagination.
//
// Campaign Monitor (createsend.com) v3.3 authenticates with HTTP Basic auth: the
// account/client API key is sent as the username and the password may be blank
// or a dummy value. The catalog exposes "username" (config) and "password"
// (secret), so this connector wires Basic(username, password).
//
// Like the other per-system connectors it self-registers with the connectors
// registry via RegisterFactory in init(); the registryset package blank-imports
// this package in the production binary to run that side effect.
package campaignmonitor

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
	registryName          = "campaign-monitor"
	cmDefaultBaseURL      = "https://api.createsend.com/api/v3.3"
	cmDefaultPageSize     = 100
	cmMaxPageSize         = 1000
	cmUserAgent           = "polymetrics-go-cli"
	cmFixtureClientID     = "cli_fixture_1"
	cmFixtureSentDate     = "2026-01-01 10:00:00"
	cmFixtureMaxFixtures  = 2
	cmResultsPath         = "Results"
	cmPageNumberPath      = "PageNumber"
	cmNumberOfPagesPath   = "NumberOfPages"
	cmDefaultDummyPasswrd = "x"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Campaign Monitor connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Campaign Monitor connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Campaign Monitor",
		IntegrationType: "api",
		Description:     "Reads Campaign Monitor clients, campaigns, subscriber lists, and suppression lists through the createsend.com v3.3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Campaign
// Monitor. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := cmBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cmUsername(cfg)) == "" {
		return errors.New("campaign-monitor connector requires config username (the API key)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the clients list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "clients.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check campaign-monitor: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: campaignMonitorStreams()}, nil
}

// Write is unsupported: Campaign Monitor is exposed as a read-only source. The
// method exists to satisfy the connectors.Connector interface and always returns
// ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "clients"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("campaign-monitor stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := cmPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := cmMaxPages(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.paged {
		return c.readArray(ctx, r, resource, endpoint, emit)
	}
	return c.harvestPaged(ctx, r, resource, endpoint, pageSize, maxPages, emit)
}

// readArray reads a bare top-level JSON array endpoint (clients, lists) in a
// single request.
func (c Connector) readArray(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read campaign-monitor %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode campaign-monitor %s: %w", resource, err)
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

// harvestPaged drives Campaign Monitor's page/NumberOfPages pagination. Paged
// list endpoints return {Results:[...], PageNumber:n, NumberOfPages:m}; the loop
// requests page+1 until PageNumber reaches NumberOfPages.
func (c Connector) harvestPaged(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for pageCount := 0; maxPages == 0 || pageCount < maxPages; pageCount++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("pagesize", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read campaign-monitor %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, cmResultsPath)
		if err != nil {
			return fmt.Errorf("decode campaign-monitor %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		numberOfPages := atoiOr(stringAt(resp.Body, cmNumberOfPagesPath), 0)
		current := atoiOr(stringAt(resp.Body, cmPageNumberPath), page)
		if numberOfPages == 0 || current >= numberOfPages {
			return nil
		}
		page = current + 1
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= cmFixtureMaxFixtures; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"ClientID":          fmt.Sprintf("%s_%d", cmFixtureClientID, i),
			"Name":              fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"CampaignID":        fmt.Sprintf("camp_fixture_%d", i),
			"Subject":           fmt.Sprintf("Fixture subject %d", i),
			"FromName":          "Fixture Sender",
			"FromEmail":         "sender@example.com",
			"ReplyTo":           "reply@example.com",
			"WebVersionURL":     "https://example.com/view",
			"WebVersionTextURL": "https://example.com/view.txt",
			"SentDate":          cmFixtureSentDate,
			"TotalRecipients":   int64(100 * i),
			"ListID":            fmt.Sprintf("list_fixture_%d", i),
			"EmailAddress":      fmt.Sprintf("fixture+%d@example.com", i),
			"Date":              cmFixtureSentDate,
			"State":             "Suppressed",
			"SuppressionType":   "Manual",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The password secret only ever flows into connsdk.Basic; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := cmBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(cmUsername(cfg))
	if username == "" {
		return nil, errors.New("campaign-monitor connector requires config username (the API key)")
	}
	password := cmPassword(cfg)
	if strings.TrimSpace(password) == "" {
		// Campaign Monitor allows a blank or dummy password alongside the API key.
		password = cmDefaultDummyPasswrd
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, password),
		UserAgent: cmUserAgent,
	}, nil
}

// resolveResource fills the client_id into a scoped endpoint's path template.
func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.scoped {
		return endpoint.resource, nil
	}
	clientID := strings.TrimSpace(cfg.Config["client_id"])
	if clientID == "" {
		return "", errors.New("campaign-monitor config client_id is required for this stream")
	}
	return fmt.Sprintf(endpoint.resource, url.PathEscape(clientID)), nil
}

func cmUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func cmPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// cmBaseURL resolves and validates the base URL. The default is
// api.createsend.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func cmBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return cmDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("campaign-monitor config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("campaign-monitor config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("campaign-monitor config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cmPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return cmDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("campaign-monitor config page_size must be an integer: %w", err)
	}
	if value < 1 || value > cmMaxPageSize {
		return 0, fmt.Errorf("campaign-monitor config page_size must be between 1 and %d", cmMaxPageSize)
	}
	return value, nil
}

func cmMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("campaign-monitor config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("campaign-monitor config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// stringAt reads a dotted path from a JSON body, swallowing decode errors as an
// empty string (callers treat empty as "absent").
func stringAt(body []byte, path string) string {
	s, err := connsdk.StringAt(body, path)
	if err != nil {
		return ""
	}
	return s
}

func atoiOr(s string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return v
}
