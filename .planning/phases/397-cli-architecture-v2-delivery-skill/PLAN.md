# PLAN: CLI Architecture v2 Delivery Skill

## Objective

Add the smallest non-duplicative project skill and routing needed to finish parent issue #397 across
its Cobra/Viper, events/TUI, observability, help/parity, stacked-delivery, and final-readiness tracks.
The skill routes to existing specialist contracts; it does not replace them or implement product
behavior.

## Scope Correction And Delivery Context

- Original task branch: `fm/cli-go-tui-development-skill-r1`.
- Original base: `origin/main` at `873cd7b251f70c4a35a607a0d4e86051ea0fbd15`.
- Captain correction: CLI Architecture v2 issue #397 is authoritative; the correct stacked base is
  `feat/cli-architecture-v2`, not `main`.
- Current parent base after `git fetch origin feat/cli-architecture-v2 --prune`:
  `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`.
- The existing plan commit was preserved by rebasing onto that parent head. This explicit plan
  revision supersedes the generic `go-tui-development` outcome.
- Final PR base: `feat/cli-architecture-v2`; use `Refs #397`; never merge parent PR #438.

## Existing Authorities To Reuse

- `.agents/skills/bubble-tea-tui-design/` owns Bubble Tea interaction, accessibility, safety, and
  TUI testing. Do not add a generic or duplicate Go TUI skill.
- `.agents/skills/caveman/SKILL.md` owns compact orchestration prose only.
- `.pi/skills/gsd-core/SKILL.md`, issue-first contracts, and parent orchestration contracts own the
  delivery lifecycle.
- ADR 0002 owns Cobra/Viper migration; ADR 0003 owns events/TUI; ADR 0004 owns observability.
- CLI help/docs/website parity and required Go skill routing remain authoritative.

## GSD And Orchestrator Activation

- `scripts/gsd doctor` and `scripts/gsd list`: passed.
- `scripts/gsd sources programming-loop`: unavailable because the 69-command registry does not
  expose `programming-loop`.
- `scripts/gsd prompt quick --full ...`: generated and executed inline before the scope correction.
- Corrected workflow: manual universal-loop fallback using this plan, `TDD-LEDGER.md`, and
  `VERIFICATION.md`.
- Direct project subagent `parent-issue-orchestrator` was unavailable in this Pi runtime. The
  project `pm-planner` performed a read-only bounded intake; runtime result is
  `not_spawned_runtime_capability_missing`, while firstmate's explicit assignment supplies the
  bounded delegation. The live parent orchestrator retains queue, shared artifact, integration,
  review-route, and readiness ownership.

## Required Skills Used

- `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-concurrency`, `golang-context`, `golang-safety`, `golang-security`,
  `golang-code-style`, and `golang-documentation`.
- Track-specific routing inspected: `golang-spf13-cobra`, `golang-spf13-viper`,
  `golang-observability`, `golang-benchmark`, `golang-performance`, and
  `bubble-tea-tui-design`.
- `no-mistakes` is the final authoritative validation/stacked-PR path. Captain explicitly excludes
  Claude and GitHub Copilot review for this task.

## Work Slices

1. Audit the live parent branch, PR #438, issue #397/subissues, integrated phase artifacts, current
   code/dependencies, routing, and existing skills.
2. Incorporate the separate read-only #397 audit's `skill-gap-spec.md` before freezing design.
3. Retain only research that informs #397 delivery; compare alternative TUI frameworks only to
   verify that issue #462's existing Bubble Tea decision does not need reopening.
4. Add `.agents/skills/cli-architecture-v2-delivery/` only if the audit confirms no equivalent.
5. Keep volatile queue/head/review status in a dated project artifact, never evergreen skill prose.
6. Route any #397 phase through the delivery skill and then to track specialists in
   `required-skills-routing.md` and `task-skill-matrix.yaml`.
7. Add a focused test-only contract for frontmatter, force triggers, internal links, YAML routing,
   contradiction markers, and non-duplication of `bubble-tea-tui-design`.
8. Run focused and repository gates, update evidence, commit, and hand off to firstmate before the
   no-mistakes shipping stage.

## Red / Green / Refactor

- **Red:** contract test expects `cli-architecture-v2-delivery`, its five focused references,
  required #397/PR #438/track triggers, routing entries, resolvable links, and preserved specialist
  boundaries before those files exist.
- **Green:** add the smallest complete orchestration/router skill, references, dated gap artifact,
  and two routing updates.
- **Refactor:** remove copied queue/head/status and specialist implementation detail; verify no
  contradiction with ADRs, parent contracts, `bubble-tea-tui-design`, or captain's no-review-bot
  route.

## Write Scope

Expected:

- `.agents/skills/cli-architecture-v2-delivery/**`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/matrices/task-skill-matrix.yaml`
- one disjoint test-only validation package
- `.planning/phases/397-cli-architecture-v2-delivery-skill/**`

Do not edit shared `.planning/phases/397-cli-architecture-v2-orchestration/**`, parent queue/state,
`AGENTS.md`, product CLI/UI code, help/manual/website surfaces, dependencies, or GitHub issues.

## Hard Boundaries

- No product feature or runtime dependency.
- No duplicate generic TUI skill and no rewrite of `bubble-tea-tui-design` without a verified gap.
- No queue arbitration, worker spawning, sub-PR integration, parent PR readiness, or parent-branch
  push owned by this worker.
- No secrets, credentialed connector checks, runtime lifecycle, generic write tools, or reverse ETL
  execution.
- #419 remains explicitly deferred; this task grants no dependency approval.
- No Claude or GitHub Copilot invocation.

## Commit Checkpoints

1. Original plan preserved on the original base.
2. Scope correction and parent-base evidence (this revision).
3. Red skill/routing contract test.
4. Green skill, references, routing, and dated gap artifact.
5. Verification/refactor evidence and final committed handoff.
