# Plan — MySQL container harness R1

## Scope and ownership

Target connector: **MySQL only**. Changed production paths are limited to the MySQL native
connector/bundle and the generic closed changefeed mechanism vocabulary required to represent its
binlog protocol honestly. The reusable harness is test-only plumbing under the native connector
test tree. It does not alter the production registry, command runner, sync contract, runtime
machine scripts, or another connector.

## Design

1. Add a small `dbtest` test-support package that runs a configured engine through an explicit
   Podman connection named explicitly by the caller. It creates collision-resistant container/volume names,
   records free disk before/after, chooses a Docker-assigned loopback host port, waits through a bounded probe, and
   tracks whether the image existed before the run. Cleanup is idempotent and runs via defer plus
   an interrupt handler: container removal, named-volume removal, then removal of only an image
   pulled by this run unless the keep-image opt-in is set. An opt-in host-disk trim of the backing machine runs only after
   Docker cleanup to reclaim host disk that Docker removal alone cannot shrink. Engines execute
   serially by default.
2. Add the MySQL Tier-3 bundle and component split: connection validation and wire-client setup;
   dynamic information-schema cataloging; bounded/safe snapshot and cursor-incremental reads; and
   a binary-log CDC executor. SQL identifiers stay strictly validated and values remain parameters.
   CDC persists binlog filename/position only after event callback acknowledgement and checkpoint
   commit.
3. Add `binlog_replication` to the shared closed declaration vocabulary and JSON Schema, then make
   `defs/mysql/changefeed.json` implemented only because `mysql.Connector` supplies the exact
   matching `ChangefeedExecutorDescriptor` and real `ReadCDC` behavior.
4. Add deterministic unit tests for config/identifier validation, catalog/read/CDC row projection,
   capability admission, bundle validity, and cleanup command ordering. Then add exactly one
   tagged MySQL end-to-end test driven through the harness. It asserts real returned records for
   check, discovery, full paging, incremental filtering, and insert/update/delete change capture.
5. Add a development guide with the single documented invocation and its cleanup/skip behavior.
   It records the next ready configurations: MariaDB (same binlog engine, version-specific
   position behavior); SQL Server (CDC/transaction-log dependency and image licensing review);
   Oracle (XE image/licensing/architecture availability); and PostgreSQL (existing connector needs
   its approved replication dependency and a logical-replication setup).

## Dependency legitimacy gate

The only new module considered is `github.com/go-mysql-org/go-mysql v1.16.0`. Its primary project
states it is a pure-Go MySQL/MariaDB network-protocol and replication library, supports binlog
streaming, and provides a simple client. pkg.go.dev lists the tagged stable MIT release published
2026-07-15. Before committing the dependency, capture its resolved transitive graph, run the Go
vulnerability scanner, and compare a clean `pm` binary built before vs after. No second module may
be added: its client package is used for normal database work as well as its replication package.

## TDD execution sequence

1. Write red unit tests for the harness resource lifecycle and missing-unscoped-Docker-context rejection;
   MySQL config/read/catalog/CDC contracts; and binlog mechanism validation.
2. Implement the test harness and MySQL connector until the focused tests pass.
3. Add the tagged real-engine test. It must fail when the engine cannot be reached, while deferred
   cleanup executes, and may only skip with a visible opt-in/scoped-connection reason before a
   live run begins.
4. Run the documented command against the explicit Podman connection. Check free-disk values,
   Podman resource absence, and the opt-in host-disk trim after the test prove no named container,
   volume, or harness-pulled image remains and the VM disk is reclaimed.
5. Run focused Go, format, vet/build, all non-suite verification gates, generated-surface checks,
   and inline review. Record results in `VERIFICATION.md` and exact PR material in `PR-BODY.md`.

## Acceptance facts

- No exposed default database port; the discovered non-default port reaches the connector only.
- Seed data has more rows than one request page and values that make incremental filtering exact.
- CDC assertion covers real row events and acknowledgement-gated source state.
- Failure and interrupt teardown use the same idempotent cleanup path.
- The default test suite stays dependency-free with a visible `SKIP` explanation for this optional
  integration test.
