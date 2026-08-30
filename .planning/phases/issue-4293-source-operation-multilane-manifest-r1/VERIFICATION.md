# Verification — issue 4293 source-operation multi-lane manifest

## Scoped checks — passed

- `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceOperationMapping' -count=1`
- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go vet ./cmd/connectorgen ./internal/connectors/engine`
- `gofmt -w` on every changed Go file and `git diff --check`
- `jq empty internal/connectors/engine/schema/source_operation_mapping.schema.json`
- `go run ./cmd/connectorgen --help` contains `source-operation-mapping <manifest> --check`
- `go run ./cmd/connectorgen declaration-admission` (`1 connector(s), 1 source operation(s), 0 finding(s)`)
- `go run ./cmd/agentcontractgen check`

The focused suite creates its own manifest and source locks and calls the
check-only command, so it exercises schema validation, strict duplicate-member
rejection, source-lock reconciliation, and artifact/cell linkage without
writing a repository artifact.

## Existing broader-check findings — not in this slice

- `go test -timeout 20m ./cmd/connectorgen -count=1` ran for 622.602 seconds
  and failed in six existing connector-parity tests:
  `TestImplementedCommandEndpointEquivalenceCoversExactFleet` (249 non-GraphQL
  aliases observed vs 246 expected; GraphQL remained 4),
  `TestOperationEvidenceGitLabSourceLockBridge` (967 source identities vs 733
  expected), `TestRetainedAsanaSourceImportRejectsReadProjectionDrift`,
  `TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation`,
  `TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant`, and
  `TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical`.
- `go run ./cmd/connectorgen operation-evidence --check` refused existing
  `internal/connectors/operation-evidence.json` drift and suggested its
  write-fixed-100 command. No artifact was rewritten.
- `go run ./cmd/connectorgen validate internal/connectors/defs` reported
  `553 connector(s) checked, 50 finding(s)` in existing Bitbucket, CircleCI,
  Docker Hub, Jira, Notion, Sentry, Stripe, and Vercel source-projection or
  body-schema data. No connector file was changed.

These commands exercise pre-existing batch-parity work outside this diff; the
new command is not in their failing call paths. They are recorded for Batch R1
integration rather than repaired here.

## CLI parity disposition

The added command is developer-only `connectorgen` help, so its main usage and
subcommand help are covered. There is no new `pm` runtime command, JSON mode,
manual page, website page, generated PM reference, completion, credential
flow, or reverse-ETL path; those surfaces are intentionally not applicable.

## Runtime boundary

This task must not alter `internal/connectors/commandrunner`, connector executors, transport execution, credential resolution, or certification runners. Any apparent need to do so is a stop-and-report condition.

Review outcome: no runtime foundation gap was found. The implementation invokes
only the existing mapping-only source-lock reader and schema validation. It
does not modify a connector definition, source lock, retained bytes, executor,
transport, credential path, or certification behavior.
