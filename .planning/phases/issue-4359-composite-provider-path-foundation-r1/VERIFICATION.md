# Issue #4359 — Verification checklist

## Targeted proof

- [x] `go test -timeout 20m ./internal/connectors/engine -run 'CircleCICompositeProviderPathIdentity|TestBundleLoadRejectsCompositeProviderPathIdentityForAnotherConnector|TestResolveImplementedCommandBinding(ProvesEveryCircleCICompositeProjectSlugBinding|RejectsUnprovenCircleCICompositeProjectSlugSubstitutions)$' -count=1`
- [x] `go test -timeout 20m ./internal/connectors/commandrunner` — pass.
- [x] `go test -timeout 20m ./internal/app` — pass (cached confirmation after full package run).
- [x] `go test -timeout 20m ./internal/cli` — pass (cached confirmation after full package run).
- [x] `go build ./cmd/pm` — pass.
- [~] Eleven credential-boundary probes — deliberately deferred to Batch 1 integration. This isolated foundation has no generated CircleCI `cli_surface.json`; the built binary returns `unknown command "circleci"` for all eleven, which Firstmate confirmed is expected and must not be worked around. No credential, fixture input, or provider I/O was supplied. Batch 1 must prove each reaches exactly `missing --credential` in a fresh project.

## Generated/source checks

- [x] `go test -timeout 20m ./cmd/connectorgen` — pass.
- [~] `go run ./cmd/connectorgen source-import circleci --out internal/connectors/defs/circleci/sources/circleci-operation-source-lock.json --check` — correctly cannot run on current main: the retained Batch-1 CircleCI `sources/` directory is intentionally absent. No source lock was fabricated or changed.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 553 connectors, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` — 553 scanned, 0 changes.
- [x] `go run ./cmd/connectorgen declaration-admission internal/connectors/defs` — 1 connector, 1 source operation, 0 findings.
- [x] `go run ./cmd/connectorgen operation-evidence . --check` — current, 1,525 rows / 5 rollups / fixed-100 passed.
- [x] Current-head CI recovery: `go run ./cmd/connectorgen certification-subject` refreshed only `internal/connectors/certifications/current-subject.json` after Verify run `33042859184` found its declaration fingerprint stale; `go run ./cmd/connectorgen certification-subject --check`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, and `make connectorgen-certification-sweep` then passed.

## Repository checks

- [x] `gofmt -w` on changed Go files.
- [x] `go vet ./internal/connectors/engine ./internal/connectors/defs ./internal/connectors/commandrunner ./internal/app ./internal/cli`.
- [x] `make tidy-check lint smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` — pass.
- [x] `make docs-check` — connector docs validated.
- [x] `make github-parity-artifacts-check connector-canon-check` — pass.
- [x] Website data/type check not applicable: no CLI, docs, website, or generated website-data source changed; `make docs-check` covers connector-document parity.
- [x] `git diff --check` before implementation commits.
- [x] CI-boundary repair rerun: focused/full engine tests; `./internal/connectors/defs`, `./internal/connectors/commandrunner`, `./internal/app`, and `./internal/cli`; `go build ./cmd/pm`; `go vet` for changed/runtime packages; `go test ./cmd/connectorgen`; `connectorgen validate`, `surface-sync --check`, `declaration-admission`, and `operation-evidence --check`; `make docs-check`; and `make tidy-check lint smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` — pass.

## Review and delivery

- [x] Fresh-context Codex audit replaces unavailable Claude audit and records immutable source SHA.
- [x] Fresh-context Codex re-review records final code SHA `56808f8d2732f9d545982af4a1934a1e9b8dba5d`, six-lane results, and no production-code finding. Its low PR-body evidence wording finding was resolved without changing code.
- [x] Fresh-context exact-head Codex re-review of `b9b2478b3b2451d632d28b9aa138a170ad835110..fbc320592427cece41e6944bcf2f476ec4f62794` — pass, zero findings. It confirms the certification recovery changes only the generator-owned declaration fingerprint, preserves all six lanes and source/CLI/App surfaces, and leaves no additional technical gate before normal push.
- [x] REST API read-back for PR #4360 confirms base `main`, remote head `fbc320592427cece41e6944bcf2f476ec4f62794`, and the corrected PR body.
