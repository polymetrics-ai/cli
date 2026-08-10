# Manual code review — PostgreSQL logical-replication CDC containment

Date: 2026-08-10
Mode: inline manual review (the issue-worker contract forbids reviewer-role
spawning for this lane).

## Reviewed scope

- The retained native replication material, source identity, slot lifecycle,
  `pgoutput` decoder, recovery classification, and the new UTF-8/replica-
  identity guards.
- The fail-closed `CDCReadRequest` boundary, bundle declaration, generated
  website connector data, documentation, and skipped historical live test.

## Findings and dispositions

1. **Historical PostgreSQL 12 compatibility:** modern `SNAPSHOT 'nothing'` syntax did
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
   or connection string. The historical live test accepts individual environment
   fields, but now skips before reading them because planned CDC must not contact
   a source.
7. **Large transaction safety:** the former protocol-v1 at-commit path has no
   bounded per-transaction stage or whole-transaction durable receipt. It is
   therefore unavailable rather than advertised. A future executor must use
   PostgreSQL 14+ v2 streaming, a hard quota, `StreamAbort` discard, and an
   explicit slot-health/rebootstrap procedure.
8. **Data fidelity and identity:** replication configuration forces/verifies
   UTF-8 and rejects malformed tuple/origin/type bytes before durable handling.
   The retained path admits only DEFAULT replica identity with a primary key and
   rejects FULL, USING INDEX, NOTHING, and unknown modes by named error.

## Verdict

Pass for containment. The executor is fail-closed: metadata and catalogue do
not advertise CDC, `ReadCDC` returns a named unsupported error before any
source operation, and the historical real PostgreSQL test skips. This is not
current end-to-end CDC conformance.
