# Overview

Reads PostgreSQL tables from a dynamically discovered catalog, snapshots tables, and supports
cursor-incremental reads on a configurable cursor column. PostgreSQL logical-replication change
capture uses PostgreSQL 14+ `pgoutput` protocol-v2 streaming. It keeps each transaction in a
bounded crash-recoverable private stage, discards `StreamAbort`, delivers only after
`StreamCommit`, and persists a whole-transaction durable receipt before checkpointing and
acknowledging the source LSN.

Stream availability is discovered at runtime; see [Streams notes](#streams-notes) for the configured
schema's scope and permission rules.

This connector is source-only. Its write capability remains false: the typed PostgreSQL
managed-target driver is not registered as a production destination, and the compatibility
`Connector.Write` method remains unsupported.

## Warehouse resumable polling transport

For warehouse-mediated sync, PostgreSQL's declared source executor is the
catalog-bound `postgres_polling_watermark` adapter. It uses the shared polling
source executor for `full_overwrite`, `full_append`, `incremental_append`,
`incremental_upsert`, `incremental_dedupe`, and `incremental_dedupe_history`.
It is neither a caller-authored SQL surface nor a fallback to the compatibility
`Connector.Read` path.

The source identity names one `database.schema.relation`. Before opening a
pool or running a page query, the adapter requires that stream's own non-null
cursor field and one distinct non-null primary/unique tie-breaker, then validates
both against the live typed catalog. A connection-level `cursor_field` setting
does not select a sync stream cursor. PostgreSQL accepts integer and timestamp
cursors; nullable, unknown, and unsupported cursor columns are refused so a
stored checkpoint can never silently become an unfiltered scan or omit NULL
rows.

The shared executor owns keyset paging, watermark checkpoints, durable
acknowledgement sequencing, and resume validation. Checkpoints are bound to the
source identity, source generation, live-catalog fingerprint, and no-snapshot
barrier. An invalid or stale checkpoint requires explicit rebootstrap instead
of restarting from page one. Hard deletes are not observable through polling.

`polling_watermark.json` remains a planned static declaration only because a
dynamic PostgreSQL catalog cannot truthfully declare a fixed cursor type,
cursor column, or tie-breaker. The production transport builds that effective
implemented declaration per stream after catalog validation and runs the
shared preflight/conformance contract before source I/O.

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

- `cursor_field` (optional, string); Legacy compatibility setting for direct incremental reads.
  Warehouse polling does not use it: every configured stream must provide and validate its own
  cursor field before source I/O.
- `cdc_publication` (optional, string); Publication used by logical-replication CDC. It must
  include the selected `schema.table` stream. Defaults to `pm_publication`.
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

Read behavior: low. No write actions are declared. The private managed-target driver does not
publish a capability until a production destination declaration and factory make it reachable.
PostgreSQL exposes no generic SQL write action and cannot write an arbitrary pre-existing target.

## Known limits

- Provider HTTP rate limiting is not applicable: PostgreSQL uses its native wire protocol and
  makes no provider HTTP API requests. Native pool, batch, statement, CDC stage, slot, and WAL
  bounds are enforced by the typed database and changefeed contracts instead.
- PostgreSQL warehouse polling is implemented through the shared polling source
  executor. Its source ordering is the live-catalog stream cursor plus a unique
  tie-breaker; its checkpoint is committed only after durable downstream
  acknowledgement, so replay is at least once. The static
  `polling_watermark.json` status stays planned because that file cannot
  represent a catalog-discovered cursor and tie-breaker. A polling source
  cannot observe a hard delete after the row disappears; a delete-aware history
  contract needs a declared cursor-advancing soft-delete mapping. Incompatible
  state, source identity changes, and schema changes require explicit
  rebootstrap, never an automatic full scan.
- Logical-replication CDC requires PostgreSQL 14+, `wal_level=logical`, a role with
  `REPLICATION`, positive `max_replication_slots` and `max_wal_senders`, a matching publication,
  and a stable primary-key replica identity. The slot is source-bound; a missing, incompatible,
  invalidated, or retention-gapped checkpoint requires explicit rebootstrap rather than a guessed
  resume.
- Each streamed transaction is constrained by private byte, record, aggregate-stage, and age
  limits. `StreamAbort` publishes, checkpoints, and acknowledges nothing. A successful
  `StreamCommit` emits its ordered transaction, records its complete durable receipt, persists the
  checkpoint, and only then acknowledges the commit LSN.
- Cursor or timestamp reconciliation is not a CDC fallback: it cannot faithfully recover hard
  deletes or transaction history. A stage-limit outcome must require explicit retry or connector-
  owned teardown/rebootstrap, with source slot health made visible.
- Query remains false. The PostgreSQL protocol is used only by the declared typed source and
  change-capture contracts; there is no caller-authored SQL query surface.
