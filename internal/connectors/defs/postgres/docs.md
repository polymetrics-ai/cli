# Overview

Reads PostgreSQL tables from a dynamically discovered catalog, snapshots tables, and supports
cursor-incremental reads on a configurable cursor column. PostgreSQL logical-replication
change capture is deliberately planned, not executable: it requires PostgreSQL 14+ `pgoutput`
protocol-v2 streaming, a bounded crash-recoverable per-transaction stage, `StreamAbort` discard,
a named `TransactionStageLimitExceeded` outcome with no source acknowledgement, and a durable
receipt for the complete transaction before any source LSN acknowledgement.

Stream availability is discovered at runtime; see [Streams notes](#streams-notes) for the configured
schema's scope and permission rules.

This connector is read-only; no write actions are declared.

## Warehouse snapshot transport

For warehouse-mediated sync, PostgreSQL provides one closed, live-only bounded
snapshot source. It accepts the logical `snapshot` stream only for
`full_append` and `full_overwrite`; it is neither a caller-authored SQL surface
nor a fallback to the compatibility `Connector.Read` path. Polling,
incremental modes, and change capture remain separate, non-executable source
paths.

The source identity must name one `database.schema.relation`. It discovers that
relation's typed catalog and reads finite pages, ordered by a declared non-null
primary or unique key, in one read-only repeatable-read transaction. Every
page supplies a candidate checkpoint bound to the source identity, typed-catalog
schema fingerprint, and PostgreSQL snapshot barrier. A prior full-snapshot
checkpoint cannot be resumed.

Before any source I/O, transport preflight requires the connector definition to
name its exact native database executor and the registry to have registered it.
Generic App composition never infers this source from broad connector capabilities.

## Auth setup

Configure a TCP `host`, `database`, and `username`. `port` defaults to 5432. Live connections use
password authentication: `password` is required, secret, and never logged. `mode=fixture` does
not open a source connection and does not require a password. Peer/socket and client-certificate
authentication, including ambient `PGSSLCERT`/`PGSSLKEY` values and the default PostgreSQL client
certificate pair, are unsupported and rejected during connection validation. Do not put credentials
in a host or URL-shaped value.

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
- `cdc_publication` (optional, string); Reserved for the planned logical-replication change
  capture path. It is not invoked while CDC is non-executable.
- `database` (required, string); Database name to connect to.
- `host` (required, string); TCP hostname or IP of the PostgreSQL server (no scheme, path, or
  credentials - a URL-shaped value is rejected). Unix-socket/peer authentication is unsupported.
- `mode` (optional, string); allowed values `fixture`.
- `password` (conditionally required, secret, string); Required for live password authentication.
  Fixture mode does not require it. Never logged.
- `port` (optional, string); TCP port, 1-65535. Defaults to 5432 when omitted.
- `read_limit` (optional, string); Maximum rows returned per Read snapshot SELECT. Defaults to
  10000; set to 0, all, or unlimited to disable the bound.
- `schema` (optional, string); PostgreSQL user/application schema to discover tables from. Defaults
  to public. `pg_catalog`, `information_schema`, `pg_toast`, `pg_toast_*`, and `pg_temp_*` are
  rejected before a live catalog connection is opened.
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

The connector discovers its catalog from PostgreSQL system catalogs on the configured service rather
than using fixed stream declarations. Discovery is limited to ordinary and partitioned base tables in one
allowed user/application `schema` (default `public`); PostgreSQL-owned `pg_catalog`,
`information_schema`, `pg_toast`, `pg_toast_*`, and `pg_temp_*` schema names are rejected before a
pool is opened. Views and all other relation kinds are excluded.

The configured role must have `USAGE` on that schema and either table-level `SELECT` or `SELECT`
on every non-dropped column. Relations that do not meet those permissions are omitted. If no eligible
relation remains, catalog discovery returns an error rather than advertising an unreadable stream.

Catalog discovery preserves only lossless typed metadata. For an otherwise eligible relation, an
unsupported relation shape, identifier, or PostgreSQL type shape fails the catalog request instead of
being coerced to a generic field type or returned as a partial catalog.

## Write actions & risks

This connector is read-only. Read behavior: low.

## Known limits

- Logical-replication CDC is planned and fails closed before opening a replication connection,
  creating/reusing a slot, consuming WAL, or advancing a checkpoint. It will not be advertised
  until PostgreSQL 14+ `pgoutput` protocol-v2 streaming can stage each transaction privately under
  a hard byte/record quota, discard `StreamAbort`, return a named
  `TransactionStageLimitExceeded` outcome without acknowledging the source, and acknowledge the
  source only after a whole-transaction durable downstream receipt.
- Cursor or timestamp reconciliation is not a CDC fallback: it cannot faithfully recover hard
  deletes or transaction history. A stage-limit outcome must require explicit retry or connector-
  owned teardown/rebootstrap, with source slot health made visible.
