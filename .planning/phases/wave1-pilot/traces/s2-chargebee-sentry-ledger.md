# Gap-loop cycle 1 — Step 2 TDD ledger (chargebee, sentry majors; vitally/bitly minors)

Executor: gsd-loop-backend. HEAD at start: dc7ad63, branch connector-architecture-v2.
Scope per GAP-LOOP-PLAN.md Step 2 + REVIEW-A.md (chargebee/sentry majors) + REVIEW-B.md
(vitally/bitly minor flags). Step 1 (engine mini-wave, HEAD dc7ad63) already landed the typed
computed_fields extraction, optional-query dialect, `last_path_segment` filter, `stop_path` cursor
paginator, and spec-default materialization this dispatch depends on (see
`traces/gaploop-s1-ledger.md`).

Files touched: `internal/connectors/defs/{chargebee,sentry,vitally,bitly}/**`,
`internal/connectors/paritytest/{chargebee,sentry,vitally,bitly}/**`. No engine/hook Go code
required changes beyond what Step 1 already landed (sentry hostname fix is docs+spec+streams only,
no hook edit needed — confirmed below).

---

## CHARGEBEE item 1 — adopt typed computed_fields; retype schemas; flip stringify test; fix TestConformance

### Starting RED (confirms Step-1 handoff, re-verified independently before any edit)

```
$ go test ./internal/connectors/paritytest/chargebee/... -run TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields -v
--- FAIL: TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields (0.01s)
    parity_test.go:541: engine customers[0].created_at = "1700000000" (json.Number), want string
    (computed_fields always stringifies; conventions.md §5 chargebee deviation)
FAIL

$ go test ./internal/connectors/conformance/... -run 'TestConformance/chargebee' -v
--- FAIL: TestConformance/chargebee (0.01s)
    records_match_schema: stream "customers": record failed schema validation:
    /created_at: value does not match type [string null]
FAIL
```

Both failures are exactly the Step-1 ledger's predicted breakage: bare `{{ record.<path> }}`
computed_fields now preserve native JSON types (numbers/bools), but chargebee's schemas still
declare `["string","null"]`-widened types and the parity test still asserts the old stringified
form. Ground truth for the real wire type of each field is
`internal/connectors/chargebee/streams.go`'s `chargebee*Fields()` catalogs (cross-checked against
`chargebee*Record()` mappers, which pass every field through unchanged from the raw
`json.Decoder.UseNumber()`-decoded envelope).

### GREEN — schemas retightened to native types

Retyped every numeric/boolean property (was `["string","null"]` + a "stringified by
computed_fields" description) back to its real wire type, per streams.go:
- integer fields (Unix-seconds timestamps AND plain integers): `["integer","null"]`
  - customers: `net_term_days`, `created_at`, `updated_at`
  - subscriptions: `plan_quantity`, `plan_amount`, `current_term_start`, `current_term_end`,
    `created_at`, `started_at`, `updated_at`
  - invoices: `total`, `amount_paid`, `amount_due`, `date`, `due_date`, `paid_at`, `updated_at`
  - plans: `price`, `period`, `created_at`, `updated_at`
  - items: `created_at`, `updated_at`
- boolean fields: `["boolean","null"]`
  - customers: `deleted`
  - subscriptions: `deleted`
  - invoices: `deleted`
  - items: `is_shippable`, `enabled_for_checkout`
- all remaining string fields (`id`, names, statuses, currency codes, period_unit, etc.) unchanged.

`x-primary-key`/`x-cursor-field`/`required` unchanged. Stale "stringified by computed_fields ...
see docs.md Known limits" descriptions removed from every retyped property (no longer true).

### GREEN — parity test flipped to native-type equality

`TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields` renamed
`TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes` (behavior it now asserts
is the opposite of its old name) — asserts the engine's `created_at` is now `json.Number("1700000000")`
(matching legacy's `json.Number` exactly, not just same-valued) and `deleted` is a native Go `bool`
`false` (matching legacy's `bool` exactly). Kept as a dedicated companion assertion per REVIEW-A.md
A2 rule 4 (a type-shape guarantee still deserves its own pinned test, even once "same as legacy"),
now proving RESOLVED instead of ledgering a known deviation.

`normalizeRecordStringify`/`stringifyAny` helpers in the parity test file are UNCHANGED (still used
by `TestParityChargebee_StreamRecords` etc. as a tolerant same-DATA compare) — with native types now
flowing through both sides, the stringify-based compare still holds (a string-form comparison of
identical typed values is trivially still equal), so no other test needed edits; verified by full
suite rerun below.

### Verification

```
$ go test ./internal/connectors/paritytest/chargebee/... -v 2>&1 | tail -30
--- PASS: TestParityChargebee_StreamRecords (all 5 streams)
--- PASS: TestParityChargebee_CustomersTwoPagePagination
--- PASS: TestParityChargebee_IncrementalUpdatedAtFromState
--- PASS: TestParityChargebee_IncrementalUpdatedAtFromStartDate
--- PASS: TestParityChargebee_BasicAuthHeader
--- PASS: TestParityChargebee_ErrorPathNon2xx
--- PASS: TestParityChargebee_WriteUnsupported
--- PASS: TestParityChargebee_CatalogSurface
--- PASS: TestParityChargebee_BundleLoadsAndValidates
--- PASS: TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes
PASS

$ go test ./internal/connectors/conformance/... -run 'TestConformance/chargebee' -v
--- PASS: TestConformance/chargebee
PASS
```

(No `TestParityChargebee_IncrementalSortByAscUpdatedAt` test was added — see item 2's STOP
disposition immediately below: the engine has no mechanism to express this correctly within this
dispatch's file scope, so no new test was added for it. Full transcripts captured again in the
final self-verify section at the end of this ledger.)

---

## CHARGEBEE item 2 — sort_by[asc]=updated_at on incremental requests (optional-query dialect)

Per REVIEW-A.md chargebee major finding 1: legacy sets `sort_by[asc]=updated_at` alongside
`updated_at[after]` on every incremental request (`chargebee.go:152-154`, ONLY inside the
`if updatedAfter != ""` branch) — never on a full-refresh read. The optional-query dialect
(Step-1 item 3) is the exact mechanism: `omit_when_absent` keyed off the SAME config reference
(`config.start_date`) that the `incremental` block's own `start_config_key` already names would not
by itself capture the STATE-cursor path (an app-persisted cursor is not a config key at all) — so
instead this uses the dialect's `omit_when_absent` against the stream's own incremental cursor
resolution. Re-reading `engine/read.go`'s `buildInitialQuery`/incremental-handling to confirm the
exact mechanism available:

### STOP — genuine engine-scope gap, not resolvable within this dispatch's file scope

Re-read `engine/read.go`'s `buildInitialQuery`/`IncrementalSpec` (Step 1's actual landed shape,
not assumed) before writing any RED test, since the fix path matters:

- Legacy's condition for sending `sort_by[asc]=updated_at` is exactly "the SAME condition that
  puts `updated_at[after]` into the request" (`chargebee.go:151-155`: both are set inside
  `if updatedAfter != ""`, where `updatedAfter` = state cursor OR `start_date` fallback,
  `incrementalLowerBound`).
- The engine computes the equivalent value as `buildInitialQuery`'s local `lower` (via
  `incrementalLowerBoundValue`, which reads `connsdk.Cursor(req.State)` first, `stream.Incremental
  .StartConfigKey`'s config value second) — but `lower` is NEVER exposed to `stream.Query` template
  resolution: `buildInitialQuery` calls `Interpolate(param.Template, vars)` with
  `vars := requestVars(req.Config, nil, "")` (Cursor hardcoded to `""`) BEFORE computing `lower`,
  and nothing wires `lower`'s value back into `vars.Cursor` (or any other Vars field) for the
  `stream.Query` map's own resolution pass.
- Confirmed independently that the app/worker layer (`internal/app/app.go:493-501`) does NOT
  round-trip the persisted cursor into `config.start_date`: it sets `readConfig.Config["since"] =
  prior.Cursor` (a DIFFERENT key chargebee's `IncrementalSpec.StartConfigKey` never names) and
  passes the cursor separately via `State["cursor"]`. So an optional-query entry templated as
  `"sort_by[asc]": {"template": "updated_at", "omit_when_absent": true}` gated on, say,
  `config.start_date`'s presence CANNOT be built — `omit_when_absent`'s absent-key detection is
  keyed to the TEMPLATE's own referenced config/secret keys resolving or not, and there is no
  config/secret key whose presence tracks "the incremental lower bound resolved" on the
  repeat-sync/state-cursor path (the common case — the very first sync after `start_date` is the
  ONLY case a `config.start_date`-gated entry would get right; every subsequent incremental sync
  driven by the persisted `State["cursor"]` would silently omit `sort_by[asc]` where legacy sends
  it, which is a worse, silent, undetected divergence than today's already-ledgered gap).
- No `hooks/chargebee` package exists (chargebee currently uses zero hook interfaces — REVIEW-A.md
  confirms `escape_hatches_justified=true (none used)`), and this dispatch's FILES list does not
  authorize creating one (only `hooks/sentry/**`, conditionally, is in scope) — a new Tier-2 hook
  package for a single static query param would also be a disproportionate escalation for what
  REVIEW-A itself frames as an engine-dialect gap, not a per-connector hook gap.
- The engine's `stream.Query` resolution pass runs BEFORE `buildInitialQuery` computes `lower`, and
  wiring `lower` into the `Vars` available to `stream.Query` templates is a change to
  `internal/connectors/engine/read.go` — outside this dispatch's FILES scope (engine/** is not
  listed; only `hooks/sentry/**` is, conditionally).

**Disposition: STOP condition — human/engine-scope gate reached for this sub-item only.** Per the
role brief's stop conditions ("missing required context... same failure repeats without new
evidence" and the requirement to route engine changes through the mandated engine mini-wave
process, conventions.md §6), this specific fix requires a small, well-scoped engine increment
(expose the resolved incremental lower bound to `stream.Query`'s own Vars — e.g. populate
`vars.Cursor` with `lower` before the `stream.Query` resolution loop, or add a dedicated
`IncrementalSpec` field such as `"extra_param_when_incremental": {"sort_by[asc]": "updated_at"}`)
that is NOT expressible via defs/paritytest-only edits and NOT authorized by this dispatch's file
scope. Ledgered here (not silently dropped) for the P-12/orchestrator queue: this is the SAME
recurrence-class gap REVIEW-A already named ("a param sent only-when-incremental is the same
conditional-query class" as REVIEW-B's optional-query gap) — Step 1 closed the
config/secret-absence half of that class (`omit_when_absent`/`default`) but not the
incremental-lower-bound-presence half, which is a distinct condition (state-cursor-aware, not
config-key-keyed). No test was added for this sub-item (adding a test the engine cannot make pass
without an out-of-scope change would just be a second, permanently-red RED with no GREEN available
in this dispatch — not useful TDD evidence). `docs.md`'s existing "Known limits" note stays as the
ledgered, documented deviation (see below) rather than being silently removed, since the underlying
divergence is unresolved.

**No behavior/test change made for chargebee item 2.** `internal/connectors/defs/chargebee/
streams.json`'s `query` blocks and `internal/connectors/paritytest/chargebee/parity_test.go` are
UNCHANGED from HEAD for this specific sub-item (item 1 and item 3 changes below are independent and
proceed normally).

`docs.md` records the open gap explicitly (new "OPEN" Known-limits bullet, see below) rather than
leaving it silently undocumented, and the old stringify-deviation bullet it previously lived next
to is now marked RESOLVED (item 1).

---

## CHARGEBEE item 3 — `site` dead config resolved (drop + require base_url, per C3 guidance)

Per REVIEW-A.md chargebee major finding 2 + C3: `site` is dead config (nothing in `streams.json`
references `config.site`; only `chargebeeBaseURL` in the LEGACY package derives
`https://{site}.chargebee.com/api/v2` from it) and `docs.md`'s claim that base_url is "derived from
the required-ish `site` config value ... matching legacy's `chargebeeBaseURL`" is false — a
`site`-only config (no `base_url`) hard-errors in this bundle (`{{ config.base_url }}` has no
fallback). Cross-checked against `docs/migration/conventions.md`'s C3 guidance (added by Step 1,
item 6/7): "For a DERIVED default (sentry's hostname-based URL, chargebee's site-based URL — the
base URL is a function of another config value, not a fixed literal) [spec-default materialization]
alone is not enough; either require base_url and drop the derivation ..., or express the derivation
as a computed_fields-style template if/when the dialect grows one ... — do not invent ad hoc Go for
it." The dialect has no such derived-default template mechanism today (`Schema.Defaults()` only
stringifies a LITERAL declared value, never a cross-key template — confirmed by reading
`engine/schema.go:488-503` directly), so per the conventions' own prescribed choice: **drop `site`,
require `base_url`.**

No RED/GREEN test cycle applies (this is a spec.json/docs.md-only change — dropping an already-dead
key and tightening `required[]`); verified instead by re-running the full chargebee suite after the
edit (no test in `paritytest/chargebee` ever set `config.site` — confirmed by grep before editing —
so this could not regress any existing assertion) and by `connectorgen validate`'s
`required`/`default_type_mismatch` rules accepting the new `spec.json` shape.

### GREEN

- `internal/connectors/defs/chargebee/spec.json`: `required` is now `["site_api_key", "base_url"]`;
  the `site` property is REMOVED entirely; `base_url`'s `description` rewritten to state it is
  required and to name the dropped derivation + why (cross-reference to docs.md).
- `internal/connectors/defs/chargebee/docs.md`: "Auth setup" paragraph rewritten (no more false
  "derived from site" claim); new "Known limits" bullet ("`site` config key dropped; `base_url` is
  now required") documenting the config-surface narrowing per conventions.md's requirement that
  every dropped/renamed config key gets a ledgered docs.md note.

### Verification

```
$ grep -rn '"site"' internal/connectors/paritytest/chargebee/ internal/connectors/conformance/
(no output — confirms no test ever set config.site; dropping it cannot regress any assertion)

$ go run ./cmd/connectorgen validate internal/connectors/defs 2>&1 | grep -i chargebee
(no output — 0 findings for chargebee)

$ go test ./internal/connectors/paritytest/chargebee/... -v 2>&1 | tail -15
--- PASS: TestParityChargebee_StreamRecords (all 5 streams)
--- PASS: TestParityChargebee_CustomersTwoPagePagination
--- PASS: TestParityChargebee_IncrementalUpdatedAtFromState
--- PASS: TestParityChargebee_IncrementalUpdatedAtFromStartDate
--- PASS: TestParityChargebee_BasicAuthHeader
--- PASS: TestParityChargebee_ErrorPathNon2xx
--- PASS: TestParityChargebee_WriteUnsupported
--- PASS: TestParityChargebee_CatalogSurface
--- PASS: TestParityChargebee_BundleLoadsAndValidates
--- PASS: TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes
PASS

$ go test ./internal/connectors/conformance/... -run 'TestConformance/chargebee' -v
--- PASS: TestConformance/chargebee
PASS
```

All chargebee items resolved except item 2 (STOPPED — genuine engine-scope gap, ledgered above and
in docs.md).

---

## SENTRY — hostname dead config resolved (drop + require base_url, per C3 guidance)

Per REVIEW-A.md sentry major finding: `hostname` is dead config in the bundle (nothing in
`streams.json` references `config.hostname`) while `docs.md`'s Auth setup claims base_url resolution
"match[es] legacy's `sentryBaseURL` resolution" — false, since legacy's `sentryBaseURL`
(`sentry.go:293-308`) derives `https://<hostname>` (default `sentry.io`) when `base_url` is unset,
and this bundle's `{{ config.base_url }}` template has no such fallback; a legacy-canonical
`hostname`-only (or fully-default, no-hostname) config hard-errors here. Identical class to
chargebee's `site` finding — same C3 guidance applies, same fix shape: **drop `hostname`, require
`base_url`.**

### Hook-touch check (per dispatch: "hooks/sentry/** only if hostname fix requires")

Confirmed NOT required before making any change:
- `grep -n "hostname" internal/connectors/hooks/sentry/*.go` → no output. `hooks/sentry/hooks.go`'s
  `StreamHook` (Link-header pagination) never reads `cfg.Config["hostname"]` or resolves the base
  URL itself — it operates on an already-configured `*connsdk.Requester`/bundle base URL supplied by
  the engine's declarative `base.url` resolution, same as every other stream. Dropping `hostname`
  from `spec.json` therefore cannot affect `hooks/sentry/**` at all.
- `grep -n '"hostname"\|"base_url"' internal/connectors/paritytest/sentry/parity_test.go` → only
  `"base_url"` (`sentryRuntimeConfig` always sets it explicitly on every test call); no test ever
  set `config.hostname`. Dropping it cannot regress any existing parity assertion.

**No `hooks/sentry/**` edit made** — confirmed unnecessary by the above, consistent with the
dispatch's conditional scope grant.

### GREEN

- `internal/connectors/defs/sentry/spec.json`: `required` is now `["auth_token", "base_url"]`; the
  `hostname` property is REMOVED entirely; `base_url`'s `description` rewritten to state it is
  required and to name the dropped derivation + why.
- `internal/connectors/defs/sentry/docs.md`: "Auth setup" paragraph rewritten (no more false
  "matching legacy's sentryBaseURL resolution" claim); new "Known limits" bullet ("`hostname` config
  key dropped; `base_url` is now required") documenting the config-surface narrowing, cross-linking
  chargebee's identical-class fix and conventions.md's derived-default guidance.

### Verification

```
$ go run ./cmd/connectorgen validate internal/connectors/defs 2>&1 | grep -i sentry
(no output — 0 findings for sentry)

$ go test ./internal/connectors/paritytest/sentry/... -v 2>&1 | tail -20
--- PASS: TestParitySentry_ProjectsStreamRecords
--- PASS: TestParitySentry_IssuesTwoPagePaginationAndResultsFalseStop
--- PASS: TestParitySentry_EventsAndReleasesStreamRecords (events, releases)
--- PASS: TestParitySentry_BearerAuthHeaderParity
--- PASS: TestParitySentry_NonSuccessStatusErrorsBothSides
--- PASS: TestParitySentry_BundleLoadsAndValidates
--- PASS: TestParitySentry_CatalogSurface
--- PASS: TestParitySentry_HostileBaseURLFailsClosedBothSides
PASS

$ go test ./internal/connectors/hooks/sentry/... -v 2>&1 | tail -10
--- PASS: TestReadStream_LinkHeaderResultsFalseStopsAfterTwoPages
--- PASS: TestReadStream_NoLinkHeaderStopsAfterOnePage
--- PASS: TestReadStream_ProjectionKeepsOnlySchemaProperties
--- PASS: TestReadStream_MissingOrganizationErrorsForScopedStream
--- PASS: TestReadStream_UnknownStreamFallsBackToDeclarative
--- PASS: TestConnectorNameAndRegistration
PASS

$ go test ./internal/connectors/conformance/... -run 'TestConformance/sentry' -v
--- PASS: TestConformance/sentry
PASS
```

Sentry major fully RESOLVED.

---

## VITALLY — docs.md Known-limits sentence for the fail-loud Check

Per REVIEW-B.md vitally minor finding 1: legacy `Check` (`vitally.go:33-47`) validates config/secret
presence OFFLINE only — it never dials the network. This bundle's `base.check`
(`streams.json:8`, `{"method":"GET","path":"/resources/accounts"}`) issues a real HTTP request.
Fail-loud improvement (a bad credential or unreachable host is now caught at `Check` time instead of
the first `Read`), zero record-data impact, but an undocumented behavior deviation per REVIEW-B.

Docs-only change; no RED/GREEN test cycle applies (adding a sentence to an existing "Known limits"
list does not change any assertable behavior). Verified by re-running the full vitally suite
(unaffected, as expected for a docs-only edit) and independently re-confirming the legacy-vs-bundle
Check behavior claim against `internal/connectors/vitally/vitally.go` and
`internal/connectors/defs/vitally/streams.json` before writing the sentence.

### GREEN

`internal/connectors/defs/vitally/docs.md`: new "Known limits" bullet ("`Check` now dials the
network; legacy's `Check` never did") inserted before the existing `status`-filter bullet, naming
the legacy line reference, the bundle's `base.check` config, and framing it explicitly as a
deliberate strictly-improving deviation (not a regression) per REVIEW-B's own characterization.

### Verification

```
$ go test ./internal/connectors/paritytest/vitally/... -v 2>&1 | tail -15
--- PASS: TestParityVitally_AccountsStreamRecords
--- PASS: TestParityVitally_NoStatusParamSentWhenUnset
--- PASS: TestParityVitally_AuthorizationHeaderByteExact
--- PASS: TestParityVitally_NonSuccessStatusErrorsOnBothSides
--- PASS: TestParityVitally_WriteUnsupportedOnBothSides
--- PASS: TestParityVitally_BundleLoadsWithSingleAccountsStream
--- PASS: TestParityVitally_CheckRequiresAuthSecretOnBothSides
PASS

$ go test ./internal/connectors/conformance/... -run 'TestConformance/vitally' -v
--- PASS: TestConformance/vitally
PASS

$ go run ./cmd/connectorgen validate internal/connectors/defs 2>&1 | grep -i vitally
(no output — 0 findings for vitally)
```

Vitally minor RESOLVED (docs-only, as scoped).

---

## BITLY — docs.md + parity-test comment false size=50-first-request-only claim

Per REVIEW-B.md bitly minor findings 1+2: `docs.md`'s "Streams notes" claimed `size=50` "is sent as
a static per-stream query value on the FIRST request only — subsequent pages are driven entirely by
the absolute `pagination.next` URL ... matching legacy's `harvest` loop". FALSE for the engine:
`readDeclarative` merges `stream.Query` into EVERY page request (`engine/read.go`'s
`mergeQuery(baseQuery, page.Query)`), and `connsdk.Requester.resolveURL` re-applies it onto the
absolute next URL (Del+Add, replacing any same-named param) — legacy explicitly clears the query to
an empty `url.Values{}` once it follows an absolute next-page URL (`bitly.go:180-183`), so `size` is
sent on page 1 only there. Verified benign in DATA terms (Bitly's own next URL already carries the
identical `size` value the engine re-applies; the replace is idempotent) but the doc claim is wrong.
`paritytest/bitly/parity_test.go`'s `bitlyTwoPageServer` doc comment made the same false claim
("asserts the request-shape (... size query param on page 1 only) matches legacy") with no such
assertion actually present in the handler (only path + `search_after` are matched by the switch).

### RED (confirms the real engine behavior the docs/comment got wrong, before any test/doc edit)

```
$ go test ./internal/connectors/paritytest/bitly/... -run TestParityBitly_BitlinksStreamPaginates -v
--- PASS: TestParityBitly_BitlinksStreamPaginates
(pre-existing test passes today — it never asserted the size-param-per-page shape either way, which
is exactly finding 2's point: no test/assertion exists for the true per-page query shape)
```

Wrote a throwaway probe (not committed as a permanent test) to confirm the real wire shape before
touching any file:
```
$ cat <<'EOF' > /tmp/probe_test.go
... (readAllBitlyRecords against bitlyTwoPageServer, capturing r.URL.RawQuery per request) ...
EOF
$ go test ./internal/connectors/paritytest/bitly/... -run TestProbeSizeParam -v
    page 2 request query = "search_after=tok2&size=50"   (confirms size=50 IS present on page 2 —
    the docs/comment claim of "page 1 only" is false, exactly as REVIEW-B found)
(probe file removed after confirming; the real fix adds a permanent, named assertion instead)
```

### GREEN

- `internal/connectors/defs/bitly/docs.md` "Streams notes": rewritten to state `size=50` is
  re-sent on every page (not first-request-only), explain the engine mechanism
  (`mergeQuery`/`resolveURL`'s Del+Add re-apply) vs legacy's query-reset-on-absolute-next-URL
  behavior, and state precisely why this is verified benign (Bitly's own next URL already carries
  the identical value; the replace is idempotent) rather than claiming false byte-identical
  behavior.
- `internal/connectors/paritytest/bitly/parity_test.go`:
  - `bitlyTwoPageServer`'s doc comment rewritten to state the real, honest scope of what the
    handler proves (group-scoped path + 2-page termination) and explicitly why it does NOT assert
    "no size on page 2" (because the engine legitimately sends one).
  - New test `TestParityBitly_BitlinksSizeParamResentOnEveryPage`: parses both requested pages'
    query strings and asserts `size=50` is present on BOTH page 1 and page 2 — the honest,
    permanent, named assertion of the actual engine behavior the old comment falsely claimed did
    not exist, per REVIEW-A.md A2 rule 4's "pinned by a companion assertion" discipline extended to
    this docs-correction case.
  - Added `net/url` + `strings` imports for the new test's query-string parsing.

### Verification

```
$ go vet ./internal/connectors/paritytest/bitly/...
(clean)

$ go test ./internal/connectors/paritytest/bitly/... -v 2>&1 | tail -25
--- PASS: TestParityBitly_GroupsStreamRecords
--- PASS: TestParityBitly_OrganizationsStreamRecords
--- PASS: TestParityBitly_CampaignsStreamRecords
--- PASS: TestParityBitly_BitlinksSizeParamResentOnEveryPage
--- PASS: TestParityBitly_BitlinksStreamPaginates
--- PASS: TestParityBitly_BitlinksAbsoluteNextURLNotRelative
--- PASS: TestParityBitly_BearerAuthHeaderParity
--- PASS: TestParityBitly_ErrorPathParity
--- PASS: TestParityBitly_BundleLoadsAndValidates
PASS

$ go test ./internal/connectors/conformance/... -run 'TestConformance/bitly' -v
--- PASS: TestConformance/bitly
PASS

$ go run ./cmd/connectorgen validate internal/connectors/defs 2>&1 | grep -i bitly
(no output — 0 findings for bitly)
```

Bitly minors RESOLVED (docs.md + parity-test comment fixed, honest companion assertion added).

---

## Final self-verify (whole Step-2 dispatch: chargebee, sentry, vitally, bitly)

```
$ go build ./...
(clean)

$ go vet ./...
(clean)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/conformance/... -run 'TestConformance/(chargebee|sentry|vitally|bitly)' -v
--- PASS: TestConformance
    --- PASS: TestConformance/bitly
    --- PASS: TestConformance/chargebee
    --- PASS: TestConformance/sentry
    --- PASS: TestConformance/vitally
PASS

$ go test ./internal/connectors/paritytest/chargebee/... ./internal/connectors/paritytest/sentry/... \
    ./internal/connectors/paritytest/vitally/... ./internal/connectors/paritytest/bitly/... -v
(all subtests PASS — see per-connector sections above for full transcripts)

$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... \
  ./internal/connectors/hooks/... ./internal/connectors/native/... \
  ./internal/connectors/conformance/... ./internal/connectors/certify/... \
  ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.

$ go test ./internal/connectors/... 2>&1 | grep -v '^ok'
?   	polymetrics.ai/internal/connectors/defs	[no test files]
?   	polymetrics.ai/internal/connectors/hooks/hookset	[no test files]
?   	polymetrics.ai/internal/connectors/native/nativeset	[no test files]
--- FAIL: TestParityGmail_ComputedFieldsStringifyLabelCountFields (0.00s)
FAIL	polymetrics.ai/internal/connectors/paritytest/gmail
```

The ONLY remaining failure anywhere in `internal/connectors/...` is gmail's own Step-1-predicted
stringify breakage — explicitly OUT OF SCOPE for this dispatch (gmail's chargebee-equivalent
schema/parity-test/docs.md re-tightening is a separate Step-2 sub-task per
`gaploop-s1-ledger.md`'s handoff list; not in this dispatch's FILES:
`defs/{chargebee,sentry,vitally,bitly}/**`). All four connectors in THIS dispatch's scope
(chargebee, sentry, vitally, bitly) are fully green: build, vet, lint, connectorgen validate (0
findings), TestConformance, and every paritytest package.

## Summary of dispositions

| Item | Status |
|---|---|
| chargebee item 1 (typed computed_fields, schema retype, stringify test flip, TestConformance fix) | RESOLVED |
| chargebee item 2 (sort_by[asc]=updated_at via optional-query dialect) | **STOPPED — engine-scope gap, out of file-scope authority; ledgered in docs.md + this file for P-12/next engine mini-wave** |
| chargebee item 3 (site dead config) | RESOLVED (dropped `site`, required `base_url`, docs fixed) |
| sentry (hostname dead config) | RESOLVED (dropped `hostname`, required `base_url`, docs fixed; no hook edit needed) |
| vitally (fail-loud Check docs note) | RESOLVED (docs-only) |
| bitly (docs.md + test comment false size=50-first-request-only claim) | RESOLVED (docs.md fixed; test comment fixed + honest companion assertion added) |
