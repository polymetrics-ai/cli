# PLAN — wave0-engine-harness

Ordered TDD ledger. Every behavior task `B-N` has a paired test task `T-N`; the test is written and
committed RED (failing) before `B-N` production code. Docs-only tasks are tagged `docs-only` and
need no test pair. Each task is sized for ONE Sonnet backend/tester agent run.

Executor model: `sonnet` (gsd-loop-backend / gsd-loop-tester). Skills base set for all engine/
certify/core tasks (per `docs/prompts/universal-programming-loop-prompts.md`): `golang-project-layout`,
`golang-code-style`, `golang-structs-interfaces`, `golang-error-handling`, `golang-naming`,
`golang-testing`, `golang-safety`, `golang-lint`; add `golang-design-patterns` where noted;
tester adds `golang-benchmark` only if a perf-sensitive path is touched (none expected).

## Rules (all tasks)

- FORBIDDEN: new go.mod dependencies; edits to `internal/app/**`, `internal/cli/**` (wave0 has no
  CLI change), `registryset/registry_gen.go` (regen only via tool), `catalog_data.json`,
  `icon_data.json`, legacy connector packages (except NONE — goldens' legacy dirs are read-only
  references), any file outside the task's listed files. Exception: task B-16 edits
  `cmd/registrygen/main.go` skip map (orchestrator-approved single legacy edit).
- A needed engine feature or dependency = typed blocker (`ENGINE_GAP`/`NEEDS_NEW_DEP`) reported to
  the coordinator; never a workaround.
- Stop conditions: missing required context · verification cannot run · human gate reached ·
  same failure repeats without new evidence.
- Every task self-verifies with its listed command AND `go build ./... && go vet ./...` before
  reporting done. Record RED evidence in `.planning/phases/wave0-engine-harness/TDD-LEDGER.md`.

---

## Wave A — engine foundations (4 tasks, fully parallel) `wave:A`

### T-01 / B-01 — draft-07 mini validator `wave:A`
- Files: `internal/connectors/engine/schema.go`, `internal/connectors/engine/schema_test.go`
- T-01 (test, RED first): table-driven compile+validate cases for every supported keyword
  (`type` incl. `["string","null"]`, `required`, `properties`, `items`, `enum`, `pattern`,
  `minProperties`, `additionalProperties:false`), annotation passthrough (`format`, `default`),
  `x-secret`→`SecretKeys()`, `x-primary-key`/`x-cursor-field` accessors, unknown-keyword compile
  error, invalid-instance error messages carrying JSON pointer-ish paths.
- B-01 (behavior): implement compiler/validator per `SPEC.md` §1.1. No reflection tricks; compile
  once, validate many.
- Acceptance: all T-01 cases pass; validator has zero imports outside stdlib.
- Verify: `go test ./internal/connectors/engine -run TestSchema -v`

### T-02 / B-02 — interpolator + when-conditions `wave:A`
- Files: `internal/connectors/engine/interpolate.go`, `interpolate_test.go`
- T-02 (RED): cases for `config./secrets./record./cursor` resolution, dotted record paths
  (`record.user.login`), filters `urlencode` (default applied to path-segment insertions —
  include injection cases: config value `a/../b`, `a?x=1`, `a b`, `%2e%2e`), `unix_seconds`,
  `base64`; CRLF/header-injection rejection (resolved values destined for headers or paths
  containing `\r` or `\n` → error, per THREAT-MODEL §2); unresolved-key error (names the key +
  source); `when` grammar `==` / `in [...]` / truthiness; type-check API
  `ResolveCheck(template, specKeys)` used later by connectorgen.
- B-02: implement per design §B.3 (~150 lines target).
- Verify: `go test ./internal/connectors/engine -run TestInterpolate -v`

### T-03 / B-03 — bundle types + loader + defs scaffold `wave:A`
- Files: `internal/connectors/engine/bundle.go`, `bundle_test.go`,
  `internal/connectors/defs/defs.go`, `internal/connectors/engine/schema/*.schema.json` (5 meta-schemas),
  loader testdata under `internal/connectors/engine/testdata/bundles/`
- T-03 (RED): `LoadAll`/`Load` over an in-test `fstest.MapFS`: happy path (full bundle), optional
  files absent (`writes.json`, `fixtures/`), `streams.json` optional iff
  `capabilities.dynamic_schema`, dir-name/name mismatch error, bad name regex error, missing
  required file error, meta-schema violation error (uses B-01 validator once available — this test
  is authored RED against the loader API regardless).
- B-03: types per design §B.2 (exact fields in `API-CONTRACT.md`), loader, `defs.go` with
  `//go:embed all:*`, meta-schema JSON files (contracts in `DATA-MODEL.md`).
- Depends: B-01 for meta-schema enforcement (loader wiring lands after; parse/layout checks are
  self-contained). Coordinator may sequence B-03 last within Wave A.
- Verify: `go test ./internal/connectors/engine -run TestBundle -v`

### T-04 / B-04 — typed errors + error_map `wave:A`
- Files: `internal/connectors/engine/errors.go`, `errors_test.go`
- T-04 (RED): `engine.Error` wraps `*connsdk.HTTPError` (`errors.As` reachable); context fields
  `{connector, stream|action, page|record_index}`; `error_map` matching by `status` +
  `match_body` substring; `class`/`hint` attach; hint surfaces verbatim in `Error()`; secrets
  never in message (feed a body containing a token; assert `safety.RedactErrorText` applied).
- B-04: implement per design §F.4.
- Verify: `go test ./internal/connectors/engine -run TestError -v`

## Wave B — auth, pagination, hooks (3 tasks, parallel; after Wave A) `wave:B`

### T-05 / B-05 — auth selection `wave:B`
- Files: `internal/connectors/engine/auth.go`, `auth_test.go`
- T-05 (RED): full mode matrix → the existing constructors in
  `internal/connectors/connsdk/auth.go` (`Bearer`:41, `Basic`:52, `APIKeyHeader`:47,
  `APIKeyQuery`:71, `OAuth2ClientCredentials`:77); `when` ordering (first match wins; github-style
  `auto|token|public` table from design §A streams.json); no-match error; `custom` mode resolves
  `AuthHook` via registry (fake hook), missing hook error; secrets flow only into authenticators.
- B-05: `selectAuth` per SPEC §1.1. Depends: T/B-02 (when-conditions), T/B-09 registry stub — use a
  local registration seam if B-09 not merged; coordinator orders B-09 first inside Wave B if same agent.
- Verify: `go test ./internal/connectors/engine -run TestAuth -v`

### T-06 / B-06 — paginator construction (all 6 types) `wave:B`
- Files: `internal/connectors/engine/paginate.go`, `paginate_test.go`
- T-06 (RED): for each of `link_header`, `page_number`, `offset_limit`, `cursor(token_path)`,
  `cursor(last_record_field+stop_path)` (stripe `starting_after`/`has_more`, replacing
  `internal/connectors/stripe/stripe.go:147` harvest), `next_url`, `none`: drive against
  multi-page `httptest.Server` fixtures and assert TERMINATION (bounded page count), exact page
  sequence, and no duplicate page fetches; malformed spec errors (unknown type, cursor with both
  token sources); `next_url` SSRF guard (THREAT-MODEL §3): a next-page URL whose host differs from
  the base URL host → error unless the PaginationSpec sets `allow_cross_host: true` (field added
  to PaginationSpec; document in DATA-MODEL.md).
- B-06: `newPaginator` mapping to `connsdk/paginate.go` types (`OffsetPaginator`:26,
  `PageNumberPaginator`:58, `CursorPaginator`:95, `LinkHeaderPaginator`:121) + two new engine-local
  `connsdk.Paginator` impls (`lastRecordCursor`, `nextURL`). connsdk is NOT modified.
- Verify: `go test ./internal/connectors/engine -run TestPaginat -v`

### T-07 / B-07 — hook interfaces + registry `wave:B`
- Files: `internal/connectors/engine/hooks.go`, `hooks_test.go`,
  `internal/connectors/hooks/hookset/hookset_gen.go` (placeholder header, regenerated by B-11)
- T-07 (RED): `RegisterHooks`/`HooksFor` round-trip, duplicate-name overwrite semantics documented,
  each of the 5 hook interfaces dispatchable via fakes, `HooksFor` unknown → nil,nil-safe.
- B-07: interfaces exactly per design §B.7 signatures (see `API-CONTRACT.md`).
- Verify: `go test ./internal/connectors/engine -run TestHooks -v`

## Wave C — read/write execution (2 tasks, parallel; after Wave B) `wave:C`

### T-08 / B-08 — read path `wave:C`
- Files: `internal/connectors/engine/read.go`, `read_test.go`
- T-08 (RED): against `httptest.Server`: initial-query build (static `query` + incremental lower
  bound from state cursor, fallback `start_config_key`; each `param_format`); records extraction
  path incl. `"."` and `single_object`; `filter.field_absent` (github issues vs PRs) and
  `field_equals`; projection `schema` (only declared properties emitted) vs `passthrough`;
  `computed_fields` nested extraction (`{{ record.user.login }}`); cursor advance == max cursor
  seen and resume re-read sends `request_param`; `client_filtered` drops old records; empty-header
  omission (conditional `Stripe-Account`); rate-limit wait invoked N-1 times with injected sleeper;
  RecordHook + StreamHook + CheckHook dispatch; error_map application on 401/403 fixtures.
- B-08: implement per SPEC §1.1 read.go + generic `InitialState`.
- Verify: `go test ./internal/connectors/engine -run TestRead -v`

### T-09 / B-09 — write path `wave:C`
- Files: `internal/connectors/engine/write.go`, `write_test.go`
- T-09 (RED): body construction default (record minus `path_fields`) for `json`; `form`
  (stripe-shape, via `Requester.DoForm` — compare `internal/connectors/stripe/write.go:134`
  `customerForm`); `none` + `body_fields` for delete-with-body; `record_schema` validation errors
  carry record index; `DryRunWrite` preview shows resolved method/path with secrets redacted;
  `kind:delete` + `missing_ok_status:[404]` → written-not-failed; non-listed 404 → failed;
  WriteHook handled/handoff; per-record loop stops on ctx cancel with correct
  `RecordsWritten/RecordsFailed` accounting (match legacy semantics in `stripe/write.go:66`).
- B-09: implement per design §B.5.
- Verify: `go test ./internal/connectors/engine -run TestWrite -v`

## Wave D — connector assembly + connectorgen + certify base (3 tasks, parallel; after Wave C) `wave:D`

### T-10 / B-10 — engine.Connector + engine.Base + Definition `wave:D`
- Files: `internal/connectors/engine/connector.go`, `connector_test.go`,
  `internal/connectors/definition.go` (new file in package connectors), `definition_test.go`
- T-10 (RED): `engine.New(bundle, hooks)` satisfies `connectors.Connector` (compile-time asserts),
  `WriteValidator`, `DryRunWriter`, `StatefulReader`, `ManifestProvider`, `DefinitionProvider`;
  `Metadata()`/`Manifest()` synthesized from bundle matches legacy shape (spot fields);
  derived sync modes: PK-only stream → dedup modes, incremental block → incremental modes, both,
  neither (§B.6 truth table); `Write` without `writes.json` → `connectors.ErrUnsupportedOperation`;
  `engine.Base` serves `Definition()` for a Tier-3 fake.
- B-10: implement; `Definition/StreamSummary/WriteActionInfo/DefinitionProvider` added WITHOUT
  touching the existing `Connector` interface or `manifest.go` (signatures in `API-CONTRACT.md`).
- Skills add: `golang-structs-interfaces`, `golang-design-patterns`.
- Verify: `go test ./internal/connectors/engine -run TestConnector -v && go test ./internal/connectors -run TestDefinition -v`

### T-11 / B-11 — cmd/connectorgen (validate | gen | new) `wave:D`
- Files: `cmd/connectorgen/main.go` (+ `validate.go`, `gen.go`, `new.go`, `main_test.go`,
  `testdata/` seeded-invalid bundles), `internal/connectors/native/nativeset/nativeset_gen.go`
  (generated output committed)
- T-11 (RED): validate accepts a known-good testdata bundle; REJECTS each seeded defect class
  (≥10 seeded, ≥8 distinct classes — the list in `EVAL-PLAN.md` §3) with a finding naming the
  file+rule; `--json` findings shape; `gen` writes deterministic hookset/nativeset files
  (byte-stable on re-run); `new x` scaffold passes `validate`.
- B-11: implement subcommands; validation composes engine loader + `ResolveCheck` + meta-schemas.
- Verify: `go test ./cmd/connectorgen -v && go run ./cmd/connectorgen validate internal/connectors/defs`
  (defs may be empty of bundles until Wave F; empty tree = pass with count 0)
- Note: seeded-invalid corpus is shared in spirit with conformance self-tests (B-13) — B-11 owns
  `cmd/connectorgen/testdata/invalid/`, B-13 has its own corpus under
  `internal/connectors/conformance/testdata/invalid/`; no cross-package sharing.

### T-12 / B-12 — certify report + cliharness `wave:D`
- Files: `internal/connectors/certify/report.go`, `cliharness.go`, `certify.go`,
  `report_test.go`, `cliharness_test.go`
- T-12 (RED): report marshal/unmarshal round-trip vs the design §A JSON shape; history append;
  harness runs `cli.Run(["init","--root",tmp,"--json"], …)` and asserts kind/exit; envelope kind
  mismatch → typed failure; secret-scan detects planted secret in exact/base64/URL-encoded forms;
  argv redaction in `stages[].cli.argv_redacted`.
- B-12: implement per SPEC §1.6 (no CLI wiring, no write/flow/schedule code).
- Depends: nothing in engine — may be dispatched as early as Wave B if capacity allows.
- Verify: `go test ./internal/connectors/certify -run 'TestReport|TestHarness' -v`

## Wave E — conformance v2 + certify source stages (2 tasks, parallel; after Wave D) `wave:E`

### T-13 / B-13 — conformance v2 `wave:E`
- Files: `internal/connectors/conformance/conformance.go`, `static.go`, `dynamic.go`,
  `replay.go` (fixture httptest server), `conformance_test.go`, `testdata/invalid/**`
- T-13 (RED): static checks each fail on a targeted invalid corpus bundle and pass on a good one;
  dynamic checks against an in-test bundle + fixtures: `pagination_terminates` (2-page fixture,
  each page served exactly once), `records_match_schema` failure on seeded type drift,
  `cursor_advances` incl. re-read param assertion, `write_request_shape` vs `expect` block,
  `delete_semantics`; `TestConformance` iterates `defs.FS` (0 bundles = pass; goldens auto-join
  in Wave F).
- B-13: implement; legacy `native_conformance.go` untouched.
- Verify: `go test ./internal/connectors/conformance -v`

### T-14 / B-14 — certify source stages vs sample `wave:E`
- Files: `internal/connectors/certify/stages_source.go`, `stages_source_test.go`
- T-14 (RED): full stage list 0–11 against `sample` in an ephemeral root (mirror Makefile `smoke`
  recipe flags, `Makefile:41`): preflight, fixture_conformance=skipped(no bundle), manual_json,
  credentials add/test (env-injected token), catalog, full_refresh_append (records>0),
  overwrite/overwrite_deduped/incremental_append_deduped via `file`-connector capture replay
  (PK dedup asserted via `pm query`), incremental_append + resume (cursor monotonic, run2 ≤ run1),
  query_contract; report `data_source: live|capture` per mode; final report `passed=true`; a
  sabotaged stage (wrong expected kind) yields `passed=false` with the failing stage named.
- B-14: implement stages; discover exact CLI flags from `internal/cli/cli.go` + docs; any missing
  CLI capability = typed blocker (do NOT edit cli).
- Verify: `go test ./internal/connectors/certify -run TestSourceStages -v`

## Wave F — golden migrations + parity (3 tasks, parallel; after Wave E) `wave:F`

### T-15 / B-15 — golden: stripe bundle + parity `wave:F`
- Files: `internal/connectors/defs/stripe/**` (metadata.json, spec.json, streams.json, writes.json,
  api_surface.json, schemas/{customers,charges,invoices,subscriptions,products}.json,
  fixtures/streams/** incl. a 2-page customers fixture, fixtures/writes/{create_customer,update_customer}.json,
  docs.md), `internal/connectors/engine/parity_stripe_test.go`
- T-15 (RED): parity test per SPEC §1.9 committed first (fails until bundle exists): same records
  engine-vs-legacy (`internal/connectors/stripe/stripe.go` `Connector{Client}` + `base_url`) for
  identical httptest input across all 5 streams incl. 2-page pagination + incremental
  `created[gte]` propagation; write parity: method/path/form-body equality vs
  `stripe/write.go`; manifest-surface equality vs `connectors.ManifestOf` (stream names, PKs,
  cursor fields, write action names). Known deviation allowed + documented in conventions.md:
  create_customer "email or name" legacy rule approximated by `minProperties: 1`.
- B-15: author the bundle; `connectorgen validate` + `TestConformance/stripe` must pass.
- Skills: `golang-code-style`, `golang-naming`, `golang-error-handling`, `golang-testing`.
- Verify: `go run ./cmd/connectorgen validate internal/connectors/defs && go test ./internal/connectors/engine -run TestParityStripe -v && go test ./internal/connectors/conformance -run 'TestConformance/stripe' -v`

### T-16 / B-16 — golden: searxng bundle + parity; registrygen skip-map `wave:F`
- Files: `internal/connectors/defs/searxng/**`, `internal/connectors/engine/parity_searxng_test.go`,
  `cmd/registrygen/main.go` (skip-map entries ONLY: `defs`, `engine`, `hooks`, `native`,
  `conformance`, `certify`) + regenerated `internal/connectors/registryset/registry_gen.go`
  (via `go run ./cmd/registrygen`, must be byte-identical to current — the skip dirs contain no
  connector packages yet at scan level; if not identical, STOP and escalate)
- T-16 (RED): parity for `search` + `reddit` streams (templated `q`, `pageno` pagination,
  `max_pages: 1` stop, no size param — legacy behavior in `internal/connectors/searxng/searxng.go:102`);
  manifest-surface equality; regression test asserting `registryset` still registers legacy
  searxng (`RegisterNativeLive` path, `searxng.go:43`).
- B-16: author bundle + skip-map edit + regen.
- Verify: `go run ./cmd/connectorgen validate internal/connectors/defs && go run ./cmd/registrygen && git diff --exit-code internal/connectors/registryset/ && go test ./internal/connectors/engine -run TestParitySearxng -v`

### T-17 / B-17 — golden: postgres Tier-3 native + bundle + parity `wave:F`
- Files: `internal/connectors/native/postgres/{connector.go,connection.go,reader.go,cataloger.go,cdc.go}`
  (+ `postgres_test.go`), `internal/connectors/defs/postgres/**` (metadata.json with
  `dynamic_schema: true`, spec.json with `password` x-secret, api_surface.json (db surface:
  excluded-category `non_data_endpoint` prose or minimal), docs.md; NO streams.json),
  `internal/connectors/native/postgres/parity_test.go`
- T-17 (RED): fixture-mode parity vs legacy `internal/connectors/postgres/postgres.go` (Check,
  Catalog stream set, Read record equality); config-validation error table parity
  (host/port/sslmode rules, `resolveConfig` at `postgres.go:119`); `Definition()` served from
  bundle via `engine.Base`; NO init() registration (grep-guard test asserts the package does not
  call `RegisterFactory`); CDC remains documented stub.
- B-17: port logic from legacy package (pgx dep already in go.mod); component split per design
  §B.7 Tier-3; legacy package untouched.
- Verify: `go test ./internal/connectors/native/postgres -v && go run ./cmd/connectorgen validate internal/connectors/defs`

## Wave G — tooling, docs, inventory (3 tasks, parallel; after Wave D; G3 anytime) `wave:G`

### T-18 / B-18 — .golangci.yml + Makefile gates `wave:G`
- Files: `.golangci.yml`, `Makefile` (add `lint`, `connectorgen-validate`; extend `verify:` line 60)
- T-18 (RED): a `Makefile`-target smoke assertion is not unit-testable — the RED artifact is a
  recorded failing run (`make lint` before config exists / lint findings on a seeded file in a
  scratch branch), logged in TDD-LEDGER per `workflow.tdd_mode` validation-artifact rule.
- B-18: minimal `.golangci.yml` (govet, staticcheck, errcheck, ineffassign, unused, misspell;
  exclusions for generated `*_gen.go`); Makefile `lint` uses the coordinator-decided acquisition
  path (SPEC §5); `verify: fmt tidy-check vet test build docs-check smoke lint connectorgen-validate`.
- Human-gate note: adding gates only (no reduction). Tool acquisition decision flagged.
- Verify: `make lint && make connectorgen-validate && make verify`

### T-19 / B-19 — cmd/inventorygen + inventory.json `wave:G`
- Files: `cmd/inventorygen/main.go`, `main_test.go`, generated `docs/migration/inventory.json`
- T-19 (RED): unit tests on bucketization (S<300/M300–699/L700–899/XL≥900 per
  `docs/migration/orchestration-plan.md` calibration), loc counting excludes `_test.go`?
  (decision: INCLUDE all `.go` — matches orchestration calibration; document), catalog join by
  `PMConnectorName`/`BareName`, stream_count from `connectors.ManifestOf` on `registryset.NewStaged()`.
- B-19: implement; run once and commit `docs/migration/inventory.json` (~558 entries).
- Verify: `go test ./cmd/inventorygen -v && go run ./cmd/inventorygen && python3 -c "import json;d=json.load(open('docs/migration/inventory.json'));assert len(d['connectors'])>500"`

### D-20 — migration docs + agent I/O schemas `wave:G` `docs-only`
- Files: `docs/migration/conventions.md`, `docs/migration/result.schema.json`,
  `docs/migration/review.schema.json`
- Content: bundle authoring recipe grounded in the three goldens; tier triggers (design §B.7);
  naming rules (§F.3); fixture recording+sanitization rules (THREAT-MODEL §4); documented
  parity-deviation ledger (starts with the stripe `minProperties` case); result/review JSON
  contracts per `docs/migration/orchestration-plan.md` §Per-agent task spec + §Verification pyramid.
- Verify: `python3 -m json.tool docs/migration/result.schema.json >/dev/null && python3 -m json.tool docs/migration/review.schema.json >/dev/null`
- Dependency: finalize AFTER Wave F merges (goldens are cited as references).

## Wave H — integration gate (1 task, serial, last) `wave:H`

### V-21 — phase gate + artifacts `wave:H`
- Owner: gsd-verifier (sonnet) under coordinator.
- Run: `go build ./... && go vet ./... && go test ./... && make verify && make lint`,
  `go run ./cmd/connectorgen validate internal/connectors/defs`,
  `go test ./internal/connectors/conformance -run TestConformance -v`,
  `go test ./internal/connectors/certify -run TestSourceStages -v`,
  `go test -cover ./internal/connectors/engine` (gate ≥85%, `EVAL-PLAN.md`),
  path-guard: `git status --porcelain` limited to planned paths.
- Update `SUMMARY.md`, `VERIFICATION.md`, `RUN-STATE.json`, `TDD-GATE.json`, PRD-COVERAGE re-run.

---

## Dependency graph (coordinator dispatch order)

```
Wave A: T/B-01  T/B-02  T/B-03  T/B-04          (parallel ×4)  + T/B-12 floats here (DECISIONS.md #2:
        certify report/cliharness has zero engine deps and lives in a disjoint package — dispatched
        alongside Wave A)
Wave B: T/B-05  T/B-06  T/B-07                  (parallel ×3; B-05 uses B-02, B-07)
Wave C: T/B-08  T/B-09                          (parallel ×2; need A+B)
Wave D: T/B-10  T/B-11                          (parallel ×2; T/B-12 moved to Wave A per DECISIONS.md)
Wave E: T/B-13  T/B-14                          (parallel ×2; need D)
Wave F: T/B-15  T/B-16  T/B-17                  (parallel ×3; need E — conformance auto-covers goldens)
Wave G: T/B-18  T/B-19  D-20                    (parallel ×3; need D; D-20 finalizes after F)
Wave H: V-21                                    (serial, last)
```

Task count: 21 (18 behavior/test pairs + 1 docs-only + 1 verification + B-16 carries the single
legacy edit). Estimated executor runs: 10–12 agents match the orchestration-plan wave0 budget when
waves A/B/C tasks are batched per agent.
