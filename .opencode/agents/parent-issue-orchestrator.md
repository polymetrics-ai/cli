---
description: Coordinate parent issues, subissue workers, stacked PRs, and automated review coverage.
mode: primary
permission:
  edit: allow
  bash: allow
  skill: allow
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

Never merge parent PR to `main` without human approval.
