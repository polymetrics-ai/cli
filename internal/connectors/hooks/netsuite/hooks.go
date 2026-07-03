// Package netsuite implements the netsuite bundle's AuthHook: an OAuth 1.0a
// Token-Based Authentication (TBA) HMAC-SHA256 request-signing
// connsdk.Authenticator, porting legacy internal/connectors/netsuite's
// oauth1/oauthSignature field-for-field (conventions.md §1's Tier-2 table:
// "signature auth (SigV4, HMAC)" — the exact AUTH_COMPLEX trigger this hook
// exists for, well under the 300-line soft target). Only one hook interface
// is implemented (AuthHook).
//
// The engine's declarative auth dialect has no mode that computes a
// signature over the outgoing request's own method/URL/query/timestamp/
// nonce — OAuth 1.0a TBA needs exactly that, so this is a genuine Tier-2
// signature-auth case, not a Tier-3 native-protocol one (NetSuite's REST
// Record API is ordinary JSON-over-HTTPS; see docs.md's tier-justification
// note). Secret values (consumer_secret, token_secret) flow ONLY into the
// HMAC digest and are never sent on the wire themselves, never logged, and
// never appear in an error string.
package netsuite

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("netsuite", func() engine.Hooks { return New() })
}

// Hooks is the netsuite hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now (mirrors legacy's
	// time.Now() call in oauth1.Apply).
	Now func() time.Time
}

// New returns a fresh netsuite Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "netsuite" }

// Authenticator resolves the OAuth 1.0a connsdk.Authenticator for spec
// (mode "custom", hook "netsuite"). Unlike a templated AuthSpec field
// (gmail's token_url/client_id shape), the five NetSuite credentials are
// read directly off cfg here by their fixed legacy key names, mirroring
// nexus-datasets' hook: none of AuthSpec's generic fields map onto
// "OAuth realm" / "consumer key" / "token secret" naturally.
func (h *Hooks) Authenticator(_ context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	realm := configOrSecret(cfg, "realm")
	if realm == "" {
		return nil, errors.New("netsuite oauth: realm is required")
	}
	consumerKey := configOrSecret(cfg, "consumer_key")
	if consumerKey == "" {
		return nil, errors.New("netsuite oauth: consumer_key is required")
	}
	consumerSecret := configOrSecret(cfg, "consumer_secret")
	if consumerSecret == "" {
		return nil, errors.New("netsuite oauth: consumer_secret is required")
	}
	tokenKey := configOrSecret(cfg, "token_key")
	if tokenKey == "" {
		return nil, errors.New("netsuite oauth: token_key is required")
	}
	tokenSecret := configOrSecret(cfg, "token_secret")
	if tokenSecret == "" {
		return nil, errors.New("netsuite oauth: token_secret is required")
	}

	return &oauth1Auth{
		realm:          realm,
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		tokenKey:       tokenKey,
		tokenSecret:    tokenSecret,
		now:            h.Now,
	}, nil
}

// configOrSecret reads key from cfg.Config first, falling back to
// cfg.Secrets, mirroring legacy netsuite.go's configOrSecret helper exactly
// (legacy accepts these five values from either map).
func configOrSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			return v
		}
	}
	if cfg.Secrets != nil {
		return strings.TrimSpace(cfg.Secrets[key])
	}
	return ""
}

// oauth1Auth implements connsdk.Authenticator for NetSuite's OAuth 1.0a
// Token-Based Authentication (TBA) HMAC-SHA256 signing scheme, mirroring
// legacy netsuite.go's oauth1/oauthSignature field-for-field: build the
// OAuth parameter set (consumer key, token, timestamp, nonce, version),
// merge in the request's own query parameters, compute the HMAC-SHA256
// signature over the canonical base string keyed by the percent-encoded
// consumer+token secrets, and set a single Authorization header carrying
// every OAuth parameter plus the signature. The secret values never appear
// on the wire themselves (only the HMAC digest does) and are never logged.
type oauth1Auth struct {
	realm, consumerKey, consumerSecret, tokenKey, tokenSecret string

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time
}

func (a *oauth1Auth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

// Apply signs req and sets the OAuth 1.0a Authorization header. Signing
// happens fresh on every call (legacy computes a new timestamp and nonce
// per request, never caching the signature — matching netsuite.go's
// oauth1.Apply exactly).
func (a *oauth1Auth) Apply(_ context.Context, req *http.Request) error {
	params := map[string]string{
		"oauth_consumer_key":     a.consumerKey,
		"oauth_token":            a.tokenKey,
		"oauth_signature_method": "HMAC-SHA256",
		"oauth_timestamp":        strconv.FormatInt(a.timeNow().Unix(), 10),
		"oauth_nonce":            strconv.FormatInt(a.timeNow().UnixNano(), 36),
		"oauth_version":          "1.0",
	}
	for key, values := range req.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	params["oauth_signature"] = oauthSignature(req.Method, req.URL, params, a.consumerSecret, a.tokenSecret)

	headerParams := []string{fmt.Sprintf("realm=%q", a.realm)}
	for _, key := range []string{"oauth_consumer_key", "oauth_token", "oauth_signature_method", "oauth_timestamp", "oauth_nonce", "oauth_version", "oauth_signature"} {
		headerParams = append(headerParams, fmt.Sprintf("%s=%q", key, percentEncode(params[key])))
	}
	req.Header.Set("Authorization", "OAuth "+strings.Join(headerParams, ","))
	return nil
}

// oauthSignature computes the OAuth 1.0a HMAC-SHA256 signature over the
// canonical base string (METHOD&percent-encoded-base-url&percent-encoded-
// sorted-param-string), keyed by percent-encode(consumerSecret)&
// percent-encode(tokenSecret) — ported verbatim from legacy netsuite.go.
func oauthSignature(method string, u *url.URL, params map[string]string, consumerSecret, tokenSecret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key != "oauth_signature" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, percentEncode(key)+"="+percentEncode(params[key]))
	}
	baseURL := *u
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	base := strings.ToUpper(method) + "&" + percentEncode(baseURL.String()) + "&" + percentEncode(strings.Join(parts, "&"))
	mac := hmac.New(sha256.New, []byte(percentEncode(consumerSecret)+"&"+percentEncode(tokenSecret)))
	_, _ = mac.Write([]byte(base))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// percentEncode applies RFC 3986 percent-encoding as OAuth 1.0a requires it
// (url.QueryEscape encodes a space as "+"; OAuth needs "%20") — ported
// verbatim from legacy netsuite.go's percent helper.
func percentEncode(s string) string { return strings.ReplaceAll(url.QueryEscape(s), "+", "%20") }

// BaseURLFromRealm derives the NetSuite REST Record API base URL from realm
// exactly like legacy netsuite.go's baseURL function's realm-derivation
// branch: lowercased, underscores replaced with hyphens, appended to the
// SuiteTalk REST host. It is exported so streams.json's spec-default
// documentation and any future config-resolution helper can reference the
// identical derivation without duplicating the string literal.
func BaseURLFromRealm(realm string) string {
	host := strings.ToLower(strings.ReplaceAll(realm, "_", "-"))
	return "https://" + host + ".suitetalk.api.netsuite.com/services/rest/record/v1"
}
