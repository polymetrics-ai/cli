# Agentic delivery system

Status: active project-local delivery policy. This package is agent-neutral and can be consumed by
Codex, Claude, OpenCode, GitHub Actions, local scripts, or future orchestration runtimes.

## Purpose

Make a GitHub issue sufficient to launch a safe implementation PR without relying on chat history.
The issue provides task-specific scope; this package provides the canonical delivery contract,
skills, guardrails, compatibility guidance, and durable checkpoint templates.

## Files

- `contracts/issue-agent-contract.md`: generic contract every implementation agent must follow.
- `contracts/issue-prompt-template.md`: issue section template that points at the generic contract.
- `contracts/code-review-disposition-template.md`: required reply format for automated review
  findings.
- `canonical/delivery-contract.json`: authoritative issue-first worker, connector overlay, delivery
  state machine, authority boundary, and projection registry.
- `contracts/parent-issue-roadmap-template.md`: parent issue format for epic-sized work with
  sub-issues and stacked PRs.
- `contracts/parent-orchestrator-contract.md`: compatibility path for single-worker parent job
  ownership; it does not activate a parent-orchestrator role.
- `contracts/worker-handoff-template.md`: durable wave checkpoint format for the active worker or a
  successor.
- `matrices/task-skill-matrix.yaml`: required skills and capabilities by task type.
- `workflows/claude-review-loop.md`: post-implementation Claude review and disposition
  loop.
- `workflows/automated-review-routing-loop.md`: routing policy for Claude primary review,
  Copilot backup review, and human fallback.
- `workflows/parent-issue-orchestration-loop.md`: legacy multi-worker procedure retained for its
  owning cleanup wave; it is not an active instruction for the canonical flow.
- `workflows/gsd-universal-runtime-loop.md`: background runtime procedure only where it agrees with
  the canonical contract.
- `workflows/codex-active-orchestration-loop.md`: legacy multi-worker Codex procedure retained for
  its owning cleanup wave; it is not an active instruction for the canonical flow.
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
- `schemas/orchestration-state.schema.yaml`: legacy multi-worker ledger schema retained for its
  owning cleanup wave; canonical jobs use `state_machine.steps` from the JSON source.
- `agents/<type>/*.agent.yaml`: legacy role definitions retained for their owning cleanup wave; the
  canonical issue-first flow does not activate them.

The `.agents/agentic-delivery/` directory holds shared contracts, conventions, and role specs.
Specialized agent families can live beside it under `.agents/<functional-area>/` while reusing the
same schema and issue-to-PR contract.

Runtime-specific files, such as `.claude/agents/*.md`, `.codex/agents/*.toml`, `.pi/agents/*.md`,
and `.opencode/agents/*.md`, are generated activation adapters. `canonical/delivery-contract.json`
is the sole owner of each registered target's path, render mode, and requiredness. The checked-in
Claude, Codex, Pi, and OpenCode adapters are complete required full-file projections. Regenerate
them from the canonical source; never hand-edit or copy GSD/TDD, review, human-gate, or connector
certification policy into an adapter.

The same canonical contract also owns the versioned, read-only connector-certification Shepherd
gate. It consumes only generated certification artifacts and emits `PROCEED`, `RETRY`, or `HALT`
with stable cell/evidence coordinates before protected connector/parent transitions. Structural
`agentcontractgen check` validates the contract and registered projections; certification is
evaluated explicitly at a protected transition, so a baseline with zero certified connectors does
not make a general contract check red.

At each protected transition, run the canonical command rendered in the registered projection with
one canonical absolute non-symlink repository root:
`go run ./cmd/agentcontractgen certification-gate --root <repository-root> --connector <connector>
--transition <transition>`. It emits the complete verdict as JSON and exits zero only for
`PROCEED`; `RETRY` and `HALT` preserve their evidence on stdout while blocking the transition.

Codex loads the project-local `.codex` configuration layer only when the repository is trusted.
The generated Codex workers therefore fail closed for project-local selection in an untrusted
repository: trust it in Codex before selecting either worker. Their generated
`agents.enabled = false` setting disables further Codex multi-agent delegation. The canonical
`codex` section owns this isolation contract, its official documentation links, and the caveat that
same-filename user/project standalone-agent precedence is undocumented.

The Claude check inventories every repository-local `.claude/agents` scope recursively while
pruning only the exact root `.git` metadata directory. It rejects scope symlinks, extra Markdown
definitions, duplicate names, and missing or non-regular registered definitions; exactly the two
registered Markdown files may define Claude agents. Whole-file comparison normalizes each
carriage-return run immediately before LF to one LF, so CRLF checkout conversion does not create
drift; every other byte remains exact. Claude workers preload only plugin-qualified trusted skills
and omit and deny runtime `Skill`; any repository-routed skill without a trusted namespace is
documented as unavailable rather than exposed through a collision-prone name.

## Design principles

- `canonical/delivery-contract.json` is the sole owner of the canonical roles, ordered delivery
  states, tracker gates, GSD/no-mistakes topology, authority boundary, and projection registry.
- Runtime-specific adapters are generated from the canonical JSON contract as their harness waves
  land.
- Issues remain the unit of work. PRs must reference issues.
- Large goals use parent issues with sub-issues. A sub-PR may advance to completed parent-branch
  integration without human approval only when all automated gates pass and no human gate is
  triggered.
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
