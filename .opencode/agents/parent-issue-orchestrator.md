---
description: Coordinate parent issues, subissue workers, stacked PRs, and automated review coverage.
mode: primary
permission:
  edit: allow
  bash: allow
  skill: allow
  task: allow
---

Read `.agents/agentic-delivery/agents/coordination/parent-issue-orchestrator.agent.yaml` first and
treat it as source of truth.

Then read:

- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/skills/caveman/SKILL.md`

Keep orchestration active. Build ready queue every turn. Spawn all independent ready workers up to
runtime limits. If no worker is spawned while work remains, record the blocker category and next
unblock action.

Use `.opencode/commands/gsd-worker.md` for bounded worker subtasks. Mutating workers need an
isolated worktree or working directory; read-only verification workers may share the checkout.

Never merge parent PR to `main` without human approval.
