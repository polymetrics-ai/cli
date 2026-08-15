# Discussion log — PostgreSQL production transport wiring club

Generated from `scripts/gsd prompt discuss-phase cli-pg-destination-wiring-club-r1 --auto` and executed inline after reading all three issues and amendments, #4090, PR #4156, the prior component phase artifacts, connector canon, migration conventions, architecture design, runtime integration, issue contract, and CLI parity contract.

## Existing state

- #3982 already supplies `DatabaseDriver`, managed-target provisioning/control/ledger, five phase-one write modes, rollback/unknown-commit behavior, and real PostgreSQL driver tests. It has no destination declaration, factory, or non-test `NewDatabaseDriver` caller.
- #3983 already supplies immutable `ChangeDeliveryWorkset`, workset-bound plan/preview/one-shot approval, `ChangeDeliveryExecutor`, and receipt-gated baselines. Tests hand-build it; production transport never does.
- #3979 already supplies `BootstrapCDC`, exported-snapshot handover, pgoutput-v2 transaction staging, gap-free checkpoints, and live tests. Production App never calls it.
- #4093 supplies definition loading, exact factories, conformance allow-listing, App registration, warehouse staging/reopen, destination apply/read-back, and checkpoint-after-ack ordering.

## Gray areas resolved

1. Dynamic PostgreSQL relations cannot honestly be represented by the literal `snapshot` stream used by the registration-only proof. The production contract needs an explicitly validated dynamic-stream declaration and executor-side catalog/identifier checks.
2. Bootstrap's native system/timeline identity differs from App's persisted credential/connection resume identity. The transport source must translate only the envelope identity at its boundary while retaining the native bootstrap barrier, LSN, publication, relation, schema fingerprint, and generation needed for resume validation.
3. The connection warehouse stage must expose its own immutable Parquet artifact to the destination without accepting a caller path. The receipt hash and reopen checks remain authoritative; the destination rehashes again when deriving `ChangeDeliveryWorkset`.
4. Managed-target ownership is derived from persisted structural IDs and never from the destination credential, display name, stream map key, or destination table text.
5. A long-lived CDC read is ended through context cancellation. Already acknowledged transactions remain resumable; an in-flight transaction and unacknowledged candidate replay. A full-snapshot production run remains the decisive finite binary proof; the bootstrap binary proof additionally observes durable effects before deliberate cancellation.

## Scope exclusions

No write capability flip, generic database command, direct PostgreSQL hop, automatic schema evolution, physical-absence delete, history target advertisement, unrelated connector change, #4125 fix, or #4158 fix.
