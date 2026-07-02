package conformance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// runDynamicChecks runs every dynamic (fixture-backed replay) check against
// an already-loaded Bundle. Per-stream/per-action checks are skipped
// (CheckResult.Skipped) when the bundle has no applicable stream/action
// (e.g. delete_semantics on a bundle with no delete write action) rather
// than silently omitted, so callers always see a stable, explainable check
// list.
func runDynamicChecks(b engine.Bundle) []CheckResult {
	var checks []CheckResult

	checks = append(checks, checkCheckFixture(b))

	for i, s := range b.Streams {
		mandatory := i == 0 // "first stream mandatory" per design §E.2
		checks = append(checks, checkReadFixtureNonempty(b, s.Name, mandatory))
	}

	checks = append(checks, checkPaginationTerminates(b, newHitTracker()))
	checks = append(checks, checkRecordsMatchSchema(b))
	checks = append(checks, checkCursorAdvances(b))
	checks = append(checks, checkWriteRequestShape(b)...)
	checks = append(checks, checkDeleteSemantics(b))

	return checks
}

// withReplayURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL. engine.Bundle is a value type (bundle.go) so this is a plain
// struct-field override, not a mutation of the caller's bundle.
func withReplayURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// runtimeConfigForEngine builds a minimal connectors.RuntimeConfig for
// dynamic checks: every spec-declared property gets a synthetic non-secret
// value (so required-field / interpolation resolution doesn't fail for want
// of a config value), and every x-secret property gets a synthetic secret
// value. Values are deliberately synthetic/non-realistic (never derived
// from real credentials) per THREAT-MODEL §4 — conformance never touches
// live secrets.
func runtimeConfigForEngine(b engine.Bundle) connectors.RuntimeConfig {
	cfg := connectors.RuntimeConfig{Config: map[string]string{}, Secrets: map[string]string{}}
	if b.Spec == nil {
		return cfg
	}
	secretSet := map[string]bool{}
	for _, k := range b.Spec.SecretKeys() {
		secretSet[k] = true
	}
	for _, name := range b.Spec.Properties() {
		if secretSet[name] {
			cfg.Secrets[name] = "synthetic-conformance-secret"
			continue
		}
		if name == "start_date" {
			cfg.Config[name] = "2020-01-01T00:00:00Z"
			continue
		}
		cfg.Config[name] = "synthetic-conformance-value"
	}
	return cfg
}

// readRequestFor builds a connectors.ReadRequest for streamName with cfg and
// an optional state map (nil = fresh/full sync).
func readRequestFor(streamName string, cfg connectors.RuntimeConfig, state map[string]string) connectors.ReadRequest {
	return connectors.ReadRequest{Stream: streamName, Config: cfg, State: state}
}

// writeRequestFor builds a connectors.WriteRequest for actionName using cfg.
func writeRequestFor(actionName string, cfg connectors.RuntimeConfig) connectors.WriteRequest {
	return connectors.WriteRequest{Action: actionName, Config: cfg}
}

// checkCheckFixture runs the bundle's declarative Check() against a replay
// server built from fixtures/check.json (a single recorded response, since
// Check() always issues exactly one request to one declared path). A
// bundle with no HTTP.Check declared, or no fixtures/check.json at all,
// trivially Skips (there is nothing to check against).
func checkCheckFixture(b engine.Bundle) CheckResult {
	const name = "check_fixture"
	if b.HTTP.Check == nil {
		return CheckResult{Name: name, Skipped: true}
	}
	fx, ok, err := loadCheckFixture(b.Fixtures)
	if err != nil {
		return CheckResult{Name: name, Error: err.Error()}
	}
	if !ok {
		return CheckResult{Name: name, Skipped: true}
	}

	srv := newCheckReplayServer(fx)
	defer srv.Close()

	rb := withReplayURL(b, srv.URL)
	err = engine.Check(context.Background(), rb, runtimeConfigForEngine(b), nil)
	return checkResultFromErr(name, err)
}

// checkReadFixtureNonempty runs a full engine.Read against streamName's
// fixture replay server and asserts at least one record was emitted. When
// mandatory is false (not the bundle's first stream) a stream with zero
// fixture pages is Skipped rather than failed — only the first stream is
// required to ship fixtures (design §E.2 / checkFixturesPresent).
func checkReadFixtureNonempty(b engine.Bundle, streamName string, mandatory bool) CheckResult {
	name := "read_fixture_nonempty:" + streamName
	pages, err := loadFixturePages(b.Fixtures, streamName)
	if err != nil {
		return CheckResult{Name: name, Error: err.Error()}
	}
	if len(pages) == 0 {
		if mandatory {
			return CheckResult{Name: name, Error: fmt.Sprintf("stream %q (first stream) has zero fixture pages", streamName)}
		}
		return CheckResult{Name: name, Skipped: true}
	}

	count := 0
	err = readRawRecords(b, streamName, nil, func(map[string]any) error {
		count++
		return nil
	})
	if err != nil {
		return CheckResult{Name: name, Error: err.Error()}
	}
	if count == 0 {
		return CheckResult{Name: name, Error: fmt.Sprintf("stream %q emitted zero records from its own fixtures", streamName)}
	}
	return CheckResult{Name: name, Passed: true}
}

// checkPaginationTerminates runs a full engine.Read against the bundle's
// FIRST stream (the one guaranteed to ship a fixture) and asserts, via
// tracker, that the read terminated (Read returned, meaning the paginator
// eventually stopped) and that pagination consumed EXACTLY one request per
// recorded fixture page (never more, never less — a page served twice
// would mean the paginator looped; fewer means fixtures were left
// unconsumed). A bundle with no streams, or whose first stream has zero
// fixtures, is Skipped.
func checkPaginationTerminates(b engine.Bundle, tracker *hitTracker) CheckResult {
	const name = "pagination_terminates"
	if len(b.Streams) == 0 {
		return CheckResult{Name: name, Skipped: true}
	}
	stream := b.Streams[0].Name
	pages, err := loadFixturePages(b.Fixtures, stream)
	if err != nil {
		return CheckResult{Name: name, Error: err.Error()}
	}
	if len(pages) == 0 {
		return CheckResult{Name: name, Skipped: true}
	}

	if err := readRawRecords(b, stream, tracker, func(map[string]any) error { return nil }); err != nil {
		return CheckResult{Name: name, Error: fmt.Sprintf("read did not terminate cleanly: %v", err)}
	}

	hits := tracker.hitsFor(stream)
	if hits != len(pages) {
		return CheckResult{Name: name, Error: fmt.Sprintf("stream %q: replay server served %d requests, want exactly %d (one per fixture page — pagination looped or under-consumed fixtures)", stream, hits, len(pages))}
	}
	return CheckResult{Name: name, Passed: true}
}

// checkRecordsMatchSchema runs a full read of every stream that has
// fixtures and validates each emitted RAW record against that stream's
// compiled schema (validation runs before projection drops undeclared
// fields, so a type-drifted field is caught even in "schema" projection
// mode). A bundle with no fixtured stream is Skipped.
func checkRecordsMatchSchema(b engine.Bundle) CheckResult {
	const name = "records_match_schema"
	anyFixtured := false
	for _, s := range b.Streams {
		pages, err := loadFixturePages(b.Fixtures, s.Name)
		if err != nil {
			return CheckResult{Name: name, Error: err.Error()}
		}
		if len(pages) == 0 {
			continue
		}
		anyFixtured = true
		sch, ok := b.Schemas[s.Name]
		if !ok {
			continue
		}

		var validateErr error
		err = readRawRecords(b, s.Name, nil, func(raw map[string]any) error {
			if validateErr == nil {
				if verr := sch.Validate(raw); verr != nil {
					validateErr = fmt.Errorf("stream %q: record failed schema validation: %w", s.Name, verr)
				}
			}
			return nil
		})
		if err != nil {
			return CheckResult{Name: name, Error: err.Error()}
		}
		if validateErr != nil {
			return CheckResult{Name: name, Error: validateErr.Error()}
		}
	}
	if !anyFixtured {
		return CheckResult{Name: name, Skipped: true}
	}
	return CheckResult{Name: name, Passed: true}
}

// checkCursorAdvances runs a full read of the first INCREMENTAL stream with
// fixtures, asserts the resulting max-observed cursor is non-empty, then
// re-reads seeded with that cursor as read state and asserts the re-read
// request actually carried the declared incremental.request_param formatted
// per param_format. A bundle with no incremental+fixtured stream is
// Skipped.
func checkCursorAdvances(b engine.Bundle) CheckResult {
	const name = "cursor_advances"
	stream, ok := firstIncrementalStreamWithFixtures(b)
	if !ok {
		return CheckResult{Name: name, Skipped: true}
	}

	sch := b.Schemas[stream.Name]
	var maxCursor string
	err := readRawRecords(b, stream.Name, nil, func(raw map[string]any) error {
		if sch != nil && sch.CursorField != "" {
			if v, ok := raw[sch.CursorField]; ok {
				if s, ok := v.(string); ok && s > maxCursor {
					maxCursor = s
				}
			}
		}
		return nil
	})
	if err != nil {
		return CheckResult{Name: name, Error: err.Error()}
	}
	if maxCursor == "" {
		return CheckResult{Name: name, Error: fmt.Sprintf("stream %q: no cursor value observed across fixture records", stream.Name)}
	}

	wantParam, err := formatCursorForAssertion(maxCursor, stream.Incremental.ParamFormat)
	if err != nil {
		return CheckResult{Name: name, Error: err.Error()}
	}

	// Re-read seeded with the observed cursor; capture the request_param
	// actually sent, independent of whether any recorded fixture page
	// happens to match the re-read's (necessarily different) query — the
	// capture server always answers 200 with an empty page so the read
	// terminates immediately after the one request this check inspects.
	capture := newParamCaptureServer(stream.Incremental.RequestParam)
	defer capture.Close()

	rb := withReplayURL(b, capture.URL)
	req := readRequestFor(stream.Name, runtimeConfigForEngine(b), map[string]string{"cursor": maxCursor})
	_ = engine.Read(context.Background(), rb, req, nil, func(connectors.Record) error { return nil })

	gotParam := capture.CapturedValue()
	if gotParam != wantParam {
		return CheckResult{Name: name, Error: fmt.Sprintf("re-read request_param %q = %q, want %q (cursor %q, param_format %q)", stream.Incremental.RequestParam, gotParam, wantParam, maxCursor, stream.Incremental.ParamFormat)}
	}
	return CheckResult{Name: name, Passed: true}
}

// checkWriteRequestShape runs, for every fixtures/writes/<action>.json, the
// engine's real write-request construction against a capture server and
// asserts the actually-sent method/path/body match the fixture's "expect"
// block. write_validate is folded into this same per-action result (a
// record that fails engine.ValidateWrite is itself a failure here) since
// both assertions apply to the same fixture file. A write action with no
// fixtures/writes/<action>.json is Skipped.
func checkWriteRequestShape(b engine.Bundle) []CheckResult {
	var out []CheckResult
	for _, action := range b.Writes {
		name := "write_request_shape:" + action.Name
		fx, err := loadWriteFixture(b.Fixtures, action.Name)
		if err != nil {
			out = append(out, CheckResult{Name: name, Skipped: true})
			continue
		}

		record := connectors.Record(fx.Record)
		ctx := context.Background()
		cfg := runtimeConfigForEngine(b)

		if verr := engine.ValidateWrite(ctx, b, writeRequestFor(action.Name, cfg), []connectors.Record{record}); verr != nil {
			out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("write_validate: fixture record failed validation: %v", verr)})
			continue
		}

		capture := newCaptureServer()
		rb := withReplayURL(b, capture.URL)
		if _, err := engine.Write(ctx, rb, writeRequestFor(action.Name, cfg), []connectors.Record{record}, nil); err != nil {
			capture.Close()
			out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("engine.Write against replay server failed: %v", err)})
			continue
		}
		got := capture.LastRequest()
		capture.Close()
		if got == nil {
			out = append(out, CheckResult{Name: name, Error: "engine.Write sent no HTTP request"})
			continue
		}

		if mismatch := compareWriteExpectation(*got, fx.Expect); mismatch != "" {
			out = append(out, CheckResult{Name: name, Error: mismatch})
			continue
		}
		out = append(out, CheckResult{Name: name, Passed: true})
	}
	return out
}

// checkDeleteSemantics exercises every kind:delete write action's
// missing_ok_status handling: a status in that allow-list must be treated
// as written, not failed. The real engine.Write is run against a server
// that always answers the FIRST allow-listed status, so this check would
// fail (RecordsFailed>0 / an error) if that handling ever regressed. A
// bundle with no such delete action, or no fixture for it, is Skipped.
func checkDeleteSemantics(b engine.Bundle) CheckResult {
	const name = "delete_semantics"
	var deleteAction *engine.WriteAction
	for i := range b.Writes {
		a := &b.Writes[i]
		if a.Kind == "delete" && a.Delete != nil && len(a.Delete.MissingOkStatus) > 0 {
			deleteAction = a
			break
		}
	}
	if deleteAction == nil {
		return CheckResult{Name: name, Skipped: true}
	}

	fx, err := loadWriteFixture(b.Fixtures, deleteAction.Name)
	if err != nil {
		return CheckResult{Name: name, Skipped: true}
	}

	status := deleteAction.Delete.MissingOkStatus[0]
	srv := newAlwaysStatusServer(status)
	defer srv.Close()

	rb := withReplayURL(b, srv.URL)
	cfg := runtimeConfigForEngine(b)
	record := connectors.Record(fx.Record)
	result, err := engine.Write(context.Background(), rb, writeRequestFor(deleteAction.Name, cfg), []connectors.Record{record}, nil)
	if err != nil {
		return CheckResult{Name: name, Error: fmt.Sprintf("delete with missing_ok_status %d returned an error instead of being treated as written: %v", status, err)}
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		return CheckResult{Name: name, Error: fmt.Sprintf("delete result = %+v, want RecordsWritten=1 RecordsFailed=0 for an allow-listed missing_ok_status %d", result, status)}
	}
	return CheckResult{Name: name, Passed: true}
}

// --- shared helpers -------------------------------------------------------

func checkResultFromErr(name string, err error) CheckResult {
	c := CheckResult{Name: name, Passed: err == nil}
	if err != nil {
		c.Error = err.Error()
	}
	return c
}

// readRawRecords is the shared engine.Read driver used by every dynamic
// read-based check: build a fresh replay server for streamName (tracker may
// be nil), point a bundle copy at it, and run the real engine, invoking
// onRecord for every emitted record.
func readRawRecords(b engine.Bundle, streamName string, tracker *hitTracker, onRecord func(map[string]any) error) error {
	srv, err := newStreamReplayServer(b.Fixtures, streamName, tracker)
	if err != nil {
		return err
	}
	defer srv.Close()

	rb := withReplayURL(b, srv.URL)
	req := readRequestFor(streamName, runtimeConfigForEngine(b), nil)
	return engine.Read(context.Background(), rb, req, nil, func(r connectors.Record) error {
		return onRecord(map[string]any(r))
	})
}

func firstIncrementalStreamWithFixtures(b engine.Bundle) (engine.StreamSpec, bool) {
	for _, s := range b.Streams {
		if s.Incremental == nil || s.Incremental.RequestParam == "" {
			continue
		}
		pages, err := loadFixturePages(b.Fixtures, s.Name)
		if err != nil || len(pages) == 0 {
			continue
		}
		return s, true
	}
	return engine.StreamSpec{}, false
}

// formatCursorForAssertion mirrors read.go's unexported formatParam so this
// package can independently assert the re-read request shape without
// reaching into engine internals (read.go does not export formatParam; this
// is a deliberate, small, documented duplication rather than a
// cross-package reach-in — it exists purely so the ASSERTION is derived
// independently of the code under test).
func formatCursorForAssertion(value, format string) (string, error) {
	switch format {
	case "", "rfc3339":
		return value, nil
	case "unix_seconds":
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return "", fmt.Errorf("param_format unix_seconds: invalid RFC3339 value %q: %w", value, err)
		}
		return fmt.Sprintf("%d", t.Unix()), nil
	case "date":
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return "", fmt.Errorf("param_format date: invalid RFC3339 value %q: %w", value, err)
		}
		return t.Format("2006-01-02"), nil
	case "github_date_range":
		return ">=" + value, nil
	default:
		return "", fmt.Errorf("unknown param_format %q", format)
	}
}

// --- write fixture parsing -------------------------------------------------

// writeFixture is fixtures/writes/<action>.json's shape (design §E.2):
// {"record": {...}, "expect": {"method","path","body"}}.
type writeFixture struct {
	Record map[string]any   `json:"record"`
	Expect writeExpectation `json:"expect"`
}

type writeExpectation struct {
	Method string         `json:"method"`
	Path   string         `json:"path"`
	Body   map[string]any `json:"body,omitempty"`
}

// loadWriteFixture reads fixtures/writes/<action>.json.
func loadWriteFixture(fixtures fs.FS, action string) (writeFixture, error) {
	if fixtures == nil {
		return writeFixture{}, fmt.Errorf("bundle has no fixtures/ directory")
	}
	p := path.Join("writes", action+".json")
	raw, err := fs.ReadFile(fixtures, p)
	if err != nil {
		return writeFixture{}, fmt.Errorf("read fixture %s: %w", p, err)
	}
	var fx writeFixture
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&fx); err != nil {
		return writeFixture{}, fmt.Errorf("parse fixture %s: %w", p, err)
	}
	return fx, nil
}

// capturedRequest is what a captureServer observed for the single write
// request it received.
type capturedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   map[string]any
}

// compareWriteExpectation compares a capturedRequest against the fixture's
// "expect" block and returns a non-empty mismatch description, or "" when
// they match. Body comparison is a subset match (every key in expect.Body
// must be present with an equal value in got.Body) since the engine may
// include additional non-path_fields record data the fixture author didn't
// bother spelling out for a DELETE/no-op-body action.
func compareWriteExpectation(got capturedRequest, want writeExpectation) string {
	if want.Method != "" && !strings.EqualFold(got.Method, want.Method) {
		return fmt.Sprintf("method = %q, want %q", got.Method, want.Method)
	}
	if want.Path != "" && got.Path != want.Path {
		return fmt.Sprintf("path = %q, want %q", got.Path, want.Path)
	}
	for k, wantVal := range want.Body {
		gotVal, ok := got.Body[k]
		if !ok {
			return fmt.Sprintf("body missing key %q (want %v)", k, wantVal)
		}
		if fmt.Sprint(gotVal) != fmt.Sprint(wantVal) {
			return fmt.Sprintf("body[%q] = %v, want %v", k, gotVal, wantVal)
		}
	}
	return ""
}

// --- capture / synthetic replay servers ------------------------------------

// captureServer is an httptest.Server that always answers 200 {} and
// records the last request it received (method/path/query/decoded JSON
// body) for write_request_shape's assertions.
type captureServer struct {
	*httptest.Server
	last *capturedRequest
}

func newCaptureServer() *captureServer {
	cs := &captureServer{}
	cs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		_ = dec.Decode(&body) // a body-less request (e.g. DELETE) decodes to nil, not an error worth surfacing
		cs.last = &capturedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Body: body}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	return cs
}

func (cs *captureServer) LastRequest() *capturedRequest { return cs.last }

// newAlwaysStatusServer returns an httptest.Server that answers every
// request with the given HTTP status and an empty JSON body — used by
// delete_semantics to simulate an idempotent-delete's "already gone"
// response.
func newAlwaysStatusServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte("{}"))
	}))
}

// paramCaptureServer is an httptest.Server that always answers 200 with an
// empty page (so a read terminates after exactly one request) and records
// the value of a single named query parameter from the request it
// received — used by cursor_advances to assert the incremental
// request_param sent on a re-read.
type paramCaptureServer struct {
	*httptest.Server
	param string
	value string
}

func newParamCaptureServer(param string) *paramCaptureServer {
	pcs := &paramCaptureServer{param: param}
	pcs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pcs.value = r.URL.Query().Get(param)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	return pcs
}

func (pcs *paramCaptureServer) CapturedValue() string { return pcs.value }
