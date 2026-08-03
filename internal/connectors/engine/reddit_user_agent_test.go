package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
)

const (
	redditTestUsername  = "polymetrics_test_bot"
	redditWantUserAgent = "go:ai.polymetrics.cli:v1 (by /u/" + redditTestUsername + ")"
)

func loadRedditBundle(t *testing.T) Bundle {
	t.Helper()
	bundle, err := Load(defs.FS, "reddit")
	if err != nil {
		t.Fatalf("load reddit bundle: %v", err)
	}
	return bundle
}

// TestRedditUserAgentDeclaredAsTemplatedHeader pins WHERE the User-Agent is
// declared. Reddit mandates "<platform>:<app ID>:<version> (by /u/<reddit
// username>)" and drastically rate-limits non-conforming clients, so the
// operator identity has to come from config at runtime. base.user_agent is
// copied to connsdk.Requester.UserAgent verbatim and never interpolated,
// while base.headers values go through InterpolateHeader — moving the value
// back to base.user_agent would silently ship "{{ config.reddit_username }}"
// as literal text.
func TestRedditUserAgentDeclaredAsTemplatedHeader(t *testing.T) {
	bundle := loadRedditBundle(t)

	if bundle.HTTP.UserAgent != "" {
		t.Fatalf("base.user_agent = %q, want empty (the field is not interpolated; use base.headers)", bundle.HTTP.UserAgent)
	}
	got := bundle.HTTP.Headers["User-Agent"]
	if got != "go:ai.polymetrics.cli:v1 (by /u/{{ config.reddit_username }})" {
		t.Fatalf("base.headers[User-Agent] = %q, want the templated reddit-conforming value", got)
	}
}

// TestRedditUserAgentSentOnReadRequests asserts the resolved header actually
// reaches the wire: connsdk applies DefaultHeaders after Requester.UserAgent,
// so the templated entry must win on every outbound request, including
// paginated follow-ups.
func TestRedditUserAgentSentOnReadRequests(t *testing.T) {
	bundle := loadRedditBundle(t)

	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("User-Agent"))
		after := "t3_page2"
		if r.URL.Query().Get("after") != "" {
			after = ""
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"after": after,
				"children": []map[string]any{
					{"data": map[string]any{"id": "abc123", "name": "t3_abc123", "title": "fixture post"}},
				},
			},
		})
	}))
	defer server.Close()

	req := connectors.ReadRequest{
		Stream: "posts",
		Config: connectors.RuntimeConfig{
			Config: map[string]string{
				"base_url":        server.URL,
				"subreddit":       "golang",
				"reddit_username": redditTestUsername,
			},
			Secrets: map[string]string{"access_token": "synthetic-conformance-secret"},
		},
	}

	var records []connectors.Record
	err := Read(context.Background(), bundle, req, nil, func(rec connectors.Record) error {
		records = append(records, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2 (one per page)", len(records))
	}
	if len(seen) != 2 {
		t.Fatalf("requests = %d, want 2 (initial page plus cursor follow-up)", len(seen))
	}
	for i, ua := range seen {
		if ua != redditWantUserAgent {
			t.Fatalf("request %d User-Agent = %q, want %q", i, ua, redditWantUserAgent)
		}
	}
}

// TestRedditUserAgentFailsClosedWithoutUsername pins the fail-closed
// behavior that makes reddit_username a genuinely required field: because it
// is in spec.json's required[], an absent value is a hard header-resolution
// error rather than an omitted header, so no request is ever sent with a
// non-conforming (or missing) User-Agent.
func TestRedditUserAgentFailsClosedWithoutUsername(t *testing.T) {
	bundle := loadRedditBundle(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request to %s without reddit_username configured", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req := connectors.ReadRequest{
		Stream: "posts",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL, "subreddit": "golang"},
			Secrets: map[string]string{"access_token": "synthetic-conformance-secret"},
		},
	}

	err := Read(context.Background(), bundle, req, nil, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read succeeded without reddit_username, want a header-resolution error")
	}
	if !strings.Contains(err.Error(), "User-Agent") || !strings.Contains(err.Error(), "reddit_username") {
		t.Fatalf("Read error = %v, want it to name the User-Agent header and reddit_username", err)
	}
}
