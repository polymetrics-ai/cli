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
was ordered too early when it had not yet proved an exact changefeed executor. The interim ordering
tried generic transport first; the binary-entry test below then showed descriptor presence is not
the same as a matching transport mode. The final selector first proves an exact implemented
changefeed before choosing native CDC, so the #4156 generic test and the PostgreSQL binary test are
both green. This keeps the shared fix owned by #4156 while preserving the whole-transaction receipt,
checkpoint, and source-LSN ordering required by PostgreSQL CDC.

The binary-entry red test refined that disposition: descriptor presence alone is not a matching
closed transport. PostgreSQL declares only `full_append` and `full_overwrite` through its bounded
snapshot transport, so generic preflight rejected `change_capture` before the native fallback. The
final selector gives precedence only to an exact implemented `ChangefeedExecutor` for the admitted
source-only CDC-to-local-warehouse route. A source without that exact changefeed still follows the
#4156 generic transport route and its full fail-closed preflight. No transport descriptor or
registry validation changed.

The issue also requires an explicit `rate_limits.json`. PostgreSQL has no provider HTTP API, so the
bundle now declares `state: not_applicable` with that exact reason. This is metadata truth, not a
claim that native database resources are unbounded; the authored connector docs point to the typed
pool, batch, statement, CDC-stage, replication-slot, and WAL bounds.
