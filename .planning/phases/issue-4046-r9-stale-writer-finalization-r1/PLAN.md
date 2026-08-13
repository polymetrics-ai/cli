# #4046 R9 — stale-writer run finalization plan

**Issue:** [#4046](https://github.com/polymetrics-ai/cli/issues/4046)
**Parent chain:** #3864 → #3862 → #4015
**Base:** `e5a55c68003e63860e8213b509a18643d5f4e3d6`
**Branch:** `fix/4046-r9-stale-writer-run-finalization`
**Plan type:** TDD, inline/manual GSD fallback

## GSD lifecycle record

The named issue phase cannot be resolved by the currently archived numeric roadmap (`phase_found: false`), and the canonical contract forbids role spawning. Execute the required lifecycle inline in this directory:

1. `scripts/gsd prompt discuss-phase issue-4046-r9-stale-writer-finalization-r1 --auto` — completed; decisions captured in `CONTEXT.md`.
2. `scripts/gsd prompt plan-phase issue-4046-r9-stale-writer-finalization-r1 --tdd` — completed; this plan is the manual planner output.
3. `scripts/gsd prompt execute-phase issue-4046-r9-stale-writer-finalization-r1` — generate and execute inline after this planning commit.
4. `scripts/gsd prompt verify-work issue-4046-r9-stale-writer-finalization-r1` — generate and record manual verification/UAT after implementation.
5. `scripts/gsd prompt code-review issue-4046-r9-stale-writer-finalization-r1` — generate and record manual code review with every finding disposition.

This fallback is limited to the non-numbered phase topology. It preserves the lifecycle, TDD sequence, review, and evidence; it does not authorize a missing role, a skipped gate, or an old-run takeover.

## Required skills

`golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-lint`, `github-issue-first-delivery`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, and `no-mistakes`.

## Plan 1 — R9 typed-conflict terminalization (TDD)

### Objective

Make a stale transport writer's generated run durably terminal and truthfully returned, while retaining the current R7/R8 per-stream CAS and current-state preservation guarantees.

### Read first

- `CONTEXT.md`
- `TDD-LEDGER.md`
- `internal/app/app.go`
- `internal/app/transport_dispatch.go`
- `internal/app/transport_dispatch_test.go`
- `.planning/phases/issue-3864-closed-transport-dispatch-r1/TDD-LEDGER.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-3864-r9-audit-r1/report.md`

### Task 1 — behavioral RED and commit

1. Strengthen or replace the stale-writer fixture with `TestRunETLTransportStaleWriterFinalizesLosingRun`.
2. Use two real `App` instances on the fixture's persisted JSON state. Pause the loser after destination acknowledgement; persist a winner advance for the same stream plus unrelated stream/checkpoint/run data; release the loser.
3. Assert the initial behavior is red: `errors.Is(losingErr, errTransportStreamStateConflict)` is true, the returned loser run is zero or not terminal, and reopening exposes the durable losing run still `running`; prove winner and unrelated snapshots remain unchanged.
4. Run the exact RED command and commit only the test/evidence as `test(4046-r9): reproduce stale writer run finalization leak`.

### Task 2 — smallest GREEN production transition and commit

1. In `internal/app/app.go`, make `failRun` recognize `errors.Is(runErr, errTransportStreamStateConflict)`.
2. Preserve the revision equality guard for every non-conflict failure.
3. For the typed conflict only, operate on the latest `current` state supplied under `JSONStore.Update` and change only the matching `running` run ID to `failed` with redacted error text and a completion timestamp.
4. Reject a missing or incompatible target status instead of replacing another terminal run.
5. Preserve the typed conflict through the returned error chain. Follow R9-T7 in `TDD-LEDGER.md`: return a transitioned run only after a successful or may-have-committed write, return `Run{}` after a definite pre-rename persistence failure, and return an already-terminal target unchanged.
6. Do not change `transport_dispatch.go`, retry a checkpoint, or assign a stale whole state.
7. Run the same witness green and commit only `app.go` plus the coherent RED/GREEN ledger update as `fix(sync): finalize stale transport writer run`.

### Task 3 — focused proof expansion

Add deterministic package-local tests for:

- an unrelated write inserted after conflict observation and before terminalization;
- restart/reopen truth;
- conflict finalization when cancellation follows acknowledgement;
- all seven modes: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture`;
- typed-error preservation and R7/R8 non-regression.

Each test must assert observable durable state, returned run identity/status, `errors.Is`, and winner/unrelated preservation as applicable. Commit the focused coverage and ledger update only after its tests pass.

## Threat model and fail-closed rules

| Threat | Mitigation and proof |
|---|---|
| Stale writer overwrites winner checkpoint | Leave `runTransportETL` target-entry CAS unchanged; test winner checkpoint identity after reopen. |
| Generic last-writer-wins replaces unrelated state | Permit latest-state mutation only when `errors.Is(runErr, errTransportStreamStateConflict)`; tests snapshot unrelated stream/checkpoint/run values. |
| Existing failures lose their revision protection | Guard a non-conflict revision mismatch with the exact current `errStateRevisionConflict` path. |
| Terminalizing the wrong run | Match the generated `runID` and require `Status == "running"` for the typed-conflict rebase. |
| Error identity or sensitive detail is lost | Use `errors.Is` and `errors.Join`; retain `safety.RedactErrorText` for persisted error text. |

## Verification sequence

The ordered commands and expected outcome live in `VERIFICATION.md`. Required milestones are:

- a real durable behavioral RED before production edits;
- matching narrow GREEN;
- repeated interleaving/restart/cancellation tests;
- all-seven-mode and `-race` proof;
- unchanged R7/R8 suite;
- `internal/app` package, vet/build, and individual repository gates;
- manual GSD verify/work and code-review dispositions;
- a fresh `no-mistakes axi run --intent ... --skip=push,pr,ci`, never `--yes`, starting at 0/5.

## Commit and delivery checkpoints

1. Planning evidence commit — no production paths.
2. Behavioral RED test/evidence commit — test must fail for the durable leaked-run symptom.
3. Minimal GREEN commit — `app.go`, tests, and evidence; test must pass.
4. Focused coverage/review-fix commits only when green.
5. After all local gates and review, start the fresh no-mistakes run. While it is active, the pipeline owns its findings and fixes; never edit around a gate.
6. Only after a successful child-local no-mistakes result, push this branch and open/update the stacked sub-PR against `feat/3862-any-to-any-transport`. Keep it unmerged. PR body uses `Refs #4046`, `Refs #3864`, `Refs #3862`, `Refs #4015`, and parent #4019; no closing keyword.

## Success criteria

- Exactly one production file, `internal/app/app.go`, implements typed-conflict-only run finalization.
- R7/R8 source identity and per-stream CAS remain unchanged.
- The RED demonstrates the historical durable `running` + zero return; the GREEN demonstrates a durable `failed` run with matching non-zero returned identity.
- Winner and unrelated state remain unchanged through reopen and intervening writes.
- The original stale conflict stays detectable by `errors.Is`.
- Cancellation, seven-mode, race, restart, package, vet/build, and repository matrices pass without any provider, credential, network, warehouse, container, or external-service operation.
- The old no-mistakes run is never queried for control, resumed, or modified; the new run remains at or below five correction loops.
