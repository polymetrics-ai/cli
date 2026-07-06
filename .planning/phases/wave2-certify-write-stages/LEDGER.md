# Certify WRITE stages (12-17) + create-then-cleanup protocol — execution ledger

Task: build certify WRITE stages (plan_preview / write_create / write_verify /
write_cleanup / cleanup_verify / approval_idempotency), the WritePairing
table, the write-ahead leak ledger, and the orphan sweeper, per
`docs/architecture/connector-certification-design.md` §A stages 12-17 + §C.

## Scope delivered

- `internal/connectors/certify/pairing.go` (+ `pairing_test.go`): tag
  convention `pm-cert-<slug>-<runid8>-<ts>` (`NewTag`/`NewRunID8`), default
  `create_X -> delete_X | close_X | archive_X` inference (`InferPairing`),
  curated `WritePairing` table for `github` (`create_label`/`delete_label`,
  `create_issue`/`close_issue`, `create_milestone`/`delete_milestone`),
  `record_schema`-driven data generation with per-connector `Overrides`
  (`GenerateRecord`, `GenerateRecordWithOverrides`).
- `internal/connectors/certify/ledger.go` (+ `ledger_test.go`): append-only,
  fsync'd write-ahead ledger (`certify-ledger.jsonl`): `RecordPlanned` before
  any live write, `RecordCleaned` after verified cleanup, `LoadLedger` folds
  planned/cleaned entries per tag, `Uncleaned()` for the sweeper, `CopyTo`
  persists a copy under `.polymetrics/certifications/ledger/<connector>.jsonl`
  even on crash.
- `internal/connectors/certify/stages_write.go` (+ `stages_write_test.go`):
  stages 12-17 wired into `Runner.Run` between `query_contract` (11) and
  `flow_roundtrip` (18):
  - `write_plan_preview` (12): `reverse plan` + `reverse preview --json`,
    asserting no non-empty `approval_token` field AND no raw token value in
    the JSON output (redaction gate).
  - `write_create` (13): ledger `RecordPlanned` BEFORE `reverse run
    --approve`, asserts `succeeded=1 failed=0`.
  - `write_verify` (14): live read-back via a dedicated connection/ETL run
    (or, self-test path, a direct outbox-file read); a miss is a warning
    (`verify: "unverified"`), never a hard stage failure.
  - `write_cleanup` (15): runs the pairing's `Cleanup` action; ledger
    `RecordCleaned` on success; a CLI-level failure itself already records a
    `leaked_resource` (design §C: "cleanup/verify fails -> leaked_resource").
  - `cleanup_verify` (16): re-checks the entity is gone; failure -> leak.
  - `approval_idempotency` (17): replays the consumed plan+token, asserts
    rejection (non-zero exit, `Error` envelope kind).
  - Two write-pairing sources: a curated `WritePairing` for the connector
    (github today) driving a real live-writes path, or — whenever no curated
    pairing exists (e.g. `sample`, which has no write capability at all) —
    an automatic self-test path against the built-in `sample`/`outbox`
    reverse-ETL round trip (per design: "if no live creds, the stage
    self-test uses the sample/outbox reverse-ETL path the Makefile smoke
    target already exercises"). The self-test seeds a dedicated one-row
    warehouse table (`cert_write_seed_<connector>`) containing the tag under
    a mappable field name, since `reverse plan --map` only renames existing
    columns — it cannot inject a constant value.
- `internal/connectors/certify/sweeper.go` (+ `sweeper_test.go`): `--sweep`
  orphan cleanup — scans the ledger for uncleaned entries older than
  `--older-than`, cleans them via the same plan/approve/run mechanics (either
  the outbox self-test tombstone path or a curated `WritePairing`'s real
  cleanup action), and marks them `RecordCleaned`.
- `report.go`: added `Leak`, `Report.Leaks`, `WriteActionResult`,
  `Capabilities.WriteActions`, `ExitCodeFor(rep) int` (0 pass / 2 fail / 3
  leaked — leaks dominate, matching design §A exit codes).
- `certify.go`: added `Options.Write bool` and the
  `SabotageCleanupVerifyEntityStillPresent` self-test seam.
- `stages_source.go`: wired the six write stages into the pipeline; extended
  `allStagesPassed` so any stage recording a documented `"skipped: ..."`
  reason (not just `fixture_conformance`) never fails the overall report.

## Concurrent-wave interaction (found, not authored, by this task)

While this task was in progress, a separate concurrent wave landed
`credsfile.go`, `batch.go` (+tests), `stages_glue.go` (flow/schedule stages
18-19, +tests), `record.go`/`replay.go` (Tier-1 record/replay, +tests), and
`internal/cli/certify_cli.go` (CLI wiring) in the same working tree. Two
concrete integration issues surfaced and were resolved:

1. **Shared-file clobber**: `certify.go`, `report.go`, and `stages_source.go`
   were each edited by both waves; a last-write-wins race briefly reverted
   this task's edits to those three files. Recovered by re-applying the
   write-stage wiring on top of the other wave's already-landed
   flow/schedule wiring (both now coexist correctly — verified by
   `TestGlueStagesAgainstSample` and `TestSourceStagesAgainstSample` both
   passing).
2. **Import cycle**: the concurrent wave's `internal/cli/certify_cli.go`
   imports `internal/connectors/certify`, which (pre-existing wave0 design)
   imports `internal/cli` from `cliharness.go` to drive `cli.Run` in-process
   — an unavoidable cycle under direct imports. The concurrent wave had
   already re-architected `cliharness.go` around a `SetCLIRunFunc` seam
   (`cliRunFunc` package variable + `certify_testmain_test.go`'s `TestMain`)
   to break it, but `cmd/pm/main.go` had not yet been updated to call
   `certify.SetCLIRunFunc(cli.Run)` at startup. Fixed with a minimal,
   purely-additive one-line wiring call in `cmd/pm/main.go` (outside this
   task's core scope but required for `go build ./...` to produce a working
   binary, per the other wave's own documented design in
   `cliharness.go`'s comments).

Also fixed 6 pre-existing `errcheck` lint findings in `record_test.go` /
`replay_test.go` (unchecked `resp.Body.Close()`) — trivial, no logic change
— so the shared `make lint` gate (which scopes
`internal/connectors/certify/...`) is green for both waves.

## Verification

- `go build ./...` — clean.
- `go vet ./...` — clean.
- `go test ./internal/connectors/certify/... -count=1` — 85+ tests pass, 0
  fail (source stages, glue stages, write stages, ledger, pairing, sweeper,
  batch, credsfile, record/replay, report, cliharness).
- `go test ./... -count=1` — full repo green.
- `make lint` (scoped: engine/defs/hooks/native/conformance/certify +
  connectorgen/inventorygen) — 0 issues.
- `make clean` — ok.

## Design decisions / trade-offs worth flagging

- **No connector currently registered implements `DefinitionProvider`** with
  a real `record_schema` outside the `engine` package (github is a
  hand-written Tier-3 connector using string action-name dispatch, not
  `engine.WriteAction`). `writeActionRecordSchema` therefore uses a small
  hand-curated `builtinWriteSchemas` map mirroring
  `internal/connectors/defs/github/writes.json`'s `create_label`/
  `delete_label`/`create_issue`/`close_issue`/`create_milestone`/
  `delete_milestone` schemas, rather than reading the real `writes.json` file
  at runtime (no loader for it exists on the hand-written-connector path).
- **outbox is append-only** (`Outbox.Write` never deletes), so the
  self-test's `cleanup_verify` cannot check "entity gone" literally. It
  checks "the last record appended for this tag reflects the cleanup action,
  not the original create action" — the append-only analogue of "gone",
  proving the write-protocol MACHINERY (ledger ordering, stage sequencing,
  leak detection) end-to-end without depending on outbox having real delete
  semantics it doesn't have.
- **`reverse plan --map` only renames existing source columns** (no constant
  injection) — the self-test path seeds a dedicated one-row warehouse table
  containing the tag under a mappable field name before planning, via the
  same file->warehouse ETL path the Makefile smoke target already exercises.
- Live-pairing write_verify/cleanup for a real connector (github) reuses the
  same `connections create` + `catalog refresh` + `etl run` pattern
  established by stages_source.go's stage 4, rather than `etl read
  --connector/--config` (which — per the design doc's own documented gotcha
  — never resolves vault-backed credential secrets and is out of this task's
  scope to fix).
