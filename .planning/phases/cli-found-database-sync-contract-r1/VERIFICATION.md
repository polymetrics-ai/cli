# Verification checklist — Issue 3810: shared database sync contract

## Required behavioral checks

- [x] Seven and only seven contract modes validate.
- [x] A contract mode is not executable without matching native executor and complete reusable
      fixture evidence.
- [x] State-envelope JSON round-trips all required fields, including a dedupe window, and opaque
      non-UTF-8 tokens byte-for-byte.
- [x] Per-partition state remains an array of independent structured positions.
- [x] Every listed recovery condition is typed and requires rebootstrap without clearing state.
- [x] State committer is called only after an explicit durable downstream acknowledgement.
- [x] Tombstones require identity/key/image/order and history deletes close `_valid_to`/
      `_is_current` validity windows with an explicit `_valid_from` protocol column.
- [x] Native contract has no REST/API-surface/raw SQL/raw HTTP/shell field.
- [x] Legacy scalar `cursor` state blocks resumption explicitly rather than silently scanning.
- [x] Existing legacy sync tests still prove append/overwrite/incremental behavior through the
      envelope adapter.

## Local command checklist

- [x] `go test ./internal/synccontract -count=1`
- [x] `go test ./internal/app -count=1`
- [x] `go test ./internal/connectors -count=1`
- [x] `go test ./internal/connectors/engine -count=1`
- [x] `go test ./internal/cli -count=1`
- [x] `gofmt -w internal/synccontract internal/app`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`
- [x] `scripts/gsd prompt verify-work cli-found-database-sync-contract-r1` fallback record / manual
      verification and `scripts/gsd prompt code-review cli-found-database-sync-contract-r1` review
      record.

## Deliberate not-applicable parity

No CLI command, flag, output, connector bundle surface, generated manual, docs, or website catalog
is changed here. #3748 and #3860 own user-facing surfacing; this foundation only creates importable
internal contracts and compatibility persistence. CLI/docs/website parity is therefore N/A for this
slice and must be addressed by those consumer lanes.

## Additional scoped lint note

`golangci-lint run ./internal/app/... ./internal/synccontract/...` was run after the repository
gate. The new materialized-final-file close warning was fixed. The remaining five findings are
pre-existing: an unrelated reader close in `internal/app/local_warehouse.go`,
`internal/app/query_engine_helpers_test.go`, `internal/app/reverse_approval_test.go`,
`internal/app/util.go`, and the existing nil-map check in `internal/app/app.go`. `make lint` itself
passes.
