---
description: PM-owned issue-first TDD lifecycle for Polymetrics
argument-hint: "<implementation-task>"
---

# Polymetrics PM Implementation Loop

Task:

$@

Run the universal implementation lifecycle under the active PM parent orchestrator.

Required reading:

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/local-codex-review-loop.md`
- `.agents/agentic-delivery/workflows/shepherd-validator.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

Run `scripts/gsd doctor`, `scripts/gsd list`, and source discovery. Use only commands in the
registry. When `programming-loop` is absent, do not invoke or invent it: `/pm-orchestrate` owns PLAN
→ RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE and records the absent command.

Plan before production edits. Capture red test or validation evidence before behavior changes.
Commit and push coherent green slices to the active issue/PR branch after local gates. Use
subagents for independent work only when each mutating worker has an isolated worktree or working
directory.

After exact-head verification, dispatch a fresh-context read-only local Codex reviewer through
`local-codex-review-loop.md`, disposition every finding, and re-review every changed head. Then run
independent `shepherd-validator.md` trajectory validation before integration. Do not request or
count Claude or GitHub Copilot as required, fallback, or substitute PM review coverage.

Do not skip TDD, verification, review disposition, Shepherd, or human gates.
