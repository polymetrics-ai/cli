# TDD ledger — T/B-14 certify source stages vs sample

Package: `internal/connectors/certify` (extends the Wave A `certify.Runner` skeleton from
T/B-12; zero imports of `internal/connectors/conformance` or `internal/connectors/engine`
per this task's scope guardrail).

## Discovery notes (ground truth, before writing tests)

- CLI dispatch: `internal/cli/cli.go` `Run` — commands `init`, `connectors inspect`,
  `credentials add|test`, `connections create`, `catalog refresh`, `etl run`, `query run`.
  No `--credential` flag exists on `etl check`/`etl read` (design doc gotcha #5) — not needed
  here because certify never calls `etl check`/`etl read` directly; live credential validation
  goes through `pm credentials test` (vault/secret-resolved) and reads go through `pm etl run`
  against a `connections create`-declared connection (source+destination pair), so the gotcha's
  prerequisite fix is genuinely out of scope for this stage list.
- `sample` connector (`internal/connectors/connectors.go:461`): NOT a `connectors.StatefulReader`
  — `Read` always emits all 3 `customers` records regardless of `req.State`/`since` config.
  Incremental filtering (dropping records below the prior cursor) happens entirely at the
  `internal/app` layer (`app.go:511`, `local_warehouse.go:162`) via `ParseSyncMode` +
  `compareCursor`. This is exactly the app-layer/connector-layer split documented in
  connector-certification-design.md's "honest answer" section — proven live end-to-end:
  run1 against a fresh `incremental_append` connection reads 3/loads 3 records; run2 (resume)
  reads 3 (source doesn't filter) but loads 1 (app-layer cursor filter drops the other two),
  cursor stays monotonic (max unchanged since no record exceeds it). This *is* the resume proof
  — no sabotage/synthetic proof needed, verified against the real CLI:
  ```
  run1: records_read=3 records_loaded=3 cursor=2026-06-22T09:15:00Z
  run2: records_read=3 records_loaded=1 cursor=2026-06-22T09:15:00Z (unchanged, i.e. monotonic)
  ```
- Deduped sync modes (`full_refresh_overwrite_deduped`, `incremental_append_deduped`) require the
  local `warehouse` destination (`app.go:441`: `runConnectorETL` rejects deduped modes outright;
  only `runWarehouseETL`, gated on `connectors.LocalWarehouseMaterializer`, supports them). So
  ALL five source stages route through a `warehouse` destination credential — this is also what
  unlocks `pm query run --table <name>` for the query_contract stage and PK-dedup verification.
- Capture-replay stages (6/7/10: overwrite / overwrite_deduped / incremental_append_deduped) use
  the built-in `file` connector reading a JSONL capture written from the stage-5 live read
  (`pm query run --table ... --json`, stripped of `_polymetrics_*` fields) — proven manually:
  `file` derives its single catalog stream name from the capture file's basename
  (`connectors.go:581`), `credentials add --connector file --config path=<file>` needs no Check
  at add-time, `credentials test` runs `File.Check` (stat) successfully once the capture file
  exists.
  - overwrite: two `etl run`s against the same capture leave `query --table` count unchanged
    (3, not 6) — truncate semantics confirmed live.
  - overwrite_deduped: a 3-line capture with one duplicate PK (newer cursor for `cus_001`)
    yields `records_read=3, records_loaded=2` and `pm query` returns exactly 2 rows, one per PK,
    keeping the newer cursor's field values — confirmed live.
- Stage 1 (`fixture_conformance`): `internal/connectors/defs/` contains ONLY `defs.go` in wave0
  (goldens land in Wave F) — `sample` (and every connector) has zero defs bundles right now.
  Per the parallel-agent boundary (conformance package owned elsewhere) and this task's explicit
  instruction, this stage is unconditionally recorded as skipped with reason "no defs bundle" for
  `sample`, without importing `internal/connectors/conformance` OR `internal/connectors/defs`.
  Real conformance-package integration is deferred to Wave F / V-21 per SPEC.md §1.6 exclusion
  list ("EXCLUDED from wave0... write protocol/ledger/sweeper... "; fixture_conformance real
  wiring is implicitly a Wave F concern once goldens exist, matching TEST-PLAN.md §1.6's
  "skip-with-reason when the connector has no defs bundle").
- Exit codes: `internal/cli/errors.go` `exitCodeFor` returns a *category*-keyed code (2 usage,
  3 validation, 4 auth, 5 connector, 6 runtime, 7 policy, 1 internal/other), not the 0/1/2/3
  scheme sketched in connector-certification-design.md §A "Exit codes" (that scheme applies to
  the not-yet-existing `pm connectors certify` subcommand itself, wired in a later phase). Stage
  assertions here check `ExitCode == 0` for happy-path stages via `Harness.MustKind`, which
  already asserts both kind and exit in one call (T/B-12).
- No secret value ever appears in `connectors inspot --json` / `credentials add|test --json`
  output (manually verified: `CredentialMeta` only carries `secret_fields` names, never values).

## T-14

Status: red-confirmed

Command:
```
go test ./internal/connectors/certify -run TestSourceStages -v
```

Output:
```
# polymetrics.ai/internal/connectors/certify_test [polymetrics.ai/internal/connectors/certify.test]
internal/connectors/certify/stages_source_test.go:211:10: undefined: certify.SabotageExpectedKind
internal/connectors/certify/stages_source_test.go:259:21: undefined: certify.LastWorkdir
FAIL	polymetrics.ai/internal/connectors/certify [build failed]
FAIL
```

Timestamp: 2026-07-02T09:07:45Z

Notes: `stages_source_test.go` authored first, covering the full stage-0..11 pipeline against the
real `sample` connector driven through the real CLI in an ephemeral `--root`
(`certify.NewRunner(...).Run(ctx)`), plus:
- a sabotage test (`certify.SabotageExpectedKind(r, "catalog", "NotTheRealKind")`) asserting
  exactly the sabotaged stage flips to failed, the overall report flips to `Passed=false`, and
  unrelated earlier stages (`preflight`) stay green;
- an ephemeral-workdir cleanup test (`certify.LastWorkdir(r)` + `os.Stat` `IsNotExist`) proving
  the runner does not leak its temp root when `KeepWork` is unset.

Compiler-error RED (undefined `SabotageExpectedKind`/`LastWorkdir`, and once those are stubbed,
`stages_source.go` itself does not exist) is the expected first-red shape for a task extending an
existing package with new exported surface, matching the T/B-12 ledger precedent.

## B-14

Status: green-confirmed

Implementation: `stages_source.go` — `Runner.Run` (replacing the T/B-12 `ErrNotImplemented`
skeleton method) drives 13 recorded stages end-to-end through `Harness.Run` in one ephemeral
`os.MkdirTemp` root per call:

- `init` (project setup; not a certification-design numbered stage but required before any
  other in-process `cli.Run` call succeeds — discovered live: `credentials add` on a bare tmp
  dir fails with "open project ... no such file or directory") + stage 0 `preflight`
  (`connectors list --json`, asserts the connector is registered + all `SecretEnv` values are
  non-empty).
- stage 1 `fixture_conformance`: unconditionally recorded `Passed=false` with reason
  `noDefsBundleReason` ("skipped: no defs bundle ..."); excluded from `allStagesPassed` and from
  the "every stage has a CLI record" self-check by name, since it makes no CLI call by design.
  Zero imports of `internal/connectors/conformance` or `internal/connectors/defs` (grep-verified).
- stage 2 `manual_json`: `connectors inspect <name> --json`, asserts kind `Connector` + scans
  stdout for the planted secret value.
- stage 3 `credentials_add` (+ a same-tier `warehouse_credentials_add` sub-stage) and
  `credentials_test`: `credentials add`/`credentials test --json`, asserting kind
  `Credential`/`CredentialTest` and re-scanning output for secret leakage. Live check goes
  through `credentials test` per design-doc gotcha #5 (never `etl check --connector`).
- stage 4 `catalog`: creates a `sample:cert-source -> warehouse:cert-warehouse` connection
  (`full_refresh_append`, table `cert_live_<stream>`) then `catalog refresh --json`; asserts
  `Catalog` kind, `Capabilities.Catalog.Streams >= 1`, and that the certified stream carries
  non-empty `primary_key`/`cursor_fields`.
- stage 5 `etl_full_refresh_append` (LIVE): `etl run --json` against the live connection;
  asserts `ETLRun` kind, records `Capabilities.Read` (stream/records) and
  `SyncModes["full_refresh_append"] = {pass, live}` (`passed_empty` fallback documented if
  records_read were 0 — not hit against `sample`, which always emits 3 customers). A paired
  `capture_write` stage (tier 1) then re-queries the just-loaded warehouse table via
  `query run --table ... --json`, strips `_polymetrics_*` bookkeeping fields, and writes the
  result as `capture_<stream>.jsonl` in the workdir — the capture source for stages 6/7/10.
- stages 6/7/10 (`etl_full_refresh_overwrite`, `etl_full_refresh_overwrite_deduped`,
  `etl_incremental_append_deduped`, all tier 1 / `data_source: capture`): register a
  `file`-connector credential pointing at the capture file once
  (`setupCaptureConnection`/`captureFileRegistered` memoization), then for each mode create a
  dedicated `file -> warehouse` connection (`--primary-key id --cursor updated_at`) and run
  `etl run`. Truncate semantics (stage 6) are proven by running twice and asserting the
  warehouse table's `pm query --table` row count is unchanged (not doubled). PK-dedup (stages
  7/10) is proven by `assertNoDuplicatePKs`, which re-queries the table and fails on any
  repeated `id` value.
- stages 8/9 (`etl_incremental_append` LIVE + `resume` LIVE run 2): a dedicated
  `incremental_append` connection is run once (recording `records_read`/cursor from the
  `ETLRun` envelope's `checkpoint` object), then run again; resume asserts run2's
  `records_read <= run1's records_read` and the cursor never regresses
  (`compareCursorStrings`, textual RFC3339 comparison — `sample`'s `updated_at` values compare
  correctly this way; documented rather than depending on `internal/app`'s unexported
  `compareCursor`). Proven against the real CLI before writing any assertion (see "Discovery
  notes" above): run1 reads 3/loads 3, run2 reads 3/loads 1, cursor unchanged (monotonic) — this
  IS the resume proof, no synthetic double-write needed.
- stage 11 `query_contract`: `query run --table cert_live_<stream> --limit 1 --json`, asserts
  `QueryResult` kind.
- Meta-stages: `finalizeJSONContract` counts every stage with a non-empty `CLI.Kind` (i.e. every
  stage that actually called the CLI) and flags `fail` if any stage's recorded error is a
  rendered `KindMismatchError` (`isKindMismatch` — a textual check, since `StageResult.Error` is
  a plain string field per the `Report` JSON schema, not a live `error` value).
  `finalizeSecretRedaction` re-scans every stage's `ArgvRedacted` for the planted secret values
  and reports `pass`/`fail`.
- `allStagesPassed` computes `Report.Passed`, treating `fixture_conformance`'s documented skip
  as non-fatal (it is the sole named exception).
- `SabotageExpectedKind(r, stage, wrongKind)` / `LastWorkdir(r)`: test-only exported hooks
  (documented as such in their doc comments) supporting the sabotage and workdir-cleanup
  self-tests without adding a production configuration surface for "expect the wrong kind".

`certify.go` changes: `Runner` gained two unexported self-test-only fields (`sabotage
*sabotage`, `lastWorkdir string`); the wave0 `ErrNotImplemented` skeleton `Run` method and its
`errors` import were removed (superseded by the real implementation) — `ErrNotImplemented` var
is gone since nothing returns it anymore and keeping a dead exported error would be a
lint/dead-code smell.

Command:
```
go build ./internal/connectors/certify/... ./cmd/... && \
go vet ./internal/connectors/certify && \
go test ./internal/connectors/certify -v
```

Output (tail):
```
=== RUN   TestSourceStagesAgainstSample
--- PASS: TestSourceStagesAgainstSample (0.60s)
=== RUN   TestSourceStagesSabotageFailsNamedStage
--- PASS: TestSourceStagesSabotageFailsNamedStage (0.56s)
=== RUN   TestSourceStagesEphemeralWorkdirCleanedUp
--- PASS: TestSourceStagesEphemeralWorkdirCleanedUp (0.60s)
PASS
ok  	polymetrics.ai/internal/connectors/certify	2.309s
```

21 tests total in the package (18 pre-existing T/B-12 report/harness tests + 3 new T/B-14
source-stage tests), 21 PASS, 0 FAIL, total runtime ~2.3s (well under the ~60s budget for a
real-CLI-driven ETL test). `gofmt -l internal/connectors/certify` empty.
`golangci-lint run ./internal/connectors/certify/...` → `0 issues` (one `errcheck` finding on
`defer os.RemoveAll(root)` fixed by wrapping in `func() { _ = os.RemoveAll(root) }()`).
`go vet $(go list ./... | grep -v /internal/connectors/conformance)` clean — the sole build
failure in the repo right now is inside `internal/connectors/conformance/` (T/B-13, a parallel
in-flight agent's own package — confirmed untracked/mid-edit via `git status --porcelain` and
file mtimes newer than this task's start; zero files under that path were read or written by
this task, and `go build ./internal/connectors/certify/... ./cmd/...` is green in isolation).

Path guard: `git status --porcelain` shows only `internal/connectors/certify/certify.go`
(modified), `internal/connectors/certify/stages_source.go` (new),
`internal/connectors/certify/stages_source_test.go` (new), and this ledger file as new/changed
paths under my ownership — plus the parallel T/B-13 agent's own untracked
`internal/connectors/conformance/` and `traces/waveE-b13-ledger.md`, neither touched by this
task.

Timestamp: 2026-07-02T09:24:00Z

## CLI gaps found

None. Every stage in scope (0-11, per PLAN.md T-14) maps onto an existing `pm` subcommand with
no missing flags:
- `pm init`, `pm connectors list --json`, `pm connectors inspect <name> --json`
- `pm credentials add <name> --connector <c> [--from-env f=ENV] [--config k=v] --json`
- `pm credentials test <name> --json`
- `pm connections create <name> --source c:cred --destination c:cred --stream s --primary-key k
  --cursor f --sync-mode m --table t --json`
- `pm catalog refresh --connection <name> --json`
- `pm etl run --connection <name> --stream <s> --json`
- `pm query run --table <t> [--limit n] --json`

Design-doc gotcha #5 (`--credential` needed on `etl check`/`etl read` because `directConnector`
never resolves vault secrets) is confirmed accurate for those two subcommands but does not
block T/B-14: certify never calls `etl check`/`etl read` directly — live credential validation
goes through `pm credentials test` (which does resolve secrets via `a.resolveCredential`), and
live reads go through `pm etl run` against a `connections create`-declared connection (which
resolves the source endpoint's credential the same way). This matches SPEC.md §1.6's explicit
wave0 exclusion of the `--credential` flag fix ("not needed for sample; documented as wave1
prerequisite").
