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
