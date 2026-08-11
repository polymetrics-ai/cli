# #4067 TDD ledger — acknowledged transport completion rebase

**Issue:** #4067  
**Starting candidate:** `883a86cf0040d559edcd4777413d1c2de20cd94a` (immutable rejected baseline)  
**Correction ledger:** 0/5 before no-mistakes begins  
**Safety:** local fake-backed JSON-state tests only; no provider, credential, network, warehouse, container, or external service.

| ID | Requirement | RED command / expected failure | GREEN evidence | Status |
|---|---|---|---|---|
| C1 | An acknowledged run interrupted by an unrelated post-checkpoint writer returns zero/non-terminal and is durably `running` after reopen on the rejected baseline. | **Observed:** `go test -json -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes$'` exited `1`. Each canonical mode first retained the acknowledged target stream and unrelated writer state, then observed zero returned `Run`, durable `running` target run, and zero completion timestamp. | **Observed:** `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes$' -v` exited `0` in every canonical mode. Returned and reopened runs are matching terminal `completed` records; acknowledged checkpoint identity and unrelated writer data remain intact. | Green |
| C2 | Latest-state completion is eligible only for a matching `running` target run and exact acknowledged target stream. | `TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges` drives changed, missing, and terminal targets. | Exit `0`: all three leave their current target state intact, return `Run{}`, and retain `errors.Is(err, errStateRevisionConflict)`. | Green |
| C3 | Eligible completion changes only target terminal fields and its own final metadata; winner/acknowledged checkpoint and unrelated state survive. | Snapshot before/after state around a second unrelated writer. | C1's all-seven witness proves target checkpoint connection/stream/generation/timestamp/opaque identity stays unchanged while only final stream/run metadata advances; unrelated stream/checkpoint/run survives. | Green |
| C4 | Returned run and reopened durable run agree; ordinary revision-conflict/error chain remains detectable; definite/committed/indeterminate persistence outcomes stay truthful. | State-store outcome fixtures fail before rule exists or expose a speculative return. | `TestRunETLTransportAcknowledgedCompletionReturnsTruthfulPersistenceOutcome` exits `0`: definite pre-commit returns zero/running; committed unlock and indeterminate directory-sync outcomes return matching terminal runs and retain `CommitOutcomeError`; fail-closed C2 retains revision-conflict detection. | Green |
| C5 | Cancellation after acknowledgement and before completion remains truthful. | Deterministic cancellation fixture. | `TestRunETLTransportCancellationAfterAcknowledgedCheckpointForAllModes` exits `0`: each mode retains the acknowledged checkpoint, returns/durably stores a failed run, and retains `errors.Is(err, context.Canceled)`. | Green |
| C6 | Same interleaving is correct in all canonical modes. | Table-driven `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, `change_capture`. | C1/C5 cover every listed mode; focused set passes repeated normal and race runs. | Green |
| C7 | #4046 typed-conflict-only behavior and R7/R8 per-stream CAS/source identity remain unchanged. | Existing focused regression commands. | Exact R7/R8 plus #4046-focused command exits `0`; the #4046 failure path remains separate from normal completion. | Green |

## Commit gates

1. Planning evidence — this commit; no production paths.
2. RED — test/evidence only, committed before production mutation; command must exit non-zero for the durable leak.
3. GREEN — smallest final-completion implementation, passing matching test.
4. Focused coverage/generator/review fixes only after their corresponding checks pass.
