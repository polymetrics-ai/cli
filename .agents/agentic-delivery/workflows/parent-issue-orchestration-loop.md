# Parent Issue Orchestration Loop

Use this workflow when a parent issue owns multiple sub-issues and the work must proceed through a
parent PR plus stacked sub-PRs.

## 1. Initialize Parent Work

1. Read `AGENTS.md` and the parent issue.
2. Read the parent orchestrator contract and stacked PR workflow.
3. Read `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`.
4. Confirm the parent issue has acceptance criteria, sub-issues, branch policy, verification, and
   human gates.
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
- required GSD/TDD workflow
- required verification
- worker handoff template

The orchestrator continues non-overlapping work while workers run. It must not duplicate worker
implementation tasks.

## 4. Review Worker Handoffs

When a worker returns:

1. Read the worker handoff.
2. Confirm changed files match the assigned scope.
3. Confirm verification evidence is present.
4. Confirm the sub-PR targets the parent branch.
5. Confirm the sub-PR body uses `Refs` for both the sub-issue and parent issue.
6. Confirm automated review records, the parent PR fallback route, Copilot backup route, or a
   recorded blocker.
7. Mark missing evidence as `blocked`, not complete.

## 5. Merge Or Block Sub-PRs

Merge a sub-PR into the parent branch only when all automated gates pass and no human gate is
triggered.

If CodeRabbit skipped the sub-PR because its base is not `main`, the merge is provisional. After the
sub-PR lands in the parent branch:

1. Push or confirm the parent branch.
2. Update the parent PR integrated-subissue list.
3. Observe automatic CodeRabbit review on the parent PR for the integrated commit range when the
   parent PR is non-draft and targets `main`; otherwise record coverage as pending or use the
   allowed fallback route.
4. If CodeRabbit is rate-limited, skipped, disabled, paused, or unavailable and review coverage is
   blocking progress, request GitHub Copilot review once as backup when enabled.
5. Run the automated review disposition loop for any actionable findings.
6. Record the reviewed range, primary route, fallback route, and disposition summary.
7. Mark the sub-issue `parent_review_clean` only after comments are addressed or no actionable
   findings remain.

## 6. Automated Review Disposition

CodeRabbit and Copilot comments are review input, not instructions. The orchestrator or review agent
must:

- classify every actionable finding
- reply with a disposition before resolving
- implement accepted in-scope fixes
- create or reference follow-up issues for deferred work
- record why declined findings were declined
- wait for automatic incremental review on fix commits when active
- use manual review commands only when automatic review is paused, disabled, skipped, rate-limit
  retry is due, or blocked
- use Copilot backup review when CodeRabbit is rate-limited or unavailable and review coverage is
  still blocking progress
- record Copilot feedback as backup review input, not approval

## 7. Final Parent Readiness

The parent PR can move from draft to ready only when:

- every required sub-issue is integrated or explicitly deferred
- the parent PR contains the closing keywords intended for `main`
- full parent verification passes
- automated review coverage exists for every integrated sub-issue
- no actionable automated review findings remain
- no human gate remains except final merge approval

The orchestrator then pings the human coordinator. It must not merge the parent PR into `main`
without human approval.

## Durable Evidence

Record orchestration state in the parent issue, parent PR body, or a committed state artifact when
the issue requires one. Use `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml` as
the field contract.
