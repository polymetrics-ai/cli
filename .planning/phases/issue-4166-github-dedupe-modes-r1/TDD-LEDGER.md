# TDD Ledger: GitHub dedupe modes r1

## Planned test contract

| Slice | Class | Named assertion |
| --- | --- | --- |
| Generator intersection | Happy | An admitted GitHub mode receives generated `declared` and `implemented` facts only when the source transport, selected destination, and apply strategy all support it. |
| Generator intersection | Bad | A mode absent from source transport has both claimed fields false despite coarse connector `read:true`; the failure is detectable before a generated file is accepted. |
| Generator intersection | Edge | Regeneration across all affected shards is deterministic: a second run has no bytes to write and a still-supported mode stays true. |
| GitHub preflight | Happy | Each named mode resolves registered source, destination, and apply strategy through the production preflight path. |
| GitHub preflight | Bad | A non-admitted mode returns its specific typed pre-I/O error and no source fetch occurs. |
| GitHub preflight | Edge | `incremental_append` follows its existing legacy route unchanged and destination action inventory remains exactly two names. |
| Dedupe runtime | Happy | Two occurrences of the same stable GitHub record key yield one current warehouse row in `incremental_dedupe`. |
| Dedupe runtime | Bad | Missing stable key/cursor or an invalid mode fails before source I/O, preserving an honest no-semantic-refusal. |
| Dedupe runtime | Edge | Replaying unchanged source data is idempotent for both current and history apply forms; history has no duplicate current interval for the same key. |
| Live provider | Happy | A fresh built `pm` reads bounded private-repository data and independent GitHub REST read-back confirms the provider object used for the proof. |
| Live provider | Bad | Deliberately invalid local mode/config is rejected before GitHub I/O with the typed error; no write action is attempted. |
| Live provider | Edge | A second run/replay against the same source key preserves dedupe/history record counts; command output and artifacts contain no credential value. |

## Actual evidence

### 2026-08-16 — red checkpoint

- Red: `go test -timeout 20m ./cmd/connectorgen -run 'TestCertificationSyncModeReadRequiresDeclaredSourceTransportMode'` failed because a deliberately narrowed GitHub source transport still generated `declared:true` and `implemented:true` for `incremental_dedupe` from coarse `read:true`.
- Red: `go test -timeout 20m ./internal/app -run 'TestGithubContractDedupeModesMaterializeCurrentAndHistoryRows'` failed before provider I/O for both modes with `sync mode ... is not executable: no matching closed source/destination transport has completed externally verified conformance`.
- Green: `go test -timeout 20m ./cmd/connectorgen -run 'TestCertificationSyncMode(ReadRequiresDeclaredSourceTransportMode|DatabaseWriteStubIsNotImplemented|ChangeCaptureRequiresDatabaseReadIntoWarehouse)'` passed after mode cells began reading their source/destination role declarations.
- Green: `go test -timeout 20m ./internal/app -run 'TestOpenRegistersDefinitionOwnedProductionTransports|TestGithubContractDedupeModesMaterializeCurrentAndHistoryRows|TestGithubPullRequestsETLSupportsLegacyExecutableModes|TestOpenRefusesPostgresUnpairedHistoryModeBeforeExecutorIO'` passed after source admission, warehouse apply strategies, and production dispatch were added.

### 2026-08-16 — changed-page replay checkpoint

- Red: after a real GitHub pull-request update, and after removing the test-only checkpoint reset, `go test -timeout 20m ./internal/app -run '^TestGithubContractDedupeModesMaterializeCurrentAndHistoryRows$'` failed both cases with `sync rebootstrap required: invalid_checkpoint: declarative stream resume page is no longer present`.
- Green: `go test -timeout 20m ./internal/app -run 'TestGithubContractDedupeModesMaterializeCurrentAndHistoryRows|TestGithubPullRequestsETLSupportsLegacyExecutableModes|TestOpenRegistersDefinitionOwnedProductionTransports'` passed after the two declared replay-safe dedupe modes re-emitted their bounded declarative page and let the identity-aware warehouse apply own replay suppression.
- Green: a fresh built `pm` ran both modes against retained private GitHub PR `karthik-sivadas/pm-truth-github-dedupe-modes-build-r1#1`; independent `pm github pr list` and `pm query run` reads proved one current dedupe row, two closed/open history rows after an update, and no extra row/version after the final replay.
- Manual GSD fallback: all required adapter prompts were resolved; inline execution is required because this task explicitly runs as one autonomous worker without compatible isolated GSD workers.
