# S4 engine mini-wave ledger — 5 dialect gaps (fan-out, keyed-object, 0-indexed pages, OAuth2 extra_params, date-only lower bound)

Evidence base: `docs/migration/quarantine.json` + `docs/migration/status.json` `partial[]` blocker
reasons. Each item below hits >=3 connectors per the recurrence threshold (conventions.md §6).

Scope: `internal/connectors/engine/**` (+tests), `cmd/connectorgen/**` (corpus/validate coverage),
`docs/migration/conventions.md` §3. `connsdk`/`defs`/legacy connector packages are READ-ONLY.

## Item 1 — 0-indexed page starts (algolia/auth0/beamer/braze/clickup-api/concord/customerly/
dolibarr/harness/hubplanner + more)

**Evidence**: `PaginationSpec.StartPage` is a plain Go `int`; `engine/paginate.go`'s `newPaginator`
does `start := spec.StartPage; if start == 0 { start = 1 }` before constructing
`connsdk.PageNumberPaginator`, AND `connsdk.PageNumberPaginator.Start()` itself does
`p.page = p.StartPage; if p.page == 0 { p.page = 1 }` — a genuine `start_page: 0` cannot survive
either coercion, because JSON-unmarshaling an absent `start_page` key produces the exact same zero
value as an explicit `"start_page": 0`.

**Fix (no connsdk edit)**: change `PaginationSpec.StartPage` to `*int` (pointer distinguishes
unset from explicit 0). Add an engine-local `pageNumberPaginator` (mirrors
`connsdk.PageNumberPaginator`'s Start/Next shape exactly, honoring an explicit 0 start) used
whenever `newPaginator` builds a `page_number` paginator — replaces the direct
`connsdk.PageNumberPaginator` construction. Meta-schema: `start_page` stays `"type": "integer"`
(pointer-ness is a Go-side concern only; JSON representation unchanged, 0 is already schema-valid
integer). `connectorgen validate` needs no new rule (no field it currently checks changes shape
other than the Go type, which validate doesn't reflect on).

- RED: `TestNewPaginatorPageNumberExplicitZeroStartHonored` (engine/paginate_test.go) — start_page:0
  pointer, first request carries `page=0`.
- RED: `TestNewPaginatorPageNumberUnsetStartDefaultsToOne` (negative/regression) — nil StartPage
  behaves exactly as before (defaults to 1).
- RED: `TestNewPaginatorPageNumberExplicitOneStillOne` — explicit 1 unaffected.
- RED (loader): `TestLoadStreamsPageNumberStartPageZeroRoundTrips` (bundle_test.go) — streams.json
  `"start_page": 0` decodes to a non-nil `*int` pointing at 0, not nil.
- RED (loader, regression): `TestLoadStreamsPageNumberStartPageAbsentIsNilPointer` — absent
  start_page decodes to a nil pointer.
- GREEN: all 6 tests pass after (1) `PaginationSpec.StartPage int` -> `*int` (bundle.go), (2) a new
  engine-local `pageNumberPaginator` (paginate.go) replacing the direct
  `connsdk.PageNumberPaginator` construction in `newPaginator`'s `"page_number"` case, honoring an
  explicit 0 via `startPageOrDefault`. No meta-schema change needed (`start_page` was already
  `"type": "integer"`, and 0 is already schema-valid). No connectorgen validate change needed (no
  rule references PaginationSpec). `go build ./...` clean; `go test
  ./internal/connectors/engine ./cmd/connectorgen -count=1` green.

## Item 2 — Sub-resource fan-out (appfollow/bigmailer/breezy-hr/campayn/eventzilla/everhour/
finnworlds/k6-cloud/metricool/cisco-meraki/configcat/... 15+ connectors)

**Evidence**: every listed connector's real read is "list N parent ids (config CSV or one
preliminary paginated request), then issue the normal per-stream request sequence once per id,
stamping the parent id onto every child record" — no declarative primitive drives a per-id repeat
of the whole read sequence.

**Fix**: `StreamSpec` gains `FanOut *FanOutSpec`:
```json
"fan_out": {
  "ids_from": {"config_key": "app_collection_ids"} | {"request": {"path": "...", "records_path": "...", "id_field": "id"}},
  "into": {"query_param": "apps_id"} | {"path_var": "parent_id"},
  "stamp_field": "parent_id"
}
```
`read.go`'s `readDeclarative` resolves the id list ONCE per `Read` call (config CSV split on `,`,
trimming whitespace, dropping empty entries; OR one preliminary GET+paginate-to-exhaustion against
`ids_from.request`, extracting `id_field` off each record), then runs the EXISTING per-stream
request/pagination/incremental/filter/project/computed_fields/hook sequence UNCHANGED, once per id:
`into.query_param` adds the id as a query parameter on every page request of that sub-sequence;
`into.path_var` makes `{{ fanout.id }}` resolvable in `stream.Path` (a new Vars field, mirroring
`Cursor`). `stamp_field`, when set, is written onto every projected record's field via the existing
`applyComputedFields` write path (post-projection) for that id's sub-sequence — implemented as an
engine-added computed field the caller doesn't have to declare twice. Pagination/incremental/
MaxPages/rate-limit all continue to apply per-id-sequence, unchanged (each id's sub-sequence is a
fresh independent paginator + fresh baseQuery). Meta-schema: new `fan_out` object on both
`base`/stream pagination-adjacent level (stream only — fan-out is inherently stream-scoped, not
declared at `base`). `connectorgen validate`: `ids_from.request.path`/`records_path`/`id_field`
templates get the same `ResolveCheck` treatment stream.Path/Query already receive.

- RED: `TestReadFanOutConfigKeyRunsOncePerID` — CSV config_key, query_param `into`, stamp_field set,
  asserts N request sequences + stamped field per record.
- RED: `TestReadFanOutConfigKeyPathVarInterpolatesFanoutID` — `into.path_var`, `{{ fanout.id }}` in
  stream.Path.
- RED: `TestReadFanOutRequestIDsPreliminaryPaginatedRequest` — `ids_from.request`, multi-page
  preliminary listing exhausted before any per-id sequence starts.
- RED: `TestReadFanOutEachIDSequenceOwnPagination` — per-id pagination advances/terminates
  independently (2 pages for id A, 1 page for id B).
- RED (negative): `TestReadFanOutEmptyIDListEmitsNothing` — CSV config_key resolves to "", zero
  requests issued, no error.
- RED (negative): `TestReadFanOutMissingIDsFromIsReadError` — fan_out declared with neither
  config_key nor request populated.
- RED (negative): `TestReadFanOutBothIDsFromFormsIsReadError` — both forms declared together.
- RED (positive, regression): `TestReadFanOutConfigKeyIDsAreTrimmed` — CSV whitespace/empty-entry
  handling.
- RED (validate): `TestResolveCheckAcceptsFanoutIDReference` (interpolate_test.go) — `fanout.id` is
  a known engine pseudo-namespace, `fanout.idx` (typo) still errors.
- RED (validate corpus): `fanout-request-path-unknown-spec-key` (cmd/connectorgen/testdata/invalid)
  — an undeclared spec key inside `fan_out.ids_from.request.path` must be caught, same as an
  ordinary stream.Path.
- RED (validate corpus, positive control): `fanout-valid` (testdata/valid-extra) — a well-formed
  fan_out block (config_key + path_var + stamp_field, `{{ fanout.id }}` in stream.Path) passes
  cleanly.
- GREEN: `StreamSpec.FanOut *FanOutSpec` (bundle.go: `FanOutSpec`/`FanOutIDsFrom`/
  `FanOutIDsRequest`/`FanOutInto`); meta-schema `fan_out` object added to
  `streams.schema.json`'s stream properties (additionalProperties:false at every level, `required:
  ["ids_from","into"]`); `read.go`'s `readDeclarative` branches on `stream.FanOut != nil` ->
  `readFanOut` (resolves ids once via `resolveFanOutIDs`, then calls the extracted
  `readOneSequence` once per id with a `fanoutContext{id, queryParam, stampField}`) — the
  non-fan-out path calls the SAME `readOneSequence` with a zero-valued `fanoutContext{}`, so
  ordinary (non-fan-out) streams are byte-for-byte unchanged. `interpolate.go` gained
  `Vars.FanoutID` + a `fanout.id` reference namespace (`resolveFanoutRef`) — HARD ERRORS on empty
  (unlike `incremental.lower_bound`, which is deliberately absent-tolerant), since `{{ fanout.id }}`
  only ever appears inside a fan_out-declared path and a missing id there is a real bug, not a
  legitimate absence. `checkNamespaceRef` (interpolate.go) extended with `knownFanoutKeys` so
  `ResolveCheck`/`connectorgen validate` accept `fanout.id` statically; `cmd/connectorgen/
  validate.go`'s `checkInterpolations` extended to walk `fan_out.ids_from.request.path` the same
  way it already walks `stream.Path`. `go build ./... && go vet ./...` clean; `go test
  ./internal/connectors/engine ./cmd/connectorgen -count=1` green; `go run ./cmd/connectorgen
  validate internal/connectors/defs` still 0 findings (411 connectors).

## Item 3 — Keyed-object flatten (appfigures/alpha-vantage/exchange-rates symbols/gutendex-adjacent)

**Evidence**: response body at `records.path` is a JSON OBJECT keyed by an arbitrary id
(`{"111": {...}, "222": {...}}`), not an array — `connsdk.RecordsAt` today only turns a bare object
into ONE record (the whole object), never explodes its values.

**Fix**: `RecordsSpec` gains `KeyedObject bool` + `KeyField string`. An engine-local
`recordsAtKeyed(body, path, keyField) ([]connsdk.Record, error)` selects the object at path (reusing
connsdk's decode+selectPath indirectly via a small package-local reimplementation since connsdk
itself is read-only) and, when `KeyedObject` is true, treats EACH VALUE (must itself decode as a
JSON object; non-object values are skipped, matching `RecordsAt`'s existing array-element tolerance)
as one record; when `KeyField` is set, the map key is stamped onto that field before projection (so
it participates in schema projection/computed_fields like any other raw field). `read.go`'s
`readDeclarative`/pagination-records-extraction call site branches on `stream.Records.KeyedObject`.
Sort order: map iteration in Go is random — records are emitted in ascending-key sorted order for
determinism (parity/test stability), not raw map order.

- RED: `TestRecordsAtKeyedObjectFlattensEachValue` — 3-key object -> 3 records, each value's own
  fields intact.
- RED: `TestRecordsAtKeyedObjectStampsKeyField` — key_field set, each record carries its source key.
- RED: `TestRecordsAtKeyedObjectSortedByKeyForDeterminism` — emission order is sorted ascending.
- RED (negative): `TestRecordsAtKeyedObjectSkipsNonObjectValues` — a value that is a scalar/array is
  dropped, not a crash/error.
- RED (negative): `TestRecordsAtKeyedObjectEmptyObjectYieldsNoRecords`.
- RED (integration): `TestReadKeyedObjectStreamEmitsFlattenedProjectedRecords` (read.go path, full
  Read() call against an httptest fixture).
- RED (loader): `TestLoadStreamsRecordsKeyedObjectRoundTrips` / `...WithoutKeyedObjectDefaultsFalse`
  (bundle_test.go).
- RED (validate corpus, positive control): `keyed-object-valid` (testdata/valid-extra).
- GREEN: `RecordsSpec` gained `KeyedObject bool`/`KeyField string` (bundle.go); meta-schema already
  had `keyed_object`/`key_field` on `records` (added alongside item 2's edit to
  `streams.schema.json`, same file). New engine-local `recordsAtKeyed` (read.go) — sorted-by-key
  deterministic flatten, non-object values skipped, empty object -> zero records — reusing
  engine-local `decodeJSONKeyed`/`selectPathKeyed` duplicates of connsdk's unexported decode/
  selectPath helpers (connsdk read-only). `read.go`'s per-page record extraction now calls
  `extractRecords(resp.Body, stream.Records)` (branches on `KeyedObject`) instead of calling
  `connsdk.RecordsAt` directly — the fan-out id-listing request (item 2) intentionally still calls
  `connsdk.RecordsAt` directly (keyed-object id-listing is out of scope; no evidence connector needs
  it). No connectorgen validate rule needed (no template field on RecordsSpec to statically check);
  added a positive-control corpus case instead. `go build ./... && go vet ./...` clean; `go test
  ./internal/connectors/engine ./cmd/connectorgen -count=1` green; validate CLI still 0 findings
  (411 connectors).

## Item 4 — OAuth2 extra params (auth0 `audience`, box `box_subject_type`/`box_subject_id`-adjacent)

**Evidence**: `connsdk.OAuth2ClientCredentials` ALREADY has an `ExtraParams url.Values` field (not a
connsdk gap at all) — the gap is purely that `engine.AuthSpec`'s `oauth2_client_credentials` mode has
no field to populate it from, so `buildOAuth2ClientCredentials` (engine/auth.go) never wires it.

**Fix**: `AuthSpec` gains `ExtraParams map[string]string` (templated values, resolved the same way
every other AuthSpec field is — via `Interpolate(v, vars)`, hard error on unresolved key exactly like
`ClientID`/`ClientSecret`). `buildOAuth2ClientCredentials` resolves each entry and sets
`connsdk.OAuth2ClientCredentials.ExtraParams` as a `url.Values` (one value per key; templated values
are singular strings, not repeated-key lists — no bundle needs multi-value form params here).
Meta-schema: `auth[].extra_params` object of string->string. `ResolveCheckAuthSpec` extended to
static-validate every `extra_params` value template.

- RED: `TestBuildOAuth2ClientCredentialsExtraParamsWiredIntoTokenRequest` — auth.go unit test
  asserting the constructed `*connsdk.OAuth2ClientCredentials.ExtraParams` carries the resolved
  `audience` value.
- RED: `TestBuildOAuth2ClientCredentialsExtraParamsTemplatedFromConfig` — value templated via
  `{{ config.base_url }}/api/v2/`-shaped derivation.
- RED (negative): `TestBuildOAuth2ClientCredentialsExtraParamsUnresolvedKeyErrors` — a
  extra_params value referencing an undeclared config key hard-errors, doesn't silently drop.
- RED (validate): `TestResolveCheckAuthSpecValidatesExtraParamsTemplates` (interpolate_test.go) —
  ResolveCheckAuthSpec surfaces a bad extra_params template.
- RED (integration): `TestReadOAuth2ClientCredentialsExtraParamsSentOnTokenRequest` — read_test.go,
  full token-exchange round trip against an httptest token endpoint, asserts the form-encoded
  request body carries `audience=...` (lives in auth_test.go alongside the other 3 RED tests above).
- RED (validate): `TestResolveCheckAuthFieldsValidatesAllTemplatedFields`'s two new subtests
  (interpolate_test.go) — extra_params values get the same static ResolveCheck coverage as
  token_url/client_id/client_secret/scopes.
- RED (validate corpus): `oauth2-extra-params-unknown-spec-key` (testdata/invalid) + positive
  control `oauth2-extra-params-valid` (testdata/valid-extra).
- GREEN: `AuthSpec.ExtraParams map[string]string` (bundle.go); meta-schema `auth[].extra_params`
  object added (streams.schema.json). `buildOAuth2ClientCredentials` (auth.go) resolves each entry
  via `resolveExtraParams` (hard errors on an unresolved config/secrets key — deliberately NOT
  given the omit_when_absent/default tolerance stream.Query has, since a misconfigured
  audience/subject param should fail loudly like ClientID/ClientSecret do) and sets the result on
  `connsdk.OAuth2ClientCredentials.ExtraParams` (a field connsdk ALREADY had — confirmed no connsdk
  edit was needed, the gap was purely AuthSpec having nothing to populate it from).
  `ResolveCheckAuthSpec` (interpolate.go) extended to validate every `extra_params` value template
  (sorted-key iteration for deterministic error messages) — this flows into `connectorgen validate`
  for free via the existing `engine.ResolveCheckAuthSpec(a, specKeys)` call in
  `checkInterpolations` (cmd/connectorgen/validate.go), no validate.go changes needed. `go build
  ./... && go vet ./...` clean; `go test ./internal/connectors/engine ./cmd/connectorgen -count=1`
  green; validate CLI still 0 findings (411 connectors).

## Item 5 — Date-only lower bounds (marketstack + any date param_format connector sending a
no-colon offset or a bare YYYY-MM-DD state cursor)

**Evidence**: `parseLowerBoundTime` (engine/read.go) accepts ONLY all-digits (Unix seconds) or strict
`time.RFC3339`. Marketstack's real wire cursor value for its `date` param_format streams is a bare
`YYYY-MM-DD` string with no time/offset component at all — neither digits-only nor RFC3339 parse it.

**Fix**: extend `parseLowerBoundTime` to also accept `"2006-01-02"` (`time.DateOnly`-shaped) as a
third input shape, tried after digits-only and RFC3339 both fail (order: digits -> RFC3339 ->
date-only, so an ambiguous value is never misclassified — a valid RFC3339 string is never
all-digits, and a valid date-only string is never RFC3339-parseable, so no shape masks another).
This applies uniformly across every `param_format` that calls `parseLowerBoundTime`
(`unix_seconds`/`date`/`github_date_range`) per the task's explicit instruction, not just `date`.

- RED: `TestParseLowerBoundTimeAcceptsDateOnly` — direct unit test, `"2026-01-02"` parses to
  midnight UTC that date.
- RED: `TestFormatParamDateAcceptsDateOnlyLowerBound` — `formatParam(v, "date")` round-trips a
  date-only input back to the same date-only string.
- RED: `TestFormatParamUnixSecondsAcceptsDateOnlyLowerBound` — date-only input converts correctly to
  Unix seconds via `param_format: unix_seconds`.
- RED: `TestFormatParamGithubDateRangeAcceptsDateOnlyLowerBound` — date-only input formats as
  `>=2026-01-02T00:00:00Z`.
- RED (negative): `TestParseLowerBoundTimeRejectsMalformedDateOnly` — `"2026-13-40"`,
  `"2026/01/02"`, `"not-a-date"` still hard-error (not silently truncated/misparsed). This one
  already passed pre-fix (baseline confirmed unaffected); the other 4 tests in this item were
  confirmed RED before the fix.
- GREEN: `parseLowerBoundTime` (read.go) now tries THREE shapes in order — all-digits (Unix
  seconds) -> strict RFC3339 -> bare `YYYY-MM-DD` (new `dateOnlyLayout` const, parsed as midnight
  UTC that date) — applied uniformly across every `param_format` that calls it
  (`unix_seconds`/`date`/`github_date_range`), per the task's explicit instruction, not just
  `date`. No ordering ambiguity: a valid RFC3339 string always contains a `T` separator (never
  all-digits, never date-only-parseable), and a valid date-only string is never RFC3339-parseable.
  No meta-schema/connectorgen validate change needed (parseLowerBoundTime is a runtime value-parsing
  function, not a declarative-field shape connectorgen statically checks). `go build ./...
  && go vet ./...` clean; `go test ./internal/connectors/engine ./cmd/connectorgen
  ./internal/connectors/conformance -count=1` all green.

## Self-verify (run after all 5 items green)

```
go build ./... && go vet ./...
go test ./internal/connectors/engine ./cmd/connectorgen ./internal/connectors/conformance -count=1
make lint
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results recorded at the bottom of this ledger once run.

## Self-verify results (2026-07-03)

- `go build ./... && go vet ./...` — clean.
- `go test ./internal/connectors/engine ./cmd/connectorgen ./internal/connectors/conformance -count=1`
  — all ok (engine 2.4s, connectorgen 1.3s, conformance 4.6s). One spurious engine failure was
  observed when `go vet` and `go test` ran CONCURRENTLY against a cold build cache
  (`TestResolveCheckAcceptsFanoutIDReference` reported the pre-item-2 "unknown namespace fanout"
  error); sequential re-runs and a `-count=20` stress of `TestResolveCheck*` all pass — build-cache
  race on the runner, not a code defect.
- `make lint` — 0 issues.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — 411 connector(s) checked,
  0 findings.

## Item 1 aftermath — parity coverage for pre-existing `start_page: 0` bundles (2026-07-03)

Three already-migrated bundles declare `"start_page": 0`: justcall, mailosaur, nasa (all four
legacy counterparts, navan included, harvest with `for page := 0`). Coverage audit against the
replay harness (conformance/replay.go matches recorded method+path+QUERY per fixture page):

- **mailosaur** (`messages`) and **nasa** (`neo_browse`): fixtures already record
  `query: {"page": "0"}` on page_1 and `{"page": "1"}` on page_2 — `TestConformance/mailosaur`
  and `TestConformance/nasa` are end-to-end proof the engine now sends the literal page 0 first.
- **justcall** (`calls`/`sms`/`users`): fixtures recorded NO `page` query key at all, so replay
  matched on path alone — the pre-fix engine's wrong first request (`page=1`) sailed through
  green. This is the wave2-flagged missed defect (navan ENGINE_GAP report). FIXED in this pass:
  all 6 fixture pages now pin the literal page value (`page=0` on page_1, `page=1` on page_2).
  A regressed engine that coerces the first request back to 1 would match page_2's fixture
  first, leave page_1 unserved, and fail the pagination_terminates check.
  `TestConformance/justcall` green with the pins.

navan itself stays quarantined per process (wave3-roster.json `quarantine_engine_gap`, 47
connectors) — item 1 closes its blocker, so the wave3 batch re-attempts it with `start_page: 0`
plus the same fixture-level page-value pins.
