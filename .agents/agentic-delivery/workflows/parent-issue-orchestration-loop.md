# Parent Issue Orchestration Loop

Use this workflow when a parent issue owns multiple sub-issues and the work must proceed through a
parent PR plus stacked sub-PRs. The loop is runtime-generic: "spawn" means the active coordinator
uses the runtime's native worker mechanism, such as Claude Code `Task`, a Codex subagent job, an
OpenCode subtask, or an equivalent future worker context.

## Active Mode

Execute this state machine until the parent issue is human-ready, blocked, or explicitly limited by
the user. Do not stop after describing next steps when required inputs and permissions are available.

At each coordinator turn:

- build the ready queue
- spawn or assign all independent ready workers up to runtime limits
- record `spawned` or a `not_spawned_*` blocker category
- keep the parent orchestrator context open until handoff, human gate, or blocker

## 1. Initialize Parent Work

1. Read `AGENTS.md` and the parent issue.
2. Read the parent orchestrator contract and stacked PR workflow.
3. Read `.agents/agentic-delivery/workflows/local-codex-review-loop.md` and
   `.agents/agentic-delivery/workflows/shepherd-validator.md`.
4. Confirm the parent issue has acceptance criteria, sub-issues, branch policy, verification,
   `max_correction_rounds` (default 4), per-range correction counters, and human gates.
5. Create or confirm the parent branch from `main`.
6. Create or confirm the parent PR from the parent branch to `main` before starting sub-issue work.
7. If GitHub cannot open the parent PR because the branch has no diff:
   - prefer a small roadmap/status scaffold when useful
   - otherwise create an empty seed commit to open the PR thread
8. Keep the parent PR draft until all required sub-issues are integrated and final verification is
   complete.

## 2. Build The Work Queue

For each sub-issue, record:

- issue number and URL
- dependency list
- expected write scope
- branch name
- PR base
- primary worker agent
- required skills
- verification commands
- human gates

Only mark a sub-issue `worker_ready` when its inputs are complete and the parent PR exists.

## 3. Spawn Or Assign Workers

Workers may run in parallel when their write scopes are disjoint and their dependencies are
satisfied. Each worker prompt must include:

- one sub-issue URL
- parent issue URL
- parent branch and parent PR
- allowed write scope
- isolated worker directory or git worktree for mutating tasks
- required GSD/TDD workflow
- required verification
- worker handoff template

The orchestrator continues non-overlapping work while workers run. It must not duplicate worker
implementation tasks. Worker prompts may use compact status language, but exact commands, code,
test output, safety gates, security warnings, destructive-action warnings, and approval gates must
remain uncompressed and unambiguous.

If subagent tooling exists and ready work is independent, parallel worker dispatch is the default.
Sequential execution needs an explicit reason in the state ledger.

Mutating workers must not share the coordinator checkout. If no isolated worktree or working
directory is available, record `not_spawned_isolation_missing` and run the slice locally or with
read-only agents.

## 4. Review Worker Handoffs

When a worker returns:

1. Read the worker handoff.
2. Confirm changed files match the assigned scope.
3. Confirm verification evidence is present.
4. Confirm the sub-PR targets the parent branch.
5. Confirm the sub-PR body uses `Refs` for both the sub-issue and parent issue.
6. Confirm exact-head local Codex findings/dispositions and independent Shepherd evidence, or a
   recorded blocker.
7. Mark missing evidence as `blocked`, not complete.

## 5. Merge Or Block Sub-PRs

Merge a sub-PR into the parent branch only when all gates pass and no human gate is triggered.
Before integration:

1. Verify local and remote candidate identities match the recorded exact base/head SHAs.
2. Run exact-head verification.
3. Compile the v4 review range and require `ready`:
   `scripts/pm-review-system.py compile --scope <validated-scope> --base <exact-base-sha> --head <exact-head-sha> > <manifest>`.
   Stop on identity drift, an authority/impact disagreement, an incomplete bound, or any blocked
   manifest.
4. For every manifest packet, run
   `scripts/pm-review-system.py render --manifest <manifest> --packet-id <packet-id>` and pass its
   stdout unchanged to one fresh-context, read-only local Codex reviewer through
   `.agents/agentic-delivery/workflows/local-codex-review-loop.md`. Preserve one v4 response per
   packet in the response directory; do not hand-build or augment prompts.
5. Run
   `scripts/pm-review-system.py synthesize --manifest <manifest> --responses-dir <responses-dir>`.
   Only one authenticated v4 synthesis that is `clean` for the same exact base/head/tree advances.
6. Record a disposition for every actionable finding. Accepted fixes return to an isolated worker,
   then repeat affected verification, v4 compilation, exact rendering, packet review, and synthesis
   at the new exact head.
7. Run independent Shepherd validation only after that exact-head synthesis is clean. Require
   `PROCEED` for the same identities.
8. Record compiler, packet, local Codex, disposition, Shepherd, CI, and remaining human-gate
   evidence.
9. Integrate only that reviewed and validated exact head. After integration, verify the resulting
   parent range and keep parent readiness pending until final parent gates pass.

## 6. Local Review Disposition And Shepherd

Local Codex findings are review input, not instructions. The orchestrator must:

- use the exact machine vocabulary
  `finding_disposition_values: [accepted, accepted_with_modification, declined, duplicate, deferred, needs_human]`
  for every actionable finding
- record a reason before considering the finding dispositioned
- implement accepted in-scope fixes through an isolated worker
- create or reference follow-up issues for deferred work
- re-run affected verification and fresh-context local Codex review after every head change
- run Shepherd independently after clean code review and before integration
- increment `rounds_by_range` when an accepted correction creates a new head
- stop with a human blocker when `max_correction_rounds` is exceeded
- stop on a `RETRY`, `REVERT`, or `HALT` verdict until the specified correction completes

Claude and GitHub Copilot are not required, requested, or fallback coverage in the canonical PM
route. Legacy bot-review documents remain historical references only.

## 7. Final Parent Readiness

The parent PR can move from draft to ready only when:

- every required sub-issue is integrated or explicitly deferred
- the parent PR contains the closing keywords intended for `main`
- full parent verification passes
- clean exact-head local Codex review and independent Shepherd coverage exist for every integrated
  sub-issue
- no actionable local Codex findings remain
- no human gate remains except final merge approval

The orchestrator then pings the human coordinator. It must not merge the parent PR into `main`
without human approval.

## Durable Evidence

Record orchestration state in the parent issue, parent PR body, or a committed state artifact when
the issue requires one. Use `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml` as
the field contract and write current records as `canonical_v2`. Preserve truthful historical
bot-review fields only as explicitly labeled read-only legacy history; never update them as current
coverage. Current PM records carry the v4 compiler/packet/synthesis evidence, exact-head
`local_codex`, and downstream `shepherd` evidence.
