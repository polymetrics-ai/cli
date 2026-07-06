package certify_test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// writeCassetteFile writes a single cassette entry JSON file at
// <dir>/<stage>/<seq>.json matching the shape RecordingTransport.Flush
// produces, for replay tests that don't need a live recording pass first.
func writeCassetteFile(t *testing.T, dir, stage, seq, json string) {
	t.Helper()
	stageDir := filepath.Join(dir, stage)
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", stageDir, err)
	}
	if err := os.WriteFile(filepath.Join(stageDir, seq+".json"), []byte(json), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
}

// TestReplayTransportServesCassetteInOrder proves ReplayTransport reads
// cassette entries back for a stage in per-stage sequence order and never
// hits the network (certification design §E: "Tier-1 replay transport +
// cassette store").
func TestReplayTransportServesCassetteInOrder(t *testing.T) {
	dir := t.TempDir()
	writeCassetteFile(t, dir, "catalog_live", "0001", `{
		"request":{"method":"GET","path":"/customers"},
		"response":{"status":200,"header":{"Content-Type":["application/json"]},"body":"{\"page\":1}"}
	}`)
	writeCassetteFile(t, dir, "catalog_live", "0002", `{
		"request":{"method":"GET","path":"/customers"},
		"response":{"status":200,"header":{"Content-Type":["application/json"]},"body":"{\"page\":2}"}
	}`)

	replay, err := certify.NewReplayTransport(dir)
	if err != nil {
		t.Fatalf("NewReplayTransport() error = %v", err)
	}
	replay.SetStage("catalog_live")

	req1, _ := http.NewRequest(http.MethodGet, "http://example.invalid/customers", nil)
	resp1, err := replay.RoundTrip(req1)
	if err != nil {
		t.Fatalf("RoundTrip(1) error = %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	_ = resp1.Body.Close()
	if string(body1) != `{"page":1}` {
		t.Errorf("body1 = %q, want page 1", string(body1))
	}

	req2, _ := http.NewRequest(http.MethodGet, "http://example.invalid/customers", nil)
	resp2, err := replay.RoundTrip(req2)
	if err != nil {
		t.Fatalf("RoundTrip(2) error = %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if string(body2) != `{"page":2}` {
		t.Errorf("body2 = %q, want page 2", string(body2))
	}
}

// TestReplayTransportSetsResponseStatusAndHeaders proves replayed responses
// carry the recorded status code and headers, not just the body.
func TestReplayTransportSetsResponseStatusAndHeaders(t *testing.T) {
	dir := t.TempDir()
	writeCassetteFile(t, dir, "stage_a", "0001", `{
		"request":{"method":"GET","path":"/x"},
		"response":{"status":404,"header":{"Content-Type":["application/json"]},"body":"{\"error\":\"not found\"}"}
	}`)

	replay, err := certify.NewReplayTransport(dir)
	if err != nil {
		t.Fatalf("NewReplayTransport() error = %v", err)
	}
	replay.SetStage("stage_a")

	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	resp, err := replay.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", resp.Header.Get("Content-Type"))
	}
}

// TestReplayTransportUnmatchedRequestFails is the design's core replay
// guarantee: "unmatched request in replay => failure (catches drift)"
// (certification design §E). An unrecognized method+path (or an
// already-exhausted sequence) must return an error, not a zero-value
// response, so a stage author can't silently pass with stale data.
func TestReplayTransportUnmatchedRequestFails(t *testing.T) {
	dir := t.TempDir()
	writeCassetteFile(t, dir, "stage_b", "0001", `{
		"request":{"method":"GET","path":"/known"},
		"response":{"status":200,"header":{},"body":"{}"}
	}`)

	replay, err := certify.NewReplayTransport(dir)
	if err != nil {
		t.Fatalf("NewReplayTransport() error = %v", err)
	}
	replay.SetStage("stage_b")

	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/unknown-path", nil)
	if _, err := replay.RoundTrip(req); err == nil {
		t.Fatalf("RoundTrip(unmatched path) error = nil, want failure")
	}
}

// TestReplayTransportExhaustedSequenceFails proves calling RoundTrip more
// times than the stage has cassette entries fails rather than looping/reusing
// the last entry silently.
func TestReplayTransportExhaustedSequenceFails(t *testing.T) {
	dir := t.TempDir()
	writeCassetteFile(t, dir, "stage_c", "0001", `{
		"request":{"method":"GET","path":"/x"},
		"response":{"status":200,"header":{},"body":"{}"}
	}`)

	replay, err := certify.NewReplayTransport(dir)
	if err != nil {
		t.Fatalf("NewReplayTransport() error = %v", err)
	}
	replay.SetStage("stage_c")

	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if _, err := replay.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip(1) error = %v, want success", err)
	}

	req2, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if _, err := replay.RoundTrip(req2); err == nil {
		t.Fatalf("RoundTrip(2, exhausted) error = nil, want failure")
	}
}

// TestReplayTransportMismatchedMethodFails proves the matcher checks method
// as well as path.
func TestReplayTransportMismatchedMethodFails(t *testing.T) {
	dir := t.TempDir()
	writeCassetteFile(t, dir, "stage_d", "0001", `{
		"request":{"method":"POST","path":"/x"},
		"response":{"status":200,"header":{},"body":"{}"}
	}`)

	replay, err := certify.NewReplayTransport(dir)
	if err != nil {
		t.Fatalf("NewReplayTransport() error = %v", err)
	}
	replay.SetStage("stage_d")

	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if _, err := replay.RoundTrip(req); err == nil {
		t.Fatalf("RoundTrip(method mismatch) error = nil, want failure")
	}
}

// TestReplayTransportUnknownStageFails proves selecting a stage with no
// recorded cassette directory fails clearly instead of panicking or
// returning an empty-but-successful response.
func TestReplayTransportUnknownStageFails(t *testing.T) {
	dir := t.TempDir()
	writeCassetteFile(t, dir, "stage_e", "0001", `{
		"request":{"method":"GET","path":"/x"},
		"response":{"status":200,"header":{},"body":"{}"}
	}`)

	replay, err := certify.NewReplayTransport(dir)
	if err != nil {
		t.Fatalf("NewReplayTransport() error = %v", err)
	}
	replay.SetStage("stage_never_recorded")

	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if _, err := replay.RoundTrip(req); err == nil {
		t.Fatalf("RoundTrip(unknown stage) error = nil, want failure")
	}
}

// TestRecordThenReplayRoundTrip proves a cassette written by
// RecordingTransport is directly consumable by ReplayTransport (the two
// halves of Tier-1 share one on-disk format).
func TestRecordThenReplayRoundTrip(t *testing.T) {
	srv := newTestServer(t, `{"id":"cus_1","name":"Ada"}`)
	dir := t.TempDir()

	rec := certify.NewRecordingTransport(http.DefaultTransport, dir, "etl_full_refresh_append")
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/customers", nil)
	resp, err := rec.RoundTrip(req)
	if err != nil {
		t.Fatalf("record RoundTrip() error = %v", err)
	}
	_ = resp.Body.Close()
	if err := rec.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	replay, err := certify.NewReplayTransport(dir)
	if err != nil {
		t.Fatalf("NewReplayTransport() error = %v", err)
	}
	replay.SetStage("etl_full_refresh_append")

	req2, _ := http.NewRequest(http.MethodGet, "http://example.invalid/customers", nil)
	resp2, err := replay.RoundTrip(req2)
	if err != nil {
		t.Fatalf("replay RoundTrip() error = %v", err)
	}
	body, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if string(body) != `{"id":"cus_1","name":"Ada"}` {
		t.Errorf("replayed body = %q, want recorded response body", string(body))
	}
}

// TestNewReplayTransportMissingDirErrors surfaces a clear error when the
// cassette root directory doesn't exist, rather than succeeding with an
// empty store that then fails every RoundTrip with a confusing message.
func TestNewReplayTransportMissingDirErrors(t *testing.T) {
	if _, err := certify.NewReplayTransport(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Fatalf("NewReplayTransport() error = nil, want error for missing directory")
	}
}
