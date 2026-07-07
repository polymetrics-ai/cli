# Prompt Notes

## Worker Prompt Contract

Use the shared runtime contract:

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md` when parent/subissue work
  is involved
- `.agents/skills/caveman/SKILL.md` only for compact status, worker prompts, and handoffs

For parent issues, the prompt must explicitly say:

```text
Keep orchestration active. Build the ready queue. Spawn or assign every independent ready worker up
to runtime limits. If no worker is spawned while work remains, record one not_spawned_* blocker.
Use caveman compact mode only for status, worker prompts, and handoffs. Preserve exact commands,
code, test output, security warnings, destructive-action warnings, ordered safety gates, and
approval gates.
```
