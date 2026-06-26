// Package emailoctopus implements the native pm EmailOctopus connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a
// Requester with api_key query auth, RecordsAt extraction over the EmailOctopus
// {data:[...],paging:{next}} envelope, and page-following pagination.
//
// It mirrors the stripe reference connector's shape and self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
//
// EmailOctopus API v1.6 is read-only here: the connector exposes lists,
// campaigns, and list_contacts streams over full_refresh syncs (the API offers
// no incremental cursor), so Capabilities.Write is false.
package emailoctopus

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
	emailOctopusDefaultBaseURL  = "https://emailoctopus.com/api/1.6"
	emailOctopusDefaultPageSize = 100
	emailOctopusMaxPageSize     = 100
	emailOctopusUserAgent       = "polymetrics-go-cli"
	// emailOctopusFixtureCreated is the deterministic created_at timestamp used
	// by fixture-mode records.
	emailOctopusFixtureCreated = "2026-01-01T00:00:00+00:00"
)

func init() {
	connectors.RegisterFactory("emailoctopus", New)
}

// New returns the EmailOctopus connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm EmailOctopus connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "emailoctopus" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "emailoctopus",
		DisplayName:     "EmailOctopus",
		IntegrationType: "api",
		Description:     "Reads EmailOctopus lists, campaigns, and list contacts through the EmailOctopus v1.6 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to
// EmailOctopus. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := emailOctopusBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(emailOctopusSecret(cfg)) == "" {
		return errors.New("emailoctopus connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the lists endpoint confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "lists", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check emailoctopus: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. EmailOctopus is exposed
// read-only here, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: emailOctopusStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "lists"
	}
	endpoint, ok := emailOctopusStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("emailoctopus stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	resource, err := resolveResource(stream, endpoint.resource, req.Config)
	if err != nil {
		return err
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := emailOctopusPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := emailOctopusMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint.mapRecord, pageSize, maxPages, emit)
}

// harvest drives EmailOctopus page-based pagination. List responses are shaped
// {data:[...], paging:{next:<url|null>, previous:<url|null>}}; the next page is
// the absolute URL at paging.next, or the loop stops when it is null. The
// Requester treats an http(s) path as absolute, so paging.next is followed
// directly; the api_key authenticator re-applies the credential on each hop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := resource
	query := url.Values{"limit": []string{strconv.Itoa(pageSize)}}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read emailoctopus %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode emailoctopus %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return fmt.Errorf("decode emailoctopus %s paging: %w", resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// Follow the absolute next URL as-is; reset the per-request query so the
		// page/limit carried in the next URL are not overwritten.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise emailoctopus credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		switch stream {
		case "campaigns":
			item = map[string]any{
				"id":         fmt.Sprintf("campaign_fixture_%d", i),
				"status":     "SENT",
				"name":       fmt.Sprintf("Fixture Campaign %d", i),
				"subject":    fmt.Sprintf("Subject %d", i),
				"from":       map[string]any{"name": "Fixture Sender", "email_address": "sender@example.com"},
				"created_at": emailOctopusFixtureCreated,
				"sent_at":    emailOctopusFixtureCreated,
			}
		case "list_contacts":
			item = map[string]any{
				"id":              fmt.Sprintf("contact_fixture_%d", i),
				"email_address":   fmt.Sprintf("fixture+%d@example.com", i),
				"status":          "SUBSCRIBED",
				"tags":            []any{"vip"},
				"fields":          map[string]any{"FirstName": fmt.Sprintf("Fixture%d", i)},
				"created_at":      emailOctopusFixtureCreated,
				"last_updated_at": emailOctopusFixtureCreated,
			}
		default: // lists
			item = map[string]any{
				"id":            fmt.Sprintf("list_fixture_%d", i),
				"name":          fmt.Sprintf("Fixture List %d", i),
				"double_opt_in": false,
				"counts":        map[string]any{"pending": 0, "subscribed": 10 * i, "unsubscribed": i},
				"created_at":    emailOctopusFixtureCreated,
			}
		}
		endpoint := emailOctopusStreamEndpoints[stream]
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api_key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := emailOctopusBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := emailOctopusSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("emailoctopus connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("api_key", secret),
		UserAgent: emailOctopusUserAgent,
	}, nil
}

// resolveResource substitutes the {list_id} path template for parent-scoped
// streams (list_contacts) from config, and validates it.
func resolveResource(stream, resource string, cfg connectors.RuntimeConfig) (string, error) {
	if !strings.Contains(resource, "{list_id}") {
		return resource, nil
	}
	listID := strings.TrimSpace(cfg.Config["list_id"])
	if listID == "" {
		return "", fmt.Errorf("emailoctopus stream %q requires config list_id", stream)
	}
	return strings.ReplaceAll(resource, "{list_id}", url.PathEscape(listID)), nil
}

func emailOctopusSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// emailOctopusBaseURL resolves and validates the base URL. The default is
// emailoctopus.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func emailOctopusBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return emailOctopusDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("emailoctopus config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("emailoctopus config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("emailoctopus config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func emailOctopusPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return emailOctopusDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("emailoctopus config page_size must be an integer: %w", err)
	}
	if value < 1 || value > emailOctopusMaxPageSize {
		return 0, fmt.Errorf("emailoctopus config page_size must be between 1 and %d", emailOctopusMaxPageSize)
	}
	return value, nil
}

func emailOctopusMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("emailoctopus config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("emailoctopus config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
