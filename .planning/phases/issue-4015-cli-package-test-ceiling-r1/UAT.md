# UAT — CLI package test-ceiling foundation

All deliverables are observable through automated evidence; no judgment-dependent UI or external-service check applies.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| One immutable production binary is shared by real-binary tests | The strengthened existing `TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle` fails red on two distinct paths and passes green on one stable fixture path while retaining its original lifecycle assertions. | pass |
| No test is lost or weakened | Before/after `go test -list '.*' ./internal/cli` inventories contain the same 263 runnable names and matching SHA-256. | pass |
| The normal verification command no longer collides with the ceiling locally | Final rebased `make verify` runs `go test -timeout 20m ./...` and reports `internal/cli` passing in 847.535s (29.4% below the ceiling). | pass |
| Fixture cleanup does not mask package failures | `TestMain` calls the existing budget guard, removes the package-owned temp directory, preserves either failure in its exit code, and only then calls `os.Exit`. | pass |

See `MEASUREMENTS.md` and `TDD-LEDGER.md` for exact commands, timing, and red/green results.
