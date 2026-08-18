# GREEN — live PostgreSQL system-schema scope

**Candidate:** fresh `pm` rebuilt from the guarded #4070 source.
**Boundary:** the same unique run-owned loopback PostgreSQL fixture used for
RED, with a newly opened temporary-table session. No ambient PostgreSQL
credential or shared database was read.

## Focused tests

```sh
go test -count=1 ./internal/connectors/native/postgres
go test -tags=databaseintegration -count=1 -run '^$' \
  ./internal/connectors/native/postgres
```

Both passed. The first runs the table-driven exact/prefix pre-connection scope
test; the second compiles the full real PostgreSQL integration path under its
opt-in build tag.

## Fresh-binary production proof

1. Rebuilt `pm` from this worktree and ran `pm help catalog` before use.
2. `pm catalog refresh --connection pg4070-alpha --json` still returned two
   live `audit_alpha_4070` base tables with their server-derived fields and
   primary keys.
3. `pm catalog refresh --connection pg4070-beta --json` still returned one
   materially different `audit_beta_4070.accounts` relation with its composite
   primary key. Independent `information_schema` queries matched both schemas'
   column order, nullability, and primary/unique-key ordering.
4. The enum-backed `audit_unsupported_4070.events` still returned the explicit
   unsupported-catalog-shape error instead of a partial/static catalog.
5. A separate session created
   `pg_temp_3.catalog_scope_temp_green_4070`; PostgreSQL independently observed
   the relation while that session remained open. `pm catalog refresh` for its
   configured `pg_temp_3` schema returned the named scope error, not a stream.
6. The same `pm catalog refresh` error was observed for live `pg_catalog`,
   `information_schema`, `pg_toast`, `pg_toast_4070`, and the held `pg_temp_3`.
   A separate `pg_catalog` credential pointed at an unused loopback port and
   still returned the named scope error, proving the PM path rejects before it
   can open a pool/attempt transport.
7. Closing the temporary-table session left zero matching temporary relations
   in `pg_catalog.pg_class` (`count = 0`).

All scope errors had exactly this safe, identifier-free message:

```text
postgres catalog schema is reserved for PostgreSQL system objects
```

The two allowed catalog refreshes use the source's live system-catalog path;
they are not fixture results or connector JSON table/column metadata. No
connector code or JSON schema changed between alpha and beta.
