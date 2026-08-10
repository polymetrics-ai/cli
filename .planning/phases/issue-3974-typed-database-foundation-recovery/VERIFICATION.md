# Verification checklist — Issue #3974 typed database foundation recovery

## Required focused checks

- [ ] `go test -timeout 20m -count=1 ./internal/connectors/database`
- [ ] `go test -timeout 20m -count=1 ./internal/warehouse`
- [ ] `go test -timeout 20m -count=1 ./internal/synccontract`
- [ ] `go test -timeout 20m -count=1 ./internal/connectors/engine -run '^TestBundleLoadPostgresDatabaseDefinitionWithoutCapabilityPromotion$'`
- [ ] `go test -timeout 20m -count=1 ./internal/connectors/native/postgres -run '^TestPostgresDatabaseDriverReferenceSeam$'`
- [ ] relevant bundle/registry and `./internal/app` suites
- [ ] `go test -timeout 20m -count=1 ./internal/cli`

## Required static/repository checks

- [ ] `gofmt -w` changed Go paths
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`
- [ ] `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`
- [ ] `make connector-boundary`, `make release-workflow-check`

## Truth boundaries

- [ ] `database.json` admits no modes.
- [ ] PostgreSQL metadata has `write:false`, `query:false`, and `cdc:false`.
- [ ] No generic SQL/query/write surface, target executor, canonical mode execution, or CDC v2 was introduced.
- [ ] #3995 status is recorded as a bounded compatible verdict or external dependency; never automatic approval when outside ancestry.

## Delivery checks

- [ ] Inline `verify-work` result recorded; gaps are repaired through `plan-phase --gaps` and `execute-phase --gaps-only` if any exist.
- [ ] Deep inline code review findings are dispositioned.
- [ ] no-mistakes gates are driven without `--yes`.
- [ ] PR targets `feat/3972-postgres-parity`, uses both Refs lines, and preserves/linkes #4014 audit history.
