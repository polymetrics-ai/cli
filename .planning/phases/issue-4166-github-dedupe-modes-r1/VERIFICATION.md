# Verification: GitHub dedupe modes r1

## Required checks

- [ ] Focused generator red/green test and matrix generation repeat-byte stability.
- [ ] Focused production-registry preflight and no-I/O-refusal tests.
- [ ] Focused warehouse-apply/dedupe and history-replay tests through the production command path.
- [ ] Fresh built `pm` happy, bad, and edge runs against a private retained GitHub repository, with an independent provider read-back.
- [ ] `go test -timeout 20m` for changed packages and `internal/cli`, separately.
- [ ] `go vet ./...`, `go build ./cmd/pm`, `git diff --check`.
- [ ] `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`, scoped `certification-matrix` generation/check, `make connector-boundary`.
- [ ] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make release-workflow-check`.
- [ ] `pnpm --dir website run gen:docs`, then repeat it and prove the second run is byte-stable.
- [ ] Runtime help/parity smoke: `pm connectors`, `pm help github`, and a GitHub command `--help`; record whether docs text changes are applicable.
- [ ] Inline `verify-work` and `code-review` evidence, with each actionable finding fixed or dispositioned.
- [ ] Direct PR opened and its API-reported base exactly equals `integration/4015-mvp-flat-r1`.

## Results

Pending implementation.
