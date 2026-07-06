# Glue stages (18-19) + meta-scan/aggregation glue (20-21) — wave H

Phase: wave0-engine-harness · Wave: H · Executor: gsd-loop (sonnet), TDD pair.

## Scope

Built `internal/connectors/certify/stages_glue.go` + `stages_glue_test.go` implementing
certification design (`docs/architecture/connector-certification-design.md`) §A "Glue stages":

- **Stage 18 `flow_roundtrip`**: registers a dedicated capture-backed connection
  (`cert_flow_conn_<connector>`, full_refresh_append, mirroring stages 6/7/10's
  replay-through-file-connector pattern), writes an ephemeral two-step flow manifest
  (`cert_flow_<connector>.json`: a `sync` step over that connection + a dependent `query` step
  reading the sync step's output table), then drives the real `pm flow` CLI surface end to end:
  `flow plan` (asserts `status=ok`, `order=[sync_step, query_step]`), `flow preview` (asserts
  `status=dry_run` AND zero side effects — the query step's output table's existence is probed via
  `pm query run --table` immediately before and after preview and must be unchanged), `flow run`
  (asserts `status=ok`, both steps `status=ok`), `flow status` (asserts both steps report
  `status=success`). Sub-stages `flow_connection_create`/`flow_plan`/`flow_preview`/`flow_run`/
  `flow_status` are individually recorded; `flow_roundtrip` itself is an aggregate meta-stage
  (no direct CLI call) summarizing them, matching `Capabilities.Flow` (`*CapabilityResult`,
  pointer so it's omitted from JSON when unset, per `TestReportMarshalJSONShape`).
- **Stage 19 `schedule_roundtrip`**: redirects `internal/schedule`'s `CrontabBackend` to an
  ephemeral file via `PM_CRONTAB_FILE` (the same seam `internal/cli/schedule_test.go` uses) for
  the stage's duration only, restored via `defer` — the real operator crontab is never touched.
  Snapshots the (empty) redirected crontab, then `schedule create` → `schedule list` (asserts
  presence) → `schedule install --crontab` (asserts `backend=crontab` and the
  `# pm-schedule-<name>` sentinel — see `internal/schedule/crontab.go:88` — is present in the
  redirected file) → `schedule remove --crontab` → asserts sentinel absent AND the file is
  byte-identical to the pre-create snapshot. Any residue (failed sub-stage, sentinel still
  present, or crontab drift) sets `Capabilities.Schedule.Residue = true` and force-removes the
  sentinel line before returning, per design §D ("harness force-removes sentinel before
  reporting"). `Capabilities.Schedule` is `*ScheduleResult{Result, Backend, Residue, Reason}`.
- **Glue stage 20 (secret_redaction) / 21 (json_contract)**: no new code — `finalizeSecretRedaction`
  and `finalizeJSONContract` (stages_source.go, wave0) already iterate `rep.Stages` /
  `rc.capturedOutputs` unconditionally, so appending `stageFlowRoundtrip`/`stageScheduleRoundtrip`
  to the `Runner.Run` pipeline (stages_source.go's `stages := []stageFunc{...}`) before those two
  finalizers run automatically extends their coverage to the new stages' CLI output with no
  separate implementation needed. Verified by `TestGlueStagesSecretLeakInFlowStdoutFailsSecretRedaction`
  (plants a known secret in `flow_run`'s captured stdout via the existing `SabotageStdoutLeak`
  seam — proves the scan sees glue-stage output, not just source-stage output) and
  `TestGlueStagesAgainstSample`'s `Capabilities.JSONContract.StagesChecked >= 14` assertion.

`report.go` gained `Flow *CapabilityResult` and `Schedule *ScheduleResult` (new `ScheduleResult`
type: `{Result, Backend, Residue, Reason}`) on `Capabilities`, both `omitempty` pointers so
`TestReportMarshalJSONShape`'s pre-existing "capabilities.flow/schedule absent in wave0" assertion
still holds for any report that never reaches these stages.

## Grounding

Read before writing: `docs/architecture/connector-certification-design.md` §A (stage list, report
artifact shape), §D (flow+schedule stage spec verbatim), §E (package layout); `internal/cli/
flow_cli.go` in full (exact `flow plan|preview|run|status|list` subcommands/flags, `RunResult`/
`FlowManifest` JSON shapes — **no `"kind"` field on any flow envelope**, confirmed by reading
`flowPlan`/`flowRun`/`flowStatus`'s `writeJSON` call sites); `internal/flow/{manifest,engine,dag}.go`
(step kinds, `BuildDAG` topological order, `RunResult.Steps[].Status` values `dry_run|ok|skipped|
failed`); `internal/cli/schedule.go` + `internal/schedule/{crontab,schedule,select}.go` in full
(exact `schedule create|list|install|remove` flags, `# pm-schedule-<name>` sentinel format,
`PM_CRONTAB_FILE` env redirect — cross-checked against `internal/cli/schedule_test.go`'s own use
of the identical seam); `internal/cli/parse.go` (`parseFlags`/`parseGlobal` bare-flag-vs-value
parsing, load-bearing for `--crontab` as a boolean flag); existing wave0
`stages_source.go`/`report.go`/`report_test.go`/`stages_source_test.go` (capture-connection
helpers reused as-is: `setupCaptureConnection`, `captureStreamName`; `recordStage`/`assertKind`/
`cliInfoFrom` conventions followed exactly).

## Key finding: `pm flow status` / `pm flow run` checkpoint-dir mismatch (legacy CLI, not fixed)

`flowRun` (`internal/cli/flow_cli.go:170-181`) builds its `flow.FileCheckpointStore` at
`a.ProjectDir()` (`<root>/.polymetrics`) whenever a real `*app.App` is present — `--flows-dir` is
only used for by-name manifest resolution, never for the checkpoint store, when `--file` is given.
`flowStatus` (`internal/cli/flow_cli.go:196-239`), by contrast, ALWAYS builds its checkpoint store
at whatever `--flows-dir` it's given (default `os.TempDir()`), and reads the manifest from
`<flows-dir>/<name>.json`. So `flow run --file X --flows-dir Y` followed by
`flow status <name> --flows-dir Y` never sees the checkpoints `flow run` just wrote if `Y !=
a.ProjectDir()` — there is no existing CLI test that chains `flow run` + `flow status` against a
real project to catch this (`flow_cli_test.go`'s `TestFlowRunByNameResolvesProjectFlowManifest`
never calls `flow status`). Per this task's ground rule ("Legacy CLI surface is ground truth" /
no edits outside `internal/connectors/certify/**`), `stageFlowRoundtrip` works AROUND this rather
than patching `flow_cli.go`: it writes a second copy of the manifest at
`<root>/.polymetrics/<name>.json` and calls `flow status <name> --flows-dir <root>/.polymetrics`
(the project dir), matching `flow run`'s actual checkpoint location. Documented inline in
`stages_glue.go` at the `projectDir`/`projectManifestPath` block. **This is a real, unexercised gap
in the legacy CLI** (worth a follow-up fix to `flow_cli.go` itself so `--flows-dir` is consistent
between `run` and `status`), out of scope here.

## TDD evidence

RED: `stages_glue_test.go` written first; compile failure confirmed
(`rep.Capabilities.Flow undefined`, `rep.Capabilities.Schedule undefined`) before any
`stages_glue.go` code existed.

GREEN: `go test ./internal/connectors/certify/... -run TestGlueStages -count=3` — 5/5 tests green,
stable across 3 repeated runs:
- `TestGlueStagesAgainstSample` — full stage-18/19 sub-stage assertions (plan order, preview
  dry_run + zero side effects, run completed with both steps `ok`, status both steps `success`;
  schedule create/list/install sentinel/remove sentinel-absent+residue=false), plus
  `json_contract.stages_checked >= 14` and `secret_redaction.result == pass`.
- `TestGlueStagesFlowPreviewHasZeroSideEffects` — dedicated preview/run pairing check.
- `TestGlueStagesScheduleRoundtripLeavesNoResidue` — install+remove leaves `Residue=false`.
- `TestGlueStagesSabotageFlowFailsNamedStage` — `SabotageExpectedKind(r, "flow_run", ...)` flips
  only `flow_run`/overall `Passed`; `schedule_install` unaffected (sabotage is scoped).
- `TestGlueStagesSecretLeakInFlowStdoutFailsSecretRedaction` — `SabotageStdoutLeak(r, "flow_run",
  ...)` flips `secret_redaction` to `fail` naming `flow_run`, without touching `flow_run`'s own
  `Passed` outcome (mirrors the wave0 M2 regression test's shape exactly).

Also re-ran the full wave0 self-test set to confirm no regression:
`TestSourceStagesAgainstSample|TestSourceStagesSabotageFailsNamedStage|
TestSourceStagesSecretLeakInStdoutFailsSecretRedactionNamingStage|
TestSourceStagesEphemeralWorkdirCleanedUp` — all green except
`TestSourceStagesAgainstSample`, whose remaining failure is entirely attributable to a
concurrently-developed sibling task's write-protocol stages (`write_plan_preview`/`write_create`/
`write_verify`/`write_cleanup`/`cleanup_verify`/`approval_idempotency` — implemented in
`stages_write.go`, not touched by this task) lacking `CLI.ArgvRedacted` in their `Options.Write ==
false` skip path; NOT caused by `flow_roundtrip`/`schedule_roundtrip` (both pass and are correctly
exempted via a `metaStagesWithoutDirectCLICall` map added to the shared assertion in
`stages_source_test.go`, the same treatment `fixture_conformance` already had). Flagged as a
separate background task (`task_88ccb243`) rather than fixed here, since `stages_write.go` is
outside this task's scope. `report.go`/`stages_source.go`/`certify.go` were being concurrently
edited by that sibling task throughout this session (write-protocol stages 12-17 + batch/ledger/
pairing/sweeper/record/replay/credsfile modules landing in parallel in the same package) — this
task's edits were re-verified stable and green after each observed settle point.

## Self-verify

```
go build ./...                                            # clean
go vet ./internal/connectors/certify/...                  # clean
gofmt -l internal/connectors/certify/stages_glue.go \
  internal/connectors/certify/stages_glue_test.go \
  internal/connectors/certify/report.go \
  internal/connectors/certify/stages_source_test.go        # empty (formatted)
golangci-lint run internal/connectors/certify/...          # 0 issues in files this task touched
                                                             # (13 pre-existing issues in sibling
                                                             # task's ledger.go/record.go/
                                                             # record_test.go/replay_test.go/
                                                             # stages_write_test.go/zzdebug_test.go/
                                                             # pairing.go — out of scope, not fixed)
go test ./internal/connectors/certify -count=1              # 2 pre-existing failures in sibling
                                                             # task's write-protocol stages/tests
                                                             # (cleanup_verify leaked_resource bug;
                                                             # see task_88ccb243) — all stages_glue.go
                                                             # coverage green
```

No `internal/cli/certify_cli.go` wiring added: no CLI subcommand (`pm connectors certify`) exists
yet in this codebase (implementation-order step 5, per design doc, is a later task) — `Runner.Run`
is still exercised directly from Go tests, matching wave0's existing pattern.
