package connsdk

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/telemetry"
)

func TestRequesterDoEmitsSecretSafeHTTPSpan(t *testing.T) {
	const marker = "pm_test_secret_token_http_span"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("api_key"); got != marker {
			t.Fatalf("api_key query = %q, want marker", got)
		}
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Fatalf("page query = %q, want 1", got)
		}
		w.Header().Set("X-Secret-Echo", marker)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, Directory: dir, RunID: "http-span"}, func(string) {})
	requester := &Requester{
		Client:     server.Client(),
		BaseURL:    server.URL,
		Auth:       APIKeyQuery("api_key", marker),
		MaxRetries: 1,
	}

	resp, err := requester.Do(ctx, http.MethodPost, "/v1/accounts", url.Values{"page": []string{"1"}}, map[string]string{"secret": marker})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("Status = %d, want 200", resp.Status)
	}
	telemetry.Shutdown(context.Background(), handle, func(string) {})

	data := readConnSDKTelemetry(t, dir)
	assertConnSDKContains(t, data, "pm.connector.http")
	assertConnSDKContains(t, data, "pm.http.scheme")
	assertConnSDKContains(t, data, "pm.http.host")
	assertConnSDKContains(t, data, "pm.http.path")
	assertConnSDKContains(t, data, "/v1/accounts")
	assertConnSDKContains(t, data, "pm.http.status_code")
	assertConnSDKContains(t, data, "pm.http.attempt")
	assertConnSDKNotContains(t, data, marker)
	assertConnSDKNotContains(t, data, "api_key")
	assertConnSDKNotContains(t, data, "page=1")
	assertConnSDKNotContains(t, data, "X-Secret-Echo")
	assertConnSDKNotContains(t, data, "request.body")
	assertConnSDKNotContains(t, data, "url.full")
}

func readConnSDKTelemetry(t *testing.T, dir string) []byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read telemetry dir: %v", err)
	}
	var out bytes.Buffer
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read telemetry file: %v", err)
		}
		out.Write(data)
	}
	if out.Len() == 0 {
		t.Fatalf("no telemetry JSONL data under %s", dir)
	}
	return out.Bytes()
}

func assertConnSDKContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if !bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output missing %q:\n%s", needle, data)
	}
}

func assertConnSDKNotContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output contains forbidden %q:\n%s", needle, data)
	}
}
