# Phase Agents and Ownership

Phase: `471-pi-agent-session-shepherd`

The parent orchestrator owns #471, PR #472, shared phase artifacts, dependency scheduling, child
handoffs, integration decisions, review coverage, human-decision routing, and final readiness.
Mutating workers own one child issue in one isolated worktree and must not edit shared parent
artifacts unless explicitly assigned.

| Issue/role | Model | Thinking | Write scope / authority |
|---|---|---|---|
| Parent orchestrator | `openai-codex/gpt-5.6-sol` | `xhigh` | shared roadmap/state/GitHub coordination; typed mutations |
| #473 foundation worker | `openai-codex/gpt-5.6-sol` | `high` | existing core/state/evidence/SDK/controller files |
| #474 policy worker | `openai-codex/gpt-5.6-sol` | `high` | dependency/policy/reconciler files |
| #475 runtime worker | `openai-codex/gpt-5.6-sol` | `high` | AgentSession runtime/tool-policy files |
| #476 Git worker | `openai-codex/gpt-5.6-sol` | `high` | workspace/Git adapter files |
| #477 decision worker | `openai-codex/gpt-5.6-sol` | `high` | human/GitHub decision broker files |
| #478-#481 implementation/correction | `openai-codex/gpt-5.6-sol` | `high` | exact issue scope in its isolated worktree |
| research/planning/verification/review/disposition | `openai-codex/gpt-5.6-sol` | `xhigh` | bounded role scope; read-only except planning artifacts |

The controller is the only session spawner. Child sessions never receive recursive spawning,
unrestricted external-write tools, secrets, or authority beyond their issue/worktree/scope. Every
turn records `spawned` or one exact `not_spawned_*` reason.
