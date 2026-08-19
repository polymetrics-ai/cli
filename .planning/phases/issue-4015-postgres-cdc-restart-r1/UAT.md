# UAT — PostgreSQL CDC restart recovery for 0.2.1

## Automated acceptance evidence

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Current failure is reproduced before repair | `traces/live-red.txt` and `traces/focused-red.txt` | Pass: fresh process rejects the durable logical-replication envelope as polling and the post-restart row is absent. |
| Checkpoint-family validation remains strict | `TestPostgresBootstrapTransportAcceptsOnlyDurableLogicalReplicationCheckpointBeforeIO` | Pass: committed logical-replication checkpoint accepted only for bootstrap/CDC; valid polling checkpoint rejected with typed rebootstrap. |
| Restart resumes from the durable LSN | `TestPMBinaryExecutesPostgresWarehousePostgres` | Pass: killed process restarts from the persisted logical-replication checkpoint and advances after the later committed transaction. |
| No loss or duplicate in the managed target | Independent PostgreSQL target queries in the same live test | Pass: CDC counts `1 → 1 → 2`, resumed key multiplicity `1`, control count `1,001`. |
| CDC durability ordering remains true | `TestPMBinaryDispatchesPostgresChangeCaptureToWarehouse` and PostgreSQL package tests | Pass: bounded stage receipt and checkpoint precede source acknowledgement. |
| Capability evidence remains honest | Audit of `postgres_cdc_r1-capability-cdc.json` against a fresh live run | Pass: its explicit facts remain true; generic delivery remains `at_least_once`. |

No human judgment is needed for behavioral acceptance. Merge approval remains the repository's human gate.
