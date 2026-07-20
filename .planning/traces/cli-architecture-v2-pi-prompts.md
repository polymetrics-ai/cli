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
and record review coverage. Integrate terminal-design gate #462/D-TUI and its accepted correction
PR/review disposition before dispatching production work for #408, #409, #411, #412, #414, #416,
#418, or chart child #463. Never merge the parent PR into main without human approval.
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

## TUI worker session

Use this prompt in each isolated Pi session for issues #408, #409, #411, #412, #414, #416,
#418, and #463 after #462/D-TUI plus its accepted correction PR/review blocker are integrated or
cleared by the parent orchestrator. Replace the placeholders with the assigned issue/worktree.

```text
Execute polymetrics-ai/cli#<ISSUE> as one bounded Bubble Tea TUI worker for parent #397.

Work only in <ABSOLUTE_WORKTREE>, on branch <BRANCH>, based on the latest
feat/cli-architecture-v2. Confirm design gate #462/D-TUI, its correction/review status, and all
GitHub blocked-by dependencies are integrated before coding. The parent orchestrator owns GitHub
blocked-by metadata updates; workers do not mutate issue metadata. Read AGENTS.md, the full issue,
docs/design/tui-ux-design.md,
docs/design/terminal-ui-research-and-design-system.md, ADR-0003, the CLI help/docs/website parity
reference, required-skills-routing.md, and the worker handoff template.

Run:
/gsd doctor
/gsd plan-phase <ISSUE> --skip-research
/gsd-programming-loop init --phase <ISSUE> --dry-run

Load and record `bubble-tea-tui-design`, then `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`,
`golang-concurrency`, and `golang-documentation` as applicable. Before production edits, put exact
RED cases in PLAN.md/TDD-LEDGER.md for bare namespace help-not-TUI behavior, TUI/Huh activation
requiring both stdin and stdout TTYs, `stdin-piped+stdout-TTY` fallback, `stdout-piped`, `CI`,
`--json`, `--plain`, `--no-input`, Normal/Filter/Edit printable-key conflicts, arrows+Vim
equivalence, focus/context help, wide/standard/compact/guard rendering, loading/empty/failure/
cancel/final states, no-color/ASCII/reduced-motion/accessibility fallback, sanitation/redaction,
approval-token non-display, cancellation/race cleanup, and unchanged plain/JSON/stdout/stderr/exit
semantics.

Follow the operator-workspace design: LazyGit panel hierarchy, fzf filter/list/preview behavior,
bpytop exact metric density, Gum focused wizard cadence, and Polymetrics' quiet pipeline-rail
language. Bubble Tea/Huh prompts activate only when stdin and stdout are TTYs; piped/non-TTY stdin
falls back to deterministic plain/noninteractive behavior and must never be consumed unexpectedly,
hang, or bypass through `/dev/tty`. Bare `pm query` and bare `pm reverse` render contextual
help/subcommand summaries and exit 0; explicit interactive subcommands are `pm query grid` and
`pm reverse guide`. Do not copy generic shell execution, shell-backed previews, unlabelled
destructive keys, generic HTTP/SQL writes, generic file writes, approval-token display, or
interactive secret entry. Mouse/OSC52/advanced graphics are optional accelerators only.

For #411 query grid and #463 charts, operate only on returned read-only rows. Query export must be a
typed read-only export with project-scoped default, clean/confined path, control-character/
traversal/broad-path/symlink race rejection, no overwrite by default, confirmation only when stdin
and stdout are TTYs, noninteractive `--output` + `--force`, sanitized command echo, exact
`--no-input` guidance, and no generic file-write or SQL-write boundary. For query chart issue #463,
keep table/text access, axes,
units,
exact selected values, deterministic bounds/downsampling disclosure, and accessibility fallback.
`ntcharts/v2` is not approved: stop at the human dependency gate unless #463 records
explicit approval and an exact pinned wrapper. Do not edit go.mod speculatively.

Implement strict RED → GREEN → refactor, commit/push coherent green slices, and run targeted,
repeated/race, help/manual/website parity, and full repository gates required by the issue. Then run:
/gsd verify-work
/gsd-code-review <ISSUE>

Open a stacked PR to feat/cli-architecture-v2 with Refs #<ISSUE> and Refs #397, route automated
review, and return the complete worker handoff. Never merge the parent PR to main.
```

### Shell launcher for a new Pi TUI session

Run from the issue's isolated worktree, not from the parent checkout:

```bash
pi --tools read,bash,edit,write,grep,find,ls,subagent --approve
```

Paste the completed **TUI worker session** block at the Pi prompt. If the project-local aliases are
not visible, trust/reload the repository and use `scripts/gsd prompt plan-phase <ISSUE>
--skip-research` to generate the same planning prompt; record any manual GSD fallback.

## Initial scheduling

1. Run only #398 until the parent branch and draft parent PR exist.
2. Run #399 → #400 → #401 → #402 serially.
3. After #402 is integrated and reviewed, launch #403, #404, #406, and #410 in four isolated Pi
   worktrees/sessions.
4. Integrate #462 before the first production TUI session.
5. Continue from GitHub blocked-by state; do not infer readiness from issue numbers alone.
