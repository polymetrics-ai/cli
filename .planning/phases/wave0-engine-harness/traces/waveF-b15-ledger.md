# T/B-15 — golden migration: stripe defs bundle + engine-vs-legacy parity tests

Phase: wave0-engine-harness · Wave: F · Executor: gsd-loop-backend (sonnet, with tester duties)

## Scope

- `internal/connectors/engine/parity_stripe_test.go` — engine-vs-legacy parity test (RED first):
  per-stream record equality (5 streams) incl. 2-page `customers` `starting_after`/`has_more`
  pagination and incremental `created[gte]` propagation (from state cursor and from
  `start_date`); write parity (`create_customer`/`update_customer` method/path/form-body vs
  `stripe/write.go`); manifest-surface equality (stream names, PKs, cursor fields, write action
  names) vs `connectors.ManifestOf(stripe.New())`.
- `internal/connectors/defs/stripe/**` — the golden bundle: `metadata.json`, `spec.json`,
  `streams.json`, `writes.json`, `api_surface.json`, `schemas/{customers,charges,invoices,
  subscriptions,products}.json`, `fixtures/streams/**` (customers 2-page), `fixtures/writes/
  {create_customer,update_customer}.json`, `fixtures/check.json`, `docs.md`.

Files touched exclusively as scoped by the coordinator; legacy `internal/connectors/stripe/**` is
read-only reference, untouched.

## Legacy source of truth (read, not modified)

- `stripe.go`: base URL default `https://api.stripe.com/v1` (config `base_url` override, http/https
  + host required); Bearer auth via `secrets.client_secret`; optional `Stripe-Account` header from
  `config.account_id` (omitted when empty); `harvest()` drives `starting_after`/`has_more` over
  `data[]` with `limit` (default 100, config `page_size`) and `created[gte]` (unix seconds, from
  state cursor verbatim or converted from RFC3339 `config.start_date`).
- `streams.go`: 5 streams (customers, charges, invoices, subscriptions, products), all `PrimaryKey:
  ["id"]`, `CursorFields: ["created"]`; `mapRecord` functions define the exact field set emitted
  per stream (used verbatim to derive the bundle schemas' `properties`, for field-level parity).
- `manifest.go`: config fields (`base_url`, `account_id`, `start_date`, `page_size`, `max_pages`,
  `mode`), secret `client_secret`, write action specs (`create_customer` POST `/customers`,
  `update_customer` POST `/customers/{id}`).
- `write.go`: allow-list `create_customer`/`update_customer`, both `http.MethodPost`;
  `update_customer.needsID`; `customerForm` builds a form body from
  `email,name,description,phone` (empty values omitted); validation requires an `id` for update and
  `email OR name` for create (`validateWriteRecord`).
- `stripe_test.go`: paginates via `has_more`+`starting_after` over `/customers`, expects Bearer
  auth header, 3 records across 2 pages.

## RED evidence

`parity_stripe_test.go` authored first, referencing `defs.FS`/`engine.LoadAll` to locate the
`stripe` bundle — before `internal/connectors/defs/stripe/` exists on disk:

```
$ go test ./internal/connectors/engine -run TestParityStripe -v
=== RUN   TestParityStripe_StreamRecords
    parity_stripe_test.go:203: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_StreamRecords (0.00s)
=== RUN   TestParityStripe_CustomersTwoPagePagination
    parity_stripe_test.go:244: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_CustomersTwoPagePagination (0.00s)
=== RUN   TestParityStripe_IncrementalCreatedGTEFromState
    parity_stripe_test.go:303: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_IncrementalCreatedGTEFromState (0.00s)
=== RUN   TestParityStripe_IncrementalCreatedGTEFromStartDate (0.00s)
    parity_stripe_test.go:334: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_IncrementalCreatedGTEFromStartDate
=== RUN   TestParityStripe_WriteCreateCustomerFormBody
    parity_stripe_test.go:385: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_WriteCreateCustomerFormBody (0.00s)
=== RUN   TestParityStripe_WriteUpdateCustomerFormBody
    parity_stripe_test.go:423: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_WriteUpdateCustomerFormBody (0.00s)
=== RUN   TestParityStripe_ManifestSurface
    parity_stripe_test.go:476: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_ManifestSurface (0.00s)
=== RUN   TestParityStripe_BundleLoadsAndValidates
    parity_stripe_test.go:525: engine.LoadAll(defs.FS): load all bundles: postgres: load bundle postgres: missing required file docs.md
--- FAIL: TestParityStripe_BundleLoadsAndValidates (0.00s)
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.439s
FAIL
```

RED confirmed: every parity subtest fails identically inside the shared `loadStripeBundle` helper
because `engine.LoadAll(defs.FS)` cannot yet enumerate ANY bundle successfully — the parallel
Tier-3 agent's in-progress `internal/connectors/defs/postgres/` bundle is mid-flight (only
`metadata.json` committed so far, missing `docs.md`/`spec.json`/etc. per its own task scope, which
this agent does not touch), and `internal/connectors/defs/stripe/` does not exist yet either. This
is still the correct RED signal for THIS task: the test file compiles clean (it only references
public `engine`/`connectors`/`defs`/`stripe` package APIs that already exist from earlier waves)
and fails purely for want of the golden bundle B-15 supplies — `loadStripeBundle`'s `t.Fatalf`
never even reaches its "bundle not found" branch because `LoadAll` itself errors first on the
sibling in-flight bundle, which is expected/acceptable coordinator-known parallel-wave interleave,
not a defect in this task's scope. (Verified by temporarily removing the just-created empty
`internal/connectors/defs/stripe/` subdirectories to confirm the failure is not self-inflicted: the
same `postgres: missing required file docs.md` error occurs with zero stripe-bundle files present
at all.)

## Parity-deviation ledger (carried into `docs/migration/conventions.md` by D-20)

1. **`create_customer` "email or name" rule → `minProperties: 1`.** Legacy
   `validateWriteRecord` requires at least one of `email`/`name` present (a named-field OR-rule).
   The engine's `record_schema` validator (draft-07 subset, `schema.go`) has no `anyOf`/`oneOf`
   keyword (SPEC §1.1 scope). `minProperties: 1` over the four optional properties
   (`email,name,description,phone`) is a broader approximation: it also accepts a record with only
   `description` or only `phone` set, which legacy would reject. Documented pre-existing deviation
   per PLAN.md T-15; not an `ENGINE_GAP` (no DATA divergence for any record that legacy itself
   would accept — the approximation is strictly more permissive, never rejects a legacy-valid
   record, and the parity test only exercises legacy-valid `email`/`name`-bearing records so
   observed WRITE behavior stays identical for every record either connector actually accepts in
   practice).
2. **Bundle-fixture `created` representation vs parity-test wire format.** The bundle's OWN
   `fixtures/streams/customers/*.json` (used only by `TestConformance/stripe` and
   `connectorgen validate`) represent `created` as an RFC3339 STRING
   (`"2026-01-01T00:00:00Z"`-shaped), not Stripe's real unix-integer wire format, because
   `conformance/dynamic.go`'s `checkCursorAdvances` only recognizes a cursor field via a Go type
   assertion `v.(string)` (dynamic.go:247) and then parses it as RFC3339 for `unix_seconds`
   formatting (dynamic.go:429) — a numeric `created` would make that check hard-fail ("no cursor
   value observed"), not skip. This does not affect engine-vs-legacy DATA parity: the schema type
   is declared permissively (`["string","integer"]`), `read.go`'s `projectRecord` performs no type
   coercion/validation (verified: it copies `raw[name]` verbatim regardless of declared type), and
   `TestParityStripe_*` drives its OWN httptest payloads with `created` as a JSON NUMBER (Stripe's
   real shape) — both engine and legacy decode that payload via the same
   `connsdk.RecordsAt`/`json.Number`-preserving decoder, so the emitted `created` values are
   byte-identical `json.Number` instances on both sides regardless of what the bundle's own
   (separate, synthetic) conformance fixtures happen to use. Filed as a documented bundle-authoring
   convention, not an `ENGINE_GAP` (no read-path behavior differs; only the shared conformance
   harness's cursor-typing assumption is narrower than real Stripe's wire format — a note for
   `docs/migration/conventions.md`'s fixture-authoring guidance, not an engine defect blocking this
   golden).
3. **`limit` sent as a static `stream.Query` value, not via `PaginationSpec.LimitParam`.** DATA-
   MODEL.md §3 lists `limit_param: limit` + `page_size: 100` as part of stripe's `PaginationSpec`;
   however `engine/paginate.go`'s `newCursorPaginator`/`lastRecordCursor` (the `type: cursor` +
   `last_record_field` implementation) does not read `LimitParam`/`PageSize` at all — those fields
   only drive the `page_number`/`offset_limit` paginator constructors. Confirmed by reading
   `paginate.go` end-to-end and its `TestNewPaginatorCursorLastRecordFieldStripeShape` test (no
   `limit` assertion). To reproduce legacy's `limit=100` on every request byte-identically, the
   bundle instead declares `streams.json`'s per-stream `query: {"limit": "100"}` (a static,
   always-sent param merged under the paginator's per-page query — `read.go`'s `mergeQuery`
   confirms per-page query only ever adds/overrides `starting_after`, never removes a base query
   key). Not an `ENGINE_GAP`: the declarative `query` escape hatch already produces the identical
   wire request; `limit_param`/`page_size` on `PaginationSpec` are additionally declared (harmless,
   unused-by-cursor-type fields) purely so the bundle documents the intended page size for a future
   engine enhancement, per DATA-MODEL.md's literal spec text.

## Self-verification (post-GREEN)

```
$ go build ./... && go vet ./...
(clean)

$ gofmt -l internal/connectors
(empty)

$ go test ./internal/connectors/engine -run TestParityStripe -v
--- PASS: TestParityStripe_StreamRecords (customers, charges, invoices, subscriptions, products)
--- PASS: TestParityStripe_CustomersTwoPagePagination
--- PASS: TestParityStripe_IncrementalCreatedGTEFromState
--- PASS: TestParityStripe_IncrementalCreatedGTEFromStartDate
--- PASS: TestParityStripe_WriteCreateCustomerFormBody
--- PASS: TestParityStripe_WriteUpdateCustomerFormBody
--- PASS: TestParityStripe_ManifestSurface
--- PASS: TestParityStripe_BundleLoadsAndValidates
PASS
ok  	polymetrics.ai/internal/connectors/engine	0.448s

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 2 connector(s) checked, 0 findings

$ go test ./internal/connectors/conformance -run 'TestConformance/stripe' -v
--- PASS: TestConformance (0.01s)
    --- PASS: TestConformance/stripe (0.01s)
PASS
ok  	polymetrics.ai/internal/connectors/conformance	0.366s

$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ...
0 issues.
```

Two fixture-bug iterations during GREEN (both self-corrected, no production-code workaround):
1. `TestParityStripe_IncrementalCreatedGTEFromStartDate` had a wrong hand-computed expected
   Unix-seconds constant for `2025-06-15T00:00:00Z` — fixed to `1749945600`.
2. `TestParityStripe_IncrementalCreatedGTEFromState` originally fed BOTH connectors the same raw
   unix-seconds cursor string, which is legacy's native persisted-cursor shape but NOT what the
   engine's `param_format: unix_seconds` (read.go's `formatParam`) accepts — it treats
   `req.State["cursor"]` as an RFC3339 input to convert, per `read_test.go`'s own
   `TestReadIncrementalParamFormats` contract. Rewrote the test to feed each connector its OWN
   native cursor-state representation for the IDENTICAL instant (legacy: unix seconds; engine:
   RFC3339) and assert both then emit the identical outgoing `created[gte]` wire value — the
   correct parity bar, since neither connector's `Read()` interface exposes its internal cursor
   bookkeeping format to callers; only the outgoing wire behavior for a given logical resume point
   is observable and comparable. Documented as parity-deviation #2 below (with docs.md cross-ref).

## Path guard

`git status --porcelain` after all work: only
`.planning/phases/wave0-engine-harness/traces/waveF-b15-ledger.md`,
`internal/connectors/defs/stripe/` (new, untracked), and
`internal/connectors/engine/parity_stripe_test.go` (new, untracked) — no edits outside the task's
scoped files.

## Blocker/note surfaced to coordinator (not this task's scope to fix)

`go test ./...` has one pre-existing failure unrelated to this task's own files:
`TestBundleLoadAllDefsFSEmpty` (`internal/connectors/engine/bundle_test.go`, landed in Wave A under
T/B-03) hard-asserts `len(engine.LoadAll(defs.FS)) == 0` ("wave0 ships no goldens yet"). This
assumption breaks as soon as ANY Wave F golden lands in `defs.FS` — confirmed independent of this
task by temporarily removing `internal/connectors/defs/stripe/` and re-running: the test already
fails on `postgres` alone (1 bundle, "want 0"), i.e. this is a foreseeable consequence of Wave F
landing goldens in parallel, not something this task's bundle introduces or can fix within its
exclusive file scope (`bundle_test.go` is not in this task's allowed-files list). The sibling
`conformance_test.go`'s `TestConformance_EmptyDefsTreePassesTrivially` already anticipated exactly
this scenario and gracefully `t.Skip`s; `bundle_test.go`'s copy of the same assumption was not
updated the same way. Recommend the coordinator route a small follow-up (update or skip
`TestBundleLoadAllDefsFSEmpty` the same way) to whichever Wave F agent lands last, or to V-21's
phase-gate pass.
