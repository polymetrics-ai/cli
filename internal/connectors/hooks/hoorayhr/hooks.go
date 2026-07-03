// Package hoorayhr is the Tier-2 escape hatch for the hoorayhr bundle: a
// single AuthHook porting legacy's custom session-token exchange
// (internal/connectors/hoorayhr/hoorayhr.go's sessionTokenAuth, read-only
// reference).
//
// HoorayHR authenticates by POSTing {email, password, strategy:"local"} JSON
// to /authentication, reading accessToken back from the response, and
// injecting that token RAW (no "Bearer " prefix) into the Authorization
// header of every subsequent request. None of the engine's declarative auth
// modes (bearer/basic/api_key_header/api_key_query/oauth2_client_credentials)
// can express "POST a fixed-shape JSON body to a fixed path, then inject an
// unprefixed response field into a header on every later request" — this is
// a genuine token-exchange auth shape (conventions.md Section 1's AuthHook
// row), the same class of escape hatch as hooks/keka and hooks/github's
// token-exchange hooks. This single hook interface, well under the ~300-line
// soft target, is the correct Tier-2 shape rather than stretching the
// declarative dialect or escalating to Tier 3 (the protocol is still plain
// HTTP/REST for every data read; only the login exchange itself is
// non-declarative).
package hoorayhr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const loginPath = "authentication"

func init() {
	engine.RegisterHooks("hoorayhr", func() engine.Hooks { return New() })
}

// Hooks implements engine.AuthHook for the hoorayhr bundle. Its only state is
// a test-injection HTTP client override (mirrors hooks/keka's Hooks shape);
// every method is otherwise a pure function of its arguments.
type Hooks struct {
	// Client overrides the HTTP client used for the login exchange; nil uses
	// http.DefaultClient (mirrors legacy's Connector.Client fallback).
	Client *http.Client
}

// New returns a fresh hoorayhr Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "hoorayhr" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

// Authenticator builds a session-token authenticator that logs in once
// (POST /authentication) and caches the raw accessToken for every subsequent
// request (matches legacy's sessionTokenAuth exactly: same login body shape,
// same uncached-until-first-use caching, same raw non-Bearer-prefixed header
// injection). ctx is honored so a caller cancellation aborts an in-flight
// login (F8-equivalent: the real caller context is threaded through Apply,
// never context.Background()).
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	username := strings.TrimSpace(cfg.Config["hoorayhrusername"])
	if username == "" {
		return nil, errors.New("hoorayhr connector requires config hoorayhrusername")
	}
	password := ""
	if cfg.Secrets != nil {
		password = cfg.Secrets["hoorayhrpassword"]
	}
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("hoorayhr connector requires secret hoorayhrpassword")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.Config["base_url"]), "/")
	if baseURL == "" {
		baseURL = "https://api.hooray.nl"
	}
	return &sessionTokenAuth{
		login: &connsdk.Requester{
			Client:    h.Client,
			BaseURL:   baseURL,
			UserAgent: "polymetrics-go-cli",
		},
		username: username,
		password: password,
	}, nil
}

// sessionTokenAuth implements connsdk.Authenticator for HoorayHR's
// session-token flow: it POSTs credentials to /authentication once, caches
// accessToken, and injects it raw into the Authorization header. The
// password and token are never logged.
type sessionTokenAuth struct {
	login    *connsdk.Requester
	username string
	password string

	mu     sync.Mutex
	cached string
}

// Apply ensures a session token has been fetched and sets it on the request.
func (a *sessionTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	return nil
}

// token returns the cached access token, fetching one via /authentication on
// first use.
func (a *sessionTokenAuth) token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cached != "" {
		return a.cached, nil
	}
	body := map[string]any{
		"email":    a.username,
		"password": a.password,
		"strategy": "local",
	}
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	if err := a.login.DoJSON(ctx, http.MethodPost, loginPath, nil, body, &out); err != nil {
		return "", fmt.Errorf("hoorayhr authentication: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("hoorayhr authentication response missing accessToken")
	}
	a.cached = out.AccessToken
	return a.cached, nil
}
