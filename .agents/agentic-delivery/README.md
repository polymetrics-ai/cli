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
- `matrices/task-skill-matrix.yaml`: required skills and capabilities by task type.
- `references/yaml-agent-best-practices.md`: research-backed rules for YAML agent specs.
- `schemas/agent-spec.schema.yaml`: lightweight schema contract for repo-local YAML agents.
- `agents/<type>/*.agent.yaml`: reusable role definitions grouped by agent type.

The `.agents/agentic-delivery/` directory holds shared contracts, conventions, and role specs.
Specialized agent families can live beside it under `.agents/<functional-area>/` while reusing the
same schema and issue-to-PR contract.

## Design principles

- Agent definitions are declarative YAML, but runtime-specific adapters stay optional.
- Issues remain the unit of work. PRs must reference issues.
- Skills are declared by capability, with preferred local skill names when available.
- Guardrails are explicit hard stops, not prose suggestions.
- Production behavior changes require test-first evidence.
- Secrets, auth scope changes, destructive actions, dependencies, and quality-gate reductions are
  human-gated.
