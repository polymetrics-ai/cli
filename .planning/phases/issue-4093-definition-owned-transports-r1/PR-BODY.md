Refs #4093
Refs #4015

## Reconciled scope

This is a residual completion PR, not a rewrite of the original issue.

- Merged #4156 (`11fd27b4b`) already delivered the versioned strict
  `sync_transport.json` loader, clone-safe `engine.Base.Definition` projection,
  definition-selected conformance evidence, exact family/identifier factory
  composition, and production PostgreSQL/GitHub hooks without connector-name
  branches in shared registration.
- Merged #4186 (`281560ca1`) moved GitHub issue-label source admission into
  `defs/github/sync_transport.json` through `destination_transport.source_bindings`.
- Merged #4187 and #4188 made the advertised history modes executable. The
  issue text's older history refusal is therefore deliberately not restored.
  The still-relevant CDC boundary remains closed: a destination uses
  `change_apply`, never independently advertises `change_capture`.

The remaining gap was a production-composition proof for a loaded second
definition, plus the post-checkpoint transient-stage window called out in the
issue's kill/reconciliation acceptance.

## Delivered

- Add `TestAppCompositionRoutesLoadedSyntheticDefinitionConnector`: a complete
  throwaway bundle declares both roles solely in its own
  `sync_transport.json`. It is loaded by `engine.Load`, discovered through the
  normal `DefinitionFactoryProvider` path, selected only by declared
  family/identifier/evidence and a named test hook, then executed by the real
  App-composed registry and generic orchestrator. It reads, stages, plans,
  applies, read-backs, and commits one record. No App, orchestrator, or dispatch
  production edit is required to register that connector; a connector-name
  branch reintroduced in those paths makes the test fail.
- Add the optional `RetirableWarehouseStage` contract. The connection-owned
  Parquet stage retires only an exact validated receipt, derives all paths from
  its owner and opaque stage ID, verifies its retained manifest, and never
  accepts a path or removes a directory tree.
- Retire committed receipts only after the durable checkpoint callback
  succeeds. Reconcile a bounded maximum of 64 owned manifest candidates on
  startup by matching their candidate checkpoint to the persisted committed
  checkpoint. An interruption during cleanup remains retryable because data is
  removed before its manifest; active, malformed, foreign, and uncommitted
  worksets are retained.

## Acceptance evidence

| Case | Evidence |
| --- | --- |
| Happy | `TestAppCompositionRoutesLoadedSyntheticDefinitionConnector` proves the definition-loaded synthetic connector end-to-end. `TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit` observes one retirement after one durable commit. |
| Bad / before I/O | Existing `TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations` proves unknown/version/role members fail closed. `TestPreflightReturnsTypedDestinationSourceIneligibleErrorBeforeExecutorAccess` asserts the specific `*DestinationSourceIneligibleError` and zero executor access. `TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration` proves malformed declarations produce zero builders/registrations. |
| CDC role boundary | `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination` permits only the declared closed `change_apply` pairing and asserts a malformed `change_capture`/`append` pair performs zero builds, registrations, reads, plans, and applies. |
| Kill-after-commit edge | `TestOpenReconcilesOnlyCommittedConnectionOwnedTransportStages` simulates death after checkpoint persistence. A fresh App removes only that receipt's manifest/WAL/Parquet paths and preserves a second active receipt. The checkpoint is the durable resume guard, so the already-applied effect cannot be repeated. |

## GSD and skills

Used the required Go skills (`golang-how-to`, testing, error handling, safety,
security, design patterns, structs/interfaces, naming, context, concurrency,
database, and lint) after the project routing reference. Ran `scripts/gsd
doctor`, resolved each required lifecycle command with `scripts/gsd sources`,
and completed `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` inline/manual because this issue's existing
phase directory is non-numeric and the adapter's role runtime is unavailable.
The RED/GREEN evidence and inline review are in the phase artifacts.

## Verification

All commands below passed (exit 0) after the recorded RED, unless noted:

```text
go test -count=1 -timeout 20m ./internal/app -run '^TestAppCompositionRoutesLoadedSyntheticDefinitionConnector$'
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit$'
go test -count=1 -timeout 20m ./internal/app -run '^TestOpenReconcilesOnlyCommittedConnectionOwnedTransportStages$'
go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadSyncTransportRefusesUnknownOrUnsafeDeclarations$'
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestPreflightReturnsTypedDestinationSourceIneligibleErrorBeforeExecutorAccess$'
go test -count=1 -timeout 20m ./internal/synctransport -run '^TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration$'
go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination$'
go test -count=1 -timeout 20m ./internal/synctransport
go test -count=1 -timeout 20m ./internal/connectors/engine
go test -count=1 -timeout 20m ./internal/connectors
go test -count=1 -timeout 20m ./internal/app
go test -count=1 -timeout 20m ./internal/cli
go test -count=1 -timeout 20m ./cmd/connectorgen
go vet ./...
go build ./cmd/pm
go run ./cmd/connectorgen boundary . --json
pnpm --dir website run gen:docs  # twice; generated 12 pages and left no diff
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
make connectorgen-certification-matrix
make connector-canon-check
make github-parity-artifacts-check
git diff --check
```

The repository's agent guidance explicitly says not to run monolithic
`go test ./...` or `make verify` under a per-command timeout. I instead ran the
changed packages plus `internal/cli` separately and every non-test `make verify`
gate; full-suite coverage remains CI.

No user-facing CLI, generated manual, or website source changed. Website docs
generation was run twice and was byte-stable.
