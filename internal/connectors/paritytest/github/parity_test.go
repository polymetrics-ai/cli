package githubparity_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	githublegacy "polymetrics.ai/internal/connectors/github"
	_ "polymetrics.ai/internal/connectors/hooks/github" // registers the AuthHook/WriteHook via init()
)

// This file is the migration parity suite for the github bundle (PLAN.md
// P-9, wave1-pilot): the Tier-2 AuthHook (github_app JWT->installation-token
// exchange) + WriteHook (compound writes) pilot, and the only pilot with
// real writes (24 legacy write actions). Both the legacy hand-written
// github.Connector (internal/connectors/github, read-only reference) and
// the engine-backed connector (engine.New(bundle, engine.HooksFor("github")))
// are driven against the SAME httptest.Server; RAW reflect.DeepEqual record
// equality (after a UseNumber-normalizing round trip) is the parity bar for
// reads, and method/path/decoded-body equality is the bar for writes. This is
// the red-first test: it fails to even compile/load until defs/github and
// hooks/github exist.

func loadGithubBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("engine.LoadAll(defs.FS): %v", err)
	}
	for _, b := range bundles {
		if b.Name == "github" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "github", names)
	return engine.Bundle{}
}

func withGithubBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func newGithubEngineConnector(b engine.Bundle) connectors.Connector {
	return engine.New(b, engine.HooksFor("github"))
}

// githubRuntimeConfig builds a RuntimeConfig usable by BOTH connectors:
// legacy reads a single "repository" ("owner/repo") config value (or "repo"
// alias) while the engine bundle reads separate "owner"/"repo" keys (see
// docs.md Known limits — InterpolatePath urlencodes each {{ }} reference as
// one opaque path segment, so a combined value can't be split
// declaratively); setting all three keeps both sides pointed at the same
// fixture repository.
func githubRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":   baseURL,
		"repository": "octocat/hello-world",
		"owner":      "octocat",
		"repo":       "hello-world",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"token": "fixture_token_placeholder"},
	}
}

func readAllGithubRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
	t.Helper()
	var out []connectors.Record
	if err := c.Read(context.Background(), req, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	}); err != nil {
		t.Fatalf("Read(%s): %v", req.Stream, err)
	}
	return out
}

// normalizeRecord re-encodes r through encoding/json with UseNumber so
// legacy's native Go types and the engine's json.Number-preserving decode
// compare equal on numeric fields (mirrors parity_stripe_test.go's helper).
func normalizeRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode record: %v", err)
	}
	return out
}

func normalizeRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeRecord(t, r)
	}
	return out
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// --- bundle smoke guard ----------------------------------------------------

func TestParityGithub_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadGithubBundle(t)

	wantStreams := []string{
		"branches", "collaborators", "commits", "contributors", "deployments",
		"issue_comments", "issues", "labels", "milestones", "pull_request_review_comments",
		"pull_requests", "releases", "repository", "stargazers", "subscribers",
		"tags", "workflow_artifacts", "workflow_runs", "workflows",
	}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v (%d), want %v (%d)", gotStreams, len(gotStreams), wantStreams, len(wantStreams))
	}

	wantWrites := []string{
		"cancel_workflow_run", "close_issue", "close_pull_request", "comment_issue",
		"create_issue", "create_label", "create_milestone", "create_or_update_file",
		"create_pull_request", "create_pull_request_review", "create_release",
		"delete_file", "delete_label", "delete_milestone", "delete_release",
		"delete_workflow_run", "dispatch_workflow", "merge_pull_request",
		"request_reviewers", "rerun_workflow_run", "update_issue", "update_label",
		"update_milestone", "update_pull_request", "update_release",
	}
	gotWrites := make([]string, 0, len(bundle.Writes))
	for _, w := range bundle.Writes {
		gotWrites = append(gotWrites, w.Name)
	}
	sort.Strings(gotWrites)
	if !reflect.DeepEqual(gotWrites, wantWrites) {
		t.Fatalf("bundle write actions = %v (%d), want %v (%d)", gotWrites, len(gotWrites), wantWrites, len(wantWrites))
	}

	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true")
	}
}

// --- per-stream record parity ---------------------------------------------

// githubStreamServer answers every stream path used by the representative
// parity subset below with a fixed page (or 2 pages for issues, exercising
// page_number pagination + the pull_request filter simultaneously). Both
// legacy and engine hit this SAME server.
func githubStreamServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/octocat/hello-world", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"id":1296269,"node_id":"MDEwOlJlcG9zaXRvcnkxMjk2MjY5","name":"hello-world","full_name":"octocat/hello-world","private":false,"description":"Fixture.","html_url":"https://github.com/octocat/hello-world","default_branch":"main","language":"Go","stargazers_count":42,"watchers_count":42,"forks_count":7,"open_issues_count":3,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z","pushed_at":"2026-01-02T00:00:00Z"}`)
	})

	mux.HandleFunc("/repos/octocat/hello-world/issues", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "", "1":
			writeJSON(w, `[
				{"id":1,"node_id":"I1","number":101,"state":"open","title":"First","body":"b1","html_url":"https://github.com/octocat/hello-world/issues/101","url":"https://api.github.com/repos/octocat/hello-world/issues/101","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T01:00:00Z"},
				{"id":2,"node_id":"I2","number":102,"state":"open","title":"PR-shaped","body":"b2","html_url":"https://github.com/octocat/hello-world/pull/102","url":"https://api.github.com/repos/octocat/hello-world/issues/102","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"pull_request":{"url":"https://api.github.com/repos/octocat/hello-world/pulls/102"},"created_at":"2026-01-01T02:00:00Z","updated_at":"2026-01-01T02:00:00Z"}
			]`)
		case "2":
			writeJSON(w, `[{"id":3,"node_id":"I3","number":103,"state":"open","title":"Second page","body":"b3","html_url":"https://github.com/octocat/hello-world/issues/103","url":"https://api.github.com/repos/octocat/hello-world/issues/103","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":1,"locked":false,"created_at":"2026-01-02T00:00:00Z","updated_at":"2026-01-02T01:00:00Z"}]`)
		default:
			writeJSON(w, `[]`)
		}
	})

	mux.HandleFunc("/repos/octocat/hello-world/pulls", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `[{"id":501,"node_id":"PR1","number":301,"state":"open","title":"Fixture PR","body":"pr body","html_url":"https://github.com/octocat/hello-world/pull/301","url":"https://api.github.com/repos/octocat/hello-world/pulls/301","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"created_at":"2026-01-03T00:00:00Z","updated_at":"2026-01-03T01:00:00Z","merged_at":null,"draft":false,"merge_commit_sha":null,"base":{"ref":"main","sha":"abc"},"head":{"ref":"feat","sha":"def"}}]`)
	})

	mux.HandleFunc("/repos/octocat/hello-world/actions/workflows", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"total_count":1,"workflows":[{"id":1201,"node_id":"W1","name":"CI","path":".github/workflows/ci.yml","state":"active","badge_url":"https://github.com/octocat/hello-world/workflows/CI/badge.svg","html_url":"https://github.com/octocat/hello-world/actions/workflows/ci.yml","created_at":"2026-01-09T00:00:00Z","updated_at":"2026-01-09T01:00:00Z"}]}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestParityGithub_StreamRecords(t *testing.T) {
	bundle := loadGithubBundle(t)
	streams := []string{"repository", "issues", "pull_requests", "workflows"}

	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := githubStreamServer(t)

			legacy := githublegacy.New()
			legacyRecs := readAllGithubRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: githubRuntimeConfig(srv.URL, nil)})

			eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
			engRecs := readAllGithubRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: githubRuntimeConfig(srv.URL, nil)})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy github emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeRecords(t, engRecs)
			wantNorm := normalizeRecords(t, legacyRecs)
			for i := range wantNorm {
				// Compare only the fields BOTH sides are expected to emit
				// (the engine bundle documents a small, ledgered set of
				// dropped/omitted fields — labels_count/assignees_count/
				// is_pull_request/assets_count — that legacy emits but the
				// dialect cannot express; see docs.md Known limits and
				// ledger items). Every remaining field must match exactly.
				for key, legacyVal := range wantNorm[i] {
					if isDocumentedDrop(key) {
						continue
					}
					engVal, ok := gotNorm[i][key]
					if !ok {
						// Schema-projection ("schema" mode, projectRecord)
						// only copies a raw key when it is PRESENT on the raw
						// record; legacy's hand-written *Record functions
						// unconditionally write every declared field
						// (item["x"] on a Go map miss is nil), so a raw
						// record that never carries an optional field (e.g.
						// state_reason/closed_at absent rather than
						// JSON-null) surfaces as legacy=nil vs engine=absent.
						// Both mean "no value" to any downstream consumer;
						// treat legacy nil + engine absent as equivalent
						// rather than a mismatch.
						if legacyVal == nil {
							continue
						}
						t.Fatalf("stream %q record %d missing field %q in engine output (legacy=%v)", stream, i, key, legacyVal)
					}
					if !reflect.DeepEqual(engVal, legacyVal) {
						t.Fatalf("stream %q record %d field %q mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, key, engVal, legacyVal)
					}
				}
			}
		})
	}
}

// TestParityGithub_NestedIDComputedFieldsEmitNativeNumbers pins G0b's
// RESOLVED state (gap-loop cycle-1 engine mini-wave item 1/REVIEW-A.md
// adjudication A1: typed computed_fields extraction). user_id/author_id/
// committer_id/workflow_run_id are all sourced via a BARE single
// "{{ record.<path> }}" computed_fields template (e.g. "{{ record.user.id
// }}", no filter, no surrounding literal text), so the engine now copies
// the raw JSON value straight through instead of stringifying it via
// Interpolate — these fields must be native json.Number, matching legacy's
// own raw-JSON-passthrough numeric type exactly (RAW equality, not the
// string-form-only comparison isStringifiedNestedID used to require before
// this engine increment landed).
func TestParityGithub_NestedIDComputedFieldsEmitNativeNumbers(t *testing.T) {
	bundle := loadGithubBundle(t)
	srv := githubStreamServer(t)

	eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
	engRecs := readAllGithubRecords(t, eng, connectors.ReadRequest{Stream: "issues", Config: githubRuntimeConfig(srv.URL, nil)})
	if len(engRecs) == 0 {
		t.Fatal("engine emitted zero issues records (test fixture bug)")
	}

	got := normalizeRecord(t, engRecs[0])["user_id"]
	n, ok := got.(json.Number)
	if !ok {
		t.Fatalf("engine issues[0].user_id = %#v (%T), want json.Number (native type, typed computed_fields extraction)", got, got)
	}
	if n.String() != "1" {
		t.Fatalf("engine issues[0].user_id = %q, want %q", n.String(), "1")
	}
}

// TestParityGithub_RepositoryMarkerFieldRestored pins G0's RESOLVED state
// (gap-loop cycle-1 engine mini-wave: config.* now resolves inside
// computed_fields; see docs/migration/conventions.md §3 "config.* in
// computed_fields"). Legacy stamps req.Config.Config["repository"] (a
// single "owner/repo" config value) onto EVERY emitted record of EVERY
// stream; this bundle now reproduces it via a computed_fields
// "{{ config.owner }}/{{ config.repo }}" template. Asserted directly
// (not folded into the generic per-field loop in TestParityGithub_
// StreamRecords, since the string this bundle derives it from --
// "owner"+"repo" -- differs from legacy's single "repository" config key,
// even though both resolve to the identical wire value for this fixture).
func TestParityGithub_RepositoryMarkerFieldRestored(t *testing.T) {
	bundle := loadGithubBundle(t)
	streams := []string{"repository", "issues", "pull_requests", "workflows"}

	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := githubStreamServer(t)

			legacy := githublegacy.New()
			legacyRecs := readAllGithubRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: githubRuntimeConfig(srv.URL, nil)})

			eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
			engRecs := readAllGithubRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: githubRuntimeConfig(srv.URL, nil)})

			if len(legacyRecs) == 0 || len(engRecs) == 0 {
				t.Fatalf("zero records for stream %q (test fixture bug)", stream)
			}

			wantRepo, ok := legacyRecs[0]["repository"].(string)
			if !ok || wantRepo == "" {
				t.Fatalf("legacy stream %q record 0 repository = %#v, want non-empty string (test fixture bug)", stream, legacyRecs[0]["repository"])
			}
			if wantRepo != "octocat/hello-world" {
				t.Fatalf("legacy stream %q record 0 repository = %q, want %q (test fixture bug)", stream, wantRepo, "octocat/hello-world")
			}

			gotRepo, ok := engRecs[0]["repository"].(string)
			if !ok {
				t.Fatalf("engine stream %q record 0 repository = %#v (%T), want string %q", stream, engRecs[0]["repository"], engRecs[0]["repository"], wantRepo)
			}
			if gotRepo != wantRepo {
				t.Fatalf("engine stream %q record 0 repository = %q, want %q (legacy)", stream, gotRepo, wantRepo)
			}
		})
	}
}

// isDocumentedDrop names the record fields legacy emits that this bundle's
// dialect genuinely cannot express (no count/length filter) — see docs.md
// Known limits. Anything else must match exactly. "repository" was
// previously here (ENGINE_GAP G0: computed_fields could not reference
// config.*) but the gap-loop cycle-1 engine mini-wave closed G0 (config.*
// now resolves inside computed_fields, secrets.* stays excluded) — the
// marker is restored via "{{ config.owner }}/{{ config.repo }}" (see
// TestParityGithub_RepositoryMarkerFieldRestored) and must no longer be
// silently stripped from the comparison.
func isDocumentedDrop(field string) bool {
	switch field {
	case "labels_count", "assignees_count", "is_pull_request", "assets_count":
		return true
	default:
		return false
	}
}

// TestParityGithub_IssuesPaginationFiltersOutPullRequests exercises
// page_number pagination (2 full pages of the bundle's fixed page_size=100)
// AND the pull_request-filter simultaneously: page 1 has 100 real issues
// plus one PR-shaped item (filtered), page 2 has one more real issue. The
// bundle's page_size is a fixed 100 (not runtime-configurable — see docs.md
// Known limits), so a genuine multi-page proof needs a full 100-item first
// page, matching exactly what fixtures/streams/issues/{page_1,page_2}.json
// already commits for conformance's pagination_terminates check; this test
// reuses that same shape against a shared httptest.Server for both
// connectors. Legacy is configured with max_pages=all (its own default,
// max_pages=1, is a disclosed deviation — see docs.md Known limits) so both
// sides exhibit the SAME effective pagination behavior.
func TestParityGithub_IssuesPaginationFiltersOutPullRequests(t *testing.T) {
	bundle := loadGithubBundle(t)
	srv := githubTwoPageIssuesServer(t)

	legacy := githublegacy.New()
	legacyRecs := readAllGithubRecords(t, legacy, connectors.ReadRequest{Stream: "issues", Config: githubRuntimeConfig(srv.URL, map[string]string{"max_pages": "all"})})

	eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
	engRecs := readAllGithubRecords(t, eng, connectors.ReadRequest{Stream: "issues", Config: githubRuntimeConfig(srv.URL, nil)})

	if len(legacyRecs) != 101 {
		t.Fatalf("legacy issues records = %d, want 101 (100 page-1 issues + 1 page-2 issue; PR filtered on both pages, test fixture bug)", len(legacyRecs))
	}
	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("engine issues records = %d, want %d (legacy)", len(engRecs), len(legacyRecs))
	}

	gotIDs := recordIDs(t, engRecs)
	wantIDs := recordIDs(t, legacyRecs)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("issues record id sequence mismatch (engine vs legacy)")
	}
}

// githubTwoPageIssuesServer builds a 100-record page 1 (+ 1 PR-shaped item
// filtered out) and a 2-record page 2 (1 real issue + 1 PR-shaped item
// filtered out), matching fixtures/streams/issues/{page_1,page_2}.json's
// shape so the SAME page_size=100 short-page-stop rule genuinely exercises
// two requests on both connectors.
func githubTwoPageIssuesServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/octocat/hello-world/issues", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "", "1":
			var items []string
			for i := 1; i <= 100; i++ {
				items = append(items, fmt.Sprintf(`{"id":%d,"node_id":"I%d","number":%d,"state":"open","title":"Issue %d","body":"b","html_url":"https://github.com/octocat/hello-world/issues/%d","url":"https://api.github.com/repos/octocat/hello-world/issues/%d","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T01:00:00Z"}`, i, i, 100+i, i, 100+i, 100+i))
			}
			items = append(items, `{"id":9001,"node_id":"IPR","number":9101,"state":"open","title":"PR-shaped","body":"b","html_url":"https://github.com/octocat/hello-world/pull/9101","url":"https://api.github.com/repos/octocat/hello-world/issues/9101","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"pull_request":{"url":"https://api.github.com/repos/octocat/hello-world/pulls/9101"},"created_at":"2026-01-01T02:00:00Z","updated_at":"2026-01-01T02:00:00Z"}`)
			writeJSON(w, "["+strings.Join(items, ",")+"]")
		case "2":
			writeJSON(w, `[
				{"id":101,"node_id":"I101","number":201,"state":"open","title":"Page 2 issue","body":"b","html_url":"https://github.com/octocat/hello-world/issues/201","url":"https://api.github.com/repos/octocat/hello-world/issues/201","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"created_at":"2026-01-02T00:00:00Z","updated_at":"2026-01-02T01:00:00Z"},
				{"id":9002,"node_id":"IPR2","number":9102,"state":"open","title":"PR-shaped page 2","body":"b","html_url":"https://github.com/octocat/hello-world/pull/9102","url":"https://api.github.com/repos/octocat/hello-world/issues/9102","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"pull_request":{"url":"https://api.github.com/repos/octocat/hello-world/pulls/9102"},"created_at":"2026-01-02T02:00:00Z","updated_at":"2026-01-02T02:00:00Z"}
			]`)
		default:
			writeJSON(w, `[]`)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func recordIDs(t *testing.T, recs []connectors.Record) []string {
	t.Helper()
	out := make([]string, len(recs))
	for i, r := range recs {
		out[i] = normalizeRecord(t, r)["id"].(json.Number).String()
	}
	return out
}

// --- since-param incremental parity (docs.md Known limits / ledger G15) --

// sinceCaptureServer answers the issues stream endpoint, recording the
// "since" query param value it received (empty string if absent), and
// returns one fixture issue.
func sinceCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotSince string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSince = r.URL.Query().Get("since")
		writeJSON(w, `[{"id":1,"node_id":"I1","number":101,"state":"open","title":"Fixture","body":"b","html_url":"https://github.com/octocat/hello-world/issues/101","url":"https://api.github.com/repos/octocat/hello-world/issues/101","user":{"login":"octocat","id":1},"author_association":"OWNER","comments":0,"locked":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T01:00:00Z"}]`)
	}))
	t.Cleanup(srv.Close)
	return srv, &gotSince
}

// TestParityGithub_SinceConfigOnlyMatchesLegacy asserts the CONFIG path:
// both connectors forward config.since to the "since" query param
// identically when no app-persisted state cursor is present — legacy's own
// (and only) "since" source (github.go Read: config "since" only, never
// req.State).
func TestParityGithub_SinceConfigOnlyMatchesLegacy(t *testing.T) {
	bundle := loadGithubBundle(t)
	const configSince = "2026-01-01T00:00:00Z"

	legacySrv, legacyGot := sinceCaptureServer(t)
	legacy := githublegacy.New()
	_ = readAllGithubRecords(t, legacy, connectors.ReadRequest{
		Stream: "issues",
		Config: githubRuntimeConfig(legacySrv.URL, map[string]string{"since": configSince}),
	})
	if *legacyGot != configSince {
		t.Fatalf("legacy since query param = %q, want %q (test fixture bug)", *legacyGot, configSince)
	}

	engSrv, engGot := sinceCaptureServer(t)
	eng := newGithubEngineConnector(withGithubBaseURL(bundle, engSrv.URL))
	_ = readAllGithubRecords(t, eng, connectors.ReadRequest{
		Stream: "issues",
		Config: githubRuntimeConfig(engSrv.URL, map[string]string{"since": configSince}),
	})
	if *engGot != *legacyGot {
		t.Fatalf("engine since query param = %q, want %q (legacy, config-only path)", *engGot, *legacyGot)
	}
}

// TestParityGithub_SinceStateCursorForwardingIsEngineOnlyBehavior asserts
// the STATE path: the engine forwards an app-persisted state cursor as
// "since" (engine/read.go's incrementalLowerBoundValue prefers req.State
// over start_config_key — engine-wide semantics, matches every other
// incremental stream in the fleet), while legacy IGNORES req.State entirely
// (github.go's Read only ever consults req.Config.Config["since"] — greps
// clean, no req.State reference anywhere in the legacy package). This is a
// documented, deliberate IMPROVEMENT (a sync with persisted state emits a
// smaller, correctly-incremental record set on the engine side) — not a
// parity bug — but docs.md must not claim "matches legacy exactly" for the
// state path, and this test pins the actual (diverging, by design) behavior
// so a future change to either side is caught.
func TestParityGithub_SinceStateCursorForwardingIsEngineOnlyBehavior(t *testing.T) {
	bundle := loadGithubBundle(t)
	const stateCursor = "2026-02-02T00:00:00Z"

	legacySrv, legacyGot := sinceCaptureServer(t)
	legacy := githublegacy.New()
	_ = readAllGithubRecords(t, legacy, connectors.ReadRequest{
		Stream: "issues",
		Config: githubRuntimeConfig(legacySrv.URL, nil), // no config "since" set either
		State:  map[string]string{"cursor": stateCursor},
	})
	if *legacyGot != "" {
		t.Fatalf("legacy since query param = %q, want empty (legacy never reads req.State — test fixture bug if this fails)", *legacyGot)
	}

	engSrv, engGot := sinceCaptureServer(t)
	eng := newGithubEngineConnector(withGithubBaseURL(bundle, engSrv.URL))
	_ = readAllGithubRecords(t, eng, connectors.ReadRequest{
		Stream: "issues",
		Config: githubRuntimeConfig(engSrv.URL, nil),
		State:  map[string]string{"cursor": stateCursor},
	})
	if *engGot != stateCursor {
		t.Fatalf("engine since query param = %q, want %q (documented engine-only state-cursor-forwarding behavior, see docs.md Known limits)", *engGot, stateCursor)
	}
}

// --- AuthHook parity: github_app JWT -> installation-token exchange ------

// TestParityGithub_AuthGithubAppInstallationTokenBearerHeader asserts BOTH
// connectors send the SAME "Authorization: Bearer <installation token>"
// header on a read request when configured for auth_type=github_app,
// exercising the shared installation-token-exchange double from both sides.
func TestParityGithub_AuthGithubAppInstallationTokenBearerHeader(t *testing.T) {
	bundle := loadGithubBundle(t)
	privKey := testRSAKeyPEM(t)
	const installationToken = "ghs_shared_fixture_installation_token"

	var legacyAuthHeader, engAuthHeader string

	mux := http.NewServeMux()
	mux.HandleFunc("/app/installations/67890/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"token":"`+installationToken+`"}`)
	})
	mux.HandleFunc("/repos/octocat/hello-world", func(w http.ResponseWriter, r *http.Request) {
		if legacyAuthHeader == "" {
			legacyAuthHeader = r.Header.Get("Authorization")
		} else {
			engAuthHeader = r.Header.Get("Authorization")
		}
		writeJSON(w, `{"id":1,"node_id":"n","name":"hello-world","full_name":"octocat/hello-world","private":false,"html_url":"https://github.com/octocat/hello-world","default_branch":"main","stargazers_count":0,"watchers_count":0,"forks_count":0,"open_issues_count":0,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	appCfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url": srv.URL, "repository": "octocat/hello-world",
			"owner": "octocat", "repo": "hello-world",
			"app_id": "12345", "installation_id": "67890",
		},
		Secrets: map[string]string{"private_key": privKey},
	}

	legacy := githublegacy.New()
	_ = readAllGithubRecords(t, legacy, connectors.ReadRequest{Stream: "repository", Config: appCfg})

	eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
	_ = readAllGithubRecords(t, eng, connectors.ReadRequest{Stream: "repository", Config: appCfg})

	want := "Bearer " + installationToken
	if legacyAuthHeader != want {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", legacyAuthHeader, want)
	}
	if engAuthHeader != legacyAuthHeader {
		t.Fatalf("engine Authorization = %q, want %q (legacy)", engAuthHeader, legacyAuthHeader)
	}
}

// TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic pins
// the fixed no-silent-fallthrough behavior (docs.md Known limits / ledger
// G14): a config with NEITHER a token secret NOR github_app config NOR the
// explicit public_access opt-in must hard-error, not silently resolve to an
// unauthenticated read. Legacy's own "auto" resolution (auth.go:73-80) DOES
// fall through to public in this exact shape (including for a caller who set
// an alias-shaped secret legacy tolerates — personalAccessToken, etc. — but
// not "token" itself) — this bundle is a documented, intentional
// STRICTER-than-legacy deviation (F4: never silently fail open) closing the
// alias/typo hazard REVIEW-A.md's major flagged. A caller that genuinely
// wants unauthenticated reads must opt in with public_access (see
// TestParityGithub_AuthExplicitPublicOptIn); the engine's `when` grammar has
// no statically-validated string-equality/membership check today (F9,
// connectorgen validate's ResolveCheck only parses bare namespace.key
// truthiness), so this bundle cannot reproduce legacy's exact
// auth_type=public/none/anonymous/unauthenticated string-value selection —
// see docs.md Known limits for the documented scope narrowing.
func TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic(t *testing.T) {
	bundle := loadGithubBundle(t)
	srv := githubStreamServer(t)

	noCredsCfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url": srv.URL,
			"owner":    "octocat",
			"repo":     "hello-world",
		},
		// No Secrets at all — not even the alias-shaped ones legacy also
		// tolerates (personalAccessToken, etc.) — this is a "no credentials
		// were ever configured" config, the exact shape that must never
		// silently reach GitHub unauthenticated.
	}

	eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
	err := eng.Read(context.Background(), connectors.ReadRequest{Stream: "repository", Config: noCredsCfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("engine Read with no credentials and no public_access opt-in = nil error, want a hard failure (must not silently fall through to unauthenticated mode:none)")
	}
}

// TestParityGithub_AuthExplicitPublicOptIn asserts a caller can still
// explicitly opt into unauthenticated reads via public_access — the
// mode:none candidate is gated ({{ config.public_access }} truthiness), not
// an unconditional catch-all, but remains reachable for the
// legacy-documented "public" auth mode (auth.go's githubAuthPublic) via this
// bundle's dedicated opt-in key (see docs.md Known limits: legacy's
// auth_type=public string-value selection itself is not reproduced 1:1).
func TestParityGithub_AuthExplicitPublicOptIn(t *testing.T) {
	bundle := loadGithubBundle(t)
	srv := githubStreamServer(t)

	publicCfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":      srv.URL,
			"owner":         "octocat",
			"repo":          "hello-world",
			"public_access": "true",
		},
	}

	eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
	var recs []connectors.Record
	err := eng.Read(context.Background(), connectors.ReadRequest{Stream: "repository", Config: publicCfg}, func(r connectors.Record) error {
		recs = append(recs, r)
		return nil
	})
	if err != nil {
		t.Fatalf("engine Read with explicit public_access opt-in: %v (want success, opt-in unauthenticated read)", err)
	}
	if len(recs) == 0 {
		t.Fatal("engine Read with explicit public_access opt-in emitted zero records (test fixture bug)")
	}
}

// TestParityGithub_AuthTypePublicEnumOptIn is the S3 engine mini-wave item 2
// restoration (wave1-pilot SUMMARY.md carried queue / REVIEW-A.md re-review
// R1 CONDITION): now that engine.ResolveCheckWhen statically validates the
// FULL when-grammar (==/in/truthiness), legacy's exact
// `auth_type=public|none|anonymous|unauthenticated` string-enum selection
// (auth.go:81-82) is restored as an ADDITIONAL, purely-additive opt-in
// alongside `public_access` (never a replacement — `public_access` stays the
// primary documented surface; EvalWhen has no `||` operator, so the two
// opt-ins are expressed as two separate `mode:none` candidates in the auth
// list rather than one OR'd condition). Every legacy synonym is asserted.
func TestParityGithub_AuthTypePublicEnumOptIn(t *testing.T) {
	bundle := loadGithubBundle(t)

	for _, synonym := range []string{"public", "none", "anonymous", "unauthenticated"} {
		synonym := synonym
		t.Run(synonym, func(t *testing.T) {
			srv := githubStreamServer(t)
			cfg := connectors.RuntimeConfig{
				Config: map[string]string{
					"base_url":  srv.URL,
					"owner":     "octocat",
					"repo":      "hello-world",
					"auth_type": synonym,
				},
			}

			eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
			var recs []connectors.Record
			err := eng.Read(context.Background(), connectors.ReadRequest{Stream: "repository", Config: cfg}, func(r connectors.Record) error {
				recs = append(recs, r)
				return nil
			})
			if err != nil {
				t.Fatalf("engine Read with auth_type=%q: %v (want success, legacy string-enum opt-in restored)", synonym, err)
			}
			if len(recs) == 0 {
				t.Fatalf("engine Read with auth_type=%q emitted zero records (test fixture bug)", synonym)
			}
		})
	}
}

// TestParityGithub_AuthTypeUnrelatedValueDoesNotGrantPublicAccess proves the
// auth_type enum restoration is NARROW (only the 4 legacy public synonyms
// arm mode:none) — an unrelated/garbage auth_type value must still hard-error
// exactly like TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic,
// not silently fall through to unauthenticated reads for every possible
// auth_type string.
func TestParityGithub_AuthTypeUnrelatedValueDoesNotGrantPublicAccess(t *testing.T) {
	bundle := loadGithubBundle(t)
	srv := githubStreamServer(t)

	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"owner":     "octocat",
			"repo":      "hello-world",
			"auth_type": "some_other_value",
		},
	}

	eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
	err := eng.Read(context.Background(), connectors.ReadRequest{Stream: "repository", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("engine Read with an unrelated auth_type value = nil error, want a hard failure (only the 4 legacy public synonyms arm mode:none)")
	}
}

func testRSAKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	der := x509.MarshalPKCS1PrivateKey(key)
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}))
}

// --- write parity: parity floor actions -----------------------------------

type writeCaptureRequest struct {
	Method string
	Path   string
	Body   map[string]any
}

// writeCaptureServer answers every request 200 with response (or {"number":N}
// style bodies as needed per-action) and records EVERY request received, in
// order (compound actions issue more than one).
func writeCaptureServer(t *testing.T, response string) (*httptest.Server, *[]writeCaptureRequest) {
	t.Helper()
	var reqs []writeCaptureRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		reqs = append(reqs, writeCaptureRequest{Method: r.Method, Path: r.URL.Path, Body: body})
		w.Header().Set("Content-Type", "application/json")
		if response == "" {
			response = "{}"
		}
		writeJSON(w, response)
	}))
	t.Cleanup(srv.Close)
	return srv, &reqs
}

func runWriteParity(t *testing.T, action string, record connectors.Record, response string) (legacyReqs, engReqs []writeCaptureRequest) {
	t.Helper()

	legacySrv, legacyGot := writeCaptureServer(t, response)
	legacy := githublegacy.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{Action: action, Config: githubRuntimeConfig(legacySrv.URL, nil)}, []connectors.Record{record}); err != nil {
		t.Fatalf("legacy Write(%s): %v", action, err)
	}

	bundle := loadGithubBundle(t)
	engSrv, engGot := writeCaptureServer(t, response)
	eng := newGithubEngineConnector(withGithubBaseURL(bundle, engSrv.URL))
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{Action: action, Config: githubRuntimeConfig(engSrv.URL, nil)}, []connectors.Record{record}); err != nil {
		t.Fatalf("engine Write(%s): %v", action, err)
	}
	return *legacyGot, *engGot
}

func TestParityGithub_WriteCreateIssue(t *testing.T) {
	record := connectors.Record{"title": "Fixture issue", "body": "Fixture body"}
	legacyReqs, engReqs := runWriteParity(t, "create_issue", record, `{"number":1}`)

	if len(legacyReqs) != 1 || len(engReqs) != 1 {
		t.Fatalf("request counts = legacy:%d engine:%d, want 1 and 1", len(legacyReqs), len(engReqs))
	}
	if legacyReqs[0].Method != http.MethodPost || legacyReqs[0].Path != "/repos/octocat/hello-world/issues" {
		t.Fatalf("legacy request = %+v (test fixture bug)", legacyReqs[0])
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

func TestParityGithub_WriteUpdateIssue(t *testing.T) {
	record := connectors.Record{"issue_number": 101, "title": "Updated", "state": "closed"}
	legacyReqs, engReqs := runWriteParity(t, "update_issue", record, "")

	if legacyReqs[0].Method != http.MethodPatch || legacyReqs[0].Path != "/repos/octocat/hello-world/issues/101" {
		t.Fatalf("legacy request = %+v (test fixture bug)", legacyReqs[0])
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

func TestParityGithub_WriteCommentIssue(t *testing.T) {
	record := connectors.Record{"issue_number": 101, "body": "Fixture comment"}
	legacyReqs, engReqs := runWriteParity(t, "comment_issue", record, "")

	if legacyReqs[0].Method != http.MethodPost || legacyReqs[0].Path != "/repos/octocat/hello-world/issues/101/comments" {
		t.Fatalf("legacy request = %+v (test fixture bug)", legacyReqs[0])
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

// TestParityGithub_WriteCreatePullRequestCompound is the compound-write bar:
// both connectors must issue the SAME sequence of requests (create POST,
// issue-metadata PATCH, reviewers POST) for a record carrying labels AND
// reviewers — method, path, AND decoded body, matching every non-compound
// write test's bar (S3 engine mini-wave carried minor — SUMMARY.md carried
// minors: "github compound-write test still compares method/path only
// (bodies compared in all non-compound tests)"). Body comparison was not a
// gap in engine BEHAVIOR — it was a gap in this test's own assertion
// strength; flipping to full reflect.DeepEqual is a strict strengthening.
func TestParityGithub_WriteCreatePullRequestCompound(t *testing.T) {
	record := connectors.Record{
		"head": "feature-1", "base": "main", "title": "Fixture PR",
		"labels": []any{"bug"}, "reviewers": []any{"octocat"},
	}
	legacyReqs, engReqs := runWriteParity(t, "create_pull_request", record, `{"number":301}`)

	if len(legacyReqs) != 3 {
		t.Fatalf("legacy request count = %d, want 3 (create, metadata, reviewers) (test fixture bug): %+v", len(legacyReqs), legacyReqs)
	}
	if len(engReqs) != len(legacyReqs) {
		t.Fatalf("engine request count = %d, want %d (legacy): engine=%+v legacy=%+v", len(engReqs), len(legacyReqs), engReqs, legacyReqs)
	}
	for i := range legacyReqs {
		if !reflect.DeepEqual(engReqs[i], legacyReqs[i]) {
			t.Fatalf("request %d = %+v, want %+v (legacy)", i, engReqs[i], legacyReqs[i])
		}
	}
}

func TestParityGithub_WriteMergePullRequest(t *testing.T) {
	record := connectors.Record{"pull_number": 301, "merge_method": "squash"}
	legacyReqs, engReqs := runWriteParity(t, "merge_pull_request", record, "")

	if legacyReqs[0].Method != http.MethodPut || legacyReqs[0].Path != "/repos/octocat/hello-world/pulls/301/merge" {
		t.Fatalf("legacy request = %+v (test fixture bug)", legacyReqs[0])
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

// TestParityGithub_WriteCreateLabelStripsLeadingHashFromColor pins the
// restored '#'-strip normalization (docs.md Known limits / ledger G16):
// legacy's githubCreateLabelPayload does
// strings.TrimPrefix(color, "#") (github.go:1120) before sending, so a
// caller-supplied "#ff0000" (a value GitHub's own docs/UI commonly show
// with the leading hash, and legacy explicitly accepts+normalizes) must
// reach the wire as "ff0000" on BOTH connectors, not the raw "#ff0000"
// (which GitHub's API 422s on).
func TestParityGithub_WriteCreateLabelStripsLeadingHashFromColor(t *testing.T) {
	record := connectors.Record{"name": "bug", "color": "#ff0000"}
	legacyReqs, engReqs := runWriteParity(t, "create_label", record, `{"id":1,"name":"bug","color":"ff0000"}`)

	if legacyReqs[0].Method != http.MethodPost || legacyReqs[0].Path != "/repos/octocat/hello-world/labels" {
		t.Fatalf("legacy request = %+v (test fixture bug)", legacyReqs[0])
	}
	if got := legacyReqs[0].Body["color"]; got != "ff0000" {
		t.Fatalf("legacy request body color = %#v, want %q (test fixture bug: legacy's TrimPrefix(color, \"#\"))", got, "ff0000")
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

// TestParityGithub_WriteCreateLabelColorWithoutHashUnaffected asserts a
// caller-supplied color with NO leading '#' (GitHub's own wire shape) is
// unaffected by the strip on both sides — the normalization is a no-op for
// an already-bare hex string.
func TestParityGithub_WriteCreateLabelColorWithoutHashUnaffected(t *testing.T) {
	record := connectors.Record{"name": "bug", "color": "ff0000"}
	legacyReqs, engReqs := runWriteParity(t, "create_label", record, `{"id":1,"name":"bug","color":"ff0000"}`)

	if got := legacyReqs[0].Body["color"]; got != "ff0000" {
		t.Fatalf("legacy request body color = %#v, want %q (test fixture bug)", got, "ff0000")
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

// TestParityGithub_WriteUpdateLabelStripsLeadingHashFromColor is
// create_label's sibling case for update_label — legacy's
// githubUpdateLabelPayload does the identical TrimPrefix (github.go:1133),
// but ONLY when a color field is actually present (update_label's color is
// optional, "at least one mutable field" shape, ledger G3).
func TestParityGithub_WriteUpdateLabelStripsLeadingHashFromColor(t *testing.T) {
	record := connectors.Record{"name": "bug", "color": "#00ff00"}
	legacyReqs, engReqs := runWriteParity(t, "update_label", record, "")

	if legacyReqs[0].Method != http.MethodPatch || legacyReqs[0].Path != "/repos/octocat/hello-world/labels/bug" {
		t.Fatalf("legacy request = %+v (test fixture bug)", legacyReqs[0])
	}
	if got := legacyReqs[0].Body["color"]; got != "00ff00" {
		t.Fatalf("legacy request body color = %#v, want %q (test fixture bug: legacy's TrimPrefix(color, \"#\"))", got, "00ff00")
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

// TestParityGithub_WriteUpdateLabelNoColorFieldOmitsColor asserts
// update_label's "at least one mutable field" shape: a record with no
// color field at all must not send a "color" body key on either side (the
// strip normalization must not invent a color key that was never present).
func TestParityGithub_WriteUpdateLabelNoColorFieldOmitsColor(t *testing.T) {
	record := connectors.Record{"name": "bug", "new_name": "bug2"}
	legacyReqs, engReqs := runWriteParity(t, "update_label", record, "")

	if _, ok := legacyReqs[0].Body["color"]; ok {
		t.Fatalf("legacy request body = %+v, want no \"color\" key (test fixture bug)", legacyReqs[0].Body)
	}
	if !reflect.DeepEqual(engReqs[0], legacyReqs[0]) {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

// TestParityGithub_WriteDeleteLabel asserts both connectors send the SAME
// DELETE request and both succeed against a 204 response. Legacy has NO
// idempotent/missing_ok delete semantics (any non-2xx, including 404, is a
// hard failure in legacy's doJSONWithAuth) — this bundle intentionally does
// NOT declare delete.missing_ok_status either, matching legacy exactly (see
// docs.md Known limits / ledger).
func TestParityGithub_WriteDeleteLabel(t *testing.T) {
	record := connectors.Record{"name": "bug"}
	legacyReqs, engReqs := runWriteParity(t, "delete_label", record, "")

	if legacyReqs[0].Method != http.MethodDelete || legacyReqs[0].Path != "/repos/octocat/hello-world/labels/bug" {
		t.Fatalf("legacy request = %+v (test fixture bug)", legacyReqs[0])
	}
	if engReqs[0].Method != legacyReqs[0].Method || engReqs[0].Path != legacyReqs[0].Path {
		t.Fatalf("engine request = %+v, want %+v (legacy)", engReqs[0], legacyReqs[0])
	}
}

// TestParityGithub_WriteDeleteLabelNotFoundFailsOnBothSides asserts a 404 on
// delete_label is a hard FAILURE on both connectors (legacy has no
// missing_ok/idempotent-delete semantics; this bundle deliberately matches
// that rather than adding new engine leniency legacy never had).
func TestParityGithub_WriteDeleteLabelNotFoundFailsOnBothSides(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	record := connectors.Record{"name": "bug"}

	legacy := githublegacy.New()
	_, legacyErr := legacy.Write(context.Background(), connectors.WriteRequest{Action: "delete_label", Config: githubRuntimeConfig(srv.URL, nil)}, []connectors.Record{record})
	if legacyErr == nil {
		t.Fatal("legacy Write(delete_label) with 404 response = nil error, want a failure (test fixture bug: legacy has no idempotent-delete semantics)")
	}

	bundle := loadGithubBundle(t)
	eng := newGithubEngineConnector(withGithubBaseURL(bundle, srv.URL))
	_, engErr := eng.Write(context.Background(), connectors.WriteRequest{Action: "delete_label", Config: githubRuntimeConfig(srv.URL, nil)}, []connectors.Record{record})
	if engErr == nil {
		t.Fatal("engine Write(delete_label) with 404 response = nil error, want a failure (matches legacy: no missing_ok_status declared)")
	}
}

// --- manifest-surface parity ------------------------------------------------

func TestParityGithub_ManifestSurface(t *testing.T) {
	bundle := loadGithubBundle(t)

	legacyManifest := connectors.ManifestOf(githublegacy.New())
	eng := newGithubEngineConnector(bundle)
	engManifest := connectors.ManifestOf(eng)

	wantStreams := manifestStreamNames(legacyManifest.Streams)
	gotStreams := manifestStreamNames(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream names = %v, want %v (legacy)", gotStreams, wantStreams)
	}

	wantWrites := writeActionNames(legacyManifest.WriteActions)
	gotWrites := writeActionNames(engManifest.WriteActions)
	if !reflect.DeepEqual(gotWrites, wantWrites) {
		t.Fatalf("write action names = %v, want %v (legacy)", gotWrites, wantWrites)
	}
}

func manifestStreamNames(streams []connectors.Stream) []string {
	out := make([]string, len(streams))
	for i, s := range streams {
		out[i] = s.Name
	}
	sort.Strings(out)
	return out
}

func writeActionNames(actions []connectors.WriteActionSpec) []string {
	out := make([]string, len(actions))
	for i, a := range actions {
		out[i] = a.Name
	}
	sort.Strings(out)
	return out
}
