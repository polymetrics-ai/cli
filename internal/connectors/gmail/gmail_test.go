package gmail_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gmail"
)

// liveConfig builds a RuntimeConfig pointed at an httptest server with the OAuth
// token endpoint overridden so the refresh-token exchange stays local. Secrets
// use the flattened last-segment names the runtime resolves the dotted
// credentials.* fields into.
func liveConfig(baseURL, tokenURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  baseURL,
			"token_url": tokenURL,
		},
		Secrets: map[string]string{
			"client_id":            "client-123",
			"client_secret":        "secret-xyz",
			"client_refresh_token": "refresh-abc",
		},
	}
}

// TestReadMessagesAuthenticatesAndPaginates is the red-first test: it asserts the
// connector exchanges the refresh token for an access token, sends it as a Bearer
// header to the Gmail API, follows pageToken/nextPageToken cursor pagination
// across two pages of messages[], and maps records.
func TestReadMessagesAuthenticatesAndPaginates(t *testing.T) {
	var (
		sawAuth       string
		tokenForm     string
		tokenRequests int
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++
		_ = r.ParseForm()
		tokenForm = r.Form.Encode()
		_, _ = w.Write([]byte(`{"access_token":"ACCESS_TOKEN_1","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"messages":[{"id":"msg_1","threadId":"thr_1"},{"id":"msg_2","threadId":"thr_1"}],"nextPageToken":"PAGE2","resultSizeEstimate":3}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"messages":[{"id":"msg_3","threadId":"thr_2"}],"resultSizeEstimate":3}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := gmail.New()
	cfg := liveConfig(srv.URL+"/v1", srv.URL+"/oauth/token")

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ACCESS_TOKEN_1" {
		t.Fatalf("Authorization = %q, want Bearer ACCESS_TOKEN_1", sawAuth)
	}
	if tokenRequests == 0 {
		t.Fatal("expected a refresh-token exchange against the token endpoint")
	}
	if !strings.Contains(tokenForm, "grant_type=refresh_token") || !strings.Contains(tokenForm, "refresh_token=refresh-abc") {
		t.Fatalf("token form = %q, want refresh_token grant", tokenForm)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["id"] != "msg_1" || got[0]["thread_id"] != "thr_1" {
		t.Fatalf("record[0] = %+v, want mapped msg_1/thr_1", got[0])
	}
	if got[2]["id"] != "msg_3" {
		t.Fatalf("record[2] id = %v, want msg_3", got[2]["id"])
	}
}

// TestReadIncludesSpamTrashAndStartDate asserts the optional include_spam_and_trash
// config is sent as includeSpamTrash and start_date becomes a Gmail q filter.
func TestReadIncludesSpamTrashAndStartDate(t *testing.T) {
	var (
		sawSpam string
		sawQ    string
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/users/me/threads", func(w http.ResponseWriter, r *http.Request) {
		sawSpam = r.URL.Query().Get("includeSpamTrash")
		sawQ = r.URL.Query().Get("q")
		_, _ = w.Write([]byte(`{"threads":[{"id":"thr_1","snippet":"hello","historyId":"42"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := liveConfig(srv.URL+"/v1", srv.URL+"/oauth/token")
	cfg.Config["include_spam_and_trash"] = "true"
	cfg.Config["start_date"] = "2024-01-02T00:00:00Z"

	c := gmail.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "threads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawSpam != "true" {
		t.Fatalf("includeSpamTrash = %q, want true", sawSpam)
	}
	if !strings.Contains(sawQ, "after:") {
		t.Fatalf("q = %q, want an after: filter derived from start_date", sawQ)
	}
	if len(got) != 1 || got[0]["id"] != "thr_1" {
		t.Fatalf("records = %+v, want one thread thr_1", got)
	}
}

// TestReadLabelsUnpaginated asserts the labels stream reads the single-page
// labels[] array and maps its records.
func TestReadLabelsUnpaginated(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/users/me/labels", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"labels":[{"id":"INBOX","name":"INBOX","type":"system"},{"id":"Label_1","name":"Work","type":"user"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := liveConfig(srv.URL+"/v1", srv.URL+"/oauth/token")
	c := gmail.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "labels", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 labels", len(got))
	}
	if got[0]["id"] != "INBOX" || got[0]["type"] != "system" {
		t.Fatalf("record[0] = %+v, want mapped INBOX/system", got[0])
	}
}

// TestFixtureModeNoNetwork confirms credential-free conformance: fixture mode
// emits deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := gmail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	for _, stream := range []string{"messages", "threads", "drafts", "labels"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(fixture, %s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCatalogStreams(t *testing.T) {
	c := gmail.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want at least 3", len(cat.Streams))
	}
	want := map[string]bool{"messages": false, "threads": false, "drafts": false, "labels": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = gmail.New() // ensure init ran
	caps := gmail.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities Write = true, want read-only connector")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("gmail"); !ok {
		t.Fatal("registry did not resolve gmail (self-registration)")
	}
}
