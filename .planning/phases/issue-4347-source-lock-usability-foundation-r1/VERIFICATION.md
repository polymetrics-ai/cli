# Verification checklist — issue 4347 source-lock usability

- [x] Focused red tests recorded for missing form pin, canonical reserialisation, 403, and undersize.
- [x] Focused red/green test recorded for the distinct 403/TLS `BOT-BLOCK` verdict.
- [x] CI failure verdict: clean `060bb7864` passes `go test -timeout 10m ./cmd/connectorgen -run '^TestSourceImportRetainedArtifactRejectsMissingAndMismatchedCopies$'`; the branch-owned root fix passes that test plus `TestSourceRetainReportsBotBlockBeforeWrongSourceOrDrift`.
- [x] `go test -timeout 10m ./cmd/connectorgen -run '^TestSourceRetain'` passes after implementation. The broader `go test -timeout 20m ./cmd/connectorgen` was stopped locally after its unrelated `TestCertificationMatrix...` fixture started `certification-matrix --all` with a fresh `GOCACHE`; it produced no source-lock failure. CI carries that broader package path.
- [x] `go vet ./cmd/connectorgen` passes.
- [x] Built `connectorgen` and ran `source-retain` sequentially for fastly, github, hubspot, pipedrive, shipstation, squarespace, woocommerce, and zendesk-support; all commands exit 0 and pre/post SHA-256 values prove their locks were unchanged.
- [x] `go run ./cmd/connectorgen surface-sync --check` passes.
- [x] Relevant independent `make verify` gates pass: `tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`. Do not run aggregate `go test ./...`.
- [x] `scripts/gsd prompt verify-work 4347` and `scripts/gsd prompt code-review 4347` ran as the recorded inline/manual fallback; `REVIEW.md` records the review disposition.
- [ ] PR base is read back through the GitHub API and exactly equals `main` after the follow-up push.

## Independent audit R1 gap closure — pending

- [x] F1 red and green: strict v3 document-owned operation-evidence identity preserves six lanes and declared/deferred rows. Exact focused green command passed in 2.822s on 2026-08-26.
- [x] F2 red and green: generic rendered publication citation rejects unless fragment or verified capture extraction binding is present. Read-only Batch 6–7 impact at `origin/fm/cli-map-batch67-r1` / `18248d233e6abd9d7ec03075a225cf35ee2f5399`: 861 generic citation rows in eight connectors are intentionally not admitted until their lock owners add a fragment or binding.
- [x] F3 red and green: HTTP MIME/body evidence rejects plausible login and `Error 503` pages plus invalid/bad MIME as wrong-source before drift, without rejecting legitimate documentation HTML.
- [x] `go test -timeout 20m ./cmd/connectorgen` passes after the final repair (158.269s on 2026-08-26).
- [x] `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, `make tidy-check`, and `make docs-check-no-build` pass after the repair.
- [x] Clean tracked archive at `9e1bfdb9b21ab346f84537bfb094a22782b0d5d5` passed `agentcontractgen check`, `connectorgen validate`, `surface-sync --check`, `operation-evidence --check` (1,525 rows; fixed-100 passed), certification subject/matrix/candidates/sweep checks, and `connectorgen boundary . --json`. Its temporary archive was deleted after the checks; it excluded the preserved live-retention artifacts.
- [x] `make lint` passes after the static-analysis repair. No aggregate `go test ./...` was run.
- [ ] PR #4350 is pushed, its API-reported base is `main`, and Firstmate is asked for a fresh independent audit. No merge is performed.
