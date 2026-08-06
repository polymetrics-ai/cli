# Manual code review — PostgreSQL logical-replication CDC

Date: 2026-08-06
Mode: inline manual review (the issue-worker contract forbids reviewer-role
spawning for this lane).

## Reviewed scope

- Native replication connection, source identity, slot lifecycle, `pgoutput`
  decode, checkpointing, acknowledgement ordering, and recovery classification.
- The durable `CDCReadRequest` checkpoint port and unchanged fail-closed
  capability projection.
- Bundle declaration, generated website connector data, documentation, and
  live integration test.

## Findings and dispositions

1. **PostgreSQL 12 compatibility:** modern `SNAPSHOT 'nothing'` syntax did
   not create a slot on the approved local PostgreSQL 12 server. Replaced it
   with PostgreSQL-12-compatible `NOEXPORT_SNAPSHOT`; the real protocol test
   passes on PostgreSQL 12.22.
2. **Resource close errors:** lint required the normal slot-inspection and
   integration-test connection closes to explicitly discard their errors.
   Updated both deferred closes; `make lint` passes.
3. **Secret safety:** no test fixture or planning artifact contains a password
   or connection string. The live test accepts individual environment fields,
   fails on an unreachable configured source, and only explicitly skips when
   the integration environment is intentionally absent.

## Verdict

Pass. The executor remains fail-closed: metadata is discoverable as CDC only
because the native connector provides the descriptor that exactly matches the
implemented bundle; invocation requires a real source and a durable checkpoint
committer.
