# Verification Checklist — Issue #4290

- [x] Batch 4 source lock/map inventory equality is green (`materialize-parity-maps.mjs check batch4`).
- [x] Batch 5 source lock/map inventory equality is green (`materialize-parity-maps.mjs check batch5`).
- [x] Every map has valid six-class parity classification and DELETE coverage (Batch 4 and Batch 5 check assertions).
- [x] Every unauthored operation is `declaration-pending`, not `foundation-gap` (Batch 4 and Batch 5 check assertions); the only recorded gap is the reverse-ETL eligibility attribute `generic-typed-destination-executor`.
- [x] `go run ./cmd/connectorgen validate` is green (552 connectors, 0 findings; final run).
- [x] `go run ./cmd/connectorgen surface-sync --check` is green (552 connectors, 0 fields corrected; final run).
- [x] `make connector-boundary` is green (final detached run, exit 0; log: `connector-boundary-final.log`).
- [x] Focused tests are green: `go test -timeout 20m ./cmd/connectorgen`, `go test -timeout 20m ./internal/connectors/engine` (7.508s), and `go test -timeout 20m ./internal/cli`.
- [x] Repository generated/docs and compile checks are green: `go vet ./...`; `go build ./cmd/pm`; `make tidy-check`; `make lint` (0 issues); `make docs-check`; `make smoke-no-build`; `make agent-contract-check`; and `make release-workflow-check`.
- [x] PR #4295 base is `main`, read from the GitHub API with `gh api /repos/polymetrics-ai/cli/pulls/4295 --jq .base.ref`.
