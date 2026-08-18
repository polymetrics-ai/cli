# TDD Ledger — PostgreSQL CDC Restart Recovery

| ID | Red behaviour | Green implementation | Status |
| --- | --- | --- | --- |
| CDC-R1 | A fresh process rejects the checkpoint emitted by the interrupted PostgreSQL CDC path and the post-interruption row never reaches the target. | Producer, durable envelope, validator, and resume dispatch agree on the canonical logical-replication mechanism and position. | Pending live red. |
| CDC-R2 | A validator-only relaxation could accept a checkpoint while resuming from the wrong LSN. | Restart test asserts the exact persisted/started LSN, source identity, and delivered transaction, not merely absence of an error. | Pending red. |
| CDC-R3 | Restart can lose a row committed after interruption. | Fresh executor resumes from the last durable LSN and delivers the later row. | Pending red. |
| CDC-R4 | Restart can replay already-applied rows and duplicate target keys. | Independent target query asserts exact total count and post-interruption key multiplicity `1`; earlier count remains unchanged at the interruption boundary. | Pending live red. |
| CDC-R5 | A failed receiver/checkpoint commit could advance PostgreSQL acknowledgement beyond durable state. | Existing and new failure-injection tests retain receipt → checkpoint → acknowledgement ordering and resume from the last successful commit. | Pending verification. |
| CDC-R6 | Capability evidence can read stronger than the fixed implementation. | Inspect the existing artifact and update its result/limitations only from observed live proof. | Pending inspection. |

## Red:

Pending. The live current-failure trace and the focused failing test will be captured before production edits.

## Green:

Pending. The same focused and live commands will be rerun after the smallest correct fix.

## Refactor:

Pending. Any cleanup must preserve the checkpoint wire contract, identity binding, and acknowledgement ordering.
