# TDD ledger — Twenty CRM

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Recovery command contract | Pending: add Twenty-local contract test; it must fail before import because `cli_surface.json` is absent. | Pending: recovered and reconciled bundle makes the test pass with 168 fully-classified commands. | Pending: record exact command output. |
| Read and pagination | Pending: first failing focused test exposes stale planned/partial command and direct-read binding. | Pending: current operation-backed command declarations pass focused/conformance tests. | Pending. |
| Write/delete safety | Pending: failing test for typed confirmation, invalid record, and delete request shape. | Pending: current declaration passes conformance and refuses unsafe input. | Pending. |
| Live round trip | Pending: no fixture may stand in for live proof. | Pending: built binary proves bounded create/read/update/delete against disposable Twenty. | Pending: record only commands, counts, statuses, and redacted identifiers. |
