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

## PR #595 convention CI repair

- [x] `HEAD_REF=fm/cli-pm-broker-contract-fixtures-r1 bash <branch-name-policy>` accepts the
  validation branch.
- [x] `HEAD_REF=feature/not-valid bash <branch-name-policy>` still rejects invalid ordinary
  branches.
- [x] `HEAD_REF=feat/valid-branch bash <branch-name-policy>` still accepts ordinary Conventional
  Commit branch names.
- [x] `require-linked-issue` still applies to `fm/*`, matching `.github/workflows/pr-issue-guard.yml`.
- [x] `go test ./cmd/prissueguard ./internal/coordination/issueguard` passes to confirm the issue
  guard behavior itself was not loosened.
- [x] `git diff --check` passes.
