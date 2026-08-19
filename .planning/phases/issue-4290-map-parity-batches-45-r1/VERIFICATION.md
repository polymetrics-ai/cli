# Verification Checklist — Issue #4290

- [x] Every Batch 4 source lock/map inventory is regenerated from a provider-complete source or carries an explicit dynamic/unavailable basis (`materialize-parity-maps.mjs check batch4`).
- [x] Every Batch 5 source lock/map inventory is regenerated from a provider-complete source or carries an explicit dynamic/unavailable basis (`materialize-parity-maps.mjs check batch5`).
- [x] Per connector source report records old `api_surface` operation count, regenerated count, and source basis (`SOURCE-INVENTORY-REPORT.md`).
- [x] Every map has valid six-class parity classification and DELETE coverage (Batch 4 and Batch 5 check assertions).
- [x] Every unauthored operation is `declaration-pending`, not `foundation-gap` (Batch 4 and Batch 5 check assertions); the only recorded gap is the reverse-ETL eligibility attribute `generic-typed-destination-executor`.
- [x] Buildkite v1 sensitive-operation provenance recovery: targeted `connectorgen validate internal/connectors/defs/buildkite --json` first reported 100 missing provider citations; after rematerialization through the pinned retrieval-artifact fallback it reports zero findings, and the ledger has 100 cited operations with zero missing sensitive citations.
- [x] Seven-surface audit: `materialize-parity-maps.mjs seven-surface-ledger --check` asserts exactly 20 assigned rows; every existing typed write has a single explicit reverse-ETL eligibility disposition.
- [ ] Foundation dependency: before final push, fetch and merge the current `origin/fm/cli-reverse-etl-destination-r1`, prove it is an ancestor, and exercise the real installed App/CLI generic destination dispatch. At SHA `c6f03c937`, that dispatch remains pending; no connector claims deployment.
- [x] `go run ./cmd/connectorgen validate` is green (552 connectors, 0 findings; final run).
- [x] `go run ./cmd/connectorgen surface-sync --check` is green (552 connectors, 0 fields corrected; final run).
- [x] `make connector-boundary` is green (final detached run, exit 0; log: `connector-boundary-final.log`).
- [x] Focused tests are green: `go test -timeout 20m ./cmd/connectorgen`, `go test -timeout 20m ./internal/connectors/engine`, and `make connector-runtime-preflight`.
- [ ] Local limitation: `go test -timeout 20m ./internal/cli` timed out after 20m while concurrent full CLI suites were active. It timed out during `TestScheduleCLI_Create_InvalidName` while loading the all-connector registry (`engine.loadOperationEndpointLedgers`); no assertion failure was reported. Log: `/tmp/cli-map-batch45-cli-test.log`. It is left to CI rather than retried concurrently.
- [x] Repository generated/docs and compile checks are green: `go vet ./...`; `go build ./cmd/pm`; `make tidy-check`; `make lint` (0 issues); `make docs-check`; `make smoke-no-build`; `make agent-contract-check`; and `make release-workflow-check`.
- [x] PR #4295 base is `main`, read from the GitHub API with `gh api /repos/polymetrics-ai/cli/pulls/4295 --jq .base.ref`.
