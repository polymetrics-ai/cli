# CLI Architecture v2 Pi Session Prompts

## Parent orchestrator session

```text
Act as the active parent issue orchestrator for polymetrics-ai/cli#397.

Read AGENTS.md, issue #397, .agents/agentic-delivery/contracts/parent-orchestrator-contract.md,
.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md, the stacked-PR workflow,
the GSD universal runtime loop, automated-review routing loop, Claude review loop,
.agents/agentic-delivery/references/gsd-pi-adapter.md, required-skills-routing.md, and the worker
handoff template. Run /gsd doctor.

Treat GitHub sub-issues and blocked-by relationships as the authoritative ready queue. First execute
Stage 0 issue #398. It must register the 22-phase milestone in active .planning artifacts, create
feat/cli-architecture-v2 from main, make a deliberate planning/roadmap seed commit, and open the
draft parent PR to main. Do not mark an implementation issue worker_ready until that parent PR exists.

After every integration, recompute dependencies and write-scope collisions. Assign each ready issue
to one isolated git worktree and one Pi session. Every worker gets exactly one issue and must use the
issue's /gsd plan-phase, /gsd-programming-loop, /gsd verify-work, and /gsd-code-review workflow.
Do not launch Phase 9 namespace issues concurrently; #421 through #437 are deliberately serialized.
The first designed parallel fan-out occurs after #402: #403, #404, #406, and #410 may run in
separate worktrees. Keep the parent context open, collect worker handoffs, arbitrate stacked PRs,
and record review coverage. Never merge the parent PR into main without human approval.
```

## Stage 0 session

```text
Execute polymetrics-ai/cli#398 as the parent-orchestrator bootstrap for parent #397.

Use the current repository checkout owned by the parent orchestrator. Read the issue and every
required file it names. Run /gsd doctor and /gsd new-milestone "CLI Architecture v2". Update active
.planning/PROJECT.md and .planning/ROADMAP.md with the source plan's 22 phases while preserving the
existing connector-parity workstreams. Record the dependency approvals, frozen cli.Run/JSON/exit-code
contracts, parallel waves, required skills, TDD expectations, and human gates.

Create feat/cli-architecture-v2 from main, commit the deliberate planning scaffold with
"docs(planning): register CLI Architecture v2 milestone", push it, and open a draft parent PR to
main using Refs #397. Update #397 with the PR URL and state ledger. Do not edit cmd/** or internal/**.
Do not merge the parent PR.
```

## Generic worker session

Replace the placeholders with one GitHub issue and its assigned isolated worktree.

```text
Execute polymetrics-ai/cli#<ISSUE> as one bounded worker for parent #397.

Work only in <ABSOLUTE_WORKTREE>, on branch <BRANCH>, based on the latest
feat/cli-architecture-v2. Confirm the draft parent PR exists before coding. Read AGENTS.md, the full
issue, every required source/ADR/reference named by the issue, and the worker handoff template.
Do not edit shared parent planning/orchestration artifacts unless #397 explicitly delegates them.

Run:
/gsd doctor
/gsd plan-phase <ISSUE> --skip-research
/gsd-programming-loop init --phase <ISSUE> --dry-run

Create/update the issue GSD plan, TDD ledger, and verification checklist before production edits.
Load and record the required skills from the issue, starting with golang-how-to for Go work. For
behavior changes, capture the failing test and exact red output before implementation. Implement only
the issue's allowed write scope, commit/push coherent green slices, and run all targeted and broader
verification specified by the issue, including CLI help/docs/website parity.

Then run:
/gsd verify-work
/gsd-code-review <ISSUE>

Open a stacked PR to feat/cli-architecture-v2 with a Conventional Commit title and body containing
Refs #<ISSUE> and Refs #397. Follow automated-review routing and disposition every actionable
finding. Return the repository worker handoff template with branch, PR, commits, changed files,
red/green/refactor evidence, exact verification, skills used, parity status, review coverage, and
remaining blockers. Do not merge the PR and never merge the parent PR to main.
```

## Initial scheduling

1. Run only #398 until the parent branch and draft parent PR exist.
2. Run #399 → #400 → #401 → #402 serially.
3. After #402 is integrated and reviewed, launch #403, #404, #406, and #410 in four isolated Pi
   worktrees/sessions.
4. Continue from GitHub blocked-by state; do not infer readiness from issue numbers alone.
