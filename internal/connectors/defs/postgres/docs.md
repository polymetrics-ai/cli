# Overview

Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and
supports cursor-incremental reads on a configurable cursor column, and consumes change events
through PostgreSQL logical replication. CDC uses a connector-owned, source-bound replication slot
and commits an LSN only after validated durable downstream acknowledgement, so restart delivery is at-least-once.

This connector discovers available streams and schemas from the configured service at runtime.

This connector is read-only; no write actions are declared.

## Auth setup

Connection fields:

- `cursor_field` (optional, string); Optional column name used for incremental reads (rows with
  cursor_field greater than the stored cursor are read, ordered by cursor_field ascending).
- `cdc_publication` (optional, string); Existing PostgreSQL publication used for logical-
  replication CDC. Required only for CDC; it must include the selected table and publish only insert,
  update, and delete changes.
- `database` (required, string); Database name to connect to.
- `host` (required, string); Bare hostname or IP of the PostgreSQL server (no scheme, path, or
  credentials - a URL-shaped value is rejected).
- `mode` (optional, string); allowed values `fixture`.
- `password` (optional, secret, string); Database role password. Never logged.
- `port` (optional, string); TCP port, 1-65535. Defaults to 5432 when omitted.
- `read_limit` (optional, string); Maximum rows returned per Read snapshot SELECT. Defaults to
  10000; set to 0, all, or unlimited to disable the bound.
- `schema` (optional, string); PostgreSQL schema to discover tables from. Defaults to public.
- `sslmode` (optional, string); allowed values `disable`, `allow`, `prefer`, `require`, `verify-ca`,
  `verify-full`; libpq sslmode. Defaults to disable when omitted.
- `username` (required, string); Database role used to authenticate.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

## Streams notes

The connector discovers catalogs and records directly from the configured service instead of using
fixed stream declarations.

## Write actions & risks

This connector is read-only. Read behavior: low.

## Known limits

- Schemas and stream availability depend on the configured service at runtime.
- CDC requires a real PostgreSQL source with `wal_level=logical`, a role permitted to use logical
  replication, and an existing `cdc_publication` that contains the selected table.
- CDC supports publications that exclude `TRUNCATE` changes and selected relations without descendant
  tables.
- CDC slots are derived from the PostgreSQL system identity, database, and fully qualified stream;
  teardown drops only that inactive connector-owned slot. Do not delete a slot while another CDC
  reader is active.
