---
coverage:
  - id: D1
    description: Closed registered polling target applies only descriptor-bounded pages.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_apply_test.go
        status: pass
      - kind: integration
        ref: TestPostgresManagedTargetWorksetDeliveryLive oversized-page re-read
        status: pass
    human_judgment: false
  - id: D2
    description: PostgreSQL applies all six native strategy states safely.
    verification:
      - kind: integration
        ref: databaseintegration TestPostgresManagedTargetWorksetDeliveryLive
        status: pass
    human_judgment: false
  - id: D3
    description: The target acknowledgement is available only after durable target/ledger evidence.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go
        status: pass
      - kind: integration
        ref: databaseintegration TestPostgresManagedTargetWorksetDeliveryLive
        status: pass
    human_judgment: false
---

# Summary — #3859 native database apply strategies

Implemented a sealed `DatabasePollingApplyExecutor` that translates a resolved
native polling descriptor and bounded page into a count-bound
`DatabaseWritePlan`, preview, one-shot approval, transaction, receipt, and
durable acknowledgement. It introduces no source reader or generic write
surface.

PostgreSQL now provisions private source-order fence storage and history
columns, supports `incremental_dedupe_history`, and applies keyed
upsert/dedupe/history events under a transaction and table lock. Explicit
tombstones physically delete non-history rows and close history rows; omitted
records remain ordinary absence.

Verification is recorded in `VERIFICATION.md`, `UAT.md`, and the trace files.
The inline/manual GSD fallback is used because the canonical issue phase is
non-numeric and the single-worker delivery contract forbids role spawning.
