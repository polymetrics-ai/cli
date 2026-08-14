# TDD ledger: issue #4087

## Slice 1 — typed aliases

- **Red:** recorded 2026-08-14. Added focused parsing and `RunETL` tests for `full_refresh_overwrite_deduped` and `incremental_append_deduped`; `go test -timeout 20m ./internal/app -run 'Test(DedupedLegacyAliasesUseTypedContractsBeforeSourceIO|CanonicalSyncModesRetainParsedContracts)$'` failed as expected. Both normal and persisted-legacy parsing returned an empty `ContractMode` and `LegacyCompatibility:true`; failures are at `sync_modes_test.go:289` before the typed pre-I/O assertion.
- **Green:** recorded 2026-08-14. `syncModeDefinitions` is the single connector-neutral authority used by both direct and persisted-legacy parsing. The aliases now return `ModeFullOverwrite` and `ModeIncrementalDedupe` with typed admission; the focused test passed and `go test -timeout 20m ./internal/app` passed in 159.862s.
- **Refactor:** recorded 2026-08-14. Removed tests that asserted the now-invalid legacy dedupe execution, retained unrelated cursor and warehouse isolation coverage on `incremental_append`, and added closed-mode control assertions for every canonical contract. The control also verifies the pre-existing `incremental_append` persisted-compatibility execution remains unchanged, while the other typed contracts retain their pre-I/O refusal without a transport.

## Slice 3 — surface and certification parity

- **Red:** recorded 2026-08-14. Before the help-source change, `go test -timeout 20m ./internal/cli -run TestETLHelpListsAllSyncModes$` failed because neither compatibility description identified the typed canonical admission. The real CLI certification sweep likewise treated the typed `Error`/exit-1 response as a failed expectation for an `ETLRun`.
- **Green:** recorded 2026-08-14. Updated generic help, generated `docs/cli` output (via `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`), website docs, and the connector-neutral certification expectation. `go test -timeout 20m ./internal/connectors/certify -run TestSourceStagesAgainstSample$` and `go test -timeout 20m ./internal/cli -run 'Test(CertifyCLISingleConnectorPassExitsZero|ETLHelpListsAllSyncModes)$'` pass.

## Planned checks

- `go test -timeout 20m ./internal/app -run 'Test(.*Legacy.*Typed|.*Canonical.*)'`
- `go test -timeout 20m ./internal/app`
- `go test -timeout 20m ./internal/synccontract ./internal/synctransport`
- `go test -timeout 20m ./internal/connectors/certify`
- `gofmt -w internal/app/sync_modes.go internal/app/sync_modes_test.go`
- `go vet ./...`
- `go build ./cmd/pm`
- individual `make verify` gates, including `connectorgen-surface-sync`
