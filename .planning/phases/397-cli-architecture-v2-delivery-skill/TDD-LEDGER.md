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

- Status: RED recorded 2026-07-23 before skill production files.
- Test: `scripts/tests/cli-architecture-v2-delivery-skill.sh`, invoked directly and through the
  focused `make cli-architecture-v2-skill-check` target. It MUST remain outside global
  `make verify` in this change.
- Required failures before implementation:
  - `cli-architecture-v2-delivery/SKILL.md`, three references, and `agents/openai.yaml` do not exist;
  - force triggers do not cover #397/PR #438, Cobra/Viper, events/TUI, observability, stacked
    orchestration, GSD/TDD, exact-head review, Shepherd, and parent readiness;
  - `AGENTS.md`, required-skills routing, task matrix, and Bubble Tea cross-link do not route the
    new skill;
  - local links cannot validate;
  - specialist-boundary and volatile-state contradiction markers are absent.
- RED command: `scripts/tests/cli-architecture-v2-delivery-skill.sh`.
- RED evidence: exit 1 with
  `AssertionError: missing required CLI Architecture v2 delivery skill file: .../SKILL.md`.

## Green: Program Router And References

- Status: GREEN recorded 2026-07-23.
- Target:
  - concise `cli-architecture-v2-delivery/SKILL.md` and `agents/openai.yaml`;
  - `state-and-dependency-model.md`;
  - `phase-delivery-checklist.md`;
  - `parent-integration-and-review.md`;
  - dated #397 source/gap evidence outside evergreen rules;
  - force-trigger, required routing, task matrix, and Bubble Tea cross-link entries;
  - focused shell validation plus focused Make target only.
- Focused commands:
  - `scripts/tests/cli-architecture-v2-delivery-skill.sh` — pass;
  - `make cli-architecture-v2-skill-check` — pass;
  - `scripts/tests/pi-model-routing.sh` — pass.
- GREEN evidence: content/link contract and YAML parsing passed; all 10 routed Pi agent model checks
  passed; focused check remains outside global `make verify`.

## Refactor

- Status: in progress after focused GREEN.
- Remove any copied queue, SHA, model, bot availability, or issue-status prose from the evergreen
  skill.
- Keep TUI implementation guidance in `bubble-tea-tui-design` and Go mechanics in routed Go skills.
- Verify #419-like human deferrals use `deferred_by_human`, parent/main merge remains human-gated,
  and active review constraints override configured reviewer routes without being hard-coded into
  evergreen prose.

## Independent Exact-Head Review And Fixes

- Reviewed head: `d81163f703c25266481a67e1072ab1fe5b09b7aa` against parent
  `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`.
- A dedicated reviewer process could not start because its isolated provider lacked authentication;
  no external review bot was substituted.
- A fresh independent verifier completed the diff/acceptance review and found:
  1. Track C verification claimed performance routing, but the skill omitted `golang-benchmark` and
     `golang-performance`.
  2. The focused script's Darwin `mktemp` template was not concurrency-safe.
  3. Stacked push/PR delivery remained pending, as expected before shipping.
- Disposition: added conditional benchmark/performance routing and assertions; changed the checker
  to use a unique temporary directory; proved simultaneous direct and Make invocations pass. The PR
  gate remains open for the no-mistakes lifecycle.
- A fresh exact-head review is required after the fix commit.
