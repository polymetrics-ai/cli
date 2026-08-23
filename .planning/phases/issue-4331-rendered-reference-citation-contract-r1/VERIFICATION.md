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

## Merged-main revalidation

After `origin/main` advanced with the cyclic source-import resolver, it was merged into this branch (no rebase). The combined head passed:

- `go test -timeout 20m ./cmd/connectorgen`: **PASS** — `ok polymetrics.ai/cmd/connectorgen 262.671s`.
- focused rendered-reference, standalone-vs-fragment, gzip-bundle, legacy, and OpenAPI byte-stability suite: **PASS**.
- `go vet ./cmd/connectorgen && go build ./cmd/connectorgen && git diff --check`: **PASS**.
- `go run ./cmd/connectorgen validate`: **PASS** — `552 connector(s) checked, 0 findings`.
- `go run ./cmd/connectorgen surface-sync --check`: **PASS** — 552 connectors scanned with zero changes.
- `go run ./cmd/agentcontractgen check`: **PASS**.

## Consumer-corpus preflight (read-only)

Using the final contract code at `be4ef9747`, source import was run read-only from this worktree against each consumer lane's `internal/connectors/defs` directory. No consumer worktree or lock was modified.

| Lane | Command result | Measured passing locks | Reason non-passing locks stop before representation validation |
| --- | --- | ---: | --- |
| Batch 2/3 (`.../46/cli`) | exit 1, 19 findings | 0 / 19 | Legacy/colliding `format` and `operation_counts` fields are rejected by strict v3 decoding. |
| Batch 6/7 (`.../12/cli`) | exit 1, 20 findings | 0 / 20 | Colliding `source_url` fields are rejected by strict v3 decoding. |
| Batch 8/9/10 (`.../54/cli`) | exit 1, 30 findings | 0 / 30 | Colliding `source_url` and `state` fields are rejected by strict v3 decoding. |
| Zoom (`.../1/cli`) | exit 1, 1 finding | 0 / 1 | Its incompatible top-level REST `retrieval` field is rejected by strict v3 decoding. |

These are migration-owned schema collisions, not missing foundation kinds: mainline v3 remains authoritative and must not accept those alternate field names. The contract already contains and tests the discovered consumer representations: JSON/YAML rendered captures and path fragments, ZIP/gzip bundles, explicit unavailable sources, required non-empty coverage confidence, and citation/evidence integrity.

## Post-mapping consumer dry-run (read-only)

The prior raw preflight intentionally stopped at names that the migration drops. The corrective dry-run used a temporary Go test (deleted before commit) to read each consumer lock, write only its decided v3 mapping to `t.TempDir`, and invoke the production `parseSourceImportLock` strict parser on that copy. It did not fetch any provider URL or write to a consumer worktree. Full `connectorgen validate` is not the right structural command for these copies because the consumer lanes have not yet produced their canonical source descriptors.

`go test -timeout 20m -run '^TestMigrationDryRunConsumerLocks$' -v ./cmd/connectorgen`: **PASS** — 47 / 70 mapped lock copies validate.

| Lane | Mapped copies passing | Remaining migration correction |
| --- | ---: | --- |
| Batch 2/3 | 12 / 19 | Five claimed OpenAPI documents have no valid recorded 3.0/3.1 pin (`amazon-sqs` service model `2012-11-05`, Gmail discovery `v1`, unpinned Google Ads discovery, Google Calendar discovery `v3`, and Slack Swagger `2.0`); `miro` has a path ending in `?`; `trello` repeats an operation identity. |
| Batch 6/7 | 19 / 20 | `iterable` repeats source operation identities. |
| Batch 8/9/10 | 15 / 30 | Fifteen machine-readable documents have no recorded 3.0/3.1 pin: `auth0`, `brex`, `calendly`, `coda`, `commercetools`, `datadog`, `dbt`, `docuseal`, `firehydrant`, `looker`, `metabase`, `mode`, `okta`, `pagerduty`, and `posthog`. |
| Zoom | 1 / 1 | None. |

The dry-run covers the 889 rendered-reference documents (including JSON/YAML and zero-operation navigation pages), the ZIP/gzip bundles, all three unavailable declarations, 63 n8n path fragments, and Zoom's 35 Next-data documents. It establishes no missing rendered/bundle/unavailable contract kind. The 23 remaining locks need migration-owned source corrections: a valid OpenAPI 3.0/3.1 provenance pin where `kind: openapi` is claimed, a valid route string, or unique source operation IDs. Those are existing v3 invariants intentionally preserved by this foundation, not compatibility names or a new weaker representation path.

## Final local regression after the dry-run refinement

The shared machine was under concurrent cold-cache load, so Go compilation was deliberately capped without changing the repository's 20-minute test timeout:

- `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen`: **PASS** — `ok polymetrics.ai/cmd/connectorgen 181.477s`.
- `GOFLAGS='-p=3' go vet ./cmd/connectorgen && GOFLAGS='-p=3' go build ./cmd/connectorgen && git diff --check`: **PASS**.
- `GOFLAGS='-p=3' go run ./cmd/connectorgen validate`: **PASS** — 552 connectors, zero findings.
- `GOFLAGS='-p=3' go run ./cmd/connectorgen surface-sync --check`: **PASS** — 552 connectors, zero changes.
- `GOFLAGS='-p=3' go run ./cmd/agentcontractgen check`, `make connector-boundary`, `make lint`, `make tidy-check`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, and `make release-workflow-check`: **PASS**.
