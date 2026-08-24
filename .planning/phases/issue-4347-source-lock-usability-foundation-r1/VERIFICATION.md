# Verification checklist — issue 4347 source-lock usability

- [ ] Focused red tests recorded for missing form pin, canonical reserialisation, 403, and undersize.
- [ ] `go test -timeout 10m ./cmd/connectorgen` passes after implementation.
- [ ] `go vet ./cmd/connectorgen` passes.
- [ ] Build `connectorgen` and run `source-retain` sequentially for fastly, github, hubspot, pipedrive, shipstation, squarespace, woocommerce, and zendesk-support; each manifest and digest-addressed file is inspected without modifying its lock.
- [ ] `go run ./cmd/connectorgen surface-sync --check` passes or an exact unrelated pre-existing failure is recorded.
- [ ] Relevant independent `make verify` gates (`tidy-check`, `lint`, `docs-check`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`) pass or are recorded with exact reason. Do not run aggregate `go test ./...`.
- [ ] `scripts/gsd prompt verify-work 4347` and `scripts/gsd prompt code-review 4347` inline outcomes are recorded; review findings are dispositioned.
- [ ] PR base is read back through the GitHub API and exactly equals `main`.
