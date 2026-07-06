# Parent Issue Orchestrator Plan

Issue: #50
Branch: `feat/50-parent-issue-orchestrator`
Mode: manual GSD fallback; `scripts/programming-loop.mjs` is not present in this clone.

## Objective

Add a generic parent issue orchestrator to `.agents/agentic-delivery` so large issue hierarchies can
run through a single parent issue, parent PR, parallel sub-issue workers, controlled sub-PR
integration, automated review coverage, and final human approval.

## Scope

- Add a parent orchestrator contract, workflow, state schema, YAML agent spec, and worker handoff
  template.
- Update existing `.agents/agentic-delivery` contracts so workers implement and report while the
  orchestrator owns shared parent artifacts and default merge decisions.
- Keep `.agents/` as the generic source of truth and describe runtime-specific agents as thin
  adapters.
- Add only short cross-agent pointers in `AGENTS.md` and `CLAUDE.md`.
- Add CodeRabbit-to-Copilot fallback routing so CodeRabbit rate limits do not cause repeated manual
  review commands or unreviewed PRs.

## Non-Goals

- No connector runtime implementation.
- No GitHub Project writes or auth scope refresh.
- No dependency changes.
- No changes to PR guard Go code in this slice unless docs validation exposes a defect.
- No parent PR merge to `main`.

## Implementation Steps

1. Add orchestration artifacts:
   - `contracts/parent-orchestrator-contract.md`
   - `contracts/worker-handoff-template.md`
   - `workflows/parent-issue-orchestration-loop.md`
   - `schemas/orchestration-state.schema.yaml`
   - `agents/coordination/parent-issue-orchestrator.agent.yaml`
2. Update existing shared docs:
   - `README.md`
   - `matrices/task-skill-matrix.yaml`
   - `schemas/agent-spec.schema.yaml`
   - `workflows/stacked-parent-subissue-workflow.md`
   - `workflows/coderabbit-review-loop.md`
   - `workflows/automated-review-routing-loop.md`
   - `contracts/issue-agent-contract.md`
   - `contracts/issue-prompt-template.md`
   - `contracts/parent-issue-roadmap-template.md`
3. Add source-backed automated review routing references for CodeRabbit primary review, Copilot
   backup review, and human fallback.
4. Add short source-of-truth pointers in `AGENTS.md` and `CLAUDE.md`.
5. Validate YAML, whitespace, JSON, and focused PR guard tests.
6. Commit, push, open PR with `Closes #50`, then wait for automatic CodeRabbit review because the
   PR targets `main` and is not draft. Manual CodeRabbit review commands are fallback-only; if
   CodeRabbit is rate-limited and review coverage blocks progress, request GitHub Copilot review
   once as backup when enabled.

## Human Gates

- Auth scope changes.
- GitHub Project creation or mutation.
- Dependencies.
- Destructive external actions.
- Quality gate reductions.
- Parent PR merge to `main`.
