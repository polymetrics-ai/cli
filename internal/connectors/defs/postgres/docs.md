# Overview

Reads PostgreSQL tables from a dynamically discovered catalog, snapshots tables, and supports
cursor-incremental reads on a configurable cursor column. PostgreSQL logical-replication change
capture uses PostgreSQL 14+ `pgoutput` protocol-v2 streaming. It keeps each transaction in a
bounded crash-recoverable private stage, discards `StreamAbort`, delivers only after
`StreamCommit`, and persists a whole-transaction durable receipt before checkpointing and
acknowledging the source LSN.

Stream availability is discovered at runtime; see [Streams notes](#streams-notes) for the configured
schema's scope and permission rules.

The write capability is the typed PostgreSQL managed-target driver. It applies a durable warehouse
workset only to a target whose private ownership and schema control records match the exact
connection. It does not expose a generic SQL action, a direct connector-to-connector hop, or the
compatibility `Connector.Write` method.

## Warehouse snapshot transport

For warehouse-mediated sync, PostgreSQL provides one closed, live-only bounded
snapshot source. It accepts the logical `snapshot` stream only for
`full_append` and `full_overwrite`; it is neither a caller-authored SQL surface
nor a fallback to the compatibility `Connector.Read` path. Polling, incremental modes, and
logical-replication change capture remain distinct source paths.

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

Read behavior: low. Write behavior: high and restricted to the typed managed-target path. The
driver validates the warehouse workset, target owner, namespace, relation, type mapping, mode,
and delivery ledger before applying a bounded batch. A durable target receipt is required before
the warehouse delivery checkpoint advances. PostgreSQL exposes no generic SQL write action and
cannot write an arbitrary pre-existing target.

## Known limits

- Provider HTTP rate limiting is not applicable: PostgreSQL uses its native wire protocol and
  makes no provider HTTP API requests. Native pool, batch, statement, CDC stage, slot, and WAL
  bounds are enforced by the typed database and changefeed contracts instead.
- `polling_watermark` is planned, not implemented. It is a bounded keyset poll
  rather than CDC or change capture, and no polling mode can be selected until
  one declared native source/object/destination binding passes runtime
  preflight with registered source and apply executors plus immutable
  conformance evidence. When such a binding exists, its source order must be a
  declared watermark plus unique tie-breaker; its checkpoint is committed only
  after durable downstream acknowledgement, so replay is at least once. A
  polling source cannot observe a hard delete after the row disappears; a
  delete-aware history contract needs a declared cursor-advancing soft-delete
  mapping. Incompatible state, source identity changes, snapshot expiry, and
  retention failure require explicit rebootstrap, never an automatic full scan.
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
- Query remains false. The PostgreSQL protocol is used only by the declared typed source,
  change-capture, and managed-target contracts; there is no caller-authored SQL query surface.
