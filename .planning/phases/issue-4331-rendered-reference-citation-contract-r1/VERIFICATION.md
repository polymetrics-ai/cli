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

`GOFLAGS='-p=3' go test -timeout 20m -run '^TestMigrationDryRunConsumerLocks$' -count=1 -v ./cmd/connectorgen`: **PASS** — 49 / 70 mapped lock copies validate on the final contract revision. The temporary test and all its copied locks were deleted before commit.

| Lane | Mapped copies passing | Remaining migration correction |
| --- | ---: | --- |
| Batch 2/3 | 15 / 19 | Amazon SQS is an explicitly excluded native AWS Query contract gap; Google Ads has rendered-reference route aliases to collapse; `miro` has a path ending in `?`; `trello` repeats an operation identity. Gmail and Google Calendar pass when correctly mapped as rendered references. Slack Swagger `2.0` validates with its explicit Swagger form pin. |
| Batch 6/7 | 18 / 20 | `iterable` repeats source operation identities; one Outreach operation cites `developers.outreach.io` while the sole captured document publishes under `api.outreach.io`. |
| Batch 8/9/10 | 15 / 30 | Fifteen machine-readable documents have no recorded 3.0/3.1 pin: `auth0`, `brex`, `calendly`, `coda`, `commercetools`, `datadog`, `dbt`, `docuseal`, `firehydrant`, `looker`, `metabase`, `mode`, `okta`, `pagerduty`, and `posthog`. |
| Zoom | 1 / 1 | None. |

The dry-run covers the 889 rendered-reference documents (including JSON/YAML and zero-operation navigation pages), the ZIP/gzip bundles, all three unavailable declarations, 63 n8n path fragments, and Zoom's 35 Next-data documents. Slack exposed the sole former contract gap and is now supported by the existing Swagger 2.0 parser with a strict form pin. The final full-corpus result is 49 PASS, one excluded native/AWS-Query contract gap (Amazon SQS), 17 mapping gaps, and three source defects. Across the 69 consumers in this PR's scope, there are **zero contract gaps**.

## Final local regression after the dry-run refinement

The shared machine was under concurrent cold-cache load, so Go compilation was deliberately capped without changing the repository's 20-minute test timeout:

- `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen`: **PASS** — `ok polymetrics.ai/cmd/connectorgen 181.477s`.
- `GOFLAGS='-p=3' go vet ./cmd/connectorgen && GOFLAGS='-p=3' go build ./cmd/connectorgen && git diff --check`: **PASS**.
- `GOFLAGS='-p=3' go run ./cmd/connectorgen validate`: **PASS** — 552 connectors, zero findings.
- `GOFLAGS='-p=3' go run ./cmd/connectorgen surface-sync --check`: **PASS** — 552 connectors, zero changes.
- `GOFLAGS='-p=3' go run ./cmd/agentcontractgen check`, `make connector-boundary`, `make lint`, `make tidy-check`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, and `make release-workflow-check`: **PASS**.

## Swagger 2.0 contract-gap correction

The consumer dry-run classified Slack's `swagger: "2.0"` source as a genuine contract gap: it is a parseable standalone source description, so the rendered-reference discriminator correctly rejects it, while the v3 OpenAPI inventory intentionally permits only OpenAPI 3.0/3.1. The source importer already has a full Swagger 2.0 parser; this refinement records its existing `artifact.swagger` form pin on an OpenAPI-kind document and excludes it from `openapi_versions`. No OpenAPI 3.0/3.1 validation was relaxed.

- `GOFLAGS='-p=3' go test -timeout 20m -run '^TestSourceImportVersion3SwaggerTwoProjectsWithoutOpenAPIVersionInventory$' -count=1 ./cmd/connectorgen`: **RED then PASS**.
- `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen`: **PASS** — `ok polymetrics.ai/cmd/connectorgen 192.468s`.
- `GOFLAGS='-p=3' go vet ./cmd/connectorgen && GOFLAGS='-p=3' go build ./cmd/connectorgen && git diff --check`: **PASS**.

## Final consumer failure classification and batch 6/7 delta

The final migration dry-run intentionally wrote every mapped lock only below
`t.TempDir` and parsed it through the production strict importer. A fresh
full-corpus rerun at `429ea1bf1` corrects the earlier masked Amazon SQS result:
49 copies pass and 21 fail. Consumer paths and line numbers name the read-only
input evidence; the final column names the production contract check which
refused the mapped copy.

### Contract gap — Amazon SQS, explicitly excluded from this PR

Slack was the one genuine REST-document contract gap discovered by the first
dry-run: its standalone `swagger: "2.0"` source was neither a rendered reference
nor a 3.0/3.1 OpenAPI document. Commit `9df9e058e` resolves that gap by retaining
the existing strict Swagger parser with an explicit `artifact.swagger` form pin.

Amazon SQS is not a REST-route mapping gap. Its Botocore service-model capture
at `.../46/cli/internal/connectors/defs/amazon-sqs/sources/amazon-sqs-operation-source-lock.json:6-10`
correctly maps as a rendered reference, but then its intentional native operation
identity `POST SQS.<Action>` (for example `SQS.AddPermission` at `:16-23`) is
rejected by `cmd/connectorgen/sourceimport.go:800-801`: v3 REST operations must
have slash-prefixed connector-relative paths. Rewriting it as a REST path would
misrepresent the native AWS Query runtime identity. Amazon SQS is therefore
removed from this PR's 69-consumer gate and is owned by
`cli-native-query-source-operation-contract-r1`; this PR does not implement that
native/action-addressed representation. Consequently, there are **zero contract
gaps across the 69 in-scope consumers**.

### Mapping gaps — 17

| Connector | Lane | Exact unmapped shape and read-only evidence | Production refusal |
| --- | --- | --- | --- |
| google-ads | 2/3 | Google Discovery JSON at `.../46/cli/internal/connectors/defs/google-ads/sources/google-ads-operation-source-lock.json:6-13` correctly maps as a rendered reference, which exposes repeated provider route aliases at `:873-890`, `:1033-1069`, and `:1433-1449`. | `cmd/connectorgen/sourceimport.go:809-810` — the mapper must collapse or otherwise resolve the duplicate routes; it must not mislabel Discovery JSON as OpenAPI. |
| outreach | 6/7 | The source artifact publishes under `https://api.outreach.io` at `.../12/cli/internal/connectors/defs/outreach/sources/outreach-operation-source-lock.json:6`, but `outreach.rest.delete.ApiV2CustomObjectsObjectNameId8` cites `https://developers.outreach.io/api/custom-objects` at `:93-100`. One rendered-reference document cannot legitimately vouch for both origins. | `cmd/connectorgen/sourceimport.go:804-805` — citation must be a well-formed absolute URL under its document's published-source origin. The mapping needs a separately captured `developers.outreach.io` document (or a retraced same-origin source), not a relaxed origin check. |
| auth0 | 8/9/10 | The migrated `documents` entry has content/hash/bytes but no retained verified form/version pin: `.../54/cli/internal/connectors/defs/auth0/sources/auth0-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| brex | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/brex/sources/brex-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| calendly | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/calendly/sources/calendly-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| coda | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/coda/sources/coda-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| commercetools | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/commercetools/sources/commercetools-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| datadog | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/datadog/sources/datadog-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| dbt | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/dbt/sources/dbt-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| docuseal | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/docuseal/sources/docuseal-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| firehydrant | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/firehydrant/sources/firehydrant-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| looker | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/looker/sources/looker-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| metabase | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/metabase/sources/metabase-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| mode | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/mode/sources/mode-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| okta | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/okta/sources/okta-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| pagerduty | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/pagerduty/sources/pagerduty-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |
| posthog | 8/9/10 | The `machine-readable-spec` document retains no verified form/version pin at `.../54/cli/internal/connectors/defs/posthog/sources/posthog-operation-source-lock.json:6-16`. | `cmd/connectorgen/sourceimport.go:764-765` — the mapped OpenAPI branch lacks its required version. |

The final 15 rows are one systematic migration omission, not 15 new contract
shapes: the mapper must retain the captured document's verified form/version
pin — `artifact.openapi` at 3.0/3.1 for an OpenAPI document, or
`artifact.swagger` for Swagger 2.0 — before choosing the document kind. The
contract must not infer either form or version from a URL, content type, or
operation list.

### Source defects — 3

| Connector | Lane | Exact malformed consumer evidence | Production refusal |
| --- | --- | --- | --- |
| miro | 2/3 | Path `/v2/boards/{board_id}/groups/{group_id}?` is malformed at `.../46/cli/internal/connectors/defs/miro/sources/miro-operation-source-lock.json:1427-1434`. | `cmd/connectorgen/sourceimport.go:800-801` — invalid REST operation path. |
| trello | 2/3 | `trello.rest.put-members-id-notificationChannelSettings-channel-blockedKeys` is duplicated at `.../46/cli/internal/connectors/defs/trello/sources/trello-operation-source-lock.json:2019`, `:2039`, and `:2049`. | `cmd/connectorgen/sourceimport.go:797-798` — duplicate REST operation identity. |
| iterable | 6/7 | `iterable.rest.delete.delete` is duplicated at `.../12/cli/internal/connectors/defs/iterable/sources/iterable-operation-source-lock.json:53` and `:63` (with further repetitions later in the source inventory). | `cmd/connectorgen/sourceimport.go:797-798` — duplicate REST operation identity. |

### Batch 6/7 19-to-18 delta: mapper correction, not a contract regression

The earlier `9f5cd8672` dry-run reported 19/20 for batch 6/7 because its
temporary copy mapper transformed each operation twice. On the first conversion
it copied an operation's legacy `source_url` into the citation. On the second,
that legacy field was already absent, so the mapper incorrectly substituted the
document's `published_source.source_url` for every citation. That substitution
made Outreach's `developers.outreach.io` citation appear to originate at
`api.outreach.io` and hid the mapping gap.

The final copy mapper preserves the original per-operation citation, so the
production same-origin check correctly rejects the operation at Outreach line
100. The source-code diff from `9f5cd8672` to `9df9e058e` is limited to Swagger
2.0 `artifact.swagger` handling in `sourceimport.go:756-765` and its projected
form in `sourceprojection.go:1869-1877`; neither executes for batch 6/7's
rendered-reference documents. That Swagger correction gained Slack in batch
2/3, while the mapper correction exposed Outreach in batch 6/7. The fresh
full-corpus rerun additionally maps Gmail and Google Calendar as rendered
references, raising the overall result to 49/70. Outreach remains deliberately
classified as a **mapping gap**, not a reason to weaken the contract's
cross-origin citation protection.
