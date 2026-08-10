# Manual code review — PostgreSQL logical-replication CDC containment

Date: 2026-08-06
Mode: inline manual review (the issue-worker contract forbids reviewer-role
spawning for this lane).

## Historical reviewed scope

- Native replication connection, source identity, slot lifecycle, `pgoutput`
  decode, checkpointing, acknowledgement ordering, and recovery classification.
- The durable `CDCReadRequest` checkpoint port and unchanged fail-closed
  capability projection.
- Bundle declaration, generated website connector data, documentation, and
  live integration test.

## Historical findings and dispositions

1. **PostgreSQL 12 compatibility:** modern `SNAPSHOT 'nothing'` syntax did
   not create a slot on the approved local PostgreSQL 12 server. Replaced it
   with PostgreSQL-12-compatible `NOEXPORT_SNAPSHOT`; the historical real
   protocol test passed on PostgreSQL 12.22.
2. **Resource close errors:** lint required the normal slot-inspection and
   integration-test connection closes to explicitly discard their errors.
   Updated both deferred closes; `make lint` passes.
3. **Secret safety:** no test fixture or planning artifact contains a password
   or connection string. The historical live test accepted individual
   environment fields and failed on an unreachable configured source; it is
   now skipped while change capture remains planned.

## Current containment verdict

Pass. Change capture is planned and non-executable: `ReadCDC` fails closed
before connecting to a source, the connector does not advertise CDC, and the
preserved PostgreSQL 12.22 integration result is historical evidence only. No
current source LSN, checkpoint, or replication-slot lifecycle claim is made.
