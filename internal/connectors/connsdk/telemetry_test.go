package connsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/telemetry"
)

func TestRequesterDoEmitsSecretSafeHTTPSpanTelemetry(t *testing.T) {
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

	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "http-span"}, func(string) {})
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

func TestRequesterDoFailedHTTPSpanHasSafeErrorAndEventAttrs(t *testing.T) {
	const marker = "pm_http_response_body_marker"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "response body contains "+marker, http.StatusServiceUnavailable)
	}))
	defer server.Close()

	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "http-failed-span"}, func(string) {})
	requester := &Requester{
		Client:      server.Client(),
		BaseURL:     server.URL,
		MaxRetries:  1,
		BaseBackoff: time.Nanosecond,
		Sleep: func(context.Context, time.Duration) error {
			return nil
		},
	}

	_, err := requester.Do(ctx, http.MethodGet, "/v1/fail", url.Values{"api_key": []string{marker}}, nil)
	if err == nil {
		t.Fatal("Do error = nil, want HTTP failure")
	}
	telemetry.Shutdown(context.Background(), handle, func(string) {})

	data := readConnSDKTelemetry(t, dir)
	assertConnSDKContains(t, data, "pm.connector.http")
	assertConnSDKContains(t, data, "pm.error.type")
	assertConnSDKContains(t, data, "pm.error.code")
	assertConnSDKContains(t, data, "http_status")
	assertConnSDKContains(t, data, "pm.error.status_code")
	assertConnSDKNotContains(t, data, "exception.")
	assertConnSDKNotContains(t, data, "connsdk.HTTPError")
	assertConnSDKNotContains(t, data, marker)
	assertConnSDKNotContains(t, data, "response body")
	assertConnSDKNotContains(t, data, "api_key")
	if !connSDKSpanEventHasAttr(t, data, "pm.connector.http.retry", "pm.http.status_code") {
		t.Fatalf("retry event missing status attr:\n%s", data)
	}
	if !connSDKSpanEventHasAttr(t, data, "pm.connector.http.retry", "pm.http.attempt") {
		t.Fatalf("retry event missing attempt attr:\n%s", data)
	}
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

func connSDKSpanEventHasAttr(t *testing.T, data []byte, eventName, attrKey string) bool {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var span map[string]any
		err := dec.Decode(&span)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("decode span JSON: %v\n%s", err, data)
		}
		events, _ := span["Events"].([]any)
		for _, rawEvent := range events {
			event, _ := rawEvent.(map[string]any)
			if event["Name"] != eventName {
				continue
			}
			attrs, _ := event["Attributes"].([]any)
			for _, rawAttr := range attrs {
				attr, _ := rawAttr.(map[string]any)
				if attr["Key"] == attrKey {
					return true
				}
			}
		}
	}
	return false
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
