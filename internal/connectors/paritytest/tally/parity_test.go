// Package paritytest_tally is tally's live-behavior proof for the ONE
// dynamic conformance check its bundle cannot exercise generically:
// cursor_advances (conventions.md §4's conformance-is-hook-aware /
// skip-marker rule — the identical marker mechanism applies here even
// though this is a Tier-1, no-hook bundle, since the gap is in the generic
// harness itself, not anything a hook could fix).
//
// checkCursorAdvances (internal/connectors/conformance/dynamic.go) re-reads
// an incremental+fixtured stream against a single httptest.Server that
// always answers 200 with an empty body, then asserts the re-read's
// captured query param matches the expected formatted lower bound. That
// harness has no fan_out awareness: tally's "submissions" stream fans out
// over every form id via a PRELIMINARY GET /forms id-listing request before
// its own per-form paginated read runs. Against the harness's
// always-empty-body capture server, the id-listing request itself finds
// zero form ids (an empty page has no "items"), so the per-form submissions
// sub-sequence — the ONLY request that would ever carry the declared
// startDate incremental request_param — is never issued at all; the
// harness's single captured-value slot would in any case only ever record
// whichever of the two request shapes happened to hit it last.
//
// This test proves the underlying behavior the generic harness cannot: a
// REAL two-phase fan_out read of "submissions", seeded with prior sync
// state (a cursor value), correctly issues BOTH the id-listing request and
// the per-form submissions request, and the latter carries
// startDate=<the seeded cursor value> exactly as streams.json's
// incremental.request_param declares.
package paritytest_tally

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func loadTallyBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "tally")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "tally", err)
	}
	return b
}

func withTallyBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// TestParityTally_SubmissionsFanOutSendsStartDateOnResumedSync is the
// authoritative substitute the submissions stream's skip_dynamic marker
// (streams.json) names.
func TestParityTally_SubmissionsFanOutSendsStartDateOnResumedSync(t *testing.T) {
	const seededCursor = "2026-01-09T00:00:00Z"

	var sawFormsRequest, sawSubmissionsRequest bool
	var gotStartDate string
	var gotFormsQueryPage, gotSubmissionsQueryPage string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/forms":
			sawFormsRequest = true
			gotFormsQueryPage = r.URL.Query().Get("page")
			fmt.Fprint(w, `{"items":[{"id":"form_fixture_1"}],"page":1,"limit":100,"total":1,"hasMore":false}`)
		case r.URL.Path == "/forms/form_fixture_1/submissions":
			sawSubmissionsRequest = true
			gotStartDate = r.URL.Query().Get("startDate")
			gotSubmissionsQueryPage = r.URL.Query().Get("page")
			fmt.Fprint(w, `{"page":1,"limit":100,"hasMore":false,"submissions":[]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	b := withTallyBaseURL(loadTallyBundle(t), srv.URL)
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "100"},
		Secrets: map[string]string{"api_key": "fixture-token"},
	}
	req := connectors.ReadRequest{
		Stream: "submissions",
		Config: cfg,
		State:  map[string]string{"cursor": seededCursor},
	}

	if err := engine.Read(context.Background(), b, req, engine.HooksFor(b.Name), func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("engine.Read(submissions): %v", err)
	}

	if !sawFormsRequest {
		t.Fatalf("fan_out id-listing GET /forms was never issued")
	}
	if gotFormsQueryPage != "1" {
		t.Fatalf("id-listing GET /forms page = %q, want \"1\"", gotFormsQueryPage)
	}
	if !sawSubmissionsRequest {
		t.Fatalf("per-form GET /forms/{id}/submissions was never issued")
	}
	if gotSubmissionsQueryPage != "1" {
		t.Fatalf("per-form submissions page = %q, want \"1\"", gotSubmissionsQueryPage)
	}
	if gotStartDate != seededCursor {
		t.Fatalf("per-form submissions startDate = %q, want %q (streams.json's incremental.request_param)", gotStartDate, seededCursor)
	}
}

// TestParityTally_SubmissionsFreshSyncOmitsStartDate proves the
// complementary case: a fresh full sync (no prior state, no start_date
// config) omits the startDate param entirely rather than sending an empty
// or zero-value one — the same "absent lower bound -> parameter omitted"
// contract every other incremental stream in this repo relies on.
func TestParityTally_SubmissionsFreshSyncOmitsStartDate(t *testing.T) {
	var gotStartDateValues []string
	var hadStartDateKey bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/forms":
			fmt.Fprint(w, `{"items":[{"id":"form_fixture_1"}],"page":1,"limit":100,"total":1,"hasMore":false}`)
		case r.URL.Path == "/forms/form_fixture_1/submissions":
			if _, ok := r.URL.Query()["startDate"]; ok {
				hadStartDateKey = true
				gotStartDateValues = append(gotStartDateValues, r.URL.Query().Get("startDate"))
			}
			fmt.Fprint(w, `{"page":1,"limit":100,"hasMore":false,"submissions":[]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	b := withTallyBaseURL(loadTallyBundle(t), srv.URL)
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "100"},
		Secrets: map[string]string{"api_key": "fixture-token"},
	}
	req := connectors.ReadRequest{Stream: "submissions", Config: cfg, State: nil}

	if err := engine.Read(context.Background(), b, req, engine.HooksFor(b.Name), func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("engine.Read(submissions): %v", err)
	}

	if hadStartDateKey {
		t.Fatalf("fresh full sync sent startDate=%v, want the param omitted entirely (no prior cursor, no start_date config)", gotStartDateValues)
	}
}
