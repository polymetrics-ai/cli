# Deep code review — PostgreSQL apply/history club

Status: passed with no Critical, Warning, or Info findings.

Review mode: `/gsd-code-review 4094 --depth=deep`, executed inline because the
repository's single-worker delivery contract forbids a spawned reviewer role.

## Scope reviewed

- `internal/connectors/engine/database_polling_apply.go`
- `internal/connectors/native/postgres/managed_target_integration_test.go`
- Planning and evidence artifacts in this phase directory

## Cross-file findings

None.

The reviewed call path is `PollingPreflight` → registered
`DatabasePollingApplyExecutor` → `NewDatabaseWritePlan` → preview → approval →
execute. History route construction occurs before preview and session I/O. The
source definition comes only from the exact source object resolved by
preflight, its driver ID must match the cloned declaration, and the complete
driver pair is still validated by `DatabaseWriteHistoryRoute.validateFor`.
Consequently a non-PostgreSQL route keeps its typed reason, while a missing or
invalid provider fails closed before any target side effect.

Non-history modes do not enter the new branch. No raw relation, SQL, endpoint,
credential, dependency, public command, or reverse-ETL bypass was introduced.
Context checks, bounded batches, clone-before-use behavior, preview/approval,
receipt-gated acknowledgement, and the existing driver/session transaction
boundary remain in place. The test exercises the public registered-adapter
path and observes PostgreSQL rows plus durable ledger identity; it cannot pass
if the adapter does no work.

`git diff --check`, focused race tests, vet, build, live PostgreSQL tests, and
the applicable repository gates were clean. The full-suite parallel container
transient and clean isolated/package reruns are disclosed in `VERIFICATION.md`.
