# TDD ledger — Issue 4095 residual route

| ID | Guarantee | Red | Green | Status |
| --- | --- | --- | --- | --- |
| R4-live | A real pgoutput insert/update/delete reaches a live PostgreSQL keyed apply/history target through durable stage, receipt, workset, and mapping. | Add the tagged whole-route test and show that the pre-existing component-only proofs cannot satisfy its real-stream/read-back assertion. | Independently query the target and durable state; assert receipt-before-LSN acknowledgement, restart restoration, and replay safety. | planned |
| Refusal-R1 | The R1 source/destination `change_capture` pairing has a typed pre-I/O refusal. | Add a named table row with source/target I/O counters and the expected concrete error. | `errors.As` identifies the declared typed error while both counters remain zero. | planned |
| Refusal-R2 | The R2 source/destination `change_capture` pairing has a typed pre-I/O refusal. | Add a named table row with source/target I/O counters and the expected concrete error. | `errors.As` identifies the declared typed error while both counters remain zero. | planned |
| Refusal-destinations | Every destination-side `change_capture` cell receives its own row. | Add rows generated from inspected GitHub and PostgreSQL transport declarations, never a catch-all. | Every row asserts its typed error and zero I/O. | planned |
| R3 | PostgreSQL CDC → API is not executable with current declarations. | N/A — do not create a destination action merely to test it. | Record `non_executable`: GitHub's destination declares neither a CDC source binding nor delete availability; `change_capture` is source-only to its connection-owned local warehouse. | non_executable |

## Red/green record

- **Red:** not yet run. The test is deliberately added before any production behavior change; its first result will be recorded verbatim with command and date.
- **Green:** not yet run. A green live route must include a real PostgreSQL 14+ run; an environment skip remains `not_run`, never `pass`.
