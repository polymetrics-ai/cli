// Package googleclassroom implements the native pm Google Classroom source
// connector. It follows the declarative-HTTP template (see internal/connectors/
// stripe): a thin package that composes the connsdk toolkit (Requester +
// RecordsAt extraction) with Google Classroom stream definitions, endpoints, and
// an OAuth2 refresh-token authenticator.
//
// connector exposes Check/Catalog/Read but no Write.
package googleclassroom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	registryName = "google-classroom"

	classroomDefaultBaseURL  = "https://classroom.googleapis.com"
	classroomDefaultTokenURL = "https://oauth2.googleapis.com/token"
	classroomDefaultPageSize = 100
	classroomMaxPageSize     = 1000
	classroomUserAgent       = "polymetrics-go-cli"
	// classroomFixtureUpdate is the deterministic updateTime used by fixture-mode
	// records.
	classroomFixtureUpdate = "2026-01-01T00:00:00Z"
)

// New returns the Google Classroom connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google Classroom source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Google Classroom",
		IntegrationType: "api",
		Description:     "Reads Google Classroom courses, teachers, students, course work, and announcements through the Classroom REST API using an OAuth2 refresh token.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Google
// Classroom. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := classroomBaseURL(cfg); err != nil {
		return err
	}
	if err := requireOAuthSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the courses list confirms the token exchange, auth, and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v1/courses", url.Values{"pageSize": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check google-classroom: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: classroomStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (full sync).
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "courses"
	}
	endpoint, ok := classroomStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("google-classroom stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := classroomPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := classroomMaxPages(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.nested {
		return c.harvest(ctx, r, endpoint, "", pageSize, maxPages, emit)
	}

	// Course-nested streams require the set of course ids first.
	courseIDs, err := c.listCourseIDs(ctx, r, pageSize, maxPages)
	if err != nil {
		return err
	}
	for _, courseID := range courseIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := c.harvest(ctx, r, endpoint, courseID, pageSize, maxPages, emit); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives Google Classroom's pageToken/nextPageToken pagination for a
// single endpoint. courseID is empty for top-level endpoints and the course id
// for nested ones. There is no body-token paginator in connsdk that injects the
// courseId into both the path and the records, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, courseID string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.pathTemplate
	if endpoint.nested {
		path = fmt.Sprintf(endpoint.pathTemplate, url.PathEscape(courseID))
	}

	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("pageSize", strconv.Itoa(pageSize))
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read google-classroom %s: %w", endpoint.recordsKey, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode google-classroom %s page: %w", endpoint.recordsKey, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if endpoint.nested && courseID != "" {
				item["courseId"] = courseID
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode google-classroom %s nextPageToken: %w", endpoint.recordsKey, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// listCourseIDs collects the course ids that nested streams iterate over.
func (c Connector) listCourseIDs(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int) ([]string, error) {
	var ids []string
	endpoint := classroomStreamEndpoints["courses"]
	err := c.harvest(ctx, r, endpoint, "", pageSize, maxPages, func(rec connectors.Record) error {
		if id := stringField(rec, "id"); id != "" {
			ids = append(ids, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors
// stripe's fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := classroomStreamEndpoints[stream]
	const fixtureCourse = "course_fixture_1"
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"courseId":      fixtureCourse,
			"userId":        fmt.Sprintf("user_fixture_%d", i),
			"name":          fmt.Sprintf("Fixture %s %d", stream, i),
			"title":         fmt.Sprintf("Fixture %s %d", stream, i),
			"text":          fmt.Sprintf("Fixture announcement %d", i),
			"section":       "A",
			"ownerId":       "owner_fixture_1",
			"courseState":   "ACTIVE",
			"state":         "PUBLISHED",
			"creatorUserId": "owner_fixture_1",
			"workType":      "ASSIGNMENT",
			"creationTime":  classroomFixtureUpdate,
			"updateTime":    classroomFixtureUpdate,
			"alternateLink": "https://classroom.google.com/fixture",
			"profile": map[string]any{
				"id":           fmt.Sprintf("user_fixture_%d", i),
				"emailAddress": fmt.Sprintf("fixture+%d@example.com", i),
				"name":         map[string]any{"fullName": fmt.Sprintf("Fixture User %d", i)},
			},
		}
		record := endpoint.mapRecord(item)
		if record["id"] == nil {
			record["id"] = item["id"]
		}
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the OAuth2 refresh-token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := classroomBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireOAuthSecrets(cfg); err != nil {
		return nil, err
	}
	auth := &refreshTokenAuth{
		tokenURL:     classroomTokenURL(cfg),
		clientID:     cfg.Secrets["client_id"],
		clientSecret: cfg.Secrets["client_secret"],
		refreshToken: cfg.Secrets["client_refresh_token"],
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: classroomUserAgent,
	}, nil
}

// refreshTokenAuth implements connsdk.Authenticator using the OAuth2 refresh-token
// grant. It exchanges the long-lived refresh token for a short-lived access token
// and caches it until shortly before expiry. It never logs secret values.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("oauth2: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("oauth2: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth2: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("oauth2: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("oauth2: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = time.Now().Add(ttl)
	return a.token, nil
}

// requireOAuthSecrets confirms all three OAuth credential fields are present.
func requireOAuthSecrets(cfg connectors.RuntimeConfig) error {
	for _, field := range []string{"client_id", "client_secret", "client_refresh_token"} {
		if cfg.Secrets == nil || strings.TrimSpace(cfg.Secrets[field]) == "" {
			return fmt.Errorf("google-classroom connector requires secret %s", field)
		}
	}
	return nil
}

// classroomBaseURL resolves and validates the base URL. The default is
// classroom.googleapis.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func classroomBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return classroomDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("google-classroom config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-classroom config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("google-classroom config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// classroomTokenURL resolves the OAuth2 token endpoint, allowing a test override.
func classroomTokenURL(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["token_url"]); v != "" {
		return v
	}
	return classroomDefaultTokenURL
}

func classroomPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return classroomDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-classroom config page_size must be an integer: %w", err)
	}
	if value < 1 || value > classroomMaxPageSize {
		return 0, fmt.Errorf("google-classroom config page_size must be between 1 and %d", classroomMaxPageSize)
	}
	return value, nil
}

func classroomMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-classroom config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("google-classroom config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Write is unsupported: Google Classroom is exposed as a read-only source.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
