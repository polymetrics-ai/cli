# Issue 389 Shepherd Proof-Recovery Plan

## Current objective

Repair the existing Shepherd implementation on `fix/389-shepherd-proof-recovery` so completion proof,
recovery, promotion, registry loading, and external effects are real, durable, restart-safe, and
authority gated. This branch is stacked under parent branch `feat/372-gsd-pi-go-shepherd` / parent PR
#390. Do not merge PR #390, do not push to `main`, and do not run live Twenty/Asana canaries until the
exact candidate head has been independently validated by GPT-5.6 Sol/high.

## Required workflow and skills loaded

- GSD command: `scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`
- GSD adapter health: `scripts/gsd doctor`, `scripts/gsd list`
- Required reading completed this cycle: `AGENTS.md`, required-skills routing, GSD Pi adapter reference,
  issue-agent contract, issue #389 body, complete issue-389 planning artifacts, and
  `agent-runtime/shepherd/README.md`.
- Skills loaded/recorded: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`,
  `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`,
  `golang-context`, `golang-concurrency`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-observability`, and `golang-lint`.

## Reconciled status at start of proof-recovery

Earlier artifacts claimed independent validation, ratification, recovery planning, final verification,
and canary readiness. Those claims are not accepted as current evidence for this repair run. Read-only
recon found production code still manufacturing Sol/high validator, `PROCEED`, and `Ratified=true`
inside `cmd/shepherd/main.go:persistSuccessProof` without calling the real authority ratifier. Current
status is therefore **not validated, not ratified, not canary-ready**.

## Orchestration decisions

1. `cycle-0/reconcile`: spawned one read-only scout subagent for code recon; no mutating workers were
   launched. Decision: `read_only_parallel_recon` because the requested scope is broad and production
   edits must wait for artifact reconciliation and RED tests.
2. `cycle-0/execution-mode`: use `local_critical_path` for production edits in this checkout; no
   overlapping mutating workers are allowed.
3. `cycle-0/safety`: do not run Asana/Twenty canaries, GitHub mutations, or credentialed checks before
   exact-head independent Sol/high validation exists.
4. `cycle-1/store-guard`: committed/pushed the coherent store guard checkpoint before broader Slice A
   production changes.
5. `cycle-2/slice-a`: stayed on `local_critical_path`; added deterministic fake-validator integration
   tests at the typed validator port boundary before production rewiring.

## Ordered implementation slices

### A. Real independent validation and ratification

RED tests first:
- completion proof fails when validator evidence is missing or GPT-5.5;
- completion proof fails for stale candidate head;
- validator `RETRY`/`HALT` does not ratify or promote;
- production success path must call `authority.Ratify` and persist the real attestation.

GREEN target: keep candidates inside attempt worktrees, dispatch a genuinely separate GPT-5.6 Sol/high
validator against exact candidate head plus bounded artifact hashes/gates, persist observed model,
thinking, session identity, verdict, gates, evidence hashes, ratify with `authority.Ratify`, and promote
only after successful ratification.

Slice A implementation status: complete for deterministic local coverage. Attempt worktrees now produce
a candidate checkpoint before canonical promotion; invalid validation evidence leaves canonical HEAD and
canonical `.gsd` unchanged; production validation is behind a typed `validation.Validator` port and the
GSD-backed validator requires bounded `.gsd/shepherd-validation.json` evidence rather than manufacturing
a verdict in `persistSuccessProof`.

### B. Durable attempt lifecycle and crash recovery

RED tests first:
- all required attempt states persist in SQLite;
- early preparation/query/runtime failures transition explicitly;
- restart reconciles database-owned orphan worktrees/branches without deleting unknown or live paths;
- retry always creates a fresh attempt worktree.

GREEN target: durable attempt identity, branch, path, base/candidate/validated heads, and lifecycle
states: `created`, `prepared`, `running`, `validated`, `ratified`, `promoting`, `promoted`,
`retained_for_recovery`, `cleanup_pending`, `cleanup_complete`, and `cleanup_blocked`.

### C. Crash-safe GSD-state promotion

RED tests first:
- injected crashes before Git promotion, after Git promotion, before state swap, and after state swap;
- restart converges to exactly one consistent Git/GSD state;
- canonical `.gsd` is never removed and repopulated in place.

GREEN target: staged state snapshot validation, SQLite/WAL-safe handling, promotion journal, atomic
rename/swap with backup/recovery, and idempotent promotion/recovery.

### D. Complete official GSD 1.11 registry loading

RED tests first:
- real pinned fixture with array spreads such as `RUN_UAT_WORKFLOW_TOOL_NAMES` resolves;
- allowed/required/forbidden tools are preserved;
- null/unknown units fail closed unless explicitly governed sidecars;
- no hard-coded substitution for official phase metadata.

GREEN target: structured normalized export from pinned official runtime, metadata-only model routing,
and fail-closed registry admission.

### E. Real Sol/high recovery planning

RED tests first:
- static recovery-planning text is rejected;
- Sol/high recovery planner output must include observed model/thinking, evidence hash, typed action,
  bounded plan, and explicit allowlisted action;
- budget exhaustion enters durable `awaiting_decision` after restart.

GREEN target: dispatch bounded GPT-5.6 Sol/high recovery-planning units, persist selected typed actions,
and enforce per-class durable budgets.

### F. Authority-gated external effects

RED tests first:
- decision summaries, questions, statuses, and future GitHub mutations cannot bypass the fenced outbox;
- outbox pending/claim/send/failure/restart/idempotent replay paths are covered;
- workers do not receive a direct GitHub mutation path.

GREEN target: remove direct `SyncDecisionComment` production paths and route all external writes through
authority-granted, idempotent outbox effects.

### G. Real supervise integration coverage

RED tests first for:
- successful implementation -> independent Sol/high validation -> ratification -> promotion ->
  `final_human_gate`;
- missing/GPT-5.5 validator evidence;
- stale candidate head;
- validator `RETRY`/`HALT`;
- crash/restart at every promotion boundary;
- retained failed attempt followed by fresh attempt;
- recovery planning and `awaiting_decision` restart;
- outbox restart and duplicate suppression;
- official registry spread metadata;
- canonical worktree unchanged on every failed path.

GREEN target: deterministic local integration tests using fakes for external services, plus full nested
module gates. Live canaries remain deferred until independent validation passes and human approval exists.

## Checkpoints

- Commit/push plan reconciliation after artifact-only edits if requested by coordinator.
- Commit/push each coherent green slice after focused and nested module gates pass.
- Final branch push and draft sub-PR target: `feat/372-gsd-pi-go-shepherd`, title Conventional Commit,
  body uses `Refs #389`.
