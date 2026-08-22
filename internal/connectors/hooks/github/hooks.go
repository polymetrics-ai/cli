// Package github is the Tier-2 escape hatch: a github_app JWT->installation-
// token AuthHook (ports auth.go) + WriteHook for compound writes/label color.
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
	"io"
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
	_ engine.Hooks                    = (*Hooks)(nil)
	_ engine.AuthHook                 = (*Hooks)(nil)
	_ engine.DeclaredRouteAuthHook    = (*Hooks)(nil)
	_ engine.RateLimitAuthProfileHook = (*Hooks)(nil)
	_ engine.WriteHook                = (*Hooks)(nil)
	_ engine.PreparedWriteHook        = (*Hooks)(nil)
)

const (
	githubInstallationTokenDeclaredPath          = "/app/installations/{installation_id}/access_tokens"
	githubInstallationRepositoryRestrictionLimit = 500
)

// Authenticator remains the base AuthHook implementation for compatibility,
// but deliberately refuses a direct GitHub App exchange. The token request is
// network-capable and must receive the engine-owned declared-route requester.
func (h *Hooks) Authenticator(context.Context, connectors.RuntimeConfig, engine.AuthSpec) (connsdk.Authenticator, error) {
	return nil, errors.New("github_app authentication requires engine declared-route admission")
}

// RateLimitAuthProfile identifies GitHub App custom authentication for
// rate-limit selection.
func (h *Hooks) RateLimitAuthProfile(_ connectors.RuntimeConfig, spec engine.AuthSpec) (string, bool) {
	if spec.Mode != "custom" || spec.Hook != "github" {
		return "", false
	}
	return "github_app", true
}

// AuthenticatorWithDeclaredRoute mints an RS256 JWT (matching legacy's
// githubAppJWT) and exchanges it for an installation token through the
// declared `POST /app/installations/{installation_id}/access_tokens` route.
// The actual path contains the escaped configured installation ID; the engine
// admits that physical path immediately before sending it.
func (h *Hooks) AuthenticatorWithDeclaredRoute(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec, requester engine.DeclaredRouteRequester) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if requester == nil {
		return nil, errors.New("github_app authentication requires engine declared-route admission")
	}
	appID := strings.TrimSpace(cfg.Config["app_id"])
	if appID == "" {
		return nil, errors.New("github auth_type=github_app requires config app_id")
	}
	installationID := strings.TrimSpace(cfg.Config["installation_id"])
	if installationID == "" {
		return nil, errors.New("github auth_type=github_app requires config installation_id")
	}
	payload, err := installationTokenPayload(cfg)
	if err != nil {
		return nil, err
	}
	key, err := parsePrivateKey(cfg)
	if err != nil {
		return nil, err
	}
	jwt, err := signAppJWT(appID, key, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	resp, err := requester.DoJSON(ctx, engine.DeclaredRouteRequest{
		Method:       http.MethodPost,
		DeclaredPath: githubInstallationTokenDeclaredPath,
		Path:         "/app/installations/" + url.PathEscape(installationID) + "/access_tokens",
		Headers: map[string]string{
			"Authorization": "Bearer " + jwt,
			"Accept":        "application/vnd.github+json",
		},
		Body: payload,
	})
	if err != nil {
		return nil, fmt.Errorf("github_app: exchange installation token: %w", err)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(resp.Body, &out); err != nil {
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

func installationTokenPayload(cfg connectors.RuntimeConfig) (map[string]any, error) {
	payload := map[string]any{}
	repos, err := installationRepositories(cfg.Config["installation_repositories"])
	if err != nil {
		return nil, err
	}
	ids, err := installationRepositoryIDs(cfg.Config["installation_repository_ids"])
	if err != nil {
		return nil, err
	}
	if len(repos)+len(ids) > githubInstallationRepositoryRestrictionLimit {
		return nil, fmt.Errorf("github installation_repositories and installation_repository_ids support at most %d repositories combined", githubInstallationRepositoryRestrictionLimit)
	}
	if len(repos) > 0 {
		payload["repositories"] = repos
	}
	if len(ids) > 0 {
		payload["repository_ids"] = ids
	}
	if permissions, err := installationPermissions(cfg.Config["installation_permissions"]); err != nil {
		return nil, err
	} else if permissions != nil {
		payload["permissions"] = permissions
	}
	return payload, nil
}

func installationRepositories(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, repository := range strings.Split(raw, ",") {
		if repository == "" || strings.TrimSpace(repository) != repository || !githubRepositoryName(repository) {
			return nil, errors.New("github installation_repositories must be a comma-separated list of repository names")
		}
		identity := strings.ToLower(repository)
		if _, duplicate := seen[identity]; duplicate {
			return nil, fmt.Errorf("github installation_repositories repeats %q", repository)
		}
		if len(out) >= githubInstallationRepositoryRestrictionLimit {
			return nil, fmt.Errorf("github installation_repositories supports at most %d repositories", githubInstallationRepositoryRestrictionLimit)
		}
		seen[identity] = struct{}{}
		out = append(out, repository)
	}
	return out, nil
}

func githubRepositoryName(value string) bool {
	if len(value) > 100 {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || character == '_' || character == '.' {
			continue
		}
		return false
	}
	return true
}

func installationRepositoryIDs(raw string) ([]int64, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	ids := make([]int64, 0, 4)
	seen := map[int64]struct{}{}
	for _, rawID := range strings.Split(raw, ",") {
		if rawID == "" || strings.TrimSpace(rawID) != rawID || !decimalIdentifier(rawID) {
			return nil, errors.New("github installation_repository_ids must be a comma-separated list of positive integers")
		}
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || id <= 0 {
			return nil, errors.New("github installation_repository_ids must be a comma-separated list of positive integers")
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, fmt.Errorf("github installation_repository_ids repeats %d", id)
		}
		if len(ids) >= githubInstallationRepositoryRestrictionLimit {
			return nil, fmt.Errorf("github installation_repository_ids supports at most %d repositories", githubInstallationRepositoryRestrictionLimit)
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func decimalIdentifier(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return value != ""
}

func installationPermissions(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("github installation_permissions must be a JSON object")
	}
	permissions := map[string]string{}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, errors.New("github installation_permissions must be a JSON object")
		}
		name, ok := keyToken.(string)
		if !ok || !githubPermissionName(name) {
			return nil, errors.New("github installation_permissions contains an invalid permission name")
		}
		allowed, supported := githubInstallationPermissionMatrix[name]
		if !supported {
			return nil, fmt.Errorf("github installation_permissions.%s is not supported", name)
		}
		if _, duplicate := permissions[name]; duplicate {
			return nil, fmt.Errorf("github installation_permissions repeats %q", name)
		}
		var level string
		if err := decoder.Decode(&level); err != nil || !githubPermissionLevel(level) {
			return nil, fmt.Errorf("github installation_permissions.%s must be read, write, or admin", name)
		}
		if allowed&githubInstallationPermissionLevel(level) == 0 {
			return nil, fmt.Errorf("github installation_permissions.%s does not support %q", name, level)
		}
		permissions[name] = level
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, errors.New("github installation_permissions must be a JSON object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("github installation_permissions must contain one JSON object")
	}
	return permissions, nil
}

func githubPermissionName(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character == '_' {
			continue
		}
		return false
	}
	return true
}

func githubPermissionLevel(value string) bool {
	return value == "read" || value == "write" || value == "admin"
}

type githubInstallationPermissionLevels uint8

const (
	githubInstallationPermissionRead githubInstallationPermissionLevels = 1 << iota
	githubInstallationPermissionWrite
	githubInstallationPermissionAdmin
)

var githubInstallationPermissionMatrix = map[string]githubInstallationPermissionLevels{
	"actions":                             githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"administration":                      githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"artifact_metadata":                   githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"attestations":                        githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"checks":                              githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"code_quality":                        githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"codespaces":                          githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"contents":                            githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"custom_properties_for_organizations": githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"dependabot_secrets":                  githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"deployments":                         githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"discussions":                         githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"email_addresses":                     githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"enterprise_custom_properties_for_organizations": githubInstallationPermissionRead | githubInstallationPermissionWrite | githubInstallationPermissionAdmin,
	"environments":                                githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"followers":                                   githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"git_ssh_keys":                                githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"gpg_keys":                                    githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"interaction_limits":                          githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"issues":                                      githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"members":                                     githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"merge_queues":                                githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"metadata":                                    githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_administration":                 githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_announcement_banners":           githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_copilot_agent_settings":         githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_copilot_seat_management":        githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_custom_org_roles":               githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_custom_properties":              githubInstallationPermissionRead | githubInstallationPermissionWrite | githubInstallationPermissionAdmin,
	"organization_custom_roles":                   githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_events":                         githubInstallationPermissionRead,
	"organization_hooks":                          githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_packages":                       githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_personal_access_token_requests": githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_personal_access_tokens":         githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_plan":                           githubInstallationPermissionRead,
	"organization_projects":                       githubInstallationPermissionRead | githubInstallationPermissionWrite | githubInstallationPermissionAdmin,
	"organization_secrets":                        githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_self_hosted_runners":            githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"organization_user_blocking":                  githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"packages":                                    githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"pages":                                       githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"profile":                                     githubInstallationPermissionWrite,
	"pull_requests":                               githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"repository_custom_properties":                githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"repository_hooks":                            githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"repository_projects":                         githubInstallationPermissionRead | githubInstallationPermissionWrite | githubInstallationPermissionAdmin,
	"secret_scanning_alerts":                      githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"secrets":                                     githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"security_events":                             githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"single_file":                                 githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"starring":                                    githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"statuses":                                    githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"vulnerability_alerts":                        githubInstallationPermissionRead | githubInstallationPermissionWrite,
	"workflows":                                   githubInstallationPermissionWrite,
}

func githubInstallationPermissionLevel(value string) githubInstallationPermissionLevels {
	switch value {
	case "read":
		return githubInstallationPermissionRead
	case "write":
		return githubInstallationPermissionWrite
	case "admin":
		return githubInstallationPermissionAdmin
	default:
		return 0
	}
}

// --- WriteHook: compound writes + label color-strip normalization --------

var metaFields = []string{"labels", "assignees", "milestone"}
var reviewerFields = []string{"reviewers", "team_reviewers"}
var pullCoreFields = []string{"title", "body", "state", "base", "maintainer_can_modify"}

// Every path below is an existing GitHub bundle declaration: action paths for
// the primary write and api_surface paths for a hook-only follow-up. Keeping
// the declarations here makes a hook request use the same rate-limit
// resolution boundary as a declarative request without exposing a generic
// request escape hatch.
const (
	githubIssueDeclaredPath              = "/repos/{owner}/{repo}/issues/{issue_number}"
	githubIssueCommentDeclaredPath       = "/repos/{owner}/{repo}/issues/{issue_number}/comments"
	githubLabelDeclaredPath              = "/repos/{owner}/{repo}/labels"
	githubLabelByNameDeclaredPath        = "/repos/{owner}/{repo}/labels/{name}"
	githubPullDeclaredPath               = "/repos/{owner}/{repo}/pulls"
	githubPullByNumberDeclaredPath       = "/repos/{owner}/{repo}/pulls/{pull_number}"
	githubRequestedReviewersDeclaredPath = "/repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers"
)

// githubWriteRequester acquires one declaration-aware requester for exactly
// one physical GitHub WriteHook send. A compound action intentionally calls it
// again for every follow-up so each request gets an independent rate-limit
// admission and completion observation lifecycle.
func githubWriteRequester(rt *engine.Runtime, method, declaredPath string) (*connsdk.Requester, error) {
	if rt == nil {
		return nil, errors.New("github write hook runtime is unavailable")
	}
	requester, err := rt.RequesterFor(method, declaredPath)
	if err != nil {
		return nil, fmt.Errorf("github write hook resolve %s %s: %w", method, declaredPath, err)
	}
	return requester, nil
}

// ExecuteWrite: 4 compound actions + label color normalization; anything
// else returns handled=false (declarative fallback).
func (h *Hooks) ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (bool, []*connsdk.Response, error) {
	switch action.Name {
	case "close_issue":
		responses, err := closeResource(ctx, rt, "issues", "issue_number", rec)
		return true, responses, err
	case "close_pull_request":
		responses, err := closeResource(ctx, rt, "pulls", "pull_number", rec)
		return true, responses, err
	case "reopen_issue":
		responses, err := reopenResource(ctx, rt, "issues", "issue_number", rec)
		return true, responses, err
	case "reopen_pull_request":
		responses, err := reopenResource(ctx, rt, "pulls", "pull_number", rec)
		return true, responses, err
	case "create_pull_request":
		responses, err := createPullRequest(ctx, rt, rec)
		return true, responses, err
	case "update_pull_request":
		responses, err := updatePullRequest(ctx, rt, rec)
		return true, responses, err
	case "create_label":
		responses, err := createLabel(ctx, rt, rec)
		return true, responses, err
	case "update_label":
		responses, err := updateLabel(ctx, rt, rec)
		return true, responses, err
	default:
		return false, nil, nil
	}
}

func (h *Hooks) HandlesWriteAction(action engine.WriteAction) bool {
	switch action.Name {
	case "close_issue", "close_pull_request", "reopen_issue", "reopen_pull_request", "create_pull_request", "update_pull_request", "create_label", "update_label":
		return true
	default:
		return false
	}
}

// PrepareWrite translates the small GitHub compound-action vocabulary into
// ordered references to existing writes.json actions. The hook supplies no
// route, verb, header, or free-form body: the engine resolves all physical
// request bytes from those declarations and seals their digest before
// approval. A create-PR follow-up binds its declared path field only from the
// preceding create receipt's bounded `number` field.
func (h *Hooks) PrepareWrite(action engine.WriteAction, records []connectors.Record) (engine.PreparedWriteHookPlan, bool, error) {
	if !h.HandlesWriteAction(action) {
		return engine.PreparedWriteHookPlan{}, false, nil
	}
	plan := engine.PreparedWriteHookPlan{Records: make([]engine.PreparedWriteHookRecord, len(records))}
	for index, record := range records {
		steps, err := githubPreparedWriteSteps(action.Name, record)
		if err != nil {
			return engine.PreparedWriteHookPlan{}, true, err
		}
		plan.Records[index].Steps = steps
	}
	return plan, true, nil
}

func githubPreparedWriteSteps(action string, rec connectors.Record) ([]engine.PreparedWriteHookStep, error) {
	step := func(action string, record connectors.Record) engine.PreparedWriteHookStep {
		return engine.PreparedWriteHookStep{Action: action, Record: record}
	}
	resolvedDeclarativeStep := func(action string, record connectors.Record) engine.PreparedWriteHookStep {
		return engine.PreparedWriteHookStep{Action: action, Record: record, ResolvedDeclarative: true}
	}
	boundStep := func(action string, record connectors.Record, target string) engine.PreparedWriteHookStep {
		return engine.PreparedWriteHookStep{
			Action: action,
			Record: record,
			ResponseBinding: &engine.PreparedWriteResponseBinding{
				SourceStep:  0,
				Field:       "number",
				TargetField: target,
			},
		}
	}

	switch action {
	case "close_issue":
		number, err := requiredInt(rec, "issue_number", "number")
		if err != nil {
			return nil, err
		}
		steps := make([]engine.PreparedWriteHookStep, 0, 2)
		if comment := optionalString(rec, "comment"); comment != "" {
			steps = append(steps, step("comment_issue", connectors.Record{"issue_number": number, "body": comment}))
		}
		update := connectors.Record{"issue_number": number, "state": "closed"}
		if reason := optionalString(rec, "state_reason"); reason != "" {
			update["state_reason"] = reason
		}
		return append(steps, step("update_issue", update)), nil
	case "close_pull_request":
		number, err := requiredInt(rec, "pull_number", "number")
		if err != nil {
			return nil, err
		}
		steps := make([]engine.PreparedWriteHookStep, 0, 2)
		if comment := optionalString(rec, "comment"); comment != "" {
			steps = append(steps, step("comment_issue", connectors.Record{"issue_number": number, "body": comment}))
		}
		return append(steps, resolvedDeclarativeStep("update_pull_request", connectors.Record{"pull_number": number, "state": "closed"})), nil
	case "reopen_issue":
		number, err := requiredInt(rec, "issue_number", "number")
		if err != nil {
			return nil, err
		}
		return []engine.PreparedWriteHookStep{step("update_issue", connectors.Record{"issue_number": number, "state": "open"})}, nil
	case "reopen_pull_request":
		number, err := requiredInt(rec, "pull_number", "number")
		if err != nil {
			return nil, err
		}
		return []engine.PreparedWriteHookStep{resolvedDeclarativeStep("update_pull_request", connectors.Record{"pull_number": number, "state": "open"})}, nil
	case "create_pull_request":
		steps := []engine.PreparedWriteHookStep{step("create_pull_request", cloneGitHubWriteRecord(rec))}
		if meta := selectFields(rec, metaFields); len(meta) > 0 {
			// This value only witnesses update_issue's required path field at
			// preparation. Engine execution overwrites it from create's receipt.
			meta["issue_number"] = 0
			steps = append(steps, boundStep("update_issue", connectors.Record(meta), "issue_number"))
		}
		if reviewers := selectFields(rec, reviewerFields); len(reviewers) > 0 {
			reviewers["pull_number"] = 0
			steps = append(steps, boundStep("request_reviewers", connectors.Record(reviewers), "pull_number"))
		}
		return steps, nil
	case "update_pull_request":
		number, err := requiredInt(rec, "pull_number", "number")
		if err != nil {
			return nil, err
		}
		steps := make([]engine.PreparedWriteHookStep, 0, 3)
		if core := selectFields(rec, pullCoreFields); len(core) > 0 {
			core["pull_number"] = number
			steps = append(steps, step("update_pull_request", connectors.Record(core)))
		}
		if meta := selectFields(rec, metaFields); len(meta) > 0 {
			meta["issue_number"] = number
			steps = append(steps, step("update_issue", connectors.Record(meta)))
		}
		if reviewers := selectFields(rec, reviewerFields); len(reviewers) > 0 {
			reviewers["pull_number"] = number
			steps = append(steps, step("request_reviewers", connectors.Record(reviewers)))
		}
		return steps, nil
	case "create_label":
		name, color := optionalString(rec, "name"), optionalString(rec, "color")
		if name == "" || color == "" {
			return nil, fmt.Errorf("name and color are required")
		}
		label := cloneGitHubWriteRecord(rec)
		label["color"] = strings.TrimPrefix(color, "#")
		return []engine.PreparedWriteHookStep{step("create_label", label)}, nil
	case "update_label":
		name := optionalString(rec, "name")
		if name == "" {
			return nil, fmt.Errorf("name is required")
		}
		label := cloneGitHubWriteRecord(rec)
		if color := optionalString(rec, "color"); color != "" {
			label["color"] = strings.TrimPrefix(color, "#")
		}
		return []engine.PreparedWriteHookStep{step("update_label", label)}, nil
	default:
		return nil, fmt.Errorf("github write hook has no prepared plan for action %q", action)
	}
}

func cloneGitHubWriteRecord(rec connectors.Record) connectors.Record {
	cloned := make(connectors.Record, len(rec))
	for key, value := range rec {
		cloned[key] = value
	}
	return cloned
}

// MapWriteRecord pins the one field that distinguishes `repo archive` from
// `repo unarchive`. Both ride PATCH /repos/{owner}/{repo}, the same endpoint as
// the generic `repo update`, which is why they exist as separate write actions
// rather than as flags: a command named "archive" that only archives when the
// caller also supplies archived=true is a command that lies.
//
// They pin the record rather than the request, unlike closeResource, so the
// declarative path stays intact: preview and execution build one body from one
// record, which is what makes the digest an operator approves the request that
// runs. The action's body_fields allow-list keeps the pinned field the only one
// sent.
func (h *Hooks) MapWriteRecord(action engine.WriteAction, rec connectors.Record) (connectors.Record, bool, error) {
	switch action.Name {
	case "archive_repo":
		return pinRepoArchived(rec, true), true, nil
	case "unarchive_repo":
		return pinRepoArchived(rec, false), true, nil
	default:
		return rec, false, nil
	}
}

func pinRepoArchived(rec connectors.Record, archived bool) connectors.Record {
	pinned := make(connectors.Record, len(rec))
	for key, value := range rec {
		pinned[key] = value
	}
	pinned["archived"] = archived
	return pinned
}

// createLabel/updateLabel reproduce githubCreateLabelPayload/
// githubUpdateLabelPayload: a leading "#" on color is stripped
// (github.go:1120,1133; ledger G3 — update_label's fields are all optional).
func createLabel(ctx context.Context, rt *engine.Runtime, rec connectors.Record) ([]*connsdk.Response, error) {
	name, color := optionalString(rec, "name"), optionalString(rec, "color")
	if name == "" || color == "" {
		return nil, fmt.Errorf("name and color are required")
	}
	payload := map[string]any{"name": name, "color": strings.TrimPrefix(color, "#")}
	if v := optionalString(rec, "description"); v != "" {
		payload["description"] = v
	}
	requester, err := githubWriteRequester(rt, http.MethodPost, githubLabelDeclaredPath)
	if err != nil {
		return nil, err
	}
	response, err := requester.Do(ctx, http.MethodPost, repoPath(rt)+"/labels", nil, payload)
	return responseSlice(response), err
}

func updateLabel(ctx context.Context, rt *engine.Runtime, rec connectors.Record) ([]*connsdk.Response, error) {
	name := optionalString(rec, "name")
	if name == "" {
		return nil, fmt.Errorf("name is required")
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
	requester, err := githubWriteRequester(rt, http.MethodPatch, githubLabelByNameDeclaredPath)
	if err != nil {
		return nil, err
	}
	response, err := requester.Do(ctx, http.MethodPatch, path, nil, payload)
	return responseSlice(response), err
}

func repoPath(rt *engine.Runtime) string {
	owner := strings.TrimSpace(rt.Config.Config["owner"])
	repo := strings.TrimSpace(rt.Config.Config["repo"])
	return "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo)
}

func closeResource(ctx context.Context, rt *engine.Runtime, resource, numberField string, rec connectors.Record) ([]*connsdk.Response, error) {
	number, err := requiredInt(rec, numberField, "number")
	if err != nil {
		return nil, err
	}
	var responses []*connsdk.Response
	if comment := optionalString(rec, "comment"); comment != "" {
		commentResponses, err := postComment(ctx, rt, "issues", number, comment)
		responses = append(responses, commentResponses...)
		if err != nil {
			return responses, err
		}
	}
	payload := map[string]any{"state": "closed"}
	if resource == "issues" {
		if reason := optionalString(rec, "state_reason"); reason != "" {
			payload["state_reason"] = reason
		}
	}
	path := fmt.Sprintf("%s/%s/%d", repoPath(rt), resource, number)
	declaredPath, err := githubResourceDeclaredPath(resource)
	if err != nil {
		return responses, err
	}
	requester, err := githubWriteRequester(rt, http.MethodPatch, declaredPath)
	if err != nil {
		return responses, err
	}
	response, err := requester.Do(ctx, http.MethodPatch, path, nil, payload)
	return append(responses, responseSlice(response)...), err
}

// reopenResource is reopen_issue/reopen_pull_request's shared shape: a
// state=open PATCH against resource ("issues" or "pulls"). It intentionally
// sends no state_reason (reopen has no reason surface in gh) and posts no
// comment, mirroring `gh issue reopen` / `gh pr reopen`.
func reopenResource(ctx context.Context, rt *engine.Runtime, resource, numberField string, rec connectors.Record) ([]*connsdk.Response, error) {
	number, err := requiredInt(rec, numberField, "number")
	if err != nil {
		return nil, err
	}
	payload := map[string]any{"state": "open"}
	path := fmt.Sprintf("%s/%s/%d", repoPath(rt), resource, number)
	declaredPath, err := githubResourceDeclaredPath(resource)
	if err != nil {
		return nil, err
	}
	requester, err := githubWriteRequester(rt, http.MethodPatch, declaredPath)
	if err != nil {
		return nil, err
	}
	response, err := requester.Do(ctx, http.MethodPatch, path, nil, payload)
	return responseSlice(response), err
}

func githubResourceDeclaredPath(resource string) (string, error) {
	switch resource {
	case "issues":
		return githubIssueDeclaredPath, nil
	case "pulls":
		return githubPullByNumberDeclaredPath, nil
	default:
		return "", fmt.Errorf("github write hook has no declared route for resource %q", resource)
	}
}

func createPullRequest(ctx context.Context, rt *engine.Runtime, rec connectors.Record) ([]*connsdk.Response, error) {
	skip := map[string]bool{"labels": true, "assignees": true, "milestone": true, "reviewers": true, "team_reviewers": true}
	payload := map[string]any{}
	for k, v := range rec {
		if !skip[k] {
			payload[k] = v
		}
	}
	requester, err := githubWriteRequester(rt, http.MethodPost, githubPullDeclaredPath)
	if err != nil {
		return nil, err
	}
	resp, err := requester.Do(ctx, http.MethodPost, repoPath(rt)+"/pulls", nil, payload)
	if err != nil {
		return responseSlice(resp), err
	}
	var created struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(resp.Body, &created); err != nil || created.Number == 0 {
		return responseSlice(resp), fmt.Errorf("github_app: create_pull_request response missing number: %w", err)
	}
	followups, err := pullRequestFollowups(ctx, rt, created.Number, rec)
	return append(responseSlice(resp), followups...), err
}

func updatePullRequest(ctx context.Context, rt *engine.Runtime, rec connectors.Record) ([]*connsdk.Response, error) {
	number, err := requiredInt(rec, "pull_number", "number")
	if err != nil {
		return nil, err
	}
	var responses []*connsdk.Response
	if core := selectFields(rec, pullCoreFields); len(core) > 0 {
		path := fmt.Sprintf("%s/pulls/%d", repoPath(rt), number)
		requester, err := githubWriteRequester(rt, http.MethodPatch, githubPullByNumberDeclaredPath)
		if err != nil {
			return responses, err
		}
		response, err := requester.Do(ctx, http.MethodPatch, path, nil, core)
		responses = append(responses, responseSlice(response)...)
		if err != nil {
			return responses, err
		}
	}
	followups, err := pullRequestFollowups(ctx, rt, number, rec)
	return append(responses, followups...), err
}

// pullRequestFollowups sends the optional issue-metadata PATCH then the
// optional reviewers POST.
func pullRequestFollowups(ctx context.Context, rt *engine.Runtime, number int, rec connectors.Record) ([]*connsdk.Response, error) {
	var responses []*connsdk.Response
	if meta := selectFields(rec, metaFields); len(meta) > 0 {
		path := fmt.Sprintf("%s/issues/%d", repoPath(rt), number)
		requester, err := githubWriteRequester(rt, http.MethodPatch, githubIssueDeclaredPath)
		if err != nil {
			return responses, err
		}
		response, err := requester.Do(ctx, http.MethodPatch, path, nil, meta)
		responses = append(responses, responseSlice(response)...)
		if err != nil {
			return responses, err
		}
	}
	reviewers := selectFields(rec, reviewerFields)
	if len(reviewers) == 0 {
		return responses, nil
	}
	path := fmt.Sprintf("%s/pulls/%d/requested_reviewers", repoPath(rt), number)
	requester, err := githubWriteRequester(rt, http.MethodPost, githubRequestedReviewersDeclaredPath)
	if err != nil {
		return responses, err
	}
	response, err := requester.Do(ctx, http.MethodPost, path, nil, reviewers)
	return append(responses, responseSlice(response)...), err
}

func postComment(ctx context.Context, rt *engine.Runtime, resource string, number int, body string) ([]*connsdk.Response, error) {
	if resource != "issues" {
		return nil, fmt.Errorf("github write hook has no declared comment route for resource %q", resource)
	}
	path := fmt.Sprintf("%s/%s/%d/comments", repoPath(rt), resource, number)
	requester, err := githubWriteRequester(rt, http.MethodPost, githubIssueCommentDeclaredPath)
	if err != nil {
		return nil, err
	}
	response, err := requester.Do(ctx, http.MethodPost, path, nil, map[string]any{"body": body})
	return responseSlice(response), err
}

func responseSlice(response *connsdk.Response) []*connsdk.Response {
	if response == nil {
		return nil
	}
	return []*connsdk.Response{response}
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
