# #4067 — acknowledged transport completion rebase context

**Gathered:** 2026-08-11
**Status:** Sol r3 F1 continuation planned; loop 4/5 TDD RED pending
**Issue:** [#4067](https://github.com/polymetrics-ai/cli/issues/4067)
**Existing branch / PR:** `feat/3864-closed-transport-dispatch-nm5` / [#4059](https://github.com/polymetrics-ai/cli/pull/4059)

## Lifecycle fallback

The GSD adapter and canonical contract were validated with `scripts/gsd doctor`, resolved `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check`. The generated prompts for `discuss-phase` and `plan-phase --tdd` were resolved. `gsd-sdk query init.phase-op issue-4067-acknowledged-completion-rebase-r1` reported `phase_found: false` because this issue phase is not a numbered roadmap entry. The repository's delivery contract also forbids lifecycle-role spawning in this custody lane. This directory is the required inline/manual GSD fallback, not a lifecycle waiver: it records discussion, plan, RED/GREEN execution, verification, and review before delivery.

## Phase boundary

Repair only the terminal aftermath of an already acknowledged transport checkpoint. Successful completion and post-acknowledgement error finalization may observe latest locked state only when the generated target run is still `running` and that target stream still exactly matches the checkpoint this run acknowledged. Completion may change only its target run's terminal fields and own final stream/run metadata; error finalization may change only the target run's failed terminal fields.

The solution must preserve unrelated fields and concurrent writers, never replay destination apply, never retry or overwrite a checkpoint, never use a generic refresh/last-writer-wins path, and retain definite-not-committed, committed, and indeterminate state-store truth plus the original detectable error chain.

## Locked decisions

- **D-01:** #4067 is distinct from #4046. The #4046 exception remains limited to `errors.Is(err, errTransportStreamStateConflict)` failure finalization; this issue must not broaden that path.
- **D-02:** The completed run is eligible for latest-state terminalization only if it remains `running` and the latest target stream exactly equals the run's acknowledged checkpoint. A changed/missing/incompatible target is fail-closed and must not be overwritten.
- **D-03:** A deterministic two-App real JSON-store witness is the RED. It must expose a zero returned run and durable reopened `running` run before production mutation, while proving the acknowledged checkpoint/winner and unrelated write already survived.
- **D-04:** GREEN proves the returned run identity/status equals the durable reopened terminal run, retains an `errors.Is`-detectable ordinary revision-conflict chain where applicable, and preserves all unrelated state.
- **D-05:** Coverage must drive the exact post-checkpoint/pre-completion interleaving through `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture`, plus restart/reopen, cancellation after acknowledgement, race, and R7/R8 regression.
- **D-06:** Generated artifacts are regenerated only through their canonical generators: `website/lib/docs.generated.ts` and `internal/connectors/certifications/flow-matrix.json`. No generated output is hand-edited.
- **D-07:** Fresh #4067 correction budget is 0/5. Do not start or compete with a no-mistakes run until focused/local gates and review are complete; never use `--yes`; never control old run `01KZQ0C1KEZRHNXX4WJFWXSCFB`.
- **D-08 (r2):** Acknowledgement is a durable witness even when the orchestrator subsequently returns an error. `runTransportETL` must preserve that witness only for post-ack errors so a dedicated, narrow failure finalizer can terminalize the exact still-running run after an unrelated revision. It must preserve the original cancellation or source-error chain and never broaden ordinary `failRun`.
- **D-09 (r2):** An acknowledged rebase that cannot find its exact target run is a fail-closed revision conflict, not an ordinary `run not found`. It returns `Run{}` and leaves reopened state unchanged.
- **D-10 (r2):** The behavioral RED is table-driven across all seven modes. Each post-ack cancellation case persists an unrelated write before release and proves one apply, preserved acknowledged/unrelated state, original error identity, and matching returned/reopened failed run. A representative post-ack source-error case follows the same path; missing-run coverage independently proves the typed chain in all modes.
- **D-11 (r2):** Two correction loops have already been consumed. Only loops 3/5 through 5/5 are available, all without `--yes`; #4059 remains the sole draft stacked PR and #4068 remains closed/unmerged.
- **D-12 (r3):** A later `errTransportStreamStateConflict` after an earlier durable page acknowledgement is owned by the existing #4046 typed-conflict terminalizer. `failAcknowledgedTransportRun` must delegate directly to `failRun` before reading an older acknowledgement witness; every non-typed-conflict path retains the r2 exact-stream/running-run guard.
- **D-13 (r3):** The RED/GREEN is a deterministic real-JSON-store two-page, two-App witness in all seven modes. Page one must be durably acknowledged; a distinct winner then advances the target stream; the loser applies page two exactly once before its checkpoint CAS rejects. The proof requires a typed error, preserved winner and unrelated state, a non-zero returned failed loser matching reopen, and no retry/replay.
- **D-14 (r3):** Transport-core and existing-connector proof are distinct gates. The deterministic persisted-state/race evidence remains primary. Connector evidence must use the real `pm` binary, current GitHub definition/hook tests, command-runner preflight, inspection, and only a repository-owned bounded harness valid at this branch.
- **D-15 (r3):** A bounded read-only GitHub smoke is conditional on an already-approved credential in its sanctioned secret channel. No token may be copied to a file, command line, report, log, or chat; no credential may be requested or created. If no sanctioned channel is available, record the limitation rather than substituting a provider action.
- **D-16 (r3):** Three correction loops are consumed. This correction may submit only loop 4/5 locally, without `--yes`, with no push, PR action, retarget, or merge. A document step may not retain an unrelated architecture rewrite.

## Canonical references

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/transport-correction-after-sol-audit.md` — controlling correction directive.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-final-sol-audit-r1/report.md` — F1 ordinary final-completion CAS witness and F2/F3 generated-drift findings.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-scope-gate-sol-r1/report.md` — candidate-scope disposition.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/report.md` — #4046/R7/R8 causal boundary.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/sol-audit-r2-correction-handoff.md` — controlling r2 F1/F2 correction route.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-4067-final-sol-audit-r2/report.md` — independent F1/F2 witness, call path, and evidence-truth findings.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/sol-r3-correction-loop4.md` — controlling r3 F1 correction and loop-4 verification route.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-4067-final-sol-audit-r3/report.md` — independent two-page typed-conflict witness and exact fail-acknowledged call path.
- `.planning/phases/issue-4046-r9-stale-writer-finalization-r1/CONTEXT.md` and `TDD-LEDGER.md` — typed-conflict-only boundary to preserve.
- `.planning/phases/issue-3864-closed-transport-dispatch-r1/CONTEXT.md`, `TDD-LEDGER.md`, and `VERIFICATION.md` — established transport and R7/R8 proof.
- `internal/app/app.go` and `internal/app/transport_dispatch.go` — finalization and acknowledged-checkpoint boundary.
- `internal/app/transport_dispatch_test.go` — real persisted two-App transport fixtures.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` and `.agents/agentic-delivery/references/gsd-pi-adapter.md` — issue-first and inline GSD process.
- `docs/architecture/connector-certification-design.md` and `docs/connector-canon/IMPLEMENTATION-PROCEDURE.md` — canonical generated-artifact rules.

## Expected code shape (to validate before GREEN)

The r2 Sol evidence locates the stale whole-state guard on the post-error `failRun` path after `runTransportETL` discards its acknowledged witness, and it locates an untyped missing-run branch in `completeRunWithAcknowledgedTransportState`. Source inspection confirms the smallest r2 repair: preserve a post-ack result across the orchestrator error; use a narrow acknowledged-failure finalizer guarded by the exact stream witness and `running` target; and wrap a missing rebased target in `errStateRevisionConflict`.

The r3 Sol evidence then locates a narrower ordering error inside `failAcknowledgedTransportRun`: it consults the older acknowledged page-one witness even when the new error is already `errTransportStreamStateConflict` from the second page. The smallest repair is the exact early branch `if errors.Is(runErr, errTransportStreamStateConflict) { return a.failRun(runID, runErr) }`, before any witness lookup. This preserves #4046's typed-conflict-only behavior and leaves every r2 ordinary-error guard intact.

## Explicit non-goals

No provider mutation, credential acquisition or disclosure, warehouse mutation, container, external service, live certification claim, no-mistakes takeover, PR retarget, force-push, or merge is authorized. A pre-approved, bounded read-only GitHub smoke is permitted only through its existing sanctioned secret channel and never substitutes for deterministic core evidence.
