# TDD LEDGER — caller-supplied identifier sets

| ID | Contract | Red evidence | Green evidence |
| --- | --- | --- | --- |
| R1 | Closed operation declaration requires min/max and a supported shape/wire | Pending targeted engine schema/load test before production change | Pending |
| R2 | Implemented direct-read command binds exactly one required `string_array` flag to every set | Pending focused connectorgen/runner test before production change | Pending |
| R3 | Flat identifier lists encode correctly for comma query, repeated query, JSON body, and path segment | Pending end-to-end test-only bundle test before production change | Pending |
| R4 | Cardinality and shape fail before networking without identifier disclosure | Pending engine/runner rejection test before production change | Pending |
| R5 | Explicit blank retains `[]` with zero minimum; absent remains rejected | Pending commandrunner test before production change | Pending |
