// Package outlook implements the outlook bundle's Tier-2 hook set
// (docs/migration/quarantine.json's "outlook" AUTH_COMPLEX entry: "OAuth2
// refresh_token grant pre-request (hook needed, gmail pattern)"). Exactly 2
// hook interfaces are implemented, within conventions.md §1's Tier-2 cap:
//
//   - AuthHook: an OAuth 2.0 refresh-token-grant connsdk.Authenticator,
//     porting legacy internal/connectors/outlook/outlook.go's
//     refreshTokenAuth almost verbatim (the engine's declarative
//     oauth2_client_credentials mode only performs client-credentials,
//     never grant_type=refresh_token). Same shape as hooks/gmail and
//     hooks/strava's AuthHooks (gmail's https-only token_url guard is
//     followed here per this task's mandate).
//   - StreamHook: Microsoft Graph's @odata.nextLink pagination cursor lives
//     at a response-body key containing a literal "." — the declarative
//     next_url type's dotted-path parser (connsdk.StringAt) cannot address
//     it (confirmed: it silently resolves to "" with no error), the same
//     gap already recorded for microsoft-entra-id/microsoft-lists/
//     microsoft-teams. Ports legacy's harvest/nextLink loop verbatim.
//
// Secret values (client_secret, the refresh token, cached access tokens)
// flow ONLY into the outgoing token-request form or the Authorization
// header; never logged, never in an error string (THREAT-MODEL.md Delta 2).
package outlook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("outlook", func() engine.Hooks { return New() })
}

// Hooks is the outlook hook set: engine.AuthHook and engine.StreamHook.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil
	// uses a default client with a 30s timeout (mirrors legacy's inline
	// default).
	Client *http.Client
}

// New returns a fresh outlook Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "outlook" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "outlook"). spec.Token is interpreted as the
// refresh token (gmail/strava's identical convention).
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	tokenURL, err := interpolateRequired(spec.TokenURL, "token_url", cfg)
	if err != nil {
		return nil, err
	}
	if err := validateHTTPSURL(tokenURL, "token_url"); err != nil {
		return nil, err
	}
	clientID, err := interpolateRequired(spec.ClientID, "client_id", cfg)
	if err != nil {
		return nil, err
	}
	clientSecret, err := interpolateRequired(spec.ClientSecret, "client_secret", cfg)
	if err != nil {
		return nil, err
	}
	refreshToken, err := interpolateRequired(spec.Token, "refresh_token", cfg)
	if err != nil {
		return nil, err
	}
	// scope is genuinely optional (legacy only sets it on the form when
	// non-empty); interpolateOptional tolerates the absent key instead of
	// hard-erroring like engine.Interpolate normally would.
	return &refreshTokenAuth{
		tokenURL: tokenURL, clientID: clientID, clientSecret: clientSecret,
		refreshToken: refreshToken, scope: interpolateOptional(spec.Scopes, cfg),
		client: h.Client, now: h.Now,
	}, nil
}

// interpolateRequired resolves tmpl via engine.Interpolate, wraps any error
// naming field, and also rejects an empty resolved value (legacy parity).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("outlook oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("outlook oauth: %s is required", field)
	}
	return val, nil
}

// interpolateOptional resolves tmpl best-effort: ANY engine.Interpolate
// error (absent key, CRLF injection, unknown filter) resolves to "" rather
// than propagating, matching gmail's interpolateOptional exactly. Its only
// call site (spec.Scopes) is an optional-when-empty POST-form value.
func interpolateOptional(tmpl string, cfg connectors.RuntimeConfig) string {
	if strings.TrimSpace(tmpl) == "" {
		return ""
	}
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return ""
	}
	return val
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPSURL fails closed on anything but a well-formed https:// URL
// with a host (THREAT-MODEL.md Delta 2: an attacker-controlled token_url
// override could otherwise exfiltrate client_secret/the refresh token).
// Stricter than legacy's validateBaseURL (which also accepted plain http);
// documented as a parity deviation in docs.md's Known limits — identical
// rule to hooks/gmail's validateHTTPSURL per this task's mandate.
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("outlook oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("outlook oauth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("outlook oauth: %s must include a host", field)
	}
	return nil
}

// refreshTokenAuth implements connsdk.Authenticator for the Microsoft Graph
// OAuth 2.0 refresh-token grant, mirroring legacy outlook.go's
// refreshTokenAuth: exchange client_id/client_secret/refresh_token for a
// short-lived bearer access token at tokenURL, cache it until 60s before
// its declared expiry (3600s TTL default when expires_in is
// absent/unparsable), then set Authorization: Bearer <token> per request.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	scope        string
	client       *http.Client

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *refreshTokenAuth) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply sets Authorization: Bearer <token> after ensuring a fresh token.
func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races (legacy parity).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("outlook oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("outlook oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("outlook oauth: client_id is required")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)
	if strings.TrimSpace(a.scope) != "" {
		form.Set("scope", a.scope)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("outlook oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("outlook oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("outlook oauth: token endpoint returned %s", resp.Status)
	}
	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("outlook oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("outlook oauth: token response missing access_token")
	}
	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}

// graphStreamPaths mirrors legacy's streamEndpoints routing table: the
// Graph collection path (relative to base_url) each stream reads from.
var graphStreamPaths = map[string]string{
	"messages":     "/me/messages",
	"mail_folders": "/me/mailFolders",
	"events":       "/me/events",
}

// ReadStream drives Microsoft Graph's @odata.nextLink pagination for every
// stream this bundle declares. handled is always true for a recognized
// stream name; an unrecognized name returns handled=false (declarative
// fallback), which should not happen for a correctly authored bundle.
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}
	name := stream.Name
	if name == "" {
		name = "messages"
	}
	path, ok := graphStreamPaths[name]
	if !ok {
		return false, nil
	}
	schema := rt.Bundle.Schemas[name]
	if schema == nil {
		return false, nil
	}
	query := url.Values{"$top": []string{strconv.Itoa(pageSizeFor(req.Config))}}
	return true, harvest(ctx, rt.Requester, path, query, maxPagesFor(req.Config), schema.Properties(), emit)
}

// harvest follows Microsoft Graph's @odata.nextLink pagination. Graph
// collections return {value:[...], "@odata.nextLink":"<absolute url>"}; the
// next page is the nextLink URL verbatim — connsdk.Requester treats an
// absolute http(s) path as-is. Port of legacy outlook.go's harvest.
func harvest(ctx context.Context, r *connsdk.Requester, firstPath string, firstQuery url.Values, maxPages int, props []string, emit func(connectors.Record) error) error {
	path := firstPath
	query := firstQuery
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("outlook: read %s: %w", firstPath, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return fmt.Errorf("outlook: decode %s page: %w", firstPath, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(connectors.Record(projectBySchema(item, props))); err != nil {
				return err
			}
		}
		// "@odata.nextLink" contains a literal dot, unaddressable via
		// connsdk.StringAt's dotted-path traversal; decode it directly.
		next, err := nextLink(resp.Body)
		if err != nil {
			return fmt.Errorf("outlook: decode %s nextLink: %w", firstPath, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// nextLink is an absolute URL carrying its own query ($skiptoken
		// etc.); pass it as the path with no extra query, used verbatim.
		path = next
		query = nil
	}
	return nil
}

// projectBySchema keeps only the schema-declared properties from raw,
// renamed from their raw Microsoft Graph (camelCase) field name via
// graphFieldNames, matching conventions.md §2's schema-as-projection rule
// and legacy's per-stream message/folder/eventRecord functions
// field-for-field. Graph itself never emits snake_case keys, so this
// renames explicitly rather than relying on schema projection's
// exact-key-match default.
func projectBySchema(raw map[string]any, props []string) map[string]any {
	out := make(map[string]any, len(props))
	for _, name := range props {
		if graphKey, ok := graphFieldNames[name]; ok {
			if v, ok := raw[graphKey]; ok {
				out[name] = v
			}
		}
	}
	return out
}

// graphFieldNames is the union of every stream's schema-field -> raw-Graph-
// field mapping. Names are unique enough across streams that a single flat
// map is unambiguous, since projectBySchema only looks up the current
// stream's own declared properties.
var graphFieldNames = map[string]string{
	"id": "id", "subject": "subject", "web_link": "webLink",
	"received_date_time": "receivedDateTime", "last_modified_date_time": "lastModifiedDateTime",
	"created_date_time": "createdDateTime", "display_name": "displayName",
	"total_item_count": "totalItemCount", "unread_item_count": "unreadItemCount",
}

// nextLink reads the literal "@odata.nextLink" key from a Graph response
// body: an absolute URL (carrying its own $skiptoken) or empty on the last
// page. Port of legacy outlook.go's nextLink.
func nextLink(body []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var envelope struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if err := dec.Decode(&envelope); err != nil {
		return "", fmt.Errorf("decode graph envelope: %w", err)
	}
	return envelope.NextLink, nil
}

// pageSizeFor mirrors legacy's pageSize: bounded [1, 999]; a malformed or
// out-of-range value falls back to the default (100).
func pageSizeFor(cfg connectors.RuntimeConfig) int {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	n, err := strconv.Atoi(raw)
	if raw == "" || err != nil || n < 1 || n > 999 {
		return 100
	}
	return n
}

// maxPagesFor mirrors legacy's maxPages: permissive parse, never errors —
// empty/"all"/"unlimited"/malformed/negative all mean unbounded (0).
func maxPagesFor(cfg connectors.RuntimeConfig) int {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
