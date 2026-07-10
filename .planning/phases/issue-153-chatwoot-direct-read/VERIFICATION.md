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

- [ ] Push branch `feat/153-chatwoot-direct-read`.
- [ ] Open stacked PR to `feat/148-chatwoot-cli-parity` with `Refs #153`.
- [ ] Route CodeRabbit/parent fallback review if the non-default-base sub-PR is skipped.
