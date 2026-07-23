# Agentic delivery system

Status: implementation planning artifact. This package is agent-neutral and can be consumed by
Codex, Claude, OpenCode, GitHub Actions, local scripts, or future orchestration runtimes.

## Purpose

Make a GitHub issue sufficient to launch a safe implementation PR without relying on chat history.
The issue provides task-specific scope; this package provides the reusable execution contract,
skills, guardrails, YAML agent definitions, and handoff rules.

## Files

- `contracts/issue-agent-contract.md`: generic contract every implementation agent must follow.
- `contracts/issue-prompt-template.md`: issue section template that points at the generic contract.
- `contracts/pm-code-review-disposition-template.md`: canonical PM exact-head local review
  disposition and gate-state format.
- `contracts/code-review-disposition-template.md`: legacy hosted-review reply format retained for
  historical and explicitly authorized non-PM use.
- `contracts/parent-issue-roadmap-template.md`: parent issue format for epic-sized work with
  sub-issues and stacked PRs.
- `contracts/parent-orchestrator-contract.md`: runtime contract for parent issue orchestration.
- `contracts/pm-worker-handoff-template.md`: canonical PM worker-to-orchestrator handoff with
  exact-head local review, Shepherd, correction-budget, and human-gate fields.
- `contracts/worker-handoff-template.md`: legacy generic handoff retained for historical records.
- `matrices/task-skill-matrix.yaml`: required skills and capabilities by task type.
- `workflows/local-codex-review-loop.md`: canonical PM fresh-context exact-head code review and
  disposition loop.
- `workflows/shepherd-validator.md`: independent orchestration trajectory validation after code
  review and before integration.
- `workflows/claude-review-loop.md` and `workflows/automated-review-routing-loop.md`: legacy
  GitHub-bot routes retained for truthful historical records and explicitly authorized non-PM work;
  they are not current PM coverage.
- `workflows/parent-issue-orchestration-loop.md`: full parent issue execution loop across workers,
  sub-PRs, parent PR review, and human readiness.
- `workflows/gsd-universal-runtime-loop.md`: cross-runtime GSD loop contract for Claude, Codex,
  OpenCode, and future runtimes.
- `workflows/codex-active-orchestration-loop.md`: Codex-specific active orchestration loop for
  parent issues, because Codex subagents must be spawned explicitly.
- `workflows/stacked-parent-subissue-workflow.md`: parent branch and sub-PR workflow for large
  issue hierarchies.
- `references/issue-roadmap-best-practices.md`: source-backed GitHub and Atlassian planning
  guidance.
- `references/claude-review-best-practices.md`: source-backed Claude review practices.
- `references/automated-review-routing-best-practices.md`: source-backed Claude-to-Copilot
  fallback policy.
- `references/caveman-token-compression.md`: compact-output guidance for long-running
  orchestration.
- `references/yaml-agent-best-practices.md`: research-backed rules for YAML agent specs.
- `references/gsd-pi-adapter.md`: repo-local official GSD Core command path for Pi and shell agents.
- `references/required-skills-routing.md`: required Go/design skill routing for agents and subagents.
- `references/runtime-rlm-website-integration.md`: required runtime/RLM/Pi-agent/website integration knowledge for Podman, PostgreSQL, DragonflyDB/Redis-compatible coordination, Temporal, RLM agent mode, and website docs.
- `references/cli-help-docs-website-parity.md`: required parity checklist for CLI help, manual docs, generated docs, and website docs.
- `schemas/agent-spec.schema.yaml`: lightweight schema contract for repo-local YAML agents.
- `schemas/orchestration-state.schema.yaml`: field contract for parent issue state ledgers.
- `agents/<type>/*.agent.yaml`: reusable role definitions grouped by agent type.

The `.agents/agentic-delivery/` directory holds shared contracts, conventions, and role specs.
Specialized agent families can live beside it under `.agents/<functional-area>/` while reusing the
same schema and issue-to-PR contract.

Runtime-specific files, such as `.codex/agents/*.toml` and `.opencode/agents/*.md`, are thin
activation adapters. They must point back to the `.agents/` YAML and Markdown contracts instead of
copying GSD/TDD, local review, Shepherd, or human-gate policy.

## Design principles

- Agent definitions are declarative YAML, but runtime-specific adapters stay optional.
- Issues remain the unit of work. PRs must reference issues.
- Large goals use parent issues with sub-issues. Sub-PRs may merge into a parent branch without
  human approval only when all automated gates pass and no human gate is triggered.
- A parent issue orchestrator owns shared parent artifacts, parent PR state, sub-PR merge
  arbitration, automated review coverage routing, and final readiness. Worker agents implement one
  assigned sub-issue and report back through the worker handoff template.
- Parent issue orchestration is active, not advisory. If ready sub-issues exist and runtime
  subagent tools are available, the orchestrator must spawn or assign all independent ready workers
  up to runtime limits, or record the blocker category and next unblock action.
- Stacked work must have a parent PR from the parent branch to `main` before sub-issues are treated
  as executable. If the parent branch has no useful file diff, use a deliberate seed commit to open
  the parent PR thread.
- Skills are declared by capability, with preferred local skill names when available.
- Guardrails are explicit hard stops, not prose suggestions.
- Run repo-local GSD registry discovery before production behavior changes. For parent/stacked
  work, `/pm-orchestrate` is the active owner. When `programming-loop` is absent, the PM owner runs
  PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE with durable evidence; agents must
  not invent the missing command or weaken test-first delivery.
- Implementation agents must plan before production edits, keep GSD/TDD/verification artifacts
  current, record the GSD command path used, record required Go/design skills loaded from
  `references/required-skills-routing.md`, and commit/push coherent green slices to the active
  issue/PR branch after local green gates.
- CLI feature work must keep runtime help, bare namespace behavior, `docs/cli/**`, website docs,
  generated help/manual artifacts, and tests in parity; follow
  `references/cli-help-docs-website-parity.md`.
- Runtime/RLM/Pi-agent work must preserve the dependency-free default, treat Podman/PostgreSQL/DragonflyDB/Temporal as optional runtime-backed services unless explicitly in scope, and follow `references/runtime-rlm-website-integration.md`.
- Current and forward PM code review uses a fresh-context read-only local Codex reviewer bound to
  exact base/head identities. Every actionable finding receives a reasoned disposition; changed
  heads require affected verification and fresh-context re-review.
- Independent Shepherd trajectory validation follows clean code review and precedes integration.
  Reviewer and Shepherd evidence are separate; neither can self-certify the other.
- Claude and GitHub Copilot are not required, requested, or fallback PM review coverage. Their
  legacy workflow files remain only for historical records or separately authorized non-PM routes.
- Secrets, auth scope changes, destructive actions, dependencies, and quality-gate reductions are
  human-gated.
- Parent PRs into `main` are always human-gated.
