# #4046 R9 — stale-writer run finalization context

**Gathered:** 2026-08-11
**Status:** Ready for TDD planning and execution
**Primary issue:** [#4046](https://github.com/polymetrics-ai/cli/issues/4046)
**Parent chain:** [#3864](https://github.com/polymetrics-ai/cli/issues/3864) → [#3862](https://github.com/polymetrics-ai/cli/issues/3862) → [#4015](https://github.com/polymetrics-ai/cli/issues/4015)

## Lifecycle fallback

The required GSD command path was resolved with `scripts/gsd doctor`, `scripts/gsd sources`, and `go run ./cmd/agentcontractgen check`. The following generated prompts were executed:

```text
/gsd-discuss-phase issue-4046-r9-stale-writer-finalization-r1 --auto
/gsd-plan-phase issue-4046-r9-stale-writer-finalization-r1 --tdd
```

`gsd-sdk query init.phase-op issue-4046-r9-stale-writer-finalization-r1` returned `phase_found: false`: the current `.planning/ROADMAP.md` is intentionally an archive entry point rather than an active numbered-roadmap registry. The canonical delivery contract also forbids spawning lifecycle roles in this custody lane. This directory is therefore the documented inline/manual GSD fallback, not a lifecycle waiver. The user-provided ship brief, accepted #4046 addendum, and R9 causal report lock the decisions that `discuss-phase --auto` would otherwise gather.

## Phase boundary

Repair only the stale per-stream-CAS aftermath: a losing transport ETL attempt must terminalize its own durably created `running` run as `failed` after `errTransportStreamStateConflict`, without changing the protected winning stream checkpoint or unrelated project state.

Production scope is strictly `internal/app/app.go`. Focused test scope is `internal/app/transport_dispatch_test.go` (and existing package-local test helpers only). Planning evidence remains under this directory. Stop rather than expand into transport orchestration, provider code, connector manifests, credentials, network behavior, warehouse state, containers, runtime services, CLI/docs, or certification.

## Locked implementation decisions

### Typed conflict boundary

- **D-01:** Detect `errTransportStreamStateConflict` with `errors.Is` in `failRun`; do not compare an error string and do not alter the original error chain.
- **D-02:** Only that typed conflict may rebase run terminalization onto the latest state read under the existing `JSONStore.Update` lock. All non-conflict failures retain their current whole-state revision guard.
- **D-03:** The rebase changes only the matching generated losing run ID from `running` to `failed`, sets a redacted failure string and completion timestamp, and rejects an absent or incompatible non-running target rather than fabricating a terminal result.

### State preservation

- **D-04:** Do not refresh or replace the losing App's whole state, retry the rejected checkpoint write, weaken the R7/R8 resume-identity or stream-entry CAS, or overwrite/merge the winner checkpoint.
- **D-05:** Winner stream state, unrelated stream state, unrelated checkpoints, credentials, connections, and other runs must carry forward from the latest locked state unchanged.
- **D-06:** The successful terminal write returns the exact persisted losing `Run`. R9-T7 in `TDD-LEDGER.md` owns commit-outcome truth: return a transitioned run only after success or a may-have-committed outcome, and return `Run{}` after a definite pre-rename persistence failure. An already-terminal target may be returned unchanged. Every error path preserves `errors.Is(err, errTransportStreamStateConflict)`.

### Proof and safety

- **D-07:** The RED is behavioral, not error-code-only: two `App` instances start from the same revision, the winner advances its stream and unrelated state, the loser loses the real CAS, then reopen proves a durable `failed` loser and non-zero matching returned run.
- **D-08:** Test fake executors and the local JSON state store only. No provider, credential, network, warehouse, container, or external-service activity is permitted.
- **D-09:** Cancellation coverage preserves the existing acknowledgement-before-cancellation ordering. All seven modes must use the same finalization path.
- **D-10:** The previous run `01KZQ0C1KEZRHNXX4WJFWXSCFB` remains immutable at its exhausted 5/5 ledger. This phase begins a new no-mistakes budget at 0/5 and never invokes `respond`, `sync`, `abort`, `rerun`, or recovery for the previous run.

## Canonical references

Downstream implementation and review must read these inputs before changing behavior:

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/report.md` — complete causal witness, R9 acceptance contract, and ordered verification matrix.
- `.planning/phases/issue-3864-closed-transport-dispatch-r1/CONTEXT.md` — prior #3864 closed-transport decisions and manual-GSD precedent.
- `.planning/phases/issue-3864-closed-transport-dispatch-r1/TDD-LEDGER.md` — R7/R8 T20/T21 RED/GREEN evidence.
- `.planning/phases/issue-3864-closed-transport-dispatch-r1/VERIFICATION.md` — prior bounded verification and honest non-certification boundary.
- `internal/app/app.go` — `updateState`, `RunETL`, and `failRun` state boundaries.
- `internal/app/transport_dispatch.go` — preserved R7/R8 checkpoint identity and per-stream CAS boundary; it is read-only for this phase.
- `internal/app/transport_dispatch_test.go` — deterministic two-App transport fixture and prior stale-writer test.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — issue-first, stacked-PR, review, and handoff rules.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — required GSD adapter command path and inline fallback rules.

## Existing-code insights

- `runTransportETL` calls `updateState` under `JSONStore.Update` and correctly returns `errTransportStreamStateConflict` when the captured stream entry no longer matches; it intentionally does not mutate the winner.
- `updateState` assigns `a.state` only on a successful or may-have-committed store update. A typed conflict therefore leaves the losing App's revision stale.
- `RunETL` routes a transport failure to `failRun`; its existing whole-state revision comparison then rejects the failure write, producing the durable `running` leak.
- The existing `TestRunETLTransportRejectsStaleCheckpointWriter` already orchestrates a deterministic winner/loser race and is the closest behavioral fixture, but it needs durable losing-run assertions and typed error checks.

## Deferred ideas

None. Generic refresh behavior, stream-checkpoint retry, provider behavior, state-store redesign, and certification are explicitly outside this phase.
