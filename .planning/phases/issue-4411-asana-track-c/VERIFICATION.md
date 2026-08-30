# Verification checklist — Issue #4411 Asana Track C

## Completed green evidence

- [x] Deliberate invalid/missing-artifact validation: `go test ./internal/connectors/defs/asana -run '^TestAsanaSourceLaneArtifactsProjectTheTrackAMatrix$' -count=1 -v`. Its internal red counterparts reject an absent `sync_transport.json`, an erased direct-read backlink, and a `mapped_unproven` ETL cell promoted solely by descriptor pagination.
- [x] `go test ./internal/connectors/defs/asana -run '^TestAsanaTrackC' -count=1 -v` (17.9s): production embed → normal registry → CLI credential boundary for direct read/write/binary upload with zero provider sends; source-bound `tasks` → trusted local response at the declared origin → local DuckDB materialization/read-back/checkpoint; credential `base_url` override rejected before I/O.
- [x] `go test ./internal/connectors/defs/asana -run '^TestGetMembershipExecutesItsSourceLockedPathBinding$' -count=1 -v`: source-locked direct read reaches its exact local fake-provider path.
- [x] `go test ./internal/app -run '^TestAsana(EventTombstoneUsesDeletedButNeverRemoved|EventTombstoneIdentityIsStableWithinTokenWindow|EventSourceBootstrapTokenPrecedesExhaustiveSnapshot|EventSourceExpiredTokenRebootstrapsBeforeSnapshot|EventSourceCoalescesCompleteWindowBeforeHydration|EventSourceRefusesPartialTokenWindow|IncrementalAppendWarehouseRowsUseProviderTokenAndPreserveDeleteMarker)$' -count=1 -v`: constrained source-token transport proves bootstrap/rebootstrap, full-window gating, hydration/coalescing, provider-token checkpoint rows, and deleted-only tombstones.
- [x] `go test ./internal/connectors/defs/asana -count=1` (67.7s): full Asana package, including every-action direct-write and DuckDB reverse-ETL plan → preview → approval → local fake execution, attachment upload, source matrix, and artifact projection.
- [x] `go test ./cmd/connectorgen -run 'TestEnabledConnectorContractsLoadThroughNormalValidation|TestEnabledConnectorContractBindsPrimaryV3RetainedEvidence|TestEnabledConnectorContractsKeepExecutableLanesImplementedWhenSourceMappingIsPartial' -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/asana` → `1 connector(s) checked, 0 findings`.
- [x] `go run ./cmd/connectorgen declaration-admission internal/connectors/defs --json` → one connector / one source operation / zero findings.
- [x] `go run ./cmd/agentcontractgen check` → canonical contract and registered projections are current.
- [x] `jq empty internal/connectors/defs/asana/{enabled_connector_contract,missing-foundation,event_source_contract,sync_transport,streams,writes,cli_surface}.json internal/connectors/defs/asana/sources/{asana-source-lane-matrix,asana-operation-source-lock,asana-operation-descriptor}.json`
- [x] `git diff --check` before final evidence updates; rerun after final staging.

## Lane receipts and bounded outcomes

| Lane | Matrix disposition count | Executed proof / boundary |
| --- | --- | --- |
| Direct read | 119 implemented; 130 N/A | Embedded `agents get-agents-for-workspace` stops at the credential boundary with zero sends; `getMembership` runs its exact source-locked fake route. |
| Direct write | 130 implemented; 119 N/A | The embedded `tasks create` command is typed-write/API-surface bound, deliberately has no invented CLI source link, and full-package every-action fake execution remains green. |
| Binary download | 249 N/A | Artifact-projection red cases reject missing/fabricated links; no executable download claim exists. |
| Binary upload | 1 implemented; 248 N/A | The embedded attachment alias reaches the credential boundary with zero sends; full-package attachment plan/execution evidence is green. |
| ETL through DuckDB | 12 implemented; 52 mapped_unproven; 185 N/A | Only source-backed `tasks` is run here through local DuckDB. The copied-matrix promotion red case keeps all 52 mapped-unproven cells non-executable. |
| Reverse ETL from DuckDB | 130 implemented; 119 N/A | Full-package bulk test materializes local source rows and runs every named action through plan, preview, approval, and exactly one fake-provider request. |
| Sync transport through DuckDB | 3 implemented; 246 N/A | Source-token fake tests prove only task event/snapshot/hydration scope, token windows, and warehouse marker behavior; they make no API-to-API transport claim. |

## Atlas/gap result

All required foundations are existing Atlas reuse (`runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, `warehouse.reverse-etl.v1`, `transport.sync-contract.v1`, `asana.event-token-source.v1`, and `source.projection-admission.v1`). No new runtime foundation or demand record is required.

## Non-scoped baseline result (not changed or masked)

- `go run ./cmd/connectorgen source-import asana --check` remains red before and after this proof-only slice: `descriptor or derived bundle projection has drifted (writes=0 cli=297)`.
- `go run ./cmd/connectorgen source-import asana --read-projection-only --check` remains red: `descriptor or derived bundle projection has drifted (writes=0 cli=0)`.
- Track C changes only a connector-local Go proof test and planning evidence. It does not modify source-lock, descriptor, matrix, definitions, or any baseline projection input; these failures are recorded separately and are not made green by scope expansion.

## Non-goals

- No live Asana I/O or certification.
- No promotion of the 52 mapped-unproven ETL candidate cells.
- No alteration of Track B's matrix, lock, descriptor, definitions, or baseline projection inputs.
