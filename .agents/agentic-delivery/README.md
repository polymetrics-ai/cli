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
- `matrices/task-skill-matrix.yaml`: required skills and capabilities by task type.
- `workflows/coderabbit-review-loop.md`: post-implementation CodeRabbit review and disposition
  loop.
- `workflows/stacked-parent-subissue-workflow.md`: parent branch and sub-PR workflow for large
  issue hierarchies.
- `references/issue-roadmap-best-practices.md`: source-backed GitHub and Atlassian planning
  guidance.
- `references/coderabbit-review-best-practices.md`: source-backed CodeRabbit review practices.
- `references/yaml-agent-best-practices.md`: research-backed rules for YAML agent specs.
- `learnings/github-cli-surface-metadata.md`: first-slice lessons for gh-inspired connector command
  metadata, validation, docs, and agent guardrails.
- `schemas/agent-spec.schema.yaml`: lightweight schema contract for repo-local YAML agents.
- `agents/<type>/*.agent.yaml`: reusable role definitions grouped by agent type.

The `.agents/agentic-delivery/` directory holds shared contracts, conventions, and role specs.
Specialized agent families can live beside it under `.agents/<functional-area>/` while reusing the
same schema and issue-to-PR contract.

## Design principles

- Agent definitions are declarative YAML, but runtime-specific adapters stay optional.
- Issues remain the unit of work. PRs must reference issues.
- Large goals use parent issues with sub-issues. Sub-PRs may merge into a parent branch without
  human approval only when all automated gates pass and no human gate is triggered.
- Skills are declared by capability, with preferred local skill names when available.
- Guardrails are explicit hard stops, not prose suggestions.
- Production behavior changes require `gsd-programming-loop`; if local GSD scripts are unavailable,
  agents must record a manual-GSD fallback and still provide test-first evidence.
- Implementation agents must plan before production edits, keep GSD/TDD/verification artifacts
  current, and commit/push coherent green slices to the active branch when repo policy permits.
- CodeRabbit review is a post-implementation gate. Every actionable review item must receive a
  reasoned disposition before it is resolved. Follow-up fix commits should rely on automatic
  incremental review when active; manual `@coderabbitai review` is only a fallback for paused,
  disabled, skipped, rate-limited, or auto-paused review states.
- Secrets, auth scope changes, destructive actions, dependencies, and quality-gate reductions are
  human-gated.
- Parent PRs into `main` are always human-gated.
