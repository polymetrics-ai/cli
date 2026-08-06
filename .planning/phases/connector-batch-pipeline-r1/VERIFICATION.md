# Verification — connector batch pipeline r1

Status: current-main implementation verification complete; external authoring remains gated.

## Evidence to collect

- `go test ./cmd/connectorgen -run '^TestBatchPlan' -count=1`: passed after
  recording the baseline red failure (unknown `batch` command).
- `go test ./cmd/connectorgen -run '^TestBatchGate' -count=1`: passed after
  recording the baseline red failure (unknown `batch gate`). Tests prove
  continuation after a malformed sibling, real runtime-preflight command
  counting, and named surface-sync drift drops.
- `go test ./cmd/connectorgen -run
  '^TestBatchGateDropsRedactingOutputPolicy$' -count=1`: recorded the red
  failure in the TDD ledger, then passed after the gate began dropping
  `json_redacted` declarations. The companion write-`redact_fields` red then
  green test is recorded in the same ledger.
- `go test ./cmd/connectorgen -count=1`: passed.
- `go vet ./cmd/connectorgen`: passed.
- `go build ./cmd/connectorgen`: passed.
- `go build ./cmd/pm`: passed.
- `go test ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`: passed.
- `go run ./cmd/connectorgen batch --help`: passed and lists `plan` and `gate`.
- `jq` over the generated manifest reports five candidates, 190 surveyed
  operations, 113 reads, and 77 writes. Its source snapshot was regenerated
  from the live ledger at 235 records; the external ledger is still growing.
- `git diff --check`: passed.
- Repository gates passed individually: `make tidy-check`, `make lint`,
  `make docs-check-no-build`, `make smoke-no-build`, `make
  agent-contract-check`, `make connectorgen-validate`, `make
  connectorgen-surface-sync`, `make connector-boundary`, and `make
  release-workflow-check`.
- The branch was rebased onto `origin/main` at `50deaade9` after #3870
  (`ee26d20fc`) and #3868 (`50deaade9`) merged, then every preceding local
  gate was rerun. `go run ./cmd/connectorgen batch --help` lists the expected
  `plan` and `gate` commands.

## External wait

The manifest and gate are complete on the current base. Actual bundle authoring
remains intentionally paused pending only merged #3869 (v2 provenance). #3870
(non-redacting direct-write policies) and #3868 (command-runner content
preservation) are present in the verified base. No connector bundle, schema,
engine, runner, credential, or provider API request was changed.
