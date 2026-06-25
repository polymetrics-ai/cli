// Package zendesktalk implements the native pm Zendesk Talk connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, copying
// the stripe reference shape: a thin package that composes connsdk.Requester
// with Basic (API token) or Bearer (OAuth) auth, connsdk.RecordsAt extraction,
// and Zendesk's next_page body pagination, plus Talk-specific stream defs.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. The directory and registry key are "zendesk-talk"; the
// Go package identifier is "zendesktalk".
package zendesktalk

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
	zendeskTalkAPIPath    = "/api/v2/channels/voice"
	zendeskTalkPageSize   = 100
	zendeskTalkUserAgent  = "polymetrics-go-cli"
	zendeskTalkFixtureTS  = "2026-01-01T00:00:00Z"
	registryKey           = "zendesk-talk"
	zendeskTalkMaxPagesNo = 0 // unlimited
)

func init() {
	connectors.RegisterFactory(registryKey, New)
}

// New returns the Zendesk Talk connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Zendesk Talk connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryKey }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryKey,
		DisplayName:     "Zendesk Talk",
		IntegrationType: "api",
		Description:     "Reads Zendesk Talk phone numbers, greetings, greeting categories, IVRs, and agent activity statistics through the Zendesk Talk (voice) REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Zendesk
// Talk. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := zendeskTalkBaseURL(cfg); err != nil {
		return err
	}
	if _, err := zendeskTalkAuth(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of greeting_categories confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "greeting_categories", nil, nil, nil); err != nil {
		return fmt.Errorf("check zendesk-talk: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: zendeskTalkStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "phone_numbers"
	}
	endpoint, ok := zendeskTalkStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("zendesk-talk stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := zendeskTalkMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, maxPages, emit)
}

// harvest drives Zendesk's next_page body pagination. A Talk list response is
// {<key>:[...], "next_page": "<absolute url or null>"}; the next page is fetched
// by requesting next_page verbatim until it is null. The connsdk Requester
// treats an absolute http(s) path as-is, so next_page is passed straight
// through. The loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := url.Values{}
	query.Set("per_page", strconv.Itoa(zendeskTalkPageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read zendesk-talk %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode zendesk-talk %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page")
		if err != nil {
			return fmt.Errorf("decode zendesk-talk %s next_page: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "null" {
			return nil
		}
		// next_page is a full URL with its own query (page/cursor); requesting it
		// verbatim means we must not re-append our own params.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise zendesk-talk credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"agent_id":            int64(i),
			"number":              fmt.Sprintf("+1555000%d", i),
			"display_number":      fmt.Sprintf("+1 (555) 000-%d", i),
			"nickname":            fmt.Sprintf("fixture-%d", i),
			"country_code":        "US",
			"toll_free":           false,
			"voice_enabled":       true,
			"sms_enabled":         false,
			"recorded":            true,
			"created_at":          zendeskTalkFixtureTS,
			"name":                fmt.Sprintf("Fixture %d", i),
			"category_id":         int64(1),
			"default":             i == 1,
			"active":              true,
			"audio_name":          fmt.Sprintf("greeting-%d.mp3", i),
			"audio_url":           fmt.Sprintf("https://example.com/audio/%d", i),
			"has_sub_settings":    false,
			"phone_number_ids":    []any{int64(i)},
			"phone_number_names":  []any{fmt.Sprintf("+1555000%d", i)},
			"avatar_url":          fmt.Sprintf("https://example.com/avatar/%d", i),
			"agent_state":         "available",
			"call_status":         "on_call",
			"available_time":      int64(100 * i),
			"away_time":           int64(10 * i),
			"online_time":         int64(110 * i),
			"calls_accepted":      int64(i),
			"calls_denied":        int64(0),
			"calls_missed":        int64(0),
			"forwarding_number":   "",
			"total_call_duration": int64(60 * i),
			"total_talk_time":     int64(55 * i),
			"total_wrap_up_time":  int64(5 * i),
			"via":                 "browser",
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

// requester builds a connsdk.Requester wired with the resolved auth and base
// URL. Secrets only ever flow into the connsdk Authenticator; they are never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := zendeskTalkBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := zendeskTalkAuth(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: zendeskTalkUserAgent,
	}, nil
}

// zendeskTalkAuth resolves the Authenticator from the configured credentials.
// API-token auth uses HTTP Basic with username "{email}/token" and password
// "{api_token}"; OAuth uses Bearer "{access_token}". Bearer takes precedence
// when an access token is present.
func zendeskTalkAuth(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	if cfg.Secrets == nil {
		return nil, errors.New("zendesk-talk connector requires credentials (access_token, or email + api_token)")
	}
	accessToken := strings.TrimSpace(secret(cfg, "credentials.access_token"))
	if accessToken != "" {
		return connsdk.Bearer(accessToken), nil
	}
	apiToken := strings.TrimSpace(secret(cfg, "credentials.api_token"))
	email := strings.TrimSpace(firstNonEmpty(secret(cfg, "credentials.email"), cfg.Config["email"]))
	if apiToken != "" {
		if email == "" {
			return nil, errors.New("zendesk-talk API token auth requires an email (credentials.email)")
		}
		return connsdk.Basic(email+"/token", apiToken), nil
	}
	return nil, errors.New("zendesk-talk connector requires credentials (access_token, or email + api_token)")
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// zendeskTalkBaseURL resolves and validates the base URL. The default is derived
// from the required subdomain config: https://{subdomain}.zendesk.com/api/v2/
// channels/voice. A base_url override (used by tests/proxies) must be an
// absolute http(s) URL with a host to bound SSRF risk; when overridden the Talk
// API path is appended automatically.
func zendeskTalkBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	override := strings.TrimSpace(cfg.Config["base_url"])
	if override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("zendesk-talk config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("zendesk-talk config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("zendesk-talk config base_url must include a host")
		}
		trimmed := strings.TrimRight(override, "/")
		if strings.Contains(parsed.Path, zendeskTalkAPIPath) {
			return trimmed, nil
		}
		return trimmed + zendeskTalkAPIPath, nil
	}

	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("zendesk-talk connector requires config subdomain")
	}
	if !validSubdomain(subdomain) {
		return "", fmt.Errorf("zendesk-talk config subdomain %q is invalid", subdomain)
	}
	return "https://" + subdomain + ".zendesk.com" + zendeskTalkAPIPath, nil
}

// validSubdomain bounds the subdomain to a DNS-safe label set to prevent host
// injection into the constructed base URL.
func validSubdomain(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func zendeskTalkMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return zendeskTalkMaxPagesNo, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("zendesk-talk config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("zendesk-talk config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Zendesk Talk exposes no safe reverse-ETL writes for this
// connector's stream set, so it is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
