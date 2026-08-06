# Overview

Reads MySQL tables through the MySQL wire protocol. It discovers tables/columns dynamically,
performs bounded full and cursor-incremental reads, and consumes row-based binary-log changes for a
selected discovered table. It is a read-only source.

## Auth setup

Configure a bare `host`, `database`, and `username`. `port` defaults to 3306. `password` is a
secret field and is never logged. Do not put credentials in a host or URL-shaped value.

For CDC, configure a positive `replication_server_id` unique among replication clients. The MySQL
server must enable binary logging with row format and full row images; it must also grant the
configured account the narrowly appropriate replication-read privilege for the deployed server.

## Streams notes

The catalog exposes every base table in the configured database as `database.table`. Set
`cursor_field` to a stable sortable column to make reads cursor-incremental and internally paged;
`page_size` and `read_limit` bound those reads. CDC targets one discovered stream and stores the
last acknowledged binary-log file and position.

## Write actions & risks

This connector is read-only. It issues discovery and read queries and opens a replication stream;
it does not create, alter, or delete database data.

## Known limits

- CDC guarantees source ordering and at-least-once delivery. A replay at the committed binary-log
  boundary is possible; downstream consumers must use the file/position dedupe identity.
- CDC requires row-based binary logging and full row images. Statement-only or minimal row-image
  binary logs are rejected as unsuitable for complete row event projection.
- TLS configuration, GTID checkpointing, schema-change event projection, and cross-database CDC
  fan-out are outside this first engine slice.

## Container integration proof

The tagged integration test is intentionally sequential and opt-in. It starts one MySQL
`docker.io/library/mysql:8.4.11` container on a Docker-assigned loopback port, seeds deterministic
multi-page data, and asserts check, discovery, full read, incremental read, and row-binlog CDC
against returned records. Run the complete proof with its required host-disk cleanup:

```bash
DOCKER_CONTEXT=colima POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_DATABASE_RESET_COLIMA=1 \
  go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql
```

Every Docker command is scoped to the supplied context. The test refuses to run without the opt-in
or explicit context, and never falls back to Docker's global default. Its deferred cleanup removes
the generated container, named volume, and an image pulled by that run; it reports free disk before
startup and after teardown. The reset opt-in then runs `colima delete` followed by `colima start`
for the disposable default profile, because Docker removal frees space inside Colima but does not
shrink its host VM disk file. It is destructive to that profile and must not be used while it holds
other work. `POLYMETRICS_DATABASE_KEEP_IMAGE=1` is an explicit local speed opt-in and cannot be
combined with the reset.

Podman is not used for this harness: three task-era Podman machines on this host failed to start,
whereas Docker via Colima was independently verified with MySQL 8.4 and row-format binary logging.
