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
- obtain parent review coverage for provisionally integrated commits when the stacked sub-PR route
  did not produce coverage
- run the full integrated parent verification and no-mistakes pipeline once all children land
- mark the draft ready only after every child and parent gate is green
- preserve final merge for explicit captain approval while the parent remains green

## State machine

Record these states in the parent issue, parent PR, or a GSD artifact:

- `planned`: issue map exists; parent branch or draft PR is not ready.
- `parent_pr_open`: parent branch, deliberate seed, and draft parent PR exist before production
  work.
- `wave_ready`: one sub-issue has complete inputs and satisfied dependencies.
- `wave_in_progress`: the canonical worker is executing that sub-issue's GSD/TDD plan inline.
- `sub_pr_open`: a sub-PR targets the parent branch.
- `sub_pr_green`: local and remote checks pass.
- `sub_pr_reviewed`: required automated review coverage exists, or an allowed parent fallback is
  explicitly recorded.
- `provisionally_integrated`: the sub-PR landed on the parent branch while parent fallback review
  coverage is still pending.
- `parent_review_pending`: integrated commits still need parent review coverage or disposition.
- `parent_review_clean`: integrated commits have no unresolved actionable review finding.
- `final_verification`: all children are integrated and full parent gates are running.
- `ready_for_captain`: the parent is green and ready, but final merge is not authorized.
- `blocked`: a dependency, human gate, failed verification, or review blocker prevents progress.
- `complete`: the captain-approved parent PR merged to the default branch.

Only one `wave_in_progress` state may exist. There is no parallel worker fan-out. When multiple
waves are dependency-ready, process them deterministically and persist the remaining order in the
parent issue.

## Installed GSD lifecycle

For each ready wave:

1. map the sub-issue to a GSD phase
2. `discuss-phase` for known decisions
3. `plan-phase --tdd`
4. `execute-phase` inline through RED, GREEN, and REFACTOR
5. `verify-work`; when needed, `plan-phase --gaps`, `execute-phase --gaps-only`, and verify again
6. `code-review` with reasoned finding disposition

Every command must resolve through `scripts/gsd sources <command>`. Do not invoke the absent
`programming-loop`. Do not use GSD ship: official ship creates a PR after verification, while this
contract already requires the draft parent PR before implementation and stacked sub-PRs to that
parent.

## no-mistakes topology

Never use `--yes`.

On a sub-issue branch, run the review/test/docs/lint loop with the argv vector
`["no-mistakes","axi","run","--intent","<issue-intent>","--skip=push,pr,ci"]`. Replace the
intent placeholder element with the complete issue intent and pass the vector directly to the
process without shell interpolation. Respond to each exact finding ID with recorded rationale. Let
the pipeline apply bounded in-scope fixes, then rerun the gate.

After local gates pass, open the sub-PR with
`["gh-axi","pr","create","--base","<parent-branch>","--head","<child-branch>","--title","<conventional-title>","--body-file","<pr-body-file>"]`.
No-mistakes v1.41.2 cannot target a non-default PR base.

On the integrated parent branch, run
`["no-mistakes","axi","run","--intent","<parent-intent>"]` against its existing draft parent PR,
replacing the placeholder element with the complete parent intent and passing the vector directly
without shell interpolation.

## Child integration gate

A sub-PR may integrate into the parent branch only when:

- it targets the parent branch and uses `Refs #<sub-issue>` plus `Refs #<parent-issue>`
- targeted and issue-level verification pass
- CI checks pass
- every actionable automated review finding is resolved or explicitly dispositioned
- review coverage exists on the sub-PR, or the allowed parent-PR fallback is recorded
- the diff remains inside the sub-issue scope and ownership boundaries
- no requested-changes review or human gate remains

A specific infrastructure blocker pauses integration, records the wave as blocked, and escalates
the unblock action. It never substitutes for passing checks.

If Claude skips a non-default-base sub-PR, integration is provisional until the main-targeted parent
PR receives Claude review covering that commit range, or the documented Copilot/human fallback is
completed. A skipped, errored, rate-limited, or never-started review is not review coverage.

## Parent readiness and merge gate

After all sub-issues are integrated:

1. run full integrated tests, lint, docs, build, and issue-specific verification
2. run the full parent no-mistakes pipeline
3. ensure every integrated range has automated review coverage and all findings are dispositioned
4. update the parent issue and draft PR with GSD/TDD/review/verification evidence
5. mark the draft ready only while all checks remain green

Ready is not approval. Never infer permission from captain absence, never merge red, and never merge
the parent to the default branch without explicit captain approval while it is still green.

## Away-mode boundary

Self-answer only a routine, reversible gate fixed by an issue decision, repo contract, or explicit
standing authority; address the exact finding and record why. Auto-fix bounded code, test, and docs
findings inside scope through the active gate, then rerun it.

Pause and preserve state for product ambiguity, destructive or irreversible actions, secrets/auth
or security-boundary changes, dependencies or production impact, generic write capabilities,
reverse-ETL execute approval, quality-gate weakening, and final merge. Absence never expands
authority.
