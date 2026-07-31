# Overview

PostgreSQL is a Tier-3 native database connector. It discovers schemas/columns from
`information_schema`, snapshots tables with bounded `SELECT` reads, supports cursor-incremental
reads on a configured cursor column, and exposes five bounded reverse-ETL row/table actions.

The connector is intentionally **not** a raw SQL interface. It does not expose arbitrary SQL reads,
arbitrary SQL writes, COPY streams, extension APIs, shell/file escapes, or protocol passthrough.
CDC remains a documented stub pending the gated `pglogrepl` dependency; fixture decoder tests do
not certify live logical replication.

## Auth setup

Connection fields:

- `host` (required, string): bare hostname or IP; URL-shaped values, paths, credentials, and
  connection-string fragments are rejected.
- `port` (optional, string): TCP port, defaults to `5432`.
- `database` (required, string): database name.
- `username` (required, string): database role.
- `password` (required, secret, string): database role password; never logged.
- `sslmode` (optional, enum): `disable`, `allow`, `prefer`, `require`, `verify-ca`, or
  `verify-full`; defaults to `disable`.
- `schema` (optional, string): default schema for discovery/read/write actions; defaults to
  `public`.
- `cursor_field` (optional, string): column used for incremental reads.
- `read_limit` (optional, string): maximum rows returned per snapshot `SELECT`; defaults to
  `10000`. Set `0`, `all`, or `unlimited` to disable the configured bound. A smaller request/CLI
  limit is applied before this bound.
- `mode=fixture` (test/conformance only): validates config but short-circuits network access.

Secret fields are redacted in logs and write previews: `password`.

## Streams notes

The connector discovers catalogs and records directly from the configured service instead of using
fixed stream declarations. Stream names are runtime `schema.table` identifiers. Reads validate and
quote schema/table/cursor identifiers and bind cursor values as parameters.

## Write actions & risks

Reverse ETL writes follow plan → preview → explicit approval → execute. Preview output includes SQL
templates with placeholders only; bound values are not printed.

Declared actions:

- `insert_row`: `INSERT INTO <schema>.<table> (<columns...>) VALUES ($1, ...)`.
- `update_row`: `UPDATE <schema>.<table> SET ... WHERE <key columns...>`; key columns are required.
- `upsert_row`: bounded single-row `MERGE` using typed source values; conflict keys must also be
  present in `values` (or provided through the closed CLI shortcut fields).
- `delete_row`: `DELETE FROM <schema>.<table> WHERE <key columns...>`; destructive confirmation is
  required by the reverse-ETL plan/approval path.
- `truncate_table`: `TRUNCATE TABLE ONLY <schema>.<table>`; `confirm_phrase` must equal
  `truncate`. `CASCADE` and `RESTART IDENTITY` are not exposed.

Every write record uses a closed schema. Programmatic writes use `schema`, `table`, `values`,
`keys`, and `confirm_phrase`; CLI convenience flags map to closed shortcut fields such as
`value_column`, `value_string`, `key_column`, and `key_int`. Column/table/schema identifiers must be
plain PostgreSQL identifiers, and values must be typed through one of `value_string`, `value_int`,
`value_float`, `value_bool`, `value_null`, or `value_json`.

## Known limits

- No generic SQL/query/write command is available.
- No COPY/binary/file transfer surface is available.
- CDC/changefeed operations are blocked until a separately gated pglogrepl implementation lands.
- Fixture behavior is not live certification.
