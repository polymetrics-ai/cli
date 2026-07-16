# Issue 400 Summary — Cobra Router Shell

Status: planning/TDD setup in progress.

## Scope summary

Replace top-level handwritten CLI switch with Cobra router shell only. Existing handlers/parsers remain underneath `DisableFlagParsing` wrappers. `cli.Run(args, stdout, stderr) int` stays unchanged. Dynamic connector passthrough, hidden extract/worker behavior, help/manual text, JSON envelopes, stdout/stderr split, and exit-code taxonomy must remain byte-identical.

## Current evidence

- Parent dependency #399 integrated via PR #439, parent commit `379cb5015335ff7c9b20e5bb780952ead22c53b2`.
- Branch `refactor/400-cobra-router-shell` starts at `origin/feat/cli-architecture-v2`.
- GSD adapter healthy; `programming-loop` command unavailable (`scripts/gsd: unknown GSD command: programming-loop`), so `.pi/prompts/pm-gsd-loop.md` manual fallback recorded.
- Selected approved dependency: `github.com/spf13/cobra v1.10.2`.

## Delivery status

- Plan/TDD/verification artifacts: created.
- Red evidence: pending.
- Implementation: pending.
- Verification: pending.
- PR: pending.
- Review route: human/parent-PR fallback pending; no Claude/Copilot request planned for this blocker window.

## Human gates

No secrets, credentialed checks, dependency deviations, generic write tools, reverse ETL execution, quality-gate reduction, or merge to `main`.
