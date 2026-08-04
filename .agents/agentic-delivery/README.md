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
activation adapters. The `pm-delivery-worker` and `pm-connector-worker` adapters project marked
blocks from `canonical/delivery-contract.json`; they must not copy GSD/TDD, review, or human-gate
policy by hand.

## Design principles

- Agent definitions are declarative YAML, but runtime-specific adapters stay optional.
- Issues remain the unit of work. PRs must reference issues.
- Large goals use parent issues with sub-issues. Sub-PRs may merge into a parent branch without
  human approval only when all automated gates pass and no human gate is triggered.
- One canonical worker owns the job and its shared parent artifacts, parent PR state, sub-PR
  integration decisions, automated review coverage, and final readiness inline. It processes ready
  sub-issues without spawning orchestrator, shepherd, planner, reviewer, verifier, or GSD roles;
  durable GitHub and GSD artifacts let a later worker resume if needed.
- Stacked work must have a parent PR from the parent branch to `main` before sub-issues are treated
  as executable. If the parent branch has no useful file diff, use a deliberate seed commit to open
  the parent PR thread.
- Skills are declared by capability, with preferred local skill names when available.
- Guardrails are explicit hard stops, not prose suggestions.
- Production behavior changes use the installed repo-local GSD sequence: `discuss-phase`,
  `plan-phase --tdd`, `execute-phase`, `verify-work` with gap planning/execution when needed, then
  `code-review`. The pinned adapter has no `programming-loop` command. Generated prompts may be
  executed inline when compatible isolated runtime agents are unavailable or the canonical
  single-worker contract forbids spawning roles; record the fallback and retain test-first evidence.
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
- A skipped Claude review is not approval. For sub-PRs whose base is not `main`, the canonical
  worker must record sub-PR review coverage or route the integrated commit range through the parent
  PR review fallback.
- GitHub Copilot review is a backup route when a Claude run errors, its quota is exhausted, the
  review did not run, or Claude is otherwise unavailable. Copilot comments must be dispositioned
  like Claude comments, but Copilot review is not approval and must not bypass human gates.
- Secrets, auth scope changes, destructive actions, dependencies, and quality-gate reductions are
  human-gated.
- Parent PRs into `main` are always human-gated.
