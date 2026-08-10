---
coverage:
  - id: D1
    description: An opt-in observer finds and publishes only settled feature-branch commits.
    verification:
      - kind: integration
        ref: scripts/tests/unpushed-work-safety-net_test.py real Git worktree suite
        status: pass
    human_judgment: false
  - id: D2
    description: The observer never force-pushes, rewrites a remote branch, or pushes the remote default branch.
    verification:
      - kind: integration
        ref: divergent, final-recheck remote-change, and actual non-fast-forward tests plus force-argument wrapper
        status: pass
    human_judgment: false
  - id: D3
    description: Active Git operations, dirty work, locks, state failures, and missed schedules are visible and recoverable.
    verification:
      - kind: integration
        ref: real operation, fcntl, corrupt-state, and heartbeat tests
        status: pass
    human_judgment: false
  - id: D4
    description: Operators can explicitly enable, schedule, inspect, and disable the safety net without enabling Git hooks.
    verification:
      - kind: integration
        ref: enable/disable/status/launchd-plist tests and docs/unpushed-work-safety-net.md
        status: pass
    human_judgment: false
---

# Summary — unpushed-work safety net r1

## Delivered

- Added `scripts/unpushed-work-safety-net.py`, a standard-library Python observer that is opt-in
  per repository and suitable for an explicit 600-second launchd schedule.
- It enumerates linked worktrees, reports dirty/detached/active-operation states, resolves the
  effective push target and live default branch from the actual push URL, and only makes a pinned,
  explicit non-force feature-branch push.
- It uses a kernel-released shared lock, atomic durable state, a pre-push rate-floor record,
  event logging, health status, and visible failures/retries rather than a stale lease.
- Added a real-Git acceptance harness and wired it into `make verify` through
  `unpushed-work-safety-net-check`.
- Added explicit installation/removal and hazard-rationale documentation without changing the
  existing `core.hooksPath` decision.

## Recorded-hazard disposition

- **Avoided structurally:** operation cleanup/past-operation reconstruction (R1/R6/R7/R11/R12/R13)
  and competing/stale hook leases (R3/R8/R9). The observer has no post-commit callback, synthetic
  operation marker, detached child, or durable lease.
- **Defended:** delayed observation-to-push timing (R4) with immediate rechecks, a pinned SHA, a
  non-force refspec, and a common-Git-database push; the final-recheck remote-change test proves
  a change is deferred then recovered; default/push-target ambiguity (R2/R5/R10)
  with live discovery from the effective single push URL and fail-closed errors.

## Remaining gate

The feature is locally reviewed and verified, then awaits the coordinator-owned no-mistakes
pipeline, PR creation, remote CI, automated review, and human merge gate. This worker has not
pushed to `main` or merged anything.
