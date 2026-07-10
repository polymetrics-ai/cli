# Verification — Issue #153 Chatwoot direct read

## Targeted checks

- [ ] `go test ./internal/connectors/engine -run DirectRead -count=1`
- [ ] `go test ./internal/connectors/commandrunner -run DirectRead -count=1`
- [ ] `go test ./internal/cli -run Chatwoot -count=1`
- [ ] `go test ./cmd/connectorgen -run Chatwoot -count=1`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- [ ] `cd website && npm run test:unit -- --run tests/api/connector-data.test.ts`
- [ ] `git diff --check`

## Full gates

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## Remote/review

- [ ] Push branch `feat/153-chatwoot-direct-read`.
- [ ] Open stacked PR to `feat/148-chatwoot-cli-parity` with `Refs #153`.
- [ ] Route CodeRabbit/parent fallback review if the non-default-base sub-PR is skipped.
