package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/credential"
	"polymetrics.ai/internal/safety"
)

// authVars builds the interpolation environment for AuthSpec fields: config
// and secret values only (no record/cursor context exists during auth
// selection).
func authVars(cfg connectors.RuntimeConfig) Vars {
	return Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// rateLimitConfigForSelectedAuth gives the resolver a hook-declared, non-secret
// profile for the first auth spec that will be selected. This lets a custom
// auth exchange be admitted before it sends its own request without changing
// the configuration passed to the authenticator.
func rateLimitConfigForSelectedAuth(cfg connectors.RuntimeConfig, specs []AuthSpec, h Hooks) connectors.RuntimeConfig {
	profileHook, ok := h.(RateLimitAuthProfileHook)
	if !ok || len(specs) == 0 {
		return cfg
	}

	vars := authVars(cfg)
	for _, spec := range specs {
		matched, err := authSpecMatches(spec, vars)
		if err != nil {
			return cfg
		}
		if !matched {
			continue
		}
		profile, ok := profileHook.RateLimitAuthProfile(cfg, spec)
		if !ok {
			return cfg
		}
		profile = strings.TrimSpace(profile)
		if profile == "" || cfg.Config["auth_type"] == profile {
			return cfg
		}
		normalized := make(map[string]string, len(cfg.Config)+1)
		for key, value := range cfg.Config {
			normalized[key] = value
		}
		normalized["auth_type"] = profile
		cfg.Config = normalized
		return cfg
	}
	return cfg
}

// selectAuth evaluates specs in declared order and returns the
// connsdk.Authenticator for the first spec whose "when" condition matches
// (a spec with no "when" always matches). mode "custom" resolves an
// AuthHook via h (the connector's registered Hooks, or nil when none). ctx
// is the caller's context, threaded through to AuthHook.Authenticator (F8,
// REVIEW.md: a github_app-style JWT->installation-token exchange is a
// network call and must honor the caller's cancellation/deadline, not run
// under context.Background()). Secret values flow only into the constructed
// Authenticator, never into error messages (mirrors stripe/stripe.go:279 —
// secrets never read from Config, only Secrets).
func selectAuth(ctx context.Context, cfg connectors.RuntimeConfig, specs []AuthSpec, h Hooks) (connsdk.Authenticator, error) {
	return selectAuthWithDeclaredRoute(ctx, cfg, specs, h, nil)
}

// selectAuthWithDeclaredRoute is selectAuth's runtime-aware counterpart. Only
// a hook opting into DeclaredRouteAuthHook receives the narrow requester;
// existing hooks retain their AuthHook contract unchanged.
func selectAuthWithDeclaredRoute(ctx context.Context, cfg connectors.RuntimeConfig, specs []AuthSpec, h Hooks, requester DeclaredRouteRequester) (connsdk.Authenticator, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("select auth: no auth specs declared")
	}

	vars := authVars(cfg)

	for _, spec := range specs {
		matched, err := authSpecMatches(spec, vars)
		if err != nil {
			return nil, fmt.Errorf("select auth: mode %q: %w", spec.Mode, err)
		}
		if !matched {
			continue
		}
		return buildAuthenticatorWithDeclaredRoute(ctx, cfg, spec, vars, h, requester)
	}

	return nil, fmt.Errorf("select auth: no auth spec matched for auth_type %q", cfg.Config["auth_type"])
}

// authSpecMatches reports whether spec's "when" condition matches vars. A
// spec with an empty "when" always matches.
func authSpecMatches(spec AuthSpec, vars Vars) (bool, error) {
	if strings.TrimSpace(spec.When) == "" {
		return true, nil
	}
	return EvalWhen(spec.When, vars)
}

func buildAuthenticatorWithDeclaredRoute(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec, vars Vars, h Hooks, requester DeclaredRouteRequester) (connsdk.Authenticator, error) {
	switch spec.Mode {
	case "none":
		return connsdk.AuthFunc(func(_ context.Context, _ *http.Request) error { return nil }), nil

	case "bearer":
		token, err := Interpolate(spec.Token, vars)
		if err != nil {
			return nil, fmt.Errorf("bearer: %w", err)
		}
		if err := credential.RequireAuthenticationValue("bearer token", token); err != nil {
			return nil, fmt.Errorf("bearer: %w", err)
		}
		return connsdk.Bearer(token), nil

	case "basic":
		username, err := Interpolate(spec.Username, vars)
		if err != nil {
			return nil, fmt.Errorf("basic: username: %w", err)
		}
		password, err := Interpolate(spec.Password, vars)
		if err != nil {
			return nil, fmt.Errorf("basic: password: %w", err)
		}
		if err := credential.RequireAuthenticationValue("basic username", username); err != nil {
			return nil, fmt.Errorf("basic: %w", err)
		}
		if err := credential.RequireAuthenticationValue("basic password", password); err != nil {
			return nil, fmt.Errorf("basic: %w", err)
		}
		return connsdk.Basic(username, password), nil

	case "api_key_header":
		value, err := Interpolate(spec.Value, vars)
		if err != nil {
			return nil, fmt.Errorf("api_key_header: %w", err)
		}
		if err := credential.RequireAuthenticationValue("API key", value); err != nil {
			return nil, fmt.Errorf("api_key_header: %w", err)
		}
		return connsdk.APIKeyHeader(spec.Header, value, spec.Prefix), nil

	case "api_key_query":
		value, err := Interpolate(spec.Value, vars)
		if err != nil {
			return nil, fmt.Errorf("api_key_query: %w", err)
		}
		if err := credential.RequireAuthenticationValue("API key", value); err != nil {
			return nil, fmt.Errorf("api_key_query: %w", err)
		}
		return connsdk.APIKeyQuery(spec.Param, value), nil

	case "oauth2_client_credentials":
		return buildOAuth2ClientCredentials(spec, vars)

	case "oauth2_refresh_token":
		return buildOAuth2RefreshToken(cfg, spec, vars)

	case "custom":
		return buildCustomAuthWithDeclaredRoute(ctx, cfg, spec, h, requester)

	default:
		return nil, fmt.Errorf("unknown auth mode %q", spec.Mode)
	}
}

func buildOAuth2ClientCredentials(spec AuthSpec, vars Vars) (connsdk.Authenticator, error) {
	tokenURL, err := Interpolate(spec.TokenURL, vars)
	if err != nil {
		return nil, fmt.Errorf("oauth2_client_credentials: token_url: %w", err)
	}
	clientID, err := Interpolate(spec.ClientID, vars)
	if err != nil {
		return nil, fmt.Errorf("oauth2_client_credentials: client_id: %w", err)
	}
	clientSecret, err := Interpolate(spec.ClientSecret, vars)
	if err != nil {
		return nil, fmt.Errorf("oauth2_client_credentials: client_secret: %w", err)
	}
	if err := credential.RequireAuthenticationValue("OAuth2 client ID", clientID); err != nil {
		return nil, fmt.Errorf("oauth2_client_credentials: %w", err)
	}
	if err := credential.RequireAuthenticationValue("OAuth2 client secret", clientSecret); err != nil {
		return nil, fmt.Errorf("oauth2_client_credentials: %w", err)
	}

	var scopes []string
	if spec.Scopes != "" {
		resolved, err := Interpolate(spec.Scopes, vars)
		if err != nil {
			return nil, fmt.Errorf("oauth2_client_credentials: scopes: %w", err)
		}
		scopes = strings.Fields(resolved)
	}

	extraParams, err := resolveExtraParams("oauth2_client_credentials", spec.ExtraParams, vars)
	if err != nil {
		return nil, err
	}

	return &connsdk.OAuth2ClientCredentials{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		ExtraParams:  extraParams,
	}, nil
}

// resolveExtraParams resolves every AuthSpec.ExtraParams entry (S4 engine
// mini-wave item 4) against vars, hard-erroring on an unresolved
// config/secrets key exactly like every other AuthSpec field — an
// extra_params value is never silently dropped on absence, unlike
// stream.Query's opt-in omit_when_absent/default dialect (that tolerance is
// deliberately NOT extended here: a mis-configured audience/subject param
// should fail loudly, matching ClientID/ClientSecret's own behavior, not
// silently omit a param a real OAuth2 provider may require). A nil/empty
// map returns a nil url.Values (connsdk.OAuth2ClientCredentials.ExtraParams
// ranges over a nil map with zero iterations, so this is a true no-op).
func resolveExtraParams(mode string, params map[string]string, vars Vars) (url.Values, error) {
	if len(params) == 0 {
		return nil, nil
	}
	out := url.Values{}
	for k, tmpl := range params {
		val, err := Interpolate(tmpl, vars)
		if err != nil {
			return nil, fmt.Errorf("%s: extra_params %q: %w", mode, k, err)
		}
		out.Set(k, val)
	}
	return out, nil
}

// buildOAuth2RefreshToken constructs the oauth2_refresh_token authenticator:
// the OAuth2 refresh-token grant (RFC 6749 §6), which renews a USER-CONTEXT
// access token.
//
// oauth2_client_credentials cannot stand in for it. That grant obtains an
// app-only token — it authenticates the application, not a user — so it cannot
// reach any endpoint acting on behalf of an end user. A connector bundle has
// nowhere to put a token exchange, an expiry clock, a rotation callback or a
// 401 retry, which is why this is shared-runtime behaviour rather than
// something a connector lane can add for itself.
//
// Field resolution is identical to buildOAuth2ClientCredentials: every
// templated field goes through Interpolate and an unresolved config/secrets key
// is a hard error, never a silently dropped credential.
//
// cfg is threaded in (unlike the client-credentials builder) only for
// cfg.SecretStore, which is how a provider-rotated refresh token reaches the
// caller's encrypted credential store. Secret VALUES still come from vars, and
// none of them reaches an error message.
func buildOAuth2RefreshToken(cfg connectors.RuntimeConfig, spec AuthSpec, vars Vars) (connsdk.Authenticator, error) {
	const mode = "oauth2_refresh_token"

	tokenURL, err := Interpolate(spec.TokenURL, vars)
	if err != nil {
		return nil, fmt.Errorf("%s: token_url: %w", mode, err)
	}
	clientID, err := Interpolate(spec.ClientID, vars)
	if err != nil {
		return nil, fmt.Errorf("%s: client_id: %w", mode, err)
	}
	clientSecret, err := Interpolate(spec.ClientSecret, vars)
	if err != nil {
		return nil, fmt.Errorf("%s: client_secret: %w", mode, err)
	}
	refreshToken, err := Interpolate(spec.RefreshToken, vars)
	if err != nil {
		return nil, fmt.Errorf("%s: refresh_token: %w", mode, err)
	}
	if err := credential.RequireAuthenticationValue("OAuth2 refresh token", refreshToken); err != nil {
		return nil, fmt.Errorf("%s: %w", mode, err)
	}

	var scopes []string
	if spec.Scopes != "" {
		resolved, err := Interpolate(spec.Scopes, vars)
		if err != nil {
			return nil, fmt.Errorf("%s: scopes: %w", mode, err)
		}
		scopes = strings.Fields(resolved)
	}

	extraParams, err := resolveExtraParams(mode, spec.ExtraParams, vars)
	if err != nil {
		return nil, err
	}

	auth := &connsdk.OAuth2RefreshToken{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
		Scopes:       scopes,
		ExtraParams:  extraParams,
	}

	// Rotation persistence is opt-in and requires BOTH a declared key and a
	// store. A bundle that declares no key never writes anything — the engine
	// does not guess which secret to overwrite (see AuthSpec.RefreshTokenStoreKey).
	// A caller with no store (conformance harnesses, tests) keeps rotation in
	// memory for the process lifetime; it is never downgraded to a plaintext
	// write.
	storeKey := strings.TrimSpace(spec.RefreshTokenStoreKey)
	if storeKey != "" {
		if err := safety.ValidateIdentifier(storeKey, mode+": refresh_token_store_key"); err != nil {
			return nil, err
		}
		if cfg.SecretStore != nil {
			store := cfg.SecretStore
			auth.OnRefreshTokenRotated = func(ctx context.Context, rotated string) error {
				return store.PutSecret(ctx, storeKey, rotated)
			}
		}
	}

	return auth, nil
}

func buildCustomAuthWithDeclaredRoute(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec, h Hooks, requester DeclaredRouteRequester) (connsdk.Authenticator, error) {
	if h == nil {
		return nil, fmt.Errorf("custom auth: hook %q not registered (no hooks provided)", spec.Hook)
	}
	authHook, ok := h.(AuthHook)
	if !ok {
		return nil, fmt.Errorf("custom auth: hook %q not registered (hooks %q does not implement AuthHook)", spec.Hook, h.ConnectorName())
	}
	if declaredRouteHook, ok := h.(DeclaredRouteAuthHook); ok {
		return declaredRouteHook.AuthenticatorWithDeclaredRoute(ctx, cfg, spec, requester)
	}
	return authHook.Authenticator(ctx, cfg, spec)
}
