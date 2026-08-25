# Verification — #4352 source-bound read execution foundation

Status: verified locally, pending PR automation and human gate.

## Focused red/green proof

- Red: `go test ./cmd/connectorgen -run '^TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL$' -count=1` failed before production changes because `get_access_requests` carried no source binding.
- Green generator coverage: `go test ./cmd/connectorgen -run '^(TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL|TestSourceProjectionLeavesIncompleteReadAsNamedFoundation|TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical|TestSourceProjectionRequiresExplicitReadOnlyNonMutationDeclaration|TestSourceProjectionRequiresReachableRESTReadOrConcreteGap)$' -count=1` passed.
- Green engine coverage: `go test ./internal/connectors/engine -run '^TestPreflightSourceBound' -count=1` passed.
- Green runner coverage: `go test ./internal/connectors/commandrunner -run '^(TestRunSourceBoundOperationDirectReadRejectsBeforeDispatch|TestRunSourceBoundReadMissingFoundationRefusesBeforeDispatch)$' -count=1` passed.
- Green Asana controls: `go test ./internal/connectors/defs/asana -run '^(TestSourceBoundReadControlsReachEnginePreflight|TestReverseETLLedgerReconciles|TestDestructiveOperationsStayBlocked|TestReverseETLWriteActionsExecute)$' -count=1` passed.

## Regression and repository gates

- `go test -timeout 20m ./internal/connectors/engine -count=1` passed.
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1` passed, including `TestEveryImplementedCommandPassesRuntimePreflight`.
- `go test -timeout 20m ./internal/connectors/defs/asana -count=1` passed.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make docs-check-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make lint`, `make connector-runtime-preflight`, `make smoke-no-build`, `make connector-canon-check`, and `npm --prefix website run typecheck` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs` reported 553 connectors and 0 findings; `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` reported no drift.
- `git diff --check` passed after the final generated-doc refresh.

`connector-boundary` and `release-installed-github-certification.sh` exceeded the local per-command runner limit. Their child processes were terminated after confirming they were this task's duplicate validation attempts; they are not recorded as passing and remain CI/PR checks.

## Credential boundary

Built a fresh temporary `pm`, initialised a fresh temporary project, and invoked `pm asana access-requests get-access-requests --root <temp> --json` without a credential. It returned `error: missing --credential` (exit 1). No provider credential was configured and no provider request was made.

## CLI/docs/website parity

- Checked `pm --help`, `pm asana --help`, and the generated Asana manual/skill output.
- `docs/connectors/asana/{MANUAL,SKILL}.md`, `docs/skills/pm-asana/SKILL.md`, the connector catalog, and website generated connector data were refreshed.
- Top-level `docs/cli/**` did not need a hand-authored command-tree change; broad unrelated generated pages were intentionally not included.
