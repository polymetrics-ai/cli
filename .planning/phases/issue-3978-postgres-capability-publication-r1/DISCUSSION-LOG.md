# Discussion log — Issues 3978 and 3977

The autonomous task brief and both complete issue threads supplied all binding decisions. The only implementation gray area was how to close #3977's audited application-dispatch residual without weakening transaction durability.

The selected boundary is a small generic committed-CDC transaction receiver consumed by the PostgreSQL native adapter and implemented by the application warehouse path. It streams events with bounded memory, returns a durable warehouse acknowledgement only after the connection-owned WAL and derived table are durable, and supplies that same acknowledgement to checkpoint persistence. The existing per-event callback remains for connector-local tests and compatibility but is not the production application route.

Rejected alternatives:

- treating `ChangefeedExecutor` method presence or `changefeed.json` as dispatch proof;
- appending per-event directly to the visible warehouse before whole-transaction commit;
- buffering an arbitrarily large transaction in memory;
- routing CDC directly into the PostgreSQL managed target;
- keeping `cdc: true` while `App.RunETL` returns `ModeNotExecutableError`.

## Post-#4156 rebase disposition

#4156 introduced the generic closed transport route for definition-owned declarations. Its
`change_capture` regression test initially exposed that this branch's native PostgreSQL fallback
was ordered too early. Dispatch now tries a matching closed transport first and reaches the native
transaction-aware warehouse receiver only when no such transport matches. Both routes have focused
green tests. This keeps the shared fix owned by #4156 while preserving the whole-transaction receipt,
checkpoint, and source-LSN ordering required by PostgreSQL CDC.

The issue also requires an explicit `rate_limits.json`. PostgreSQL has no provider HTTP API, so the
bundle now declares `state: not_applicable` with that exact reason. This is metadata truth, not a
claim that native database resources are unbounded; the authored connector docs point to the typed
pool, batch, statement, CDC-stage, replication-slot, and WAL bounds.
