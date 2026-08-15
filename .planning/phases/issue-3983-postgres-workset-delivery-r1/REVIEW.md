# Code review — Issue 3983

Manual inline review was required because the canonical single-worker and
Firstmate direct-PR contracts prohibit role spawning. The generated
`code-review` prompt was resolved before this review.

| Area | Result | Evidence |
| --- | --- | --- |
| Approval before mutation | Pass | The controller validates/reopens the workset and consumes the exact workset-bound approval before `DatabaseWriteExecutor` can open a session; the stale-approval test observes all session/ledger/baseline mutation counters at zero. |
| Effects | Pass | The input is made only from `ReadDelta` and explicit `Tombstones`; tombstone key fields are mapped through `MappingContractV1`. No projection-difference or full-overwrite path exists. |
| Receipt and recovery | Pass | Baseline promotion follows a ledger-backed `DeliveryReceiptV1`; failed and unknown outcomes do not call the baseline store, and unknown commit maps to replay-required. |
| Target binding | Pass | The sealed plan hash covers workset identity/content, owner, database, namespace/relation, native identity/OID, schema fingerprint, mapping types, ordered keys, effects, and bounded batch size. |
| Destination isolation | Pass | File-store directories are hashes of the complete managed ledger key, candidates are content-addressed, and pointer publication is temp-file + fsync + rename + directory sync. |
| Safety and scope | Pass | No raw connection strings, credentials, generic SQL, direct database handle, connector manifest, or excluded issue files were added or changed. |

No actionable findings remain.
