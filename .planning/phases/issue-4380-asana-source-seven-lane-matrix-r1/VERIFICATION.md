# Verification checklist — Issue #4380

## Completed green evidence

- [x] `jq empty internal/connectors/defs/asana/sources/asana-source-lane-matrix.json internal/connectors/defs/asana/enabled_connector_contract.json`
- [x] Deliberate red: an in-memory copy missing `asana.rest.addCustomFieldSettingForGoal` / `sync_transport` failed with `missing lane cell`.
- [x] Green reconciliation: 249 unique lock IDs, exact source method/path/operation/citation parity, seven cells per row, 64 descriptor pagination candidates, and the expected lane-disposition counts.
- [x] `go test ./cmd/connectorgen -run 'TestParseSourceImportLockAcceptsAsanaV3DocumentOwnedInventory|TestSourceImportRetainedAsanaV3EventInventoryResolvesExactlyFiveSchemas|TestSourceImportRetainedAsanaPreservesLockedRESTOperationIDs|TestRetainedAsanaSourceImportSelectsSourceBackedFanOutETLStreams|TestRetainedAsanaMultipartActionsCoverLockedAttachmentOperation' -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/asana`
- [x] `go run ./cmd/connectorgen declaration-admission internal/connectors/defs --json` → one connector / one source operation / zero findings in the current admission inventory.
- [x] `go test ./cmd/connectorgen -run 'TestEnabledConnectorContractsLoadThroughNormalValidation|TestEnabledConnectorContractBindsPrimaryV3RetainedEvidence|TestEnabledConnectorContractsKeepExecutableLanesImplementedWhenSourceMappingIsPartial' -count=1`
- [x] `go test ./internal/connectors/defs/asana -count=1`
- [x] `git diff --check`; changed-path review limits production changes to the Asana contract, matrix, local validation test, and the contract expectation correction.

## Non-blocking current-tree boundaries

- `source-import asana --check` and `--read-projection-only --check` report current derived bundle-projection drift (`writes=0 cli=297`). The matrix is not a source-projection input and this lane does not alter generated streams, writes, or CLI surface.
- The broader `TestRetainedAsanaSourceImportRejectsReadProjectionDrift` currently fails before its mutation cases because the current source-projection state reports `partial mutation coverage disposition source operation "asana.rest.createCustomField" has no implemented declared action`. This is source-projection/runtime-admission work outside the matrix lane; the independent retained-lock, event-inventory, operation-ID, fan-out, and attachment source tests above pass.
- `source-materialize asana --check` intentionally refuses the v3 lock (`v4 ... materialization block is required`); source-import is the applicable source check.
- `surface-sync internal/connectors/defs --check` reports existing Asana CLI projection drift (`writes=0 cli=297`) and a Bitbucket missing descriptor. Those global/source-projection repairs are outside the Asana matrix lane and were not changed.

## Inline lifecycle and review

- Resolved `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review` through `scripts/gsd sources`; executed their planning, TDD, verification, and review work inline because this assignment prohibits spawning runtime roles.
- Manual code review checked the lane-classification boundaries: pagination rather than GET drives ETL; API-surface/write-action evidence rather than method drives direct/reverse write; the attachment keeps two action variants under one source ID; event contract selects exactly three sync IDs; and every other cell remains explicit.
