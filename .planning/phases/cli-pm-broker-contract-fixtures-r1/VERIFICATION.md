# Verification Checklist

- [x] `go test ./internal/pmbroker/contract/v1`
- [x] `gofmt -w internal/pmbroker/contract/v1`
- [x] `go test ./internal/pmbroker/...`
- [x] `git diff --check`
- [x] Broader local gate: `go test ./...`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [ ] Branch committed and pushed to `fm/cli-pm-broker-contract-fixtures-r1`
- [ ] PR opened against `integration/pm-broker-production-program`
- [ ] no-mistakes / CI-ready validation attached to authoritative integration-base PR
