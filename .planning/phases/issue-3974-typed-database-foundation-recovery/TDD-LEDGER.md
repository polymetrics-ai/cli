# TDD ledger — Issue #3974 typed database foundation recovery

| ID | Guarantee | Red evidence | Green evidence |
| --- | --- | --- | --- |
| R1 | Recovery is necessary | `./internal/connectors/database` is absent on the parent branch. | The package and complete focused contract suite compile and pass after the ordered replay. |
| R2 | Strict typed definitions | Unknown/ambiguous members, invalid numeric limits, opaque mappings, and cancellation are refused. | `TestDatabaseDefinitionStrictLoadAndDefensiveProjection`, ambiguity/numeric/cancellation tests pass. |
| R3 | Exact types and bounds | Unsafe type conversion, unlimited resource values, or mutable projections can enter the contract. | Lossless-only logical plan and resource-policy tests pass with defensive copies. |
| R4 | Real native admission | A declaration or cross-leg/cross-driver admission can be used as execution authority. | Registry and synccontract tests require registered compatible, distinct per-leg admissions. |
| R5 | Warehouse mediation | A database value can carry both source and target or bypass the shared artifact. | Warehouse artifact and isolated inbound/outbound leg tests pass. |
| R6 | PostgreSQL stays non-executing | Recovery can accidentally alter TLS/CDC behavior or raise a capability. | PostgreSQL reference seam and embedded definition tests retain `write:false`, `query:false`, `cdc:false`. |
| R7 | Newer canon wins | Stale #4014 docs/connector wording can replace #4003 or current PostgreSQL hardening. | Recorded conflict dispositions retain current canon and fail-closed behavior. |
| R8 | Generated certification projection stays attributable | `make connectorgen-certification-matrix` detects a source-location drift after the #3974 replay. | Correction child #4026 records ownership; regenerated matrix and capability-false binary evidence are green without capability promotion. |

## Commands and evidence

**Red:** `go test -timeout 20m -count=1 ./internal/connectors/database -run '^TestDatabaseDefinitionStrictLoadAndDefensiveProjection$'` before replay. Its output is retained in `traces/typed-admission-red.txt`.

**Red:** semantic replay conflict status and the three overlap paths are retained in `traces/semantic-replay-red.txt` before manual reconciliation.

**Red:** correction round 1/5 records the failed derived-artifact check and branch-versus-parent attribution in `traces/capability-matrix-red.txt`; #4026 is its sole correction child.

**Green:** focused tests, runtime metadata assertions, and repository gates are recorded in `VERIFICATION.md` and `traces/` after replay.
