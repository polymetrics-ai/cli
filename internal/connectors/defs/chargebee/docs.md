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
The API host is either an explicit `base_url` override (tests/proxies) or derived from the
required-ish `site` config value as `https://{site}.chargebee.com/api/v2` — `base_url` wins when
both are set, matching legacy's `chargebeeBaseURL`.

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
`chargebeeStreams()` catalog uniformly.

**Envelope unwrap via per-field `computed_fields`** (conventions.md §2 schema-as-projection):
Chargebee wraps every list item in a single-key resource envelope, so plain schema projection
(which looks up each schema property directly on the raw extracted record) sees only that one
wrapper key and produces an empty record. Every schema property is therefore populated by a
`computed_fields` entry reaching into the envelope (e.g. `"id": "{{ record.customer.id }}"`,
`"created_at": "{{ record.customer.created_at }}"`), matching legacy's `chargebeeCustomerRecord`
(and its 4 sibling `mapRecord` functions in streams.go) field-for-field.

## Write actions & risks

None — Chargebee is exposed as a read-only source connector in both legacy
(`chargebee.go:258-260`'s `Write` returns `connectors.ErrUnsupportedOperation`) and this bundle
(`metadata.json`'s `capabilities.write: false`, no `writes.json` file at all, matching the
searxng read-only-variant pattern in conventions.md §1).

## Known limits

- Full Chargebee API surface (coupons, credit notes, addons, hosted pages, events, webhooks) is out
  of scope for wave1-pilot; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 5 legacy-parity streams are implemented.
- **Documented parity deviation — computed_fields envelope unwrap stringifies numeric/boolean
  fields.** Every schema field on every stream (`customers`, `subscriptions`, `invoices`, `plans`,
  `items`) is derived via a per-field `computed_fields` template reaching into the raw resource
  envelope (`{{ record.customer.id }}`, etc. — see "Streams notes" above), because plain schema
  projection cannot see past the envelope's single wrapper key. The engine's `computed_fields`
  mechanism resolves every template through `engine.Interpolate`, which always returns a Go
  `string` (see `internal/connectors/engine/interpolate.go`'s `resolveExpr`/`stringify`) —
  regardless of the raw JSON value's real type. Concretely: every Unix-seconds timestamp field
  (`created_at`, `updated_at`, `current_term_start`, `current_term_end`, `date`, `due_date`,
  `paid_at`, `started_at`), every integer field (`net_term_days`, `plan_quantity`, `plan_amount`,
  `total`, `amount_paid`, `amount_due`, `price`, `period`), and every boolean field (`deleted`,
  `is_shippable`, `enabled_for_checkout`) is emitted by this bundle as a STRING
  (`"1700000000"`/`"false"`), while legacy's hand-written `mapRecord` functions
  (`internal/connectors/chargebee/streams.go`) emit the native Go type decoded straight off the raw
  JSON (`int64`/`bool`). Every schema property above is typed `["string", "null"]` (not widened to
  a multi-type union with `integer`/`boolean` — a single honest type matching what this bundle
  ACTUALLY emits, per conventions.md's tight-typing guidance), with an inline `description`
  documenting the real wire type. This never changes the DATA an accepted input carries (a digit
  string and the same-valued JSON number are textually identical information; `"false"` and `false`
  likewise), so it is an ACCEPTABLE, ledgered deviation (conventions.md §5), not a blocker — and it
  is asserted EXPLICITLY (never silently absorbed by a coercing test helper) by
  `paritytest/chargebee/parity_test.go`'s
  `TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields`.
  - **Why not a `RecordHook` instead** (SPEC §5.4's suggested fallback for cases computed_fields
    cannot reproduce exactly): `internal/connectors/conformance/dynamic.go`'s dynamic checks
    (`checkReadFixtureNonempty`, `checkPaginationTerminates`, `checkRecordsMatchSchema`,
    `checkCursorAdvances`) all call `engine.Read`/`engine.Check` with a literal `nil` Hooks
    parameter — a `RecordHook` would never fire during conformance, so `checkRecordsMatchSchema`
    would validate the schema against the still-envelope-wrapped raw record (one top-level key)
    instead of a flattened one, failing hard for every stream regardless of hook correctness.
    `computed_fields` is therefore the only mechanism whose output conformance actually exercises,
    and is used here even though it cannot preserve non-string raw types. See
    `.planning/phases/wave1-pilot/traces/p6-chargebee-ledger.md` for the full design-decision
    trace.
  - `checkCursorAdvances` still passes cleanly against a string-typed `updated_at` cursor field:
    `param_format: unix_seconds`'s digit-string acceptance (B1 fix, `read.go`'s
    `parseLowerBoundTime`) and conformance's own `cursorValueString` (which explicitly tolerates a
    plain `string` cursor value, comparing lexicographically) both already accommodate this shape
    without any additional change.
- `metadata.json` declares no `rate_limit` block: legacy chargebee enforces no client-side rate
  limiting (no `rate_limit`/throttle field anywhere in `chargebee.go`), so this bundle adds none
  either, matching conventions.md §3's "informational vs. enforced" rate-limit rule (an absent
  block, not merely an unenforced one, since Chargebee's public rate limit was never documented in
  the legacy package to carry forward informationally).
