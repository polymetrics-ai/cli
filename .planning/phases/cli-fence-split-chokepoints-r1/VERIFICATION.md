# Verification checklist — Fence split chokepoints r1

## Behaviour-preservation checks

- [ ] GitHub and PostgreSQL shard union reconstructs their exact old aggregate payload.
- [ ] A normal scoped generation run leaves the other allowlisted shard byte-identical.
- [ ] All generated anchors name source symbols and no `:<digits>` source anchor remains.
- [ ] `connectorgen certification-matrix --check` detects allowlisted shard drift and source
      disappearance while ignoring non-allowlisted certification claims.
- [ ] A one-line shared-file insertion above an anchor produces zero generated shard changes after
      the refactor (baseline old-matrix diff count is recorded below).
- [ ] Existing app Open/ETL tests and PostgreSQL metadata/manifest tests pass unchanged.
- [ ] No capability values, error text, CLI output, or connector definition semantics change.

## Local commands

- [ ] `go test -count=1 ./cmd/connectorgen`
- [ ] `go test -count=1 ./internal/app`
- [ ] `go test -count=1 ./internal/connectors/native/postgres`
- [ ] `go run ./cmd/connectorgen certification-matrix --check`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connectorgen-certification-matrix`
- [ ] `make connector-boundary`
- [ ] `make connector-canon-check`
- [ ] `make release-workflow-check`
- [ ] no-mistakes pipeline returns `checks-passed`.

## Required report measurements

- Database test package on base: absent (`internal/connectors/database` does not exist on
  `2df18ee`).
- Old one-line-insertion capability-matrix diff: pending measurement before implementation.
- Shards produced / largest shard / aggregate consumers / post-change insertion diff: pending final
  regeneration.
- Flow equivalent share for GitHub and PostgreSQL, complete 556-connector consumer inventory, and
  final certification-surface size: pending final regeneration.
