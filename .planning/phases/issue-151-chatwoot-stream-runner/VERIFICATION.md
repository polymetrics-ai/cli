# Verification — Issue #151 Chatwoot stream runner

## Planned targeted checks

- [ ] `go test ./internal/connectors/engine -run TestReadFanOutRequestPaginationOverrideAllowsDifferentChildPagination -count=1`
- [ ] `go test ./internal/connectors/conformance -run TestChatwootStreamRunnerSweep -count=1`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
- [ ] `go test ./cmd/connectorgen -run Chatwoot -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `git diff --check`

## Planned full gates

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## Review / remote

- [ ] Push branch `feat/151-chatwoot-stream-runner`.
- [ ] Open stacked PR to `feat/148-chatwoot-cli-parity` with `Refs #151`.
- [ ] If CodeRabbit skips non-default base, request/record parent PR #223 fallback review after integration.
