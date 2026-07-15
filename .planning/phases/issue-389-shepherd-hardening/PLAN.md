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
  `golang-database`, `golang-observability`, and `golang-lint`.

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
6. `cycle-3/retry-correction`: previous Slice A completion at `19d051c6` is recorded as a false green;
   production validation had no real proof producer and trusted worktree-local `.gsd/shepherd-validation.json`.
7. `cycle-4/live-pi-correction`: `99604d48` is a second Slice A false green because it invoked the
   unsupported invented `gsd headless shepherd-validate` verb. Replaced it with a separately configured,
   capability-probed Pi executable using the exact installed non-interactive JSON/read-only interface.
8. `cycle-5/slice-b`: Slice A is accepted GREEN at `95a17f18`. Slice B used `local_critical_path`
   because store, workspace, and supervise share mutation paths; no overlapping mutating workers.
9. `cycle-6/slice-c`: Slice B is independently accepted GREEN at
   `1a050692f9e47b5b4d3d74cfb38e56c67d461399`. Slice C uses `local_critical_path` because the
   promotion journal, Shepherd SQLite, canonical Git, staged GSD snapshot, and filesystem rename
   protocol form one critical transaction. Slice D onward, PR creation, final Sol review, and
   canaries remain blocked.

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

Slice A implementation status: **RETRY / false green at `19d051c6`**. The prior code improved
candidate-before-promotion flow, but the production validator had no real result producer, used the
canonical `validate-milestone` workflow unit as a generic validator, trusted a worker-controlled
`.gsd/shepherd-validation.json`, fabricated fallback session identity, used generation as state version,
hard-coded PR base as `main`, and blindly required/claimed UAT gates. The corrected Slice A must add RED
tests against the real production validator and keep canaries, PR creation, and Slice B blocked.

Retry RED tests proved: no validation-result producer, stale pre-existing result, no new validator
session, GPT-5.5 model, non-high thinking, result head/evidence/request nonce mismatch, candidate moves
during validation, stale base/governance version, RETRY/HALT/missing gates, and unchanged canonical Git
and `.gsd` on rejected command paths. The `99604d48` correction was still false green because its
subprocess did not exist in official GSD. Final corrected production behavior now invokes a configured Pi
executable with `--mode json --print`, exact Sol/high, `read,bash,grep,find,ls`, disabled project resources,
a dedicated protected session directory, a bounded capability probe, fresh nonce directories per retry,
and redacted process errors. Protected evidence binding, ratification order, delayed promotion, durable
state version, full attestation persistence, and metadata-derived gates remain enforced.

### B. Durable attempt lifecycle and crash recovery — COMPLETE / GREEN at `1a050692`

Independent checkpoint acceptance: `1a050692f9e47b5b4d3d74cfb38e56c67d461399`.
Slice A acceptance: GREEN at `95a17f18274c87ed0e3fde825b41257039b757de`; preserve its Pi validator,
protected evidence, ratification, and delayed promotion behavior.

RED tests first:
- all required attempt states persist through SQLite reopen and existing databases migrate intact;
- duplicate identity cannot rebind branch/path/base; illegal, terminal, and stale-owner transitions fail closed;
- real create/prepare/dispatch/validate/ratify/promote/failure/cleanup paths transition explicitly;
- restart reconciliation cleans only exact database-owned, non-live attempts and is idempotent;
- unknown, mismatched, checked-out, and live worktrees/branches are untouched and blocked;
- retry after retention always creates a fresh branch/path.

GREEN delivered: durable attempt identity, positively confirmed branch/path ownership, controller
owner/epoch, base/candidate/validated heads, bounded diagnostics, timestamps, and the exact 11-state
lifecycle. Repository-global locking and SQLite fencing cover bootstrap, query, execution, promotion,
and cleanup. Startup safely reconciles confirmed non-live resources, interrupts stale delivery/unit
runs, preserves ambiguous running/promoting/unconfirmed resources, and permits human-gated convergence
only after exact resources are proven absent. No broad worktree prune, broad branch deletion, unproven
path removal, or `RemoveAll` was introduced. Slice C promotion journaling and atomic `.gsd` state swap
remain explicitly excluded.

### C. Crash-safe GSD-state promotion — GREEN (candidate diff)

RED tests first:
- journal close/reopen persistence and exact identity/proof/state/snapshot binding;
- failures before Git promotion, immediately after Git promotion, before backup rename, between the
  backup and install renames, and after install before completion;
- restart at every boundary converges idempotently with no duplicate effect;
- canonical Git and `.gsd` never finish mixed; moved/dirty canonical state fails closed;
- missing, changed, corrupt, unknown, or symlinked stage/backup resources are preserved and blocked;
- expired/mismatched proof blocks before Git promotion, while already-promoted valid journals finish
  forward without a new model verdict;
- SQLite WAL test data survives a consistent staged snapshot; installed `gsd.db` passes integrity
  checking and has no stale WAL/SHM dependency;
- all Slice A/B validator, ratification, lifecycle, cleanup, and restart tests remain green.

GREEN target: a protected SQLite promotion journal with states `journal_created`, `state_staged`,
`git_promoting`, `git_promoted`, `state_swap_started`, `state_installed`, `complete`, and `blocked`;
a bounded deterministic no-symlink GSD manifest; SQLite-safe `gsd.db` snapshotting and integrity checks;
same-filesystem stage/backup rename installation with parent-directory fsync; forward-only recovery
once Git reaches the candidate; exact journal-owned cleanup; and startup recovery before canonical GSD
query or dispatch. Slice D onward remains excluded.

Implemented with journal intent before staging, full proof/attestation identity binding, bounded
root-confined copies, SQLite online backup/integrity verification, exact pre-Git ownership rechecks,
crash-safe two-rename installation, rooted cleanup tombstones, and universal blocked-journal gates.
Exact commit evidence is recorded after the Slice C checkpoint commit.

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
