# Issue #4359 — Verification checklist

## Targeted proof

- [x] `go test -timeout 20m ./internal/connectors/engine -run 'CircleCICompositeProviderPathIdentity|TestBundleLoadRejectsCompositeProviderPathIdentityForAnotherConnector|TestResolveImplementedCommandBinding(ProvesEveryCircleCICompositeProjectSlugBinding|RejectsUnprovenCircleCICompositeProjectSlugSubstitutions)$' -count=1`
- [x] `go test -timeout 20m ./internal/connectors/commandrunner` — pass.
- [x] `go test -timeout 20m ./internal/app` — pass (cached confirmation after full package run).
- [ ] `go test -timeout 20m ./internal/cli` — running at this checkpoint.
- [x] `go build ./cmd/pm` — pass.
- [~] Eleven credential-boundary probes — deliberately deferred to Batch 1 integration. This isolated foundation has no generated CircleCI `cli_surface.json`; the built binary returns `unknown command "circleci"` for all eleven, which Firstmate confirmed is expected and must not be worked around. No credential, fixture input, or provider I/O was supplied. Batch 1 must prove each reaches exactly `missing --credential` in a fresh project.

## Generated/source checks

- [ ] `go test -timeout 20m ./cmd/connectorgen -count=1`
- [~] `go run ./cmd/connectorgen source-import circleci --out internal/connectors/defs/circleci/sources/circleci-operation-source-lock.json --check` — correctly cannot run on current main: the retained Batch-1 CircleCI `sources/` directory is intentionally absent. No source lock was fabricated or changed.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 553 connectors, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` — 553 scanned, 0 changes.
- [x] `go run ./cmd/connectorgen declaration-admission internal/connectors/defs` — 1 connector, 1 source operation, 0 findings.
- [x] `go run ./cmd/connectorgen operation-evidence . --check` — current, 1,525 rows / 5 rollups / fixed-100 passed.

## Repository checks

- [x] `gofmt -w` on changed Go files.
- [x] `go vet ./internal/connectors/engine ./internal/connectors/defs ./internal/connectors/commandrunner ./internal/app ./internal/cli`.
- [ ] `make tidy-check` / `make lint` / `make smoke-no-build` / `make agent-contract-check` / `make connector-boundary` / `make release-workflow-check` — running together as their individual Make targets at this checkpoint.
- [x] `make docs-check` — connector docs validated.
- [x] Website data/type check not applicable: no CLI, docs, website, or generated website-data source changed; `make docs-check` covers connector-document parity.
- [x] `git diff --check` before implementation commits.

## Review and delivery

- [x] Fresh-context Codex audit replaces unavailable Claude audit and records immutable source SHA.
- [x] Fresh-context Codex re-review records final code SHA `880c3c452274b227d91450aa5680188087f95a0e`, six-lane results, and PASS/no code findings.
- [ ] PR base API response is exactly `main`.
