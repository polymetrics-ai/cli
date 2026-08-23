# TDD Ledger — Issue 4316

| Slice | State | Red evidence | Green evidence | Refactor / notes |
| --- | --- | --- | --- | --- |
| Declaration-owned resolver | green | **Red:** `go test -timeout 20m ./internal/connectors/engine -run '^TestHelpScoutV3DirectReadsUseTheirDeclaredRoute$' -count=1` failed for all five operation IDs (not found; 0 provider requests). | **Green:** the same command passes after the declaration resolver and Help Scout route entries land. | One resolver is used by selected operation, stream, and write declarations. |
| Missing/conflicting route diagnostics | green | **Red:** no selected-route declaration existed, so the Help Scout operation test could not reach a provider request. | `TestOperationRoutesFailClosedBeforeProviderIO` and `TestOperationRoutesRejectConflictingBasesBeforeProviderIO` prove source-traced missing-foundation/error conflict results and zero hits. | No fallback or caller route/URL channel was added. |
| Direct read/write cross-surface | green | **Red:** no operation metadata could bind Help Scout v3 into the shared operation executor. | `TestOperationRoutesUseOneDeclaredRouteForDirectReadAndWrite` captures exact `/v3/read` and `/v3/write` provider requests after preview/approval. | Reverse-ETL approval remains mandatory; write routes reuse the same resolver. |
| Help Scout v3 acceptance | green | **Red:** five real Help Scout direct reads failed before the definition declared their operations. | `TestHelpScoutV3DirectReadsUseTheirDeclaredRoute` passes and captures every source-locked v3 path with a configured `/v2` base. | `surface-sync` regenerated CLI/endpoint projections and website data was regenerated. |

Red and green command output will be appended as each slice runs. No production `cmd/` or `internal/` change may precede its corresponding red evidence.
