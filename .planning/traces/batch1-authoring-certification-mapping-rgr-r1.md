# Batch 1 authoring admission / certification separation trace

## Task Delivery Header

- Issue: Refs the active Batch 1 source-lock integration; the canonical integration owner retains the issue and pull-request linkage.
- Base branch: `origin/fm/cli-top100-declaration-batch-r1-source-locks-r1`.
- Merges into: the existing Batch 1 branch, then `main` only through its canonical integration path.
- Delivery: Two ordered local commits for review; no push, merge, source-lock update, or runtime-foundation change in this task.
- Working branch: `fm/cli-batch1-authoring-rgr-r1`.
- Exact base commit: `7212f14bf8b602c317f30c6e0addcfb6655d88c4`.
- Exact fixed implementation commit: `90265feba13a6249de88c2176097d13007db43d4`.
- Task: Make source-backed declaration admission and mapping projection independent of certification-overlay availability, certification credentials, and retention/hash representation while preserving strict source-import and certification validation.
- Verification: Focused Red/Green tests, the complete `cmd/connectorgen` package test, authoring commands over checked-in connector definitions, Go formatting/vetting, JSON validation, and repository contract checks.

## Scope boundary

This is an authoring-only `cmd/connectorgen` change. It changes mapping admission,
source projection, surface sync, their tests, and the authoring/Atlas contract.
It does not change any connector source-lock byte, connector definition, generic
runtime engine, transport, warehouse behavior, credential loader, or
certification evaluator. Runtime and certification continue to load the full
declarative bundle. Strict source import continues to validate retained bytes,
byte counts, and digests.

The implementation commit owns exactly these files:

- `cmd/connectorgen/authoringfs.go`
- `cmd/connectorgen/declarationadmission.go`
- `cmd/connectorgen/declarationadmission_test.go`
- `cmd/connectorgen/sourceprojection.go`
- `cmd/connectorgen/sourceprojection_test.go`
- `cmd/connectorgen/surfacesync.go`
- `cmd/connectorgen/surfacesync_test.go`
- `docs/connector-canon/DECLARATION-ADMISSION.md`
- `docs/connector-canon/IMPLEMENTATION-PROCEDURE.md`
- `docs/connector-canon/foundations/catalog.json`

This trace is the only file in the second commit. Clean integration must preserve
both commits in order so the implementation SHA remains the immutable fixed
reference recorded here.

## Manual GSD evidence

This bounded Red-Green-Refactor slice was executed within the already active
Batch 1 integration, with disjoint ownership from the source-build and F22
workers. No additional lifecycle agent was spawned. The project-local GSD
adapter was checked before delivery:

- `scripts/gsd doctor` passed all repository, docs, registry, lock, Pi adapter,
  prompt, and command-count checks (`commands=69`).
- `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`,
  `verify-work`, and `code-review` each resolved the checked-in GSD command
  registry, upstream lock, and official command documentation.
- `go run ./cmd/agentcontractgen check` passed: the canonical contract and all
  registered projections are current.

## Red

The initial focused command was:

```text
go test -timeout 10m ./cmd/connectorgen -run 'TestDeclarationAdmissionMappingDoesNotRequireCertificationOverlay|TestSyncRuntimeOperationEndpointLedgerCreatesCompactProjection|TestSourceProjectionRestoresRequiredPathFlagForSourceBoundDirectRead|TestSourceProjectionRestoresReachableSourceBoundRead|TestSourceProjectionMappingIgnoresRetentionAndEmbeddedSourceOperation'
```

It failed for all five missing behaviors:

1. Declaration admission attempted to parse a malformed
   `certification.json` through `engine.Load`.
2. Endpoint-ledger sync attempted the same certification parse.
3. Source projection rejected the mapping lock's embedded
   `source_operation` because it used the strict retention parser.
4. Required direct-read path flags were not restored without a certification
   cohort.
5. Projection execution-surface discovery attempted to parse the malformed
   certification overlay.

The citation-only compatibility safety Red was:

```text
go test -timeout 3m ./cmd/connectorgen -run '^TestSourceReferenceSurfaceSyncRejectsTamperedDescriptor$' -count=1
```

It failed because the first mapping-only parser revision did not yet account
for legacy `operations_found` or schema-v3 `source_reference`. That test drove
the compatibility correction without weakening the closed unavailable-source
contract.

## Green

The final focused command was:

```text
go test -timeout 8m ./cmd/connectorgen -run 'TestDeclarationAdmissionMappingDoesNotRequireCertificationOverlay|TestSyncRuntimeOperationEndpointLedgerCreatesCompactProjection|TestSourceProjectionRestoresRequiredPathFlagForSourceBoundDirectRead|TestSourceProjectionRestoresReachableSourceBoundRead|TestSourceProjectionMappingIgnoresRetentionAndEmbeddedSourceOperation|TestSourceProjectionSourceReferenceIgnoresRetentionButPreservesClosedGap|TestSourceReferenceSurfaceSyncRejectsTamperedDescriptor|TestSourceReferenceSurfaceSyncRejectsMarkerBypass|TestSourceProjection_MissingOperationOrFieldFailsValidateAndSurfaceCheck|TestSurfaceSyncAcceptsSchema3SourceDescriptor' -count=1
```

Result: `ok polymetrics.ai/cmd/connectorgen 6.072s`.

Additional Green checks:

| Command | Result |
| --- | --- |
| `go run ./cmd/connectorgen declaration-admission --json` | Passed with no findings: `connectors_checked=1`, `source_operations=1`. |
| `go run ./cmd/connectorgen certification-matrix --check` | Passed: `connectors=3 capability_complete=0 certified=0`; certification remains a separate proof surface. |
| `go vet ./cmd/connectorgen` | Passed. |
| `gofmt -d` over all changed Go files | Empty output. |
| `jq empty docs/connector-canon/foundations/catalog.schema.json docs/connector-canon/foundations/catalog.json` | Passed. |
| Atlas stable-ID uniqueness query | Passed. |
| `git diff --check` | Passed. |

`go run ./cmd/connectorgen surface-sync --check` reached the exact known
pre-integration source-build state: Asana would update one write and one CLI
surface, then Bitbucket failed because its canonical source descriptor is
missing. This task neither owns nor hides those generated source-build changes.

## Complete-package result and baseline separation

The fixed-tree package command was:

```text
go test -timeout 20m ./cmd/connectorgen -count=1
```

It ran for `320.531s`. Every failure introduced during this RGR cycle was fixed.
The remaining failures were:

- `TestOperationEvidenceFixed100RejectsEveryRegression`
- `TestOperationEvidenceCheckRunsFixed100Gate`
- `TestRetainedAsanaSourceImportRejectsReadProjectionDrift`
- `TestRetainedAsanaSourceImportSelectsSourceBackedFanOutETLStreams`
- `TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation`

The first pair reports the pre-existing Asana ETL classification regression for
`asana.rest.getCustomFieldsForWorkspace`. The retained-import tests use the
strict source-import parser and reject the already-enriched
`source_operation` field; that parser is deliberately outside this mapping-only
change.

A clean detached worktree at exact base
`7212f14bf8b602c317f30c6e0addcfb6655d88c4` reproduced the direct fixed-100
failure and all three retained-import failures with:

```text
go test -timeout 8m ./cmd/connectorgen -run 'TestRetainedAsanaSourceImportRejectsReadProjectionDrift|TestRetainedAsanaSourceImportSelectsSourceBackedFanOutETLStreams|TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation|TestOperationEvidenceFixed100RejectsEveryRegression' -count=1
```

The `TestOperationEvidenceCheckRunsFixed100Gate` wrapper has the same underlying
fixed-100 error. These are baseline/integration-chain failures, not regressions
from commit `90265feba13a6249de88c2176097d13007db43d4`.

## Evidence classification

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Mapping admission does not require a valid certification overlay | fake | Hermetic fixtures deliberately omit or corrupt `certification.json` and exercise the real authoring load path; provider credentials and network I/O are outside this task. |
| Mapping projection ignores retention/hash representation | fake | Controlled source-lock fixtures vary retention metadata while preserving stable provider facts, making the boundary deterministic. |
| Citation-only unavailable sources remain closed against invention | fake | Hermetic v2/v3 fixtures prove tamper and marker-bypass rejection without relying on a mutable provider document. |
| Checked-in declaration admission remains sound | live | The real checked-in connector definitions and source locks pass `declaration-admission --json` with zero findings. |
| Certification remains separate and callable | live | The real checked-in definitions pass `certification-matrix --check`; no credential was supplied or required for mapping admission. |

No provider request, live credential, runtime transport, warehouse operation, or
source-lock rewrite was used as evidence for this authoring-only slice.
