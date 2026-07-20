# TDD Ledger — Issue 398 CLI Architecture v2 Bootstrap

## Classification

Planning-only parent-orchestrator bootstrap. No production behavior changes and no Go source edits are allowed.

## Red / Initial Evidence

| Evidence | Command / Source | Result |
|---|---|---|
| Parent branch missing | `git ls-remote --heads origin feat/cli-architecture-v2` | no output |
| Parent PR missing | `gh pr list --head feat/cli-architecture-v2 --base main --state all --json ...` | `[]` |
| Parent issue says parent PR absent | `gh issue view 397 --json body` | issue body records Stage 0 PR blocker |
| Active roadmap lacks CLI Architecture v2 milestone | `.planning/ROADMAP.md` before edit | connector-parity roadmap only |
| No production source edits at baseline | `git status --short --branch`; `git diff --name-only -- cmd internal` | no tracked `cmd/**` or `internal/**` edits |
| GSD programming-loop shell command missing | `scripts/gsd prompt programming-loop init --phase 398 --dry-run` | `scripts/gsd: unknown GSD command: programming-loop` |

## Green Evidence Targets

| Target | Verification |
|---|---|
| GSD adapter healthy | `scripts/gsd doctor` exits 0 |
| GSD planning prompts generated | `scripts/gsd prompt new-milestone "CLI Architecture v2"` and `scripts/gsd prompt plan-phase 398 --skip-research` produce non-empty prompts |
| Active planning names CLI Architecture v2 | `rg -n "CLI Architecture v2|feat/cli-architecture-v2|#397|#398" .planning/PROJECT.md .planning/ROADMAP.md .planning/phases/398-cli-architecture-v2-bootstrap .planning/traces` |
| 22-phase roadmap recorded | `rg -n "Golden transcript|Cobra router|typed Viper|OpenTelemetry metrics|Architecture v2 cleanup" .planning/ROADMAP.md` |
| Connector-parity workstreams preserved | `rg -n "Polymetrics CLI Connector Parity|Inventory and Surface Reconciliation|Durable Read and ETL Parity" .planning/PROJECT.md .planning/ROADMAP.md` |
| No Go source changed | `git diff --name-only -- cmd internal` returns no output |
| Diff clean | `git diff --check` exits 0 |
| Parent PR exists | `gh pr view <number> --json number,url,isDraft,baseRefName,headRefName` shows draft PR `feat/cli-architecture-v2` → `main` |

## Red/Green/Refactor Notes

No failing Go test is appropriate for this planning-only issue. The red evidence is missing parent branch/PR and absent active GSD milestone. Green evidence is planning state, PR existence, and scope guards.

## Skills Recorded

- `gsd-core`
- `caveman`
- issue-first/parent-orchestrator contracts
- Required Go skills: not applicable (no Go implementation)
