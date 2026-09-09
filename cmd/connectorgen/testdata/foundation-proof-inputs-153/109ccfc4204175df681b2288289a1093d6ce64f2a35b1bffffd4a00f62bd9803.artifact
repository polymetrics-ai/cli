package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
)

// TestHelpScoutV3DirectReadsUseTheirDeclaredRoute exercises the shipped Help
// Scout definition through the actual operation direct-read executor. The
// connection's v2 base deliberately points at the fixture's /v2 path: a v3
// declaration must select its own declared route and reach /v3, never /v2/v3.
func TestHelpScoutV3DirectReadsUseTheirDeclaredRoute(t *testing.T) {
	t.Helper()

	var (
		mu   sync.Mutex
		seen []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.URL.Path)
		mu.Unlock()
		if r.URL.Path == "/v2/v3" || len(r.URL.Path) >= len("/v2/v3/") && r.URL.Path[:len("/v2/v3/")] == "/v2/v3/" {
			t.Fatalf("request path = %q, route was incorrectly joined to v2 base", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"fixture"}`))
	}))
	t.Cleanup(srv.Close)

	bundle, err := Load(defs.FS, "help-scout")
	if err != nil {
		t.Fatalf("Load(help-scout): %v", err)
	}
	// Fixture transport replaces authentication only; operations, source URLs,
	// canonical endpoint mapping, and the configured-v2-base shape remain the
	// real shipped Help Scout definition.
	bundle.HTTP.Auth = nil

	tests := []struct {
		name       string
		operation  string
		pathParams map[string]string
		wantPath   string
	}{
		{name: "conversation", operation: "help-scout.v3_get_conversation", pathParams: map[string]string{"conversationId": "conversation-1"}, wantPath: "/v3/conversations/conversation-1"},
		{name: "conversation threads", operation: "help-scout.v3_list_conversation_threads", pathParams: map[string]string{"conversationId": "conversation-1"}, wantPath: "/v3/conversations/conversation-1/threads"},
		{name: "customers", operation: "help-scout.v3_list_customers", wantPath: "/v3/customers"},
		{name: "system users", operation: "help-scout.v3_list_system_users", wantPath: "/v3/system-users"},
		{name: "system user", operation: "help-scout.v3_get_system_user", pathParams: map[string]string{"systemUserId": "system-user-1"}, wantPath: "/v3/system-users/system-user-1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation:  test.operation,
				PathParams: test.pathParams,
				Config: connectors.RuntimeConfig{Config: map[string]string{
					"base_url": srv.URL + "/v2",
				}},
			}, nil)
			if err != nil {
				t.Fatalf("OperationDirectRead(%q): %v", test.operation, err)
			}
		})
	}

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != len(tests) {
		t.Fatalf("provider requests = %d, want %d (%v)", len(seen), len(tests), seen)
	}
	for i, test := range tests {
		if seen[i] != test.wantPath {
			t.Fatalf("request %d path = %q, want %q", i, seen[i], test.wantPath)
		}
	}
}
