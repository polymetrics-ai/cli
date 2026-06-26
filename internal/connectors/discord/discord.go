// Package discord implements the native pm Discord source connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a
// Requester with Discord's `Authorization: Bot <token>` auth, top-level-array
// and single-object extraction, and Discord's snowflake `after`=highest-id
// cursor pagination for the members stream.
//
// It mirrors the stripe reference connector's shape and self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
//
// Discord is read-only here (the upstream source supports full_refresh only), so
// Capabilities.Write is false.
package discord

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
	discordDefaultBaseURL  = "https://discord.com/api/v10"
	discordDefaultPageSize = 100
	discordMaxPageSize     = 1000
	// discordUserAgent follows Discord's required DiscordBot User-Agent format;
	// requests without a valid User-Agent may be blocked by Cloudflare.
	discordUserAgent = "DiscordBot (https://polymetrics.ai, 2.0) polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("discord", New)
}

// New returns the Discord connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Discord source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "discord" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "discord",
		DisplayName:     "Discord",
		IntegrationType: "api",
		Description:     "Reads Discord guild, channel, role, and member data through the Discord REST API using a bot token.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Discord. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := discordBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(discordSecret(cfg)) == "" {
		return errors.New("discord connector requires secret bot_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// GET /users/@me confirms the bot token authenticates without needing a guild
	// id or mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users/@me", nil, nil, nil); err != nil {
		return fmt.Errorf("check discord: %w", err)
	}
	return nil
}

// Write is unsupported: Discord is a read-only source connector. The method
// exists to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: discordStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "guilds"
	}
	endpoint, ok := discordStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("discord stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	guildID, err := discordGuildID(req.Config)
	if err != nil {
		return err
	}
	path := strings.ReplaceAll(endpoint.pathTemplate, "{guild_id}", url.PathEscape(guildID))

	switch endpoint.pagination {
	case pageSingle:
		return c.readSingle(ctx, r, path, endpoint, emit)
	case pageNone:
		return c.readArray(ctx, r, path, nil, endpoint, emit)
	case pageAfterID:
		pageSize, err := discordPageSize(req.Config)
		if err != nil {
			return err
		}
		maxPages, err := discordMaxPages(req.Config)
		if err != nil {
			return err
		}
		return c.readMembers(ctx, r, path, pageSize, maxPages, endpoint, emit)
	default:
		return fmt.Errorf("discord stream %q has unknown pagination", stream)
	}
}

// readSingle reads a single-object resource (e.g. a guild) and emits one record.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read discord %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode discord %s: %w", path, err)
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

// readArray reads an unpaginated top-level array resource (e.g. channels, roles).
func (c Connector) readArray(ctx context.Context, r *connsdk.Requester, path string, query url.Values, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return fmt.Errorf("read discord %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode discord %s: %w", path, err)
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

// readMembers drives Discord's snowflake `after`=highest-id cursor pagination.
// The members list returns a top-level array of member objects in ascending user
// id order; the next page is requested with after=<last member's user id>. The
// loop stops on a short page (fewer than pageSize records) or when maxPages is
// reached. There is no connsdk paginator for the user-id-from-record shape, so
// the loop lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) readMembers(ctx context.Context, r *connsdk.Requester, path string, pageSize, maxPages int, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read discord %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode discord %s page: %w", path, err)
		}
		lastAfter := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if id := memberAfterID(item); id != "" {
				lastAfter = id
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize || lastAfter == "" {
			return nil
		}
		after = lastAfter
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise discord credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	count := 2
	if endpoint.pagination == pageSingle {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", stream, i)
		item := map[string]any{
			"id":                         id,
			"name":                       fmt.Sprintf("Fixture %s %d", stream, i),
			"guild_id":                   "guild_fixture_1",
			"owner_id":                   "user_fixture_1",
			"type":                       0,
			"position":                   i - 1,
			"color":                      0,
			"permissions":                "0",
			"approximate_member_count":   int64(2),
			"approximate_presence_count": int64(1),
			"premium_tier":               int64(0),
			"preferred_locale":           "en-US",
			"nick":                       fmt.Sprintf("Fixture %d", i),
			"joined_at":                  "2026-01-01T00:00:00Z",
			"roles":                      []any{},
			"user": map[string]any{
				"id":          id,
				"username":    fmt.Sprintf("fixture_user_%d", i),
				"global_name": fmt.Sprintf("Fixture User %d", i),
				"bot":         false,
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Discord's Bot auth header, the
// resolved base URL, and the required DiscordBot User-Agent. The secret only ever
// flows into the Authorization header; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := discordBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := discordSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("discord connector requires secret bot_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "Bot "),
		UserAgent: discordUserAgent,
	}, nil
}

func discordSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["bot_token"]
}

// discordGuildID resolves the required guild_id config for guild-scoped streams.
func discordGuildID(cfg connectors.RuntimeConfig) (string, error) {
	id := strings.TrimSpace(cfg.Config["guild_id"])
	if id == "" {
		return "", errors.New("discord connector requires config guild_id")
	}
	return id, nil
}

// discordBaseURL resolves and validates the base URL. The default is
// discord.com/api/v10; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func discordBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return discordDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("discord config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("discord config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("discord config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func discordPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return discordDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("discord config page_size must be an integer: %w", err)
	}
	if value < 1 || value > discordMaxPageSize {
		return 0, fmt.Errorf("discord config page_size must be between 1 and %d", discordMaxPageSize)
	}
	return value, nil
}

func discordMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("discord config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("discord config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
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
