package engine

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// authVars builds the interpolation environment for AuthSpec fields: config
// and secret values only (no record/cursor context exists during auth
// selection).
func authVars(cfg connectors.RuntimeConfig) Vars {
	return Vars{Config: cfg.Config, Secrets: cfg.Secrets}
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
		return buildAuthenticator(ctx, cfg, spec, vars, h)
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

// buildAuthenticator constructs the connsdk.Authenticator for a matched
// spec, interpolating every templated field first.
func buildAuthenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec, vars Vars, h Hooks) (connsdk.Authenticator, error) {
	switch spec.Mode {
	case "none":
		return connsdk.AuthFunc(func(_ context.Context, _ *http.Request) error { return nil }), nil

	case "bearer":
		token, err := Interpolate(spec.Token, vars)
		if err != nil {
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
		return connsdk.Basic(username, password), nil

	case "api_key_header":
		value, err := Interpolate(spec.Value, vars)
		if err != nil {
			return nil, fmt.Errorf("api_key_header: %w", err)
		}
		return connsdk.APIKeyHeader(spec.Header, value, spec.Prefix), nil

	case "api_key_query":
		value, err := Interpolate(spec.Value, vars)
		if err != nil {
			return nil, fmt.Errorf("api_key_query: %w", err)
		}
		return connsdk.APIKeyQuery(spec.Param, value), nil

	case "oauth2_client_credentials":
		return buildOAuth2ClientCredentials(spec, vars)

	case "custom":
		return buildCustomAuth(ctx, cfg, spec, h)

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

	var scopes []string
	if spec.Scopes != "" {
		resolved, err := Interpolate(spec.Scopes, vars)
		if err != nil {
			return nil, fmt.Errorf("oauth2_client_credentials: scopes: %w", err)
		}
		scopes = strings.Fields(resolved)
	}

	return &connsdk.OAuth2ClientCredentials{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
	}, nil
}

// buildCustomAuth resolves an AuthHook for spec.Hook via h. A nil Hooks, or
// a Hooks that does not implement AuthHook, is a typed error naming the
// missing hook rather than a workaround (per PLAN.md: a needed hook is a
// blocker, never silently skipped). ctx is the caller's context (F8):
// AuthHook.Authenticator is invoked with it directly, never
// context.Background(), so a hook that performs a network call (e.g. a
// github_app JWT->installation-token exchange) honors the caller's
// cancellation/deadline.
func buildCustomAuth(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec, h Hooks) (connsdk.Authenticator, error) {
	if h == nil {
		return nil, fmt.Errorf("custom auth: hook %q not registered (no hooks provided)", spec.Hook)
	}
	authHook, ok := h.(AuthHook)
	if !ok {
		return nil, fmt.Errorf("custom auth: hook %q not registered (hooks %q does not implement AuthHook)", spec.Hook, h.ConnectorName())
	}
	return authHook.Authenticator(ctx, cfg, spec)
}
