# Verification checklist — issue 4347 source-lock usability

- [x] Focused red tests recorded for missing form pin, canonical reserialisation, 403, and undersize.
- [x] Focused red/green test recorded for the distinct 403/TLS `BOT-BLOCK` verdict.
- [x] `go test -timeout 10m ./cmd/connectorgen -run '^TestSourceRetain'` passes after implementation. The broader `go test -timeout 20m ./cmd/connectorgen` was stopped locally after its unrelated `TestCertificationMatrix...` fixture started `certification-matrix --all` with a fresh `GOCACHE`; it produced no source-lock failure. CI carries that broader package path.
- [x] `go vet ./cmd/connectorgen` passes.
- [x] Built `connectorgen` and ran `source-retain` sequentially for fastly, github, hubspot, pipedrive, shipstation, squarespace, woocommerce, and zendesk-support; all commands exit 0 and pre/post SHA-256 values prove their locks were unchanged.
- [x] `go run ./cmd/connectorgen surface-sync --check` passes.
- [x] Relevant independent `make verify` gates pass: `tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`. Do not run aggregate `go test ./...`.
- [x] `scripts/gsd prompt verify-work 4347` and `scripts/gsd prompt code-review 4347` ran as the recorded inline/manual fallback; `REVIEW.md` records the review disposition.
- [ ] PR base is read back through the GitHub API and exactly equals `main` after the follow-up push.
