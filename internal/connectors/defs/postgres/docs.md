# Overview

Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and
supports cursor-incremental reads on a configurable cursor column, and consumes change events
through PostgreSQL logical replication. CDC maps insert, update, delete, and full-table truncate
events for the selected stream. It uses a connector-owned, source-bound replication slot and
commits an LSN only after downstream acknowledgement, so restart delivery is at-least-once.

This connector discovers available streams and schemas from the configured service at runtime.

This connector is read-only; no write actions are declared.

## Auth setup

Configure a bare `host`, `database`, and `username`. `port` defaults to 5432. `password` is a
secret field and is never logged. Do not put credentials in a host or URL-shaped value.

`sslmode` uses the same transport-security shape as the MySQL connector and is honestly enforced
for both local and remote servers:

| `sslmode` | Encrypts | Falls back to plaintext | Verifies chain | Verifies server name |
| --- | --- | --- | --- | --- |
| `disabled` (default) | no | n/a | no | no |
| `preferred` | when offered | yes, only when the server advertises no TLS | no | no |
| `required` | yes | never | no | no |
| `verify-ca` | yes | never | yes | no |
| `verify-identity` | yes | never | yes | yes |

`preferred` is the only canonical mode that may fall back. `required` and both verifying modes
fail closed when the server offers no TLS. `sslrootcert` names an absolute PEM CA path for the
verifying modes and defaults to the host trust store; `sslservername` overrides the verified name
under `verify-identity`, for example when connecting to an IP address. Compatibility spellings
`disable`/`prefer`/`require`/`verify-full` and `verify_ca`/`verify_identity` are accepted. The
legacy libpq `allow` spelling is also accepted unchanged for existing PostgreSQL configurations; it
tries plaintext before TLS, so new portable configurations should use `preferred` instead.

Connection fields:

- `cursor_field` (optional, string); Optional column name used for incremental reads (rows with
  cursor_field greater than the stored cursor are read, ordered by cursor_field ascending).
- `cdc_publication` (optional, string); Existing PostgreSQL publication used for logical-
  replication CDC. Required only for CDC; it must include the selected table.
- `database` (required, string); Database name to connect to.
- `host` (required, string); Bare hostname or IP of the PostgreSQL server (no scheme, path, or
  credentials - a URL-shaped value is rejected).
- `mode` (optional, string); allowed values `fixture`.
- `password` (optional, secret, string); Database role password. Never logged.
- `port` (optional, string); TCP port, 1-65535. Defaults to 5432 when omitted.
- `read_limit` (optional, string); Maximum rows returned per Read snapshot SELECT. Defaults to
  10000; set to 0, all, or unlimited to disable the bound.
- `schema` (optional, string); PostgreSQL schema to discover tables from. Defaults to public.
- `sslmode` (optional, string); allowed values `disabled`, `preferred`, `required`, `verify-ca`,
  `verify-identity`, `disable`, `allow`, `prefer`, `require`, `verify-full`, `verify_ca`,
  `verify_identity`; transport security. Defaults to disabled when omitted.
- `sslrootcert` (optional, string); Absolute path to a PEM CA bundle used by verify-ca and
  verify-identity. Defaults to the host trust store when omitted.
- `sslservername` (optional, string); Server name to verify under verify-identity when it differs
  from host, such as an IP-addressed endpoint fronting a named certificate.
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
- CDC slots are derived from the PostgreSQL system identity, database, and fully qualified stream;
  restart requires the matching durable checkpoint. An existing slot without that checkpoint is
  refused until it is explicitly torn down and re-snapshotted. Teardown drops only that inactive
  connector-owned slot. Do not delete a slot while another CDC reader is active.
