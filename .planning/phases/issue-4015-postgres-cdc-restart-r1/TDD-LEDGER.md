# TDD Ledger — PostgreSQL CDC Restart Recovery

| ID | Red behaviour | Green implementation | Status |
| --- | --- | --- | --- |
| CDC-R1 | A fresh process rejects the checkpoint emitted by the interrupted PostgreSQL CDC path and the post-interruption row never reaches the target. | Producer, durable envelope, validator, and resume dispatch agree on the canonical logical-replication mechanism and position. | Red — live; `traces/live-red.txt`. |
| CDC-R2 | A validator-only relaxation could accept a checkpoint while resuming from the wrong LSN. | Focused checkpoint-family test requires the durable logical-replication envelope and still rejects a polling envelope; existing CDC tests assert restored LSN/source identity, and the live restart asserts the delivered transaction. | Red — focused; `traces/focused-red.txt`. |
| CDC-R3 | Restart can lose a row committed after interruption. | Fresh executor resumes from the last durable LSN and delivers the later row. | Pending red. |
| CDC-R4 | Restart can replay already-applied rows and duplicate target keys. | Independent target query asserts exact total count and post-interruption key multiplicity `1`; earlier count remains unchanged at the interruption boundary. | Pending live red. |
| CDC-R5 | A failed receiver/checkpoint commit could advance PostgreSQL acknowledgement beyond durable state. | Existing and new failure-injection tests retain receipt → checkpoint → acknowledgement ordering and resume from the last successful commit. | Pending verification. |
| CDC-R6 | Capability evidence can read stronger than the fixed implementation. | Inspect the existing artifact and update its result/limitations only from observed live proof. | Pending inspection. |

## Red:

The live binary harness reproduced `invalid_checkpoint: PostgreSQL polling checkpoint mechanism is not resumable` after process death. It had already proven the persisted bootstrap checkpoint contained `logical_replication`; the post-restart row never reached the independently queried target. The exact command and observable assertions are retained in `traces/live-red.txt`.

The focused checkpoint-family test fails for the same reason as live and is retained in `traces/focused-red.txt`. It accepts only the durable logical-replication family and explicitly requires a polling checkpoint to remain rejected on the bootstrap path.

## Green:

Pending. The same focused and live commands will be rerun after the smallest correct fix.

## Refactor:

Pending. Any cleanup must preserve the checkpoint wire contract, identity binding, and acknowledgement ordering.
