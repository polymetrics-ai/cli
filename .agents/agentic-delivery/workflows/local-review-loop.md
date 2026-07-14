# Local automated review loop

Use this workflow after implementation is complete and local verification has passed. Automated
review coverage is local evidence from an independent reviewer, verifier, debugger, or
security-auditor pass in the active agent runtime. Remote PR-bot review is not part of the default
gate.

## Required steps

1. Record the exact candidate head SHA, issue, branch, base branch, and local verification already
   run.
2. Run at least one independent local review pass appropriate to the change:
   - `reviewer` for general code quality and scope;
   - `security-auditor` for auth, secrets, filesystem, external effects, or untrusted input;
   - `verifier` for command evidence and reproducibility;
   - `debugger` for failing or flaky behavior.
3. The review pass must be read-only unless it is explicitly launched as a fix worker with its own
   bounded scope. Reviewer prompts must forbid credentials, GitHub mutations, merges, and broad
   cleanup.
4. Record review coverage in the phase artifact, PR body, or worker handoff:
   - reviewer role and runtime;
   - exact head or diff range reviewed;
   - files/scope reviewed;
   - findings summary;
   - disposition for every actionable finding;
   - follow-up issue for deferred work, if any.
5. Implement accepted in-scope fixes, rerun focused tests, and rerun a local review pass when the
   fixes materially change the reviewed code.
6. Repeat until no unresolved actionable local review findings remain, or record a human-gated
   blocker.
7. Stop at final human gates, including parent PR merge to `main`.

## Stacked PRs

For stacked PRs, the sub-issue worker records local review coverage on the sub-issue head. The
parent orchestrator records local review coverage for integrated parent-branch batches when needed.
A parent PR to `main` is still required as the human integration target, but remote PR-bot review is
not required for sub-PR or parent-PR commit ranges.

## Hard stops

Stop for human approval before acting on any review finding that requires auth-scope changes,
secrets, new dependencies, destructive external actions, production deploys, broad generated
rewrites, quality-gate reductions, generic write tools, reverse ETL execution, or merging to `main`.
