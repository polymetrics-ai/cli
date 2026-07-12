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

- [x] Push branch `feat/155-chatwoot-sensitive-admin-policy`.
- [x] Open stacked PR #264 to `feat/148-chatwoot-cli-parity` with `Refs #155`.
- [ ] Route CodeRabbit parent fallback review on PR #223 because non-default-base PR #264 was not reviewed.

## Parent integration

- [x] PR #264 remote checks passed.
- [x] PR #264 squash-merged into parent branch as `debf010a`.
- [x] Parent post-integration local gates passed: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`.
- [ ] Parent PR #223 CodeRabbit fallback review for integrated #155 commits.
