# Overview

Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and
supports cursor-incremental reads on a configurable cursor column.

This connector discovers available streams and schemas from the configured service at runtime.

This connector is read-only; no write actions are declared.

## Auth setup

Connection fields:

- `cursor_field` (optional, string); Optional column name used for incremental reads (rows with
  cursor_field greater than the stored cursor are read, ordered by cursor_field ascending).
- `cdc_publication` (optional, string); Existing PostgreSQL publication reserved for the planned
  logical-replication CDC executor; change capture is not currently executable.
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
- Logical-replication CDC is not executable until it has bounded crash-recoverable streamed
  transaction staging.
