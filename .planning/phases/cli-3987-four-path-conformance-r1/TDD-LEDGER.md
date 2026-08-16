# TDD ledger — #3987 four-path warehouse conformance

| Slice | Red | Green | Refactor | Status |
| --- | --- | --- | --- | --- |
| Exact four-direction contracts | Pending: add focused matrix test before its helper/invariant exists. | Pending. | Pending. | planned |
| Warehouse mediation and current mode matrix | Pending: assertions must distinguish a source-owned stage from a destination workset and reject stale `incremental_dedupe_history` expectations. | Pending. | Pending. | planned |
| Change-capture source-only restriction | Pending: PostgreSQL target-mode interpretation must produce a typed pre-I/O refusal. | Pending. | Pending. | planned |
| Direction-specific failure demonstration | Pending: schema-valid API→API source-binding scratch defect. | Pending restoration and focused green rerun. | N/A. | planned |

The implementation records only observed red/green output. No skipped database or provider check is counted as a pass.
