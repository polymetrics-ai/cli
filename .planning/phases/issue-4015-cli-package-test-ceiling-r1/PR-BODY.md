Refs #4015

## Intent

Remove `internal/cli`'s collision with Go's 20-minute per-test-binary timeout without weakening a single test.

## What changed

- Replaced 18 repeated `go build -o <test-tempdir>/pm ./cmd/pm` calls with one lazy, package-scoped, read-only `pm` fixture protected by `sync.Once`.
- Kept every real-binary test's independent project root, subprocess entry point, provider fixture, approval flow, readback, and assertion unchanged.
- Made `TestMain` remove the package-owned temporary fixture directory after `m.Run()` while preserving its existing invocation-budget failure semantics.
- Strengthened the existing real-binary lifecycle proof with an identity assertion: it first failed on two distinct executable paths, then passed when both calls shared one path.

## Measurement and completeness

| Evidence | Before | After |
| --- | ---: | ---: |
| Diagnostic `internal/cli` package time | 623.128s | 532.694s |
| Diagnostic wall time | 627.73s | 537.29s |
| Final rebased `make verify` package time | — | 847.535s |
| Normal 1200s timeout headroom on final branch | 19.018s on the documented base CI result | 352.465s / 29.4% locally |
| Runnable top-level test names | 263 | 263 |

The normalized before/after inventories have the identical SHA-256 `198d206817b81ff1f9480421f2aad978958f275bce83c1e105eb28023a4e61b6`. `TestMain` is the 264th declared `Test…` function and is intentionally absent from `go test -list`. No test was renamed, deleted, skipped, shortened, build-tagged out, or mode-excluded.

The original full timing trace showed a long tail, not an isolated defect: the named Bahmni matrix test completed in 39.78s and is unchanged. The chosen mechanism follows evidence: package-wide parallelism is unsafe without a broad rewrite because current tests use process-global `t.Setenv` and `t.Chdir`; sharding would retain repeated linking and add partition-completeness risk; moving tests or raising `-timeout` would not fix the directly measured cause.

## Red / green

- **Red:** `go test -timeout 20m ./internal/cli -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$'` failed because two fixture requests returned `…/001/pm` and `…/002/pm`.
- **Green:** the same test passed with a stable shared path and its original real-binary lifecycle assertions intact.
- **Concurrency:** `go test -race -timeout 20m ./internal/cli -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$'` passed (61.353s).

## Verification

- Final rebased `make verify` passed. Its unchanged `go test -timeout 20m ./...` stage reported `internal/cli` at 847.535s; generated-file, smoke, lint, connector boundary, canonical, and release checks also passed.
- `go test -timeout 20m ./cmd/connectorgen` passed (94.783s).
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make docs-check-no-build`, `make smoke-no-build`, `make lint`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connectorgen-certification-sweep`, `make connector-boundary`, `make connector-canon-check`, and `make release-workflow-check` passed.
- After rebasing onto `db494bc8f`, both `pnpm --dir website run gen:docs` and `pnpm --dir website run gen:website-data` ran twice; `git diff --exit-code -- website` confirmed byte stability.
- `security/snyk` was not run locally because it fails identically on the base branch, as documented in the task brief.

## Delivery record

- Base: `integration/4015-mvp-flat-r1` (`db494bc8f` at the final rebase)
- GSD: resolved and executed `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` inline. This is the documented non-Pi/manual fallback; no GSD role agents were spawned.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-benchmark`, `golang-performance`, `golang-troubleshooting`, `golang-safety`, and `golang-lint`.
- CLI help/manual/website behavior: not applicable; the CLI surface does not change. Website generator stability was nevertheless verified after the rebase.
- Safety: no credentials, production writes, reverse-ETL execution, dependencies, or connector-specific shared-Go identifiers were added.
- Commits: `2de30451a`, `e6d755c9a`.

## Automated review

Primary route: `claude_auto` pending for this trusted-author, non-draft stacked PR. Copilot is fallback-only if Claude cannot provide review coverage. No automated findings are available yet.
