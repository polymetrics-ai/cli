# Phase 462 Plan — Terminal UI research and design gate

Issue: #462
Parent: #397
Branch: `docs/462-terminal-ui-design-research`
Starting commit: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`
Classification: documentation, planning, and repo-local skill only; no production Go behavior.

## Objective

Freeze an evidence-backed Bubble Tea interaction and visual design contract before production
work starts in #408, #409, #411, #412, #414, #416, or #418. Give Pi/GSD workers one required
skill and prompt that makes Vim navigation, responsive layout, chart safety, accessibility,
plain/JSON parity, and dependency gates testable.

## GSD and skills

- `scripts/gsd doctor` passes with 69 registered commands.
- `scripts/gsd prompt plan-phase 462 --skip-research` generated a 10,704-byte official prompt;
  this plan executes it locally because the requested primary-source and hands-on research is
  already complete.
- Required skills used: `github-issue-first-delivery`, `gsd-plan-phase`, `golang-how-to`,
  `golang-cli`, `golang-testing`, `golang-documentation`, `golang-security`, `skill-creator`,
  and the newly authored repo-local `bubble-tea-tui-design`.
- The available `opentui` skill was inspected and rejected as implementation authority because
  it targets Bun/Zig/React/Solid rather than this repository's Go/Bubble Tea v2 stack.

## Scope and ownership

Owned files:

- `docs/design/terminal-ui-research-and-design-system.md`
- `docs/design/tui-ux-design.md`
- `docs/adr/0003-interactive-tui-layer.md`
- `docs/plans/cli-architecture-v2-improvement-plan.md`
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md`
- `.agents/skills/bubble-tea-tui-design/**`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- delegated parent planning/traces and this issue's phase artifacts
- live issue planning sections for #397, #408, #409, #411, #412, #414, #416, and #418

No `cmd/**`, `internal/**`, `go.mod`, `go.sum`, generated CLI help, website page, connector
definition, credential, remote write, or production TUI implementation is in scope.

## RED → GREEN → refactor tasks

1. **RED — contract inventory**
   - Prove the base commit lacks the research document, Bubble Tea design skill, GSD phase,
     modal key contract, chart dependency decision, and Pi TUI worker prompt.
   - Record the exact failure evidence in `TDD-LEDGER.md`.
2. **GREEN — evidence and normative design**
   - Record primary-source and isolated interaction findings for every requested application.
   - Freeze the operator-workspace reference, Normal/Filter/Edit modes, responsive classes,
     visual hierarchy, motion policy, charts/dashboard grammar, and Bubble Tea architecture.
3. **GREEN — reusable worker instructions**
   - Create and validate `.agents/skills/bubble-tea-tui-design` with focused references.
   - Route it from required skills and require it in all TUI Pi/GSD prompts.
4. **GREEN — program integration**
   - Update the ADR, source plan, execution prompt, roadmap, issue backlog, and Pi prompt trace.
   - Create query-chart child issue #463; keep `ntcharts/v2` behind an explicit human gate.
   - Update live affected issue bodies and parent status without overwriting unrelated content.
5. **REFACTOR — consistency and verification**
   - Remove contradictions between 80×24 enhancement and compact/guard behavior.
   - Check links/references, skill validation, issue mentions, Markdown whitespace, GSD health,
     dependency/scope diffs, and repository docs gates.

## Acceptance checklist

- [x] Research contract names an appropriate primary reference and adopt/adapt/avoid decisions.
- [x] Local isolated versions, interactions, screenshots, and chart compatibility are recorded.
- [x] Modal Vim-style navigation never steals printable input and has non-Vim alternatives.
- [x] Responsive, focus, help, motion, and accessibility rules are explicit.
- [x] Query chart/dashboard grammar retains exact table/text representation and read-only safety.
- [x] NTCharts remains a proposed dependency with a separate human gate.
- [x] Repo-local skill is created and routed.
- [x] GSD roadmap/backlog/execution/Pi prompts require the design gate and skill.
- [x] Live production UI issues point to the contract, skill, and exact RED matrix.
- [x] Targeted documentation, skill, GSD, and scope verification passes.
- [x] Changes are committed, pushed, and opened as stacked PR #465 to the parent branch.

## Review correction plan — 2026-07-20

Branch: `docs/462-terminal-ui-design-review-fixes`
Base: `feat/cli-architecture-v2` at `c91b90cf9671b5caabc0ef4ec24d81897f870458`
Prior delivery context: PR #465, head `6853fee28e0208381b49931fb1f5dfec42ee50ef`, squashed into
parent as `a5474bcb`; review coverage remains blocked because Claude is disabled and Copilot quota
was exhausted. New accepted-correction PR will target the parent branch; do not merge.

### Loaded skills and routing

- `gsd-core` via `.pi/skills/gsd-core/SKILL.md`.
- `caveman` via `.agents/skills/caveman/SKILL.md` for compact handoff only.
- `bubble-tea-tui-design` plus all four references.
- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-documentation`, `golang-security`.
- Additional routed context loaded for TUI docs: `golang-safety`, `golang-context`,
  `golang-concurrency` from required-skills routing.
- `.pi/skills/go-implementation/SKILL.md` is absent; used available repo/global skill paths and
  recorded the mismatch here and in the TDD ledger.

### Docs-only correction slices

1. **Planning checkpoint** — reopen this phase's `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`,
   `SUMMARY.md`, `PROMPTS.md`, `RUN-STATE.md`, and `RUN-STATE.json` with the accepted review
   corrections, manual universal-loop fallback, scope, checks, and review blocker before editing
   design/program docs.
2. **RED inventory** — keep grep evidence for current contradictions:
   - bare `pm query` wording incorrectly launches a TUI from a bare namespace;
   - guided reverse text prints/teaches approval tokens;
   - affected TUI dependency rows omit direct `#462`/`D-TUI` blockers;
   - phase status claims delivery/automated-review pending instead of provisional/review blocked;
   - query export lacks the typed path confinement/overwrite/noninteractive contract.
3. **GREEN docs correction** — update only delegated docs and phase artifacts:
   - require bare `pm query` and bare `pm reverse` to render contextual help/subcommand summaries
     and exit 0; interactive surfaces use explicit documented subcommands (`pm query grid`,
     `pm reverse guide`);
   - mark approval tokens as sensitive one-time authorization values that may live only
     ephemerally in memory through plan → preview → approval → execute and are never printed,
     logged, transcripted, screenshot, accessibility-output, JSON-output, or shell-equivalent text;
   - encode #462/D-TUI directly in each affected TUI row (#408, #409, #411, #412, #414, #416,
     #418, and #463 where listed) across roadmap/backlog/prompt rosters;
   - state #462 as provisionally integrated / review blocked with PR #465, its head SHA, Claude
     disabled, Copilot quota exhausted, human fallback, and correction PR #467 starting
     head/status captured explicitly;
   - add the read-only query export path contract: clean/confined project-scoped paths, reject
     control characters/traversal/broad paths/symlink races, no overwrite by default, TTY
     confirmation, noninteractive `--force`, sanitized command echo, exact `--no-input` guidance,
     and no generic file-write/SQL-write boundary change.
4. **Terminal evidence checkpoint** — run docs-contract greps, link/marker checks,
   `scripts/gsd doctor`, `git diff --check`, exact no-forbidden-scope diff, skill quick validation,
   Markdown/YAML/JSON checks, and `make docs-check` if feasible. Commit/push the evidence.

### Commit/push checkpoints

- `docs(gsd): plan terminal ui review corrections` — phase artifact reopen only.
- `docs(ui): apply terminal design review corrections` — delegated docs/program corrections.
- `docs(gsd): record terminal ui correction verification` — final verification/status evidence.

## Correction PR #467 accepted review findings — 2026-07-20

Branch remains `docs/462-terminal-ui-design-review-fixes` at start head
`e8286ea83a76ac2c6f6257c6e2d40fd21af81640`, based on `feat/cli-architecture-v2` at
`c91b90cf9671b5caabc0ef4ec24d81897f870458`; do not merge. Starting PR state to record in this
phase: correction PR #467 open at that head with CI green at that head; review status is
human/parent pending rather than a generic pending placeholder. Git/GitHub remain the current
source of truth after this starting snapshot, so this run must not invent a self-referential final
head.

### Loaded skills and routing for this bounded correction

- Re-read `AGENTS.md`, issue #462, GSD runtime contracts, this phase's `PLAN.md`,
  `TDD-LEDGER.md`, and `VERIFICATION.md` before edits.
- `gsd-core`, `caveman`, `bubble-tea-tui-design` plus four references.
- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-documentation`, `golang-security`,
  `golang-safety`, `golang-context`, `golang-concurrency`, and `golang-error-handling` for CLI/TUI
  docs that describe stdin/stdout gating, cancellation, secrets, paths, and tests.
- `.pi/skills/go-implementation/SKILL.md` remains absent; use available repo/global skills and
  record the mismatch in the ledger/handoff.

### Bounded correction slice

1. **Planning checkpoint** — reopen `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`,
   `PROMPTS.md`, `RUN-STATE.json`, and `RUN-STATE.md` for accepted PR #467 findings before
   changing delegated source/design/skill docs.
2. **RED inventory** — docs-contract grep must fail on stdout-only or ambiguous TUI gates and on
   missing future RED markers for `stdin-piped+stdout-TTY` and `stdout-piped` fallback.
3. **GREEN docs correction** — align every delegated source, skill/reference, roadmap, backlog,
   and worker/execution prompt mention so Bubble Tea/Huh/prompt activation requires **both stdin
   and stdout TTY** plus the existing `--json`/`--plain`/`--no-input`/`PM_NO_TUI`/`CI`/`TERM=dumb`
   disables. With piped or non-TTY stdin, fall back to deterministic plain/noninteractive behavior;
   never consume scripted stdin unexpectedly, hang, or bypass through `/dev/tty`.
4. **Future RED test contract** — require production TUI issues to add tests for
   `stdin-piped+stdout-TTY` fallback, `stdout-piped`, `CI`, `--json`, `--plain`, and `--no-input`,
   while preserving explicit `pm query grid`, `pm reverse guide`, read-only query export, reverse
   approval-token secrecy, and accessibility/plain behavior.
5. **State honesty** — update run-state/summary/checklists to say #467 was open at the starting
   head with CI green and human/parent review pending; replace open-PR next steps with local
   finding disposition, human review, and parent integration gates.
6. **Verification checkpoint** — rerun contradiction grep, JSON parse, skill validation,
   direct dependency/token/export contract checks, `git diff --check`, exact scope check,
   `scripts/gsd doctor`, and `make docs-check`; commit/push planning/docs fix/evidence.
