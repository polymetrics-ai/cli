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

- [ ] F1 red and green: strict v3 document-owned operation-evidence identity preserves six lanes and declared/deferred rows.
- [ ] F2 red and green: generic rendered publication citation rejects unless fragment or verified capture extraction binding is present. Read-only Batch 6–7 impact at `origin/fm/cli-map-batch67-r1` / `18248d233e6abd9d7ec03075a225cf35ee2f5399`: 861 generic citation rows in eight connectors are intentionally not admitted until their lock owners add a fragment or binding.
- [ ] F3 red and green: HTTP MIME/body evidence rejects plausible login and `Error 503` pages plus bad MIME as wrong-source before drift, without rejecting legitimate documentation HTML.
- [ ] `go test -timeout 20m ./cmd/connectorgen` passes after the repair, or any unrelated local limitation is recorded exactly.
- [ ] `go vet ./cmd/connectorgen` passes after the repair.
- [ ] `go run ./cmd/connectorgen operation-evidence --check` and `go run ./cmd/connectorgen certification-subject --check` pass from a clean tracked worktree; regenerated checked-in provenance is byte-stable on a second run.
- [ ] Applicable non-aggregate repository gates pass separately with explicit timeouts, including source import/projection and generated snapshot checks.
- [ ] PR #4350 is pushed, its API-reported base is `main`, and Firstmate is asked for a fresh independent audit. No merge is performed.
