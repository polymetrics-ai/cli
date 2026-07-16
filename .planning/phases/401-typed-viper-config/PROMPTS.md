# Issue 401 Prompts — Typed Viper Configuration

## Kickoff

Task: Execute polymetrics-ai/cli#401 as one bounded mutating worker for parent #397.

Downstream artifact: `.planning/phases/401-typed-viper-config/PLAN.md`; PR #441 opened; prior review-fix pushed; final website caveat fixed locally.
Verification result: final caveat local gates passed; website package scripts blocked by missing `website/node_modules` where applicable; commit/push and PR body update pending; human/parent review fallback pending.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 401 --skip-research >/tmp/gsd-plan-phase-401.prompt
scripts/gsd prompt programming-loop init --phase 401 --dry-run >/tmp/gsd-programming-loop-401.prompt
```

`programming-loop` adapter result:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Fallback prompt loaded: `.pi/prompts/pm-gsd-loop.md`.

## Active fallback prompt

Manual GSD programming loop from `.pi/prompts/pm-gsd-loop.md`:

- Read required repo/GSD contracts and issue acceptance.
- Plan before production edits.
- Capture red test or validation evidence before behavior changes.
- Commit and push coherent green slices.
- Do not skip TDD, verification, review disposition, or human gates.

## Worker prompt source summary

- Issue: #401 `feat(config): add typed Viper configuration`.
- Parent: #397 / parent PR #438.
- Dependency integrated: #400 via PR #440, parent commit `8900db141cc289b65491365d2ebcab490af57789`.
- Branch: `feat/401-typed-viper-config`.
- Allowed scope: `internal/config/**`, focused tests, minimal CLI config load/bind/error mapping, ADR-approved Viper dependency delta, config docs/website, issue-local planning artifacts.
- Review route: Claude disabled manually; Copilot quota exhausted; record human/parent fallback pending.
