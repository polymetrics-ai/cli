# Phase 18 design refinement plan

Issues: #416 and child #469
Delivery PR: #468 (docs/design only)
Implementation: future isolated issue branches; no production Go in this plan

## Objective

Turn the user-testing findings around unintuitive credential and connection commands into one
implementation-ready Bubble Tea contract, one bounded setup issue, and synchronized GSD/Pi docs.

## Required process and skills

- `scripts/gsd doctor` and `scripts/gsd prompt ui-phase 18 --text` were run.
- Native GSD phase discovery did not find Phase 18 because this sibling roadmap stores it as an
  issue row; the documented manual-GSD fallback was used and recorded.
- The GSD UI researcher produced `18-UI-SPEC.md`; the UI checker initially blocked four generic CTA
  labels, the researcher revised them, and the checker returned `UI-SPEC VERIFIED` with all six
  dimensions passing.
- Skills loaded: `bubble-tea-tui-design`, `gsd-ui-phase`, `github-issue-first-delivery`,
  `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, and
  `golang-spf13-cobra`.

## Slices

1. RED: inventory ambiguous behavior, prompt gating, secret entry, duplicate recovery, and agent
   invocation gaps in the existing Phase 18 contract.
2. GREEN: research primary terminal/CLI sources and freeze activation, state, accessibility,
   secret, duplicate, and automation contracts in `18-UI-SPEC.md`.
3. GREEN: split #469 from #416; update GitHub parent/blocked-by edges and synchronize roadmap,
   issue backlog, Pi prompts, execution prompt, ADR, design docs, and the repo-local TUI skill.
4. REFACTOR: remove duplicate/contradictory wording, verify markers and scope, run docs/GSD/skill
   gates, push PR #468, and leave merge to human review.

## Future implementation order

- #416 and #469 may start only after #462/PR #468 is reviewed and integrated, their direct blockers
  are closed, and the parent orchestrator assigns isolated non-colliding worktrees.
- Each issue runs its own `gsd-programming-loop` with RED → GREEN → refactor checkpoints.
- #469 must implement the complete activation matrix before form rendering, then credential setup,
  connection setup, duplicate-race recovery, accessibility, and CLI/docs/website parity.
- #417 help/man convergence and #418 accessibility convergence follow #469.

## Human gates

- No dependency change, plaintext secret flow, credentialed external test, overwrite behavior,
  automated issue merge, stacked PR merge, or parent-to-main merge is authorized here.
- PR #468 requires human review because Claude is disabled and Copilot quota is exhausted for the
  current review window.
