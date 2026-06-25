// Package slack implements the native pm Slack connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: it
// composes a Requester + Bearer (xoxb bot token) auth + Slack's cursor
// pagination + per-stream record mappers. It mirrors the stripe reference
// connector's shape, swapping in Slack's base URL, auth, pagination style, and
// streams.
//
// Slack's Web API is read-only here (the connector exposes no reverse-ETL
// writes). Like stripe/github it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package slack

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
	slackDefaultBaseURL  = "https://slack.com/api"
	slackDefaultPageSize = 200
	slackMaxPageSize     = 1000
	slackUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("slack", New)
}

// New returns the Slack connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Slack connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "slack" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "slack",
		DisplayName:     "Slack",
		IntegrationType: "api",
		Description:     "Reads Slack workspace users, channels, and channel messages through the Slack Web API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Slack. In
// fixture mode it short-circuits without a network call. Otherwise it calls
// auth.test, which validates the token without reading any data.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := slackBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(slackSecret(cfg)) == "" {
		return errors.New("slack connector requires secret api_token (xoxb- bot token)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, "auth.test", nil, nil)
	if err != nil {
		return fmt.Errorf("check slack: %w", err)
	}
	if err := slackOK(resp.Body, "auth.test"); err != nil {
		return fmt.Errorf("check slack: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Slack is read-only here:
// the connector exposes no reverse-ETL write actions, so every write is refused.
// Capabilities.Write is false, and no WriteValidator/DryRunWriter is implemented.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: slackStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	endpoint, ok := slackStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("slack stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := slackPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := slackMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	// channel_messages reads a single channel's history and needs a channel id.
	if stream == "channel_messages" {
		channel := strings.TrimSpace(req.Config.Config["channel_id"])
		if channel == "" {
			return errors.New("slack stream channel_messages requires config channel_id")
		}
		base.Set("channel", channel)
	}
	return c.harvest(ctx, r, endpoint, base, maxPages, emit)
}

// harvest drives Slack's cursor pagination. List methods return
// {ok:true, <listKey>:[...], response_metadata:{next_cursor}}. The next page is
// requested with cursor=<next_cursor>; the loop stops when next_cursor is empty.
// A Slack API error is signalled by ok:false with an error code at HTTP 200, so
// every page is checked via slackOK. There is no off-the-shelf connsdk
// paginator that also enforces the ok-flag, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read slack %s: %w", endpoint.resource, err)
		}
		if err := slackOK(resp.Body, endpoint.resource); err != nil {
			return err
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.listKey)
		if err != nil {
			return fmt.Errorf("decode slack %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "response_metadata.next_cursor")
		if err != nil {
			return fmt.Errorf("decode slack %s next_cursor: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise slack credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(stream, "s"), i),
			"ts":          fmt.Sprintf("170000000%d.000100", i),
			"team_id":     "T_FIXTURE",
			"team":        "T_FIXTURE",
			"name":        fmt.Sprintf("fixture-%d", i),
			"real_name":   fmt.Sprintf("Fixture %d", i),
			"deleted":     false,
			"is_bot":      false,
			"is_channel":  true,
			"is_archived": false,
			"is_private":  false,
			"created":     int64(1700000000 + i),
			"updated":     int64(1700000000 + i),
			"num_members": int64(i),
			"type":        "message",
			"user":        "U_FIXTURE",
			"text":        fmt.Sprintf("fixture message %d", i),
			"profile": map[string]any{
				"email":        fmt.Sprintf("fixture+%d@example.com", i),
				"display_name": fmt.Sprintf("fixture-%d", i),
			},
			"topic":   map[string]any{"value": "fixture topic"},
			"purpose": map[string]any{"value": "fixture purpose"},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := slackBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := slackSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("slack connector requires secret api_token (xoxb- bot token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: slackUserAgent,
	}, nil
}

// slackOK inspects a Slack Web API response body for the ok flag. Slack returns
// HTTP 200 even for logical failures, carrying {ok:false, error:"<code>"}.
func slackOK(body []byte, method string) error {
	ok, err := connsdk.StringAt(body, "ok")
	if err != nil {
		return fmt.Errorf("decode slack %s response: %w", method, err)
	}
	if ok == "true" {
		return nil
	}
	code, _ := connsdk.StringAt(body, "error")
	if strings.TrimSpace(code) == "" {
		code = "unknown_error"
	}
	return fmt.Errorf("slack %s returned error: %s", method, code)
}

// slackSecret resolves the bot token from secrets. Both api_token and
// access_token are accepted to match the catalog's two auth shapes (API Token
// Credentials and OAuth2 access_token).
func slackSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{"api_token", "access_token", "credentials.api_token", "credentials.access_token"} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// slackBaseURL resolves and validates the base URL. The default is
// slack.com/api; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func slackBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return slackDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("slack config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("slack config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("slack config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func slackPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return slackDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("slack config page_size must be an integer: %w", err)
	}
	if value < 1 || value > slackMaxPageSize {
		return 0, fmt.Errorf("slack config page_size must be between 1 and %d", slackMaxPageSize)
	}
	return value, nil
}

func slackMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("slack config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("slack config max_pages must be 0 for unlimited or a positive integer")
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
