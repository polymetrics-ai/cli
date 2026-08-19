# Verification: GitHub dedupe modes r1

## Required checks

- [x] Focused generator red/green test and matrix generation repeat-byte stability.
- [x] Focused production-registry preflight and no-I/O-refusal tests.
- [x] Focused warehouse-apply/dedupe and history-replay tests through the production command path.
- [x] Fresh built `pm` happy, bad, and edge runs against retained private GitHub repository `karthik-sivadas/pm-truth-github-dedupe-modes-build-r1`, with independent `pm github pr list` and warehouse read-back.
- [x] `go test -timeout 20m` for changed packages and `internal/cli`, separately.
- [x] `go vet ./...`, `go build ./cmd/pm`, `git diff --check`.
- [x] `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`, certification-matrix generation/check, `go run ./cmd/connectorgen boundary . --json`.
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`, `make agent-contract-check`, `make release-workflow-check`.
- [x] `pnpm --dir website run gen:docs`, then repeat it and prove the second run is byte-stable.
- [x] CI help gap: unchanged `TestETLHelpListsAllSyncModes` now finds the complete history-mode description in rendered `pm help etl`; regenerated `docs/cli/etl.md` and golden transcripts are current.
- [x] Runtime help/parity smoke: `pm connectors inspect github --json`, `pm help etl`, `pm connections`, and `pm github pr list --help`; help/manual/website text changed and generated outputs were refreshed.
- [x] Inline `verify-work` and `code-review` evidence, with each actionable finding fixed or dispositioned (manual fallback recorded in `UAT.md` and `REVIEW.md`).
- [x] Direct PR [#4188](https://github.com/polymetrics-ai/cli/pull/4188) opened; `gh-axi pr list --head fm/cli-truth-github-dedupe-modes-build-r1 --base integration/4015-mvp-flat-r1` returned exactly that PR.

## Results

All required local verification and direct-PR base checks passed. The one CI
failure was closed with a regenerated help/manual/golden change, then the full
scoped local gate set and repeat website generator were rerun successfully.
