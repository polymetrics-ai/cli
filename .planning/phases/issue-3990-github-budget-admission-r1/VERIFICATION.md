# Verification: Issue #3990

## Required checks

- [x] `go test -timeout 20m ./internal/connectors/engine/...`
- [x] `go test -timeout 20m ./internal/coordination/...`
- [x] `go test -timeout 20m ./internal/connectors/certify/...`
- [x] `go test -timeout 20m ./internal/connectors/connsdk`
- [x] Focused CLI certification-report test and `pm help connectors`/`pm connectors` parity checks.
- [x] Focused multi-process shared-budget proof with the provider-double second-send count equal to zero.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go run ./cmd/connectorgen surface-sync --check`
- [x] `git diff --check`
- [x] `go vet ./...`
- [ ] Push and open a PR explicitly based on `integration/4015-mvp-flat-r1`; verify the API-reported base. CI must report the protected integration checks (verify, govulncheck, and CodeQL) green. `no-mistakes axi run` is explicitly excluded by captain delivery recovery because it retargets to `main`.

## Completed local evidence

- `go run ./cmd/connectorgen certification-matrix --connector github` regenerated the allowlisted GitHub shard; its subsequent `--check` passed.
- `POLYMETRICS_COORDINATION_INTEGRATION=1 go test -tags coordinationintegration -timeout 20m -v ./internal/connectors/engine -run '^TestGitHubCertificationSharedBudgetCoordinatesSeparateProcesses$'` passed: two OS processes shared capacity one, the first sent once, and the second sent zero times.
- `make lint`, `make docs-check`, `make github-parity-artifacts-check`, `make connector-canon-check`, and `make release-workflow-check` passed.
- The local `agent-contract-check` sees a pre-existing duplicate project-agent file beneath untracked workspace `.tmp/`; it is not in this branch or an integration CI checkout, so it is recorded as environment noise rather than changed here.

## Live evidence rule

Every failing-admission test asserts its physical provider-double request count is zero. Every success case asserts the exact request/event/ledger state claimed; no test treats an error-free return as proof.

## Out-of-scope finding protocol

Append `needs-decision:` to this record for any unrelated review finding rather than editing it in this branch. Known exclusion: #4125, `internal/coordination/shared_rate_limits.go` duration overflow. The invalid `sample` certification sweep target is routed to #4136; this issue runs GitHub certification evidence only.
