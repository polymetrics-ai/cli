# Verification checklist — Issue #3981 managed-target ownership and provisioning

## Inline verify-work status

Pending implementation. This checklist will record command output and any
gap/correction loop; no claim of a green result is made before the tests run.

## Required focused checks

- [x] `go test -timeout 20m -count=1 ./internal/connectors/database -run '^TestManagedTargetProvisioningTruthTable$'`
- [x] `go test -race -timeout 20m -count=1 ./internal/connectors/database -run '^TestManagedTargetProvisioningTruthTable$'`
- [ ] `go test -timeout 20m -count=1 ./internal/connectors/database`
- [ ] `go test -timeout 20m -count=1 ./internal/warehouse`
- [ ] `go test -timeout 20m -count=1 ./internal/synccontract`
- [ ] `go test -timeout 20m -count=1 ./internal/app`
- [ ] `go test -timeout 20m -count=1 ./internal/cli`

## Required static/repository checks

- [ ] `gofmt -d` on changed Go files and `git diff --check`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] Individual applicable make gates: tidy, lint, docs/smoke, agent contract,
  connectorgen validation/surface sync, connector boundary, and release workflow.

## Truth and delivery boundaries

- [ ] State-transition assertions cover every required truth-table row; no test
  relies only on exit status.
- [ ] No generic SQL, PostgreSQL driver/DDL, transport, write/query/CDC, CLI,
  docs parity, or capability promotion change is present.
- [ ] Inline `verify-work`, any gaps-only loop, code review,
  Shepherd-compatible verdict, and no-mistakes result are recorded.
- [ ] Child PR targets exactly `feat/3972-postgres-parity`, is draft, includes
  `Refs #3981` and `Refs #3972`, and is not merged.
