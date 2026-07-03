package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// widgetsSchema compiles a minimal record schema with a primary key and
// cursor field, used across read.go tests as the "declared" projection
// target.
func widgetsSchema(t *testing.T) *StreamSchema {
	t.Helper()
	raw := json.RawMessage(`{
		"type": "object",
		"x-primary-key": ["id"],
		"x-cursor-field": "updated_at",
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"},
			"updated_at": {"type": "string"}
		}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	return &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}
}

// newTestBundle builds a minimal Bundle wired against srv, with a single
// "widgets" stream described by streamOverride (Name/Records/etc pre-filled
// with sane defaults the caller may override).
func newTestBundle(t *testing.T, srv *httptest.Server, stream StreamSpec) Bundle {
	t.Helper()
	if stream.Name == "" {
		stream.Name = "widgets"
	}
	if stream.Path == "" {
		stream.Path = "/widgets"
	}
	if stream.Records.Path == "" {
		stream.Records.Path = "data"
	}
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{
			URL: srv.URL,
		},
		Streams: []StreamSpec{stream},
		Schemas: map[string]*StreamSchema{
			stream.Name: widgetsSchema(t),
		},
	}
}

func readAll(t *testing.T, ctx context.Context, b Bundle, req connectors.ReadRequest, h Hooks) ([]connectors.Record, error) {
	t.Helper()
	var out []connectors.Record
	err := Read(ctx, b, req, h, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	})
	return out, err
}

func jsonServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// --- initial query construction ---

func TestReadStaticQuery(t *testing.T) {
	var gotQuery string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Query: map[string]QueryParam{"sort": {Template: "asc"}, "state": {Template: "all"}}})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotQuery != "sort=asc&state=all" {
		t.Fatalf("query = %q, want sort=asc&state=all", gotQuery)
	}
}

// --- optional-query dialect (gap-loop item 3, REVIEW-B.md adjudication 2) ---

// TestReadOptionalQueryOmittedWhenConfigAbsent proves an object-form query
// entry with omit_when_absent:true is silently left off the request when its
// referenced config key is unset — the vitally `status`/bitly
// `page_size` shape a plain string entry cannot express without a hard
// error.
func TestReadOptionalQueryOmittedWhenConfigAbsent(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"status": {Template: "{{ config.status }}", OmitWhenAbsent: true},
		},
	})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v (an omit_when_absent entry referencing an absent config key must never hard-error)", err)
	}
	if _, present := gotQuery["status"]; present {
		t.Fatalf("query = %v, want status param OMITTED when config.status is unset", gotQuery)
	}
}

// TestReadOptionalQuerySentWhenConfigPresent proves the SAME object-form
// entry sends the resolved value normally once the config key IS set —
// omit_when_absent only tolerates ABSENCE, it never suppresses a genuinely
// present value.
func TestReadOptionalQuerySentWhenConfigPresent(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"status": {Template: "{{ config.status }}", OmitWhenAbsent: true},
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"status": "active"}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotQuery.Get("status") != "active" {
		t.Fatalf("query = %v, want status=active", gotQuery)
	}
}

// TestReadOptionalQueryDefaultAppliedWhenConfigAbsent proves the `default`
// field closes the calendly page_size gap: an object-form entry with a
// default literal sends that default when the referenced config key is
// absent, instead of hard-erroring OR omitting the param entirely.
func TestReadOptionalQueryDefaultAppliedWhenConfigAbsent(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"count": {Template: "{{ config.page_size }}", Default: "100"},
		},
	})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotQuery.Get("count") != "100" {
		t.Fatalf("query = %v, want count=100 (default applied)", gotQuery)
	}
}

// TestReadOptionalQueryDefaultOverriddenByConfig proves a `default`-bearing
// entry still resolves the templated config value normally when it IS set —
// default is a fallback, never a fixed override.
func TestReadOptionalQueryDefaultOverriddenByConfig(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"count": {Template: "{{ config.page_size }}", Default: "100"},
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "25"}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotQuery.Get("count") != "25" {
		t.Fatalf("query = %v, want count=25 (config value overrides default)", gotQuery)
	}
}

// TestReadQueryStringEntryStillHardErrorsOnAbsentConfig locks in that a
// PLAIN STRING query entry (the pre-existing dialect shape, zero migration
// risk) keeps today's exact hard-error semantics — no blanket
// absent-key-falsy tolerance leaks into string entries.
func TestReadQueryStringEntryStillHardErrorsOnAbsentConfig(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"status": {Template: "{{ config.status }}"},
		},
	})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want hard error for a plain string query entry referencing an absent config key, got nil")
	}
}

// --- spec.json default materialization (gap-loop item 6, REVIEW-A.md C3) ---

// TestReadBaseURLDefaultMaterializedWhenConfigAbsent is the RED test for C3:
// a spec.json property carrying a "default" (e.g. base_url:
// "https://api.example.com") must be materialized into RuntimeConfig.Config
// when the caller's config omits that key entirely — closing the
// hard-error-on-every-legacy-default-shaped-config gap C3 names (github
// api.github.com, gmail base+token URLs, monday api.monday.com/v2, etc).
func TestReadBaseURLDefaultMaterializedWhenConfigAbsent(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	spec, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"base_url": {"type": "string", "default": "` + srv.URL + `"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.Spec = spec
	b.HTTP.URL = "{{ config.base_url }}"

	// No base_url in config at all — must resolve via spec.json's default,
	// not hard-error "unresolved key base_url in config".
	_, err = readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v (spec.json default for base_url must be materialized into config)", err)
	}
}

// TestReadConfigDefaultDoesNotOverrideExplicitValue proves a materialized
// default never clobbers a value the caller DID supply — defaults only fill
// genuinely absent keys.
func TestReadConfigDefaultDoesNotOverrideExplicitValue(t *testing.T) {
	var gotHost string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	otherSrv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	spec, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"base_url": {"type": "string", "default": "` + otherSrv.URL + `"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.Spec = spec
	b.HTTP.URL = "{{ config.base_url }}"

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}}}
	_, err = readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotHost == "" {
		t.Fatalf("request never reached the explicitly-configured server — default incorrectly took priority")
	}
}

// TestCheckBaseURLDefaultMaterializedWhenConfigAbsent proves Check's config
// path materializes spec defaults too (same single mechanism, not just
// Read's).
func TestCheckBaseURLDefaultMaterializedWhenConfigAbsent(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	spec, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"base_url": {"type": "string", "default": "` + srv.URL + `"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b := Bundle{
		Name: "acme",
		Spec: spec,
		HTTP: HTTPBase{URL: "{{ config.base_url }}", Check: &RequestSpec{Method: "GET", Path: "/ping"}},
	}

	err = Check(context.Background(), b, connectors.RuntimeConfig{}, nil)
	if err != nil {
		t.Fatalf("Check: %v (spec.json default for base_url must be materialized into config)", err)
	}
}

func TestReadIncrementalLowerBoundFromStateCursor(t *testing.T) {
	var gotSince string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotSince = r.URL.Query().Get("since")
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "since", ParamFormat: "rfc3339"},
	})

	req := connectors.ReadRequest{Stream: "widgets", State: map[string]string{"cursor": "2026-01-01T00:00:00Z"}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotSince != "2026-01-01T00:00:00Z" {
		t.Fatalf("since = %q, want 2026-01-01T00:00:00Z", gotSince)
	}
}

func TestReadIncrementalLowerBoundFallsBackToStartConfigKey(t *testing.T) {
	var gotSince string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotSince = r.URL.Query().Get("since")
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "since", ParamFormat: "rfc3339", StartConfigKey: "start_date"},
	})

	req := connectors.ReadRequest{
		Stream: "widgets",
		Config: connectors.RuntimeConfig{Config: map[string]string{"start_date": "2025-06-01T00:00:00Z"}},
	}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotSince != "2025-06-01T00:00:00Z" {
		t.Fatalf("since = %q, want 2025-06-01T00:00:00Z", gotSince)
	}
}

func TestReadIncrementalParamFormats(t *testing.T) {
	cases := []struct {
		name      string
		format    string
		cursor    string
		wantParam string
	}{
		{"rfc3339", "rfc3339", "2026-03-01T12:00:00Z", "2026-03-01T12:00:00Z"},
		{"unix_seconds", "unix_seconds", "2026-03-01T12:00:00Z", "1772366400"},
		{"date", "date", "2026-03-01T12:00:00Z", "2026-03-01"},
		{"github_date_range", "github_date_range", "2026-03-01T12:00:00Z", ">=2026-03-01T12:00:00Z"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
				got = r.URL.Query().Get("since")
				_, _ = w.Write([]byte(`{"data":[]}`))
			})
			b := newTestBundle(t, srv, StreamSpec{
				Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "since", ParamFormat: tc.format},
			})
			req := connectors.ReadRequest{Stream: "widgets", State: map[string]string{"cursor": tc.cursor}}
			_, err := readAll(t, context.Background(), b, req, nil)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if got != tc.wantParam {
				t.Fatalf("since = %q, want %q", got, tc.wantParam)
			}
		})
	}
}

// --- incremental.lower_bound query-var dialect (S3 engine mini-wave item 1,
// wave1-pilot SUMMARY.md carried queue / REVIEW-A.md re-review R2): expose the
// RESOLVED (state cursor or start_config_key, post-formatParam) incremental
// lower bound to stream.Query template resolution, so a param legacy sends
// ONLY when the incremental lower bound resolves (chargebee's
// sort_by[asc]=updated_at, chargebee.go:151-155) is expressible via the
// existing optional-query dialect instead of requiring a new engine
// primitive. ---

// TestReadIncrementalLowerBoundQueryVarOmittedOnFreshFullSync proves a
// stream.Query entry templated as "{{ incremental.lower_bound }}" with
// omit_when_absent:true is left off the request entirely on a fresh full
// sync (no state cursor, no start_config_key value) — incremental.lower_bound
// resolves to "" (genuinely absent), and the dialect's existing
// omit_when_absent semantics apply to that absence exactly like a
// config/secrets absence.
func TestReadIncrementalLowerBoundQueryVarOmittedOnFreshFullSync(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"sort_by[asc]": {Template: "{{ incremental.lower_bound }}", OmitWhenAbsent: true},
		},
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "updated_at[after]", ParamFormat: "unix_seconds", StartConfigKey: "start_date"},
	})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v (an omit_when_absent incremental.lower_bound entry must never hard-error on a fresh full sync)", err)
	}
	if _, present := gotQuery["sort_by[asc]"]; present {
		t.Fatalf("query = %v, want sort_by[asc] OMITTED when no incremental lower bound resolves", gotQuery)
	}
}

// TestReadIncrementalLowerBoundQueryVarSentFromStateCursor proves the SAME
// entry sends a literal value once a state cursor resolves the lower bound —
// this is exactly the state-cursor-driven repeat-sync path the S-2 STOP
// identified as inexpressible (a config.start_date-gated entry would get
// this path wrong, since a persisted State["cursor"] is not a config key at
// all).
func TestReadIncrementalLowerBoundQueryVarSentFromStateCursor(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"sort_by[asc]": {Template: "{{ incremental.lower_bound }}", OmitWhenAbsent: true},
		},
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "updated_at[after]", ParamFormat: "unix_seconds", StartConfigKey: "start_date"},
	})

	req := connectors.ReadRequest{Stream: "widgets", State: map[string]string{"cursor": "1700000100"}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotQuery.Get("sort_by[asc]") != "1700000100" {
		t.Fatalf("query = %v, want sort_by[asc]=1700000100 (the RESOLVED, post-formatParam lower bound value, from the state cursor)", gotQuery)
	}
}

// TestReadIncrementalLowerBoundQueryVarSentFromStartConfigKey proves the same
// dialect resolves from the start_config_key fallback (fresh sync, config
// start_date set) and formats the value per the stream's own param_format —
// the value seen by the query var is the SAME resolved/formatted value sent
// as the request_param, not the raw unformatted cursor.
func TestReadIncrementalLowerBoundQueryVarSentFromStartConfigKey(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"sort_by[asc]": {Template: "{{ incremental.lower_bound }}", OmitWhenAbsent: true},
		},
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "updated_at[after]", ParamFormat: "unix_seconds", StartConfigKey: "start_date"},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"start_date": "2023-11-14T22:15:00Z"}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotQuery.Get("sort_by[asc]") != "1700000100" {
		t.Fatalf("query = %v, want sort_by[asc]=1700000100 (start_date formatted as unix_seconds, matching the request_param's own formatted value)", gotQuery)
	}
	if gotQuery.Get("updated_at[after]") != gotQuery.Get("sort_by[asc]") {
		t.Fatalf("sort_by[asc] (%q) and updated_at[after] (%q) must carry the identical formatted lower-bound value", gotQuery.Get("sort_by[asc]"), gotQuery.Get("updated_at[after]"))
	}
}

// TestReadIncrementalLowerBoundQueryVarAbsentIncrementalSpecOmits proves a
// stream with NO incremental spec at all still resolves
// "{{ incremental.lower_bound }}" as genuinely absent (never a hard error,
// never a literal "incremental.lower_bound"-shaped string) — the query var
// is defined even for a non-incremental stream, it just always resolves
// empty there.
func TestReadIncrementalLowerBoundQueryVarAbsentIncrementalSpecOmits(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"sort_by[asc]": {Template: "{{ incremental.lower_bound }}", OmitWhenAbsent: true},
		},
	})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if _, present := gotQuery["sort_by[asc]"]; present {
		t.Fatalf("query = %v, want sort_by[asc] OMITTED for a stream with no incremental spec at all", gotQuery)
	}
}

// --- B1 (REVIEW.md BLOCK): formatParam must accept a digits-only
// (unix-seconds) cursor input as well as RFC3339, since that is the ONLY
// shape internal/app actually ever persists for a numeric cursor field
// (internal/app/sync_modes.go recordCursor -> toComparableString stringifies
// a json.Number/int64 verbatim, never converting it to RFC3339). ---

func TestFormatParamUnixSecondsAcceptsDigitsPassthrough(t *testing.T) {
	got, err := formatParam("1700000100", "unix_seconds")
	if err != nil {
		t.Fatalf("formatParam(digits, unix_seconds) error = %v, want digits-passthrough (matches legacy verbatim-forward semantics)", err)
	}
	if got != "1700000100" {
		t.Fatalf("formatParam(digits, unix_seconds) = %q, want %q (verbatim)", got, "1700000100")
	}
}

func TestFormatParamUnixSecondsStillAcceptsRFC3339(t *testing.T) {
	got, err := formatParam("2026-03-01T12:00:00Z", "unix_seconds")
	if err != nil {
		t.Fatalf("formatParam(rfc3339, unix_seconds) error = %v", err)
	}
	if got != "1772366400" {
		t.Fatalf("formatParam(rfc3339, unix_seconds) = %q, want 1772366400", got)
	}
}

func TestFormatParamDateAcceptsDigitsPassthrough(t *testing.T) {
	// 1700000100 unix seconds = 2023-11-14T22:15:00Z -> date "2023-11-14".
	got, err := formatParam("1700000100", "date")
	if err != nil {
		t.Fatalf("formatParam(digits, date) error = %v, want the digits interpreted as unix seconds then formatted", err)
	}
	if got != "2023-11-14" {
		t.Fatalf("formatParam(digits, date) = %q, want 2023-11-14", got)
	}
}

func TestFormatParamGithubDateRangeAcceptsDigitsPassthrough(t *testing.T) {
	got, err := formatParam("1700000100", "github_date_range")
	if err != nil {
		t.Fatalf("formatParam(digits, github_date_range) error = %v, want the digits interpreted as unix seconds then formatted as an RFC3339 lower bound", err)
	}
	if got != ">=2023-11-14T22:15:00Z" {
		t.Fatalf("formatParam(digits, github_date_range) = %q, want >=2023-11-14T22:15:00Z", got)
	}
}

func TestFormatParamRFC3339PassesThroughVerbatimUnaffected(t *testing.T) {
	got, err := formatParam("2026-03-01T12:00:00Z", "rfc3339")
	if err != nil {
		t.Fatalf("formatParam(rfc3339, rfc3339) error = %v", err)
	}
	if got != "2026-03-01T12:00:00Z" {
		t.Fatalf("formatParam(rfc3339, rfc3339) = %q, want verbatim passthrough", got)
	}
}

// --- S4 engine mini-wave item 5: date-only lower bounds (marketstack) ------
//
// parseLowerBoundTime accepted only all-digits (Unix seconds) or strict
// RFC3339. Marketstack's real wire cursor value for its date param_format
// streams (eod/splits/dividends) is a bare YYYY-MM-DD string with no
// time/offset component at all — neither shape parses it. Tried order:
// digits -> RFC3339 -> date-only, so no shape masks another (a valid RFC3339
// string is never all-digits; a valid date-only string is never
// RFC3339-parseable).

func TestParseLowerBoundTimeAcceptsDateOnly(t *testing.T) {
	got, err := parseLowerBoundTime("2026-01-02")
	if err != nil {
		t.Fatalf("parseLowerBoundTime(date-only) error = %v, want a bare YYYY-MM-DD value to parse", err)
	}
	want := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parseLowerBoundTime(date-only) = %v, want %v (midnight UTC that date)", got, want)
	}
}

func TestFormatParamDateAcceptsDateOnlyLowerBound(t *testing.T) {
	got, err := formatParam("2026-01-02", "date")
	if err != nil {
		t.Fatalf("formatParam(date-only, date) error = %v", err)
	}
	if got != "2026-01-02" {
		t.Fatalf("formatParam(date-only, date) = %q, want round-trip to the same date-only string %q", got, "2026-01-02")
	}
}

func TestFormatParamUnixSecondsAcceptsDateOnlyLowerBound(t *testing.T) {
	got, err := formatParam("2026-01-02", "unix_seconds")
	if err != nil {
		t.Fatalf("formatParam(date-only, unix_seconds) error = %v", err)
	}
	want := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).Unix()
	if got != strconv.FormatInt(want, 10) {
		t.Fatalf("formatParam(date-only, unix_seconds) = %q, want %d", got, want)
	}
}

func TestFormatParamGithubDateRangeAcceptsDateOnlyLowerBound(t *testing.T) {
	got, err := formatParam("2026-01-02", "github_date_range")
	if err != nil {
		t.Fatalf("formatParam(date-only, github_date_range) error = %v", err)
	}
	if got != ">=2026-01-02T00:00:00Z" {
		t.Fatalf("formatParam(date-only, github_date_range) = %q, want %q", got, ">=2026-01-02T00:00:00Z")
	}
}

func TestParseLowerBoundTimeRejectsMalformedDateOnly(t *testing.T) {
	for _, bad := range []string{"2026-13-40", "2026/01/02", "not-a-date"} {
		if _, err := parseLowerBoundTime(bad); err == nil {
			t.Fatalf("parseLowerBoundTime(%q) error = nil, want an error (malformed value must never silently misparse)", bad)
		}
	}
}

// toComparableStringLikeApp is a deliberate, LOCAL copy of
// internal/app/sync_modes.go's toComparableString (read-only reference; the
// engine package must not import internal/app per PLAN.md). It exists only so
// this test can prove the exact app-persisted cursor SHAPE without importing
// app, per the task's "mimic recordCursor stringification" instruction.
func toComparableStringLikeApp(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		if v == float64(int64(v)) {
			return strconv_FormatInt(int64(v))
		}
		return strconv_FormatFloat(v)
	default:
		return fmt.Sprint(v)
	}
}

func strconv_FormatInt(v int64) string {
	return fmt.Sprintf("%d", v)
}

func strconv_FormatFloat(v float64) string {
	return fmt.Sprintf("%v", v)
}

// TestReadAppLevelCursorRoundTrip is the honest B1 parity bar: read a stream
// whose "created" cursor field arrives as a numeric (json.Number) wire value
// (Stripe's real shape), derive the persisted state cursor EXACTLY the way
// internal/app/sync_modes.go's recordCursor does (stringify the raw
// json.Number verbatim, never converting to RFC3339), feed that cursor back
// into a second Read as req.State["cursor"], and assert the resumed read
// sends the correct unix-seconds lower-bound query value. Before the B1 fix
// this fails at the second Read call with formatParam's RFC3339-parse error.
func TestReadAppLevelCursorRoundTrip(t *testing.T) {
	var gotSince []string
	page := 0
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotSince = append(gotSince, r.URL.Query().Get("created_gte"))
		page++
		switch page {
		case 1:
			_, _ = w.Write([]byte(`{"data":[{"id":"cus_1","name":"a","updated_at":1700000000},{"id":"cus_2","name":"b","updated_at":1700000100}]}`))
		default:
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:     RecordsSpec{Path: "data"},
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "created_gte", ParamFormat: "unix_seconds"},
	})
	// widgetsSchema declares "updated_at" as a string property; the wire value
	// here is numeric, matching connsdk's json.Number-preserving decode of a
	// real Stripe-shaped "created" field. Schema validation is not invoked on
	// read (only on write), so this is safe.

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read (first, full sync): %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("records = %+v, want 2", recs)
	}

	// Derive the persisted cursor exactly the way internal/app does: the MAX
	// across emitted records of recordCursor(record, "updated_at") stringified
	// via toComparableString. Both records here carry a json.Number
	// "updated_at" (connsdk decodes with UseNumber), so this reproduces the
	// real "1700000100" persisted-cursor shape from B1's failure scenario.
	persisted := ""
	for _, r := range recs {
		cursor := toComparableStringLikeApp(r["updated_at"])
		if persisted == "" || cursor > persisted {
			persisted = cursor
		}
	}
	if persisted != "1700000100" {
		t.Fatalf("persisted cursor = %q, want %q (app-persisted unix-seconds string shape)", persisted, "1700000100")
	}

	// Resume: second Read with the app-persisted cursor must succeed AND send
	// the correct unix-seconds value (itself, verbatim) as created_gte.
	page = 0
	gotSince = nil
	_, err = readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets", State: map[string]string{"cursor": persisted}}, nil)
	if err != nil {
		t.Fatalf("Read (resume with app-persisted cursor %q): %v (B1: formatParam must accept digits-only unix_seconds input)", persisted, err)
	}
	if len(gotSince) == 0 || gotSince[0] != "1700000100" {
		t.Fatalf("resume created_gte = %v, want first request to carry %q", gotSince, "1700000100")
	}
}

// --- records extraction ---

func TestReadRecordsPathDot(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}]`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "."}})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["id"] != "1" {
		t.Fatalf("records = %+v", recs)
	}
}

func TestReadSingleObject(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"1","name":"solo","updated_at":"2026-01-01T00:00:00Z"}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: ".", SingleObject: true}})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["id"] != "1" {
		t.Fatalf("records = %+v", recs)
	}
}

// --- filters ---

func TestReadFilterFieldAbsent(t *testing.T) {
	// github issues-vs-PRs: a record carrying "pull_request" is excluded.
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[
			{"id":"1","name":"issue","updated_at":"2026-01-01T00:00:00Z"},
			{"id":"2","name":"pr","updated_at":"2026-01-01T00:00:00Z","pull_request":{"url":"x"}}
		]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data", Filter: &FilterSpec{FieldAbsent: "pull_request"}}})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["id"] != "1" {
		t.Fatalf("records = %+v, want only the issue record", recs)
	}
}

func TestReadFilterFieldEquals(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[
			{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z","status":"open"},
			{"id":"2","name":"b","updated_at":"2026-01-01T00:00:00Z","status":"closed"}
		]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data", Filter: &FilterSpec{FieldEquals: map[string]any{"status": "open"}}}})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["id"] != "1" {
		t.Fatalf("records = %+v, want only status=open", recs)
	}
}

// --- projection ---

func TestReadProjectionSchemaDropsUndeclaredFields(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z","secret_internal":"x"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("records = %+v", recs)
	}
	if _, ok := recs[0]["secret_internal"]; ok {
		t.Fatalf("records[0] = %+v, undeclared field should be dropped in schema projection", recs[0])
	}
	if recs[0]["id"] != "1" || recs[0]["name"] != "a" {
		t.Fatalf("records[0] = %+v, declared fields should survive", recs[0])
	}
}

func TestReadProjectionPassthroughKeepsUndeclaredFields(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z","extra":"x"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}, Projection: "passthrough"})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["extra"] != "x" {
		t.Fatalf("records = %+v, want passthrough field preserved", recs)
	}
}

// --- computed_fields ---

func TestReadComputedFieldsNestedExtraction(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z","user":{"login":"octocat"}}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:        RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{"user_login": "{{ record.user.login }}"},
	})
	// computed_fields must survive schema projection even though "user_login"
	// is not a declared property, so extend the schema.
	raw := json.RawMessage(`{
		"type": "object",
		"x-primary-key": ["id"],
		"x-cursor-field": "updated_at",
		"properties": {
			"id": {"type": "string"}, "name": {"type": "string"},
			"updated_at": {"type": "string"}, "user_login": {"type": "string"}
		}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Schemas["widgets"] = &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["user_login"] != "octocat" {
		t.Fatalf("records = %+v, want user_login=octocat", recs)
	}
}

func TestReadComputedFieldsMissingIntermediateDoesNotPanic(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:        RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{"user_login": "{{ record.user.login }}"},
	})
	raw := json.RawMessage(`{
		"type": "object", "x-primary-key": ["id"], "x-cursor-field": "updated_at",
		"properties": {"id": {"type":"string"}, "name":{"type":"string"}, "updated_at":{"type":"string"}, "user_login":{"type":["string","null"]}}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Schemas["widgets"] = &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v (missing intermediate must not panic/error)", err)
	}
	if len(recs) != 1 {
		t.Fatalf("records = %+v", recs)
	}
	if v, ok := recs[0]["user_login"]; ok && v != nil {
		t.Fatalf("user_login = %v, want nil/absent for missing intermediate", v)
	}
}

// TestReadComputedFieldsStaticLiteralNoTemplate proves a computed_fields
// entry whose value has NO {{ }} template markers at all (F7 meta-rule
// enablement: a static literal, e.g. "source_system": "searxng") is emitted
// as that exact literal string, not treated as an interpolation error or
// dropped. This already works via Interpolate's existing "no {{ }} match
// means no replacement" semantics (interpolate()'s ReplaceAllStringFunc is a
// no-op when the template has nothing to replace) — this test locks it in as
// a named regression guard rather than leaving the behavior implicit.
func TestReadComputedFieldsStaticLiteralNoTemplate(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:        RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{"source_system": "searxng"},
	})
	raw := json.RawMessage(`{
		"type": "object", "x-primary-key": ["id"], "x-cursor-field": "updated_at",
		"properties": {"id":{"type":"string"},"name":{"type":"string"},"updated_at":{"type":"string"},"source_system":{"type":"string"}}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Schemas["widgets"] = &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["source_system"] != "searxng" {
		t.Fatalf("records = %+v, want source_system=searxng (static literal, no {{ }} markers)", recs)
	}
}

// TestReadComputedFieldsJoinFilterArrayField proves the join:<sep> filter
// (F7 meta-rule enablement) is reachable through the full Read path: an
// array-valued raw record field joined into a separator-delimited string
// computed field, without changing the record's OTHER emitted fields.
func TestReadComputedFieldsJoinFilterArrayField(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z","engines":["google","bing","ddg"]}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:        RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{"engines_csv": "{{ record.engines | join:, }}"},
	})
	raw := json.RawMessage(`{
		"type": "object", "x-primary-key": ["id"], "x-cursor-field": "updated_at",
		"properties": {"id":{"type":"string"},"name":{"type":"string"},"updated_at":{"type":"string"},"engines_csv":{"type":"string"}}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Schemas["widgets"] = &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["engines_csv"] != "google,bing,ddg" {
		t.Fatalf("records = %+v, want engines_csv=google,bing,ddg", recs)
	}
}

// TestReadComputedFieldsBareRecordPathPreservesNativeType is the RED test for
// the wave1-pilot gap-loop A1 adjudication (chargebee ~30 fields / gmail 4
// fields / github 4 fields all forced to widen a real integer/boolean schema
// type to string because computed_fields' Interpolate always stringifies).
// A computed_fields template that is a SINGLE bare `{{ record.<path> }}`
// reference with NO filter stage must copy the RAW typed JSON value
// (number/bool/null preserved) into the projected record, bypassing
// Interpolate's stringify entirely. A template with a filter, or with more
// than one `{{ }}` segment / surrounding text, keeps today's string
// semantics unchanged (see the next test).
func TestReadComputedFieldsBareRecordPathPreservesNativeType(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z","meta":{"count":42,"active":true,"missing":null}}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{
			"count_typed":  "{{ record.meta.count }}",
			"active_typed": "{{ record.meta.active }}",
		},
	})
	raw := json.RawMessage(`{
		"type": "object", "x-primary-key": ["id"], "x-cursor-field": "updated_at",
		"properties": {
			"id": {"type": "string"}, "name": {"type": "string"}, "updated_at": {"type": "string"},
			"count_typed": {"type": "integer"}, "active_typed": {"type": "boolean"}
		}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Schemas["widgets"] = &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("records = %+v", recs)
	}
	if _, isString := recs[0]["count_typed"].(string); isString {
		t.Fatalf("count_typed = %#v (%T), want a native number (json.Number/float64/int), not a string", recs[0]["count_typed"], recs[0]["count_typed"])
	}
	if _, isString := recs[0]["active_typed"].(string); isString {
		t.Fatalf("active_typed = %#v (%T), want a native bool, not a string", recs[0]["active_typed"], recs[0]["active_typed"])
	}
	if b, ok := recs[0]["active_typed"].(bool); !ok || b != true {
		t.Fatalf("active_typed = %#v, want bool true", recs[0]["active_typed"])
	}
}

// TestReadComputedFieldsFilteredOrMixedTemplateKeepsStringSemantics locks in
// the OTHER half of A1's ruling: a template with a filter stage (join:,
// unix_seconds, etc.) or with more than a single bare {{ record.path }}
// reference (mixed/multi-part text) keeps producing a STRING, exactly as
// before — only the single-bare-reference shape gets typed extraction.
func TestReadComputedFieldsFilteredOrMixedTemplateKeepsStringSemantics(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z","count":42,"engines":["google","bing"]}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{
			"engines_csv": "{{ record.engines | join:, }}",
			"count_label": "count={{ record.count }}",
		},
	})
	raw := json.RawMessage(`{
		"type": "object", "x-primary-key": ["id"], "x-cursor-field": "updated_at",
		"properties": {
			"id": {"type": "string"}, "name": {"type": "string"}, "updated_at": {"type": "string"},
			"engines_csv": {"type": "string"}, "count_label": {"type": "string"}
		}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Schemas["widgets"] = &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["engines_csv"] != "google,bing" {
		t.Fatalf("records = %+v, want engines_csv=google,bing (string, filtered template)", recs)
	}
	if recs[0]["count_label"] != "count=42" {
		t.Fatalf("records = %+v, want count_label=count=42 (string, mixed template)", recs)
	}
}

// TestReadComputedFieldsConfigReference proves A3/G0: applyComputedFields'
// Vars must wire Config (not Secrets) so a computed field can stamp a
// config-derived marker onto every record (github's dropped `repository`
// field, e.g. "{{ config.owner }}/{{ config.repo }}").
func TestReadComputedFieldsConfigReference(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:        RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{"repository": "{{ config.owner }}/{{ config.repo }}"},
	})
	raw := json.RawMessage(`{
		"type": "object", "x-primary-key": ["id"], "x-cursor-field": "updated_at",
		"properties": {
			"id": {"type": "string"}, "name": {"type": "string"}, "updated_at": {"type": "string"},
			"repository": {"type": "string"}
		}
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Schemas["widgets"] = &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}

	cfg := connectors.RuntimeConfig{Config: map[string]string{"owner": "octo-org", "repo": "hello-world"}}
	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets", Config: cfg}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["repository"] != "octo-org/hello-world" {
		t.Fatalf("records = %+v, want repository=octo-org/hello-world", recs)
	}
}

// TestReadComputedFieldsSecretsNotAccessible proves the A3 threat-model line:
// Secrets must stay EXCLUDED from applyComputedFields' Vars — a computed
// field must never be able to copy a secret value into an emitted record.
// A template referencing secrets.* must still hard-error (unresolved key),
// exactly like today, never silently resolve a secret into record data.
func TestReadComputedFieldsSecretsNotAccessible(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:        RecordsSpec{Path: "data"},
		ComputedFields: map[string]string{"leaked": "{{ secrets.token }}"},
	})
	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"token": "sk_should_never_appear"}}
	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets", Config: cfg}, nil)
	if err == nil {
		t.Fatalf("Read: want error (secrets.* must never resolve inside computed_fields), got nil")
	}
}

// --- cursor advance + resume ---

func TestReadCursorAdvancesToMaxSeenAndResumeSendsRequestParam(t *testing.T) {
	// The engine tracks MaxCursor internally per design §B.4 ("track MaxCursor
	// ... emit"); the app layer (internal/app/local_warehouse.go) derives the
	// persisted state cursor from each emitted record's declared cursor field
	// exactly the same way, via connsdk.MaxCursor, since connectors.Connector.Read
	// has no side channel for a "final cursor" return. This test proves both
	// halves of that contract: (1) the emitted records carry the raw cursor
	// field values needed for MaxCursor derivation, (2) a resumed Read sends the
	// advanced cursor as the request_param.
	var seenSinceValues []string
	page := 0
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenSinceValues = append(seenSinceValues, r.URL.Query().Get("since"))
		page++
		switch page {
		case 1:
			_, _ = w.Write([]byte(`{"data":[
				{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"},
				{"id":"2","name":"b","updated_at":"2026-01-03T00:00:00Z"}
			], "has_more": false}`))
		}
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:     RecordsSpec{Path: "data"},
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "since", ParamFormat: "rfc3339"},
	})

	state, err := InitialState(context.Background(), b, "widgets", connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("InitialState: %v", err)
	}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets", State: state}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	finalCursor := ""
	for _, rec := range recs {
		if v, ok := rec["updated_at"].(string); ok {
			finalCursor = connsdk.MaxCursor(finalCursor, v)
		}
	}
	if finalCursor != "2026-01-03T00:00:00Z" {
		t.Fatalf("finalCursor = %q, want max cursor seen (2026-01-03T00:00:00Z)", finalCursor)
	}

	// Resume: re-read with the advanced cursor should send it as the request param.
	page = 0
	_, err = readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets", State: map[string]string{"cursor": finalCursor}}, nil)
	if err != nil {
		t.Fatalf("Read (resume): %v", err)
	}
	if seenSinceValues[len(seenSinceValues)-1] != finalCursor {
		t.Fatalf("resume since = %q, want %q", seenSinceValues[len(seenSinceValues)-1], finalCursor)
	}
}

// --- client_filtered ---

func TestReadClientFilteredDropsOldRecords(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[
			{"id":"1","name":"old","updated_at":"2025-01-01T00:00:00Z"},
			{"id":"2","name":"new","updated_at":"2026-06-01T00:00:00Z"}
		]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:     RecordsSpec{Path: "data"},
		Incremental: &IncrementalSpec{CursorField: "updated_at", ClientFiltered: true},
	})

	req := connectors.ReadRequest{Stream: "widgets", State: map[string]string{"cursor": "2026-01-01T00:00:00Z"}}
	recs, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["id"] != "2" {
		t.Fatalf("records = %+v, want only the record newer than cursor", recs)
	}
}

// --- conditional header omission ---

func TestReadHeaderOmittedWhenInterpolatedValueEmpty(t *testing.T) {
	var gotHeader string
	sawHeader := false
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Stripe-Account")
		_, sawHeader = r.Header["Stripe-Account"]
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.Headers = map[string]string{"Stripe-Account": "{{ config.account_id }}"}

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawHeader {
		t.Fatalf("Stripe-Account header present (%q) with empty interpolated value, want omitted", gotHeader)
	}

	// Non-empty config value: header must be sent.
	req.Config.Config["account_id"] = "acct_123"
	_, err = readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotHeader != "acct_123" {
		t.Fatalf("Stripe-Account header = %q, want acct_123", gotHeader)
	}
}

// --- error_map ---

func TestReadErrorMap401HintSurfaces(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`unauthorized`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.ErrorMap = []ErrorRule{{Status: 401, Hint: "token is missing or expired; re-run pm credentials add acme"}}

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want error for 401")
	}
	if !containsStr(err.Error(), "token is missing or expired; re-run pm credentials add acme") {
		t.Fatalf("error = %q, want hint text", err.Error())
	}
}

func TestReadErrorMap403MatchBodyRateLimited(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`you have hit the rate limit`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.ErrorMap = []ErrorRule{{Status: 403, MatchBody: "rate limit", Class: "rate_limited"}}

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want error for 403")
	}
	var engineErr *Error
	if !errAs(err, &engineErr) {
		t.Fatalf("error = %v, want *engine.Error", err)
	}
	if engineErr.Class != "rate_limited" {
		t.Fatalf("Class = %q, want rate_limited", engineErr.Class)
	}
}

// --- rate limiting ---

func TestReadRateLimitSleeperInvokedNMinus1Times(t *testing.T) {
	page := 0
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		switch {
		case page < 3:
			_, _ = fmt.Fprintf(w, `{"data":[{"id":"%d","name":"a","updated_at":"2026-01-01T00:00:00Z"}], "has_more": true}`, page)
		default:
			_, _ = w.Write([]byte(`{"data":[], "has_more": false}`))
		}
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{
			Type: "cursor", CursorParam: "starting_after", LastRecordField: "id", StopPath: "has_more",
		},
	})
	b.HTTP.RateLimit = &RateLimitSpec{RequestsPerMinute: 600}

	var sleeps int
	var mu sync.Mutex
	sleeper := func(ctx context.Context, d time.Duration) error {
		mu.Lock()
		sleeps++
		mu.Unlock()
		return nil
	}

	recs, err := readAllWithSleeper(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil, sleeper)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("records = %+v, want 2 across 3 requests", recs)
	}
	// 3 requests total -> sleeper invoked between them: N-1 = 2 times.
	if sleeps != 2 {
		t.Fatalf("sleeps = %d, want 2 (N-1 for 3 requests)", sleeps)
	}
}

func readAllWithSleeper(t *testing.T, ctx context.Context, b Bundle, req connectors.ReadRequest, h Hooks, sleeper func(context.Context, time.Duration) error) ([]connectors.Record, error) {
	t.Helper()
	var out []connectors.Record
	err := ReadWithSleeper(ctx, b, req, h, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	}, sleeper)
	return out, err
}

// --- hooks ---

func TestReadRecordHookMutateAndDrop(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[
			{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"},
			{"id":"2","name":"b","updated_at":"2026-01-01T00:00:00Z"}
		]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})

	callCount := 0
	h := &recordHookFunc{
		fn: func(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error) {
			callCount++
			if projected["id"] == "2" {
				return nil, false, nil // drop
			}
			projected["mutated"] = true
			return projected, true, nil
		},
	}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, h)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("RecordHook called %d times, want 2", callCount)
	}
	if len(recs) != 1 || recs[0]["id"] != "1" {
		t.Fatalf("records = %+v, want only id=1 (id=2 dropped)", recs)
	}
	if recs[0]["mutated"] != true {
		t.Fatalf("records[0] = %+v, want mutated=true", recs[0])
	}
}

func TestReadStreamHookHandledBypassesDeclarative(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("declarative HTTP request should not happen when StreamHook handles the read")
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})

	h := &streamHookFunc{
		fn: func(ctx context.Context, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, emit func(connectors.Record) error) (bool, error) {
			_ = emit(connectors.Record{"id": "hook-emitted"})
			return true, nil
		},
	}

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, h)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 || recs[0]["id"] != "hook-emitted" {
		t.Fatalf("records = %+v, want hook-emitted record only", recs)
	}
}

func TestReadCheckHookHandled(t *testing.T) {
	called := false
	h := &checkHookFunc{
		fn: func(ctx context.Context, cfg connectors.RuntimeConfig, rt *Runtime) (bool, error) {
			called = true
			return true, nil
		},
	}
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: "http://example.invalid", Check: &RequestSpec{Method: "GET", Path: "/check"}}}
	err := Check(context.Background(), b, connectors.RuntimeConfig{}, h)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !called {
		t.Fatalf("CheckHook was not invoked")
	}
}

// --- limit / ctx cancellation ---

func TestReadCtxCancelMidPage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		cancel()
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}], "has_more": true}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{
			Type: "cursor", CursorParam: "starting_after", LastRecordField: "id", StopPath: "has_more",
		},
	})

	_, err := readAll(t, ctx, b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want context.Canceled error")
	}
}

func TestReadLimitEmitterStopsEarly(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[
			{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"},
			{"id":"2","name":"b","updated_at":"2026-01-01T00:00:00Z"}
		]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})

	var out []connectors.Record
	emit := connectors.LimitEmitter(1, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	})
	err := Read(context.Background(), b, connectors.ReadRequest{Stream: "widgets", Limit: 1}, nil, emit)
	if err := connectors.IgnoreReadLimit(err); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("out = %+v, want exactly 1 record (limit reached)", out)
	}
}

// --- stream not found ---

func TestReadUnknownStreamErrors(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "does-not-exist"}, nil)
	if err == nil {
		t.Fatalf("Read: want error for unknown stream")
	}
}

func TestReadAuthSelectionFailureSurfaces(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should be sent when auth selection fails")
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.Auth = []AuthSpec{{Mode: "bearer", Token: "{{ secrets.missing_token }}"}}

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want error when the auth spec's token cannot be resolved")
	}
}

func TestReadBaseURLInterpolationFailureSurfaces(t *testing.T) {
	b := Bundle{
		Name:    "acme",
		HTTP:    HTTPBase{URL: "{{ config.missing_base_url }}"},
		Streams: []StreamSpec{{Name: "widgets", Path: "/widgets", Records: RecordsSpec{Path: "data"}}},
	}
	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want error when the base url template cannot be resolved")
	}
}

// --- generic InitialState ---

func TestInitialStateEmptyCursor(t *testing.T) {
	b := Bundle{Name: "acme"}
	state, err := InitialState(context.Background(), b, "widgets", connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("InitialState: %v", err)
	}
	if connsdk.Cursor(state) != "" {
		t.Fatalf("InitialState cursor = %q, want empty", connsdk.Cursor(state))
	}
}

// --- test-only hook adapter helpers ---

// recordHookFunc adapts a function literal to the RecordHook interface (and
// the base Hooks interface) for table-driven-style hook tests.
type recordHookFunc struct {
	fn func(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error)
}

func (r *recordHookFunc) ConnectorName() string { return "record-hook-func-test" }
func (r *recordHookFunc) MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error) {
	return r.fn(stream, raw, projected)
}

type streamHookFunc struct {
	fn func(ctx context.Context, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, emit func(connectors.Record) error) (bool, error)
}

func (s *streamHookFunc) ConnectorName() string { return "stream-hook-func-test" }
func (s *streamHookFunc) ReadStream(ctx context.Context, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, emit func(connectors.Record) error) (bool, error) {
	return s.fn(ctx, stream, req, rt, emit)
}

type checkHookFunc struct {
	fn func(ctx context.Context, cfg connectors.RuntimeConfig, rt *Runtime) (bool, error)
}

func (c *checkHookFunc) ConnectorName() string { return "check-hook-func-test" }
func (c *checkHookFunc) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *Runtime) (bool, error) {
	return c.fn(ctx, cfg, rt)
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}

func errAs(err error, target any) bool {
	return errors.As(err, target)
}

// --- declarative Check (no CheckHook) ---

func TestCheckDeclarativeRequestSucceeds(t *testing.T) {
	var gotPath string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{Method: "GET", Path: "/status"}}}

	if err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gotPath != "/status" {
		t.Fatalf("path = %q, want /status", gotPath)
	}
}

func TestCheckNoDeclaredCheckIsNoop(t *testing.T) {
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: "http://example.invalid"}}
	if err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil); err != nil {
		t.Fatalf("Check: %v, want nil (no HTTP.Check declared)", err)
	}
}

func TestCheckDeclarativeRequestErrorMapApplied(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`unauthorized`))
	})
	b := Bundle{
		Name: "acme",
		HTTP: HTTPBase{
			URL:      srv.URL,
			Check:    &RequestSpec{Method: "GET", Path: "/status"},
			ErrorMap: []ErrorRule{{Status: 401, Hint: "re-run pm credentials add acme"}},
		},
	}

	err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil)
	if err == nil {
		t.Fatalf("Check: want error for 401")
	}
	if !containsStr(err.Error(), "re-run pm credentials add acme") {
		t.Fatalf("error = %q, want hint text", err.Error())
	}
}

func TestCheckHookErrorPropagates(t *testing.T) {
	h := &checkHookFunc{fn: func(ctx context.Context, cfg connectors.RuntimeConfig, rt *Runtime) (bool, error) {
		return false, fmt.Errorf("boom")
	}}
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: "http://example.invalid"}}
	err := Check(context.Background(), b, connectors.RuntimeConfig{}, h)
	if err == nil {
		t.Fatalf("Check: want error from CheckHook")
	}
}

// --- PaginationSpec.MaxPages wiring (ENGINE_GAP repair, wave0-engine-harness) ---

// TestReadMaxPagesHardStopsRequestCount proves readDeclarative enforces a
// hard request-count cap independent of page fullness: a source that ALWAYS
// returns a full page (so the short-page stop signal never fires on its own)
// must still stop after exactly MaxPages requests.
func TestReadMaxPagesHardStopsRequestCount(t *testing.T) {
	var hits int
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		// Always full (PageSize=2, always return exactly 2 records): the
		// short-page stop signal never fires by itself.
		_, _ = fmt.Fprintf(w, `{"data":[{"id":"%d-a","name":"a","updated_at":"2026-01-01T00:00:00Z"},{"id":"%d-b","name":"b","updated_at":"2026-01-01T00:00:00Z"}]}`, hits, hits)
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{
			Type: "page_number", PageParam: "pageno", StartPage: intPtr(1), PageSize: 2, MaxPages: 2,
		},
	})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if hits != 2 {
		t.Fatalf("requests issued = %d, want exactly 2 (MaxPages=2 hard stop against an always-full-page source)", hits)
	}
	if len(recs) != 4 {
		t.Fatalf("records = %d, want 4 (2 pages x 2 records)", len(recs))
	}
}

// TestReadMaxPagesZeroIsUnbounded proves MaxPages==0 (the zero value/absent
// case) preserves today's behavior: pagination is bounded only by the
// short/empty-page stop signal, never by a request-count cap.
func TestReadMaxPagesZeroIsUnbounded(t *testing.T) {
	var hits int
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		switch {
		case hits < 3:
			_, _ = fmt.Fprintf(w, `{"data":[{"id":"%d-a","name":"a","updated_at":"2026-01-01T00:00:00Z"},{"id":"%d-b","name":"b","updated_at":"2026-01-01T00:00:00Z"}]}`, hits, hits)
		default:
			// Short (empty) page: the ordinary stop signal fires here.
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{
			Type: "page_number", PageParam: "pageno", StartPage: intPtr(1), PageSize: 2, MaxPages: 0,
		},
	})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if hits != 3 {
		t.Fatalf("requests issued = %d, want 3 (MaxPages=0 means unbounded; only the short-page stop applies)", hits)
	}
	if len(recs) != 4 {
		t.Fatalf("records = %d, want 4 (2 full pages x 2 records, 3rd page empty)", len(recs))
	}
}

// TestReadMaxPagesAbsentPaginationSpecIsUnbounded proves a stream with NO
// pagination spec at all (nil) is unaffected by the MaxPages wiring — same
// as MaxPages==0, existing short-page-stop-only behavior must stay green.
func TestReadMaxPagesAbsentPaginationSpecIsUnbounded(t *testing.T) {
	var hits int
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if hits != 1 {
		t.Fatalf("requests issued = %d, want 1 (nonePaginator issues exactly one request)", hits)
	}
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
}

// TestReadMaxPagesStreamLevelOverridesBase proves the stream-level
// PaginationSpec's MaxPages overrides the base-level spec's MaxPages,
// matching read.go's existing "stream overrides base" merge semantics (the
// stream spec is used wholesale when non-nil; there is no field-by-field
// merge between base and stream pagination specs).
func TestReadMaxPagesStreamLevelOverridesBase(t *testing.T) {
	var hits int
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = fmt.Fprintf(w, `{"data":[{"id":"%d-a","name":"a","updated_at":"2026-01-01T00:00:00Z"},{"id":"%d-b","name":"b","updated_at":"2026-01-01T00:00:00Z"}]}`, hits, hits)
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{
			Type: "page_number", PageParam: "pageno", StartPage: intPtr(1), PageSize: 2, MaxPages: 3,
		},
	})
	// Base declares MaxPages=1; if the stream-level override did not win, the
	// engine would stop after only 1 request instead of the stream's 3.
	b.HTTP.Pagination = &PaginationSpec{
		Type: "page_number", PageParam: "pageno", StartPage: intPtr(1), PageSize: 2, MaxPages: 1,
	}

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if hits != 3 {
		t.Fatalf("requests issued = %d, want 3 (stream-level MaxPages=3 must override base-level MaxPages=1)", hits)
	}
}

// --- next_url pagination through a real Read call (exercises requesterHost) ---

func TestReadNextURLPaginationSetsBaseHostFromRequester(t *testing.T) {
	var srv *httptest.Server
	page := 0
	srv = jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		switch page {
		case 1:
			_, _ = fmt.Fprintf(w, `{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}],"meta":{"next_page_link":%q}}`, srv.URL+"/widgets?page=2")
		default:
			_, _ = w.Write([]byte(`{"data":[{"id":"2","name":"b","updated_at":"2026-01-01T00:00:00Z"}],"meta":{"next_page_link":""}}`))
		}
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{
			Type: "next_url", NextURLPath: "meta.next_page_link",
		},
	})

	recs, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("records = %+v, want 2 across both pages", recs)
	}
}

// TestReadNextURLPaginationCrossHostBlocked proves the SSRF guard is wired
// live through Read (BaseHost derived from requester.BaseURL via
// requesterHost), not just unit-tested against newPaginator directly.
func TestReadNextURLPaginationCrossHostBlocked(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}],"meta":{"next_page_link":"https://evil.example.com/steal"}}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:    RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{Type: "next_url", NextURLPath: "meta.next_page_link"},
	})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want cross-host next_url blocked")
	}
	if !containsStr(err.Error(), "cross-host") {
		t.Fatalf("error = %q, want cross-host guard message", err.Error())
	}
}

// --- F1 (REVIEW.md high flag / SECURITY-REVIEW.md m3): stream.Path and
// HTTP.Check.Path must be run through InterpolatePath, exactly like write.go
// already does for action.Path. Before the fix, `{{ }}` markers in a stream
// path are sent to the live API LITERALLY. ---

func TestReadStreamPathIsInterpolated(t *testing.T) {
	var gotPath string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Path: "/repos/{{ config.repo }}", Records: RecordsSpec{Path: "data"}})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"repo": "acme/widgets"}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// InterpolatePath urlencodes by default (path segments are the primary
	// injection surface, THREAT-MODEL §2): "/" in the config value is encoded
	// on the wire (net/http decodes r.URL.Path for handler convenience, so the
	// escaped form is what proves the literal-vs-interpolated distinction).
	if gotPath != "/repos/acme%2Fwidgets" {
		t.Fatalf("escaped path = %q, want /repos/acme%%2Fwidgets (templated path must be interpolated+urlencoded, not sent literally)", gotPath)
	}
}

func TestReadStreamPathUnresolvedKeyErrors(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should be sent when the stream path template cannot be resolved")
	})
	b := newTestBundle(t, srv, StreamSpec{Path: "/repos/{{ config.missing_repo }}", Records: RecordsSpec{Path: "data"}})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want error when the stream path template cannot be resolved")
	}
}

// TestReadStreamPathStaticGoldenUnaffected locks in that a static
// (non-templated) path — what every wave0 golden bundle uses — round-trips
// unchanged through InterpolatePath (no `{{ }}` markers means no
// replacement, hence no accidental encoding of literal path characters).
func TestReadStreamPathStaticGoldenUnaffected(t *testing.T) {
	var gotPath string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Path: "/widgets/all", Records: RecordsSpec{Path: "data"}})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotPath != "/widgets/all" {
		t.Fatalf("path = %q, want /widgets/all (static path unaffected)", gotPath)
	}
}

func TestCheckPathIsInterpolated(t *testing.T) {
	var gotPath string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{Method: "GET", Path: "/accounts/{{ config.account_id }}/status"}}}

	cfg := connectors.RuntimeConfig{Config: map[string]string{"account_id": "acct_123"}}
	if err := Check(context.Background(), b, cfg, nil); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gotPath != "/accounts/acct_123/status" {
		t.Fatalf("path = %q, want /accounts/acct_123/status (check path must be interpolated)", gotPath)
	}
}

func TestCheckPathUnresolvedKeyErrors(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should be sent when the check path template cannot be resolved")
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{Method: "GET", Path: "/accounts/{{ config.missing_account_id }}/status"}}}

	err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil)
	if err == nil {
		t.Fatalf("Check: want error when the check path template cannot be resolved")
	}
}

// --- ENGINE DIALECT ADDITION (checkquery-ledger.md): RequestSpec.Query ----
//
// 148 bundles under internal/connectors/defs declare "base.check.query"
// (e.g. limit=1 probe params legacy checks genuinely send) that RequestSpec
// (Method/Path only) could not express at all — the hardening ledger's own
// follow-up note names exactly this gap. RequestSpec gains a Query
// map[string]QueryParam field (mirroring StreamSpec.Query's existing
// string-or-object dialect verbatim, per hardening-ledger.md's suggested
// follow-up shape) and Check() resolves+sends it using the SAME semantics as
// stream.Query: Interpolate against requestVars, hard error on an unresolved
// config/secrets key for a plain-string entry, OmitWhenAbsent/Default honored
// for an object-form entry.

// TestCheckQuerySentOnRequest proves a static (non-templated) check.query
// entry is resolved and actually sent on the wire — the exact defect class
// the ledger names: RequestSpec previously had nowhere to put this value at
// all, so it was either a load-time meta-schema rejection or (pre-hardening)
// a silently dropped no-op.
func TestCheckQuerySentOnRequest(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{
		Method: "GET",
		Path:   "/status",
		Query:  map[string]QueryParam{"limit": {Template: "1"}},
	}}}

	if err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gotQuery.Get("limit") != "1" {
		t.Fatalf("query = %v, want limit=1", gotQuery)
	}
}

// TestCheckQueryTemplateResolvedAgainstConfig proves a templated check.query
// value resolves against the same config Check() already threads through
// requestVars for the check path.
func TestCheckQueryTemplateResolvedAgainstConfig(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{
		Method: "GET",
		Path:   "/status",
		Query:  map[string]QueryParam{"page_size": {Template: "{{ config.page_size }}"}},
	}}}

	cfg := connectors.RuntimeConfig{Config: map[string]string{"page_size": "5"}}
	if err := Check(context.Background(), b, cfg, nil); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gotQuery.Get("page_size") != "5" {
		t.Fatalf("query = %v, want page_size=5", gotQuery)
	}
}

// TestCheckQueryUnresolvedKeyErrorsHard proves a plain-string check.query
// entry referencing an absent config/secrets key is a hard error — the SAME
// semantics as stream.Query's plain-string entries (buildInitialQuery), not
// a silent omission (F4's fail-open class the engine deliberately rejects
// elsewhere).
func TestCheckQueryUnresolvedKeyErrorsHard(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should be sent when a check query template cannot be resolved")
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{
		Method: "GET",
		Path:   "/status",
		Query:  map[string]QueryParam{"page_size": {Template: "{{ config.missing_page_size }}"}},
	}}}

	err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil)
	if err == nil {
		t.Fatalf("Check: want error when a check query template cannot be resolved")
	}
}

// TestCheckQueryOmitWhenAbsentDropsParam proves the object-form
// omit_when_absent dialect works identically to stream.Query's: an
// unresolved key is tolerated and the param is left off the request
// entirely, rather than erroring.
func TestCheckQueryOmitWhenAbsentDropsParam(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{
		Method: "GET",
		Path:   "/status",
		Query:  map[string]QueryParam{"status": {Template: "{{ config.status }}", OmitWhenAbsent: true}},
	}}}

	if err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gotQuery.Has("status") {
		t.Fatalf("query = %v, want status omitted (unresolved + omit_when_absent)", gotQuery)
	}
}

// TestCheckQueryDefaultSentWhenAbsent proves the object-form default dialect
// works identically to stream.Query's: an unresolved key falls back to the
// literal default value instead of erroring.
func TestCheckQueryDefaultSentWhenAbsent(t *testing.T) {
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{
		Method: "GET",
		Path:   "/status",
		Query:  map[string]QueryParam{"count": {Template: "{{ config.missing_count }}", Default: "100"}},
	}}}

	if err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gotQuery.Get("count") != "100" {
		t.Fatalf("query = %v, want count=100 (default)", gotQuery)
	}
}

// TestCheckNoQueryDeclaredSendsNoQuery locks in that a bundle with no
// check.query at all (every pre-existing golden) is completely unaffected —
// Check() sends no query string, same as before this dialect existed.
func TestCheckNoQueryDeclaredSendsNoQuery(t *testing.T) {
	var gotRawQuery string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	b := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL, Check: &RequestSpec{Method: "GET", Path: "/status"}}}

	if err := Check(context.Background(), b, connectors.RuntimeConfig{}, nil); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gotRawQuery != "" {
		t.Fatalf("raw query = %q, want empty (no check.query declared)", gotRawQuery)
	}
}

// --- F4 (REVIEW.md flag / SECURITY-REVIEW.md finding): resolveHeaders
// swallowed ANY unresolved-key error uniformly (isUnresolvedKey substring
// match), so a header referencing an absent secrets.* key was SILENTLY
// OMITTED — sending the request unauthenticated instead of failing. ---

// specSchemaWithRequired compiles a minimal spec.json-shaped Schema (as
// Bundle.Spec) declaring optionalKey as a plain declared property and
// requiredKey (when non-empty) as also present in "required".
func specSchemaWithRequired(t *testing.T, optionalKey, requiredKey string) *Schema {
	t.Helper()
	props := `"` + optionalKey + `": {"type": "string"}`
	req := ""
	if requiredKey != "" {
		props += `, "` + requiredKey + `": {"type": "string"}`
		req = `, "required": ["` + requiredKey + `"]`
	}
	raw := json.RawMessage(`{"type": "object", "properties": {` + props + `}` + req + `}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	return sch
}

func TestReadHeaderAbsentSecretAuthorizationErrors(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should be sent unauthenticated when a secrets.* header reference cannot be resolved")
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.Headers = map[string]string{"Authorization": "Bearer {{ secrets.token }}"}

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{}, Secrets: map[string]string{}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err == nil {
		t.Fatalf("Read: want error when an Authorization header's secrets.* reference cannot be resolved (never silently send unauthenticated)")
	}
}

func TestReadHeaderOptionalDeclaredCustomHeaderOmitted(t *testing.T) {
	// Stripe-Account pattern: a declared-but-not-required config key (account_id)
	// with no runtime value must still OMIT the header, not error.
	var sawHeader bool
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, sawHeader = r.Header["Stripe-Account"]
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.Headers = map[string]string{"Stripe-Account": "{{ config.account_id }}"}
	b.Spec = specSchemaWithRequired(t, "account_id", "")

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawHeader {
		t.Fatalf("Stripe-Account header present, want omitted (account_id is declared-optional, absent at runtime)")
	}
}

func TestReadHeaderRequiredConfigKeyErrors(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should be sent when a REQUIRED config header key is absent")
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.Headers = map[string]string{"X-Tenant": "{{ config.tenant_id }}"}
	b.Spec = specSchemaWithRequired(t, "unrelated", "tenant_id")

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err == nil {
		t.Fatalf("Read: want error when a REQUIRED config key referenced by a header is absent")
	}
}

func TestReadHeaderUndeclaredConfigKeyErrors(t *testing.T) {
	// A config.* reference to a key not declared in spec.json's properties AT
	// ALL (not merely optional) must also hard-error, not be silently omitted
	// — undeclared is a stronger signal of a bug than declared-optional.
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should be sent when a header references an undeclared config key")
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}})
	b.HTTP.Headers = map[string]string{"X-Custom": "{{ config.totally_undeclared }}"}
	b.Spec = specSchemaWithRequired(t, "account_id", "")

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err == nil {
		t.Fatalf("Read: want error when a header references a config key not declared in spec.json at all")
	}
}

// TestReadHeaderOptionalOmittedNoSpecOnBundle locks in backward compatibility
// for the many existing test bundles (newTestBundle et al.) that never set
// b.Spec at all: with no spec to consult, a header referencing an absent
// config.* key still OMITS (cannot distinguish declared-optional from
// undeclared without a spec; the safer-by-default reading for a bundle with
// literally no declared config surface is to preserve the prior
// omit-on-absent-config behavior, exactly as TestReadHeaderOmittedWhenInterpolatedValueEmpty
// already asserts and continues to assert unmodified below).
func TestReadHeaderOptionalOmittedNoSpecOnBundle(t *testing.T) {
	var sawHeader bool
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, sawHeader = r.Header["X-Custom"]
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{Records: RecordsSpec{Path: "data"}}) // b.Spec is nil
	b.HTTP.Headers = map[string]string{"X-Custom": "{{ config.account_id }}"}

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawHeader {
		t.Fatalf("X-Custom header present, want omitted (no spec on bundle: preserve prior omit-on-absent-config behavior)")
	}
}

// --- S4 engine mini-wave item 2: sub-resource fan-out -----------------------

// TestReadFanOutConfigKeyRunsOncePerID proves the config_key + query_param
// shape (appfollow's app_collection_ids -> apps_id): the CSV config value is
// split into 2 ids, the declarative read sequence runs once per id (both
// hitting the SAME stream path), the id is sent as a query param on every
// request of its sub-sequence, and stamp_field writes the id onto every
// emitted record.
func TestReadFanOutConfigKeyRunsOncePerID(t *testing.T) {
	var gotIDs []string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("apps_id")
		gotIDs = append(gotIDs, id)
		_, _ = w.Write([]byte(`{"data":[{"id":"rec-` + id + `","name":"n"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		FanOut: &FanOutSpec{
			IDsFrom:    FanOutIDsFrom{ConfigKey: "app_collection_ids"},
			Into:       FanOutInto{QueryParam: "apps_id"},
			StampField: "app_id",
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"app_collection_ids": "111,222"}}}
	records, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(gotIDs) != 2 || gotIDs[0] != "111" || gotIDs[1] != "222" {
		t.Fatalf("gotIDs = %v, want [111 222] (once per configured id, in order)", gotIDs)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2 (one per id sub-sequence)", len(records))
	}
	if records[0]["app_id"] != "111" || records[1]["app_id"] != "222" {
		t.Fatalf("records = %+v, want app_id stamped from the fan-out id", records)
	}
}

// TestReadFanOutConfigKeyPathVarInterpolatesFanoutID proves the path_var form:
// {{ fanout.id }} in stream.Path resolves to the current fan-out id.
func TestReadFanOutConfigKeyPathVarInterpolatesFanoutID(t *testing.T) {
	var gotPaths []string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"n"}]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Path:    "/projects/{{ fanout.id }}/tasks",
		Records: RecordsSpec{Path: "data"},
		FanOut: &FanOutSpec{
			IDsFrom:    FanOutIDsFrom{ConfigKey: "project_ids"},
			Into:       FanOutInto{PathVar: "parent_id"},
			StampField: "project_id",
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"project_ids": "a,b"}}}
	records, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := []string{"/projects/a/tasks", "/projects/b/tasks"}
	if len(gotPaths) != 2 || gotPaths[0] != want[0] || gotPaths[1] != want[1] {
		t.Fatalf("gotPaths = %v, want %v", gotPaths, want)
	}
	if len(records) != 2 || records[0]["project_id"] != "a" || records[1]["project_id"] != "b" {
		t.Fatalf("records = %+v, want project_id stamped a/b", records)
	}
}

// TestReadFanOutRequestIDsPreliminaryPaginatedRequest proves the
// ids_from.request form: a preliminary GET is issued (paginated to
// exhaustion using the stream's OWN pagination spec) BEFORE any per-id
// sub-sequence starts, and the extracted id_field values drive the fan-out.
func TestReadFanOutRequestIDsPreliminaryPaginatedRequest(t *testing.T) {
	var seq []string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/projects" && r.URL.Query().Get("page") == "" || r.URL.Path == "/projects" && r.URL.Query().Get("page") == "1":
			seq = append(seq, "list-page-1")
			_, _ = w.Write([]byte(`{"data":[{"id":"p1"},{"id":"p2"}]}`))
		case r.URL.Path == "/projects" && r.URL.Query().Get("page") == "2":
			seq = append(seq, "list-page-2")
			_, _ = w.Write([]byte(`{"data":[{"id":"p3"}]}`))
		case r.URL.Path == "/projects/p1/tasks":
			seq = append(seq, "tasks-p1")
			_, _ = w.Write([]byte(`{"data":[{"id":"t1"}]}`))
		case r.URL.Path == "/projects/p2/tasks":
			seq = append(seq, "tasks-p2")
			_, _ = w.Write([]byte(`{"data":[{"id":"t2"}]}`))
		case r.URL.Path == "/projects/p3/tasks":
			seq = append(seq, "tasks-p3")
			_, _ = w.Write([]byte(`{"data":[{"id":"t3"}]}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.URL.Path, r.URL.RawQuery)
		}
	})
	b := newTestBundle(t, srv, StreamSpec{
		Path:       "/projects/{{ fanout.id }}/tasks",
		Records:    RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{Type: "page_number", PageParam: "page", PageSize: 2},
		FanOut: &FanOutSpec{
			IDsFrom: FanOutIDsFrom{Request: &FanOutIDsRequest{
				Path:        "/projects",
				RecordsPath: "data",
				IDField:     "id",
			}},
			Into:       FanOutInto{PathVar: "parent_id"},
			StampField: "project_id",
		},
	})
	// The child stream's own pagination (page_number, page_size 2) would make
	// each per-id tasks sub-sequence request 2 pages if it ever saw a full
	// page — but the id-listing pass has NOTHING to do with the child
	// stream's pagination being reused for the CHILD requests below (this
	// test's per-id tasks responses are always short/1-record, so the child
	// sub-sequences stop after page 1 regardless).

	req := connectors.ReadRequest{Stream: "widgets"}
	records, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantSeq := []string{"list-page-1", "list-page-2", "tasks-p1", "tasks-p2", "tasks-p3"}
	if len(seq) != len(wantSeq) {
		t.Fatalf("seq = %v, want %v", seq, wantSeq)
	}
	for i := range wantSeq {
		if seq[i] != wantSeq[i] {
			t.Fatalf("seq = %v, want %v (id-listing must fully exhaust BEFORE any per-id sub-sequence starts)", seq, wantSeq)
		}
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (one task per project)", len(records))
	}
}

// TestReadFanOutEachIDSequenceOwnPagination proves each id's sub-sequence
// paginates independently: id A's sequence needs 2 pages, id B's needs 1.
func TestReadFanOutEachIDSequenceOwnPagination(t *testing.T) {
	pageHitsByID := map[string]int{}
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("apps_id")
		page := r.URL.Query().Get("page")
		pageHitsByID[id+":"+page]++
		switch {
		case id == "A" && (page == "" || page == "1"):
			_, _ = w.Write([]byte(`{"data":[{"id":"1"},{"id":"2"}]}`))
		case id == "A" && page == "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"3"}]}`)) // short page: stop
		case id == "B" && (page == "" || page == "1"):
			_, _ = w.Write([]byte(`{"data":[{"id":"1"}]}`)) // short page: stop
		default:
			t.Fatalf("unexpected request id=%q page=%q", id, page)
		}
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records:    RecordsSpec{Path: "data"},
		Pagination: &PaginationSpec{Type: "page_number", PageParam: "page", PageSize: 2},
		FanOut: &FanOutSpec{
			IDsFrom: FanOutIDsFrom{ConfigKey: "ids"},
			Into:    FanOutInto{QueryParam: "apps_id"},
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"ids": "A,B"}}}
	records, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if pageHitsByID["A:1"] != 1 || pageHitsByID["A:2"] != 1 {
		t.Fatalf("pageHitsByID = %v, want id A to fetch exactly page 1 and page 2 once each", pageHitsByID)
	}
	if pageHitsByID["B:1"] != 1 {
		t.Fatalf("pageHitsByID = %v, want id B to fetch exactly page 1 once", pageHitsByID)
	}
	if len(records) != 4 {
		t.Fatalf("got %d records, want 4 (3 from id A across 2 pages + 1 from id B)", len(records))
	}
}

// TestReadFanOutEmptyIDListEmitsNothing proves an empty CSV config value
// resolves to zero fan-out ids: no request is ever issued and Read returns
// no error (an empty configured id list, e.g. no apps configured yet, is not
// itself a failure).
func TestReadFanOutEmptyIDListEmitsNothing(t *testing.T) {
	var hits int
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		FanOut: &FanOutSpec{
			IDsFrom: FanOutIDsFrom{ConfigKey: "app_collection_ids"},
			Into:    FanOutInto{QueryParam: "apps_id"},
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"app_collection_ids": ""}}}
	records, err := readAll(t, context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("Read: %v (an empty fan-out id list must never error)", err)
	}
	if hits != 0 {
		t.Fatalf("hits = %d, want 0 (no requests when the id list resolves empty)", hits)
	}
	if len(records) != 0 {
		t.Fatalf("records = %+v, want none", records)
	}
}

// TestReadFanOutMissingIDsFromIsReadError proves a fan_out block declaring
// NEITHER config_key nor request is a read-time error (mirroring cursor
// pagination's mutual-exclusivity error shape), not a silent no-op.
func TestReadFanOutMissingIDsFromIsReadError(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should ever be issued: %s", r.URL.String())
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		FanOut:  &FanOutSpec{Into: FanOutInto{QueryParam: "apps_id"}},
	})

	_, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err == nil {
		t.Fatalf("Read: want error when fan_out.ids_from declares neither config_key nor request")
	}
}

// TestReadFanOutBothIDsFromFormsIsReadError proves declaring BOTH config_key
// and request is also a read-time error, not a silently-picks-one shape.
func TestReadFanOutBothIDsFromFormsIsReadError(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request should ever be issued: %s", r.URL.String())
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		FanOut: &FanOutSpec{
			IDsFrom: FanOutIDsFrom{
				ConfigKey: "ids",
				Request:   &FanOutIDsRequest{Path: "/x", RecordsPath: "data", IDField: "id"},
			},
			Into: FanOutInto{QueryParam: "apps_id"},
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"ids": "1"}}}
	_, err := readAll(t, context.Background(), b, req, nil)
	if err == nil {
		t.Fatalf("Read: want error when fan_out.ids_from declares BOTH config_key and request")
	}
}

// TestReadFanOutConfigKeyIDsAreTrimmed proves the CSV split trims whitespace
// and drops empty entries (e.g. a trailing comma or accidental spaces).
func TestReadFanOutConfigKeyIDsAreTrimmed(t *testing.T) {
	var gotIDs []string
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotIDs = append(gotIDs, r.URL.Query().Get("apps_id"))
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: "data"},
		FanOut: &FanOutSpec{
			IDsFrom: FanOutIDsFrom{ConfigKey: "ids"},
			Into:    FanOutInto{QueryParam: "apps_id"},
		},
	})

	req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"ids": " 111 , 222,,333 "}}}
	if _, err := readAll(t, context.Background(), b, req, nil); err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := []string{"111", "222", "333"}
	if len(gotIDs) != len(want) {
		t.Fatalf("gotIDs = %v, want %v", gotIDs, want)
	}
	for i := range want {
		if gotIDs[i] != want[i] {
			t.Fatalf("gotIDs = %v, want %v", gotIDs, want)
		}
	}
}

// --- S4 engine mini-wave item 3: keyed-object flatten -----------------------

// TestRecordsAtKeyedObjectFlattensEachValue proves each value of the JSON
// object at path becomes its own record (appfigures' products/mine shape:
// {"111":{...},"222":{...}}).
func TestRecordsAtKeyedObjectFlattensEachValue(t *testing.T) {
	body := []byte(`{"products":{"111":{"name":"a"},"222":{"name":"b"}}}`)
	records, err := recordsAtKeyed(body, "products", "")
	if err != nil {
		t.Fatalf("recordsAtKeyed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	names := map[string]bool{}
	for _, r := range records {
		names[fmt.Sprint(r["name"])] = true
	}
	if !names["a"] || !names["b"] {
		t.Fatalf("records = %+v, want values a and b both present", records)
	}
}

// TestRecordsAtKeyedObjectStampsKeyField proves the map key is stamped onto
// keyField before projection.
func TestRecordsAtKeyedObjectStampsKeyField(t *testing.T) {
	body := []byte(`{"products":{"111":{"name":"a"},"222":{"name":"b"}}}`)
	records, err := recordsAtKeyed(body, "products", "product_id")
	if err != nil {
		t.Fatalf("recordsAtKeyed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	byID := map[string]connsdk.Record{}
	for _, r := range records {
		byID[fmt.Sprint(r["product_id"])] = r
	}
	if byID["111"]["name"] != "a" || byID["222"]["name"] != "b" {
		t.Fatalf("records = %+v, want product_id stamped 111/222", records)
	}
}

// TestRecordsAtKeyedObjectSortedByKeyForDeterminism proves emission order is
// sorted ascending by key (map iteration is random in Go; parity/test
// stability requires deterministic order).
func TestRecordsAtKeyedObjectSortedByKeyForDeterminism(t *testing.T) {
	body := []byte(`{"products":{"c":{"n":3},"a":{"n":1},"b":{"n":2}}}`)
	records, err := recordsAtKeyed(body, "products", "key")
	if err != nil {
		t.Fatalf("recordsAtKeyed: %v", err)
	}
	want := []string{"a", "b", "c"}
	if len(records) != len(want) {
		t.Fatalf("got %d records, want %d", len(records), len(want))
	}
	for i, k := range want {
		if fmt.Sprint(records[i]["key"]) != k {
			t.Fatalf("records[%d][key] = %v, want %q (sorted ascending)", i, records[i]["key"], k)
		}
	}
}

// TestRecordsAtKeyedObjectSkipsNonObjectValues proves a value that is a
// scalar/array is dropped, not a crash/error.
func TestRecordsAtKeyedObjectSkipsNonObjectValues(t *testing.T) {
	body := []byte(`{"products":{"a":{"n":1},"b":"not an object","c":[1,2,3],"d":null}}`)
	records, err := recordsAtKeyed(body, "products", "key")
	if err != nil {
		t.Fatalf("recordsAtKeyed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1 (only key %q is an object)", len(records), "a")
	}
	if fmt.Sprint(records[0]["key"]) != "a" {
		t.Fatalf("records[0][key] = %v, want %q", records[0]["key"], "a")
	}
}

// TestRecordsAtKeyedObjectEmptyObjectYieldsNoRecords proves an empty object
// at path yields zero records, not an error.
func TestRecordsAtKeyedObjectEmptyObjectYieldsNoRecords(t *testing.T) {
	body := []byte(`{"products":{}}`)
	records, err := recordsAtKeyed(body, "products", "")
	if err != nil {
		t.Fatalf("recordsAtKeyed: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("got %d records, want 0", len(records))
	}
}

// TestReadKeyedObjectStreamEmitsFlattenedProjectedRecords proves the full
// Read() path wires records.keyed_object/key_field through extractRecords,
// including projection and computed_fields on the exploded records.
func TestReadKeyedObjectStreamEmitsFlattenedProjectedRecords(t *testing.T) {
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"111":{"name":"widget-a"},"222":{"name":"widget-b"}}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Records: RecordsSpec{Path: ".", KeyedObject: true, KeyField: "id"},
	})

	records, err := readAll(t, context.Background(), b, connectors.ReadRequest{Stream: "widgets"}, nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0]["id"] != "111" || records[0]["name"] != "widget-a" {
		t.Fatalf("records[0] = %+v, want id=111 name=widget-a", records[0])
	}
	if records[1]["id"] != "222" || records[1]["name"] != "widget-b" {
		t.Fatalf("records[1] = %+v, want id=222 name=widget-b", records[1])
	}
}
