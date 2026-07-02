// Package github is the Tier-2 escape hatch: a github_app JWT->installation-
// token AuthHook (ports auth.go) + WriteHook for compound writes/label color.
package github

import (
	"bytes"
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

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("github", func() engine.Hooks { return New() })
}

// Hooks is the github bundle's stateless Tier-2 hook set.
type Hooks struct{}

// New returns a fresh Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "github" }

var (
	_ engine.Hooks     = (*Hooks)(nil)
	_ engine.AuthHook  = (*Hooks)(nil)
	_ engine.WriteHook = (*Hooks)(nil)
)

// Authenticator mints an RS256 JWT (matches legacy's githubAppJWT) and
// exchanges it for an installation token via POST
// /app/installations/{id}/access_tokens, returning a Bearer authenticator.
// ctx is honored (a real network call); uncached, matching legacy's own
// re-mint-on-every-call behavior.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	appID := strings.TrimSpace(cfg.Config["app_id"])
	if appID == "" {
		return nil, errors.New("github auth_type=github_app requires config app_id")
	}
	installationID := strings.TrimSpace(cfg.Config["installation_id"])
	if installationID == "" {
		return nil, errors.New("github auth_type=github_app requires config installation_id")
	}
	key, err := parsePrivateKey(cfg)
	if err != nil {
		return nil, err
	}
	jwt, err := signAppJWT(appID, key, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.Config["base_url"]), "/")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	endpoint := baseURL + "/app/installations/" + url.PathEscape(installationID) + "/access_tokens"
	body, err := json.Marshal(installationTokenPayload(cfg))
	if err != nil {
		return nil, fmt.Errorf("github_app: encode installation token request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("github_app: build installation token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github_app: exchange installation token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github_app: installation token exchange returned status %d", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("github_app: decode installation token response: %w", err)
	}
	if strings.TrimSpace(out.Token) == "" {
		return nil, errors.New("github_app: installation token response did not include token")
	}
	return connsdk.Bearer(out.Token), nil
}

func signAppJWT(issuer string, key *rsa.PrivateKey, now time.Time) (string, error) {
	header, err := base64JSON(map[string]string{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := base64JSON(map[string]any{
		"iat": now.Add(-60 * time.Second).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": issuer,
	})
	if err != nil {
		return "", err
	}
	signingInput := header + "." + payload
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("github_app: sign jwt: %w", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func base64JSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("github_app: encode jwt segment: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func parsePrivateKey(cfg connectors.RuntimeConfig) (*rsa.PrivateKey, error) {
	material := strings.TrimSpace(cfg.Secrets["private_key"])
	if material == "" {
		encoded := strings.TrimSpace(cfg.Secrets["private_key_base64"])
		if encoded == "" {
			return nil, errors.New("github auth_type=github_app requires private_key or private_key_base64 secret")
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("github_app: decode private_key_base64: %w", err)
		}
		material = strings.TrimSpace(string(decoded))
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
		return nil, fmt.Errorf("github_app: parse private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("github private key must be RSA")
	}
	return key, nil
}

func installationTokenPayload(cfg connectors.RuntimeConfig) map[string]any {
	payload := map[string]any{}
	if repos := compactSplit(cfg.Config["installation_repositories"]); len(repos) > 0 {
		payload["repositories"] = repos
	}
	if idsRaw := compactSplit(cfg.Config["installation_repository_ids"]); len(idsRaw) > 0 {
		ids := make([]int, 0, len(idsRaw))
		for _, raw := range idsRaw {
			if n, err := strconv.Atoi(raw); err == nil {
				ids = append(ids, n)
			}
		}
		if len(ids) > 0 {
			payload["repository_ids"] = ids
		}
	}
	if raw := strings.TrimSpace(cfg.Config["installation_permissions"]); raw != "" {
		var perms map[string]string
		if json.Unmarshal([]byte(raw), &perms) == nil {
			payload["permissions"] = perms
		}
	}
	return payload
}

func compactSplit(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	out := make([]string, 0, 4)
	for _, p := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// --- WriteHook: compound writes + label color-strip normalization --------

var metaFields = []string{"labels", "assignees", "milestone"}
var reviewerFields = []string{"reviewers", "team_reviewers"}
var pullCoreFields = []string{"title", "body", "state", "base", "maintainer_can_modify"}

// ExecuteWrite: 4 compound actions + label color normalization; anything
// else returns handled=false (declarative fallback).
func (h *Hooks) ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (bool, error) {
	switch action.Name {
	case "close_issue":
		return true, closeResource(ctx, rt, "issues", "issue_number", rec)
	case "close_pull_request":
		return true, closeResource(ctx, rt, "pulls", "pull_number", rec)
	case "create_pull_request":
		return true, createPullRequest(ctx, rt, rec)
	case "update_pull_request":
		return true, updatePullRequest(ctx, rt, rec)
	case "create_label":
		return true, createLabel(ctx, rt, rec)
	case "update_label":
		return true, updateLabel(ctx, rt, rec)
	default:
		return false, nil
	}
}

// createLabel/updateLabel reproduce githubCreateLabelPayload/
// githubUpdateLabelPayload: a leading "#" on color is stripped
// (github.go:1120,1133; ledger G3 — update_label's fields are all optional).
func createLabel(ctx context.Context, rt *engine.Runtime, rec connectors.Record) error {
	name, color := optionalString(rec, "name"), optionalString(rec, "color")
	if name == "" || color == "" {
		return fmt.Errorf("name and color are required")
	}
	payload := map[string]any{"name": name, "color": strings.TrimPrefix(color, "#")}
	if v := optionalString(rec, "description"); v != "" {
		payload["description"] = v
	}
	_, err := rt.Requester.Do(ctx, http.MethodPost, repoPath(rt)+"/labels", nil, payload)
	return err
}

func updateLabel(ctx context.Context, rt *engine.Runtime, rec connectors.Record) error {
	name := optionalString(rec, "name")
	if name == "" {
		return fmt.Errorf("name is required")
	}
	payload := map[string]any{}
	for _, key := range []string{"new_name", "color", "description"} {
		if v := optionalString(rec, key); v != "" {
			if key == "color" {
				v = strings.TrimPrefix(v, "#")
			}
			payload[key] = v
		}
	}
	path := fmt.Sprintf("%s/labels/%s", repoPath(rt), url.PathEscape(name))
	_, err := rt.Requester.Do(ctx, http.MethodPatch, path, nil, payload)
	return err
}

func repoPath(rt *engine.Runtime) string {
	owner := strings.TrimSpace(rt.Config.Config["owner"])
	repo := strings.TrimSpace(rt.Config.Config["repo"])
	return "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo)
}

// closeResource is close_issue/close_pull_request's shared shape: an
// optional comment POST (always against the issues comments endpoint — a PR
// IS an issue in GitHub's data model) then a state=closed PATCH against
// resource ("issues" or "pulls").
func closeResource(ctx context.Context, rt *engine.Runtime, resource, numberField string, rec connectors.Record) error {
	number, err := requiredInt(rec, numberField, "number")
	if err != nil {
		return err
	}
	if comment := optionalString(rec, "comment"); comment != "" {
		if err := postComment(ctx, rt, "issues", number, comment); err != nil {
			return err
		}
	}
	payload := map[string]any{"state": "closed"}
	if resource == "issues" {
		if reason := optionalString(rec, "state_reason"); reason != "" {
			payload["state_reason"] = reason
		}
	}
	path := fmt.Sprintf("%s/%s/%d", repoPath(rt), resource, number)
	_, err = rt.Requester.Do(ctx, http.MethodPatch, path, nil, payload)
	return err
}

func createPullRequest(ctx context.Context, rt *engine.Runtime, rec connectors.Record) error {
	skip := map[string]bool{"labels": true, "assignees": true, "milestone": true, "reviewers": true, "team_reviewers": true}
	payload := map[string]any{}
	for k, v := range rec {
		if !skip[k] {
			payload[k] = v
		}
	}
	resp, err := rt.Requester.Do(ctx, http.MethodPost, repoPath(rt)+"/pulls", nil, payload)
	if err != nil {
		return err
	}
	var created struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(resp.Body, &created); err != nil || created.Number == 0 {
		return fmt.Errorf("github_app: create_pull_request response missing number: %w", err)
	}
	return pullRequestFollowups(ctx, rt, created.Number, rec)
}

func updatePullRequest(ctx context.Context, rt *engine.Runtime, rec connectors.Record) error {
	number, err := requiredInt(rec, "pull_number", "number")
	if err != nil {
		return err
	}
	if core := selectFields(rec, pullCoreFields); len(core) > 0 {
		path := fmt.Sprintf("%s/pulls/%d", repoPath(rt), number)
		if _, err := rt.Requester.Do(ctx, http.MethodPatch, path, nil, core); err != nil {
			return err
		}
	}
	return pullRequestFollowups(ctx, rt, number, rec)
}

// pullRequestFollowups sends the optional issue-metadata PATCH then the
// optional reviewers POST.
func pullRequestFollowups(ctx context.Context, rt *engine.Runtime, number int, rec connectors.Record) error {
	if meta := selectFields(rec, metaFields); len(meta) > 0 {
		path := fmt.Sprintf("%s/issues/%d", repoPath(rt), number)
		if _, err := rt.Requester.Do(ctx, http.MethodPatch, path, nil, meta); err != nil {
			return err
		}
	}
	reviewers := selectFields(rec, reviewerFields)
	if len(reviewers) == 0 {
		return nil
	}
	path := fmt.Sprintf("%s/pulls/%d/requested_reviewers", repoPath(rt), number)
	_, err := rt.Requester.Do(ctx, http.MethodPost, path, nil, reviewers)
	return err
}

func postComment(ctx context.Context, rt *engine.Runtime, resource string, number int, body string) error {
	path := fmt.Sprintf("%s/%s/%d/comments", repoPath(rt), resource, number)
	_, err := rt.Requester.Do(ctx, http.MethodPost, path, nil, map[string]any{"body": body})
	return err
}

func selectFields(rec connectors.Record, keys []string) map[string]any {
	out := map[string]any{}
	for _, k := range keys {
		if v, ok := rec[k]; ok {
			out[k] = v
		}
	}
	return out
}

func requiredInt(rec connectors.Record, keys ...string) (int, error) {
	for _, k := range keys {
		v, ok := rec[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case int:
			return t, nil
		case int64:
			return int(t), nil
		case float64:
			return int(t), nil
		case json.Number:
			if n, err := t.Int64(); err == nil {
				return int(n), nil
			}
		}
	}
	return 0, fmt.Errorf("%s is required", strings.Join(keys, " or "))
}

func optionalString(rec connectors.Record, key string) string {
	v, ok := rec[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return strings.TrimSpace(s)
}
