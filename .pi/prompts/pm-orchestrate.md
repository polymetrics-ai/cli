# Polymetrics Parent Orchestration

Task or parent issue:

{{task}}

Run active parent issue orchestration for Polymetrics.

Required reading before action:

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`

Operate as the live parent orchestrator in the main Pi session. Build the ready queue, create or
confirm the parent branch and parent PR, and delegate independent ready work through the
`subagent` tool using project agents from `.pi/agents/`.

Use compact caveman-style status for progress and handoffs, but keep commands, tests, code,
security warnings, destructive-action warnings, and human gates exact.

Hard stops:

- Do not request, print, store, summarize, or invent secrets.
- Do not push to `main`.
- Do not merge a parent PR to `main` without human approval.
- Do not resolve CodeRabbit or Copilot comments until every actionable item has a written
  disposition.
- If no worker is spawned while ready work remains, record the exact `not_spawned_*` blocker and
  the next unblock action.
