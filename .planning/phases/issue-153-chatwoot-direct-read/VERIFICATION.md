# Verification — Issue #153 Chatwoot direct read

## Targeted checks

- [x] `go test ./internal/connectors/engine -run DirectRead -count=1`
- [x] `go test ./internal/connectors/commandrunner -run DirectRead -count=1`
- [x] `go test ./internal/cli -run Chatwoot -count=1`
- [x] `go test ./cmd/connectorgen -run Chatwoot -count=1`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- [x] `cd website && npm run test:unit -- --run tests/api/connector-data.test.ts`
- [x] `git diff --check`

## Full gates

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`

## Remote/review

- [x] Push branch `feat/153-chatwoot-direct-read`.
- [x] Open stacked PR #249 to `feat/148-chatwoot-cli-parity` with `Refs #153`.
- [x] PR #249 remote checks passed.
- [x] CodeRabbit skipped non-default base; parent PR #223 fallback review was requested after integration and replied `Review finished` with no inline findings returned by GitHub API.
- [x] Parent PR #223 checks passed after #153 integration.
