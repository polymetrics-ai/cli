# Verification Checklist

- [x] `go test ./internal/pmbroker/contract/v1`
- [x] `gofmt -w internal/pmbroker/contract/v1`
- [x] `go test ./internal/pmbroker/...`
- [x] `git diff --check`
- [x] Broader gate: `go test ./...`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] Branch committed on `fm/cli-pm-broker-client-contract-r1`
- [ ] PR opened against `integration/pm-broker-production-program`
- [ ] no-mistakes / CI-ready validation attached to the authoritative integration-base PR

## Review-Fix Verification

- [x] `gofmt -w internal/pmbroker/contract/v1`
- [x] `go test ./internal/pmbroker/contract/v1`
