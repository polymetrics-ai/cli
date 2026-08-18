# Verification — DB → API PostgreSQL → GitHub route R1

## Results

### Focused and generated checks — passed

- `go test -timeout 20m ./internal/app` — passed.
- `go test -timeout 20m ./internal/cli` — passed.
- `go test -timeout 20m ./internal/synctransport ./internal/connectors/native/postgres` — passed.
- `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` — passed; regenerated `docs/cli` from the changed help.
- `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m -run '^TestGoldenTranscripts$' ./internal/cli` — passed; generated transcript updated.
- `./pm help etl`; `./pm etl`; `./pm etl transport`; `./pm etl transport github-issue-label --help` — passed and show the PostgreSQL source/mode/mapping contract.
- `go vet ./...` — passed.
- `go build ./cmd/pm` — passed.
- `gofmt -w cmd internal` and `git diff --check` — passed.

### Docker PostgreSQL and GitHub evidence — passed

- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -count=1 -timeout 20m -tags databaseintegration -run '^TestPMBinaryExecutesPostgresWarehouseGitHubIssueLabels$' ./internal/cli` — passed. This is deterministic simulated-GitHub coverage only: live PostgreSQL + real binary/warehouse/acknowledgement/read-back/checkpoint through a faithful local HTTP boundary.
- With `POLYMETRICS_GITHUB_TOKEN` exported only by `gh auth token` command substitution, `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_GITHUB_ISSUE_LABEL_LIVE_PROOF=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -count=1 -v -timeout 20m -tags databaseintegration -run '^TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels$' ./internal/cli` — passed. The test uses a real Colima Docker PostgreSQL source, a freshly built `pm`, real GitHub HTTPS, durable receipt and checkpoint assertions, and a separate authenticated labels API client. It asserts issue #1 exactly `[pm-db-api-live-add]`, issue #2 exactly `[pm-db-api-live-set]`, and the keyed `set_issue_labels` replay leaves issue #2 unchanged.
- `gh-axi issue list -R karthik-sivadas/pm-parity-proof-db-to-api --state all --limit 10 --fields labels,url,updatedAt` — independently confirmed the retained private proof repository has `pm-db-api-live-add` on issue #1 and `pm-db-api-live-set` on issue #2. No credential value appears in this record.

### Repository verification gates — passed individually

- `make tidy-check`
- `make lint`
- `make docs-check-no-build`
- `make smoke-no-build`
- `make agent-contract-check`
- `make connectorgen-validate`
- `make connectorgen-surface-sync`
- `make github-parity-artifacts-check`
- `make connectorgen-certification-matrix`
- `make connector-boundary`
- `make connector-canon-check`
- `make release-workflow-check`

### Definition-owned admission correction — passed

- `go test -count=1 -timeout 20m -run '^(TestPreflightReturnsTypedDestinationSourceIneligibleErrorBeforeExecutorAccess|TestSyncTransportDescriptorResolvesDeclaredApplyStrategy)$' ./internal/synctransport ./internal/connectors` — passed. An unlisted source receives typed `DestinationSourceIneligibleError` before source read, warehouse stage, destination plan/apply, or checkpoint I/O.
- `go test -count=1 -timeout 20m -run '^(TestOpenSelectsPostgresIssueLabelDestinationTransport|TestIssueLabelTransportContractUsesDefinitionOwnedActionBindings|TestPostgresIssueLabelTransportRefusesBadInputsBeforeProviderWrite)$' ./internal/app` — passed.
- `go test -count=1 -timeout 20m ./internal/app ./internal/cli ./internal/connectors ./internal/connectors/engine ./internal/synctransport` — passed.
- `go run ./cmd/connectorgen validate` — passed: `552 connectors, 0 findings`.
- `make connector-boundary`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make docs-check-no-build`; `go vet ./...`; `go build ./cmd/pm`; the four ETL help/namespace probes; and `git diff --check` — passed.
- `pnpm --dir website run gen:docs` was run twice with a SHA-256 comparison of `website/lib/docs.generated.ts`; the repeat was byte-stable. `pnpm --dir website run typecheck` passed and `pnpm --dir website run lint` passed with 13 pre-existing warnings and no errors.

The complete `go test ./...` / aggregate `make verify` commands were deliberately not run as one process because this task runner applies a per-command wall limit and the repository contract directs agents to run the changed-package tests and all other `make verify` gates individually. All individually runnable gates above passed.

### Inline GSD verification and review

`verify-work` was executed as the documented manual GSD fallback against the acceptance table in `CONTEXT.md`: live proof, simulated boundary proof, typed refusal coverage, replay/resume/deletes edges, generated surface, and local gates are recorded above. `code-review` was completed again after the boundary finding: the destination definition owns source admission and bounded mappings, so no shared provider policy, generic writer, credential persistence, Arrow/full-overwrite change, or API-to-API quadrant edit remains in the diff.

### Delivery header base verification — passed

PR #4186 is open from `fm/cli-quadrant-db-to-api-postgres-github-r1`. The API query `gh-axi pr list -R polymetrics-ai/cli --state open --base integration/4015-mvp-flat-r1 --head fm/cli-quadrant-db-to-api-postgres-github-r1 --limit 10 --fields url` returned only that PR, confirming its required non-default base.
