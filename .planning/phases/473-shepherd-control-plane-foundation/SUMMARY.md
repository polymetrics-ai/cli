# Summary: #473

Status: implementation and all local/root gates green; review-driven blockers corrected; two
independent final reviews CLEAN.

The foundation now has first-wins lifecycle cancellation, bounded joined SDK teardown, stable
repository/worktree/PR evidence, private durable state, exact state invariants, root path+inode
pinning, and an append-only lease journal that revalidates the authoritative highest epoch before
accepting or cleaning anything. Orphan writers cannot return a lease, reuse an owner token, or
delete a newer epoch. Reserved epoch names and maximum exhaustion fail closed.

Exact PR evidence is explicitly bound to the local repository. Crash recovery marks every
unfinished lane interrupted, stop cannot be acknowledged before a durable checkpoint exists, and
timed-out AgentSessions remain quarantined with mutator ownership fenced until genuine settlement.
Shutdown during initialization produces an interrupted checkpoint before lane dispatch, and a
rejected settlement permanently poisons that runner instance rather than admitting more work.
Persisted lane combinations now match controller output, extended-year timestamps compare by
instant, interrupted checkpoints contain no pending work, and terminal provider/model drift fails
closed after execution.

The macOS boundary is deliberately trusted same-user local automation with a private state root.
The implementation does not claim hostile same-UID swap-and-restore resistance without native
descriptor-relative operations.

This issue is only the control-plane foundation slice; it is not the product boundary and it does
not preserve the abandoned read-only/Go-Shepherd design. The complete autonomous replacement adds
policy, scoped mutating AgentSessions, typed Git, durable GitHub decisions, parent orchestration,
shared controller wiring, recovery, and the #397/#438 canary through #474-#481.
