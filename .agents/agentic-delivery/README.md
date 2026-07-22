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
- `contracts/code-review-disposition-template.md`: required reply format for automated review
  findings.
- `contracts/parent-issue-roadmap-template.md`: parent issue format for epic-sized work with
  sub-issues and stacked PRs.
- `contracts/parent-orchestrator-contract.md`: runtime contract for parent issue orchestration.
- `contracts/worker-handoff-template.md`: required worker-to-orchestrator handoff format.
- `matrices/task-skill-matrix.yaml`: required skills and capabilities by task type.
- `workflows/claude-review-loop.md`: post-implementation Claude review and disposition
  loop.
- `workflows/automated-review-routing-loop.md`: routing policy for Claude primary review,
  Copilot backup review, and human fallback.
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
copying GSD/TDD, Claude, or human-gate policy.

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
- Production behavior changes require `gsd-programming-loop` through the repo-local Pi adapter:
  use `/gsd-programming-loop ...` in Pi or `scripts/gsd prompt programming-loop ...` from shell. If
  the adapter is unavailable, agents must record a manual-GSD fallback and still provide test-first
  evidence.
- Implementation agents must plan before production edits, keep GSD/TDD/verification artifacts
  current, record the GSD command path used, record required Go/design skills loaded from
  `references/required-skills-routing.md`, and commit/push coherent green slices to the active
  issue/PR branch after local green gates.
- CLI feature work must keep runtime help, bare namespace behavior, `docs/cli/**`, website docs,
  generated help/manual artifacts, and tests in parity; follow
  `references/cli-help-docs-website-parity.md`.
- Runtime/RLM/Pi-agent work must preserve the dependency-free default, treat Podman/PostgreSQL/DragonflyDB/Temporal as optional runtime-backed services unless explicitly in scope, and follow `references/runtime-rlm-website-integration.md`.
- Claude review is a post-implementation gate. Every actionable review item must receive a
  reasoned disposition before it is resolved. Non-draft PRs targeting `main` from trusted authors
  are reviewed automatically on open, reopen, or ready-for-review. Follow-up fix commits need a
  single `@claude review` to re-review the new unreviewed commits; do not comment `@claude review`
  after every push. A manual `@claude review` is also required when the automatic review did not
  run, such as for an untrusted or first-time author.
- Claude automatic review is the primary automated review route. GitHub Copilot review is
  fallback-only when a Claude run errors, its quota is exhausted, the automatic review did not run,
  or Claude is otherwise unavailable and review coverage is blocking progress.
- A skipped Claude review is not approval. For sub-PRs whose base is not `main`, the
  orchestrator must record sub-PR review coverage or route the integrated commit range through the
  parent PR review fallback.
- GitHub Copilot review is a backup route when a Claude run errors, its quota is exhausted, the
  review did not run, or Claude is otherwise unavailable. Copilot comments must be dispositioned
  like Claude comments, but Copilot review is not approval and must not bypass human gates.
- Secrets, auth scope changes, destructive actions, dependencies, and quality-gate reductions are
  human-gated.
- Parent PRs into `main` are always human-gated.
