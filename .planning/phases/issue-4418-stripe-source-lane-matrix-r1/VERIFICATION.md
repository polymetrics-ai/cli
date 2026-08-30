# Verification plan — #4418 Stripe source-to-seven-lane matrix

## Required commands

```sh
gofmt -w internal/connectors/defs/stripe/source_lane_matrix_test.go
go test ./internal/connectors/defs/stripe -run TestStripeSourceLaneMatrix -count=1
go test ./internal/connectors/defs/stripe -count=1
jq empty internal/connectors/defs/stripe/sources/stripe-source-lane-matrix.json
go run ./cmd/connectorgen source-import stripe --defs internal/connectors/defs --check
go test ./cmd/connectorgen -run 'TestSourceImport(CheckedInGitHubArtifactsAreRetainedAndLockVerified|V3SourceReferenceUsesTheSameClosedProjection)|TestOperationEvidence(ConnectorFilterIsBounded|GitLabSourceLockBridge)' -count=1
go test ./... -count=1
```

The broad suite is a baseline observation only. Any unrelated failures will be listed by test identifier/count and will not broaden this source-only change.

## Acceptance evidence

- Matrix source-lock reference matches the pinned Stripe source and counts all 589 source IDs exactly once.
- Seven cells exist for every source row; their states, source locations, and stable typed reasons are valid.
- The exact 128 paging candidates have explicit ETL and sync dispositions; list-shaped rows without an operation continuation stay visible and non-applicable.
- All 326 mutation rows, including 32 DELETE rows, have visible direct-write and reverse-ETL mapping dispositions.
- Every artifact link resolves to an existing source row, cell, and current artifact record.
- No runtime behavior, credential, transport, executor, or shared foundation code is changed.

## Results

| Check | Result |
| --- | --- |
| `gofmt -d internal/connectors/defs/stripe/source_lane_matrix_test.go` | Pass; no diff. |
| `jq empty internal/connectors/defs/stripe/sources/stripe-source-lane-matrix.json` | Pass. |
| `go vet ./internal/connectors/defs/stripe` | Pass. |
| `go test ./internal/connectors/defs/stripe -run TestStripeSourceLaneMatrix -count=1` | Pass. |
| `go test -race ./internal/connectors/defs/stripe -run TestStripeSourceLaneMatrix -count=1` | Pass. |
| `go test ./internal/connectors/defs/stripe -count=1` | Pass. |
| Selected `cmd/connectorgen` source-import and operation-evidence tests | Pass. |
| `go run ./cmd/connectorgen validate internal/connectors/defs --connector stripe` | Pre-existing non-scoped failure: `stripe: sources/stripe-operation-descriptor.json: [source_projection] canonical source descriptor is missing` (one finding). This task neither creates nor changes descriptors. |
| `go run ./cmd/connectorgen source-import stripe --defs internal/connectors/defs --check` | Not applicable to this retained legacy Stripe source lock: it requires `sources/stripe-retained-artifacts.json`, which is absent at the base commit. The canonical descriptor is also absent at the base. This task neither creates nor changes source-projection artifacts. |
| `go test ./... -count=1` | Exit 1, non-scoped baseline only. Six packages failed (seven top-level test IDs): `cmd/connectorgen` `TestCertificationCheckIgnoresMalformedNonAllowlistedRuntimeLedgerEntry` timed out; `internal/app` `TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume` timed out; `internal/cli` `TestExternalProofFreshChildRefusesNoHTTPSWithoutArtifact` timed out; `internal/connectors/conformance` `TestConformance` reported Asana `create_batch_request` missing provider response `data` and GitLab invalid `\\u` schema-regexp escape; `internal/connectors/defs` failed `TestProductionEmbedDeclaresOnlyGithubSourceLockException` and `TestProductionEmbedInventoryIsDeterministicAndAttributed` on an unclassified Asana artifact; `internal/connectors/engine` failed `TestBinaryDownloadNoOverwritePublicationIsCrashAndRaceSafe`. `internal/connectors/defs/stripe` passed in the same broad run. |

## Integration readiness

Ready for independent review as a connector-local source matrix. Its only composition dependency is that the Batch R1 parent may incorporate the checked Stripe matrix into #4293's root-level multi-connector manifest after landing both scoped branches. No runtime foundation gap, executor work, credential action, retained-artifact manifest, or descriptor work is requested by this task.
