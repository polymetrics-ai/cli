# Pre-#390 Merge Implementation Plan

## Purpose

This document is the implementation reference for the remaining autonomous Shepherd work that must be completed before parent PR #390 (`feat/372-gsd-pi-go-shepherd` -> `main`) is considered ready for human merge.

The plan is intentionally ordered. Do **not** remove Podman leftovers, migrate `.gsd/.planning` state, or clean obsolete files until the real `shepherd supervise` integration path and merge-disabled Twenty/Asana canaries have reached `final_human_gate`.

## Non-negotiable parent PR rule

Do **not** merge parent PR #390 until all of these are true:

1. Shepherd uses official GSD Pi 1.11 unit/phase metadata as the routing authority.
2. Every mutating attempt runs in a disposable worktree with explicit promote/discard/recovery semantics.
3. Final readiness requires real artifact proof, exact-head continuity, Sol/high independent validation, ratification, and authority-gated outbox execution.
4. `awaiting_decision` is durable and restart-safe.
5. The GitHub question/reply broker is complete, idempotent, bounded, marker-owned, and authorization checked.
6. Recovery is classed, budgeted, and uses Sol/high recovery planning where needed.
7. Real `shepherd supervise` integration tests pass.
8. Merge-disabled Twenty and Asana canaries reach `final_human_gate`.
9. Only after canaries pass: Podman leftovers are removed, `.gsd/.planning` state is migrated, and obsolete files are cleaned.
10. Parent PR #390 remains human-gated for final merge to `main`.

## Required workflow before implementation

For each sub-issue or implementation slice:

1. Read `AGENTS.md`.
2. Read `.agents/agentic-delivery/references/required-skills-routing.md`.
3. Read `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
4. For runtime/RLM/Pi-agent/Podman/GitHub orchestration work, read `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.
5. Run the programming loop prompt path and record it in the issue artifacts:

   ```bash
   scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run
   ```

6. Update the phase plan, TDD ledger, verification checklist, and run-state artifact before production edits.
7. Follow RED -> GREEN -> refactor for every behavior-changing slice.
8. Commit only coherent green checkpoints.
9. Push only issue/PR branches; never push to `main`.
10. Stop at documented human gates.

## Required skills to record

At minimum, record these skills in the issue artifact or PR body for implementation slices:

- `gsd-core`
- `polymetrics-issue-delivery`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-context`
- `golang-concurrency`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-observability`

Add task-specific skills when applicable, for example `golang-database` for durable store work or `golang-documentation` for docs/CLI surface updates.

---

# Proposed sub-issue and PR split

Use stacked sub-PRs into `feat/372-gsd-pi-go-shepherd`. Keep one primary issue per PR and use `Refs #N` for stacked/incremental work. Parent PR #390 into `main` remains human-gated.

Recommended order:

1. Issue A — GSD 1.11 unit-registry compatibility and official phase model routing.
2. Issue B — Disposable per-attempt worktrees.
3. Issue C — Artifact proof, head continuity, Sol/high validation, ratification, and authority-gated outbox.
4. Issue D — Durable `awaiting_decision` and GitHub question/reply broker.
5. Issue E — Per-class recovery budgets and Sol/high recovery planning.
6. Issue F — Real `shepherd supervise` integration tests.
7. Issue G — Merge-disabled Twenty and Asana canaries.
8. Issue H — Post-canary Podman cleanup.
9. Issue I — Post-canary `.gsd/.planning` state migration.
10. Issue J — Post-canary obsolete file cleanup.

---

# Issue A — GSD 1.11 unit-registry compatibility and model routing

## Goal

Make Shepherd derive canonical unit identity, phase/lane, expected artifacts, and model routing from official GSD Pi 1.11 metadata instead of hard-coded command names.

## Key implementation areas

- `agent-runtime/shepherd/internal/gsd/`
- `agent-runtime/shepherd/internal/supervisor/`
- `agent-runtime/shepherd/cmd/shepherd/main.go`
- tests under `agent-runtime/shepherd/internal/gsd`, `agent-runtime/shepherd/internal/supervisor`, and `agent-runtime/shepherd/cmd/shepherd`

## Required behavior

- Load official GSD Pi 1.11 unit registry/phase metadata from the active runtime.
- Validate registry shape and version.
- Validate prompt-advertised tools against the official unit/tool registry.
- Fail closed on:
  - missing registry,
  - unknown unit,
  - unknown phase,
  - unknown lane,
  - missing or partial metadata,
  - symlink/path escape in registry discovery,
  - runtime contract mismatch,
  - model mismatch,
  - thinking mismatch.
- Route by official metadata:
  - planning / coordination / recovery planning / validation / UAT / ratification -> `openai-codex/gpt-5.6-sol`, `high`
  - implementation / execution / delegated subagent work -> `openai-codex/gpt-5.5`, `high`
- Prevent subagent model events from overwriting top-level model proof.
- Persist observed model/thinking evidence per unit attempt.

## RED tests

Add or extend tests proving current behavior is insufficient:

- Registry advertises a unit whose tool is unavailable -> `runtime_contract_mismatch`.
- Registry omits metadata for canonical next unit -> fail closed.
- Unit metadata maps validation to a non-Sol model -> fail closed.
- Unit metadata maps implementation to Sol/high when implementation should use GPT-5.5/high -> fail closed.
- Runtime prompt advertises a tool not allowed by registry -> `runtime_contract_mismatch`.
- Subagent emits model event inconsistent with parent -> child evidence rejected, parent proof preserved.
- Official metadata changes shape -> compatibility failure with typed evidence.
- Registry path attempts traversal or symlink escape -> fail closed.

## GREEN implementation

- Introduce a version-qualified registry loader.
- Introduce typed metadata structs for unit, phase, lane, expected artifacts, and routing class.
- Introduce a routing resolver that accepts official metadata and returns expected model/thinking.
- Wire resolver into `supervise` dispatch.
- Store observed model/thinking per attempt.
- Keep runtime compatibility checks side-effect free.

## Refactor constraints

- Do not infer canonical GSD workflow state from `.planning`.
- Treat official GSD Pi state and metadata as authoritative for unit routing.
- Keep Shepherd SQLite as controller truth, not a replacement for GSD Pi workflow truth.

## Verification

```bash
cd agent-runtime/shepherd
go test ./internal/gsd ./internal/supervisor ./cmd/shepherd
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/shepherd
make verify
```

---

# Issue B — Disposable per-attempt worktrees

## Goal

Ensure failed, interrupted, stale, or cancelled attempts cannot mutate the canonical issue worktree.

## Key implementation areas

- new `agent-runtime/shepherd/internal/workspace/`
- `agent-runtime/shepherd/internal/git/`
- `agent-runtime/shepherd/internal/store/`
- `agent-runtime/shepherd/cmd/shepherd/main.go`

## Attempt identity

Every mutating attempt is keyed by:

```text
repo
issue
generation
unit
attempt
base_head
```

## Attempt lifecycle

Allowed states:

```text
created
running
validated
promoted
discarded
retained_for_recovery
cleanup_pending
cleanup_complete
cleanup_blocked
```

## Required behavior

For each mutating unit attempt:

1. Create a disposable attempt worktree under a Shepherd-owned attempt root.
2. Bind the worktree to issue, generation, unit, attempt, and base head.
3. Run the unit only inside the attempt worktree.
4. Persist attempt state before process launch.
5. On success, validate output before promotion.
6. On failure, interruption, cancellation, or uncertainty, retain or discard according to policy.
7. Never delete unknown or non-Shepherd-owned worktrees.
8. Never allow failed attempt dirtiness to contaminate the canonical worktree.

## Promotion requires

- canonical head still equals attempt base head,
- attempt produced an expected candidate head or expected artifact-only output,
- attempt output is inside declared write scope,
- no disallowed dirty or untracked files exist,
- expected artifacts exist,
- GSD state advanced as expected,
- model proof is valid,
- Sol/high validation and ratification permit promotion where required,
- current lease/generation is still valid.

## Discard requires

- no live process owns the attempt worktree,
- no live child/subagent owns the attempt worktree,
- retained evidence is redacted and bounded,
- cleanup path is Shepherd-owned,
- cleanup is idempotent.

## Recovery requires

- mark attempt `retained_for_recovery`,
- persist failure class and concise reason,
- do not reuse retained worktree for new attempts,
- create a fresh worktree for retry,
- Sol/high recovery planner can inspect only bounded/redacted evidence.

## RED tests

- Failed attempt dirties files; canonical worktree remains unchanged.
- Cancelled attempt leaves canonical worktree unchanged.
- Interrupted process leaves attempt `retained_for_recovery` or `cleanup_pending`, not silently discarded.
- Promotion with stale canonical head is rejected.
- Promotion with out-of-scope file is rejected.
- Promotion with missing artifact is rejected.
- Restart reconciles an orphaned attempt worktree.
- Unknown worktree is never deleted automatically.
- Duplicate promotion is idempotent.
- Attempt cleanup cannot path-traverse outside the Shepherd attempt root.

## GREEN implementation

- Add `internal/workspace` worktree manager.
- Add attempt state persistence in store.
- Wire mutating supervise units through the worktree manager.
- Add promotion/discard/recovery transitions.
- Keep cleanup conservative and auditable.

## Verification

```bash
cd agent-runtime/shepherd
go test ./internal/workspace ./internal/git ./internal/store ./cmd/shepherd
go test ./...
go test -race ./...
```

## Human gate

Do not delete legacy, unknown, shared, or live worktrees without explicit human approval.

---

# Issue C — Artifact proof, head continuity, Sol/high validation, ratification, and authority-gated outbox

## Goal

Make “unit complete” and “ready for final gate” require durable, replayable proof rather than process exit alone.

## Proof model

Persist a proof record containing:

```text
issue
unit
generation
attempt
start_head
attempt_head
candidate_head
validated_head
expected_artifacts
artifact_hashes
gsd_state_before
gsd_state_after
model_observations
validator_model_observation
local_gate_results
ratification_result
outbox_effects_requested
outbox_effects_authorized
outbox_effects_executed
```

## Required behavior

A unit can advance only if:

- start head matches canonical head at attempt creation,
- candidate head is known,
- validation runs at exact candidate head,
- validator is `openai-codex/gpt-5.6-sol` with `high`,
- expected artifacts exist,
- expected artifacts hash to stored values,
- GSD Pi state advanced consistently,
- no live child/subagent remains,
- no out-of-scope diff exists,
- local gates required by metadata pass,
- ratification is stored,
- any external effect is authorized through Shepherd outbox.

## Outbox rules

Outbox is the only path for external write effects:

- PR summary update,
- GitHub decision question comment,
- branch push,
- PR creation/update,
- status publication,
- any future write-capable external effect.

Every effect must persist:

```text
effect_id
issue
unit
generation
head
authority_class
idempotency_key
requested_by
authorized_by_policy
executed_at
result
```

No worker directly mutates GitHub or Git outside approved ports.

## Authority classes

Suggested authority classes:

```text
local_read
local_write_attempt
local_promote
git_commit
git_push
github_comment
github_pr_update
github_pr_create
final_human_gate
forbidden_main_merge
```

`forbidden_main_merge` must never be executable by Shepherd.

## RED tests

- Zero-exit unit missing artifact fails.
- Artifact exists but hash changes after validation -> invalid.
- Validation runs on stale head -> invalid.
- Sol/high validation missing -> invalid.
- Validation model proof from GPT-5.5 -> invalid.
- Ratification missing -> invalid.
- Outbox request without authority grant -> blocked.
- Duplicate outbox execution is idempotent.
- Later commit invalidates prior validation.
- Worker attempts direct GitHub write bypassing outbox -> blocked or test fake detects forbidden path.

## GREEN implementation

- Add durable proof records.
- Add artifact hashing and expected artifact checks.
- Add exact-head validation binding.
- Wire Sol/high validator into final proof.
- Add ratification record and gate.
- Add outbox authorization and idempotent execution ledger.

## Verification

```bash
cd agent-runtime/shepherd
go test ./internal/store ./internal/authority ./cmd/shepherd
go test ./...
go test -race ./...
```

---

# Issue D — Durable `awaiting_decision` and GitHub question/reply broker

## Goal

Human questions become durable, generation-bound state, not transient terminal failures.

## Required state

Add or complete:

```text
RunAwaitingDecision
```

A decision request persists:

```text
request_id
issue
pr
unit
generation
head
question_kind
evidence_summary
options
recommended_option
safe_default
expires_at
github_comment_id
status
accepted_answer
accepted_by
accepted_at
consumed_at
```

## GitHub question contract

Question comment must include:

- stable request ID,
- issue/PR,
- canonical unit,
- generation,
- exact head,
- concise redacted evidence,
- bounded named options,
- recommended option when safe,
- expiry,
- safe default,
- `@karthik-sivadas`,
- exact syntax:

```text
/shepherd decide <request-id> <option>
```

## Reply acceptance rules

Accept only if:

- author is configured allowlisted human,
- comment is not from a bot,
- syntax is exact,
- request ID exists,
- generation matches,
- head matches,
- request is still open,
- reply is not edited after creation,
- option is valid,
- answer has not already been consumed.

Reject and record:

- stale replies,
- duplicate replies,
- unauthorized users,
- bots,
- edited comments,
- malformed syntax,
- expired requests,
- stale heads,
- stale generations.

## Request lifecycle

Allowed request states:

```text
open
published
answered
consumed
expired
cancelled
rejected
```

## RED tests

- Restart preserves pending request.
- Duplicate answer is consumed once.
- Unauthorized answer rejected.
- Bot answer rejected.
- Edited answer rejected.
- Stale generation rejected.
- Stale head rejected.
- Malformed answer rejected.
- One marker-owned question comment is reused, not duplicated.
- Expired request applies safe default only when allowed.
- Expired request without safe default remains blocked or escalates.

## GREEN implementation

- Add durable decision request store APIs.
- Add GitHub question publisher port.
- Add GitHub reply reader port.
- Add idempotent marker-owned comment handling.
- Add exact reply parser.
- Wire `RunAwaitingDecision` into supervise loop.

## Verification

```bash
cd agent-runtime/shepherd
go test ./internal/store ./internal/github ./cmd/shepherd
go test ./...
go test -race ./...
```

## Human gate

Live GitHub publishing tests require explicit sandbox approval. Do not request, print, or store secret values.

---

# Issue E — Per-class recovery budgets and Sol/high recovery planning

## Goal

Recover automatically only when safe, bounded, and authorized.

## Failure classes

At minimum:

```text
runtime_contract_mismatch
model_mismatch
thinking_mismatch
artifact_missing
artifact_invalid
stale_head
dirty_tree
write_scope_breach
dead_worker
silent_tool
interrupted
github_publish_uncertain
outbox_uncertain
validation_failed
ratification_failed
retry_exhausted
human_required
unknown
```

## Budget dimensions

Budget by:

```text
issue
unit
generation
head
failure_class
```

Persist:

```text
attempt_count
max_attempts
backoff
last_failure
last_recovery_plan
next_retry_at
exhausted_at
```

## Sol/high recovery planning

Use Sol/high recovery planning for complex recoverable classes:

- dead worker,
- silent tool,
- artifact missing,
- validation failed,
- GitHub publish uncertainty,
- outbox uncertainty.

Do not auto-recover:

- model mismatch,
- thinking mismatch,
- write-scope breach,
- stale head unless explicitly reconciled,
- unknown failure,
- exhausted budget,
- authority expansion,
- dependency changes,
- auth changes,
- secret handling,
- destructive requests.

## Recovery action types

Suggested typed actions:

```text
retry_same_unit
retry_after_backoff
run_recovery_plan
await_decision
block
final_human_gate
```

## RED tests

- Dead worker retries within budget.
- Silent tool retries with backoff.
- Artifact missing triggers Sol/high recovery plan.
- Validation failure triggers Sol/high recovery plan when reversible.
- GitHub publish uncertainty uses outbox idempotency before retry.
- Model mismatch blocks immediately.
- Thinking mismatch blocks immediately.
- Write-scope breach blocks immediately.
- Unknown failure awaits human decision.
- Exhausted budget enters `awaiting_decision`.
- Restart preserves budget state.
- Recovery plan cannot widen authority.

## GREEN implementation

- Add failure classifier table.
- Add durable per-class budget records.
- Add backoff calculation.
- Add recovery planner dispatch through Sol/high.
- Add fail-closed policy for unknown/unsafe classes.
- Wire budget exhaustion to `RunAwaitingDecision` when human decision is useful; otherwise `RunBlocked`.

## Verification

```bash
cd agent-runtime/shepherd
go test ./internal/supervisor ./internal/store ./cmd/shepherd
go test ./...
go test -race ./...
```

---

# Issue F — Real `shepherd supervise` integration tests

## Goal

Test the real command path, not just pure policy units.

## Layer 1 — Deterministic fake runtime

Use fake GSD/GitHub ports but real `shepherd supervise`.

Must cover:

- bootstrap,
- unit selection,
- attempt worktree creation,
- model routing,
- artifact proof,
- validation,
- ratification,
- outbox,
- awaiting decision,
- restart,
- final human gate.

Expected terminal state:

```text
final_human_gate
```

## Layer 2 — Restart/recovery integration

Kill/restart supervise during:

- running unit,
- pending outbox effect,
- awaiting decision,
- retained failed attempt,
- validation,
- final gate.

Assert:

- no duplicate dispatch,
- no duplicate GitHub comment,
- no duplicate outbox execution,
- no lost pending decision,
- no canonical worktree contamination,
- no stale validation reuse after head change.

## Layer 3 — Merge-disabled GitHub sandbox

Requires explicit approval.

Run with:

- real GitHub issue/PR comments,
- merge disabled,
- no credentialed connector checks,
- no parent merge,
- exact outbox idempotency,
- exact `awaiting_decision` resume behavior.

## RED tests

- Current supervise cannot complete fake runtime path to `final_human_gate` with disposable worktrees and durable decision broker enabled.
- Restart during awaiting decision currently loses or duplicates state.
- Restart during outbox uncertainty currently risks duplicate effect or lost effect.

## GREEN implementation

- Add gated integration package, for example `agent-runtime/shepherd/integration/`.
- Add fake runtime harness.
- Add fake GitHub broker harness.
- Add restart harness.
- Add merge-disabled sandbox harness behind explicit environment flag.

## Verification

Deterministic/local:

```bash
cd agent-runtime/shepherd
go test ./integration/... -run TestSuperviseFakeRuntime
go test ./integration/... -run TestSuperviseRestart
```

Full nested module:

```bash
cd agent-runtime/shepherd
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/shepherd
make verify
```

Live GitHub sandbox only after human approval:

```bash
POLYMETRICS_SHEPHERD_INTEGRATION=1 go test ./integration/... -run TestSuperviseGitHubSandbox
```

---

# Issue G — Merge-disabled Twenty and Asana canaries

## Goal

Prove Shepherd can drive real representative work to `final_human_gate` without manual per-unit commands.

## Shared canary constraints

- Merge disabled.
- No autonomous merge to `main`.
- No credential reads.
- No credentialed connector checks.
- No production external effects.
- All GitHub writes go through outbox.
- Human decisions use GitHub broker.
- Final state must be `final_human_gate`.
- Canaries must use host-local runtime by default.
- Podman assets must not be removed before these pass.

## Canary evidence to capture

For each canary, capture:

```text
issue
branch
parent branch
start head
final candidate head
units completed
attempt worktrees created
attempts promoted
attempts discarded
attempts retained_for_recovery
recovery plans used
questions asked
answers accepted
answers rejected
artifact proofs
validation model proof
ratification proof
outbox ledger
local gates
final_human_gate timestamp
```

## Twenty canary

Purpose:

- validates realistic issue orchestration,
- exercises artifact proof,
- likely exercises recovery and question broker,
- validates final human gate stop.

Pass condition:

```text
shepherd supervise ... -> final_human_gate
```

No manual unit command selection is allowed.

## Asana canary

Purpose:

- validates connector/CLI-adjacent workflow,
- tests write-scope protection,
- tests docs/verification parity if applicable,
- confirms the plan is implementable with an Asana implementation path.

Pass condition:

```text
shepherd supervise ... -> final_human_gate
```

No manual unit command selection is allowed.

## Required canary safety checks

Before running each canary:

- confirm branch and base,
- confirm merge-disabled target,
- confirm no credentialed connector check will run,
- confirm no secret values are requested or logged,
- confirm outbox dry-run/merge-disabled mode is active,
- confirm final parent merge authority is unavailable.

## Verification after each canary

```bash
cd agent-runtime/shepherd
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/shepherd
make verify
```

If repo-root files changed:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

---

# Issue H — Post-canary Podman cleanup

## Start condition

Do **not** start until:

- Twenty canary reached `final_human_gate`,
- Asana canary reached `final_human_gate`,
- evidence is recorded,
- human explicitly approves cleanup.

## Goal

Remove obsolete Shepherd-specific Podman leftovers while preserving unrelated project runtime assets.

## Important distinction

Do not remove general project runtime Podman assets if they are still used by:

- `pm runtime`,
- RLM agent mode,
- Temporal/Postgres/Dragonfly integration tests,
- unrelated runtime docs.

Only remove Shepherd-specific Podman leftovers proven obsolete by host-local canaries.

## Work

- Remove obsolete Podman-only Shepherd docs/scripts/config.
- Preserve host-local default in `agent-runtime/shepherd/README.md` and `shepherd.example.json`.
- Update examples.
- Add grep checks for stale Shepherd Podman default references.

## Verification

```bash
rg -n "podman|Podman" agent-runtime/shepherd .agents .pi docs website
cd agent-runtime/shepherd && go test ./...
```

Every remaining match must be intentional, current, and documented.

---

# Issue I — Post-canary `.gsd/.planning` state migration

## Start condition

Do **not** start until:

- Twenty canary reached `final_human_gate`,
- Asana canary reached `final_human_gate`,
- post-canary cleanup has explicit approval,
- no live Shepherd/GSD process owns the current state.

## Goal

Separate tracked durable evidence from local runtime state.

## Target rule

Tracked evidence moves to stable docs paths, for example:

```text
docs/agentic-delivery/issues/<issue>/<slug>/
docs/agentic-delivery/archive/
```

Runtime-local state becomes ignored:

```text
.gsd/
.planning/
```

## Required behavior

- Do not delete local runtime state blindly.
- Use `git mv` for tracked files.
- Add migration map.
- Preserve historical evidence.
- Do not commit SQLite DBs, sessions, raw prompts, raw tool output, secrets, or credentials.
- Update references in:
  - `AGENTS.md`,
  - `CLAUDE.md`,
  - `.agents/agentic-delivery/**`,
  - `.pi/**`,
  - `scripts/gsd`,
  - Shepherd docs/tests,
  - website/docs if applicable.

## Suggested target tree

```text
docs/agentic-delivery/
├── issues/
│   ├── 372/gsd-pi-go-shepherd/
│   └── 389/shepherd-hardening/
├── archive/
│   ├── planning-v1/
│   └── gsd-pi/
└── state-migration/
    ├── PATH-MAP.md
    └── tracked-manifest.json
```

## Suggested ignored runtime tree

```text
.gsd/
.planning/
.gsd-bootstrap-*/
.planning-bootstrap-*/
```

## Verification

```bash
git ls-files .gsd .planning
scripts/gsd doctor
scripts/gsd list
git diff --check
go test ./...
make verify
```

Expected after migration:

- no tracked runtime `.gsd`/`.planning` state,
- prompt adapter still works,
- issue evidence remains tracked elsewhere,
- historical links resolve through migration map.

---

# Issue J — Post-canary obsolete file cleanup

## Start condition

Do **not** start until:

- Twenty canary passed,
- Asana canary passed,
- Podman cleanup decision is complete,
- `.gsd/.planning` migration is complete or explicitly deferred.

## Goal

Remove or update stale references and obsolete scaffolding.

## Work

Remove or update stale references to:

- operator-driven Shepherd flow,
- old review artifacts,
- Podman-as-default Shepherd runtime,
- old `.gsd/.planning` tracked locations,
- obsolete canary scaffolding,
- stale runtime contract assumptions,
- stale GSD adapter paths.

## Verification grep gates

```bash
rg -n "claude-review-loop|@claude|Copilot review"
rg -n "Podman|podman" agent-runtime/shepherd .agents .pi docs website
rg -n "\.planning/phases|\.gsd/phases" AGENTS.md CLAUDE.md .agents .pi docs website scripts agent-runtime/shepherd
rg -n "operator.*select|manual.*unit|per-unit command" agent-runtime/shepherd .agents .pi docs website
```

Each remaining match must be one of:

- current and intentional,
- historical archive,
- documented exception.

## Verification

```bash
git diff --check
go test ./...
make verify
```

---

# Final integrated readiness gate for parent PR #390

Before marking PR #390 human-ready, all issue/PR slices above must be merged into the parent branch and the parent branch must pass integrated verification.

## Required local gates

Nested Shepherd module:

```bash
cd agent-runtime/shepherd
gofmt -w cmd internal
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/shepherd
make verify
```

Root repo:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Website/docs if touched:

```bash
cd website
pnpm run gen:website-data
pnpm run typecheck
pnpm run test:unit
pnpm run build
```

## Required Shepherd evidence

Record in the final parent PR body or linked evidence artifact:

- GSD 1.11 unit-registry compatibility proof.
- Official metadata model routing proof.
- Attempt worktree promote/discard/recovery proof.
- Artifact proof and hashes.
- Exact-head continuity proof.
- Sol/high independent validation proof.
- Ratification proof.
- Outbox idempotency proof.
- Awaiting decision restart proof.
- GitHub question/reply broker proof.
- Per-class recovery budget proof.
- Twenty canary to `final_human_gate`.
- Asana canary to `final_human_gate`.
- Cleanup/migration proof for Podman and `.gsd/.planning`, if performed.

## Final human gate

Only after all evidence is complete:

1. Update PR #390 body with full evidence.
2. Record unresolved risks or explicitly state none.
3. Keep PR #390 human-gated.
4. Do **not** merge to `main` autonomously.

---

# Implementation checklist summary

Use this as the high-level progress checklist.

- [ ] Issue A: GSD 1.11 unit-registry compatibility implemented and verified.
- [ ] Issue A: official phase metadata routes models correctly.
- [ ] Issue B: disposable per-attempt worktrees implemented and verified.
- [ ] Issue B: promote/discard/recovery semantics are durable and restart-safe.
- [ ] Issue C: artifact proof implemented and verified.
- [ ] Issue C: exact-head continuity implemented and verified.
- [ ] Issue C: Sol/high validation and ratification implemented and verified.
- [ ] Issue C: authority-gated outbox implemented and verified.
- [ ] Issue D: durable `awaiting_decision` implemented and verified.
- [ ] Issue D: GitHub question/reply broker implemented and verified.
- [ ] Issue E: per-class recovery budgets implemented and verified.
- [ ] Issue E: Sol/high recovery planning implemented and verified.
- [ ] Issue F: real `shepherd supervise` fake-runtime integration reaches `final_human_gate`.
- [ ] Issue F: restart/recovery integration passes.
- [ ] Issue G: merge-disabled Twenty canary reaches `final_human_gate`.
- [ ] Issue G: merge-disabled Asana canary reaches `final_human_gate`.
- [ ] Issue H: post-canary Shepherd Podman cleanup completed or explicitly deferred.
- [ ] Issue I: post-canary `.gsd/.planning` migration completed or explicitly deferred.
- [ ] Issue J: obsolete file cleanup completed.
- [ ] Parent branch integrated gates pass.
- [ ] PR #390 body updated with evidence.
- [ ] PR #390 stopped at human merge gate.
