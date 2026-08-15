---
coverage:
  - id: D1
    description: Approval binds and consumes the complete sealed plan before session mutation.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go::TestDatabaseWriteExecutorBindsEveryApprovalPlanDimensionBeforeMutation
        status: pass
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go::TestDatabaseWriteExecutorConsumesApprovalBeforeOnePinnedBoundedSession
        status: pass
    human_judgment: false
  - id: D2
    description: One pinned session uses bounded batches, rolls back failures/cancellation, and cannot use legacy Connector.Write.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go::TestDatabaseWriteExecutorConsumesApprovalBeforeOnePinnedBoundedSession
        status: pass
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go::TestDatabaseWriteExecutorRollsBackFailuresAndCancellationWithoutCheckpointAuthority
        status: pass
    human_judgment: false
  - id: D3
    description: Commit uncertainty is terminal and only a confirmed, ledger-recorded receipt creates checkpoint authority.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go::TestDatabaseWriteExecutorTreatsUnknownCommitAsTerminalWithoutRetryOrRollback
        status: pass
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go::TestDatabaseWriteExecutorConsumesApprovalBeforeOnePinnedBoundedSession
        status: pass
    human_judgment: false
  - id: D4
    description: Atomic overwrite and canonical append/upsert/dedupe session strategies are enforced.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go::TestDatabaseWriteExecutorRequiresAtomicOverwriteAndPinsCanonicalStrategies
        status: pass
    human_judgment: false
---

# Summary — Issue 3973: transactional database write sessions

The shared `internal/connectors/database` layer now seals a write plan against
an asserted managed target control record, admitted mode, canonical strategy,
keys, exact count, batch limit, and overwrite effect. Native drivers provide a
preview then one pinned session; the executor consumes approval before begin,
batches only through that session, and records a confirmed target receipt in
the managed-target delivery ledger before making durable checkpoint authority
available.

Commit uncertainty is explicit and terminal. Cancellation and pre-commit
failures roll back the same session. `full_overwrite` requires native atomic
publish capability, while append/upsert/dedupe retain the existing closed
strategy vocabulary. `DriverRegistry.ResolveWriteDriver` adds exact registered
definition/driver admission without making a descriptor alone executable.

## Delivery evidence

- GSD lifecycle was run inline because the canonical single-worker contract
  forbids role spawning here: `discuss-phase` → `plan-phase --tdd` →
  `execute-phase` → `verify-work` → `code-review`.
- Red/green proof is retained in `traces/write-session-red.txt` and
  `traces/write-session-green.txt`.
- Required skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  and `golang-database`.
- There is no driver/native SQL/DDL/credential/live-target test in this
  foundation; real PostgreSQL proof and capability promotion remain #3982 and
  #3978. PostgreSQL’s `write=false`/unsupported-write test remains green.
