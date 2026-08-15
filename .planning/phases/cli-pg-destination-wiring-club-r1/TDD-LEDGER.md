# TDD ledger — PostgreSQL production transport wiring club

| ID | Criterion / edge | Red evidence to capture | Green observable evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | #3982 destination registered | App-open preflight cannot resolve PostgreSQL as destination. | Exact bundle reference builds/registers the native destination; public write remains false. | planned |
| R2 | #3983 warehouse workset route | Production transport never constructs `ChangeDeliveryWorkset`/executor. | Stage-owned Parquet is rehashed, planned, approved, applied, receipt/baseline persisted, read back, then checkpointed. | planned |
| R3 | #3979 bootstrap bridge | App incremental route never invokes `BootstrapCDC`. | Before/during/after rows and tombstones appear via WAL/Parquet manifests and target receipts; LSN checkpoint advances after ack. | planned |
| R4 | Missing/stale/replayed approval | Forged or reused transport approval can reach destination preparation. | Typed approval refusal; zero target rows, zero control/ledger rows, unchanged baseline/checkpoint. | planned |
| R5 | Cancellation mid-operation | Cancellation can strand or advance undurable state. | Pre-commit cancel rolls back/does not checkpoint; post-receipt cancel preserves receipt/baseline and resumes exactly. | planned |
| R6 | Connection/process death | Source or target death can yield silent success or blind retry. | Typed unavailable/unknown outcome; no unproven checkpoint/baseline; identical immutable workset required on recovery. | planned |
| R7 | Empty/single/large | Zero rows may yield no checkpoint; page boundary may drop/duplicate rows. | Empty emits durable checkpoint with zero target effects, single emits one row, large spans bounded pages with exact row count. | planned |
| R8 | Duplicate/out-of-order | Replayed or late items may double-write or regress state. | Merge/ledger and source ordering keep exact target rows; acknowledged replay is no-op/reconciled and checkpoint never regresses. | planned |
| R9 | Schema drift | Changed source or replaced target may auto-evolve/adopt. | Typed schema/OID/owner refusal before mutation; target/control/baseline/checkpoint snapshots unchanged. | planned |
| R10 | Auth/permission refusal | Bad credentials or unreadable control state may partially provision. | Typed/safe error with no secret text; zero data/control/ledger and no checkpoint. | planned |
| R11 | Concurrent same target | Two runs may race approval, provision, baseline, or checkpoint. | One-shot approval/state CAS/target locks serialize or refuse; exact rows and one coherent baseline/checkpoint. | planned |
| R12 | Resume interruption | Durable pages may be reread or skipped after restart. | Fresh App/binary resumes native LSN/barrier, reuses target receipt semantics, and reaches exact final rows. | planned |
| R13 | Binary and command surface | Component-only tests repeat the audited defect. | Built `pm` plan/preview/approval/run call chain reaches real source, warehouse, target; command help/docs/goldens match. | planned |

## Red / Green protocol

Red output is captured before each production slice exists and must fail for the missing behavior, not for a broken fixture. Green output records the focused command and observable assertions. Live refusal cases snapshot target rows, managed control/ledger, baseline, and source checkpoint before and after.
