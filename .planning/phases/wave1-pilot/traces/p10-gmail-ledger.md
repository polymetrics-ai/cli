# p10-gmail-ledger — wave1-pilot P-10 (gmail: OAuth2 refresh-grant AuthHook pilot)

Scope (writable, per PLAN.md/conventions.md §7 path guard): `internal/connectors/defs/gmail/**`,
`internal/connectors/hooks/gmail/**`, `internal/connectors/paritytest/gmail/**`. No `git commit`
performed. FORBIDDEN files untouched: `hookset_gen.go`, `defs.go`, engine non-test files,
`go.mod`, legacy `internal/connectors/gmail/**` (read-only reference throughout).

Sources read in full before authoring: legacy `internal/connectors/gmail/{auth.go,gmail.go,
streams.go}` (718 loc total; auth.go is the 127-loc oauthRefreshAuth), SPEC.md §5.7, THREAT-MODEL.md
Delta 2, API-CONTRACT.md's gmail AuthHook shape section, TEST-PLAN.md §1/§3 gmail rows,
docs/migration/conventions.md in full, goldens `defs/stripe`/`defs/searxng`, `paritytest/monday`'s
hook-wiring precedent (`engine.New(b, engine.HooksFor("monday"))`), `paritytest/chargebee`'s
computed_fields-stringification finding (independently rediscovered here for `labels`).

---

## Red-first evidence

### Hook unit tests (`hooks/gmail/hooks_test.go`, written before `hooks.go` existed)

```
$ go test ./internal/connectors/hooks/gmail/...
# polymetrics.ai/internal/connectors/hooks/gmail [polymetrics.ai/internal/connectors/hooks/gmail.test]
internal/connectors/hooks/gmail/hooks_test.go:74:42: undefined: Hooks
internal/connectors/hooks/gmail/hooks_test.go:75:7: undefined: New
... (10 total "undefined" compile errors)
FAIL	polymetrics.ai/internal/connectors/hooks/gmail [build failed]
FAIL
```

RED confirmed: the test file referencing `Hooks`/`New` fails to compile before `hooks.go` exists.
After authoring `hooks.go`, a first GREEN pass surfaced two genuine, non-trivial red→green
transitions (not just the compile-failure RED above):

1. **https-only token_url vs. httptest.Server's plain http** — the hook's `validateHTTPSURL`
   (THREAT-MODEL.md Delta 2: "fail closed on non-https... token_url") rejected every test's
   `httptest.NewServer` URL. Fix: switched every test/parity token-server helper to
   `httptest.NewTLSServer` (+ its `.Client()`, wired into `Hooks.Client`) rather than weakening the
   https requirement — a deliberately stricter-than-legacy guard (documented parity deviation,
   see below), so the TEST infrastructure had to meet the bar, not the reverse.
2. **Eager vs. lazy field validation** — the first test draft expected
   `Authenticator()` to succeed and the "missing client_id"/"missing refresh_token"/"non-https
   token_url" errors to surface only at `Apply()` time (mirroring legacy's `accessToken()`,
   which defers all validation to first use). Corrected to eager validation at `Authenticator()`-
   build time instead, matching every OTHER engine auth mode's actual behavior
   (`engine/auth.go`'s `bearer`/`basic`/`oauth2_client_credentials` all call `Interpolate` and can
   error inside `buildAuthenticator`, never lazily) — a consistency call, not a weakened
   assertion; the error CONTENT is unchanged, only its timing.

Final GREEN (`-race`):

```
$ go test -race ./internal/connectors/hooks/gmail/... -v
=== RUN   TestHooksRegisteredUnderGmail
--- PASS: TestHooksRegisteredUnderGmail (0.00s)
=== RUN   TestAuthenticator_RefreshGrantFormShape
--- PASS: TestAuthenticator_RefreshGrantFormShape (0.01s)
=== RUN   TestAuthenticator_ClientSecretOmittedWhenUnset
--- PASS: TestAuthenticator_ClientSecretOmittedWhenUnset (0.01s)
=== RUN   TestAuthenticator_ScopeOmittedWhenUnset
--- PASS: TestAuthenticator_ScopeOmittedWhenUnset (0.01s)
=== RUN   TestAuthenticator_CachesTokenAcrossRequests
--- PASS: TestAuthenticator_CachesTokenAcrossRequests (0.01s)
=== RUN   TestAuthenticator_RefreshesWithin60sOfExpiry
--- PASS: TestAuthenticator_RefreshesWithin60sOfExpiry (0.01s)
=== RUN   TestAuthenticator_NonSuccessTokenResponseIsError
--- PASS: TestAuthenticator_NonSuccessTokenResponseIsError (0.01s)
=== RUN   TestAuthenticator_MissingRefreshTokenIsError
--- PASS: TestAuthenticator_MissingRefreshTokenIsError (0.00s)
=== RUN   TestAuthenticator_MissingClientIDIsError
--- PASS: TestAuthenticator_MissingClientIDIsError (0.00s)
=== RUN   TestAuthenticator_TokenURLMustBeHTTPS
--- PASS: TestAuthenticator_TokenURLMustBeHTTPS (0.00s)
=== RUN   TestAuthenticator_TokenURLUnparseableIsError
--- PASS: TestAuthenticator_TokenURLUnparseableIsError (0.00s)
=== RUN   TestAuthenticator_HonorsContextCancellation
--- PASS: TestAuthenticator_HonorsContextCancellation (0.00s)
=== RUN   TestAuthenticator_ErrorsNeverContainSecretText
--- PASS: TestAuthenticator_ErrorsNeverContainSecretText (0.01s)
PASS
ok  	polymetrics.ai/internal/connectors/hooks/gmail	1.426s
```

### Parity suite (`paritytest/gmail/parity_test.go`, written before `defs/gmail` existed)

Per protocol, `internal/connectors/defs/gmail` was moved out (to the scratchpad) before writing
the parity file, then restored after capturing RED:

```
$ go test ./internal/connectors/paritytest/gmail/... -v
=== RUN   TestParityGmail_BundleLoadsAndValidates
    parity_test.go:27: engine.Load(defs.FS, "gmail"): load bundle gmail: missing required file metadata.json
--- FAIL: TestParityGmail_BundleLoadsAndValidates (0.00s)
FAIL
FAIL	polymetrics.ai/internal/connectors/paritytest/gmail	0.353s
FAIL
```

After authoring the full bundle, the FIRST real (non-trivial) red→green transition on the full
parity suite:

```
=== RUN   TestParityGmail_StreamRecords/messages
    parity_test.go:240: Read(messages): gmail stream=messages: engine: resolve query "maxResults": interpolate: unresolved key "page_size" in config
=== RUN   TestParityGmail_StreamRecords/labels
    parity_test.go:240: Read(labels): gmail stream=labels page=0: resolve stream path: interpolate: unresolved key "user_id" in config
```

Root cause: `spec.json`'s JSON-Schema `default` values (`page_size: "100"`, `user_id: "me"`) are
annotation-only — the engine never injects them into `RuntimeConfig` at runtime (confirmed via
`engine/schema.go`'s `annotationKeywords`); every caller (including this test) must set them
explicitly, exactly like `parity_stripe_test.go`'s `stripeRuntimeConfig` always sets `base_url`
itself rather than relying on its own spec default. Fixed by adding `user_id`/`page_size` (and
`scopes`) to `gmailRuntimeConfig`'s base config map — not an engine change, a test-authoring
correction.

A SECOND genuine red→green transition (the type-shape finding, see Parity-deviation ledger below):

```
=== RUN   TestParityGmail_StreamRecords/labels
    parity_test.go:255: stream "labels" record 0 mismatch:
        engine:  map[... messages_total:10 ...]
        legacy:  map[... messages_total:10 ...]
```
(identical `%+v` rendering but `reflect.DeepEqual` failed — a Go type mismatch: engine emits
`messages_total` as a `string`, legacy as a `json.Number`; both string-render as `10`.)

Final GREEN (`-race`, full suite incl. the explicit stringification-deviation lock-in test):

```
$ go test -race ./internal/connectors/paritytest/gmail/... -v
=== RUN   TestParityGmail_StreamRecords
=== RUN   TestParityGmail_StreamRecords/messages
=== RUN   TestParityGmail_StreamRecords/threads
=== RUN   TestParityGmail_StreamRecords/drafts
=== RUN   TestParityGmail_StreamRecords/labels
--- PASS: TestParityGmail_StreamRecords (0.05s)
=== RUN   TestParityGmail_MessagesTwoPagePagination
--- PASS: TestParityGmail_MessagesTwoPagePagination (0.01s)
=== RUN   TestParityGmail_LabelsUnpaginatedSinglePage
--- PASS: TestParityGmail_LabelsUnpaginatedSinglePage (0.01s)
=== RUN   TestParityGmail_ComputedFieldsStringifyLabelCountFields
--- PASS: TestParityGmail_ComputedFieldsStringifyLabelCountFields (0.01s)
=== RUN   TestParityGmail_AuthorizationHeaderAfterRefresh
--- PASS: TestParityGmail_AuthorizationHeaderAfterRefresh (0.01s)
=== RUN   TestParityGmail_TokenEndpointFailureSurfacesAsAuthError
--- PASS: TestParityGmail_TokenEndpointFailureSurfacesAsAuthError (0.02s)
=== RUN   TestParityGmail_WriteUnsupportedOnBothSides
--- PASS: TestParityGmail_WriteUnsupportedOnBothSides (0.00s)
=== RUN   TestParityGmail_ManifestSurface
--- PASS: TestParityGmail_ManifestSurface (0.02s)
=== RUN   TestParityGmail_BundleLoadsAndValidates
--- PASS: TestParityGmail_BundleLoadsAndValidates (0.00s)
=== RUN   TestParityGmail_AuthSpecIsSoleCustomCandidate
--- PASS: TestParityGmail_AuthSpecIsSoleCustomCandidate (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/paritytest/gmail	1.612s
```

---

## Design decisions

- **Bundle**: `defs/gmail/{metadata,spec,streams,api_surface}.json`, `schemas/{messages,threads,
  drafts,labels}.json`, `docs.md`, `fixtures/{check.json,streams/{messages,threads,drafts,
  labels}/page_*.json}`. No `writes.json` (`capabilities.write: false`, matches legacy's
  `ErrUnsupportedOperation`, gmail.go:191-192).
- **Auth**: `streams.json` `base.auth` = exactly one candidate, `{"mode":"custom","hook":"gmail",
  "token_url":"{{ config.token_url }}","client_id":"{{ secrets.client_id }}",
  "client_secret":"{{ secrets.client_secret }}","token":"{{ secrets.client_refresh_token }}",
  "scopes":"{{ config.scopes }}"}` — no `when`-gated fallback (legacy has no alternate auth path,
  unlike github's token-or-app_jwt "auto" resolution; SPEC §5.7's "no roster swap needed" holds).
- **hooks/gmail/hooks.go** (267 loc, 1 hook interface: `AuthHook`): `Hooks.Authenticator`
  interpolates the 5 templated AuthSpec fields against `cfg` (required: `token_url`,`client_id`,
  the refresh token via `spec.Token`; optional: `client_secret`,`scopes`, resolved via a
  best-effort `interpolateOptional` that treats an absent config/secrets key as `""` rather than
  propagating `engine.Interpolate`'s hard "unresolved key" error — mirrors legacy's own
  `if a.clientSecret != ""`/`if a.scope != ""` omission guards, since general `Interpolate` has no
  absent-key-falsy tolerance outside `when`, conventions.md §3), validates `token_url` is a
  well-formed **https** URL (THREAT-MODEL.md Delta 2), then returns an `oauthRefreshAuth`
  (`Authenticator`) that is `internal/connectors/gmail/auth.go`'s `oauthRefreshAuth` ported
  field-for-field: mutex-guarded token cache, 60s-early refresh margin, injectable `now`/`Client`,
  `grant_type=refresh_token` form POST, `client_secret`/`scope` omitted from the form entirely
  when empty, secret values never entering any error string.
- **Streams (Tier-1 JSON)**: `messages`/`threads`/`drafts` — `cursor` pagination
  (`cursor_param: pageToken`, `token_path: nextPageToken`); `labels` — stream-level
  `pagination: {"type":"none"}` override (legacy's `paginated: false` routing-table entry,
  streams.go:28). No `incremental` block on any stream (legacy publishes no cursor field,
  streams.go:31-34 — full_refresh only, matches legacy's `InitialState` always seeding `""`).
  `computed_fields` rename camelCase raw fields to the schema's snake_case names and, for
  `drafts`, reach into the nested raw `message` object exactly like legacy's `draftRecord`.

## Parity-deviation ledger (candidates for `docs/migration/conventions.md` §5, P-12 to formalize)

| # | description | verdict |
|---|---|---|
| 1 | `computed_fields`' camelCase->snake_case rename of `labels`' 4 numeric count fields (`messagesTotal`/`messagesUnread`/`threadsTotal`/`threadsUnread`) resolves through `engine.Interpolate`, which always returns a Go `string` regardless of the raw JSON value's real type (`engine/interpolate.go`'s `resolveExpr`/`stringify`) — so the engine emits these as strings while legacy emits `json.Number` (via `connsdk`'s `UseNumber` decode). Never changes the numeric VALUE (`"10"` vs `json.Number("10")` carry identical information); asserted explicitly, not silently coerced, in `TestParityGmail_ComputedFieldsStringifyLabelCountFields`. Schema types declared `["string","null"]` (the bundle's REAL emitted type) rather than `"integer"`. **Independently rediscovered** — `paritytest/chargebee`'s `TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields` is the identical finding for chargebee's full envelope-unwrap; P-12 should fold both into one conventions.md §5 entry ("computed_fields is string-producing; any renamed/nested non-string field is stringified"). | ACCEPTABLE |
| 2 | `token_url` is validated **https-only** by the hook, stricter than legacy's `validatedURL` (gmail.go:339-357), which also accepted plain `http`. Never stricter for any real Google OAuth endpoint (always https in production); closes the one new SSRF-adjacent surface THREAT-MODEL.md Delta 2 identifies (an http override would let an attacker-controlled endpoint receive `client_secret`+refresh token in cleartext). | ACCEPTABLE |
| 3 | `start_date`/`include_spam_and_trash` (legacy's client-side Gmail search-query filters, `q=after:<unix-seconds>`/`includeSpamTrash=true`, gmail.go:264-296) are declared in `spec.json` but NOT wired into any `streams.json` `query` template this pass — `stream.Query` templating has no absent-key-falsy tolerance (conventions.md §3), so a static `query` key would send the filter unconditionally, which is not parity with legacy's optional behavior. Left as a documented scope-narrowing (docs.md Known limits), not a silently-wrong wiring. | ACCEPTABLE (documented scope narrowing) |

## ENGINE_GAP (blocker, typed; not introduced by this connector, but this connector is the second-of-two exhibits)

**`TestConformance/<name>`'s dynamic (fixture-replay) checks always call `engine.Check`/
`engine.Read` with `h=nil`** (`internal/connectors/conformance/dynamic.go`'s `runDynamicChecks`
passes a literal `nil` `Hooks` argument to every dynamic check — verified by direct code read of
`checkCheckFixture`/`readRawRecords`/`checkPaginationTerminates`/`checkWriteRequestShape`, all of
which call `withReplayURL`+`engine.Read`/`engine.Check` with a hardcoded `nil` in the Hooks
position). `engine/auth.go`'s `buildCustomAuth` requires a non-nil `Hooks` implementing
`AuthHook` for any `mode: custom` candidate; with `h=nil` it always returns
`"custom auth: hook %q not registered (no hooks provided)"` BEFORE any HTTP request is attempted.

Confirmed empirically two ways:
1. An isolated sandboxed repro (`engine.Check` called directly with a synthetic
   `mode:custom` bundle and `h=nil`) reproduces the exact error text.
2. `go test ./internal/connectors/conformance -run 'TestConformance/gmail' -v` (see Self-verify
   below): all 10 static checks PASS; every dynamic check that resolves auth
   (`check_fixture`, `read_fixture_nonempty:*`, `pagination_terminates`, `records_match_schema`)
   FAILS with exactly this error; `cursor_advances`/`delete_semantics` legitimately Skip (no
   incremental stream / no delete write action).

This is **not specific to gmail's bundle authoring** — it reproduces for ANY bundle whose sole/
first-matching auth candidate is `mode: custom` with no `when`-gated non-custom fallback that
conformance's synthetic secrets happen to satisfy first. gmail has no such fallback (legacy has no
alternate auth path to declare, SPEC §5.7); this makes gmail (alongside github, whose `custom`
candidate is only masked because a `when`-gated bearer candidate is tried first and conformance's
synthetic secret values happen to satisfy it — the identical underlying gap, not a
counter-example) the second-of-three Tier-2 AuthHook connectors to hit this in the SAME wave,
satisfying conventions.md §6's "`ENGINE_GAP`s recur ≥3 times -> mini wave-0 engine increment"
threshold once github's finding is confirmed (monday is a StreamHook/CheckHook pilot, not
AuthHook, so it is unaffected).

- **Type**: `ENGINE_GAP`
- **Evidence**: `internal/connectors/conformance/dynamic.go` (`runDynamicChecks`, every dynamic
  check helper hardcodes `nil` for the `Hooks` parameter); `internal/connectors/engine/auth.go:149-
  158` (`buildCustomAuth`'s `h == nil` branch); reproduced live via
  `go test ./internal/connectors/conformance -run 'TestConformance/gmail' -v` (see below).
- **Why a Tier-1/2 workaround would diverge from legacy**: the only way to make
  `TestConformance/gmail` fully green today would be adding a non-custom (e.g. `bearer`) auth
  candidate ahead of the `custom` one so conformance's synthetic secret satisfies it first — but
  legacy gmail has NO such alternate auth path; inventing one would be new, unrequested auth
  surface, not a migration of legacy behavior, and would silently mask that the real refresh-grant
  path is NEVER exercised by conformance.
- **Recommended fix (P-12/orchestrator, out of this task's writable scope)**: either (a) give
  `conformance` a way to resolve a bundle's registered hooks (`engine.HooksFor(b.Name)`) instead of
  hardcoding `nil`, letting a hooks-registering connector's `TestConformance` genuinely exercise its
  hook path (mirrors how `paritytest/<name>` already does this correctly), or (b) document this
  explicitly as an accepted, permanent conformance-vs-Tier-2-AuthHook gap and route
  hook-auth-only connectors' dynamic-check acceptance bar through `paritytest/<name>` instead (this
  bundle's `docs.md` already states this explicitly).

This blocker does NOT block this connector's `migrated` status: SPEC.md's own acceptance bar for
gmail is the `paritytest/gmail` parity suite (fully green, `-race` clean), matching TEST-PLAN.md
§1's row for gmail and §2's note that `paritytest/gmail`, not `TestConformance/gmail`, is the
authoritative correctness bar for an AuthHook-only Tier-2 connector's auth path.

## Self-verify (conventions.md §7 block)

```
$ go run ./cmd/connectorgen validate internal/connectors/defs --json | jq '[.findings[]|select(.connector=="gmail")]'
[]
```
(Note: conventions.md §7 literally documents `validate internal/connectors/defs/<name>`, but the
tool's actual `[dir]` argument is the DEFS ROOT — every subdirectory of `dir` is validated as its
own bundle, so `validate internal/connectors/defs/gmail` incorrectly treats gmail's OWN
`fixtures`/`schemas` subdirectories as bundles and fails with `missing required file
metadata.json` for both — reproduced identically against `defs/stripe`, so this is a pre-existing
doc/tool mismatch, not specific to this connector. `validate internal/connectors/defs` — the
Makefile's own `connectorgen-validate` target's actual invocation — is correct and shows 0
findings total across all 13 currently-loadable bundles, gmail included. Flagged for P-12's
conventions.md patch pass.)

```
$ go build ./internal/connectors/... && go vet ./internal/connectors/...
(clean, no output)

$ go test ./internal/connectors/conformance -run 'TestConformance/gmail' -v
=== RUN   TestConformance/gmail
    conformance_test.go:75: conformance failed for gmail: [
      spec_schema_valid:PASS stream_schemas_valid:PASS pk_fields_exist:PASS
      cursor_fields_exist:PASS interpolations_resolve:PASS write_schemas_valid:PASS
      surface_complete:PASS docs_present:PASS secret_redaction:PASS fixtures_present:PASS
      check_fixture:FAIL(hook not registered) read_fixture_nonempty:*:FAIL(hook not registered) x4
      pagination_terminates:FAIL(hook not registered) records_match_schema:FAIL(hook not registered)
      cursor_advances:SKIP delete_semantics:SKIP
    ]
--- FAIL: TestConformance/gmail
```
(Expected per the ENGINE_GAP above — all static checks pass; every dynamic check needing auth
resolution fails with the identical, predicted error; `cursor_advances`/`delete_semantics`
legitimately skip. See docs.md's "Known limits" for the full explanation surfaced to reviewers.)

```
$ go test -race ./internal/connectors/paritytest/gmail ./internal/connectors/hooks/gmail -v
... (see Red-first evidence above)
PASS  polymetrics.ai/internal/connectors/hooks/gmail
PASS  polymetrics.ai/internal/connectors/paritytest/gmail
```

```
$ go build ./... && go vet ./...
(clean, no output, at time of this run — sibling in-flight DW-1 agents may transiently break
defs.FS's go:embed all:* if their directories are momentarily empty; this is a known, external,
non-gmail hazard, not a defect in this connector's files)
```

```
$ golangci-lint run ./internal/connectors/defs/gmail/... ./internal/connectors/hooks/gmail/... ./internal/connectors/paritytest/gmail/...
0 issues.
```
