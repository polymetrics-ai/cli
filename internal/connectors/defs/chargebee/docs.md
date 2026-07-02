# Overview

Chargebee is a wave1-pilot Tier-1 declarative migration (PLAN.md P-6, SPEC.md §5.4). It reads
Chargebee customers, subscriptions, invoices, plans, and items through the Chargebee v2 REST API.
This bundle is engine-vs-legacy parity-tested against `internal/connectors/chargebee` (the
hand-written connector it migrates); the legacy package stays registered and unchanged until
wave6's registry flip. Chargebee is read-only in both legacy and this bundle (no `writes.json`).

## Auth setup

Provide a Chargebee site API key via the `site_api_key` secret; it is used as the HTTP Basic
username with an empty password (`auth: [{"mode":"basic","username":"{{ secrets.site_api_key }}",
"password":""}]`), matching legacy's `connsdk.Basic(secret, "")` exactly (chargebee.go:262-264,278).
The API host is `base_url`, which is **required** in this bundle (`spec.json`'s
`required: ["site_api_key", "base_url"]`) — e.g. `https://{site}.chargebee.com/api/v2`. Legacy
instead DERIVES the host from a `site` config value (`chargebeeBaseURL`,
`"https://" + site + ".chargebee.com/api/v2"` when `base_url` is unset); this bundle does not
reproduce that derivation (see "Known limits" below for why and the config-surface change this
implies for an operator migrating a legacy-shaped config).

## Streams notes

All 5 streams (`customers`, `subscriptions`, `invoices`, `plans`, `items`) share the same shape:
`GET` against the Chargebee list endpoint, records at the top-level `list` array, each element
wrapped in a single-key resource envelope (e.g. `{"customer": {...}}`). Pagination follows
Chargebee's `offset`/`next_offset` convention (`pagination.type: cursor` with `cursor_param: offset`
and `token_path: next_offset`): the next page's `offset` query value is read verbatim from the
previous response body's `next_offset` field, and pagination stops when `next_offset` is absent —
identical to legacy's `harvest()` loop (chargebee.go:148-196). Every request sends `limit=100`
(matches legacy's default `page_size`) via each stream's static `query: {"limit": "100"}`.
Incremental reads send `updated_at[after]` as a Unix-seconds value (`param_format: unix_seconds`),
computed either from the sync's persisted cursor or, on a fresh sync, from the RFC3339 `start_date`
config value — identical to legacy `incrementalLowerBound`/`formatParam`. Primary key is `["id"]`
and the incremental cursor field is `["updated_at"]` across every stream, matching legacy's
`chargebeeStreams()` catalog uniformly. **RESOLVED — `sort_by[asc]=updated_at` is now reproduced**:
legacy also sends `sort_by[asc]=updated_at` alongside `updated_at[after]` on every incremental
request (`chargebee.go:152-154`), never on a full-refresh read; this bundle now expresses that via
the `incremental.lower_bound` query-var dialect (S3 engine mini-wave item 1) — see "Known limits"
below for the fix and `paritytest/chargebee/parity_test.go`'s
`TestParityChargebee_SortByAscSentOnIncrementalFromState`/
`TestParityChargebee_SortByAscSentOnIncrementalFromStartDate`/
`TestParityChargebee_SortByAscOmittedOnFullSync`.

**Envelope unwrap via per-field `computed_fields`** (conventions.md §2 schema-as-projection):
Chargebee wraps every list item in a single-key resource envelope, so plain schema projection
(which looks up each schema property directly on the raw extracted record) sees only that one
wrapper key and produces an empty record. Every schema property is therefore populated by a
`computed_fields` entry reaching into the envelope (e.g. `"id": "{{ record.customer.id }}"`,
`"created_at": "{{ record.customer.created_at }}"`), matching legacy's `chargebeeCustomerRecord`
(and its 4 sibling `mapRecord` functions in streams.go) field-for-field — including TYPE, not just
value: every computed_fields entry here is a single bare `{{ record.<envelope>.<field> }}`
reference with no filter stage, so the engine's typed computed_fields extraction (gap-loop cycle-1
item 1) copies the raw JSON value straight through (numeric/boolean fields preserve their native
type instead of being stringified). Schemas declare the real wire type
(`integer`/`boolean`/`string`) per field, matching `chargebeeStreams()`'s field catalog exactly; see
`paritytest/chargebee/parity_test.go`'s
`TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes`.

## Write actions & risks

None — Chargebee is exposed as a read-only source connector in both legacy
(`chargebee.go:258-260`'s `Write` returns `connectors.ErrUnsupportedOperation`) and this bundle
(`metadata.json`'s `capabilities.write: false`, no `writes.json` file at all, matching the
searxng read-only-variant pattern in conventions.md §1).

## Known limits

- Full Chargebee API surface (coupons, credit notes, addons, hosted pages, events, webhooks) is out
  of scope for wave1-pilot; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 5 legacy-parity streams are implemented.
- **RESOLVED — computed_fields envelope unwrap now preserves native numeric/boolean types.**
  Previously (pre gap-loop-cycle-1), every schema field derived via a `computed_fields` envelope
  unwrap was stringified by `engine.Interpolate` regardless of the raw JSON value's real type,
  which forced every numeric/boolean schema property to a widened `["string", "null"]` type. The
  engine's typed computed_fields extraction (gap-loop cycle-1 item 1: a bare
  `{{ record.<path> }}` template with no filter stage copies the raw typed value instead of
  stringifying) now applies to every computed_fields entry in this bundle, so schemas declare the
  real wire type (`integer` for Unix-seconds timestamps and plain integers, `boolean` for booleans)
  matching legacy's `chargebeeStreams()` field catalog and `mapRecord` functions exactly, TYPE
  included. Asserted by `paritytest/chargebee/parity_test.go`'s
  `TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes`.
  - **Why not a `RecordHook` instead** (SPEC §5.4's suggested fallback for cases computed_fields
    cannot reproduce exactly): `internal/connectors/conformance/dynamic.go`'s dynamic checks
    (`checkReadFixtureNonempty`, `checkPaginationTerminates`, `checkRecordsMatchSchema`,
    `checkCursorAdvances`) all call `engine.Read`/`engine.Check` with a literal `nil` Hooks
    parameter — a `RecordHook` would never fire during conformance, so `checkRecordsMatchSchema`
    would validate the schema against the still-envelope-wrapped raw record (one top-level key)
    instead of a flattened one, failing hard for every stream regardless of hook correctness.
    `computed_fields` is therefore the only mechanism whose output conformance actually exercises;
    with typed extraction it now ALSO preserves the real wire type, closing the gap this note
    originally documented. See `.planning/phases/wave1-pilot/traces/p6-chargebee-ledger.md` and
    `.planning/phases/wave1-pilot/traces/gaploop-s1-ledger.md`/`s2-chargebee-sentry-ledger.md` for
    the full design-decision trace.
- ~~**OPEN — `sort_by[asc]=updated_at` is not sent on incremental requests.**~~ **RESOLVED (S3 engine
  mini-wave item 1).** Legacy sets `sort_by[asc]=updated_at` alongside `updated_at[after]` on every
  incremental request whenever the computed lower bound is non-empty (`chargebee.go:151-155`), never
  on a full-refresh read. The engine now exposes the RESOLVED, post-`formatParam` incremental lower
  bound to `stream.Query` template resolution as `{{ incremental.lower_bound }}` (populated in
  `buildInitialQuery` BEFORE the query-template resolution loop runs, so it reflects EITHER the
  persisted `state.cursor` OR the `start_config_key` fallback — exactly the same value/precedence
  `updated_at[after]` itself uses). Composed with the existing `omit_when_absent` dialect and the new
  `const:<value>` filter (send a FIXED literal iff a reference resolves, without depending on the
  reference's own value), each stream's `query` now declares:
  ```json
  "sort_by[asc]": { "template": "{{ incremental.lower_bound | const:updated_at }}", "omit_when_absent": true }
  ```
  — present with the constant value `updated_at` iff the incremental lower bound resolves (state
  cursor or `start_date`), absent on a full-refresh read, exactly matching legacy's
  `if updatedAfter != ""` gate. See `paritytest/chargebee/parity_test.go`'s
  `TestParityChargebee_SortByAscSentOnIncrementalFromState`/
  `TestParityChargebee_SortByAscSentOnIncrementalFromStartDate`/
  `TestParityChargebee_SortByAscOmittedOnFullSync` and
  `.planning/phases/wave2-fanout-http-sm/traces/s3-engine-ledger.md` for the full design trace; the
  original STOP analysis remains at
  `.planning/phases/wave1-pilot/traces/s2-chargebee-sentry-ledger.md`'s chargebee item 2 section for
  historical reference.
- **`site` config key dropped; `base_url` is now required.** Legacy derives the API host from a
  `site` config value (`https://{site}.chargebee.com/api/v2`) when `base_url` is unset
  (`chargebeeBaseURL`). The engine's spec-default materialization (gap-loop cycle-1 item 6, C3)
  only fills in a LITERAL per-key default — it cannot express "derive `base_url` from `site`", a
  cross-key template. Per `docs/migration/conventions.md`'s guidance for this exact shape (sentry's
  `hostname` hit the identical class), this bundle drops `site` entirely and requires `base_url`
  instead: an operator migrating a legacy `site`-only config must now supply the fully-formed
  `https://{site}.chargebee.com/api/v2` URL as `base_url`. This is a documented config-surface
  narrowing (every legacy-accepted `site` value has an operator-reachable `base_url` equivalent; no
  request/data change once configured), not a data-shape regression.
- `metadata.json` declares no `rate_limit` block: legacy chargebee enforces no client-side rate
  limiting (no `rate_limit`/throttle field anywhere in `chargebee.go`), so this bundle adds none
  either, matching conventions.md §3's "informational vs. enforced" rate-limit rule (an absent
  block, not merely an unenforced one, since Chargebee's public rate limit was never documented in
  the legacy package to carry forward informationally).
