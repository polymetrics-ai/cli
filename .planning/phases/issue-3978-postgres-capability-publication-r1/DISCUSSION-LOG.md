# Discussion log — Issues 3978 and 3977

The autonomous task brief and both complete issue threads supplied all binding decisions. The only implementation gray area was how to close #3977's audited application-dispatch residual without weakening transaction durability.

The selected boundary is a small generic committed-CDC transaction receiver consumed by the PostgreSQL native adapter and implemented by the application warehouse path. It streams events with bounded memory, returns a durable warehouse acknowledgement only after the connection-owned WAL and derived table are durable, and supplies that same acknowledgement to checkpoint persistence. The existing per-event callback remains for connector-local tests and compatibility but is not the production application route.

Rejected alternatives:

- treating `ChangefeedExecutor` method presence or `changefeed.json` as dispatch proof;
- appending per-event directly to the visible warehouse before whole-transaction commit;
- buffering an arbitrarily large transaction in memory;
- routing CDC directly into the PostgreSQL managed target;
- keeping `cdc: true` while `App.RunETL` returns `ModeNotExecutableError`.
