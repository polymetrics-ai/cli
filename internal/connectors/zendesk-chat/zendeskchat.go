// Package zendeskchat implements the native pm Zendesk Chat connector. It follows
// the declarative-HTTP per-system template established by the stripe package: a
// thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + cursor state) with Zendesk-Chat-specific stream
// definitions, endpoints, and a fixture mode.
//
// Zendesk Chat authenticates with an OAuth access token sent as
// "Authorization: Bearer <token>". The current base URL is
// https://{subdomain}.zendesk.com/api/v2/chat. Simple list endpoints (agents,
// departments, shortcuts, triggers) return a top-level JSON array; the chats
// endpoint is an incremental-export shape ({chats:[...], next_url:..., count:N})
// that is followed by chasing next_url until it is empty.
//
// Like stripe and github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The registry key is the bare
// system name "zendesk-chat" even though the Go package identifier is
// "zendeskchat" (Go package names cannot contain hyphens).
package zendeskchat

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	zendeskChatName      = "zendesk-chat"
	zendeskChatUserAgent = "polymetrics-go-cli"
	zendeskChatMaxPages  = 1000
	// zendeskChatAccessTokenSecret is the secret key carrying the OAuth access
	// token. The catalog nests it under credentials.access_token.
	zendeskChatAccessTokenSecret = "credentials.access_token"
)

func init() {
	connectors.RegisterFactory(zendeskChatName, New)
}

// New returns the Zendesk Chat connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Zendesk Chat connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return zendeskChatName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            zendeskChatName,
		DisplayName:     "Zendesk Chat",
		IntegrationType: "api",
		Description:     "Reads Zendesk Chat agents, chats, departments, shortcuts, and triggers through the Zendesk Chat REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Zendesk Chat.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := zendeskChatBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(zendeskChatSecret(cfg)) == "" {
		return errors.New("zendesk-chat connector requires secret credentials.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the agents list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "agents", nil, nil, nil); err != nil {
		return fmt.Errorf("check zendesk-chat: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: zendeskChatStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (full sync), which the start_date config can raise at read
// time for the chats incremental-export stream.
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
		stream = "agents"
	}
	endpoint, ok := zendeskChatStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("zendesk-chat stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	switch endpoint.pagination {
	case paginateChats:
		return c.harvestChats(ctx, r, endpoint, req, emit)
	default:
		return c.harvestArray(ctx, r, endpoint, emit)
	}
}

// harvestArray reads a single-page top-level-array endpoint (agents,
// departments, shortcuts, triggers).
func (c Connector) harvestArray(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read zendesk-chat %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode zendesk-chat %s: %w", endpoint.resource, err)
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

// harvestChats drives the incremental-export pagination of the chats endpoint.
// The first request carries start_time (derived from the cursor or start_date);
// each response advertises the next page via an absolute next_url that is chased
// until it is empty. There is no connsdk paginator for this exact body shape, so
// the loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvestChats(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	startTime, err := chatsStartTime(req)
	if err != nil {
		return err
	}
	query := url.Values{}
	if startTime != "" {
		query.Set("start_time", startTime)
	}

	path := endpoint.resource
	for page := 0; page < zendeskChatMaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read zendesk-chat %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode zendesk-chat %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_url")
		if err != nil {
			return fmt.Errorf("decode zendesk-chat %s next_url: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		// Zendesk repeats the final next_url when the export is fully drained; a
		// short or empty page terminates the loop to avoid an infinite spin.
		if next == "" || len(records) == 0 {
			return nil
		}
		// next_url is an absolute URL; the Requester treats http(s)-prefixed paths
		// as absolute. Query params are already embedded in next_url, so clear the
		// outgoing query to avoid duplicating start_time.
		path = next
		query = url.Values{}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type":          "chat",
			"timestamp":     fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"display_name":  fmt.Sprintf("Fixture Agent %d", i),
			"first_name":    "Fixture",
			"last_name":     fmt.Sprintf("Agent %d", i),
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"role_id":       int64(i),
			"enabled":       true,
			"create_date":   "2026-01-01T00:00:00Z",
			"last_login":    "2026-01-02T00:00:00Z",
			"name":          fmt.Sprintf("Fixture %d", i),
			"description":   "fixture record",
			"department_id": int64(i),
			"rating":        "good",
			"comment":       "fixture",
			"duration":      int64(60 * i),
			"options":       "fixture options",
			"message":       "fixture message",
			"connector":     zendeskChatName,
			"fixture":       true,
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
	base, err := zendeskChatBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := zendeskChatSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("zendesk-chat connector requires secret credentials.access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: zendeskChatUserAgent,
	}, nil
}

// chatsStartTime returns the unix-seconds start_time for the chats incremental
// export, derived from the incremental cursor (if any) or else the start_date
// config. An empty result means no lower bound (full export).
func chatsStartTime(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		// A numeric cursor is already a unix timestamp; pass it through. A
		// timestamp cursor is converted below.
		if _, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			return cursor, nil
		}
		if t, err := time.Parse(time.RFC3339, cursor); err == nil {
			return strconv.FormatInt(t.Unix(), 10), nil
		}
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	t, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		return "", fmt.Errorf("zendesk-chat config start_date must be RFC3339: %w", err)
	}
	return strconv.FormatInt(t.Unix(), 10), nil
}

// zendeskChatSecret resolves the OAuth access token from the secrets map. Both
// the nested catalog key (credentials.access_token) and a flat access_token are
// accepted for convenience.
func zendeskChatSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets[zendeskChatAccessTokenSecret]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Secrets["access_token"])
}

// zendeskChatBaseURL resolves and validates the base URL. With a subdomain
// configured the default is https://{subdomain}.zendesk.com/api/v2/chat. Any
// explicit override must be an absolute http(s) URL with a host to bound SSRF
// risk.
func zendeskChatBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		subdomain := strings.TrimSpace(cfg.Config["subdomain"])
		if subdomain == "" {
			return "", errors.New("zendesk-chat connector requires config subdomain or base_url")
		}
		if strings.ContainsAny(subdomain, "/:") || strings.Contains(subdomain, " ") {
			return "", fmt.Errorf("zendesk-chat config subdomain must be a bare subdomain, got %q", subdomain)
		}
		return fmt.Sprintf("https://%s.zendesk.com/api/v2/chat", subdomain), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("zendesk-chat config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("zendesk-chat config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("zendesk-chat config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Zendesk Chat exposes no obvious safe reverse-ETL writes
// for this connector, so it is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
