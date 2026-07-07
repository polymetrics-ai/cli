# Polymetrics GSD Programming Loop

Task:

{{task}}

Run the mandatory GSD universal programming loop for this implementation or behavior-changing
task.

Required reading:

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

Plan before production edits. Capture red test or validation evidence before behavior changes.
Commit and push coherent green slices to the active issue/PR branch after local gates. Use
subagents for independent work only when each mutating worker has an isolated worktree or working
directory.

Do not skip TDD, verification, review disposition, or human gates.
