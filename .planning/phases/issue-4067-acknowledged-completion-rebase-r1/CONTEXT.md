# #4067 — acknowledged transport completion rebase context

**Gathered:** 2026-08-11
**Status:** Ready for TDD planning and execution
**Issue:** [#4067](https://github.com/polymetrics-ai/cli/issues/4067)
**Existing branch / PR:** `feat/3864-closed-transport-dispatch-nm5` / [#4059](https://github.com/polymetrics-ai/cli/pull/4059)

## Lifecycle fallback

The GSD adapter and canonical contract were validated with `scripts/gsd doctor`, resolved `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check`. The generated prompts for `discuss-phase` and `plan-phase --tdd` were resolved. `gsd-sdk query init.phase-op issue-4067-acknowledged-completion-rebase-r1` reported `phase_found: false` because this issue phase is not a numbered roadmap entry. The repository's delivery contract also forbids lifecycle-role spawning in this custody lane. This directory is the required inline/manual GSD fallback, not a lifecycle waiver: it records discussion, plan, RED/GREEN execution, verification, and review before delivery.

## Phase boundary

Repair only the successful terminal-completion aftermath of an already acknowledged transport checkpoint. A final completion may observe latest locked state only when the generated target run is still `running` and that target stream still exactly matches the checkpoint this run acknowledged. It may change only the target run's terminal fields and its own final stream/run metadata.

The solution must preserve unrelated fields and concurrent writers, never replay destination apply, never retry or overwrite a checkpoint, never use a generic refresh/last-writer-wins path, and retain definite-not-committed, committed, and indeterminate state-store truth plus the original detectable error chain.

## Locked decisions

- **D-01:** #4067 is distinct from #4046. The #4046 exception remains limited to `errors.Is(err, errTransportStreamStateConflict)` failure finalization; this issue must not broaden that path.
- **D-02:** The completed run is eligible for latest-state terminalization only if it remains `running` and the latest target stream exactly equals the run's acknowledged checkpoint. A changed/missing/incompatible target is fail-closed and must not be overwritten.
- **D-03:** A deterministic two-App real JSON-store witness is the RED. It must expose a zero returned run and durable reopened `running` run before production mutation, while proving the acknowledged checkpoint/winner and unrelated write already survived.
- **D-04:** GREEN proves the returned run identity/status equals the durable reopened terminal run, retains an `errors.Is`-detectable ordinary revision-conflict chain where applicable, and preserves all unrelated state.
- **D-05:** Coverage must drive the exact post-checkpoint/pre-completion interleaving through `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture`, plus restart/reopen, cancellation after acknowledgement, race, and R7/R8 regression.
- **D-06:** Generated artifacts are regenerated only through their canonical generators: `website/lib/docs.generated.ts` and `internal/connectors/certifications/flow-matrix.json`. No generated output is hand-edited.
- **D-07:** Fresh #4067 correction budget is 0/5. Do not start or compete with a no-mistakes run until focused/local gates and review are complete; never use `--yes`; never control old run `01KZQ0C1KEZRHNXX4WJFWXSCFB`.

## Canonical references

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/transport-correction-after-sol-audit.md` — controlling correction directive.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-final-sol-audit-r1/report.md` — F1 ordinary final-completion CAS witness and F2/F3 generated-drift findings.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-scope-gate-sol-r1/report.md` — candidate-scope disposition.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/report.md` — #4046/R7/R8 causal boundary.
- `.planning/phases/issue-4046-r9-stale-writer-finalization-r1/CONTEXT.md` and `TDD-LEDGER.md` — typed-conflict-only boundary to preserve.
- `.planning/phases/issue-3864-closed-transport-dispatch-r1/CONTEXT.md`, `TDD-LEDGER.md`, and `VERIFICATION.md` — established transport and R7/R8 proof.
- `internal/app/app.go` and `internal/app/transport_dispatch.go` — finalization and acknowledged-checkpoint boundary.
- `internal/app/transport_dispatch_test.go` — real persisted two-App transport fixtures.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` and `.agents/agentic-delivery/references/gsd-pi-adapter.md` — issue-first and inline GSD process.
- `docs/architecture/connector-certification-design.md` and `docs/connector-canon/IMPLEMENTATION-PROCEDURE.md` — canonical generated-artifact rules.

## Expected code shape (to validate before GREEN)

The Sol evidence locates the stale whole-state guard at normal `completeRun` finalization. Source inspection must identify the smallest way to bind final completion to the exact acknowledged checkpoint without changing `runTransportETL` CAS behavior. No design may introduce a generic App-wide state refresh or change the #4046 typed failure path.

## Explicit non-goals

No provider, credential, network, warehouse, container, external service, live certification, no-mistakes takeover, PR retarget, force-push, or merge is authorized.
