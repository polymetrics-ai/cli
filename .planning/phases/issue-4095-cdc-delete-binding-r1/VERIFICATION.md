# Verification checklist — Issue 4095

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Explicit CDC delete reaches PostgreSQL history close | passed live | `TestPostgresManagedTargetIncrementalDedupeHistoryLive` reads the real target; the current version remains stored with `_is_current=false` and `_valid_to` set after its CDC-derived tombstone. |
| Physical absence is never a deletion instruction | passed live | `TestPostgresManagedTargetWorksetDeliveryLive` reads the real target after source omission (row retained) and after its explicit CDC-derived tombstone (only that row removed). |
| Shared mapping does not accept ambiguous delete keys | passed fake | `TestMappingContractV1MapsOnlyDeclaredTombstoneKeys` checks exact source→target keys and refuses missing/extra source fields before any target input. |
| Required package and live harness regression | passed | The specified package command and complete tagged native PostgreSQL dbtest command passed; retained output summaries are under `traces/`. |

## Planned non-test gates

- `gofmt -w` on changed Go files
- `go vet` on changed packages
- `go build ./cmd/pm`
- relevant individual `make verify` non-test gates, per repository policy
