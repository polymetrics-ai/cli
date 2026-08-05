# Verification checklist — Issue 3810: shared database sync contract

## Required behavioral checks

- [ ] Seven and only seven contract modes validate.
- [ ] A contract mode is not executable without matching native executor and complete reusable
      fixture evidence.
- [ ] State-envelope JSON round-trips all required fields and opaque non-UTF-8 tokens byte-for-byte.
- [ ] Per-partition state remains an array of independent structured positions.
- [ ] Every listed recovery condition is typed and requires rebootstrap without clearing state.
- [ ] State committer is called only after an explicit durable downstream acknowledgement.
- [ ] Tombstones require identity/key/image/order and history deletes close validity windows.
- [ ] Native contract has no REST/API-surface/raw SQL/raw HTTP/shell field.
- [ ] Legacy scalar `cursor` state blocks resumption explicitly rather than silently scanning.
- [ ] Existing legacy sync tests still prove append/overwrite/incremental behavior through the
      envelope adapter.

## Local command checklist

- [ ] `go test ./internal/synccontract -count=1`
- [ ] `go test ./internal/app -count=1`
- [ ] `go test ./internal/connectors -count=1`
- [ ] `go test ./internal/connectors/engine -count=1`
- [ ] `go test ./internal/cli -count=1`
- [ ] `gofmt -w internal/synccontract internal/app`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] `scripts/gsd prompt verify-work cli-found-database-sync-contract-r1` fallback record / manual
      verification and `scripts/gsd prompt code-review cli-found-database-sync-contract-r1` review
      record.

## Deliberate not-applicable parity

No CLI command, flag, output, connector bundle surface, generated manual, docs, or website catalog
is changed here. #3748 and #3860 own user-facing surfacing; this foundation only creates importable
internal contracts and compatibility persistence. CLI/docs/website parity is therefore N/A for this
slice and must be addressed by those consumer lanes.
