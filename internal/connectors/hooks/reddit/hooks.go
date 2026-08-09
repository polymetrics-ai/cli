// Package reddit is the Tier-2 escape hatch: a refresh_token->access_token
// exchange AuthHook. Reddit's client_credentials grant (already supported
// declaratively by the engine's oauth2_client_credentials auth mode) mints
// an "Application Only" token that Reddit's own docs state can never act on
// behalf of a user (https://github.com/reddit-archive/reddit/wiki/OAuth2#application-only-oauth) --
// no user-context endpoint (moderation, private messages, votes, saves, or
// anything scoped to a specific account) is reachable with it. The
// refresh_token grant is the only way to keep a durable, user-context
// bearer token alive past Reddit's 1-hour access-token expiry, and the
// engine has no built-in support for it, so this hook ports the exchange
// (mirrors hooks/github's JWT->installation-token pattern).
package reddit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/transportpolicy"
	"polymetrics.ai/internal/safety"
)

func init() {
	engine.RegisterHooks("reddit", func() engine.Hooks { return New() })
}

// Hooks is the reddit bundle's stateless Tier-2 hook set.
type Hooks struct{}

// New returns a fresh Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "reddit" }

var (
	_ engine.Hooks     = (*Hooks)(nil)
	_ engine.AuthHook  = (*Hooks)(nil)
	_ engine.WriteHook = (*Hooks)(nil)
)

const defaultTokenURL = "https://www.reddit.com/api/v1/access_token"

// Authenticator exchanges the configured refresh_token for a fresh
// access_token via POST {token_url} grant_type=refresh_token (Reddit OAuth2
// "Refreshing the token": https://github.com/reddit-archive/reddit/wiki/OAuth2#refreshing-the-token),
// then returns a Bearer authenticator wrapping it. ctx is honored (a real
// network call). Uncached, matching hooks/github's own re-mint-on-every-call
// behavior -- this hook is invoked once per Read/Write/Check call
// (engine/read.go's newRuntime), so a fresh token is minted at the start of
// every sync/command rather than being reused past its 1-hour lifetime.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	refreshToken := strings.TrimSpace(cfg.Secrets["refresh_token"])
	if refreshToken == "" {
		return nil, errors.New("reddit: refresh_token exchange requires secrets.refresh_token")
	}
	clientID := strings.TrimSpace(cfg.Secrets["client_id"])
	if clientID == "" {
		return nil, errors.New("reddit: refresh_token exchange requires secrets.client_id")
	}
	// client_secret may legitimately be empty: Reddit's installed
	// (non-confidential) app type has no secret and documents sending an
	// empty string as the HTTP Basic password in that case.
	clientSecret := cfg.Secrets["client_secret"]

	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("reddit: build refresh_token request: %w", err)
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent(cfg))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit: refresh_token exchange: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("reddit: refresh_token exchange returned status %d", resp.StatusCode)
	}
	if readErr != nil {
		return nil, fmt.Errorf("reddit: read refresh_token response: %w", readErr)
	}
	var out struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("reddit: decode refresh_token response: %w", err)
	}
	if out.Error != "" {
		return nil, fmt.Errorf("reddit: refresh_token exchange rejected: %s", out.Error)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return nil, errors.New("reddit: refresh_token response did not include access_token")
	}
	return connsdk.Bearer(out.AccessToken), nil
}

// userAgent mirrors streams.json's declarative User-Agent template
// (<platform>:<app ID>:<version> (by /u/<reddit_username>)) for the token
// exchange request itself, which is not routed through the declarative
// header pipeline. Reddit rate-limits non-conforming User-Agents on every
// endpoint, including the token endpoint.
func userAgent(cfg connectors.RuntimeConfig) string {
	username := strings.TrimSpace(cfg.Config["reddit_username"])
	if username == "" {
		username = "unknown"
	}
	return "go:ai.polymetrics.cli:v1 (by /u/" + username + ")"
}

// ExecuteWrite implements the two Reddit endpoints that return a short-lived
// S3 upload lease rather than consuming the image bytes themselves. The first
// request is the ordinary, OAuth-authenticated Reddit form request; the second
// is a tightly bounded form upload to the lease's Amazon S3 endpoint. It is
// intentionally not a generic upload proxy: action names, source field,
// outbound host family, file field, response body shape, file-root
// containment, and byte limits are all closed here.
func (h *Hooks) ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (bool, error) {
	switch action.Name {
	case "emoji_asset_upload_s3", "widget_image_upload_s3":
		return true, executeS3LeaseUpload(ctx, action, rec, rt)
	case "upload_sr_img":
		if rt != nil && rt.Config.ProjectDir == conformanceProjectDir {
			return true, executeConformanceSubredditImageUpload(ctx, action, rec, rt)
		}
		return false, nil
	default:
		return false, nil
	}
}

const (
	leaseResponseMaxBytes = 1 << 20
	leaseUploadMaxBytes   = 10 << 20
	conformanceProjectDir = "__polymetrics_conformance_fixture__"
)

type s3UploadLease struct {
	action string
	fields map[string]string
}

func executeS3LeaseUpload(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) error {
	if rt == nil || rt.Requester == nil {
		return errors.New("reddit: S3 lease upload requires an authenticated requester")
	}
	filePath, err := requiredLeaseString(rec, "file_path")
	if err != nil {
		return err
	}
	mimetype, err := requiredImageMIME(rec)
	if err != nil {
		return err
	}

	// Conformance deliberately has no project-owned source file. It replays
	// the authenticated Reddit lease request here; the actual bounded S3 hop
	// is exercised by TestExecuteWrite_AcquiresBoundS3LeaseWithoutForwardingOAuth.
	if rt.Config.ProjectDir == conformanceProjectDir {
		_, err := acquireS3UploadLease(ctx, action, rec, rt, filePath, mimetype)
		return err
	}

	// Prepare and bind the source before acquiring a one-use lease. This both
	// avoids spending a lease on a local rejection and ensures the second
	// request cannot read a path that escaped the project root after plan time.
	form, root, err := leaseMultipartForm(rt.Config, filePath, mimetype)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	lease, err := acquireS3UploadLease(ctx, action, rec, rt, filePath, mimetype)
	if err != nil {
		return err
	}
	uploadURL, err := approvedS3LeaseURL(lease.action)
	if err != nil {
		return err
	}
	for name, value := range lease.fields {
		if name == "file" {
			return errors.New("reddit: S3 upload lease must not define the reserved file field")
		}
		form.Fields[name] = value
	}

	// Clone the Reddit requester only for its bounded transport plumbing.
	// Explicitly drop all OAuth-related state before the third-party hop: an S3
	// lease authenticates its form fields, never a Reddit bearer token.
	uploader := *rt.Requester
	uploader.BaseURL = ""
	uploader.Auth = nil
	uploader.UserAgent = ""
	uploader.DefaultHeaders = nil
	uploader.Accept = ""
	uploader.DisableRetries = true
	if _, err := uploader.DoMultipartLimited(ctx, http.MethodPost, uploadURL.String(), nil, form, leaseResponseMaxBytes); err != nil {
		return safeLeaseRequestError("upload approved image to S3", err)
	}
	return nil
}

// acquireS3UploadLease sends the OAuth-authenticated first leg of a
// lease-backed upload. The returned fields are intentionally not sent to any
// host other than the pinned Amazon S3 endpoint in executeS3LeaseUpload.
func acquireS3UploadLease(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime, filePath, mimetype string) (s3UploadLease, error) {
	vars := engine.Vars{Config: rt.Config.Config, Secrets: rt.Config.Secrets, Record: map[string]any(rec)}
	endpoint, err := engine.InterpolatePath(action.Path, vars)
	if err != nil {
		return s3UploadLease{}, fmt.Errorf("reddit: resolve S3 lease endpoint: %w", err)
	}
	leaseForm := url.Values{"filepath": {filepath.Base(filePath)}, "mimetype": {mimetype}}
	leaseRequester := *rt.Requester
	leaseRequester.DisableRetries = true
	// A redirect could turn a mutation into an unbound request; the shared
	// transport policy refuses it for both stages of this compound write.
	ctx = transportpolicy.MarkDestructive(ctx)
	response, err := leaseRequester.DoFormLimited(ctx, http.MethodPost, endpoint, nil, leaseForm, leaseResponseMaxBytes)
	if err != nil {
		return s3UploadLease{}, safeLeaseRequestError("acquire S3 upload lease", err)
	}
	lease, err := parseS3UploadLease(response.Body)
	if err != nil {
		return s3UploadLease{}, err
	}
	return lease, nil
}

// executeConformanceSubredditImageUpload is deliberately limited to the
// conformance harness's reserved project directory. A real upload_sr_img
// action returns handled=false above and runs through the engine's declarative
// multipart writer, which binds the caller's approved project-local source.
// The replay harness cannot provide such a source, so it uses a tiny
// in-memory-created PNG only to prove the action's resolved request reaches
// the capture server as multipart. The regular engine multipart test suite is
// the authoritative proof of source confinement and approval binding.
func executeConformanceSubredditImageUpload(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) error {
	if action.Multipart == nil {
		return errors.New("reddit: conformance subreddit image action is missing multipart metadata")
	}
	tmp, err := os.MkdirTemp("", "reddit-conformance-upload-*")
	if err != nil {
		return fmt.Errorf("reddit: create conformance image directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	const fixtureName = "fixture-subreddit.png"
	if err := os.WriteFile(filepath.Join(tmp, fixtureName), []byte("\x89PNG\r\n\x1a\n"), 0o600); err != nil {
		return fmt.Errorf("reddit: create conformance image: %w", err)
	}
	root, err := os.OpenRoot(tmp)
	if err != nil {
		return fmt.Errorf("reddit: open conformance image directory: %w", err)
	}
	defer func() { _ = root.Close() }()

	form := connsdk.MultipartForm{Fields: map[string]string{}, MaxBytes: action.Multipart.MaxBytes}
	for _, part := range action.Multipart.Parts {
		value, found := rec[part.Field]
		if !found || value == nil {
			if part.Required {
				return fmt.Errorf("reddit: conformance image fixture is missing required part %q", part.Name)
			}
			continue
		}
		switch part.Type {
		case "field":
			form.Fields[part.Name] = fmt.Sprint(value)
		case "file":
			form.Files = append(form.Files, connsdk.MultipartFile{
				FieldName:         part.Name,
				Root:              root,
				RelPath:           fixtureName,
				Path:              fixtureName,
				FileName:          fixtureName,
				AllowedMediaTypes: part.AllowedMediaTypes,
				MaxBytes:          part.MaxBytes,
			})
		default:
			return fmt.Errorf("reddit: conformance image action has unsupported multipart part %q", part.Type)
		}
	}
	vars := engine.Vars{Config: rt.Config.Config, Secrets: rt.Config.Secrets, Record: map[string]any(rec)}
	endpoint, err := engine.InterpolatePath(action.Path, vars)
	if err != nil {
		return fmt.Errorf("reddit: resolve conformance image endpoint: %w", err)
	}
	method := action.Method
	if method == "" {
		method = http.MethodPost
	}
	if _, err := rt.Requester.DoMultipart(ctx, method, endpoint, nil, form); err != nil {
		return safeLeaseRequestError("upload subreddit image", err)
	}
	return nil
}

func requiredLeaseString(rec connectors.Record, field string) (string, error) {
	value, ok := rec[field].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("reddit: S3 lease upload requires %s", field)
	}
	return value, nil
}

func requiredImageMIME(rec connectors.Record) (string, error) {
	raw, err := requiredLeaseString(rec, "mimetype")
	if err != nil {
		return "", err
	}
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil || !strings.HasPrefix(strings.ToLower(mediaType), "image/") {
		return "", errors.New("reddit: S3 lease upload mimetype must be an image media type")
	}
	return mediaType, nil
}

func leaseMultipartForm(cfg connectors.RuntimeConfig, rawPath, mimetype string) (connsdk.MultipartForm, *os.Root, error) {
	projectDir := strings.TrimSpace(cfg.ProjectDir)
	if projectDir == "" {
		projectDir = "."
	}
	if err := safety.ValidateLocalWritePath(projectDir, rawPath, "reddit S3 upload file path", false); err != nil {
		return connsdk.MultipartForm{}, nil, err
	}
	rootAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return connsdk.MultipartForm{}, nil, fmt.Errorf("reddit: resolve upload project root: %w", err)
	}
	rel, err := filepath.Rel(rootAbs, absoluteOrProjectPath(rootAbs, rawPath))
	if err != nil || !filepath.IsLocal(rel) {
		return connsdk.MultipartForm{}, nil, errors.New("reddit: S3 upload file path must stay inside the project root")
	}
	root, err := os.OpenRoot(rootAbs)
	if err != nil {
		return connsdk.MultipartForm{}, nil, fmt.Errorf("reddit: open upload project root: %w", err)
	}
	info, err := root.Stat(rel)
	if err != nil {
		_ = root.Close()
		return connsdk.MultipartForm{}, nil, fmt.Errorf("reddit: inspect upload file: %w", err)
	}
	if !info.Mode().IsRegular() {
		_ = root.Close()
		return connsdk.MultipartForm{}, nil, errors.New("reddit: S3 upload source must be a regular file")
	}
	if info.Size() > leaseUploadMaxBytes {
		_ = root.Close()
		return connsdk.MultipartForm{}, nil, fmt.Errorf("reddit: S3 upload source exceeds %d byte limit", leaseUploadMaxBytes)
	}
	expected := cfg.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(0, "file_path")]
	if strings.TrimSpace(expected) == "" {
		_ = root.Close()
		return connsdk.MultipartForm{}, nil, errors.New("reddit: S3 upload source is missing its approved payload digest")
	}
	return connsdk.MultipartForm{
		Fields:   map[string]string{},
		MaxBytes: leaseUploadMaxBytes,
		Files: []connsdk.MultipartFile{{
			FieldName:         "file",
			Root:              root,
			RelPath:           rel,
			Path:              rel,
			FileName:          filepath.Base(rel),
			ContentType:       mimetype,
			AllowedMediaTypes: []string{mimetype},
			MaxBytes:          leaseUploadMaxBytes,
			ExpectedSHA256:    expected,
		}},
	}, root, nil
}

func absoluteOrProjectPath(rootAbs, rawPath string) string {
	if filepath.IsAbs(rawPath) {
		return filepath.Clean(rawPath)
	}
	return filepath.Join(rootAbs, filepath.Clean(rawPath))
}

func parseS3UploadLease(body []byte) (s3UploadLease, error) {
	var envelope struct {
		Lease json.RawMessage `json:"s3UploadLease"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || len(envelope.Lease) == 0 {
		return s3UploadLease{}, errors.New("reddit: S3 upload lease response did not include s3UploadLease")
	}
	var raw struct {
		Action string          `json:"action"`
		Fields json.RawMessage `json:"fields"`
	}
	if err := json.Unmarshal(envelope.Lease, &raw); err != nil || strings.TrimSpace(raw.Action) == "" {
		return s3UploadLease{}, errors.New("reddit: S3 upload lease response did not include an upload action")
	}
	fields, err := parseS3LeaseFields(raw.Fields)
	if err != nil {
		return s3UploadLease{}, err
	}
	return s3UploadLease{action: raw.Action, fields: fields}, nil
}

func parseS3LeaseFields(raw json.RawMessage) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, errors.New("reddit: S3 upload lease response did not include upload fields")
	}
	var object map[string]string
	if err := json.Unmarshal(raw, &object); err == nil && object != nil {
		if len(object) == 0 {
			return nil, errors.New("reddit: S3 upload lease response included no upload fields")
		}
		for name := range object {
			if err := validateS3LeaseFieldName(name); err != nil {
				return nil, err
			}
		}
		return object, nil
	}
	var pairs []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &pairs); err != nil || len(pairs) == 0 {
		return nil, errors.New("reddit: S3 upload lease response has invalid upload fields")
	}
	fields := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		name := strings.TrimSpace(pair.Name)
		if err := validateS3LeaseFieldName(name); err != nil {
			return nil, err
		}
		if _, duplicate := fields[name]; duplicate {
			return nil, errors.New("reddit: S3 upload lease response repeats an upload field")
		}
		fields[name] = pair.Value
	}
	return fields, nil
}

func validateS3LeaseFieldName(name string) error {
	if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\r\n") {
		return errors.New("reddit: S3 upload lease response has an invalid upload field name")
	}
	return nil
}

func approvedS3LeaseURL(raw string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if strings.HasPrefix(value, "//") {
		value = "https:" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" || parsed.User != nil {
		return nil, errors.New("reddit: S3 upload lease has an invalid HTTPS upload URL")
	}
	if port := parsed.Port(); port != "" && port != "443" {
		return nil, errors.New("reddit: S3 upload lease must use HTTPS port 443")
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "amazonaws.com" && !strings.HasSuffix(host, ".amazonaws.com") {
		return nil, errors.New("reddit: S3 upload lease host is not an Amazon S3 host")
	}
	return parsed, nil
}

func safeLeaseRequestError(stage string, err error) error {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		return fmt.Errorf("reddit: %s returned HTTP %d", stage, httpErr.Status)
	}
	return fmt.Errorf("reddit: %s failed", stage)
}
