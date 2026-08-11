# Verification checklist — Issue #3981 managed-target ownership and provisioning

## Inline verify-work status

The manual inline `verify-work` fallback is green after one bounded correction
round. #4038 was created and linked from #3981 before fixing the
cross-provisioner target-lock gap. Its RED/GREEN evidence is retained in
`traces/cross-provisioner-lock-{red,green}.txt`; it stayed within the shared
driver-neutral contract and added no PostgreSQL/SQL or capability work.

## Required focused checks

- [x] `go test -timeout 20m -count=1 ./internal/connectors/database -run '^TestManagedTargetProvisioningTruthTable$'`
- [x] `go test -race -timeout 20m -count=1 ./internal/connectors/database -run '^TestManagedTargetProvisioningTruthTable$'`
- [x] `go test -timeout 20m -count=1 ./internal/connectors/database`
- [x] `go test -timeout 20m -count=1 ./internal/warehouse`
- [x] `go test -timeout 20m -count=1 ./internal/synccontract`
- [x] `go test -timeout 20m -count=1 ./internal/app` (123.767s)
- [x] `go test -timeout 20m -count=1 ./internal/cli` (273.648s)

## Required static/repository checks

- [x] `gofmt -d` on changed Go files and `git diff --check`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`; `golangci-lint run ./internal/connectors/database`; `make lint`
- [x] `make docs-check-no-build`; `make smoke-no-build`; `make agent-contract-check`
- [x] `make connectorgen-validate`; `make connectorgen-surface-sync`
- [x] `make github-parity-artifacts-check`; `make connectorgen-certification-matrix`
- [x] `make connector-runtime-preflight`; `make connector-canon-check`
- [x] `make connector-boundary` (clean); `make release-workflow-check`

## Truth and delivery boundaries

- [x] State-transition assertions cover every required truth-table row; no test
  relies only on exit status.
- [x] No generic SQL, PostgreSQL driver/DDL, transport, write/query/CDC, CLI,
  docs parity, or capability promotion change is present.
- [x] Inline `verify-work` and code review are recorded; no gaps-only loop is
  needed after #4038.
- [x] Bounded #3995 Shepherd-compatible `RETRY` verdict is recorded in
  `SHEPHERD-COMPATIBILITY.json`; it is not automatic approval and does not
  consume a correction round.
- [ ] no-mistakes result is recorded.
- [ ] Child PR targets exactly `feat/3972-postgres-parity`, is draft, includes
  `Refs #3981` and `Refs #3972`, and is not merged.
