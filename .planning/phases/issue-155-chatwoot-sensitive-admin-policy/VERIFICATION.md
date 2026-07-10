# Verification — Issue #155 Chatwoot sensitive/admin/destructive policy

## Targeted checks

- [x] `go test ./internal/connectors/engine -run Write -count=1`
- [x] `go test ./internal/connectors/commandrunner -run Write -count=1`
- [x] `go test ./cmd/connectorgen -run Chatwoot -count=1`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- [x] `go test ./internal/connectors/bundleregistry -run Chatwoot -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- [x] `cd website && pnpm run gen:website-data`
- [x] `cd website && pnpm run test:unit -- --run tests/api/connector-data.test.ts`
- [x] `git diff --check`

## Full gates

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`

## Remote/review

- [ ] Push branch `feat/155-chatwoot-sensitive-admin-policy`.
- [ ] Open stacked PR to `feat/148-chatwoot-cli-parity` with `Refs #155`.
- [ ] Route CodeRabbit/parent fallback review if the non-default-base sub-PR is skipped.
