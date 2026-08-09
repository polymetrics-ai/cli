# Manual code review — PostgreSQL logical-replication CDC

Date: 2026-08-10
Mode: inline manual review (the issue-worker contract forbids reviewer-role
spawning for this lane).

## Reviewed scope

- Native replication connection, source identity, slot lifecycle, `pgoutput`
  decode, checkpointing, acknowledgement ordering, recovery classification,
  and TLS parity for ordinary and replication connections.
- The durable `CDCReadRequest` checkpoint port and unchanged fail-closed
  capability projection.
- Bundle declaration, generated website connector data, documentation, and
  live integration test.

## Findings and dispositions

1. **PostgreSQL 12 compatibility:** modern `SNAPSHOT 'nothing'` syntax did
   not create a slot on the approved local PostgreSQL 12 server. Replaced it
   with PostgreSQL-12-compatible `NOEXPORT_SNAPSHOT`; the final real protocol
   test passes on PostgreSQL 12.22.
2. **Recovery lower bound:** `confirmed_flush_lsn` is an acknowledgement high
   water mark, not a guaranteed replay boundary. Recovery now compares the
   durable checkpoint only with `restart_lsn`, so it cannot skip an
   uncheckpointed transaction.
3. **Unsafe reuse and publication scope:** an existing derived slot with no
   durable checkpoint could otherwise omit prior WAL, and a multi-table
   publication could emit unrelated records. Existing uncheckpointed slots
   now return a typed rebootstrap requirement; publication membership is
   checked before slot creation and the decoder filters every relation to the
   source-bound table. Selected-table truncate is explicitly represented.
4. **Resource close errors:** lint required the normal slot-inspection and
   integration-test connection closes to explicitly discard their errors.
   Updated both deferred closes; `make lint` passes.
5. **Post-rebase CLI behavior:** inspected main's #3964 golden transcript
   update rather than copying it into this lane. No PostgreSQL help fixture
   changed; the actual focused `internal/cli` suite passes and the unresolved
   connector `--help --json` path returns usage exit code 2.
6. **Secret safety:** no test fixture or planning artifact contains a password
   or connection string. The live test accepts individual environment fields,
   fails on an unreachable configured source, and only explicitly skips when
   the integration environment is intentionally absent.

## Verdict

Pass. The executor remains fail-closed: metadata is discoverable as CDC only
because the native connector provides the descriptor that exactly matches the
implemented bundle; invocation requires a real source and a durable checkpoint
committer. The real PostgreSQL conformance test proves lifecycle cleanup and
restart behavior, not a mocked replication protocol.
