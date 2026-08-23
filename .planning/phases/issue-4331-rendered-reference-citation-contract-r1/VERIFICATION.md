# Verification — Issue 4331

## Planned commands

- `go run ./cmd/connectorgen validate` — batch 6/7 red reproduction, then full contract validation.
- `go test -timeout 20m ./cmd/connectorgen -run 'Test.*Source.*(Import|Projection|Lock)'` — focused behavior and legacy regression coverage.
- `go test -timeout 20m ./cmd/connectorgen` — affected package suite.
- `go vet ./cmd/connectorgen`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`
- `go run ./cmd/agentcontractgen check`
- `git diff --check`

The full `go test ./...` and `make verify` commands are intentionally not run as one per-command operation because repository guidance says their duration exceeds the command window; CI carries the full suite. No `no-mistakes` command is permitted for this direct-PR delivery.

## Executed evidence

- `go run ./cmd/connectorgen validate` on the base checkout: **PASS** — `552 connector(s) checked, 0 findings`. Batch 6/7 locks are not on `main`, so this command cannot demonstrate their missing-contract failure.
- `git archive origin/fm/cli-map-batch67-r1 internal/connectors/defs | tar -x -C <temporary-worktree-dir>` then `go run ./cmd/connectorgen validate <temporary-worktree-dir>/internal/connectors/defs`: **RED / expected exit 1** — 20 batch source-projection lock parse failures, including `unknown field "source_kind"` for rendered-reference locks. The archive was removed immediately after the read-only validation; no connector lock was modified.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3(RenderedReferenceProjectsCapturedCitation|MixedOpenAPIAndRenderedReferenceKeepsOpenAPIProjectionBytes|RenderedReferenceRejectsUnverifiableEvidenceAndCitations|BundleRejectsArchiveHashMismatch|UnavailableSourceProjectsBlockingGap)$|^TestSourceImportRenderedReferenceKeepsSchemaOneAndTwoLocksValid$' -count=1`: **PASS** — focused red/green/refactor contract suite.
- `go test -timeout 20m ./cmd/connectorgen`: **PASS** — `ok polymetrics.ai/cmd/connectorgen 193.908s`.
- `go vet ./cmd/connectorgen && go build ./cmd/connectorgen`: **PASS**.
- `make tidy-check && make docs-check && make smoke-no-build && make agent-contract-check`: **PASS**.
- `make lint && make connectorgen-validate && make connectorgen-surface-sync && make connector-boundary && make release-workflow-check`: **PASS** — including a clean 552-connector boundary report and installed GitHub certification archive proof.
- `git diff --check`: **PASS**.

## Final contract refinements

- `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3BundleProjectsGzipCapture$' -count=1`: **RED then PASS** — the red test showed `application/x-gzip` was incorrectly rejected; ZIP, registered gzip, and the provider-published gzip alias now share the existing hash-verified bundle path.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3RenderedReferenceRejectsStandaloneOpenAPIDescription$' -count=1`: **RED then PASS** — a full OpenAPI 3.0.3 capture cannot now be mislabeled `rendered_reference`.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3RenderedReferenceProjectsYAMLPathFragment$' -count=1`: **PASS** — a structured `application/yaml` OpenAPI path fragment remains a rendered reference and projects with its citation.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3(RenderedReferenceProjectsCapturedCitation|RenderedReferenceProjectsYAMLPathFragment|RenderedReferenceRejectsStandaloneOpenAPIDescription|MixedOpenAPIAndRenderedReferenceKeepsOpenAPIProjectionBytes|RenderedReferenceRejectsUnverifiableEvidenceAndCitations|BundleRejectsArchiveHashMismatch|BundleProjectsGzipCapture|UnavailableSourceProjectsBlockingGap)$|^TestSourceImportRenderedReferenceKeepsSchemaOneAndTwoLocksValid$' -count=1`: **PASS**.
- `go test -timeout 20m ./cmd/connectorgen`: **PASS** — cached final run after the package suite completed.
- `go vet ./cmd/connectorgen && go build ./cmd/connectorgen && git diff --check`: **PASS**.
- `go run ./cmd/connectorgen validate`: **PASS** — `552 connector(s) checked, 0 findings`.
- `go run ./cmd/connectorgen surface-sync --check`: **PASS** — `552 connector(s) scanned`, zero changes.
- `make connector-boundary && make lint && make tidy-check && make docs-check && make smoke-no-build && make agent-contract-check && make release-workflow-check`: **PASS**. The release check was rerun in an attached terminal session and exited 0 with `installed GitHub certification archive proof passed`.
