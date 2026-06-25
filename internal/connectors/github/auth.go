package github

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics/internal/connectors"
)

type githubAuthMode string

const (
	githubAuthAuto    githubAuthMode = "auto"
	githubAuthPublic  githubAuthMode = "public"
	githubAuthToken   githubAuthMode = "token"
	githubAuthApp     githubAuthMode = "github_app"
	githubAuthOAuth   githubAuthMode = "oauth"
	githubAuthActions githubAuthMode = "github_actions"
	githubAuthInstall githubAuthMode = "installation_token"
)

func (g Connector) authorizationHeader(ctx context.Context, cfg connectors.RuntimeConfig) (string, error) {
	mode, err := githubResolveAuthMode(cfg)
	if err != nil {
		return "", err
	}
	switch mode {
	case githubAuthPublic:
		return "", nil
	case githubAuthToken, githubAuthOAuth, githubAuthActions, githubAuthInstall:
		token := githubToken(cfg)
		if token == "" {
			return "", fmt.Errorf("github auth_type=%s requires a token secret", mode)
		}
		return "Bearer " + token, nil
	case githubAuthApp:
		token, err := g.githubAppInstallationToken(ctx, cfg)
		if err != nil {
			return "", err
		}
		return "Bearer " + token, nil
	default:
		return "", fmt.Errorf("unsupported github auth_type %q", mode)
	}
}

func githubResolveAuthMode(cfg connectors.RuntimeConfig) (githubAuthMode, error) {
	raw := strings.TrimSpace(strings.ToLower(firstNonEmptyString(
		cfg.Config["auth_type"],
		cfg.Config["auth"],
		cfg.Config["authentication"],
	)))
	if raw == "" {
		raw = string(githubAuthAuto)
	}
	raw = strings.ReplaceAll(raw, "-", "_")
	raw = strings.ReplaceAll(raw, " ", "_")
	switch githubAuthMode(raw) {
	case githubAuthAuto:
		if githubToken(cfg) != "" {
			return githubAuthToken, nil
		}
		if githubHasAppConfig(cfg) {
			return githubAuthApp, nil
		}
		return githubAuthPublic, nil
	case githubAuthPublic, "none", "anonymous", "unauthenticated":
		return githubAuthPublic, nil
	case githubAuthToken, "pat", "personal_access_token", "fine_grained_pat", "classic_pat":
		return githubAuthToken, nil
	case githubAuthOAuth, "oauth_token":
		return githubAuthOAuth, nil
	case githubAuthActions, "actions", "github_token":
		return githubAuthActions, nil
	case githubAuthInstall, "installation":
		return githubAuthInstall, nil
	case githubAuthApp, "app", "github_app_installation":
		return githubAuthApp, nil
	default:
		return "", fmt.Errorf("unsupported github auth_type %q", raw)
	}
}

func githubHasWriteAuth(cfg connectors.RuntimeConfig) bool {
	mode, err := githubResolveAuthMode(cfg)
	if err != nil {
		return false
	}
	switch mode {
	case githubAuthToken, githubAuthOAuth, githubAuthActions, githubAuthInstall:
		return githubToken(cfg) != ""
	case githubAuthApp:
		return githubHasAppConfig(cfg) && githubPrivateKeyMaterial(cfg) != ""
	default:
		return false
	}
}

func githubHasAppConfig(cfg connectors.RuntimeConfig) bool {
	return githubAppID(cfg) != "" && strings.TrimSpace(cfg.Config["installation_id"]) != ""
}

func (g Connector) githubAppInstallationToken(ctx context.Context, cfg connectors.RuntimeConfig) (string, error) {
	appID := githubAppID(cfg)
	if appID == "" {
		return "", errors.New("github auth_type=github_app requires config app_id")
	}
	installationID := strings.TrimSpace(cfg.Config["installation_id"])
	if installationID == "" {
		return "", errors.New("github auth_type=github_app requires config installation_id")
	}
	privateKey, err := githubParsePrivateKey(cfg)
	if err != nil {
		return "", err
	}
	jwt, err := githubAppJWT(appID, privateKey, time.Now().UTC())
	if err != nil {
		return "", err
	}
	endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/app/installations/%s/access_tokens", url.PathEscape(installationID)), nil)
	if err != nil {
		return "", err
	}
	payload, err := githubInstallationTokenPayload(cfg)
	if err != nil {
		return "", err
	}
	var response struct {
		Token string `json:"token"`
	}
	if err := g.doJSONWithAuth(ctx, cfg, http.MethodPost, endpoint, payload, &response, "Bearer "+jwt); err != nil {
		return "", fmt.Errorf("create GitHub App installation token: %w", err)
	}
	if strings.TrimSpace(response.Token) == "" {
		return "", errors.New("GitHub App installation token response did not include token")
	}
	return response.Token, nil
}

func githubAppJWT(issuer string, privateKey *rsa.PrivateKey, now time.Time) (string, error) {
	header, err := githubBase64JSON(map[string]string{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := githubBase64JSON(map[string]any{
		"iat": now.Add(-60 * time.Second).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": issuer,
	})
	if err != nil {
		return "", err
	}
	signingInput := header + "." + payload
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("sign GitHub App JWT: %w", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func githubBase64JSON(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode GitHub App JWT: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func githubParsePrivateKey(cfg connectors.RuntimeConfig) (*rsa.PrivateKey, error) {
	material := githubPrivateKeyMaterial(cfg)
	if material == "" {
		return nil, errors.New("github auth_type=github_app requires private_key or private_key_base64 secret")
	}
	block, _ := pem.Decode([]byte(material))
	if block == nil {
		return nil, errors.New("github private key must be PEM encoded")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse GitHub App private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("github private key must be RSA")
	}
	return key, nil
}

func githubPrivateKeyMaterial(cfg connectors.RuntimeConfig) string {
	if value := firstNonEmptyString(
		cfg.Secrets["private_key"],
		cfg.Secrets["privateKey"],
		cfg.Secrets["githubAppPrivateKey"],
	); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	encoded := firstNonEmptyString(
		cfg.Secrets["private_key_base64"],
		cfg.Secrets["privateKeyBase64"],
		cfg.Secrets["githubAppPrivateKeyBase64"],
	)
	if strings.TrimSpace(encoded) == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(decoded))
}

func githubInstallationTokenPayload(cfg connectors.RuntimeConfig) (map[string]any, error) {
	payload := map[string]any{}
	if repositories := compactStrings(strings.Split(cfg.Config["installation_repositories"], ",")); len(repositories) > 0 {
		payload["repositories"] = repositories
	}
	if ids := compactStrings(strings.Split(cfg.Config["installation_repository_ids"], ",")); len(ids) > 0 {
		out := make([]int, 0, len(ids))
		for _, raw := range ids {
			id, err := strconv.Atoi(raw)
			if err != nil {
				return nil, fmt.Errorf("installation_repository_ids must be comma-separated integers: %w", err)
			}
			out = append(out, id)
		}
		payload["repository_ids"] = out
	}
	if raw := strings.TrimSpace(cfg.Config["installation_permissions"]); raw != "" {
		var permissions map[string]string
		if err := json.Unmarshal([]byte(raw), &permissions); err != nil {
			return nil, fmt.Errorf("installation_permissions must be a JSON object: %w", err)
		}
		payload["permissions"] = permissions
	}
	return payload, nil
}

func githubAppID(cfg connectors.RuntimeConfig) string {
	return strings.TrimSpace(firstNonEmptyString(cfg.Config["app_id"], cfg.Config["client_id"], cfg.Config["github_app_id"]))
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func githubAuthModeSpecs() []connectors.AuthModeSpec {
	return []connectors.AuthModeSpec{
		{
			Name:         "public",
			Description:  "Unauthenticated public repository reads. Writes are not allowed.",
			ConfigFields: []string{"repository", "base_url"},
			Read:         true,
			Write:        false,
		},
		{
			Name:         "token",
			Description:  "Bearer-token auth for classic PATs, fine-grained PATs, OAuth tokens, GitHub Actions GITHUB_TOKEN, or pre-generated installation tokens.",
			ConfigFields: []string{"repository", "base_url", "auth_type=token"},
			SecretFields: []string{"token", "personalAccessToken", "oauthToken", "accessToken", "installationToken", "githubToken"},
			Read:         true,
			Write:        true,
		},
		{
			Name:         "github_app",
			Description:  "Server-to-server auth. pm signs a GitHub App JWT and exchanges it for a one-hour installation access token.",
			ConfigFields: []string{"repository", "base_url", "auth_type=github_app", "app_id", "installation_id", "installation_repositories", "installation_repository_ids", "installation_permissions"},
			SecretFields: []string{"private_key", "private_key_base64"},
			Read:         true,
			Write:        true,
		},
	}
}
