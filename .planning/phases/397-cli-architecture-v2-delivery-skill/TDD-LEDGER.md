# TDD Ledger

Phase: `397-cli-architecture-v2-delivery-skill`

## Superseded Baseline

- The original plan assumed `origin/main` and a missing generic Go TUI skill.
- Captain correction established parent issue #397 and `feat/cli-architecture-v2` as the target.
- The parent already contains `bubble-tea-tui-design` plus four focused references and two terminal
  design documents from issue #462/PRs #465, #467, and #468.
- Therefore a `.agents/skills/go-tui-development/` skill would duplicate an existing authority and
  is forbidden by the corrected scope.

## Corrected Baseline

- No project skill force-triggers on issue #397, PR #438, CLI Architecture v2, or all three program
  tracks.
- Required routing covers generic Go/CLI work and Bubble Tea work, but not program-level track,
  stacked-integration, exact-head evidence, or parent-readiness routing.
- The task-skill matrix has no `cli_architecture_v2` group or `cli-architecture-v2` task type.
- Volatile parent state exists in issue/PR and phase artifacts and must not be copied into an
  evergreen skill.

## Red: Program Delivery Skill Contract

- Status: planned.
- Test: test-only Go package, disjoint from product CLI/UI packages.
- Required failures before implementation:
  - `cli-architecture-v2-delivery/SKILL.md` and focused references do not exist;
  - force triggers do not cover #397/PR #438, Cobra/Viper, events/TUI, observability, stacked
    orchestration, GSD/TDD, exact-head review, and parent readiness;
  - routing does not name the new skill or task group;
  - YAML and local links cannot validate;
  - specialist-boundary and volatile-state contradiction markers are absent.
- Command: `go test ./internal/agentdocs` (or the final disjoint test-only package selected before
  RED).

## Green: Program Router And References

- Status: pending.
- Target:
  - concise `cli-architecture-v2-delivery/SKILL.md`;
  - phase/track routing and dependency rules;
  - machine-readable CLI/parity invariants;
  - stacked branch/write-scope/integration procedure;
  - GSD/TDD/exact-head evidence contract;
  - final parent readiness/human-gate checklist;
  - dated #397 research/gap artifact;
  - required routing and task matrix entries.
- Focused command: `go test ./internal/agentdocs`.

## Refactor

- Status: pending.
- Remove any copied queue, SHA, model, bot availability, or issue-status prose from the evergreen
  skill.
- Keep TUI implementation guidance in `bubble-tea-tui-design` and Go mechanics in routed Go skills.
- Verify #419 remains deferred, parent/main merge remains human-gated, and review-bot exclusions for
  this task are recorded only in dated delivery evidence.
