package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/credential"
)

// declaredSessionAuthenticator is intentionally narrower than a generic
// password grant: it sends only the declared username/password JSON object to
// the fixed /session route, extracts only the top-level id, and never refreshes
// or replays the exchange.
type declaredSessionAuthenticator struct {
	requester DeclaredRouteRequester
	username  string
	password  string

	mu      sync.Mutex
	session string
}

// buildDeclaredSessionAuthenticator admits only a literal HTTPS /session
// declaration. The exchange itself uses the runtime's already-resolved
// declaration-bound base URL, so neither its origin nor its route is caller
// controlled.
func buildDeclaredSessionAuthenticator(spec AuthSpec, vars Vars, requester DeclaredRouteRequester) (connsdk.Authenticator, error) {
	if strings.Contains(spec.TokenURL, "{{") || strings.Contains(spec.TokenURL, "}}") {
		return nil, errors.New("declared_session: token_url must be static")
	}
	u, err := url.Parse(spec.TokenURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.Path != "/session" || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("declared_session: token_url must be one fixed HTTPS /session route")
	}
	if requester == nil {
		return nil, errors.New("declared_session: declared route requester is unavailable")
	}
	username, err := interpolateCredential(spec.Username, vars)
	if err != nil {
		return nil, fmt.Errorf("declared_session: username: %w", err)
	}
	if err := credential.RequireAuthenticationValue("username", username); err != nil {
		return nil, fmt.Errorf("declared_session: %w", err)
	}
	password, err := interpolateCredential(spec.Password, vars)
	if err != nil {
		return nil, fmt.Errorf("declared_session: password: %w", err)
	}
	if err := credential.RequireAuthenticationValue("password", password); err != nil {
		return nil, fmt.Errorf("declared_session: %w", err)
	}
	return &declaredSessionAuthenticator{requester: requester, username: username, password: password}, nil
}

func (a *declaredSessionAuthenticator) Apply(ctx context.Context, request *http.Request) error {
	session, err := a.sessionToken(ctx)
	if err != nil {
		return err
	}
	request.Header.Set("X-Metabase-Session", session)
	return nil
}

func (a *declaredSessionAuthenticator) sessionToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session != "" {
		return a.session, nil
	}
	response, err := a.requester.DoJSON(ctx, DeclaredRouteRequest{
		Method:       http.MethodPost,
		DeclaredPath: "/session",
		Path:         "/session",
		Headers:      map[string]string{"Accept": "application/json"},
		Body: struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{Username: a.username, Password: a.password},
	})
	if err != nil {
		return "", fmt.Errorf("declared_session: request session: %w", err)
	}
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return "", errors.New("declared_session: decode session response")
	}
	if err := credential.RequireAuthenticationValue("session token", payload.ID); err != nil {
		return "", fmt.Errorf("declared_session: %w", err)
	}
	a.session = payload.ID
	return a.session, nil
}
