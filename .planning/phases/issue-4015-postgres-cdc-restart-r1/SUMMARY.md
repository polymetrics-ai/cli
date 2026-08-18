# Summary — PostgreSQL CDC restart recovery for 0.2.1

The diagnosed symptom was correct, but the proposed cause was stale. PostgreSQL already implements logical replication and persists a `logical_replication` checkpoint; `changefeed.json` is `implemented` on the target base. The actual defect was the polling transport wrapper applying polling-only checkpoint validation before its existing `transport_bootstrap` branch dispatched to the logical-replication CDC path.

The wrapper now classifies bootstrap/CDC requests first and validates their existing logical-replication checkpoint protocol plus sealed native identity. Ordinary polling requests retain their original strict validation, and polling checkpoints are explicitly rejected on the CDC path. The live CDC reader still validates source/system/timeline/publication/relation/schema identity and retained LSN before resuming.

Red reproduced the process-death failure and absent post-restart row. Green killed the sync after a durable position, restarted it, inserted a later row, and independently queried the target: CDC rows were `1` before interruption, `1` at interruption, and `2` after restart; the later key appeared exactly once and the 1,001-row control table was unchanged.

The PostgreSQL CDC capability artifact remains unchanged because its explicit bounded-stage, warehouse-receipt, readback, and source-acknowledgement facts still pass. Delivery remains honestly declared `at_least_once`; the keyed managed-target result is not presented as a generic exactly-once guarantee.
