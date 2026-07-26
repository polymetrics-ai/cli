# Parent integration and review

The parent orchestrator owns this procedure. Workers provide immutable evidence and never promote
their own heads unless the orchestrator explicitly delegates that action.

## Pre-review identity

1. Fetch default, parent, worker, and PR refs.
2. Record the worker head and reviewed commit range against the current parent merge-base.
3. Confirm the PR targets the parent branch, references the issue and parent, and does not claim
   default-branch closure.
4. Run independent VERIFY against that exact head.

A review of a different head, a skipped/failed run, or a stale base is not exact-head coverage.
Any new commit requires review of the new range or a documented equivalent range that includes it.

## Findings and active constraints

Every actionable finding requires a fix or reasoned disposition. Resolve a thread only after that
disposition exists. Follow the active task and repository review constraints; if the configured
route is prohibited, unavailable, or cannot bind to the required head, stop and record the conflict
rather than silently substituting or claiming coverage.

Fresh exact-head review complements but does not replace:

- independent verification;
- Shepherd trajectory validation of plan, implementation, evidence, and scope;
- parent-branch reruns after promotion;
- final human review and merge approval.

## Promotion

Before promotion, re-fetch and compare the worker and parent heads to recorded identities. Reject
head drift, unreviewed commits, dependency changes without approval, write-scope collisions, and
missing dispositions. Integrate with the repository's ordinary parent-owned method; never force
push the parent.

After promotion, rerun affected focused checks plus parent gates, record the exact parent commit,
and update shared orchestration state. GitHub issue OPEN state may remain until the parent reaches
the default branch.

## Historical process debt

When code is already integrated but a stacked PR, review, or evidence is missing:

- inspect and verify the actual parent commit range;
- obtain exact-range review and disposition or an explicit authorized waiver;
- record `integrated_review_debt` until the debt is resolved;
- do not reimplement the feature, manufacture a historical PR, or rewrite commit history.

## Parent readiness

The parent is ready for a human gate only when all required slices are parent-satisfied or explicitly
human-deferred, review debt is closed, full parent verification passes, documentation/help parity is
current, dependency and safety decisions are recorded, and a final remote-head drift check is clean.
The parent PR remains draft until a human authorizes readiness, and merge to the default branch is
human-only.
