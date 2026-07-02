# Overview

PostgreSQL is a Tier-3 native database source connector (design §B.7): it speaks the PostgreSQL
wire protocol directly via [pgx](https://github.com/jackc/pgx), not a declarative HTTP bundle. It
discovers tables and columns from `information_schema` at runtime (`capabilities.dynamic_schema:
true` — there is no static `streams.json`), snapshots a table with `SELECT * FROM <schema>.<table>`,
and supports cursor-incremental reads on a configurable column. It is read-only for wave0 parity
with the legacy `internal/connectors/postgres` package.

## Auth setup

Provide `host`, `database`, and `username` in config, and the database role's `password` as a
secret (never logged, never included in `Definition().Spec` verbatim — only the `x-secret: true`
marker is exposed). `sslmode` defaults to `disable`; set it to `require`, `verify-ca`, or
`verify-full` for encrypted/verified connections. `host` must be a bare hostname or IP — a
URL-shaped value (scheme, path, query, credentials) is rejected to bound SSRF /
connection-string-injection risk.

## Streams notes

Streams are discovered dynamically, one per base table in the target `schema` (default `public`),
named `<schema>.<table>`. Primary keys are derived from `information_schema.table_constraints`
when a `PRIMARY KEY` constraint exists. Incremental reads require a `cursor_field` config value
naming an existing column; rows are read in ascending `cursor_field` order, filtered to values
greater than the stored cursor. A `read_limit` config value (default 10000) bounds how many rows a
single snapshot `Read` call returns; set it to `0`, `all`, or `unlimited` to disable the bound.

## Write actions & risks

None. This is a read-only source connector — `capabilities.write` is `false` and `Write` always
returns `ErrUnsupportedOperation`.

## Known limits

CDC (change data capture) is a **documented stub**: `ReadCDC` returns `ErrUnsupportedOperation`
because full logical-replication CDC requires the `pglogrepl` dependency, which is not present in
`go.mod` (a gated add pending approval). The recorded plan for a future implementation: enable
`wal_level=logical` on the server, create a logical replication slot with the `pgoutput` plugin,
create a `PUBLICATION` for the target tables, and stream `START_REPLICATION` messages, persisting
the last committed LSN as read state. A `mode=fixture` config value short-circuits all network
access (Check succeeds, Catalog returns two canned tables, Read emits canned rows) — this is a
test/conformance-harness affordance only and must never be set in production config.
