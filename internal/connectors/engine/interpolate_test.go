package engine

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func baseVars() Vars {
	return Vars{
		Config: map[string]string{
			"base_url":   "https://api.example.com",
			"repository": "a/../b",
			"query_char": "a?x=1&y=2",
			"space":      "a b",
			"unicode":    "héllo",
			"double_enc": "%2e%2e",
			"auth_type":  "token",
			"empty":      "",
		},
		Secrets: map[string]string{
			"token": "sekret-token",
		},
		Record: map[string]any{
			"user": map[string]any{
				"login": "octocat",
			},
			"created_at": "2024-01-02T03:04:05Z",
		},
		Cursor: "cursor-value",
	}
}

func TestInterpolateResolution(t *testing.T) {
	vars := baseVars()

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"config", "{{ config.base_url }}", "https://api.example.com"},
		{"secrets", "{{ secrets.token }}", "sekret-token"},
		{"record dotted path", "{{ record.user.login }}", "octocat"},
		{"cursor", "{{ cursor }}", "cursor-value"},
		{"literal text passthrough", "prefix-{{ cursor }}-suffix", "prefix-cursor-value-suffix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Interpolate(tt.template, vars)
			if err != nil {
				t.Fatalf("Interpolate error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Interpolate(%q) = %q, want %q", tt.template, got, tt.want)
			}
		})
	}
}

func TestInterpolateMissingKey(t *testing.T) {
	vars := baseVars()
	_, err := Interpolate("{{ config.nope }}", vars)
	if err == nil {
		t.Fatalf("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "config") {
		t.Fatalf("error %q does not name key+namespace", err.Error())
	}

	_, err = Interpolate("{{ secrets.nope }}", vars)
	if err == nil {
		t.Fatalf("expected error for missing secret key")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "secrets") {
		t.Fatalf("error %q does not name key+namespace", err.Error())
	}
}

func TestInterpolatePathDefaultURLEncode(t *testing.T) {
	vars := baseVars()

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "path traversal encoded",
			template: "/repos/{{ config.repository }}",
			want:     "/repos/a%2F..%2Fb",
		},
		{
			name:     "query metachars encoded",
			template: "/x/{{ config.query_char }}",
			want:     "/x/a%3Fx%3D1%26y%3D2",
		},
		{
			name:     "space and unicode encoded",
			template: "/{{ config.space }}/{{ config.unicode }}",
			want:     "/a%20b/h%C3%A9llo",
		},
		{
			name:     "double encode guard: percent literal is re-encoded",
			template: "/{{ config.double_enc }}",
			want:     "/%252e%252e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InterpolatePath(tt.template, vars)
			if err != nil {
				t.Fatalf("InterpolatePath error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("InterpolatePath(%q) = %q, want %q", tt.template, got, tt.want)
			}
		})
	}
}

func TestInterpolateFilters(t *testing.T) {
	vars := baseVars()

	t.Run("unix_seconds on rfc3339", func(t *testing.T) {
		got, err := Interpolate("{{ record.created_at | unix_seconds }}", vars)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "1704164645" {
			t.Fatalf("got %q, want unix seconds for 2024-01-02T03:04:05Z", got)
		}
	})

	t.Run("unix_seconds on bad input errors", func(t *testing.T) {
		v := baseVars()
		v.Config["bad_date"] = "not-a-date"
		_, err := Interpolate("{{ config.bad_date | unix_seconds }}", v)
		if err == nil {
			t.Fatalf("expected error for bad date input")
		}
	})

	t.Run("base64", func(t *testing.T) {
		got, err := Interpolate("{{ secrets.token | base64 }}", vars)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "c2VrcmV0LXRva2Vu" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("explicit urlencode filter in non-path context", func(t *testing.T) {
		got, err := Interpolate("{{ config.space | urlencode }}", vars)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "a%20b" {
			t.Fatalf("got %q", got)
		}
	})
}

// --- F9 (REVIEW.md flag): chained filters must apply in sequence, not
// silently truncate after the first pipe segment. ---

func TestInterpolateMultipleFiltersChained(t *testing.T) {
	vars := baseVars()
	// space -> urlencode ("a%20b") -> base64 of THAT string. Before the fix,
	// resolveExpr only looks at parts[1] ("urlencode"), silently dropping the
	// "base64" stage, so the result would just be "a%20b" (unchained).
	got, err := Interpolate("{{ config.space | urlencode | base64 }}", vars)
	if err != nil {
		t.Fatalf("Interpolate chained filters: unexpected error: %v", err)
	}
	want := base64Encode("a%20b")
	if got != want {
		t.Fatalf("Interpolate(chained urlencode|base64) = %q, want %q (both filters must apply in order)", got, want)
	}
}

func TestInterpolateMultipleFiltersUnknownNameStillErrors(t *testing.T) {
	vars := baseVars()
	_, err := Interpolate("{{ config.space | urlencode | bogus_filter }}", vars)
	if err == nil {
		t.Fatalf("Interpolate: expected error for unknown filter name in a chain, got nil (no silent truncation/skip)")
	}
	if !strings.Contains(err.Error(), "bogus_filter") {
		t.Fatalf("error %q does not name the unknown filter", err.Error())
	}
}

func TestApplyFilterUnknownFilterNameErrors(t *testing.T) {
	_, err := applyFilter("nonexistent", "x")
	if err == nil {
		t.Fatalf("applyFilter: expected error for unknown filter name")
	}
}

// --- join:<sep> filter (F7 meta-rule enablement: array -> separator-joined
// string, e.g. searxng's "engines" array vs. legacy's comma-joined string). ---

func TestApplyFilterJoinSeparator(t *testing.T) {
	vars := baseVars()
	vars.Record["tags"] = []any{"a", "b", "c"}
	got, err := Interpolate("{{ record.tags | join:, }}", vars)
	if err != nil {
		t.Fatalf("Interpolate join filter: unexpected error: %v", err)
	}
	if got != "a,b,c" {
		t.Fatalf("Interpolate(join:,) = %q, want a,b,c", got)
	}
}

func TestApplyFilterJoinCustomSeparator(t *testing.T) {
	// "|" itself cannot be used as the join separator: it is the filter-chain
	// delimiter in this dialect's grammar, so "join:|" would be ambiguous
	// with "| <next filter>". Any other separator (including multi-char) is
	// fine.
	vars := baseVars()
	vars.Record["tags"] = []any{"x", "y"}
	got, err := Interpolate("{{ record.tags | join:; }}", vars)
	if err != nil {
		t.Fatalf("Interpolate join filter: unexpected error: %v", err)
	}
	if got != "x;y" {
		t.Fatalf("Interpolate(join:;) = %q, want x;y", got)
	}
}

func TestApplyFilterJoinNonArrayErrors(t *testing.T) {
	vars := baseVars()
	_, err := Interpolate("{{ config.space | join:, }}", vars)
	if err == nil {
		t.Fatalf("Interpolate join filter on a non-array value: expected error, got nil")
	}
}

// --- last_path_segment filter (gap-loop item 4, REVIEW-B.md finding 1 /
// cross-cutting adjudication 1): calendly's dropped derived `id` field
// (legacy idFromURI(uri) — the trailing segment of a HAL/URI-shaped field)
// and every other URI-keyed API. ---

func TestApplyFilterLastPathSegment(t *testing.T) {
	vars := baseVars()
	vars.Record = map[string]any{"uri": "https://api.calendly.com/scheduled_events/AAAAAAAAAAAAAAAA"}
	got, err := Interpolate("{{ record.uri | last_path_segment }}", vars)
	if err != nil {
		t.Fatalf("Interpolate last_path_segment: unexpected error: %v", err)
	}
	if got != "AAAAAAAAAAAAAAAA" {
		t.Fatalf("Interpolate(last_path_segment) = %q, want AAAAAAAAAAAAAAAA", got)
	}
}

// TestApplyFilterLastPathSegmentTrailingSlashIgnored proves a trailing
// slash on the source value does not produce an empty last segment (a
// defensive edge case no legacy URI is expected to hit, but worth locking
// down since idFromURI-style helpers commonly get this wrong).
func TestApplyFilterLastPathSegmentTrailingSlashIgnored(t *testing.T) {
	vars := baseVars()
	vars.Record = map[string]any{"uri": "https://api.calendly.com/scheduled_events/AAAAAAAAAAAAAAAA/"}
	got, err := Interpolate("{{ record.uri | last_path_segment }}", vars)
	if err != nil {
		t.Fatalf("Interpolate last_path_segment: unexpected error: %v", err)
	}
	if got != "AAAAAAAAAAAAAAAA" {
		t.Fatalf("Interpolate(last_path_segment) = %q, want AAAAAAAAAAAAAAAA (trailing slash ignored)", got)
	}
}

// TestApplyFilterLastPathSegmentNoSlashReturnsWholeValue proves a value with
// no "/" at all (nothing to split) passes through unchanged rather than
// erroring — the filter degrades gracefully for a bare-id source field.
func TestApplyFilterLastPathSegmentNoSlashReturnsWholeValue(t *testing.T) {
	vars := baseVars()
	vars.Record = map[string]any{"id": "AAAAAAAAAAAAAAAA"}
	got, err := Interpolate("{{ record.id | last_path_segment }}", vars)
	if err != nil {
		t.Fatalf("Interpolate last_path_segment: unexpected error: %v", err)
	}
	if got != "AAAAAAAAAAAAAAAA" {
		t.Fatalf("Interpolate(last_path_segment) = %q, want AAAAAAAAAAAAAAAA (no slash, whole value)", got)
	}
}

// TestApplyFilterLastPathSegmentKnownToResolveCheck proves the new filter
// name is accepted by ResolveCheck's static filter-name validation (F9) —
// connectorgen validate must not flag a bundle using last_path_segment as an
// "unknown filter" typo.
func TestApplyFilterLastPathSegmentKnownToResolveCheck(t *testing.T) {
	if err := ResolveCheck("{{ record.uri | last_path_segment }}", map[string]bool{}); err != nil {
		t.Fatalf("ResolveCheck: unexpected error for known filter last_path_segment: %v", err)
	}
}

// --- const:<value> filter (S3 engine mini-wave item 1, wave1-pilot
// SUMMARY.md carried queue / REVIEW-A.md re-review R2): send a FIXED literal
// value iff a reference resolves, without leaking or otherwise depending on
// the resolved value itself — chargebee's sort_by[asc]=updated_at is always
// the constant string "updated_at", gated on whether the incremental lower
// bound resolves at all, not templated FROM the lower bound's own value. ---

func TestApplyFilterConstReplacesResolvedValueWithFixedLiteral(t *testing.T) {
	vars := baseVars()
	got, err := Interpolate("{{ config.base_url | const:updated_at }}", vars)
	if err != nil {
		t.Fatalf("Interpolate const filter: unexpected error: %v", err)
	}
	if got != "updated_at" {
		t.Fatalf("Interpolate(const:updated_at) = %q, want updated_at (fixed literal, ignoring the resolved base_url value)", got)
	}
}

func TestApplyFilterConstStillFailsWhenReferenceUnresolved(t *testing.T) {
	vars := baseVars()
	_, err := Interpolate("{{ config.nope | const:updated_at }}", vars)
	if err == nil {
		t.Fatalf("Interpolate const filter: expected error for an unresolved reference (const only replaces the VALUE, it never suppresses a resolution failure)")
	}
}

func TestApplyFilterConstComposesWithOmitWhenAbsentQueryDialect(t *testing.T) {
	// This is the actual production shape: gate on incremental.lower_bound's
	// presence (via omit_when_absent), but SEND a fixed literal, not the
	// lower bound's own value.
	var gotQuery url.Values
	srv := jsonServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	b := newTestBundle(t, srv, StreamSpec{
		Query: map[string]QueryParam{
			"sort_by[asc]": {Template: "{{ incremental.lower_bound | const:updated_at }}", OmitWhenAbsent: true},
		},
		Incremental: &IncrementalSpec{CursorField: "updated_at", RequestParam: "updated_at[after]", ParamFormat: "unix_seconds"},
	})

	req := connectors.ReadRequest{Stream: "widgets", State: map[string]string{"cursor": "1700000100"}}
	if _, err := readAll(t, context.Background(), b, req, nil); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if gotQuery.Get("sort_by[asc]") != "updated_at" {
		t.Fatalf("query sort_by[asc] = %q, want the fixed literal %q (not the lower bound value %q)", gotQuery.Get("sort_by[asc]"), "updated_at", "1700000100")
	}
}

func TestApplyFilterConstKnownToResolveCheck(t *testing.T) {
	if err := ResolveCheck("{{ incremental.lower_bound | const:updated_at }}", map[string]bool{}); err != nil {
		t.Fatalf("ResolveCheck: unexpected error for known filter const: %v", err)
	}
}

// --- static-literal computed_fields (no {{ }} markers at all): already
// supported by Interpolate's no-op-on-no-match semantics; locked in here as
// a named regression test per F7's meta-rule enablement. ---

func TestInterpolateStaticLiteralNoTemplateMarkersPassesThroughVerbatim(t *testing.T) {
	got, err := Interpolate("searxng", baseVars())
	if err != nil {
		t.Fatalf("Interpolate(no markers): unexpected error: %v", err)
	}
	if got != "searxng" {
		t.Fatalf("Interpolate(no markers) = %q, want verbatim literal %q", got, "searxng")
	}
}

// --- F9: InterpolatePath must reject a resolved segment that is exactly
// ".." (or a raw "/../" survives after decode-normalization), closing the
// SECURITY-REVIEW.md F9b/m3 gap where urlencodeSegment leaves bare "." (and
// therefore "..") unescaped, so a literal ".." segment round-trips intact. ---

func TestInterpolatePathRejectsDotDotSegment(t *testing.T) {
	vars := baseVars()
	vars.Config["traversal_id"] = ".."
	_, err := InterpolatePath("/customers/{{ config.traversal_id }}", vars)
	if err == nil {
		t.Fatalf("InterpolatePath: expected error for a resolved value of exactly \"..\", got nil (path traversal must not survive as an intact segment)")
	}
}

func TestInterpolatePathRejectsDotDotAmongMultipleSegments(t *testing.T) {
	vars := baseVars()
	vars.Config["mid"] = ".."
	_, err := InterpolatePath("/a/{{ config.mid }}/b", vars)
	if err == nil {
		t.Fatalf("InterpolatePath: expected error for a \"..\" segment in the middle of a path template")
	}
}

func TestInterpolatePathSingleDotSegmentStillAllowed(t *testing.T) {
	// A single "." (not "..") is not a traversal primitive on its own in the
	// same dangerous sense; only ".." is rejected outright per the fix scope.
	vars := baseVars()
	vars.Config["dot"] = "."
	got, err := InterpolatePath("/customers/{{ config.dot }}", vars)
	if err != nil {
		t.Fatalf("InterpolatePath: unexpected error for a lone \".\" segment: %v", err)
	}
	if got != "/customers/." {
		t.Fatalf("InterpolatePath(lone dot) = %q, want /customers/.", got)
	}
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func TestInterpolateHeaderCRLFInjectionRejected(t *testing.T) {
	vars := baseVars()
	vars.Config["evil"] = "value\r\nX-Injected: true"

	_, err := InterpolateHeader("{{ config.evil }}", vars)
	if err == nil {
		t.Fatalf("expected error for CRLF injection in header value")
	}

	vars.Config["evil_path"] = "a\r\nb"
	_, err = InterpolatePath("/{{ config.evil_path }}", vars)
	if err == nil {
		t.Fatalf("expected error for CRLF injection in path value")
	}
}

func TestEvalWhen(t *testing.T) {
	vars := baseVars()

	tests := []struct {
		name string
		cond string
		want bool
	}{
		{"equality true", "{{ config.auth_type == 'token' }}", true},
		{"equality false", "{{ config.auth_type == 'public' }}", false},
		{"in list true", "{{ config.auth_type in ['auto', 'token'] }}", true},
		{"in list false", "{{ config.auth_type in ['auto', 'public'] }}", false},
		{"truthiness non-empty", "{{ config.base_url }}", true},
		{"truthiness empty", "{{ config.empty }}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalWhen(tt.cond, vars)
			if err != nil {
				t.Fatalf("EvalWhen error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("EvalWhen(%q) = %v, want %v", tt.cond, got, tt.want)
			}
		})
	}
}

func TestEvalWhenUnknownOperatorIsCompileError(t *testing.T) {
	vars := baseVars()
	_, err := EvalWhen("{{ config.auth_type >= 'token' }}", vars)
	if err == nil {
		t.Fatalf("expected compile error for unknown operator")
	}
}

// TestEvalWhenAbsentKeyEvaluatesFalsy proves that, unlike general template
// interpolation (Interpolate/InterpolatePath/InterpolateHeader — which still
// hard-error on any unresolved config/secrets key), a `when` condition
// referencing a config/secrets key that is entirely ABSENT from vars (not
// merely empty-string) evaluates as falsy rather than erroring. This is the
// OPTIONAL-credential pattern: `when: "{{ secrets.api_key }}"` must be able to
// gate an auth spec off when the caller never populated that secret at all,
// without the bundle author needing a companion "is this key present"
// primitive the dialect doesn't have.
func TestEvalWhenAbsentKeyEvaluatesFalsy(t *testing.T) {
	vars := baseVars() // vars.Secrets has "token" only; vars.Config has no "api_key" or "missing_cfg" key.

	tests := []struct {
		name string
		cond string
		want bool
	}{
		{"truthiness: absent secret key", "{{ secrets.api_key }}", false},
		{"truthiness: absent config key", "{{ config.missing_cfg }}", false},
		{"equality: absent secret key compares as empty string, mismatch", "{{ secrets.api_key == 'sekret-token' }}", false},
		{"equality: absent secret key compares as empty string, match empty literal", "{{ secrets.api_key == '' }}", true},
		{"equality: absent config key compares as empty string", "{{ config.missing_cfg == 'anything' }}", false},
		{"in: absent secret key is not contained in any non-empty list", "{{ secrets.api_key in ['a', 'b'] }}", false},
		{"in: absent config key is not contained even in a list containing empty string", "{{ config.missing_cfg in ['', 'x'] }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalWhen(tt.cond, vars)
			if err != nil {
				t.Fatalf("EvalWhen(%q) unexpected error: %v (absent key must evaluate falsy in a when-condition, not error)", tt.cond, err)
			}
			if got != tt.want {
				t.Fatalf("EvalWhen(%q) = %v, want %v", tt.cond, got, tt.want)
			}
		})
	}
}

// TestEvalWhenAbsentKeyDoesNotLeakIntoGeneralInterpolation proves the fix is
// scoped to when-evaluation only: ordinary Interpolate (and its path/header
// variants) must STILL hard-error on an unresolved config/secrets key. A
// when-condition's absent-key tolerance must never leak into general
// template resolution (bearer tokens, URLs, headers, query params, etc.).
func TestEvalWhenAbsentKeyDoesNotLeakIntoGeneralInterpolation(t *testing.T) {
	vars := baseVars()

	if _, err := Interpolate("{{ secrets.api_key }}", vars); err == nil {
		t.Fatalf("Interpolate: expected error for unresolved secrets key outside a when-condition")
	}
	if _, err := Interpolate("{{ config.missing_cfg }}", vars); err == nil {
		t.Fatalf("Interpolate: expected error for unresolved config key outside a when-condition")
	}
}

// TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates proves static
// validation is unaffected by the when-absent-key runtime fix: a `when`
// template referencing a key NOT declared in spec.json's properties (a typo)
// is still caught at connectorgen validate / ResolveCheck time, exactly like
// any other config reference. Only RUNTIME absence of a spec-KNOWN key is
// tolerated by EvalWhen — spec-unknown keys remain a hard validate-time
// error.
func TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates(t *testing.T) {
	specKeys := map[string]bool{"api_key": true, "base_url": true}

	// A when-template referencing a spec-KNOWN key (api_key) passes static
	// validation even though it may be absent at runtime (that's exactly the
	// optional-credential pattern this fix enables).
	if err := ResolveCheck("{{ config.api_key }}", specKeys); err != nil {
		t.Fatalf("ResolveCheck: unexpected error for spec-known when-template key: %v", err)
	}

	// A when-template referencing a spec-UNKNOWN key (typo) must still fail
	// static validation.
	err := ResolveCheck("{{ config.api_kay }}", specKeys)
	if err == nil {
		t.Fatalf("ResolveCheck: expected validation finding for spec-unknown when-template key (typo)")
	}
	if !strings.Contains(err.Error(), "api_kay") {
		t.Fatalf("error %q does not name the offending key", err.Error())
	}
}

func TestResolveCheck(t *testing.T) {
	specKeys := map[string]bool{"repository": true, "base_url": true}

	if err := ResolveCheck("/repos/{{ config.repository }}", specKeys); err != nil {
		t.Fatalf("unexpected error for known key: %v", err)
	}

	err := ResolveCheck("/repos/{{ config.unknown_key }}", specKeys)
	if err == nil {
		t.Fatalf("expected validation finding for unknown spec key")
	}
	if !strings.Contains(err.Error(), "unknown_key") {
		t.Fatalf("error %q does not name the offending key", err.Error())
	}

	// record/cursor/secrets references are not checked against specKeys.
	if err := ResolveCheck("{{ record.user.login }}", specKeys); err != nil {
		t.Fatalf("unexpected error for record reference: %v", err)
	}
	if err := ResolveCheck("{{ cursor }}", specKeys); err != nil {
		t.Fatalf("unexpected error for cursor reference: %v", err)
	}
}

// TestResolveCheckAcceptsIncrementalLowerBoundReference proves ResolveCheck
// statically accepts "{{ incremental.lower_bound }}" (S3 engine mini-wave
// item 1): it is not a spec.json-declared key (like record/cursor/secrets,
// it is not checked against specKeys), and an unknown key under the
// "incremental" namespace (a typo, e.g. "incremental.lowre_bound") is still a
// validate-time error naming the offending reference.
func TestResolveCheckAcceptsIncrementalLowerBoundReference(t *testing.T) {
	specKeys := map[string]bool{"repository": true}

	if err := ResolveCheck("{{ incremental.lower_bound }}", specKeys); err != nil {
		t.Fatalf("ResolveCheck: unexpected error for incremental.lower_bound reference: %v", err)
	}

	err := ResolveCheck("{{ incremental.lowre_bound }}", specKeys)
	if err == nil {
		t.Fatalf("ResolveCheck: expected validation finding for unknown incremental key (typo)")
	}
	if !strings.Contains(err.Error(), "lowre_bound") {
		t.Fatalf("error %q does not name the offending key", err.Error())
	}
}

// --- ResolveCheckWhen: full when-grammar (==, in, truthiness) static parsing
// (S3 engine mini-wave item 2, wave1-pilot SUMMARY.md carried queue /
// REVIEW-A.md re-review R1/R3): ResolveCheck's bare namespace.key-only
// parsing previously forced a `when: "{{ config.auth_type == 'public' }}"`
// clause to hard-fail validate even when auth_type IS a declared spec key,
// since ResolveCheck split the WHOLE inner expression (including "== 'public'")
// on "." as if it were one reference. ResolveCheckWhen parses the identical
// grammar EvalWhen evaluates at runtime and statically validates only the
// left-hand-side reference (plus literal/list syntax), so `==`/`in`-shaped
// when clauses referencing a real spec key now pass, while a typo'd/
// spec-unknown key or malformed literal/list syntax is still a validate-time
// error. ---

func TestResolveCheckWhenEqualityAgainstSpecKnownKeyPasses(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	if err := ResolveCheckWhen("{{ config.auth_type == 'public' }}", specKeys); err != nil {
		t.Fatalf("ResolveCheckWhen: unexpected error for spec-known key in == comparison: %v", err)
	}
}

func TestResolveCheckWhenInMembershipAgainstSpecKnownKeyPasses(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	if err := ResolveCheckWhen("{{ config.auth_type in ['public', 'none', 'anonymous'] }}", specKeys); err != nil {
		t.Fatalf("ResolveCheckWhen: unexpected error for spec-known key in `in` comparison: %v", err)
	}
}

func TestResolveCheckWhenTruthinessStillWorksUnchanged(t *testing.T) {
	specKeys := map[string]bool{"public_access": true}
	if err := ResolveCheckWhen("{{ config.public_access }}", specKeys); err != nil {
		t.Fatalf("ResolveCheckWhen: unexpected error for bare truthiness reference: %v", err)
	}
}

func TestResolveCheckWhenEqualityAgainstSpecUnknownKeyFails(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	err := ResolveCheckWhen("{{ config.auth_typo == 'public' }}", specKeys)
	if err == nil {
		t.Fatalf("ResolveCheckWhen: expected validation finding for spec-unknown key in == comparison")
	}
	if !strings.Contains(err.Error(), "auth_typo") {
		t.Fatalf("error %q does not name the offending key", err.Error())
	}
}

func TestResolveCheckWhenInMembershipAgainstSpecUnknownKeyFails(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	err := ResolveCheckWhen("{{ config.auth_typo in ['public','none'] }}", specKeys)
	if err == nil {
		t.Fatalf("ResolveCheckWhen: expected validation finding for spec-unknown key in `in` comparison")
	}
	if !strings.Contains(err.Error(), "auth_typo") {
		t.Fatalf("error %q does not name the offending key", err.Error())
	}
}

func TestResolveCheckWhenMalformedLiteralFails(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	err := ResolveCheckWhen("{{ config.auth_type == public }}", specKeys) // missing quotes
	if err == nil {
		t.Fatalf("ResolveCheckWhen: expected validation finding for unquoted == literal")
	}
}

func TestResolveCheckWhenMalformedListFails(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	err := ResolveCheckWhen("{{ config.auth_type in 'public','none' }}", specKeys) // missing brackets
	if err == nil {
		t.Fatalf("ResolveCheckWhen: expected validation finding for malformed `in` list (missing brackets)")
	}
}

func TestResolveCheckWhenUnsupportedOperatorFails(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	err := ResolveCheckWhen("{{ config.auth_type != 'public' }}", specKeys)
	if err == nil {
		t.Fatalf("ResolveCheckWhen: expected validation finding for unsupported operator !=")
	}
}

// TestResolveCheckAuthSpecWhenUsesFullGrammar proves ResolveCheckAuthSpec's
// `when` field is checked via the SAME when-grammar-aware path (not plain
// ResolveCheck), so an ==/in-shaped when clause on a real AuthSpec passes
// static validation when its key is spec-known.
func TestResolveCheckAuthSpecWhenUsesFullGrammar(t *testing.T) {
	specKeys := map[string]bool{"auth_type": true}
	spec := AuthSpec{Mode: "none", When: "{{ config.auth_type == 'public' }}"}
	if err := ResolveCheckAuthSpec(spec, specKeys); err != nil {
		t.Fatalf("ResolveCheckAuthSpec: unexpected error for spec-known key in == when clause: %v", err)
	}

	badSpec := AuthSpec{Mode: "none", When: "{{ config.auth_typo == 'public' }}"}
	err := ResolveCheckAuthSpec(badSpec, specKeys)
	if err == nil {
		t.Fatalf("ResolveCheckAuthSpec: expected validation finding for spec-unknown key in == when clause")
	}
	if !strings.Contains(err.Error(), "auth_typo") {
		t.Fatalf("error %q does not name the offending key", err.Error())
	}
}

// TestResolveCheckFilterNameValidation proves ResolveCheck also statically
// validates every filter stage in a (possibly chained) pipeline (F9): known
// filters (including the join:<sep> prefix form) pass; an unknown filter
// name anywhere in the chain is a validate-time error naming the offending
// filter.
func TestResolveCheckFilterNameValidation(t *testing.T) {
	specKeys := map[string]bool{"repository": true}

	cases := []struct {
		name      string
		template  string
		wantError bool
	}{
		{"single known filter", "/repos/{{ config.repository | urlencode }}", false},
		{"chained known filters", "/repos/{{ config.repository | urlencode | base64 }}", false},
		{"join filter prefix form", "{{ record.tags | join:, }}", false},
		{"unknown single filter", "/repos/{{ config.repository | bogus }}", true},
		{"unknown filter in a chain", "/repos/{{ config.repository | urlencode | bogus }}", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ResolveCheck(tc.template, specKeys)
			if tc.wantError && err == nil {
				t.Fatalf("ResolveCheck(%q): expected error for unknown filter, got nil", tc.template)
			}
			if !tc.wantError && err != nil {
				t.Fatalf("ResolveCheck(%q): unexpected error: %v", tc.template, err)
			}
		})
	}
}

// --- F9 (REVIEW.md flag): ResolveCheck validated only Token/Value/When for
// an AuthSpec (cmd/connectorgen/validate.go's checkInterpolations), leaving
// username/password/token_url/client_id/client_secret/scopes typos to pass
// static validation and fail at runtime. ResolveCheckAuthSpec closes that gap
// at the engine layer (connectorgen wiring is out of this task's editable
// scope — cmd/connectorgen is not in the allowed file list for this pass). ---

func TestResolveCheckAuthFieldsValidatesAllTemplatedFields(t *testing.T) {
	specKeys := map[string]bool{"user": true, "base_url": true}

	t.Run("basic username+password: known keys pass", func(t *testing.T) {
		spec := AuthSpec{Mode: "basic", Username: "{{ config.user }}", Password: "{{ secrets.pass }}"}
		if err := ResolveCheckAuthSpec(spec, specKeys); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("basic username: unknown spec key is caught", func(t *testing.T) {
		spec := AuthSpec{Mode: "basic", Username: "{{ config.usr_typo }}", Password: "{{ secrets.pass }}"}
		err := ResolveCheckAuthSpec(spec, specKeys)
		if err == nil {
			t.Fatalf("expected validation finding for unknown spec key in username template")
		}
		if !strings.Contains(err.Error(), "usr_typo") {
			t.Fatalf("error %q does not name the offending key", err.Error())
		}
	})

	t.Run("oauth2_client_credentials: token_url/client_id/client_secret/scopes all checked", func(t *testing.T) {
		bad := AuthSpec{
			Mode:         "oauth2_client_credentials",
			TokenURL:     "{{ config.token_url_typo }}",
			ClientID:     "{{ config.user }}",
			ClientSecret: "{{ secrets.client_secret }}",
			Scopes:       "{{ config.scopes_typo }}",
		}
		err := ResolveCheckAuthSpec(bad, specKeys)
		if err == nil {
			t.Fatalf("expected validation finding for unknown spec key in token_url/scopes templates")
		}
	})

	t.Run("api_key_header: header value template checked", func(t *testing.T) {
		spec := AuthSpec{Mode: "api_key_header", Header: "X-API-Key", Value: "{{ config.value_typo }}"}
		err := ResolveCheckAuthSpec(spec, specKeys)
		if err == nil {
			t.Fatalf("expected validation finding for unknown spec key in api_key_header value template")
		}
	})

	t.Run("when condition still checked", func(t *testing.T) {
		spec := AuthSpec{Mode: "none", When: "{{ config.when_typo == 'x' }}"}
		err := ResolveCheckAuthSpec(spec, specKeys)
		if err == nil {
			t.Fatalf("expected validation finding for unknown spec key in when template")
		}
	})
}
