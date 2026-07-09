---
description: Run GSD universal programming loop
agent: gsd-loop-orchestrator
---

Run the GSD universal programming loop for this request.

Arguments: `$ARGUMENTS`

Use:

- `AGENTS.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`
- `.agents/skills/caveman/SKILL.md`

Use active orchestration when the work has independent subtasks.

After running any GSD helper/preflight, immediately build the ready queue and invoke worker subtasks
through `.opencode/commands/gsd-worker.md` when scopes are independent. If no worker is invoked
while work remains, record one exact `not_spawned_*` blocker plus next unblock action.
