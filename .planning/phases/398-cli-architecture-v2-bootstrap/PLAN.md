# Issue 398 Plan — CLI Architecture v2 Bootstrap

**Issue:** [#398](https://github.com/polymetrics-ai/cli/issues/398)
**Parent:** [#397](https://github.com/polymetrics-ai/cli/issues/397)
**Mode:** parent-orchestrator bootstrap / planning-only
**GSD command path:** `scripts/gsd doctor`, `scripts/gsd prompt new-milestone "CLI Architecture v2"`, `scripts/gsd prompt plan-phase 398 --skip-research`; `scripts/gsd prompt programming-loop init --phase 398 --dry-run` was unavailable (`unknown GSD command: programming-loop`), so the Pi-local `/pm-gsd-loop` prompt contract is the recorded manual GSD programming-loop fallback for this planning-only slice.

## Objective

Register CLI Architecture v2 in active GSD planning without disturbing connector-parity workstreams, create/confirm parent branch `feat/cli-architecture-v2`, and open a draft parent PR to `main` for the parent orchestrator.

## Scope

- Update `.planning/PROJECT.md` and `.planning/ROADMAP.md` with CLI Architecture v2 milestone context and the 22-phase dependency wave.
- Add issue-local GSD plan, TDD ledger, verification checklist, summary, run-state, and orchestration ledger artifacts.
- Commit the existing CLI Architecture v2 planning source documents and ADRs:
  - `docs/plans/cli-architecture-v2-improvement-plan.md`
  - `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md`
  - `docs/design/tui-ux-design.md`
  - `docs/adr/0002-cobra-viper-cli-framework.md`
  - `docs/adr/0003-interactive-tui-layer.md`
  - `docs/adr/0004-opentelemetry-observability.md`
  - `.planning/traces/cli-architecture-v2-issue-backlog.md`
  - `.planning/traces/cli-architecture-v2-pi-prompts.md`
- Create/push `feat/cli-architecture-v2` from `origin/main`.
- Open/update draft parent PR to `main` with `Refs #397`.
- Record parent issue/PR state and ready queue.

## Non-goals

- No `cmd/**` or `internal/**` production source edits.
- No Go implementation for phases #399–#420.
- No dependency changes.
- No credentialed connector checks.
- No reverse ETL execution.
- No merge to `main`.

## Required skills / references loaded

- `gsd-core`
- `caveman`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- parent-orchestrator, stacked PR, universal runtime, automated review, Claude review, issue-agent, worker handoff contracts
- No Go implementation skills required for this planning-only slice.

## Execution plan

1. Confirm parent issue #397 and Stage 0 #398 contracts.
2. Confirm parent branch/PR missing; create `feat/cli-architecture-v2` from `origin/main`.
3. Create issue #398 planning artifacts before editing production planning docs.
4. Update `.planning/PROJECT.md` and `.planning/ROADMAP.md` while preserving connector-parity milestone.
5. Add durable orchestration state ledger.
6. Verify:
   - `scripts/gsd doctor`
   - generated GSD prompts exist
   - no `cmd/**` or `internal/**` changes
   - `git diff --check`
   - planning/docs grep for CLI Architecture v2, phase count, parent branch, human gates
7. Commit and push planning seed.
8. Open draft parent PR to `main`; update ledger/parent issue/PR body with state.
9. Build ready queue. If only #399 becomes ready after parent PR exists, spawn or record next decision.

## Spawn decision for this cycle

`local_critical_path`: Stage 0 owns shared parent branch/PR and shared planning artifacts, so the parent orchestrator executes this bootstrap inline instead of labeling it spawned.
