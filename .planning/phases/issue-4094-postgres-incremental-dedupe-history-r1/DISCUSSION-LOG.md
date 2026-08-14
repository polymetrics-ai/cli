# Discussion log — Issue 4094

## Inputs resolved

- Issue #4094 owns the native PostgreSQL history target and cites #3810 for
  generic history semantics, #3859 for generic strategies, and #3982 for the
  managed PostgreSQL driver/session.
- Issue #4097 makes the PostgreSQL-to-PostgreSQL route exclusive and requires
  every other route to refuse before I/O with a route-specific typed reason.
- The task brief requires proof of zero provider/database operations on each
  refusal and real database assertions of history row state.

## Decisions

1. Keep the implementation in the existing database/PostgreSQL managed-target
   seams; do not create a second mapping or receipt type.
2. Test route admission before driver/session acquisition, with instrumented
   fakes proving zero reads, writes, and mutations on each rejected route.
3. Add a tagged live PostgreSQL proof that queries history validity-window and
   soft-delete state rather than accepting process completion as evidence.
4. Preserve the mandated plan → preview → approval → execute guard for the
   successful managed-target write path.
