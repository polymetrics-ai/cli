// Package feishu implements the declared tenant-token authentication extension
// for the Feishu Bitable execution bundle.
package feishu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/credential"
)

const tenantAccessTokenPath = "/open-apis/auth/v3/tenant_access_token/internal"

func init() {
	engine.RegisterHooks("feishu", func() engine.Hooks { return New() })
}

// Hooks owns only the source-declared Feishu tenant token exchange; the shared
// engine owns Bitable request construction, pagination, and record handling.
type Hooks struct{}

// New returns a fresh Feishu authentication hook set.
func New() *Hooks { return &Hooks{} }

func (*Hooks) ConnectorName() string { return "feishu" }

var (
	_ engine.Hooks                 = (*Hooks)(nil)
	_ engine.AuthHook              = (*Hooks)(nil)
	_ engine.DeclaredRouteAuthHook = (*Hooks)(nil)
)

// Authenticator rejects direct token exchanges because the engine must admit
// every network-capable authentication request through the declared route.
func (*Hooks) Authenticator(context.Context, connectors.RuntimeConfig, engine.AuthSpec) (connsdk.Authenticator, error) {
	return nil, errors.New("feishu tenant authentication requires engine declared-route admission")
}

// AuthenticatorWithDeclaredRoute exchanges the configured app credentials for
// one tenant access token through the fixed provider token route.
func (*Hooks) AuthenticatorWithDeclaredRoute(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec, requester engine.DeclaredRouteRequester) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if spec.Mode != "custom" || spec.Hook != "feishu" {
		return nil, errors.New("feishu tenant authentication requires the declared feishu custom auth hook")
	}
	if requester == nil {
		return nil, errors.New("feishu tenant authentication requires engine declared-route admission")
	}

	appID := strings.TrimSpace(cfg.Secrets["app_id"])
	if err := credential.RequireAuthenticationValue("Feishu app ID", appID); err != nil {
		return nil, fmt.Errorf("feishu tenant authentication: %w", err)
	}
	appSecret := strings.TrimSpace(cfg.Secrets["app_secret"])
	if err := credential.RequireAuthenticationValue("Feishu app secret", appSecret); err != nil {
		return nil, fmt.Errorf("feishu tenant authentication: %w", err)
	}

	response, err := requester.DoJSON(ctx, engine.DeclaredRouteRequest{
		Method:       http.MethodPost,
		DeclaredPath: tenantAccessTokenPath,
		Path:         tenantAccessTokenPath,
		Headers: map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json; charset=utf-8",
		},
		Body: map[string]string{
			"app_id":     appID,
			"app_secret": appSecret,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("feishu tenant token exchange: %w", err)
	}

	var payload struct {
		Code              json.Number `json:"code"`
		TenantAccessToken string      `json:"tenant_access_token"`
	}
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return nil, fmt.Errorf("feishu tenant token response: %w", err)
	}
	if code := payload.Code.String(); code != "" && code != "0" {
		return nil, fmt.Errorf("feishu tenant token exchange failed with code %s", code)
	}
	if strings.TrimSpace(payload.TenantAccessToken) == "" {
		return nil, errors.New("feishu tenant token response did not include tenant_access_token")
	}
	return connsdk.Bearer(payload.TenantAccessToken), nil
}
