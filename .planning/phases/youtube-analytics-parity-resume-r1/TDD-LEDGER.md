# TDD Ledger

Phase: youtube-analytics-parity-resume-r1

Record failing test evidence before production code for every behavior-adding task.

| Slice | Red command/evidence | Implementation | Green command/evidence |
| --- | --- | --- |
| `media.download` promotion | `go run ./cmd/pm youtube-analytics reports download --json` exited 7: `availability=planned: operation download_report executor is not implemented in this slice`. | Promoted only `download_report` to `binary_download`, using the existing bounded executor and a safe multi-segment `resourceName` mapping. | `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` passed; a built `pm youtube-analytics reports download` accepts the declared flags and reaches the normal missing-credential boundary rather than a planned-operation block. |
| Citation evidence | Documentation matrix had no per-request-field evidence record for the recovered bundle. | Recorded provider-owned field citations and the explicit tier-4 rationale for `media.download`; this temporary research record will be transferred without reinterpretation when the shared machine-readable convention lands. | 15 citation rows cover 42 request-field uses across all 16 documented operations; no tier-5 deferrals. Connector validation passed. |
| Typed stream query selectors | Source call-path review showed `buildInitialQuery` resolving `{{ query.mine }}` and `{{ query.groupId }}` before making `ReadRequest.Query` available, producing an unresolved-key error before any provider request. | Threaded `ReadRequest.Query` into the shared query-template interpolation environment while retaining mandatory plain-string templates and the final caller-query overlay. | `go test ./internal/connectors/engine -run 'TestReadRequestQuery(ResolvesStreamQueryTemplate|TemplateMissingFailsBeforeRequest)$' -count=1` covers successful selector resolution and fail-closed missing-selector handling. |

Historical red/green evidence from preserved phase `youtube-analytics-parity-3456` is recovery context only, never evidence for current main.
