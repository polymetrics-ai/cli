# TDD ledger: issue #4087

## Slice 1 — typed aliases

- **Red:** recorded 2026-08-14. Added focused parsing and `RunETL` tests for `full_refresh_overwrite_deduped` and `incremental_append_deduped`; `go test -timeout 20m ./internal/app -run 'Test(DedupedLegacyAliasesUseTypedContractsBeforeSourceIO|CanonicalSyncModesRetainParsedContracts)$'` failed as expected. Both normal and persisted-legacy parsing returned an empty `ContractMode` and `LegacyCompatibility:true`; failures are at `sync_modes_test.go:289` before the typed pre-I/O assertion.
- **Green:** pending. Centralize mapping and use the closed full-overwrite/incremental-dedupe contracts so each alias returns the existing typed pre-I/O refusal before source I/O when no transport is registered.
- **Refactor:** pending. Keep the mapping connector-neutral and single-sourced; retain unchanged legacy and canonical behavior under explicit control tests.

## Planned checks

- `go test -timeout 20m ./internal/app -run 'Test(.*Legacy.*Typed|.*Canonical.*)'`
- `go test -timeout 20m ./internal/app`
- `go test -timeout 20m ./internal/synccontract ./internal/synctransport`
- `gofmt -w internal/app/sync_modes.go internal/app/sync_modes_test.go`
- `go vet ./...`
- `go build ./cmd/pm`
- individual `make verify` gates, including `connectorgen-surface-sync`
