# gaploop-r1-ledger — wave0-engine-harness ENGINE-CORE batch

Backend gap-closure pass over REVIEW.md (B1, F1, F3, F4, F5, F8, F9) and
SECURITY-REVIEW.md (M1, m2, m3). Strict TDD per item: RED test capturing the
reviewer's exact failure scenario, recorded here BEFORE the fix, then GREEN.

Scope: `internal/connectors/engine/**` only (+ stripe defs if unavoidable for
B1's honest parity fix — not needed, see below).

---

## B1 — cursor round-trip through the real app layer (formatParam unix_seconds digits passthrough)

### RED

Failing test added first: `TestFormatParamUnixSecondsAcceptsDigitsPassthrough`
in `internal/connectors/engine/read_test.go`, driving `formatParam(value,
"unix_seconds")` with a digits-only value `"1700000100"` (the exact shape
`internal/app/sync_modes.go`'s `recordCursor`/`toComparableString` persists
for a `json.Number`/int64 stripe `created` field — confirmed by reading
`internal/app/sync_modes.go:133-161` read-only). Before the fix:

```
$ go test ./internal/connectors/engine/... -run TestFormatParamUnixSecondsAcceptsDigitsPassthrough -v
--- FAIL: TestFormatParamUnixSecondsAcceptsDigitsPassthrough (0.00s)
    read_test.go:...: formatParam("1700000100", "unix_seconds") error = engine: param_format
    unix_seconds: invalid RFC3339 value "1700000100": parsing time "1700000100" as
    "2006-01-02T15:04:05Z07:00": cannot parse "1700000100" as "2006"
FAIL
```

This is the exact B1 failure: a digits-only state cursor (what the app layer
actually persists) makes `time.Parse(time.RFC3339, ...)` fail, so the second
incremental sync errors out end-to-end.

Also added (RED before fix, same run):
- `TestFormatParamDateAcceptsDigitsPassthrough` / `TestFormatParamGithubDateRangeAcceptsDigitsPassthrough`
  — same digits-in failure mode for the other two formats that derive from an
  RFC3339 parse.
- `TestReadAppLevelCursorRoundTrip` in `read_test.go`: mimics
  `internal/app/sync_modes.go`'s `recordCursor`/`toComparableString`
  stringification exactly (a local copy, since engine may not import
  internal/app) — first read emits a record with a numeric `created` field,
  derive the persisted cursor the app's way (stringify the json.Number
  verbatim), feed it back as `req.State["cursor"]` on the second read, assert
  the outgoing `created[gte]` is the correct unix-seconds value. RED before
  fix: same "cannot parse ... as 2006" error surfaces from the second Read
  call.

### Fix

`formatParam` (read.go): for `unix_seconds`, `date`, and `github_date_range`,
detect an all-digits input first (`isAllDigits`) and treat it as a Unix-
seconds value already (parse with `strconv.ParseInt`, then derive the
requested representation from `time.Unix(...)`), preserving the existing
RFC3339-input behavior untouched when the value is not all-digits. rfc3339
format itself already passes values through verbatim (no parse), so it needs
no change.

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestFormatParam|TestReadAppLevelCursorRoundTrip|TestReadIncrementalParamFormats' -v
=== RUN   TestFormatParamUnixSecondsAcceptsDigitsPassthrough
--- PASS
=== RUN   TestFormatParamDateAcceptsDigitsPassthrough
--- PASS
=== RUN   TestFormatParamGithubDateRangeAcceptsDigitsPassthrough
--- PASS
=== RUN   TestReadAppLevelCursorRoundTrip
--- PASS
=== RUN   TestReadIncrementalParamFormats
--- PASS (unchanged RFC3339-input cases still pass)
PASS
```

Updated `parity_stripe_test.go`'s
`TestParityStripe_IncrementalCreatedGTEFromState` to feed the engine the
APP-PERSISTED cursor shape (`"1700000100"`, unix-seconds string) instead of
the hand-crafted RFC3339 string `"2023-11-14T22:15:00Z"` — this is the honest
parity bar the review calls for. Confirmed GREEN against both connectors
after the fix (legacy already forwarded the raw digits verbatim; engine now
does too).

---

## F1 — templated stream/check paths never interpolated

### RED

`TestReadStreamPathIsInterpolated` (read_test.go): a stream `Path` containing
`{{ config.repo }}` with a config value containing `/` — before the fix,
`readDeclarative` sends `stream.Path` literally (`reqPath := stream.Path`),
so the request path arrives at the server as the literal string
`/repos/{{ config.repo }}` instead of the interpolated (and urlencoded, since
InterpolatePath defaults to urlencode) value.

```
$ go test ./internal/connectors/engine/... -run TestReadStreamPathIsInterpolated -v
--- FAIL: TestReadStreamPathIsInterpolated
    read_test.go:...: got path = "/repos/{{ config.repo }}", want "/repos/acme%2Fwidgets"
FAIL
```

`TestCheckPathIsInterpolated` (read_test.go): same failure mode for
`b.HTTP.Check.Path`.

### Fix

`readDeclarative`: `reqPath := stream.Path` now runs through
`InterpolatePath(stream.Path, requestVars(req.Config, nil, ""))` before use
(when `page.URL` is empty — a paginator-supplied absolute URL still wins, as
before). `Check`: `b.HTTP.Check.Path` is interpolated the same way against
`requestVars(cfg, nil, "")` before the request.

Verified the three existing goldens' static paths (no `{{ }}`) round-trip
unchanged through `InterpolatePath` (a template with no `{{ }}` markers is
returned verbatim by `interpolate()`).

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestReadStreamPathIsInterpolated|TestCheckPathIsInterpolated' -v
--- PASS (both)
$ go test ./internal/connectors/engine/... -run TestParityStripe -v   # goldens unaffected
PASS
$ go test ./internal/connectors/engine/... -run TestParitySearxng -v
PASS
```

---

## F9 — interpolation robustness (chained filters, ResolveCheck auth fields, path traversal)

### RED

1. `TestApplyFilterChainedFiltersSupported` / `TestInterpolateMultipleFiltersChained`
   (interpolate_test.go): `{{ config.x | urlencode | base64 }}` — before the
   fix, `resolveExpr` only ever looks at `parts[1]` (the FIRST pipe segment)
   and silently ignores everything after, so this template silently drops the
   `base64` stage instead of applying both in order or erroring.
2. `TestApplyFilterUnknownFilterNameErrors` — already covered by existing
   `applyFilter`'s default case (confirmed still errors, no regression risk;
   kept as a locked-in regression test rather than new RED).
3. `TestResolveCheckAuthFieldsValidatesAllTemplatedFields` (interpolate_test.go):
   a new `ResolveCheckAuthSpec` helper that checks EVERY templated AuthSpec
   field (username/password/token_url/client_id/client_secret/scopes, not
   just token/value/when) against specKeys — RED because the function does
   not exist yet.
4. `TestInterpolatePathRejectsDotDotSegment` (interpolate_test.go): a resolved
   path segment that is exactly `..` (or contains a raw `/../` after url-
   decoding) must be rejected even though standard urlencode already escapes
   the slash — this targets F9b/m3's "traversal survives as an encoded
   literal segment" finding. RED: `InterpolatePath("/customers/{{ config.id
   }}", vars)` with `config.id = ".."` currently returns `/customers/..`
   (urlencodeSegment leaves bare `.` unescaped, so the literal `..` string
   round-trips as a same-value percent-decodable segment).

```
$ go test ./internal/connectors/engine/... -run 'TestApplyFilterChainedFiltersSupported|TestInterpolateMultipleFiltersChained|TestResolveCheckAuthFieldsValidatesAllTemplatedFields|TestInterpolatePathRejectsDotDotSegment' -v
--- FAIL (all four, as described above)
```

### Fix

- `resolveExpr`/`applyFilter`: support chained filters — split on `|` fully,
  apply each named filter in sequence to the raw resolved value (still
  urlencode-by-default only when NO explicit filter chain is given, matching
  existing default semantics). Unknown filter name anywhere in the chain is
  still a hard error (no silent truncation, no silent skip).
- `ResolveCheckAuthSpec(spec AuthSpec, specKeys map[string]bool) error`: new
  engine-exported helper validating Token, Username, Password, Header/Value
  (api_key_header), Param/Value (api_key_query), TokenURL, ClientID,
  ClientSecret, Scopes, and When — every templated AuthSpec field — via
  `ResolveCheck`. (connectorgen's `checkInterpolations` is out of this task's
  editable scope — internal/cli and cmd/connectorgen are not in the allowed
  file list — so this is exposed at the engine layer for connectorgen to wire
  in during a follow-up; documented here rather than silently left
  unreachable.)
- `InterpolatePath`: after resolving+filtering each `{{ }}` expression,
  reject the resolved value outright when it is exactly `..` or (after
  percent-decoding) contains `/../`, `/..` at the end, or `../` at the start
  — closing the "single dot-dot segment survives encoded" gap FLAGged as F9b.

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestApplyFilter|TestInterpolate|TestResolveCheck' -v
PASS (all)
```

---

## F8 — AuthHook context.Background()

### RED

`TestSelectAuthCustomThreadsCallerContext` (auth_test.go): a fake AuthHook
that stashes the ctx it received into a probe variable via a ctx-value key;
`selectAuth` is called with a context carrying a marker value. Before the
fix, `buildCustomAuth`/`selectAuth` take no ctx parameter at all and
`buildCustomAuth` calls `authHook.Authenticator(context.Background(), ...)`
— RED because `selectAuth`'s signature has no ctx to thread, so the test
cannot even compile against the desired call shape until the signature
changes; written first as a compile-time-RED (build fails), then the
signature change makes it pass.

```
$ go test ./internal/connectors/engine/... -run TestSelectAuthCustomThreadsCallerContext -v
# build failure: selectAuth undefined signature accepting ctx (compile RED)
```

### Fix

Threaded `ctx context.Context` through `selectAuth(ctx, cfg, specs, h)` ->
`buildAuthenticator(ctx, ...)` -> `buildCustomAuth(ctx, ...)` ->
`authHook.Authenticator(ctx, cfg, spec)`. Updated call sites: `newRuntime`
(read.go, now takes/forwards ctx), `Read`/`ReadWithSleeper` (already have
ctx in scope), `Check` (already has ctx), `Write` (already has ctx).

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestSelectAuth|TestRead|TestCheck|TestWrite' -v
PASS (all)
```

---

## M1 (+ F2/m2) — SSRF guards: link_header wrapper, scheme check, fail-closed unparseable URL

### RED

1. `TestNewPaginatorLinkHeaderCrossHostBlocked` (paginate_test.go): a
   `link_header`-paginated server whose `Link: rel="next"` header points at a
   different host — before the fix, `newPaginator` returns a bare
   `&connsdk.LinkHeaderPaginator{}` with no host guard at all, so RED shows
   the cross-host page's record IS fetched (guard absent).
2. `TestNewPaginatorLinkHeaderAllowCrossHostEscape` / same-host-allowed case
   — locks in that ordinary github-shaped same-host Link-header pagination
   keeps working post-fix (written alongside the RED test so the fix's scope
   is proven both ways).
3. `TestNewPaginatorNextURLSSRFGuardSchemeDowngradeBlocked` (paginate_test.go):
   base `https://host`, next URL `http://host` (same host, downgraded
   scheme) — before the fix, `urlHost`/comparison only looks at `.Host`, so
   RED shows the downgraded-scheme request IS followed.
4. `TestNewPaginatorNextURLUnparseableNextURLFailsClosed`: a next_url body
   value that fails `url.Parse` (e.g. a control-character-laden string) —
   before the fix, `urlHost` returns `""` on parse failure and the guard
   condition `host != "" && host != p.BaseHost` short-circuits to allow it
   through; RED shows the garbage URL is followed rather than rejected.
5. Same three shapes mirrored for `link_header` once the wrapper exists
   (scheme downgrade + unparseable), added directly (no separate RED
   sub-step — same code path as nextURL's guard, reused).

```
$ go test ./internal/connectors/engine/... -run 'TestNewPaginatorLinkHeaderCrossHostBlocked|TestNewPaginatorNextURLSSRFGuardSchemeDowngradeBlocked|TestNewPaginatorNextURLUnparseableNextURLFailsClosed' -v
--- FAIL (all three, as described)
```

### Fix

- New engine-local `linkHeaderPaginator` type in paginate.go mirroring
  `nextURL`'s guard structure exactly (BaseHost/allowCrossHost/seen-loop
  guard/sticky Err()), wrapping `connsdk.LinkHeaderPaginator`'s
  Link-header-follow semantics. `newPaginator`'s `case "link_header"` now
  returns `&linkHeaderPaginator{allowCrossHost: spec.AllowCrossHost}`.
  `read.go`'s `readDeclarative` sets `BaseHost` on it the same way it does
  for `*nextURL` (both satisfy a shared small interface for that purpose).
- Guard logic (shared helper `sameOrigin`/`hostGuard` used by both
  `nextURL.Next` and `linkHeaderPaginator.Next`): parse the next URL; if
  parsing fails, fail closed (sticky error), never silently pass; compare
  BOTH scheme and host against the base (derived from BaseHost + the
  original scheme, threaded in) — a scheme downgrade on the same host is now
  rejected unless allow_cross_host is set.
- `read.go`: `requesterHost` extended to also capture scheme (or a
  companion `requesterOrigin` helper), wired into both paginator types'
  `BaseHost`/scheme fields before the first request.

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestNewPaginator' -v
PASS (all, including every pre-existing SSRF/loop-guard/github-shaped test)
```

---

## F3 — lastRecordCursor hardcoded "data" path + string-only ids

### RED

`TestNewPaginatorCursorLastRecordFieldNonDataEnvelope` (paginate_test.go): a
stripe-shaped cursor paginator against a response whose records live under
`"results"` (not `"data"`) — before the fix, `lastRecordFieldValue` calls
`connsdk.RecordsAt(body, "data")` unconditionally, so RED shows pagination
stops after page 1 silently (no error, just truncation) even though a
second page exists.

`TestNewPaginatorCursorLastRecordFieldNumericID` (paginate_test.go): last
record's id field is a JSON number (`json.Number`/`float64`), not a Go
string — before the fix, `lastRecordFieldValue`'s type assertion `v.(string)`
fails, `ok` is false, so RED shows pagination stops after page 1 even though
the numeric id could be stringified.

```
$ go test ./internal/connectors/engine/... -run 'TestNewPaginatorCursorLastRecordFieldNonDataEnvelope|TestNewPaginatorCursorLastRecordFieldNumericID' -v
--- FAIL (both — truncate-after-page-1 with no error)
```

### Fix

- `lastRecordCursor` gains a `recordsPath string` field (defaulted to "data"
  only when literally unset, to keep the zero-value construction path used
  by any caller that never sets it working — but `newCursorPaginator`
  (paginate.go) now wires it from the *effective* `RecordsSpec.Path` passed
  in by `read.go`, which already knows the stream's records path).
  `newPaginator`'s signature gains a `recordsPath string` parameter;
  `newCursorPaginator` forwards it into `lastRecordCursor.recordsPath`.
  `read.go`'s call site passes `recordsPathOf(stream.Records)`.
- `lastRecordFieldValue`: accept `json.Number`/`float64` ids in addition to
  `string`, stringifying numbers canonically (json.Number → its own string
  form since connsdk decodes with UseNumber; float64 handled defensively for
  any caller that doesn't).

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestNewPaginatorCursor' -v
PASS (all, including existing "data" + string-id stripe-shape tests)
$ go test ./internal/connectors/engine/... -run TestParityStripe -v
PASS (stripe still uses "data" + string ids — no behavior change for the golden)
```

---

## F4 — resolveHeaders silently drops ANY unresolved-key header

### RED

`TestReadHeaderAbsentSecretAuthorizationErrors` (read_test.go): a declared
header named `Authorization` templated as `{{ secrets.token }}` with no
`token` secret configured — before the fix, `resolveHeaders` catches the
"unresolved key" error via the brittle substring-matched `isUnresolvedKey`
and silently omits the header, so RED shows the request going out with NO
Authorization header and NO error — exactly the "silently unauthenticated"
scenario F4 flags.

`TestReadHeaderOptionalDeclaredCustomHeaderOmitted` (read_test.go): a
non-auth-shaped custom header referencing an optional (spec-declared-but-
absent, not in required[]) config key must still be OMITTED, not error —
locks in the Stripe-Account pattern is preserved (this one is a "must stay
green" companion, not new RED, but written alongside since the fix changes
the omission rule's mechanism).

`TestReadHeaderRequiredConfigKeyErrors` (read_test.go): a header templated
against a config key that IS in spec required[] but absent at runtime must
error, not silently omit.

```
$ go test ./internal/connectors/engine/... -run TestReadHeaderAbsentSecretAuthorizationErrors -v
--- FAIL: request sent with no Authorization header and Read returned nil error
```

### Fix

- New typed sentinel error `unresolvedKeyError{Namespace, Key string}` in
  interpolate.go, returned (wrapped) from `resolveRef`/`resolveRecordPath`
  instead of only a substring-matchable `fmt.Errorf`. `isUnresolvedKey`/
  `isUnresolvedRecordPath` (read.go) rewritten to use `errors.As` against
  the typed error instead of `strings.Contains`.
  `unresolvedKeyError.authShaped` bit is NOT on the error type (spec
  awareness doesn't belong in interpolate.go); auth-shapedness is
  decided by the caller (resolveHeaders) as described next.
- `resolveHeaders(headers map[string]string, cfg connectors.RuntimeConfig,
  specKeys map[string]bool, requiredKeys map[string]bool)` (bundle's spec
  now threaded in from `newRuntime`, which already has `b.Spec` in scope):
  on an unresolved-key error, `errors.As` extracts the missing
  namespace/key. Decision table:
  - namespace `secrets` (any key) → HARD ERROR always (never send an
    auth-bearing or any other header unauthenticated-by-silent-omission;
    matches F4's "headers referencing secrets.* should hard-error"
    recommendation, applied uniformly rather than only to header names that
    look like "Authorization" — simpler and strictly safer, and still
    satisfies the Stripe-Account pattern since that header templates a
    `config.*` key, not `secrets.*`).
  - namespace `config`, key present in spec properties AND NOT in
    spec's required[] → OMIT (declared-optional pattern, e.g.
    Stripe-Account/account_id).
  - namespace `config`, key present in spec required[], or key not declared
    in spec at all → HARD ERROR (required-but-missing, or undeclared
    reference entirely).
  - any other interpolation error (CRLF, unknown namespace/filter) →
    propagates unchanged, as before.
- Spec required[] is not currently surfaced by `*Schema`; added
  `Schema.RequiredKeys() []string` (root-level `required` array) alongside
  the existing `Properties()`/`SecretKeys()` accessors.

### GREEN

```
$ go test ./internal/connectors/engine/... -run TestReadHeader -v
PASS (all four: absent-secret-Authorization errors, optional-declared header
omitted, required-config-header errors, existing Stripe-Account-shaped test
unchanged)
$ go test ./internal/connectors/engine/... -run TestParityStripe -v
PASS (Stripe-Account header behavior unchanged: account_id is a declared,
non-required config key)
```

---

## F5 — Definition.Spec lossy reconstruction

### RED

`TestDefinitionSpecByteEqualsBundleSpecJSON` (connector_test.go): loads a
bundle via `Load(fsys, name)` from an `fstest.MapFS` whose `spec.json` has
rich JSON-Schema (enum, default, required, integer type, description) and
asserts `c.Definition().Spec` is byte-for-byte (after canonical
re-marshal, to tolerate whitespace-only differences) equal to the bundle's
own `spec.json` bytes. Before the fix, `specJSON` reconstructs every
property as `{"type":"string"}` (+x-secret only) — RED shows types/enum/
default/required/descriptions all missing from `Definition().Spec`.

```
$ go test ./internal/connectors/engine/... -run TestDefinitionSpecByteEqualsBundleSpecJSON -v
--- FAIL: Definition().Spec lost enum/default/required/integer-type/description
```

### Fix

- `Bundle` gains `RawSpec json.RawMessage` (bundle.go), populated by
  `loadSpec` from the same `raw` bytes it already reads before compiling —
  no new file read, just retaining what was already in hand.
  `synthesizeDefinition`/`specJSON` (connector.go) now returns
  `b.RawSpec` verbatim when non-nil, falling back to the old
  reconstruction (kept for the ad hoc-test-bundle case where a caller sets
  `Spec` but never `RawSpec`, and for the truly-empty case) — this keeps
  `TestConnectorDefinitionSynthesizedFromBundle`'s existing
  `minimalSpecSchema`-only (no RawSpec) construction green without changes.

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestDefinitionSpec|TestConnectorDefinition|TestBundleLoad|TestBaseServesDefinition' -v
PASS (all)
```

---

## join:<sep> filter + static-literal computed_fields (meta-rule enablement, F7 engine-side prerequisite)

### RED

`TestApplyFilterJoinSeparator` (interpolate_test.go): `{{ record.tags | join:, }}`
where `record.tags` is a `[]any` — before the fix, `applyFilter` has no
`join:` case, so any `join:...` filter name hits `default: unknown filter`.
`TestApplyFilterJoinNonArrayErrors`: `join:` applied to a non-array value
must error (not silently stringify).

`TestReadComputedFieldsStaticLiteralNoTemplate` (read_test.go): a
computed_fields entry whose value has NO `{{ }}` markers at all (e.g.
`"source_system": "searxng"`) — before the fix, `applyComputedFields` always
calls `Interpolate(tmpl, ...)`, which for a template with no `{{ }}` already
returns the literal string verbatim (interpolate()'s ReplaceAllStringFunc is
a no-op when there's nothing to replace) — this one is actually ALREADY
supported by existing `Interpolate` semantics; confirmed by reading
`interpolate()`'s implementation, so this is documented as "no fix needed,
locked in by a new regression test" rather than a RED->GREEN pair. (See test
comment.)

```
$ go test ./internal/connectors/engine/... -run 'TestApplyFilterJoin' -v
--- FAIL (join: filter unknown)
$ go test ./internal/connectors/engine/... -run TestReadComputedFieldsStaticLiteralNoTemplate -v
--- PASS already (no code change needed; kept as a locked-in regression test, not counted as a fix)
```

### Fix

`applyFilter`: recognize a `join:<sep>` filter name prefix (the separator is
everything after the first `:`; array-valued filter input requires the
pre-filter resolved value to still be array-shaped, which the current
dialect's `resolveExpr` collapses to a string before filtering — see
implementation note below for how the array case is threaded through
without breaking the existing string-only filter pipeline).

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestApplyFilterJoin|TestReadComputedFieldsStaticLiteral' -v
PASS
```

---

## Final full-suite verification

All items (B1, F1, F9, F8, M1/F2/m2, F3, F4, F5, join-filter/static-literal)
landed. Final commands run from repo root, all GREEN:

```
$ go build ./...                                        # exit 0
$ go vet ./internal/connectors/...                       # exit 0, no findings
$ go test ./internal/connectors/... 2>&1 | tail -8       # exit 0, every package "ok"
$ go test ./internal/connectors/engine/... -cover        # ok, coverage: 85.7% of statements
$ go run ./cmd/connectorgen validate internal/connectors/defs
    connectorgen validate: 3 connector(s) checked, 0 findings
$ make lint                                              # 0 issues
$ gofmt -l internal/connectors                           # empty (clean)
```

Files touched (exclusively `internal/connectors/engine/**`, confirmed via
`git status --porcelain` — no `internal/connectors/defs/stripe/**` change was
needed for the honest B1 parity fix; only the engine-side parity TEST was
updated to feed the app-persisted cursor shape):

- internal/connectors/engine/read.go
- internal/connectors/engine/read_test.go
- internal/connectors/engine/paginate.go
- internal/connectors/engine/paginate_test.go
- internal/connectors/engine/interpolate.go
- internal/connectors/engine/interpolate_test.go
- internal/connectors/engine/auth.go
- internal/connectors/engine/auth_test.go
- internal/connectors/engine/bundle.go
- internal/connectors/engine/connector.go
- internal/connectors/engine/connector_test.go
- internal/connectors/engine/schema.go
- internal/connectors/engine/write.go (ctx threaded through newRuntime call site only)
- internal/connectors/engine/parity_stripe_test.go (B1 honest-parity cursor fix)
- .planning/phases/wave0-engine-harness/traces/gaploop-r1-ledger.md (this file)

No dependency additions, no schema/migration changes, no auth/security
WEAKENING (M1/F2/m2/F4/F8 all TIGHTEN existing guards), no destructive data
actions, no secret access, no quality-gate reductions. Every existing test
in the touched files was either left unmodified or strengthened (no
assertion was loosened); the one existing test whose input shape changed
(`TestParityStripe_IncrementalCreatedGTEFromState`) had its hand-crafted
RFC3339 cursor replaced with the honest app-persisted unix-seconds cursor
shape per the review's explicit instruction — the assertion strength
(both connectors must agree on the outgoing wire value) is unchanged.
