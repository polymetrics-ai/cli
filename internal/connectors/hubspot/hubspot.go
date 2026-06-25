// Package hubspot implements the native pm HubSpot connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + after-cursor pagination) with HubSpot CRM v3 stream
// definitions, endpoints, and write actions.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package hubspot

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
	hubspotDefaultBaseURL  = "https://api.hubapi.com"
	hubspotDefaultPageSize = 100
	hubspotMaxPageSize     = 100
	hubspotUserAgent       = "polymetrics-go-cli"
	hubspotObjectsBase     = "crm/v3/objects"
	// hubspotAccessTokenSecret is the config.Secrets key carrying the private-app
	// Bearer access token (mirrors the catalog secret field
	// credentials.access_token).
	hubspotAccessTokenSecret = "credentials.access_token"
)

func init() {
	connectors.RegisterFactory("hubspot", New)
}

// New returns the HubSpot connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm HubSpot connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "hubspot" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "hubspot",
		DisplayName:     "HubSpot",
		IntegrationType: "api",
		Description:     "Reads HubSpot CRM contacts, companies, deals, and tickets, and writes approved reverse ETL contact actions through the HubSpot CRM v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: true},
	}
}

// Check verifies the connector is configured well enough to talk to HubSpot. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := hubspotBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(hubspotSecret(cfg)) == "" {
		return errors.New("hubspot connector requires secret credentials.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contacts list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, hubspotObjectsBase+"/contacts", q, nil, nil); err != nil {
		return fmt.Errorf("check hubspot: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: hubspotStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a HubSpot stream starts with
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
		stream = "contacts"
	}
	endpoint, ok := hubspotStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("hubspot stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := hubspotPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := hubspotMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives HubSpot CRM v3 after-cursor pagination. List endpoints return
// {results:[...], paging:{next:{after}}}; the next page is requested with
// after=<paging.next.after>. The loop stops when paging.next.after is absent.
// It is built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	base.Set("archived", "false")
	for _, prop := range endpoint.properties {
		base.Add("properties", prop)
	}

	path := hubspotObjectsBase + "/" + endpoint.object
	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read hubspot %s: %w", endpoint.object, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode hubspot %s page: %w", endpoint.object, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next.after")
		if err != nil {
			return fmt.Errorf("decode hubspot %s paging: %w", endpoint.object, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		after = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise hubspot credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		props := map[string]any{
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"firstname":           fmt.Sprintf("Fixture %d", i),
			"lastname":            "Tester",
			"phone":               "+15555550100",
			"company":             "Polymetrics",
			"lifecyclestage":      "lead",
			"name":                fmt.Sprintf("Fixture Co %d", i),
			"domain":              "example.com",
			"industry":            "Software",
			"city":                "Remote",
			"country":             "US",
			"numberofemployees":   "10",
			"dealname":            fmt.Sprintf("Fixture Deal %d", i),
			"amount":              strconv.Itoa(1000 * i),
			"dealstage":           "appointmentscheduled",
			"pipeline":            "default",
			"closedate":           "2026-01-01T00:00:00Z",
			"subject":             fmt.Sprintf("Fixture Ticket %d", i),
			"content":             "Fixture ticket body",
			"hs_pipeline":         "0",
			"hs_pipeline_stage":   "1",
			"hs_ticket_priority":  "HIGH",
			"createdate":          "2026-01-01T00:00:00Z",
			"lastmodifieddate":    "2026-02-01T00:00:00Z",
			"hs_lastmodifieddate": "2026-02-01T00:00:00Z",
		}
		item := map[string]any{
			"id":         fmt.Sprintf("%s_fixture_%d", endpoint.object, i),
			"properties": props,
			"createdAt":  "2026-01-01T00:00:00Z",
			"updatedAt":  fmt.Sprintf("2026-02-0%dT00:00:00Z", i),
			"archived":   false,
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := hubspotBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := hubspotSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("hubspot connector requires secret credentials.access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: hubspotUserAgent,
	}, nil
}

// hubspotSecret resolves the private-app access token from the config secrets.
// It accepts both the dotted catalog key (credentials.access_token) and a bare
// access_token fallback for convenience.
func hubspotSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets[hubspotAccessTokenSecret]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Secrets["access_token"])
}

// hubspotBaseURL resolves and validates the base URL. The default is
// api.hubapi.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func hubspotBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return hubspotDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("hubspot config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("hubspot config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("hubspot config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func hubspotPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return hubspotDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hubspot config page_size must be an integer: %w", err)
	}
	if value < 1 || value > hubspotMaxPageSize {
		return 0, fmt.Errorf("hubspot config page_size must be between 1 and %d", hubspotMaxPageSize)
	}
	return value, nil
}

func hubspotMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hubspot config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("hubspot config max_pages must be 0 for unlimited or a positive integer")
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

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
