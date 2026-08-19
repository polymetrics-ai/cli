# Verification checklist — sync-contract durability defects

## State-load audit

| Call site | Assignment to `a.state` | Verdict |
| --- | --- | --- |
| `internal/app/app.go:188` (`load`) | Yes, via `normalizeLoadedState` | Normalized: initializes state maps, marks legacy adapters, and applies credential coordination migration. |
| `internal/app/app.go:2035` (`loadReversePlan`) | Yes, via `normalizeLoadedState` | Normalized: the reverse-plan lookup cannot replace migrated in-memory state with raw persisted state. |

`rg -n 'a\.store\.Load\(' internal/app -g '*.go'` found no other `App` store-load call sites.

## Required behavior

- [x] Exact persisted-legacy Open → unknown reverse plan → ETL regression failed before the fix.
- [x] Exact persisted-legacy regression passes after the fix.
- [x] Directory-chain sync regression failed before the fix.
- [x] Directory-chain sync regression passes after the fix.
- [x] Every `a.store.Load()` assignment audit result is recorded above.
- [x] No secret, connection string, or warehouse content was printed or stored.

## Local gates

- [x] Focused affected-package tests.
- [x] `go test ./internal/app -count=1`.
- [x] `go test ./internal/durability -count=1`.
- [x] `go test ./internal/cli -count=1`.
- [x] `gofmt -w internal/app`.
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] Scoped `golangci-lint run ./internal/app/...` has no finding in this change; it reports only
      two pre-existing test-only `errcheck` findings in `query_engine_helpers_test.go` and
      `reverse_approval_test.go`.
- [x] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
      `make agent-contract-check`, `make connectorgen-validate`,
      `make connectorgen-surface-sync`, `make connector-boundary`, and
      `make release-workflow-check`.
- [x] Inline `execute-phase`, `verify-work`, and `code-review` prompt execution recorded.

## Durability test boundary

`TestRunWarehouseETLSyncsNewDirectoryParentChainBeforeAcknowledgement` proves that a successful
run has invoked the directory sync primitive for `_pm_raw`, the warehouse leaf, every newly created
parent, and the first pre-existing ancestor before its checkpoint is acknowledged. It cannot
simulate an actual kernel/filesystem power loss; `durability.SyncDirectory` remains the tested
platform primitive for the actual directory flush.
