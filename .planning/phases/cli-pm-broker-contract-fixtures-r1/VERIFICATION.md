# Verification Checklist

- [x] `go test ./internal/pmbroker/contract/v1`
- [x] `gofmt -w internal/pmbroker/contract/v1`
- [x] `go test ./internal/pmbroker/...`
- [x] `git diff --check`
- [x] Broader local gate: `go test ./...`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] Branch committed and pushed to `fm/cli-pm-broker-contract-fixtures-r1`
- [x] PR opened against `integration/pm-broker-production-program`: https://github.com/polymetrics-ai/cli/pull/594
- [ ] no-mistakes / CI-ready validation attached to authoritative integration-base PR
