# Issue #397 PM Orchestrator Scope Extension Plan

Status: active
Owner: PR #495 Wave 1 integration worker
Branch: `chore/cli-architecture-v2-wave1-parent-sync-r1`
Stacked base: `feat/cli-architecture-v2`
Draft PR: #495
Starting head: `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`
Scope: canonical current/forward PM/Pi parent orchestration documentation, prompts, role agents, state schema, and focused validation. No #408 or product implementation.

## Authority and invariants

Captain approved an additive extension to PR #495. Preserve the reviewed synchronization history and use ordinary commits only: no amend, rebase, reset, stash, force-push, or merge. Keep parent PR #438 draft/human-only and PR #493 separate.

The canonical forward route must have one owner and one review chain:

1. `/pm-orchestrate` is the active parent owner when the repo-local GSD registry lacks `programming-loop`.
2. The PM owner reconciles durable state and runs PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE without inventing an unavailable command.
3. Mutating workers are isolated; critical-path inline work is recorded honestly.
4. `pm-verifier` verifies the exact candidate head.
5. A fresh-context local Codex `pm-reviewer` reviews the exact base/head; every finding is dispositioned and any changed head is re-reviewed.
6. Independent Shepherd trajectory validation passes after exact-head review and before integration.
7. CI and human merge authority remain separate gates. Claude and GitHub Copilot are not required, requested, or fallback coverage for this PM route.

## GSD and skills

- `scripts/gsd doctor`: pass.
- `scripts/gsd list`: 69 commands.
- `scripts/gsd sources plan-phase` and `scripts/gsd sources code-review`: pinned registry/lock/official docs resolved.
- `scripts/gsd sources programming-loop`: unavailable (`unknown GSD command or prompt: programming-loop`).
- Lifecycle for this extension: PM-owned manual PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE.
- Loaded/consulted: `cli-architecture-v2-delivery` at PR #493 head without absorbing it, `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-spf13-cobra`, `golang-spf13-viper`, and `golang-documentation`.

## PR #493 collision boundary

Fetched PR #493 head: `e21e56339390c5e1946eb4cfaf276eb80a889f29`.

Do not edit any PR #493-owned path:

- `AGENTS.md`
- `Makefile`
- `.agents/agentic-delivery/matrices/task-skill-matrix.yaml`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/skills/bubble-tea-tui-design/SKILL.md`
- `.agents/skills/cli-architecture-v2-delivery/**`
- `.planning/phases/397-cli-architecture-v2-delivery-skill/**`
- `scripts/tests/cli-architecture-v2-delivery-skill.sh`

Those routing/skill files were inspected. PR #495 owns the canonical orchestration route they may reference after Wave 1 integrates.

### Subsequent PR #493 migration gate

After Wave 1 lands, PR #493 must merge the resulting parent normally, reconcile its owned
`AGENTS.md`/required-skills/task-matrix/skill guidance to this canonical PM route, rerun its focused
validation, and integrate before another CLI Architecture v2 implementation worker starts. Until
then, record `not_spawned_dependency_blocked`; do not follow contradictory legacy routing, duplicate
PR #493 files in PR #495, or claim the routing migration is complete.

## Allowed write scope

- `.agents/agentic-delivery/README.md`
- `.agents/agentic-delivery/contracts/{issue-agent-contract,parent-orchestrator-contract}.md`
- new PM-specific worker-handoff and code-review-disposition contract templates; keep generic bot-era templates unchanged
- `.agents/agentic-delivery/workflows/{automated-review-routing-loop,claude-review-loop,codex-active-orchestration-loop,gsd-universal-runtime-loop,parent-issue-orchestration-loop,pi-active-orchestration-loop,pi-autonomous-orchestration-loop,stacked-parent-subissue-workflow}.md`
- new `.agents/agentic-delivery/workflows/local-codex-review-loop.md`
- `.agents/agentic-delivery/prompts/shepherd-validator-prompt.md`
- new `.agents/agentic-delivery/prompts/local-codex-review-prompt.md`
- `.agents/agentic-delivery/agents/coordination/parent-issue-orchestrator.agent.yaml`
- `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.pi/README.md`, `.pi/prompts/{pm-orchestrate,pm-auto-loop,pm-gsd-loop,pm-review-loop}.md`
- relevant `.pi/agents/pm-*.md` role corrections
- `docs/prompts/universal-programming-loop-prompts.md`
- `scripts/tests/pi-model-routing.sh`
- new `scripts/tests/pm-orchestrator-contract.sh` and focused review-state fixtures
- this phase directory, authoritative `.planning/traces/cli-architecture-v2-orchestration-state.yaml`, and narrow #397/Wave 1 machine-state/summary evidence updates

## Slices

### 1. RED — focused forward-route contract

Add a focused validation script and wire it through the existing Pi model-routing check (not the PR #493-owned Makefile). It must fail against the current route because canonical PM files require unavailable `programming-loop`/Claude/Copilot paths and lack an exact-head local Codex workflow.

### 2. GREEN — canonical ownership and review route

Add one local Codex review workflow/prompt. Update runtime-neutral parent/stacked contracts plus Pi/Codex adapters so PM orchestration owns the unavailable-command fallback, exact-head local Codex review is required, Shepherd remains independent, and human merge authority is unchanged.

### 3. REFACTOR — adapters, roles, migration pointers

Update PM prompts/agents/operator docs and state schema consistently. Retain Claude/Copilot documents and the legacy disposition agent only as explicitly deprecated, non-PM historical routes that point to the canonical local Codex workflow. Do not globally replace truthful historical phase records.

### 4. Correction round 2 — authoritative gate, transitive templates, correction lineage

After the first corrected exact-head re-review, make the PR #493 post-Wave1 migration gate durable in the authoritative #397 queue, route all current PM required-reading to PM-specific handoff/disposition templates, and align the autonomous driver with stable exact-base/candidate-lineage correction counters. Add RED assertions for each finding before guidance/state edits. Preserve legacy fields/templates as read-only input and do not edit PR #493-owned paths.

### 5. VERIFY

Run the focused contract first, then:

```bash
scripts/tests/pm-orchestrator-contract.sh
scripts/tests/pi-model-routing.sh
git diff --check
gofmt -w cmd internal
git diff --exit-code -- cmd internal
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
go mod verify
go mod tidy -diff
make verify
```

Also verify YAML/JSON parsing, PR #493 path disjointness, no dependency delta, no CLI/help/docs/website product applicability, and no historical phase rewrite.

### 6. REVIEW / INTEGRATE

Commit and push coherent additive checkpoints. At the final exact head, run a fresh-context read-only local Codex review bound to parent base and task head, disposition every finding, rerun affected gates/review after any change, then run independent Shepherd trajectory validation. Update draft PR #495 title/body and inspect checks/reviews. Never request Claude/Copilot, mark ready, merge PR #495, or edit/merge PR #438.

## Checkpoints

1. plan-only checkpoint;
2. focused RED validation checkpoint;
3. canonical workflow/role/docs GREEN checkpoint;
4. verification/evidence checkpoint;
5. review-fix checkpoint if needed;
6. final exact-head PR evidence.
