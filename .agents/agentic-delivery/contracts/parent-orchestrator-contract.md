# Parent Job Ownership Contract

This compatibility path no longer defines or activates a dedicated parent-orchestrator role. Use
it when one delivery job owns a parent issue, outcome-sized sub-issues, a parent branch, and stacked
PRs. The one canonical worker owns that state inline and follows
`.agents/agentic-delivery/canonical/delivery-contract.json`.

GitHub issues, branches, PRs, review records, and GSD artifacts are the durable handoff. Never spawn
an orchestrator, shepherd, planner, reviewer, verifier, GSD role, or additional worker for the job.
If the active worker stops, its successor resumes from those artifacts rather than from a role
handoff.

## Required input

The parent issue must provide:

- objective, background, acceptance criteria, and non-goals
- parent branch name and draft parent PR URL or exact blocker
- a navigable sub-issue roster whose children are named as deliverable outcomes, sized to one worker
  session, and linked to explicit dependencies and decisions
- verification commands and integrated parent gates
- human gates and source links
- the automated review coverage route

The worker also reads:

- `AGENTS.md`
- `.agents/agentic-delivery/canonical/delivery-contract.json`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/claude-review-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`

## Activation and ownership

When a task references a parent issue, sub-issues, stacked PRs, a parent branch, a parent PR, or
automated review coverage, the current canonical worker owns the parent state as part of the same
job. Activation does not create another role.

The worker must:

- create or confirm the parent branch from the default branch
- create a deliberate parent seed and draft parent PR before any production edit
- maintain the parent issue as the navigable index and record dependency/decision state
- process one ready sub-issue at a time through the installed GSD lifecycle
- keep each sub-issue to one primary branch, scoped diff, and sub-PR to the parent branch
- decide child integration only after the checks and review contract below passes
- obtain parent review coverage for provisionally landed commits when the stacked sub-PR route
  did not produce coverage
- run the full integrated parent verification and no-mistakes pipeline once all children land
- mark the draft ready only after every child and parent gate is green
- preserve final merge for explicit captain approval while the parent remains green

## State machine

The ordered IDs under `state_machine.steps` in
`.agents/agentic-delivery/canonical/delivery-contract.json` are the sole state vocabulary. Record
those IDs verbatim in the parent issue, parent PR, or GSD artifact; do not maintain a second state
machine here. The canonical single-worker and no-delegation fields require ready waves to run one
at a time, with any remaining order persisted in durable parent state.

## Installed GSD lifecycle

The exact lifecycle and ship exclusion live under `gsd` in the canonical source. Resolve and
execute that sequence through
`.agents/agentic-delivery/references/gsd-pi-adapter.md`; do not maintain a second command list here.

## no-mistakes topology

The `no_mistakes` object in the canonical source owns the verified version, forbidden flags, exact
child and parent argv vectors, gate-response boundary, and sub-PR action. Execute those vectors
directly after replacing placeholder elements; never reconstruct them as shell strings or copy them
into this compatibility document.

## Child integration gate

The `tracker.integrate_when` array in the canonical source is the sole integration checklist.
Evaluate every current entry; the stacked workflow owns PR-shape and review-routing mechanics. Do
not reproduce either list here.

A specific infrastructure blocker pauses integration, records the exact blocker without advancing
the canonical state, and escalates the unblock action. It never substitutes for passing checks.

If Claude skips a non-default-base sub-PR, do not record the canonical `integrate_sub_pr` state. A
technical parent-branch landing used only to expose the commit range for parent-PR fallback review
is provisional transport, not completed integration. Obtain Claude coverage on that range, or
complete the documented Copilot/human fallback, before advancing the canonical state. A skipped,
errored, rate-limited, or never-started review is not review coverage.

## Parent readiness and merge gate

The canonical `tracker.ready_when` and `tracker.final_merge` fields are the sole readiness and merge
criteria. Record evidence for every current entry while executing the canonical
`integrated_parent_gates`, `ready_parent`, and `captain_merge` states; do not restate the criteria
here.

## Away-mode boundary

The canonical `authority` object is the sole owner of self-answer, auto-fix, pause, and invariant
rules. Apply its IDs and instructions verbatim; do not create local variants in this compatibility
document.
