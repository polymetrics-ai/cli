# Issue 389 Shepherd Proof-Recovery Plan

## Current objective

Repair the existing Shepherd implementation on `fix/389-shepherd-proof-recovery` so completion proof,
recovery, promotion, registry loading, and external effects are real, durable, restart-safe, and
authority gated. This branch is stacked under parent branch `feat/372-gsd-pi-go-shepherd` / parent PR
#390. Do not merge PR #390, do not push to `main`, and do not run live Twenty/Asana canaries until the
exact candidate head has been independently validated by GPT-5.6 Sol/high.

## Required workflow and skills loaded

- Active GSD command: `scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`
- GSD adapter health/provenance: `scripts/gsd doctor`, `scripts/gsd list`, and
  `scripts/gsd sources programming-loop`.
- Required reading completed for Slice F: `AGENTS.md`, required-skills routing, GSD Pi adapter,
  issue-agent contract, universal runtime loop, local review loop, runtime/RLM integration reference,
  issue #389 planning artifacts, and the applicable project/Go skills.
- Skills loaded/recorded: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`,
  `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-safety`,
  `golang-security`, `golang-context`, `golang-concurrency`, `golang-database`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-observability`, `golang-lint`,
  `golang-code-style`, `golang-naming`, and `golang-documentation`.
- Immutable Slice F base: `9556cb24412f3598b2b8a94a3089b61ef3d1dd91`; local and remote heads
  matched and the worktree was clean before these planning-only edits.

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
   `1a050692f9e47b5b4d3d74cfb38e56c67d461399`. Slice C used `local_critical_path` because the
   promotion journal, Shepherd SQLite, canonical Git, staged GSD snapshot, and filesystem rename
   protocol form one critical transaction.
10. `cycle-9/slice-d`: Slice C is accepted GREEN at
    `f0fbf47f54c688792a5d53edfa4b680b38b39eed`. Slice D uses `local_critical_path`; official
    runtime validation, normalized export, registry admission, model routing, and command startup
    are one sequential trust boundary. Read-only reviewers are allowed; no overlapping mutating
    worker is allowed.
11. `cycle-11/slice-e`: Slice D is accepted GREEN at
    `cacb32e8e16b7ba70742cc5365cb83fffd74ca35`. Slice E uses `local_critical_path` because typed
    failure classification, separate Sol/high planner execution, protected session evidence, atomic
    SQLite budget reservation, backoff, and supervise retry state share one trust boundary. Read-only
    testing/reliability/review/security sidecars may overlap.
12. `cycle-15/slice-f`: Slice E is accepted GREEN at
    `9556cb24412f3598b2b8a94a3089b61ef3d1dd91`. Slice F uses `local_critical_path` because effect
    schemas, controller-derived authorization, durable outbox state, fenced claims, executor capability
    isolation, reconciliation, and command integration form one sequential trust boundary. Only
    read-only tester/reviewer/reliability/security sidecars may overlap. Slice G onward, PR creation,
    live GitHub mutation, canaries, cleanup/migration, and parent PR #390 merge remain blocked.

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

### C. Crash-safe GSD-state promotion — COMPLETE / GREEN at `f0fbf47f`

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
Accepted checkpoint: `f0fbf47f54c688792a5d53edfa4b680b38b39eed`.

### D. Complete official GSD 1.11 registry loading — COMPLETE / GREEN

Orchestration: `local_critical_path`; read-only recon/review sidecars only.

Required RED tests:
- realistic pinned 1.11 metadata containing array spreads resolves through a normalized exporter;
- exact allowed/required/forbidden tools and every forbidden reason survive normalization;
- missing, malformed, partial, oversized, duplicate, and unexpected JSON fields return
  `runtime_contract_mismatch`;
- missing registry/runtime, wrong version, symlink/path escape, and validated-source drift fail closed;
- official null `phaseChain` or tool contract never receives built-in fallback data;
- unknown units/phases fail closed; supported sidecars use separate versioned policy and are never
  represented as official metadata;
- official planning/research/discussion/completion/validation/UAT/ratification phases route to
  `openai-codex/gpt-5.6-sol`/`high`;
- official execution/execution_simple/delegated implementation phases route to
  `openai-codex/gpt-5.5`/`high`, without unit-name substitution;
- subagent model events cannot overwrite top-level observed model/thinking evidence;
- prompt-advertised tools are checked against the complete normalized tool contract;
- production startup cannot silently fall back to `BuiltinUnitRegistry`;
- host execution uses an absolute hash-pinned Node binary and a complete, deterministic, privately
  snapshotted GSD package tree outside canonical/attempt worktrees;
- registry imports consume only bounded no-follow verified bytes, eliminating verify/import TOCTOU;
- every top-level model/thinking transition and any session fallback are bound to the current run start;
- session headers and runtime source reads reject symlinks, duplicate/unknown/trailing fields, stale
  evidence, file replacement, and bounded-size violations;
- every observed `gsd_*` tool start is checked against the normalized unit contract;
- exporter process groups use bounded cancellation/WaitDelay cleanup; retained Podman execution is
  rebound to an immutable local image ID or fails closed.

GREEN target: import/evaluate only the validated pinned official 1.11 runtime through argv-based,
timeout/cancellation-aware process execution; emit bounded strict JSON preserving kind, scope class,
phase chain, allowed/required/forbidden tools and reasons; bind output to validated version/source
identity; compare all governance fields; route models from official phase metadata; keep sidecar policy
separate and versioned; and remove all production fallback to built-in/sample registries.

Implemented and verified on 2026-07-15. Host admission now pins the complete v3 package manifest,
absolute Node binary, prompt/model policy, active prompt tree, current-run session identity, and durable
unit attempt evidence. Canonical aliases use disposable worktrees and fresh checkpoints. Every observed
workflow tool is contract-checked; exporter/runner/validator process trees are bounded and synchronously
cleaned. Retained Podman and unqualified continuation fail closed. The installed official runtime,
focused/full/race/vet/build/make/boundary/list/diff gates, and the 28-finding no-differential lint
baseline pass. Slice E then became the only active slice.

Skills used for Slice D: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`,
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-context`, `golang-concurrency`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-observability`, and `golang-lint`.

### E. Real Sol/high recovery planning — COMPLETE / GREEN

Orchestration: `local_critical_path`; read-only tester/reliability/reviewer/security sidecars only.
GSD command: `scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`.
Skills: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-context`, `golang-concurrency`, `golang-database`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-observability`, `golang-lint`, `golang-code-style`,
`golang-naming`, and `golang-documentation`.

RED tests first:
- the static recovery sentence is rejected as non-evidence;
- all required failure classes are typed and exhaustive; unknown input fails closed;
- unsafe classes block without invoking a planner; external-effect uncertainty awaits decision;
- dead worker, silent tool, missing/invalid artifact, interruption, and reversible validation failure
  invoke a genuinely separate GPT-5.6 Sol/high planner only while their class budget permits;
- strict bounded JSON rejects missing/duplicate/unknown/partial/oversized/trailing fields, replay,
  nonce/head/class/hash/authority mismatch, stale/expired evidence, non-Sol/non-high identity,
  fabricated sessions, unknown or class-forbidden actions, and executable free-form instructions;
- class budgets are independently keyed by delivery/generation/unit/head/class, atomically reserved,
  fenced against concurrent duplicate consumption, immutable in policy, restart-safe, and enforce
  deterministic bounded exponential backoff;
- exhaustion durably selects `await_decision` when human input can help and `block` otherwise;
- planner timeout/cancellation and normal exit synchronously terminate descendants;
- planning and every rejected path leave canonical Git and canonical `.gsd` unchanged;
- every accepted retry uses a fresh Slice B attempt worktree.

GREEN target: introduce a narrow `internal/recovery` sidecar package and durable SQLite records. Launch
Pi directly in protected JSON mode with a fresh cryptographic nonce/session directory and exact
`openai-codex/gpt-5.6-sol`/`high`; bind one current-run top-level session and strict result to delivery,
generation, unit, attempt, head, failure class, evidence/authority hashes, action, bounded typed steps,
backoff, issue/expiry times, and replay state. Persist only redacted bounded diagnostics and structured
evidence. The controller—not model text—selects and executes allowlisted actions. No external effect is
retried in Slice E; GitHub/outbox uncertainty remains `await_decision`/blocked until Slice F.

Implemented and verified on 2026-07-15. `internal/recovery` now launches a fresh no-tool Pi process
pinned to `openai-codex/gpt-5.6-sol`/`high`, validates a strict lifecycle-correlated stream and durable
no-tool session, and accepts only nonce/evidence/authority-bound typed recommendations. Controller
policy makes unsafe or ambiguous joined failures dominant, validates class-specific actions and typed
primitives, and executes no model prose. SQLite persists globally ordered, per-class budgets with
owner/lease-epoch fencing, deterministic backoff, expiry, crash-safe dispatch disposition, legacy
exhaustion gating, and fresh-attempt retry ownership. GitHub/outbox uncertainty is typed and blocks
without another write. Live Sol/high smoke, focused/full/race/vet/build/make/boundary/list/diff gates,
the exact 28-finding no-differential lint baseline, and independent Sol/high correctness/security
review cycles pass with all actionable findings dispositioned. Accepted checkpoint:
`9556cb24412f3598b2b8a94a3089b61ef3d1dd91`.

### F. Authority-gated external effects — GREEN / REVIEWED / CHECKPOINT READY

Orchestration: `local_critical_path`; no overlapping mutating worker. Read-only tester, reliability,
reviewer, and security sidecars may overlap. No live GitHub mutation is permitted in this slice.

RED tests first:
- architecture checks reject direct `SyncDecisionComment`, `SyncQuestionComment`, or any other
  write-capable GitHub call outside the outbox GitHub executor;
- reply polling receives a read-only GitHub port, while only the executor receives write capability;
- workers, GSD units, validators, recovery planners, reviewers, enqueue helpers, and executors cannot
  mint grants or expand authority;
- controller-derived grants are immutable and bind delivery, repository, issue/PR, capability,
  generation, candidate head, epoch, and exact target;
- `forbidden_main_merge`, `merge.main`, `pr.merge`, unsupported effect types, changed targets, stale
  heads, stale epochs, and capability mismatch fail closed before enqueue or execution;
- strict versioned payload decoding rejects missing, duplicate, case-duplicate, unknown, partial,
  oversized, trailing, secret-bearing, command-bearing, and unbounded diagnostic fields;
- canonical payload bytes/hash, stable idempotency identity, grant identity, claim identity, and bounded
  typed result/error survive SQLite reopen without mutation;
- the explicit `pending`, `claimed`, `sent`, `failed`, `uncertain`, `blocked`, and `cancelled` graph
  rejects illegal, stale-owner, stale-epoch, and terminal transitions;
- expired claim recovery is fenced and restart-safe; definite pre-send failure may retry only within
  policy, while ambiguous post-send failure becomes `uncertain` and is never blindly replayed;
- deterministic Shepherd-owned comment markers reconcile exact marker/payload identity before write,
  suppress exact duplicates, and block duplicate markers, conflicting payloads, changed targets, or
  ambiguous read results;
- older claimed decision-summary revisions cannot overwrite a newer ledger revision;
- question effects remain bound to request ID, generation, unit, candidate head, and external comment ID;
- every decision-ledger/outbox crash boundary converges through deterministic startup reconciliation
  without claiming cross-database atomicity;
- telemetry covers `effect_requested`, `effect_authorized`, `effect_enqueued`, `effect_claimed`,
  `effect_execution_started`, `effect_sent`, `effect_failed`, `effect_uncertain`, `effect_reconciled`,
  `effect_blocked`, and `effect_claim_recovered` without secrets or unbounded payloads.

GREEN target: add a narrow `internal/outbox` package with strict immutable effect/grant/result records,
fenced durable claims, an allowlisted executor registry, deterministic GitHub comment reconciliation,
and bounded typed telemetry. Central controller admission derives grants; enqueue helpers and executors
only validate them. Replace decision-summary and decision-question direct publication with durable
request/enqueue and successful-result projection. Preserve read-only reply polling and fail closed for
all unsupported future PR/issue/push/merge effects. Use only fakes, in-memory runners, and `httptest`.

Focused gate:
`cd agent-runtime/shepherd && go test ./internal/outbox ./internal/store ./internal/github ./internal/domain ./internal/recovery ./cmd/shepherd -count=1`.
Full gates: nested tests/race/vet/build/`make verify`, root `make verify`, module boundary,
`git diff --check`, root `go list ./...`, and exact 28-finding lint baseline with zero differential.
GREEN achieved on 2026-07-16. RED preceded production edits. Focused/full/race/vet/build, nested/root
verification, module-boundary, root package-list, formatting, diff, and exact 28-finding lint-baseline
gates pass. Repeated independent GPT-5.6 Sol/xhigh correctness, security, and reliability findings are
fully dispositioned, including final exact-head correctness/restart closure with no actionable findings.
No live GitHub mutation occurred. The coherent checkpoint is ready; push only this branch, confirm
local/remote equality and cleanliness, then stop before Slice G.

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

- Do not checkpoint planning-only Slice F activation separately; production RED evidence follows next.
- Commit/push one coherent GREEN Slice F checkpoint after focused/full gates and exact-head independent
  reviews pass; preferred subject: `feat(shepherd): fence external effects through outbox`.
- Confirm exact local/remote branch equality, clean worktree, and no generated binaries, then stop before
  Slice G, PR creation, canaries, live GitHub mutation, cleanup/migration, or parent PR #390 merge.
