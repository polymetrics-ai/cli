# Summary

Status: historical MVP; superseded by
`../479-shepherd-production-matrix/VERIFICATION.md` (17/17 functional PASS at code head
`91692415`).

Issue #479 now delivers the bounded in-process Shepherd controller. `/pm-shepherd start` and
`resume` load a repository manifest, schedule dependency-safe non-colliding children at bounded
parallelism, run implementation/verification/review in embedded Pi AgentSessions, persist status,
join stop, resume unfinished work, and end in `waiting_human` without a merge-main capability.
The explicit schema-v1 read-only canary remains separately routable.

Each autonomous child owns a separate embedded runtime, while the scheduler and scoped workspace
tools enforce disjoint writes in the current working tree. Implementation uses
`gpt-5.6-sol/high`; verification and review use `gpt-5.6-sol/xhigh`.

The single blocker-only review is complete and its four MVP blockers were fixed: the runtime now
receives an exact plain workspace capability, parallel child mutation leases are isolated,
the concurrency test no longer depends on subprocess completion order, and status follows an
active canary instead of stale autonomous state. A child failure also aborts and joins live
siblings before settlement.

The items that were deliberately deferred from this historical MVP—GitHub issue/PR publication,
isolated child worktrees, exact child integration, the production human-decision adapter, final
human-merge observation, and exhaustive crash/CAS coverage—are implemented and verified in the
production-matrix phase. No parent-to-main merge was attempted.
