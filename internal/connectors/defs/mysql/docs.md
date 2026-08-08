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
`verify-identity`, for example when connecting to an IP address. Compatibility spellings
(`disable`/`allow`/`prefer`/`require`/`verify-full` and `verify_ca`/`verify_identity`) are accepted;
use the canonical modes in the table for a portable policy across SQL connectors. The mode applies
to discovery, reads, and the binary-log replication stream alike.

For CDC, configure a positive `replication_server_id` unique among replication clients. The MySQL
server must be **8.4 or newer**, enable binary logging with row format and full row images, and
grant the configured account the narrowly appropriate replication-read privilege for the deployed
server. Reads, discovery, and checks carry no such version bound.

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
- CDC requires MySQL 8.4 or newer. A new subscription reads its starting position with
  `SHOW BINARY LOG STATUS`, which replaced `SHOW MASTER STATUS` in 8.4; against 8.0.x or 5.7 that
  first read fails rather than silently capturing a wrong position. MySQL 8.4.11 is the verified
  server; pre-8.4 support is tracked separately and is not part of this connector today.
- CDC requires row-based binary logging and full row images. A runtime binlog format or row-image
  change is rejected wherever it happens, because it silences row events for every schema at once.
  Other statement events, including DDL, are rejected only when they can reach the configured
  database: the binary log is server-wide, so unrelated schema activity does not end a changefeed,
  while a statement that names the configured database — including one qualified from another
  default schema, in backquotes or in `sql_mode=ANSI_QUOTES` double quotes — does. A statement that
  cannot be attributed to a schema is rejected too. New CDC subscriptions capture their binary-log
  position before loading column metadata.
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
  POLYMETRICS_DATABASE_OWN_MACHINE=1 \
  go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql
```

`POLYMETRICS_DATABASE_OWN_MACHINE=1` makes the run create its own uniquely named Podman machine,
use it, and delete it again. This is the mode that proves host-disk reclaim, because only a machine
this process created is trimmable. Pass `POLYMETRICS_PODMAN_CONNECTION=<your-machine>` instead to
run against an existing machine; one of the two is mandatory, and there is no default, because a
bare `podman` command targets whichever machine is the global default, which on a shared host
belongs to someone else. Every Podman command is scoped with `--connection`. The test skips with a
visible reason when the opt-in or the connection is missing, and fails rather than passing when the
engine is unreachable.

Cleanup runs on every exit path, including a failed assertion and an interrupt: the generated
container, its named volume, the run-owned image tag, the pulled source image, and any machine this
run created are all removed. `POLYMETRICS_DATABASE_KEEP_IMAGE=1` retains the source image for a
subsequent run.

Cleanup also trims the backing machine so freed guest blocks are punched out of the host's sparse
disk file. This matters on macOS: removing an image frees space inside the VM but leaves the host
file inflated. The trim runs **twice**, because a single pass immediately after an image removal was
measured to return only about a quarter of the space, while a second pass returns effectively all of
it. Measured on Podman 5.3.2 with applehv, one full run moved host free space by under 2 MiB end to
end.

The trim runs only against a machine this process created through `dbtest.NewMachine` and still
holds an ownership record for. A matching name is not ownership: `fstrim -av` reaches every
filesystem on a machine, so a caller-supplied, pre-existing, shared, or remote machine is never
trimmed no matter what it is called. Those runs report the reason together with the host bytes still
reclaimable instead. Two weaker checks follow the ownership record as defence in depth — the scoped
connection must address that machine (`<machine>` or `<machine>-root`), and `podman machine inspect`
must still resolve it locally. `POLYMETRICS_PODMAN_MACHINE` names the machine behind a
caller-supplied connection whose name differs; it does not confer ownership.

### Adding a second engine

Supply another `dbtest.Config` — image, container port, data-volume path, container and engine
arguments. No new code in `internal/connectors/native/dbtest` is required. Engines run one at a
time by default; `dbtest.SetMaxConcurrentEngines` is the bounded opt-in for parallel runs, and
parallel database containers multiply peak disk and memory.
