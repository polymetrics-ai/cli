package conformance

import (
	// Register all per-connector hooks so dynamic checks exercise the real
	// hook-dispatching engine paths (gap found by the wave1 Tier-2 pilots).
	_ "polymetrics.ai/internal/connectors/hooks/hookset"

	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
		for _, operation := range b.Operations {
			if operation.Kind == "rest_write" && operation.REST != nil && operation.REST.Multipart != nil {
				checks = append(checks, CheckResult{Name: "operation_multipart_request_shape:" + operation.ID, Skipped: true, Error: reason})
			}
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
	checks = append(checks, checkOperationMultipartRequestShape(b)...)
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
// dynamic checks. ProjectDir carries the reserved conformance-fixture sentinel
// used by custom auth hooks to avoid token exchanges; every spec-declared
// property gets a synthetic non-secret value (so required-field/interpolation
// resolution does not fail for want of config), and every x-secret property
// gets a synthetic secret value. Values are deliberately synthetic and never
// derived from real credentials per THREAT-MODEL §4 — conformance never
// touches live secrets or provider endpoints.
func runtimeConfigForEngine(b engine.Bundle) connectors.RuntimeConfig {
	cfg := connectors.RuntimeConfig{ProjectDir: "__polymetrics_conformance_fixture__", Config: map[string]string{}, Secrets: map[string]string{}}
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

// approvedFixtureWriteRequest carries synthetic, non-secret approval evidence
// for the conformance replay server. It is intentionally derived from the
// engine's real no-network preview so fixture execution exercises the same
// gate as production without granting a bypass to arbitrary callers.
func approvedFixtureWriteRequest(ctx context.Context, b engine.Bundle, actionName string, cfg connectors.RuntimeConfig, records []connectors.Record, hooks engine.Hooks) (connectors.WriteRequest, error) {
	fixtureReq := writeRequestFor(actionName, cfg)
	digests, err := engine.ApprovedMultipartPayloadSHA256ForWrite(ctx, b, fixtureReq, records, hooks)
	if err != nil {
		return connectors.WriteRequest{}, fmt.Errorf("bind conformance fixture multipart payloads: %w", err)
	}
	cfg.ApprovedPayloadSHA256 = digests
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		return connectors.WriteRequest{}, err
	}
	cfg.CredentialRevision, err = authority.CredentialRevision("conformance:"+b.Name, cfg.Secrets)
	if err != nil {
		return connectors.WriteRequest{}, err
	}
	cfg.ConfigurationDigest, err = authority.ConfigurationDigest("conformance:"+b.Name, cfg.Config)
	if err != nil {
		return connectors.WriteRequest{}, err
	}
	cfg.WriteApprovalScope = connectors.WriteApprovalScopeFixture
	req := writeRequestFor(actionName, cfg)
	preview, err := engine.DryRunWrite(ctx, b, req, records, hooks)
	if err != nil {
		return connectors.WriteRequest{}, err
	}
	planPayload, err := json.Marshal(struct {
		Connector string                   `json:"connector"`
		Action    string                   `json:"action"`
		Config    connectors.RuntimeConfig `json:"config"`
		Records   []connectors.Record      `json:"records"`
	}{Connector: b.Name, Action: actionName, Config: cfg, Records: records})
	if err != nil {
		return connectors.WriteRequest{}, fmt.Errorf("marshal conformance fixture plan: %w", err)
	}
	planHash := sha256.Sum256(planPayload)
	// Only a target the engine gate will actually stop carries a confirmation,
	// and a grant exists solely to satisfy that stop — IssueWriteGrant refuses
	// to mint one for anything else. A safe action reaches the same dispatch
	// seam with no evidence, so requesting a grant here would be asking the
	// authority to authorize a write nothing is holding back.
	if preview.ApprovalTarget.Confirmation.Kind == "" {
		return req, nil
	}
	token := "conformance-fixture-approval"
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID:        "rplan_conformance_fixture",
		PlanHash:      fmt.Sprintf("%x", planHash),
		PreviewDigest: preview.Digest,
		ApprovalToken: token,
		Target:        preview.ApprovalTarget,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		return connectors.WriteRequest{}, err
	}
	req.Approval, err = authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID:        grant.PlanID,
		PlanHash:      grant.PlanHash,
		PreviewDigest: grant.PreviewDigest,
		ApprovalToken: token,
		Target:        grant.Target,
		Confirmation:  grant.Confirmation,
	})
	if err != nil {
		return connectors.WriteRequest{}, err
	}
	return req, nil
}

// stagedFixtureMultipartPayloads materializes only the bytes declared by a
// write action's multipart file parts into a private, temporary project root.
// Assets live beside their write fixture under
// fixtures/writes/<action>.payloads/<declared-record-field>; the record still
// carries the bounded project-relative destination path that execution uses.
// No payload content is retained in plans, errors, or conformance reports.
type stagedFixtureMultipartPayloads struct {
	root string
}

func (s *stagedFixtureMultipartPayloads) Close() {
	if s != nil && s.root != "" {
		_ = os.RemoveAll(s.root)
	}
}

func stageFixtureMultipartPayloads(b engine.Bundle, action engine.WriteAction, record connectors.Record) (*stagedFixtureMultipartPayloads, error) {
	if action.Multipart == nil {
		return nil, nil
	}
	if b.Fixtures == nil {
		return nil, fmt.Errorf("fixture payload staging requires a fixture filesystem")
	}
	root, err := os.MkdirTemp("", "polymetrics-conformance-multipart-")
	if err != nil {
		return nil, fmt.Errorf("create fixture multipart project root: %w", err)
	}
	staged := &stagedFixtureMultipartPayloads{root: root}
	fail := func(err error) (*stagedFixtureMultipartPayloads, error) {
		staged.Close()
		return nil, err
	}
	seenPaths := make(map[string]struct{})
	for _, part := range action.Multipart.Parts {
		if part.Type != "file" {
			continue
		}
		value, present := fixtureRecordPathValue(record, part.Field)
		if !present || value == nil {
			if part.Required {
				return fail(fmt.Errorf("fixture multipart part %q is missing declared record field %q", part.Name, part.Field))
			}
			continue
		}
		raw, ok := value.(string)
		if !ok || strings.TrimSpace(raw) == "" {
			return fail(fmt.Errorf("fixture multipart part %q declared record field %q must be a non-empty file path", part.Name, part.Field))
		}
		if strings.TrimSpace(part.Field) == "" || strings.ContainsAny(part.Field, "/\\") {
			return fail(fmt.Errorf("fixture multipart part %q has an unsafe declared record field %q", part.Name, part.Field))
		}
		assetPath := path.Join("writes", action.Name+".payloads", part.Field)
		if !fs.ValidPath(assetPath) {
			return fail(fmt.Errorf("fixture multipart part %q has invalid payload asset path", part.Name))
		}
		payload, err := readBoundedFixturePayload(b.Fixtures, assetPath, part.MaxBytes)
		if err != nil {
			return fail(fmt.Errorf("fixture multipart part %q: %w", part.Name, err))
		}
		rel := filepath.Clean(filepath.FromSlash(raw))
		if filepath.IsAbs(rel) || !filepath.IsLocal(rel) {
			return fail(fmt.Errorf("fixture multipart part %q destination must be project-relative", part.Name))
		}
		if _, duplicate := seenPaths[rel]; duplicate {
			return fail(fmt.Errorf("fixture multipart parts reuse source path %q", rel))
		}
		seenPaths[rel] = struct{}{}
		destination := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			return fail(fmt.Errorf("fixture multipart part %q create destination: %w", part.Name, err))
		}
		if err := os.WriteFile(destination, payload, 0o600); err != nil {
			return fail(fmt.Errorf("fixture multipart part %q stage payload: %w", part.Name, err))
		}
	}
	return staged, nil
}

func stageFixtureOperationMultipartPayloads(b engine.Bundle, operation engine.OperationSpec, record connectors.Record) (*stagedFixtureMultipartPayloads, error) {
	if operation.REST == nil || operation.REST.Multipart == nil {
		return nil, nil
	}
	if b.Fixtures == nil {
		return nil, fmt.Errorf("fixture payload staging requires a fixture filesystem")
	}
	root, err := os.MkdirTemp("", "polymetrics-conformance-operation-multipart-")
	if err != nil {
		return nil, fmt.Errorf("create fixture multipart project root: %w", err)
	}
	staged := &stagedFixtureMultipartPayloads{root: root}
	fail := func(err error) (*stagedFixtureMultipartPayloads, error) {
		staged.Close()
		return nil, err
	}
	seenPaths := make(map[string]struct{})
	for _, part := range operation.REST.Multipart.Parts {
		if part.Type != "file" {
			continue
		}
		value, present := fixtureRecordPathValue(record, part.Field)
		if !present || value == nil {
			if part.Required {
				return fail(fmt.Errorf("fixture multipart part %q is missing declared record field %q", part.Name, part.Field))
			}
			continue
		}
		raw, ok := value.(string)
		if !ok || strings.TrimSpace(raw) == "" {
			return fail(fmt.Errorf("fixture multipart part %q declared record field %q must be a non-empty file path", part.Name, part.Field))
		}
		if strings.TrimSpace(part.Field) == "" || strings.ContainsAny(part.Field, "/\\") {
			return fail(fmt.Errorf("fixture multipart part %q has an unsafe declared record field %q", part.Name, part.Field))
		}
		assetPath := path.Join("operations", operation.ID+".payloads", part.Field)
		if !fs.ValidPath(assetPath) {
			return fail(fmt.Errorf("fixture multipart part %q has invalid payload asset path", part.Name))
		}
		payload, err := readBoundedFixturePayload(b.Fixtures, assetPath, part.MaxBytes)
		if err != nil {
			return fail(fmt.Errorf("fixture multipart part %q: %w", part.Name, err))
		}
		rel := filepath.Clean(filepath.FromSlash(raw))
		if filepath.IsAbs(rel) || !filepath.IsLocal(rel) {
			return fail(fmt.Errorf("fixture multipart part %q destination must be project-relative", part.Name))
		}
		if _, duplicate := seenPaths[rel]; duplicate {
			return fail(fmt.Errorf("fixture multipart parts reuse source path %q", rel))
		}
		seenPaths[rel] = struct{}{}
		destination := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			return fail(fmt.Errorf("fixture multipart part %q create destination: %w", part.Name, err))
		}
		if err := os.WriteFile(destination, payload, 0o600); err != nil {
			return fail(fmt.Errorf("fixture multipart part %q stage payload: %w", part.Name, err))
		}
	}
	return staged, nil
}

func fixtureRecordPathValue(record connectors.Record, field string) (any, bool) {
	var current any = map[string]any(record)
	for _, segment := range strings.Split(field, ".") {
		if strings.TrimSpace(segment) == "" {
			return nil, false
		}
		values, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		value, ok := values[segment]
		if !ok {
			return nil, false
		}
		current = value
	}
	return current, true
}

func readBoundedFixturePayload(fsys fs.FS, name string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("declares no positive byte cap")
	}
	file, err := fsys.Open(name)
	if err != nil {
		return nil, fmt.Errorf("read declared payload asset %q: %w", name, err)
	}
	defer func() { _ = file.Close() }()
	payload, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxBytes {
		return nil, fmt.Errorf("payload asset exceeds declared byte cap %d", maxBytes)
	}
	return payload, nil
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
// incremental.request_param formatted per param_format/operator_prefix. A
// bundle with no incremental+fixtured stream at all is Skipped; a bundle whose
// ONLY incremental+fixtured candidate(s) carry a skip_dynamic marker Skips
// with that marker's reason instead (R3).
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

	wantParam, err := formatCursorForAssertion(maxCursor, stream.Incremental)
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

		func() {
			record := connectors.Record(fx.Record)
			ctx := context.Background()
			cfg := runtimeConfigForEngine(b)
			staged, stageErr := stageFixtureMultipartPayloads(b, action, record)
			if stageErr != nil {
				out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("fixture multipart payload: %v", stageErr)})
				return
			}
			if staged != nil {
				defer staged.Close()
				cfg.ProjectDir = staged.root
			}

			if verr := engine.ValidateWrite(ctx, b, writeRequestFor(action.Name, cfg), []connectors.Record{record}); verr != nil {
				out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("write_validate: fixture record failed validation: %v", verr)})
				return
			}

			capture.Reset(fx.Response)
			rb := withReplayURL(b, capture.URL)
			hooks := engine.HooksFor(b.Name)
			writeReq, err := approvedFixtureWriteRequest(ctx, rb, action.Name, cfg, []connectors.Record{record}, hooks)
			if err != nil {
				out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("engine.DryRunWrite against replay server failed: %v", err)})
				return
			}
			if _, err := engine.Write(ctx, rb, writeReq, []connectors.Record{record}, hooks); err != nil {
				out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("engine.Write against replay server failed: %v", err)})
				return
			}
			got := capture.LastRequest()
			if got == nil {
				out = append(out, CheckResult{Name: name, Error: "engine.Write sent no HTTP request"})
				return
			}

			if mismatch := compareWriteExpectation(*got, fx.Expect); mismatch != "" {
				out = append(out, CheckResult{Name: name, Error: mismatch})
				return
			}
			out = append(out, CheckResult{Name: name, Passed: true})
		}()
	}
	return out
}

func checkOperationMultipartRequestShape(b engine.Bundle) []CheckResult {
	var out []CheckResult
	capture := newCaptureServer(nil)
	defer capture.Close()
	for _, operation := range b.Operations {
		if operation.Kind != "rest_write" || operation.REST == nil || operation.REST.Multipart == nil {
			continue
		}
		name := "operation_multipart_request_shape:" + operation.ID
		fx, err := loadOperationFixture(b.Fixtures, operation.ID)
		if err != nil {
			out = append(out, CheckResult{Name: name, Skipped: true})
			continue
		}
		func() {
			record := connectors.Record(fx.Record)
			staged, stageErr := stageFixtureOperationMultipartPayloads(b, operation, record)
			if stageErr != nil {
				out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("fixture multipart payload: %v", stageErr)})
				return
			}
			if staged != nil {
				defer staged.Close()
			}
			rb := withReplayURL(b, capture.URL)
			cfg := runtimeConfigForEngine(rb)
			if staged != nil {
				cfg.ProjectDir = staged.root
			}
			capture.ResetMultipart(fx.Response, operation.REST.Multipart.MaxBytes)
			req, requestErr := approvedFixtureOperationRequest(context.Background(), rb, operation, cfg, record, engine.HooksFor(rb.Name))
			if requestErr != nil {
				out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("operation preview: %v", requestErr)})
				return
			}
			if _, err := engine.OperationDirectWrite(context.Background(), rb, req, engine.HooksFor(rb.Name)); err != nil {
				out = append(out, CheckResult{Name: name, Error: fmt.Sprintf("operation execution: %v", err)})
				return
			}
			got := capture.LastRequest()
			if got == nil {
				out = append(out, CheckResult{Name: name, Error: "operation execution sent no HTTP request"})
				return
			}
			if mismatch := compareWriteExpectation(*got, fx.Expect); mismatch != "" {
				out = append(out, CheckResult{Name: name, Error: mismatch})
				return
			}
			if mismatch := compareOperationMultipartCapture(*got, operation, record, req); mismatch != "" {
				out = append(out, CheckResult{Name: name, Error: mismatch})
				return
			}
			out = append(out, CheckResult{Name: name, Passed: true})
		}()
	}
	return out
}

func approvedFixtureOperationRequest(ctx context.Context, b engine.Bundle, operation engine.OperationSpec, cfg connectors.RuntimeConfig, record connectors.Record, hooks engine.Hooks) (connectors.OperationDirectWriteRequest, error) {
	req := connectors.OperationDirectWriteRequest{Operation: operation.ID, Config: cfg, Body: map[string]any(record), OutputPolicy: operation.OutputPolicy}
	digests, err := engine.ApprovedMultipartPayloadSHA256ForOperation(ctx, b, req, hooks)
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, fmt.Errorf("bind conformance fixture multipart payloads: %w", err)
	}
	req.Config.ApprovedPayloadSHA256 = digests
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	req.Config.CredentialRevision, err = authority.CredentialRevision("conformance:"+b.Name, req.Config.Secrets)
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	req.Config.ConfigurationDigest, err = authority.ConfigurationDigest("conformance:"+b.Name, req.Config.Config)
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	req.Config.WriteApprovalScope = connectors.WriteApprovalScopeFixture
	preview, err := engine.PreviewOperationDirectWrite(ctx, b, req, hooks)
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	if preview.ApprovalTarget.Confirmation.Kind == "" {
		return req, nil
	}
	planPayload, err := json.Marshal(struct {
		Connector string                   `json:"connector"`
		Operation string                   `json:"operation"`
		Config    connectors.RuntimeConfig `json:"config"`
		Body      map[string]any           `json:"body"`
	}{Connector: b.Name, Operation: operation.ID, Config: req.Config, Body: req.Body})
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, fmt.Errorf("marshal conformance operation plan: %w", err)
	}
	planHash := sha256.Sum256(planPayload)
	token := "conformance-fixture-operation-approval"
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: "rplan_conformance_operation_fixture", PlanHash: fmt.Sprintf("%x", planHash), PreviewDigest: preview.Digest,
		ApprovalToken: token, Target: preview.ApprovalTarget,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	req.Approval, err = authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID: grant.PlanID, PlanHash: grant.PlanHash, PreviewDigest: grant.PreviewDigest, ApprovalToken: token,
		Target: grant.Target, Confirmation: grant.Confirmation,
	})
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	req.PreviewDigest = preview.Digest
	return req, nil
}

// checkDeleteSemantics exercises every kind:delete write action's
// missing_ok_status handling: a status in that allow-list must be treated
// as an expected no-op, not a completed write or a failure. The real engine.Write is run against a server
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
	ctx := context.Background()
	hooks := engine.HooksFor(b.Name)
	writeReq, err := approvedFixtureWriteRequest(ctx, rb, deleteAction.Name, cfg, []connectors.Record{record}, hooks)
	if err != nil {
		return CheckResult{Name: name, Error: fmt.Sprintf("delete preview failed: %v", err)}
	}
	result, err := engine.Write(ctx, rb, writeReq, []connectors.Record{record}, hooks)
	if err != nil {
		return CheckResult{Name: name, Error: fmt.Sprintf("delete with missing_ok_status %d returned an error instead of being treated as written: %v", status, err)}
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 0 || result.RecordsUnchanged != 1 {
		return CheckResult{Name: name, Error: fmt.Sprintf("delete result = %+v, want RecordsWritten=0 RecordsFailed=0 RecordsUnchanged=1 for an allow-listed missing_ok_status %d", result, status)}
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
	if len(pages) > 0 && len(pages[0].ReadQuery) > 0 {
		req.Query = cloneStringMap(pages[0].ReadQuery)
	}
	return engine.Read(context.Background(), rb, req, engine.HooksFor(b.Name), func(r connectors.Record) error {
		return onRecord(map[string]any(r))
	})
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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

// formatCursorForAssertion mirrors read.go's unexported incremental
// formatting so this package can independently assert the re-read request
// shape without reaching into engine internals. This is a deliberate, small,
// documented duplication rather than a cross-package reach-in: the assertion
// is derived independently of the code under test.
func formatCursorForAssertion(value string, incremental *engine.IncrementalSpec) (string, error) {
	if incremental == nil {
		return value, nil
	}
	formatted, err := formatCursorValueForAssertion(value, incremental.ParamFormat)
	if err != nil {
		return "", err
	}
	return incremental.OperatorPrefix + formatted, nil
}

func formatCursorValueForAssertion(value, format string) (string, error) {
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
	case "rfc3339_utc":
		t, err := parseLowerBoundTimeForAssertion(value)
		if err != nil {
			return "", fmt.Errorf("param_format rfc3339_utc: %w", err)
		}
		return t.UTC().Format(time.RFC3339), nil
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

// writeFixture is fixtures/writes/<action>.json's shape (migration conventions §4):
// {"record": {...}, "expect": {"method","path","query"?,"body"?}}.
type writeFixture struct {
	Record   map[string]any   `json:"record"`
	Expect   writeExpectation `json:"expect"`
	Response *fixtureResponse `json:"response,omitempty"` // optional: what the capture server should answer with (see newCaptureServer)
}

type writeExpectation struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query,omitempty"`
	Body   map[string]any    `json:"body,omitempty"`
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

func loadOperationFixture(fixtures fs.FS, operation string) (writeFixture, error) {
	if fixtures == nil {
		return writeFixture{}, fmt.Errorf("bundle has no fixtures/ directory")
	}
	p := path.Join("operations", operation+".json")
	if !fs.ValidPath(p) {
		return writeFixture{}, fmt.Errorf("operation fixture path is invalid")
	}
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
	Method         string
	Path           string
	Query          url.Values
	Body           map[string]any
	Multipart      bool
	MultipartParts []capturedMultipartPart
	MultipartError string
}

type capturedMultipartPart struct {
	Name        string
	FileName    string
	ContentType string
	SHA256      string
	Bytes       int64
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
	for key, wantValue := range want.Query {
		if gotValue := got.Query.Get(key); gotValue != wantValue {
			return fmt.Sprintf("query %q = %q, want %q", key, gotValue, wantValue)
		}
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

func compareOperationMultipartCapture(got capturedRequest, operation engine.OperationSpec, record connectors.Record, req connectors.OperationDirectWriteRequest) string {
	if !got.Multipart {
		return "operation execution did not send multipart/form-data"
	}
	if got.MultipartError != "" {
		return "operation multipart capture: " + got.MultipartError
	}
	if operation.REST == nil || operation.REST.Multipart == nil {
		return "operation has no multipart declaration"
	}
	type expectedPart struct {
		file         bool
		fileName     string
		contentType  string
		allowedTypes []string
		digest       string
	}
	expected := make(map[string]expectedPart, len(operation.REST.Multipart.Parts))
	for _, part := range operation.REST.Multipart.Parts {
		value, present := fixtureRecordPathValue(record, part.Field)
		if !present || value == nil {
			if part.Required {
				return fmt.Sprintf("operation multipart required part %q is absent from fixture record", part.Name)
			}
			continue
		}
		if _, duplicate := expected[part.Name]; duplicate {
			return fmt.Sprintf("operation multipart declaration duplicates part %q", part.Name)
		}
		switch part.Type {
		case "field":
			expected[part.Name] = expectedPart{digest: conformanceMultipartValueSHA256(value)}
		case "file":
			filePath, ok := value.(string)
			if !ok || strings.TrimSpace(filePath) == "" {
				return fmt.Sprintf("operation multipart file part %q does not have a fixture path", part.Name)
			}
			digest := req.Config.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(0, part.Field)]
			if digest == "" {
				return fmt.Sprintf("operation multipart file part %q has no approved digest", part.Name)
			}
			expected[part.Name] = expectedPart{file: true, fileName: filepath.Base(filePath), contentType: part.ContentType, allowedTypes: part.AllowedMediaTypes, digest: digest}
		default:
			return fmt.Sprintf("operation multipart part %q has unsupported type %q", part.Name, part.Type)
		}
	}
	if len(got.MultipartParts) != len(expected) {
		return fmt.Sprintf("operation multipart part count = %d, want %d", len(got.MultipartParts), len(expected))
	}
	seen := make(map[string]struct{}, len(got.MultipartParts))
	for _, actual := range got.MultipartParts {
		want, ok := expected[actual.Name]
		if !ok {
			return fmt.Sprintf("operation multipart sent undeclared part %q", actual.Name)
		}
		if _, duplicate := seen[actual.Name]; duplicate {
			return fmt.Sprintf("operation multipart sent duplicate part %q", actual.Name)
		}
		seen[actual.Name] = struct{}{}
		if actual.SHA256 != want.digest {
			return fmt.Sprintf("operation multipart part %q did not preserve approved payload digest", actual.Name)
		}
		if want.file {
			if actual.FileName != want.fileName {
				return fmt.Sprintf("operation multipart file part %q filename = %q, want %q", actual.Name, actual.FileName, want.fileName)
			}
			if len(want.allowedTypes) == 0 {
				if !strings.EqualFold(strings.TrimSpace(actual.ContentType), strings.TrimSpace(want.contentType)) {
					return fmt.Sprintf("operation multipart file part %q content type does not match declaration", actual.Name)
				}
			} else if !capturedMultipartContentTypeAllowed(actual.ContentType, want.allowedTypes) {
				return fmt.Sprintf("operation multipart file part %q content type is not declared", actual.Name)
			}
			continue
		}
		if actual.FileName != "" {
			return fmt.Sprintf("operation multipart field part %q unexpectedly has a filename", actual.Name)
		}
	}
	return ""
}

func conformanceMultipartValueSHA256(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		raw = []byte(fmt.Sprint(value))
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		raw = []byte(text)
	}
	digest := sha256.Sum256(raw)
	return fmt.Sprintf("%x", digest[:])
}

func capturedMultipartContentTypeAllowed(raw string, declared []string) bool {
	actual, _, err := mime.ParseMediaType(strings.TrimSpace(raw))
	if err != nil || actual == "" {
		return false
	}
	actualParts := strings.Split(actual, "/")
	if len(actualParts) != 2 {
		return false
	}
	for _, value := range declared {
		allowed, _, err := mime.ParseMediaType(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		allowedParts := strings.Split(allowed, "/")
		if len(allowedParts) == 2 && strings.EqualFold(actualParts[0], allowedParts[0]) && (allowedParts[1] == "*" || strings.EqualFold(actualParts[1], allowedParts[1])) {
			return true
		}
	}
	return false
}

// --- capture / synthetic replay servers ------------------------------------

// captureServer is an httptest.Server that answers every request with a
// fixed response (200 {} by default, or a fixture-declared response — see
// newCaptureServer) and records the last request it received (method/path/
// query/decoded JSON body) for write_request_shape's assertions.
type captureServer struct {
	*httptest.Server
	mu                sync.Mutex
	resp              *fixtureResponse
	multipartMaxBytes int64
	last              *capturedRequest
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
		cs.mu.Lock()
		resp := cs.resp
		multipartMaxBytes := cs.multipartMaxBytes
		cs.mu.Unlock()
		body, multipartParts, multipartRequest, multipartErr := captureRequestBody(r, multipartMaxBytes)

		cs.mu.Lock()
		cs.last = &capturedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Body: body, Multipart: multipartRequest, MultipartParts: multipartParts, MultipartError: multipartErr}
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
	cs.multipartMaxBytes = 0
	cs.last = nil
}

func (cs *captureServer) ResetMultipart(resp *fixtureResponse, maxBytes int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.resp = resp
	cs.multipartMaxBytes = maxBytes
	cs.last = nil
}

func captureRequestBody(r *http.Request, multipartMaxBytes int64) (map[string]any, []capturedMultipartPart, bool, string) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if multipartMaxBytes > 0 && err == nil && strings.EqualFold(mediaType, "multipart/form-data") {
		parts, captureErr := captureMultipartParts(r.Body, params["boundary"], multipartMaxBytes)
		if captureErr != nil {
			return nil, nil, true, captureErr.Error()
		}
		return nil, parts, true, ""
	}
	var body map[string]any
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	_ = dec.Decode(&body)
	return body, nil, false, ""
}

func captureMultipartParts(body io.Reader, boundary string, maxBytes int64) ([]capturedMultipartPart, error) {
	if boundary == "" || maxBytes <= 0 {
		return nil, fmt.Errorf("invalid multipart capture bounds")
	}
	reader := multipart.NewReader(body, boundary)
	parts := make([]capturedMultipartPart, 0)
	var total int64
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := part.FormName()
		if name == "" {
			_ = part.Close()
			return nil, fmt.Errorf("multipart part has no field name")
		}
		remaining := maxBytes - total
		if remaining < 0 {
			_ = part.Close()
			return nil, fmt.Errorf("multipart payload exceeds declared byte cap")
		}
		digest := sha256.New()
		written, readErr := io.Copy(digest, io.LimitReader(part, remaining+1))
		closeErr := part.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		total += written
		if total > maxBytes {
			return nil, fmt.Errorf("multipart payload exceeds declared byte cap")
		}
		parts = append(parts, capturedMultipartPart{Name: name, FileName: part.FileName(), ContentType: part.Header.Get("Content-Type"), SHA256: fmt.Sprintf("%x", digest.Sum(nil)), Bytes: written})
	}
	return parts, nil
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
