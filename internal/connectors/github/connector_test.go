package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestGithubCheckSendsAuthAndValidatesRepository(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/widgets" {
			t.Fatalf("path = %s, want /repos/acme/widgets", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Fatalf("Accept = %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("X-GitHub-Api-Version") == "" {
			t.Fatal("missing GitHub API version header")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"full_name": "acme/widgets"})
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"repository": "acme/widgets", "base_url": server.URL},
		Secrets: map[string]string{"personalAccessToken": "secret-token"},
	}
	if err := (Connector{}).Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("Authorization = %q, want Bearer secret-token", gotAuth)
	}
}

func TestGithubReadIssuesFiltersPullRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/widgets/issues" {
			t.Fatalf("path = %s, want /repos/acme/widgets/issues", r.URL.Path)
		}
		if got := r.URL.Query().Get("state"); got != "all" {
			t.Fatalf("state = %q, want all", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "100" {
			t.Fatalf("per_page = %q, want 100", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 1, "node_id": "I_1", "number": 10, "state": "open", "title": "real issue",
				"html_url": "https://github.test/acme/widgets/issues/10", "user": map[string]any{"login": "ada", "id": 7},
				"comments": 2, "locked": false, "labels": []any{"bug"}, "assignees": []any{},
				"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z",
			},
			{
				"id": 2, "number": 11, "state": "open", "title": "pr from issues endpoint",
				"pull_request": map[string]any{"url": "https://api.github.test/pulls/11"},
			},
		})
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"repository": "acme/widgets", "base_url": server.URL}}
	var records []connectors.Record
	err := (Connector{}).Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	if records[0]["title"] != "real issue" {
		t.Fatalf("title = %v", records[0]["title"])
	}
	if records[0]["repository"] != "acme/widgets" {
		t.Fatalf("repository = %v", records[0]["repository"])
	}
	if records[0]["labels_count"] != 1 {
		t.Fatalf("labels_count = %v, want 1", records[0]["labels_count"])
	}
}

func TestGithubReadPullRequestsNormalizesBranchRefs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/widgets/pulls" {
			t.Fatalf("path = %s, want /repos/acme/widgets/pulls", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 3, "node_id": "PR_3", "number": 12, "state": "open", "title": "add feature",
				"html_url": "https://github.test/acme/widgets/pull/12", "user": map[string]any{"login": "grace", "id": 8},
				"comments": 1, "locked": false, "draft": true, "merge_commit_sha": "abc123",
				"base":       map[string]any{"ref": "main", "sha": "base-sha"},
				"head":       map[string]any{"ref": "feature", "sha": "head-sha"},
				"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"repository": "acme/widgets", "base_url": server.URL}}
	var records []connectors.Record
	err := (Connector{}).Read(context.Background(), connectors.ReadRequest{Stream: "pull_requests", Config: cfg}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	if records[0]["base_ref"] != "main" || records[0]["head_ref"] != "feature" {
		t.Fatalf("refs = base %v head %v", records[0]["base_ref"], records[0]["head_ref"])
	}
	if records[0]["draft"] != true {
		t.Fatalf("draft = %v, want true", records[0]["draft"])
	}
}

func TestGithubReadPullRequestsSupportsUnlimitedPages(t *testing.T) {
	requestedPages := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/widgets/pulls" {
			t.Fatalf("path = %s, want /repos/acme/widgets/pulls", r.URL.Path)
		}
		requestedPages = append(requestedPages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "1":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 1, "number": 1, "title": "first", "user": map[string]any{"login": "ada"}},
				{"id": 2, "number": 2, "title": "second", "user": map[string]any{"login": "grace"}},
			})
		case "2":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 3, "number": 3, "title": "third", "user": map[string]any{"login": "katherine"}},
			})
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{
		"repository": "acme/widgets",
		"base_url":   server.URL,
		"per_page":   "2",
		"max_pages":  "0",
	}}
	var records []connectors.Record
	err := (Connector{}).Read(context.Background(), connectors.ReadRequest{Stream: "pull_requests", Config: cfg}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("records len = %d, want 3", len(records))
	}
	if got, want := requestedPages, []string{"1", "2"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("requested pages = %v, want %v", got, want)
	}
}

func TestGithubWriteCreateIssuePostsMappedFields(t *testing.T) {
	var gotAuth string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/repos/acme/widgets/issues" {
			t.Fatalf("path = %s, want /repos/acme/widgets/issues", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Fatalf("Accept = %q", r.Header.Get("Accept"))
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"number": 23, "html_url": "https://github.test/acme/widgets/issues/23"})
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"repository": "acme/widgets", "base_url": server.URL},
		Secrets: map[string]string{"token": "secret-token"},
	}
	result, err := (Connector{}).Write(context.Background(), connectors.WriteRequest{Action: "create_issue", Config: cfg}, []connectors.Record{{
		"title":     "Bug from warehouse",
		"body":      "Generated by reverse ETL",
		"labels":    "bug,help wanted",
		"assignees": []any{"ada", "grace"},
		"milestone": 2,
	}})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("Authorization = %q, want Bearer secret-token", gotAuth)
	}
	if payload["title"] != "Bug from warehouse" || payload["body"] != "Generated by reverse ETL" {
		t.Fatalf("payload title/body = %+v", payload)
	}
	if got := fmt.Sprint(payload["labels"]); got != "[bug help wanted]" {
		t.Fatalf("labels = %s", got)
	}
	if got := fmt.Sprint(payload["assignees"]); got != "[ada grace]" {
		t.Fatalf("assignees = %s", got)
	}
	if payload["milestone"].(float64) != 2 {
		t.Fatalf("milestone = %v", payload["milestone"])
	}
}

func TestGithubWriteCreatePullRequestAddsMetadataAndReviewers(t *testing.T) {
	type seenRequest struct {
		Method string
		Path   string
		Body   map[string]any
	}
	var seen []seenRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body for %s %s: %v", r.Method, r.URL.Path, err)
		}
		seen = append(seen, seenRequest{Method: r.Method, Path: r.URL.Path, Body: body})
		switch r.URL.Path {
		case "/repos/acme/widgets/pulls":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 42, "html_url": "https://github.test/acme/widgets/pull/42"})
		case "/repos/acme/widgets/issues/42":
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 42})
		case "/repos/acme/widgets/pulls/42/requested_reviewers":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 42})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"repository": "acme/widgets", "base_url": server.URL},
		Secrets: map[string]string{"token": "secret-token"},
	}
	result, err := (Connector{}).Write(context.Background(), connectors.WriteRequest{Action: "create_pull_request", Config: cfg}, []connectors.Record{{
		"title":                 "Ship feature",
		"body":                  "Ready for review",
		"head":                  "feature-branch",
		"base":                  "main",
		"draft":                 true,
		"maintainer_can_modify": false,
		"labels":                []any{"enhancement", "agentic"},
		"assignees":             "ada",
		"milestone":             7,
		"reviewers":             "grace,katherine",
		"team_reviewers":        []any{"platform"},
	}})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if len(seen) != 3 {
		t.Fatalf("request count = %d, want 3: %+v", len(seen), seen)
	}
	if seen[0].Method != http.MethodPost || seen[0].Path != "/repos/acme/widgets/pulls" {
		t.Fatalf("first request = %+v", seen[0])
	}
	if seen[0].Body["title"] != "Ship feature" || seen[0].Body["head"] != "feature-branch" || seen[0].Body["base"] != "main" {
		t.Fatalf("create PR body = %+v", seen[0].Body)
	}
	if seen[0].Body["draft"] != true || seen[0].Body["maintainer_can_modify"] != false {
		t.Fatalf("create PR booleans = %+v", seen[0].Body)
	}
	if seen[1].Method != http.MethodPatch || seen[1].Path != "/repos/acme/widgets/issues/42" {
		t.Fatalf("metadata request = %+v", seen[1])
	}
	if fmt.Sprint(seen[1].Body["labels"]) != "[enhancement agentic]" || fmt.Sprint(seen[1].Body["assignees"]) != "[ada]" {
		t.Fatalf("metadata body = %+v", seen[1].Body)
	}
	if seen[2].Method != http.MethodPost || seen[2].Path != "/repos/acme/widgets/pulls/42/requested_reviewers" {
		t.Fatalf("reviewer request = %+v", seen[2])
	}
	if fmt.Sprint(seen[2].Body["reviewers"]) != "[grace katherine]" || fmt.Sprint(seen[2].Body["team_reviewers"]) != "[platform]" {
		t.Fatalf("reviewer body = %+v", seen[2].Body)
	}
}

func TestGithubWriteCommentAndMergeActions(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		record     connectors.Record
		wantMethod string
		wantPath   string
	}{
		{
			name:       "comment pr through issue comments",
			action:     "comment_pr",
			record:     connectors.Record{"number": 12, "body": "Looks good"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/issues/12/comments",
		},
		{
			name:       "merge pr",
			action:     "merge_pr",
			record:     connectors.Record{"pull_number": 13, "merge_method": "squash", "commit_title": "Squash merge"},
			wantMethod: http.MethodPut,
			wantPath:   "/repos/acme/widgets/pulls/13/merge",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.wantMethod || r.URL.Path != tt.wantPath {
					t.Fatalf("%s %s, want %s %s", r.Method, r.URL.Path, tt.wantMethod, tt.wantPath)
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			}))
			defer server.Close()

			cfg := connectors.RuntimeConfig{
				Config:  map[string]string{"repository": "acme/widgets", "base_url": server.URL},
				Secrets: map[string]string{"token": "secret-token"},
			}
			result, err := (Connector{}).Write(context.Background(), connectors.WriteRequest{Action: tt.action, Config: cfg}, []connectors.Record{tt.record})
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
				t.Fatalf("result = %+v", result)
			}
		})
	}
}

func TestGithubValidateWriteRejectsUnsafePlansBeforeNetwork(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"repository": "acme/widgets", "base_url": server.URL},
		Secrets: map[string]string{"token": "secret-token"},
	}
	err := (Connector{}).ValidateWrite(context.Background(), connectors.WriteRequest{Action: "create_pull_request", Config: cfg}, []connectors.Record{{
		"title": "Missing branches",
	}})
	if err == nil || !strings.Contains(err.Error(), "head") || !strings.Contains(err.Error(), "base") {
		t.Fatalf("ValidateWrite() error = %v, want missing head/base", err)
	}
	if called {
		t.Fatal("validation made an HTTP request")
	}

	cfg.Secrets = nil
	err = (Connector{}).ValidateWrite(context.Background(), connectors.WriteRequest{Action: "create_issue", Config: cfg}, []connectors.Record{{
		"title": "Needs token",
	}})
	if err == nil || !strings.Contains(err.Error(), "requires token") {
		t.Fatalf("ValidateWrite() error = %v, want token requirement", err)
	}
}

func TestGithubRejectsInvalidRepository(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"repository": "not/a/repo"}}
	if err := (Connector{}).Check(context.Background(), cfg); err == nil {
		t.Fatal("Check() error = nil, want invalid repository error")
	}
}

func TestGithubAuthTokenAliases(t *testing.T) {
	tests := []struct {
		name    string
		secrets map[string]string
	}{
		{name: "token", secrets: map[string]string{"token": "secret-token"}},
		{name: "personal access token", secrets: map[string]string{"personalAccessToken": "secret-token"}},
		{name: "oauth token", secrets: map[string]string{"oauthToken": "secret-token"}},
		{name: "installation token", secrets: map[string]string{"installationToken": "secret-token"}},
		{name: "actions token", secrets: map[string]string{"githubToken": "secret-token"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth = r.Header.Get("Authorization")
				_ = json.NewEncoder(w).Encode(map[string]any{"full_name": "acme/widgets"})
			}))
			defer server.Close()

			cfg := connectors.RuntimeConfig{
				Config:  map[string]string{"repository": "acme/widgets", "base_url": server.URL},
				Secrets: tt.secrets,
			}
			if err := (Connector{}).Check(context.Background(), cfg); err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if gotAuth != "Bearer secret-token" {
				t.Fatalf("Authorization = %q, want Bearer secret-token", gotAuth)
			}
		})
	}
}

func TestGithubAppInstallationAuthExchangesJWTAndUsesInstallationToken(t *testing.T) {
	privateKey := mustTestRSAKey(t)
	privatePEM := mustTestPrivateKeyPEM(t, privateKey)
	var gotJWT string
	var gotRepoAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/installations/99/access_tokens":
			gotJWT = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if gotJWT == "" || strings.Count(gotJWT, ".") != 2 {
				t.Fatalf("Authorization = %q, want Bearer JWT", r.Header.Get("Authorization"))
			}
			var request map[string]any
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode installation token request: %v", err)
			}
			if got := request["repositories"]; got == nil {
				t.Fatalf("installation token request missing repositories: %+v", request)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "installation-token", "expires_at": "2026-06-25T12:00:00Z"})
		case "/repos/acme/widgets":
			gotRepoAuth = r.Header.Get("Authorization")
			_ = json.NewEncoder(w).Encode(map[string]any{"full_name": "acme/widgets"})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"repository":                "acme/widgets",
			"base_url":                  server.URL,
			"auth_type":                 "github_app",
			"app_id":                    "12345",
			"installation_id":           "99",
			"installation_repositories": "widgets",
		},
		Secrets: map[string]string{"private_key": privatePEM},
	}
	if err := (Connector{}).Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if gotJWT == "" {
		t.Fatal("GitHub App JWT was not sent")
	}
	if gotRepoAuth != "Bearer installation-token" {
		t.Fatalf("repository Authorization = %q, want Bearer installation-token", gotRepoAuth)
	}
	assertJWTIssuer(t, gotJWT, "12345")
}

func TestGithubAppPrivateKeyBase64AndValidateWriteWithoutNetwork(t *testing.T) {
	privatePEM := mustTestPrivateKeyPEM(t, mustTestRSAKey(t))
	encoded := base64.StdEncoding.EncodeToString([]byte(privatePEM))
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"repository":      "acme/widgets",
			"base_url":        server.URL,
			"auth_type":       "github_app",
			"app_id":          "12345",
			"installation_id": "99",
		},
		Secrets: map[string]string{"private_key_base64": encoded},
	}
	err := (Connector{}).ValidateWrite(context.Background(), connectors.WriteRequest{Action: "create_issue", Config: cfg}, []connectors.Record{{"title": "from app"}})
	if err != nil {
		t.Fatalf("ValidateWrite() error = %v", err)
	}
	if called {
		t.Fatal("ValidateWrite made a network request")
	}
}

func mustTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	return key
}

func mustTestPrivateKeyPEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return string(pem.EncodeToMemory(block))
}

func assertJWTIssuer(t *testing.T, jwt string, want string) {
	t.Helper()
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		t.Fatalf("jwt parts = %d, want 3", len(parts))
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode jwt payload: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("unmarshal jwt payload: %v", err)
	}
	if payload["iss"] != want {
		t.Fatalf("jwt iss = %v, want %s", payload["iss"], want)
	}
	if payload["iat"] == nil || payload["exp"] == nil {
		t.Fatalf("jwt missing iat/exp: %+v", payload)
	}
}
