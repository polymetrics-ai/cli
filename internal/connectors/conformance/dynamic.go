package conformance

import (
	// Register all per-connector hooks so dynamic checks exercise the real
	// hook-dispatching engine paths (gap found by the wave1 Tier-2 pilots).
	_ "polymetrics.ai/internal/connectors/hooks/hookset"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
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
//
// Skip markers (R3, docs/migration/conventions.md §4): a bundle whose real
// behavior lives entirely behind a Tier-2 hook that a declarative fixture
// replay cannot exercise (e.g. a custom-auth-only AuthHook whose token_url
// conformance's synthetic config can never populate) may declare an
// explicit, reason-carrying bundle-level marker
// (Metadata.Conformance.SkipDynamic) that Skips every auth-dependent dynamic
// check outright — check_fixture, every read_fixture_nonempty:<stream>,
// pagination_terminates, records_match_schema, cursor_advances, and (Pass B
// gmail full-surface expansion fix, below) every write_request_shape/
// delete_semantics check too — rather than attempting them and reporting a
// predictable, uninformative failure. A narrower per-stream marker
// (StreamSpec.Conformance.SkipDynamic) has the same effect scoped to exactly
// one stream: that stream's read_fixture_nonempty Skips, and the stream is
// excluded from every other check's candidate-stream selection
// (pagination_terminates' first-stream pick, records_match_schema's
// per-stream iteration, cursor_advances' first-incremental-stream pick) as
// if it did not exist for dynamic-check purposes; the per-stream marker does
// NOT affect write checks (a stream-level auth problem says nothing about a
// sibling write action's own auth resolution). Neither marker affects STATIC
// checks (checkFixturesPresent etc.) — those never resolve auth at all.
//
// Historical note (superseded by this fix): this comment previously claimed
// the bundle-level marker "does not affect write checks... no shipped
// bundle needs that combination today (a marked bundle/stream is always
// read-only in this wave)". gmail's Pass B full-surface expansion is the
// first bundle to combine a bundle-level skip_dynamic marker (mode:custom,
// no when-gated non-custom fallback) with a non-empty writes.json: every one
// of its 35 write actions shares the identical bundle-wide base.auth the
// marker's reason already describes, so checkWriteRequestShape/
// checkDeleteSemantics would otherwise fail identically and uninformatively
// for every single write action, for the exact same underlying reason the
// marker already documents for reads (conformance's synthetic config can
// never carry a real https token_url, and the AuthHook fails closed on
// anything else). The marker's job widens here from "describe which READ
// behavior is hook-only" to "describe which dynamic behavior (read OR
// write) is hook-only" — still narrowly a description of what's
// auth-blocked, never a blanket exemption from every conformance guarantee
// (the marker must name hook/native tests or archived pre-deletion parity
// evidence as its substitute proof).
func runDynamicChecks(b engine.Bundle) []CheckResult {
	var checks []CheckResult

	if reason, ok := bundleSkipReason(b); ok {
		checks = append(checks, CheckResult{Name: "check_fixture", Skipped: true, Error: reason})
		for _, s := range b.Streams {
			checks = append(checks, CheckResult{Name: "read_fixture_nonempty:" + s.Name, Skipped: true, Error: reason})
		}
		checks = append(checks, CheckResult{Name: "pagination_terminates", Skipped: true, Error: reason})
		checks = append(checks, CheckResult{Name: "records_match_schema", Skipped: true, Error: reason})
		checks = append(checks, CheckResult{Name: "cursor_advances", Skipped: true, Error: reason})
		for _, a := range b.Writes {
			checks = append(checks, CheckResult{Name: "write_request_shape:" + a.Name, Skipped: true, Error: reason})
		}
		checks = append(checks, CheckResult{Name: "delete_semantics", Skipped: true, Error: reason})
		return checks
	}

	checks = append(checks, checkCheckFixture(b))

	readReplay := newReusableStreamReplayServer()
	defer readReplay.Close()

	for i, s := range b.Streams {
		if reason, ok := streamSkipReason(s); ok {
			checks = append(checks, CheckResult{Name: "read_fixture_nonempty:" + s.Name, Skipped: true, Error: reason})
			continue
		}
		mandatory := i == 0 // "first stream mandatory" per design §E.2
		checks = append(checks, checkReadFixtureNonemptyWithReplay(b, s.Name, mandatory, readReplay))
	}

	checks = append(checks, checkPaginationTerminatesWithReplay(b, newHitTracker(), readReplay))
	checks = append(checks, checkRecordsMatchSchemaWithReplay(b, readReplay))
	checks = append(checks, checkCursorAdvancesWithReplay(b, readReplay))
	checks = append(checks, checkWriteRequestShape(b)...)
	checks = append(checks, checkDeleteSemantics(b))

	return checks
}

// bundleSkipReason reports the bundle-level conformance skip marker's reason
// text, and whether one is present with SkipDynamic set. A marker with
// SkipDynamic false (or absent Metadata.Conformance) is "no marker" — ok is
// false — even if Reason happens to be non-empty (connectorgen validate
// enforces the inverse: SkipDynamic implies a non-empty Reason, but an
// author is free to leave a stale Reason on a marker they've since flipped
// off; that is not this package's concern).
func bundleSkipReason(b engine.Bundle) (reason string, ok bool) {
	m := b.Metadata.Conformance
	if m == nil || !m.SkipDynamic {
		return "", false
	}
	return m.Reason, true
}

// streamSkipReason mirrors bundleSkipReason for a single StreamSpec's own
// marker.
func streamSkipReason(s engine.StreamSpec) (reason string, ok bool) {
	m := s.Conformance
	if m == nil || !m.SkipDynamic {
		return "", false
	}
	return m.Reason, true
}

// streamIsSkipped is streamSkipReason without the reason text, for callers
// that only need the boolean (candidate-stream exclusion in
// checkPaginationTerminates/checkRecordsMatchSchema/firstIncrementalStreamWithFixtures).
func streamIsSkipped(s engine.StreamSpec) bool {
	_, ok := streamSkipReason(s)
	return ok
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

	checkQueryKeys := make([]string, 0, len(b.HTTP.Check.Query))
	for k := range b.HTTP.Check.Query {
		checkQueryKeys = append(checkQueryKeys, k)
	}
	srv := newCheckReplayServer(fx, checkQueryKeys)
	defer srv.Close()

	rb := withReplayURL(b, srv.URL)
	err = engine.Check(context.Background(), rb, runtimeConfigForEngine(b), engine.HooksFor(b.Name))
	return checkResultFromErr(name, err)
}

// checkReadFixtureNonempty runs a full engine.Read against streamName's
// fixture replay server and asserts at least one record was emitted. When
// mandatory is false (not the bundle's first stream) a stream with zero
// fixture pages is Skipped rather than failed — only the first stream is
// required to ship fixtures (design §E.2 / checkFixturesPresent).
func checkReadFixtureNonempty(b engine.Bundle, streamName string, mandatory bool) CheckResult {
	readReplay := newReusableStreamReplayServer()
	defer readReplay.Close()
	return checkReadFixtureNonemptyWithReplay(b, streamName, mandatory, readReplay)
}

func checkReadFixtureNonemptyWithReplay(b engine.Bundle, streamName string, mandatory bool, readReplay *reusableStreamReplayServer) CheckResult {
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
	err = readRawRecordsWithReplay(b, streamName, nil, readReplay, func(map[string]any) error {
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
// FIRST non-marker-excluded stream (the one guaranteed to ship a fixture)
// and asserts, via tracker, that the read terminated (Read returned, meaning
// the paginator eventually stopped) and that pagination consumed EXACTLY one
// request per recorded fixture page (never more, never less — a page served
// twice would mean the paginator looped; fewer means fixtures were left
// unconsumed). A bundle with no eligible streams, or whose candidate stream
// has zero fixtures, is Skipped; a bundle whose ONLY stream(s) carry a
// skip_dynamic marker Skips with that marker's reason instead (R3).
func checkPaginationTerminates(b engine.Bundle, tracker *hitTracker) CheckResult {
	readReplay := newReusableStreamReplayServer()
	defer readReplay.Close()
	return checkPaginationTerminatesWithReplay(b, tracker, readReplay)
}

func checkPaginationTerminatesWithReplay(b engine.Bundle, tracker *hitTracker, readReplay *reusableStreamReplayServer) CheckResult {
	const name = "pagination_terminates"
	if len(b.Streams) == 0 {
		return CheckResult{Name: name, Skipped: true}
	}
	if reason, ok := firstStreamSkipReasonIfAllExcluded(b.Streams); ok {
		return CheckResult{Name: name, Skipped: true, Error: reason}
	}
	first, ok := firstEligibleStream(b.Streams)
	if !ok {
		return CheckResult{Name: name, Skipped: true}
	}
	stream := first.Name
	pages, err := loadFixturePages(b.Fixtures, stream)
	if err != nil {
		return CheckResult{Name: name, Error: err.Error()}
	}
	if len(pages) == 0 {
		return CheckResult{Name: name, Skipped: true}
	}

	if err := readRawRecordsWithReplay(b, stream, tracker, readReplay, func(map[string]any) error { return nil }); err != nil {
		return CheckResult{Name: name, Error: fmt.Sprintf("read did not terminate cleanly: %v", err)}
	}

	hits := tracker.hitsFor(stream)
	if hits != len(pages) {
		return CheckResult{Name: name, Error: fmt.Sprintf("stream %q: replay server served %d requests, want exactly %d (one per fixture page — pagination looped or under-consumed fixtures)", stream, hits, len(pages))}
	}
	return CheckResult{Name: name, Passed: true}
}

// firstEligibleStream returns the first stream with no skip_dynamic marker,
// mirroring "the bundle's first stream" for every dynamic check that used to
// hardcode b.Streams[0] before marker-exclusion existed (R3).
func firstEligibleStream(streams []engine.StreamSpec) (engine.StreamSpec, bool) {
	for _, s := range streams {
		if !streamIsSkipped(s) {
			return s, true
		}
	}
	return engine.StreamSpec{}, false
}

// firstStreamSkipReasonIfAllExcluded reports the first marked stream's
// reason when EVERY declared stream carries a skip_dynamic marker (so a
// pagination_terminates/cursor_advances-style "pick the first eligible
// stream" check has literally no candidate left) — this lets the resulting
// Skip name the authoritative substitute instead of degrading to the
// pre-existing generic "no streams"/"no fixtures" Skip (which carries no
// reason at all).
func firstStreamSkipReasonIfAllExcluded(streams []engine.StreamSpec) (reason string, ok bool) {
	if len(streams) == 0 {
		return "", false
	}
	for _, s := range streams {
		if !streamIsSkipped(s) {
			return "", false
		}
	}
	reason, _ = streamSkipReason(streams[0])
	return reason, true
}

// checkRecordsMatchSchema runs a full read of every non-marker-excluded
// stream that has fixtures and validates each emitted RAW record against
// that stream's compiled schema (validation runs before projection drops
// undeclared fields, so a type-drifted field is caught even in "schema"
// projection mode). A bundle with no eligible fixtured stream is Skipped; a
// bundle whose ONLY stream(s) carry a skip_dynamic marker Skips with that
// marker's reason instead (R3).
func checkRecordsMatchSchema(b engine.Bundle) CheckResult {
	readReplay := newReusableStreamReplayServer()
	defer readReplay.Close()
	return checkRecordsMatchSchemaWithReplay(b, readReplay)
}

func checkRecordsMatchSchemaWithReplay(b engine.Bundle, readReplay *reusableStreamReplayServer) CheckResult {
	const name = "records_match_schema"
	if reason, ok := firstStreamSkipReasonIfAllExcluded(b.Streams); ok {
		return CheckResult{Name: name, Skipped: true, Error: reason}
	}
	anyFixtured := false
	for _, s := range b.Streams {
		if streamIsSkipped(s) {
			continue
		}
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
		err = readRawRecordsWithReplay(b, s.Name, nil, readReplay, func(raw map[string]any) error {
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

// checkCursorAdvances runs a full read of the first non-marker-excluded
// INCREMENTAL stream with fixtures, asserts the resulting max-observed
// cursor is non-empty, then re-reads seeded with that cursor as read state
// and asserts the re-read request actually carried the declared
// incremental.request_param formatted per param_format. A bundle with no
// incremental+fixtured stream at all is Skipped; a bundle whose ONLY
// incremental+fixtured candidate(s) carry a skip_dynamic marker Skips with
// that marker's reason instead (R3).
func checkCursorAdvances(b engine.Bundle) CheckResult {
	readReplay := newReusableStreamReplayServer()
	defer readReplay.Close()
	return checkCursorAdvancesWithReplay(b, readReplay)
}

func checkCursorAdvancesWithReplay(b engine.Bundle, readReplay *reusableStreamReplayServer) CheckResult {
	const name = "cursor_advances"
	if reason, ok := incrementalStreamSkipReasonIfOnlyCandidatesExcluded(b); ok {
		return CheckResult{Name: name, Skipped: true, Error: reason}
	}
	stream, ok := firstIncrementalStreamWithFixtures(b)
	if !ok {
		return CheckResult{Name: name, Skipped: true}
	}

	sch := b.Schemas[stream.Name]
	var maxCursor string
	var maxCursorNumeric bool // true once maxCursor holds a numeric-cursor value, so comparisons stay numeric-aware
	err := readRawRecordsWithReplay(b, stream.Name, nil, readReplay, func(raw map[string]any) error {
		if sch != nil && sch.CursorField != "" {
			if v, ok := raw[sch.CursorField]; ok {
				if s, numeric, ok := cursorValueString(v); ok {
					switch {
					case maxCursor == "":
						maxCursor, maxCursorNumeric = s, numeric
					case numeric && maxCursorNumeric:
						if cursorNumericGreater(s, maxCursor) {
							maxCursor = s
						}
					case s > maxCursor:
						maxCursor = s
					}
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
	_ = engine.Read(context.Background(), rb, req, engine.HooksFor(b.Name), func(connectors.Record) error { return nil })

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
	capture := newCaptureServer(nil)
	defer capture.Close()

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

		capture.Reset(fx.Response)
		rb := withReplayURL(b, capture.URL)
		if _, err := engine.Write(ctx, rb, writeRequestFor(action.Name, cfg), []connectors.Record{record}, engine.HooksFor(b.Name)); err != nil {
			out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("engine.Write against replay server failed: %v", err)})
			continue
		}
		got := capture.LastRequest()
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
//
// engine.HooksFor(b.Name) is passed here (Pass B gmail full-surface
// expansion fix) for consistency with checkCheckFixture/checkReadFixtureNonempty
// generalizations/checkWriteRequestShape, all of which already resolve the
// bundle's real registered hooks — every OTHER delete-capable bundle before
// gmail declared purely declarative (bearer/basic/apikey) auth, so a nil
// Hooks argument never mattered; gmail is the first delete-capable bundle
// whose sole auth candidate is mode:custom (an AuthHook), so a nil Hooks
// argument here made this check hard-fail with "hook not registered" for
// every one of gmail's delete_message/delete_thread/delete_draft/etc.
// actions, independent of their own correctness. This was a conformance
// harness inconsistency, not a gmail-specific workaround.
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
	result, err := engine.Write(context.Background(), rb, writeRequestFor(deleteAction.Name, cfg), []connectors.Record{record}, engine.HooksFor(b.Name))
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

// readRawRecordsWithReplay is the shared engine.Read driver used by every
// dynamic read-based check. It points a bundle copy at readReplay and runs the
// real engine, invoking onRecord for every emitted record.
func readRawRecordsWithReplay(b engine.Bundle, streamName string, tracker *hitTracker, readReplay *reusableStreamReplayServer, onRecord func(map[string]any) error) error {
	pages, err := loadFixturePages(b.Fixtures, streamName)
	if err != nil {
		return err
	}
	readReplay.reset(streamName, pages, tracker)
	defer readReplay.reset("", nil, nil)

	rb := withReplayURL(b, readReplay.URL)
	req := readRequestFor(streamName, runtimeConfigForEngine(b), nil)
	return engine.Read(context.Background(), rb, req, engine.HooksFor(b.Name), func(r connectors.Record) error {
		return onRecord(map[string]any(r))
	})
}

func firstIncrementalStreamWithFixtures(b engine.Bundle) (engine.StreamSpec, bool) {
	for _, s := range b.Streams {
		if streamIsSkipped(s) {
			continue
		}
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

// incrementalStreamSkipReasonIfOnlyCandidatesExcluded reports a marked
// stream's reason when at least one incremental+fixtured stream exists but
// EVERY such candidate is marker-excluded (R3) — this lets
// checkCursorAdvances name the authoritative substitute instead of
// degrading to the pre-existing generic "no incremental stream" Skip (no
// reason at all). Returns ok=false when there is no incremental+fixtured
// stream at all (marked or not), which is the pre-existing, unrelated Skip
// case checkCursorAdvances already handles.
func incrementalStreamSkipReasonIfOnlyCandidatesExcluded(b engine.Bundle) (reason string, ok bool) {
	sawCandidate := false
	for _, s := range b.Streams {
		if s.Incremental == nil || s.Incremental.RequestParam == "" {
			continue
		}
		pages, err := loadFixturePages(b.Fixtures, s.Name)
		if err != nil || len(pages) == 0 {
			continue
		}
		sawCandidate = true
		if !streamIsSkipped(s) {
			return "", false
		}
		if reason == "" {
			reason, _ = streamSkipReason(s)
		}
	}
	if !sawCandidate {
		return "", false
	}
	return reason, true
}

// cursorValueString extracts a comparable/formattable string form from a raw
// cursor field value decoded from fixture JSON, plus whether that form is
// numeric (so callers can compare numerically rather than lexicographically
// — lexicographic comparison of digit strings is wrong in general, e.g. "9" >
// "10"). Every JSON decoder used across this codebase (connsdk, engine)
// decodes numbers with UseNumber, so json.Number is the real-world shape for
// a numeric cursor field (Stripe's `created`, etc.); float64 is accepted
// defensively for any caller that doesn't. Any other type (bool, object,
// array, nil) is not a valid cursor value and returns ok=false, matching the
// prior string-only behavior for non-string/non-numeric values.
func cursorValueString(v any) (s string, numeric bool, ok bool) {
	switch t := v.(type) {
	case string:
		return t, false, true
	case json.Number:
		return t.String(), true, true
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), true, true
	default:
		return "", false, false
	}
}

// cursorNumericGreater reports whether digit-string a represents a strictly
// greater numeric value than digit-string b. Both inputs are guaranteed
// numeric-shaped by cursorValueString's callers (json.Number/float64
// canonical forms), so a big.Float parse failure indicates a genuine bug
// rather than untrusted input; on failure this falls back to lexicographic
// comparison rather than panicking.
func cursorNumericGreater(a, b string) bool {
	af, aok := new(big.Float).SetString(a)
	bf, bok := new(big.Float).SetString(b)
	if !aok || !bok {
		return a > b
	}
	return af.Cmp(bf) > 0
}

// formatCursorForAssertion mirrors read.go's unexported formatParam so this
// package can independently assert the re-read request shape without
// reaching into engine internals (read.go does not export formatParam; this
// is a deliberate, small, documented duplication rather than a
// cross-package reach-in — it exists purely so the ASSERTION is derived
// independently of the code under test).
//
// N1 (wave0 REVIEW.md re-review "New findings"): the github_date_range
// branch used to return ">=" + value VERBATIM, while the engine's real
// formatParam parses value (digits-passthrough-as-Unix-seconds or RFC3339,
// parseLowerBoundTime) and re-emits ">=" + t.UTC().Format(time.RFC3339) —
// i.e. it ALWAYS normalizes to UTC second precision regardless of which
// shape the input cursor arrived in (a bare digit string, the app-persisted
// numeric-cursor shape, or a non-UTC-offset/fractional-second RFC3339
// string). The verbatim version only happened to match when maxCursor was
// already exactly-UTC-normalized RFC3339 with no fractional seconds; a
// numeric cursor or a non-UTC offset would falsely FAIL cursor_advances.
// This branch now shares parseLowerBoundTimeForAssertion with unix_seconds/
// date above, so all three timestamp-parsing formats normalize identically.
func formatCursorForAssertion(value, format string) (string, error) {
	switch format {
	case "", "rfc3339":
		return value, nil
	case "unix_seconds":
		t, err := parseLowerBoundTimeForAssertion(value)
		if err != nil {
			return "", fmt.Errorf("param_format unix_seconds: %w", err)
		}
		return fmt.Sprintf("%d", t.Unix()), nil
	case "date":
		t, err := parseLowerBoundTimeForAssertion(value)
		if err != nil {
			return "", fmt.Errorf("param_format date: %w", err)
		}
		return t.Format("2006-01-02"), nil
	case "github_date_range":
		t, err := parseLowerBoundTimeForAssertion(value)
		if err != nil {
			return "", fmt.Errorf("param_format github_date_range: %w", err)
		}
		return ">=" + t.UTC().Format(time.RFC3339), nil
	default:
		return "", fmt.Errorf("unknown param_format %q", format)
	}
}

// parseLowerBoundTimeForAssertion mirrors read.go's unexported
// parseLowerBoundTime (R1's B1 fix): a digits-only value is treated as
// Unix-seconds already (the real-world shape a numeric cursor field's
// max-observed value takes, per cursorValueString above); otherwise it is
// parsed as RFC3339 (the shape a string cursor field's max-observed value
// takes). Documented duplication, same rationale as formatCursorForAssertion.
func parseLowerBoundTimeForAssertion(value string) (time.Time, error) {
	if isAllDigitsForAssertion(value) {
		secs, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid unix-seconds value %q: %w", value, err)
		}
		return time.Unix(secs, 0).UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid RFC3339 value %q: %w", value, err)
	}
	return t, nil
}

// isAllDigitsForAssertion mirrors read.go's unexported isAllDigits.
func isAllDigitsForAssertion(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[0] == '-' {
		i = 1
	}
	if i == len(s) {
		return false
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// --- write fixture parsing -------------------------------------------------

// writeFixture is fixtures/writes/<action>.json's shape (design §E.2):
// {"record": {...}, "expect": {"method","path","body"}}.
type writeFixture struct {
	Record   map[string]any   `json:"record"`
	Expect   writeExpectation `json:"expect"`
	Response *fixtureResponse `json:"response,omitempty"` // optional: what the capture server should answer with (see newCaptureServer)
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

// captureServer is an httptest.Server that answers every request with a
// fixed response (200 {} by default, or a fixture-declared response — see
// newCaptureServer) and records the last request it received (method/path/
// query/decoded JSON body) for write_request_shape's assertions.
type captureServer struct {
	*httptest.Server
	mu   sync.Mutex
	resp *fixtureResponse
	last *capturedRequest
}

// newCaptureServer builds a captureServer. When resp is non-nil, every
// request is answered with resp's declared status/body (defaulting an unset
// status to 200 and an empty body to "{}", mirroring newCheckReplayServer's
// same defaulting) — this lets write_request_shape assert against a
// WriteHook whose follow-up logic reads its own write response (e.g.
// github's createPullRequest decoding the POST response's "number" field
// before issuing follow-up requests). When resp is nil (no fixture
// "response" block declared), the pre-existing hardcoded 200 {} behavior is
// unchanged byte-for-byte, so every write fixture that never needed a
// custom response is unaffected.
func newCaptureServer(resp *fixtureResponse) *captureServer {
	cs := &captureServer{resp: resp}
	cs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		_ = dec.Decode(&body) // a body-less request (e.g. DELETE) decodes to nil, not an error worth surfacing

		cs.mu.Lock()
		cs.last = &capturedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Body: body}
		resp := cs.resp
		cs.mu.Unlock()

		status := http.StatusOK
		respBody := []byte("{}")
		if resp != nil {
			if resp.Status != 0 {
				status = resp.Status
			}
			if len(resp.Body) > 0 {
				respBody = resp.Body
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(respBody)
	}))
	return cs
}

func (cs *captureServer) Reset(resp *fixtureResponse) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.resp = resp
	cs.last = nil
}

func (cs *captureServer) LastRequest() *capturedRequest {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.last
}

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
