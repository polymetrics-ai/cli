# TDD ledger: Issue #3866

## Planned cycle

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Family matrix contract | `go test -count=1 -timeout 20m ./internal/synctransport -run '^TestTransportFamilyHalfPathConformance$'` failed with `undefined: transportFamilyConformanceCases` in `family_conformance_test.go`. | Pending: minimal private fixture supports the four family half-paths and exact records. | The missing helper is deliberately test-only; no production surface is required. |
| Mode and invariant rows | Pending: add each named happy/bad/edge case before implementation. | Pending: assert typed pre-I/O errors, counts, durable checkpoint results, and no replay. | Pending. |
| Sensitivity | Pending: schema-valid scratch binding failure for one named path. | Pending: restore exact binding and rerun the named path. | Scratch change is never committed. |

No production edit has been made before this ledger.
