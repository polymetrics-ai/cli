# TDD Ledger — Zoom Quality Management documented-operation parity, R1

## RED — captured before production declarations

The red checkpoint must change only `internal/connectors/defs/zoom/command_surface_test.go` and
synthetic Quality Management fixtures. It asserts the documented six-operation target before any
Zoom production declaration changes:

- covered operations: `12 → 18`
- locally blocked: `1830 → 1824`
- direct reads: `8 → 13`
- writes: `1 → 2`
- five fixed GET paths, two required detail IDs, no invented paging flags, response redaction, and
  one exact POST JSON body with a successful fixture `201`.

The test and synthetic fixture-only state was run against the pre-Quality-Management Zoom bundle:

```text
$ go test -count=1 ./internal/connectors/defs/zoom/...
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:147: executable rows = 12, want 18
    command_surface_test.go:150: operations awaiting Zoom-local contracts = 1830, want 1824
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
    command_surface_test.go:245: reachable direct_read operation commands = 8, want 13
    command_surface_test.go:246: reachable reverse_etl write commands = 1, want 2
--- FAIL: TestQualityManagementCommandsExecuteWithFixtures (0.22s)
    --- FAIL: TestQualityManagementCommandsExecuteWithFixtures/list_automated_evaluations (0.05s)
        command_surface_test.go:829: Run("quality-management automated-evaluations list") = connector command "quality-management automated-evaluations list" is blocked: unknown command
    --- FAIL: TestQualityManagementCommandsExecuteWithFixtures/list_evaluations (0.04s)
        command_surface_test.go:829: Run("quality-management evaluations list") = connector command "quality-management evaluations list" is blocked: unknown command
    --- FAIL: TestQualityManagementCommandsExecuteWithFixtures/get_evaluation (0.03s)
        command_surface_test.go:829: Run("quality-management evaluations get") = connector command "quality-management evaluations get" is blocked: unknown command
    --- FAIL: TestQualityManagementCommandsExecuteWithFixtures/list_interactions (0.03s)
        command_surface_test.go:829: Run("quality-management interactions list") = connector command "quality-management interactions list" is blocked: unknown command
    --- FAIL: TestQualityManagementCommandsExecuteWithFixtures/get_interaction (0.04s)
        command_surface_test.go:829: Run("quality-management interactions get") = connector command "quality-management interactions get" is blocked: unknown command
    --- FAIL: TestQualityManagementCommandsExecuteWithFixtures/create_interaction_plans_before_mutation_and_accepts_created (0.03s)
        command_surface_test.go:896: BuildWriteCommand = connector command "quality-management interactions create" is blocked: unknown command
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom    1.049s
FAIL
```

This red state contains only the command-surface test, synthetic fixtures, and delivery evidence.
`operations.json`, `writes.json`, `cli_surface.json`, `api_surface.json`, metadata, generated
ledger, and connector documentation production files remain untouched. Commit and push it before
creating any of those declarations.

## GREEN — pending

Record the focused test, surface/validator, binary, docs, and review evidence after the declarations
exist and the red test becomes green.
