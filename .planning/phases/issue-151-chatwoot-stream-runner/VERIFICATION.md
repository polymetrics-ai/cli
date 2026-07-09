# Verification — Issue #151 Chatwoot stream runner

## Planned targeted checks

- [x] `go test ./internal/connectors/engine -run TestReadFanOutRequestPaginationOverrideAllowsDifferentChildPagination -count=1`
- [x] `go test ./internal/connectors/engine -run 'TestReadFanOutRequestPaginationOverrideAllowsDifferentChildPagination|TestLoadStreamsFanOutRequestPaginationRoundTrips' -count=1`
- [x] `go test ./internal/connectors/conformance -run TestChatwootStreamRunnerSweep -count=1`
- [x] `go test ./internal/connectors/conformance -run 'TestChatwootStreamRunnerSweep|TestConformance/chatwoot' -count=1`
- [x] `go test ./cmd/connectorgen -run Chatwoot -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `cd website && npm run test:unit -- --run tests/api/connector-data.test.ts`
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- [x] `git diff --check`

## Planned full gates

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`

## Review / remote

- [ ] Push branch `feat/151-chatwoot-stream-runner`.
- [ ] Open stacked PR to `feat/148-chatwoot-cli-parity` with `Refs #151`.
- [ ] If CodeRabbit skips non-default base, request/record parent PR #223 fallback review after integration.
