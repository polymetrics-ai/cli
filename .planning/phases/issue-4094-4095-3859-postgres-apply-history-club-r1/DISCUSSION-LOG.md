# Discussion log — PostgreSQL apply/history club

`scripts/gsd prompt discuss-phase 4094 --auto` was resolved and executed inline.
The autonomous launch brief and the three issue contracts leave no product gray
area requiring human input.

## Decisions applied

1. Use the exact staging base and direct-PR branch named by the launch brief.
2. Treat #4094 and #4095's staging implementations as foundations already on
   the base; audit and preserve them rather than rebuilding them.
3. Fix only #3859's stated residual: the database polling adapter cannot
   currently construct the history plan because it omits the sealed route.
4. Bind source and destination driver declarations from loaded database
   definitions, never from caller text or executor-reference naming.
5. Put the new success proof on the real PostgreSQL adapter path and require
   target-row/receipt observations. Use fakes only for forbidden-route zero-I/O
   proof, where contacting the rejected engine would defeat the criterion.
6. Record #4158's known base failure without changing its production path.

No scope was deferred and no decision was escalated.
