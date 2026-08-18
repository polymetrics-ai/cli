# TDD Ledger — PostgreSQL CDC Restart Recovery

| ID | Red behaviour | Green implementation | Status |
| --- | --- | --- | --- |
| CDC-R1 | A fresh process rejects the checkpoint emitted by the interrupted PostgreSQL CDC path and the post-interruption row never reaches the target. | The transport wrapper identifies bootstrap/CDC requests before applying checkpoint-family validation and dispatches the durable logical-replication envelope to the sealed CDC resume path. | Green — focused and live. |
| CDC-R2 | A validator-only relaxation could accept a checkpoint while resuming from the wrong LSN. | Focused checkpoint-family test requires the durable logical-replication envelope and still rejects a polling envelope; existing CDC tests assert restored LSN/source identity, and the live restart asserts the delivered transaction. | Green — focused and live. |
| CDC-R3 | Restart can lose a row committed after interruption. | Fresh executor resumes from the last durable LSN and delivers the later row. | Green — target `1 → 1 → 2`. |
| CDC-R4 | Restart can replay already-applied rows and duplicate target keys. | Independent target query asserts exact total count and post-interruption key multiplicity `1`; earlier count remains unchanged at the interruption boundary. | Green — resumed key count `1`; control `1001`. |
| CDC-R5 | A failed receiver/checkpoint commit could advance PostgreSQL acknowledgement beyond durable state. | Existing failure-injection/package tests retain receipt → checkpoint → acknowledgement ordering and resume from the last successful commit. | Green — package and capability live proof. |
| CDC-R6 | Capability evidence can read stronger than the fixed implementation. | Audit the immutable record's explicit facts against a fresh binary run; document that PostgreSQL CDC remains at-least-once and that exact final multiplicity is route/target proof, not a generic claim. | Green — artifact unchanged with written disposition. |

## Red:

The live binary harness reproduced `invalid_checkpoint: PostgreSQL polling checkpoint mechanism is not resumable` after process death. It had already proven the persisted bootstrap checkpoint contained `logical_replication`; the post-restart row never reached the independently queried target. The exact command and observable assertions are retained in `traces/live-red.txt`.

The focused checkpoint-family test fails for the same reason as live and is retained in `traces/focused-red.txt`. It accepts only the durable logical-replication family and explicitly requires a polling checkpoint to remain rejected on the bootstrap path.

## Green:

The focused checkpoint-family regression and full PostgreSQL package pass. The same live pipeline now reports CDC target row counts `1` before interruption, `1` at interruption, and `2` after restart, with the new key exactly once and the 1,001-row control table unchanged. `traces/focused-green.txt` and `traces/live-green.txt` retain the commands and observations.

## Refactor:

The shared checkpoint protocol check was extracted from `validateCDCResume` and reused by the transport wrapper. This keeps one logical-replication mechanism/protocol rule and preserves the later live source/generation/publication/schema validation. No checkpoint wire field changed.
