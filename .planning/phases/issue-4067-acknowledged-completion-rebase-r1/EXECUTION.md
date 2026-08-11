# #4067 execution record

**Status:** RED recorded — no production mutation has occurred.

## TDD gate

The first execution action is a behavioral RED in `internal/app` against the real persisted JSON state path. It must demonstrate the Sol F1 durable `running` leak and zero/non-terminal returned run after an acknowledged checkpoint and an unrelated post-checkpoint writer. No `internal/` production file may change until that RED command and the test-only commit are recorded below.

## Planned checkpoints

| Checkpoint | Required evidence | Status |
|---|---|---|
| Planning | #4067 issue/readback, manual GSD context/plan/TDD ledger, clean rejected baseline custody | Complete before this commit |
| RED | Focused test, non-zero exit due to durable symptom, test-only commit | Observed; commit pending |
| GREEN | Minimal completion-boundary implementation and same test passing | Pending |
| Focused expansion | all modes, reopen, cancellation, fail-closed eligibility, race, #4046/R7/R8 | Pending |
| Generated remediation | canonical generator commands and candidate-owned diff only | Pending |
| Heavy validation | only after required user-facing window notification | Pending |
| Review/no-mistakes/CI | all findings dispositioned; fresh 0/5 no-mistakes; exact-head CI | Pending |

## RED — all-seven-mode acknowledged-completion leak

- Command: `go test -json -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes$'`
- Exit: `1` (expected RED)
- Fixture: a source-executor test seam pauses only after the real checkpoint callback has returned. A second App then persists unrelated stream/checkpoint/run data; release lets the original source report exhaustion and reach ordinary `completeRun`.
- Result: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture` each retained the acknowledged target stream and unrelated writer first, then failed with `RunETL returned zero run`, durable target status `running`, and a zero terminal timestamp. The test also required `errors.Is(runErr, errStateRevisionConflict)`.
- Boundary: this checkpoint includes only `internal/app/transport_dispatch_test.go` and planning evidence. `internal/app/app.go` remains untouched until the RED commit is made.
