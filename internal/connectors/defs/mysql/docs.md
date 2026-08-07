# Overview

Reads MySQL tables through the MySQL wire protocol. It discovers tables/columns dynamically,
performs bounded full and cursor-incremental reads, and consumes row-based binary-log changes for a
selected discovered table. It is a read-only source registered through the native MySQL factory.

## Auth setup

Configure a bare `host`, `database`, and `username`. `port` defaults to 3306. `password` is a
secret field and is never logged. Do not put credentials in a host or URL-shaped value.

`sslmode` chooses transport security and is honestly enforced for both local and remote servers:

| `sslmode` | Encrypts | Falls back to plaintext | Verifies chain | Verifies server name |
| --- | --- | --- | --- | --- |
| `disabled` | no | n/a | no | no |
| `preferred` (default) | when offered | yes, only when the server advertises no TLS | no | no |
| `required` | yes | never | no | no |
| `verify-ca` | yes | never | yes | no |
| `verify-identity` | yes | never | yes | yes |

`preferred` is the only mode that may fall back. `required` and both verifying modes fail closed
when the server offers no TLS. `sslrootcert` names an absolute PEM CA path for the verifying modes
and defaults to the host trust store; `sslservername` overrides the verified name under
`verify-identity`, for example when connecting to an IP address. libpq spellings
(`disable`/`prefer`/`require`/`verify-full`) are accepted so a value means the same thing here as on
the PostgreSQL connector. The mode applies to discovery, reads, and the binary-log replication
stream alike.

For CDC, configure a positive `replication_server_id` unique among replication clients. The MySQL
server must enable binary logging with row format and full row images; it must also grant the
configured account the narrowly appropriate replication-read privilege for the deployed server.

## Streams notes

The catalog exposes every base table in the configured database as `database.table`. Complete
read paging requires a single-column primary key. Without `cursor_field`, a full snapshot is
unfiltered and pages by that primary key. Set `cursor_field` only to a non-null single-column
primary or unique key for incremental reads; pages order by `(cursor_field, primary_key)` and
resume strictly after a present stored cursor, including an empty or whitespace text value.
`page_size` and `read_limit` bound every read. Textual and temporal wire values are emitted as
strings, while binary values are copied before emission. CDC targets one discovered stream and stores
the last acknowledged binary-log file, position, and ordered-column schema fingerprint. A saved CDC
position without that fingerprint, or one whose fingerprint no longer matches live metadata, requires
a resnapshot rather than projecting rows against a changed schema.

## Write actions & risks

This connector is read-only. It issues discovery and read queries and opens a replication stream;
it does not create, alter, or delete database data.

## Known limits

- CDC guarantees source ordering and at-least-once delivery. A replay at the committed binary-log
  boundary is possible; downstream consumers must use the file/position/row-ordinal dedupe
  identity.
- CDC requires row-based binary logging and full row images. Statement events, including DDL and
  runtime binlog format or row-image changes, are rejected before a later row or checkpoint is
  emitted. New CDC subscriptions capture their binary-log position before loading column metadata.
- Tables and columns whose identifiers this connector cannot safely quote are omitted from the
  catalog rather than advertised, because a Read against them would always fail.
- GTID checkpointing, client certificate authentication, schema-change event projection, and
  cross-database CDC fan-out are outside this first engine slice.

## Container integration proof

The tagged integration test is opt-in and sequential. It starts one
`docker.io/library/mysql:8.4.11` container in Podman on a dynamically assigned loopback port,
configures row-based binary logging, seeds deterministic multi-page data, then asserts check,
schema discovery, a full read, an incremental read, every `sslmode`, and row binary-log CDC against
the records actually returned.

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_PODMAN_CONNECTION=<your-machine> \
  POLYMETRICS_DATABASE_RECLAIM_DISK=1 \
  go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql
```

`POLYMETRICS_PODMAN_CONNECTION` is mandatory and has no default. Every Podman command is scoped to
it with `--connection`, because a bare `podman` command targets whichever machine is the global
default, which on a shared host belongs to someone else. The test skips with a visible reason when
the opt-in or the connection is missing, and fails rather than passing when the engine is
unreachable.

Cleanup runs on every exit path, including a failed assertion and an interrupt: the generated
container, its named volume, the run-owned image tag, and the pulled source image are all removed.
`POLYMETRICS_DATABASE_KEEP_IMAGE=1` retains the source image for a subsequent run.

`POLYMETRICS_DATABASE_RECLAIM_DISK=1` additionally trims the backing machine so freed guest blocks
are punched out of the host's sparse disk file. This matters on macOS: removing an image frees space
inside the VM but leaves the host file inflated. The trim runs **twice**, because a single pass
immediately after an image removal was measured to return only about a quarter of the space, while a
second pass returns effectively all of it. Measured on Podman 5.3.2 with applehv, one full run moved
host free space by under 2 MiB end to end.

### Adding a second engine

Supply another `dbtest.Config` — image, container port, data-volume path, container and engine
arguments. No new code in `internal/connectors/native/dbtest` is required. Engines run one at a
time by default; `dbtest.SetMaxConcurrentEngines` is the bounded opt-in for parallel runs, and
parallel database containers multiply peak disk and memory.
