# Verification — connector batch pipeline r1

Status: implementation verification complete; external authoring remains gated.

## Evidence to collect

- `go test ./cmd/connectorgen -run '^TestBatchPlan' -count=1`: passed after
  recording the baseline red failure (unknown `batch` command).
- `go test ./cmd/connectorgen -run '^TestBatchGate' -count=1`: passed after
  recording the baseline red failure (unknown `batch gate`). Tests prove
  continuation after a malformed sibling, real runtime-preflight command
  counting, and named surface-sync drift drops.
- `go test ./cmd/connectorgen -count=1`: passed.
- `go vet ./cmd/connectorgen`: passed.
- `go build ./cmd/connectorgen`: passed.
- `go build ./cmd/pm`: passed.
- `go test ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`: passed.
- `go run ./cmd/connectorgen batch --help`: passed and lists `plan` and `gate`.
- `jq` over the generated manifest reports five candidates, 190 surveyed
  operations, 113 reads, and 77 writes. Its source snapshot had 224 records;
  the external ledger was re-read live and is still growing.
- `git diff --check`: passed.
- Repository gates passed individually: `make tidy-check`, `make lint`,
  `make docs-check-no-build`, `make smoke-no-build`, `make
  agent-contract-check`, `make connectorgen-validate`, `make
  connectorgen-surface-sync`, `make connector-boundary`, and `make
  release-workflow-check`.

## External wait

The manifest and gate are complete on the current base. Actual bundle authoring
remains intentionally paused pending merged #3869 (v2 provenance), #3870
(non-redacting output policies), and associated #3868 verification. No
connector bundle, schema, engine, runner, credential, or provider API request
was changed.
