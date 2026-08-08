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

## GREEN — declarations and focused runtime proof

After the red commit `91b7526a5` was pushed, the five GET declarations, typed POST declaration,
source-synchronised endpoint coverage, generated Zoom documentation, and synthetic fixtures were
added. The first conformance pass exposed that the POST fixture did not model the write fixture
contract's `record`/`expect`/`response` structure. The fixture was corrected to carry the complete
documented request body and a `201` response; the assertion remains an exact body comparison and
was not weakened.

```text
$ go test -count=1 ./internal/connectors/defs/zoom/...
ok  \tpolymetrics.ai/internal/connectors/defs/zoom\t1.082s

$ go test -count=1 -v -run '^TestConformance/zoom$' ./internal/connectors/conformance
=== RUN   TestConformance/zoom
--- PASS: TestConformance/zoom
PASS
ok  \tpolymetrics.ai/internal/connectors/conformance\t2.316s

$ go test -count=1 -timeout 20m ./internal/connectors/conformance/...
ok  \tpolymetrics.ai/internal/connectors/conformance\t18.479s

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ go run ./cmd/connectorgen validate internal/connectors/defs/zoom
connectorgen validate: 1 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen validate
connectorgen validate: 551 connector(s) checked, 0 findings

$ go vet ./...
exit 0

$ go build ./cmd/pm
exit 0
```

The built binary proved all five direct reads reach Zoom (each safely stopped at provider `401`
with an environment-only synthetic credential and no `unknown command` result). The POST command
was exercised only as `--preview --json`: it produced the typed
`create_quality_management_interaction` plan without a network request.

The first full CLI package pass made the expected root-help drift visible: the newly generated Zoom
tagline appeared in nine root-help transcripts while the tracked expected fixture still carried the
old tagline. This is a documentation parity failure, not an unrelated test issue. The repository's
own golden generator was used to regenerate `internal/cli/testdata/golden_transcripts.json`; it
changed only those nine root variants. The regular (non-update) golden test and then the full CLI
package passed:

```text
$ POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m -run '^TestGoldenTranscripts$' ./internal/cli
ok  \tpolymetrics.ai/internal/cli\t51.443s

$ go test -count=1 -timeout 20m -run '^TestGoldenTranscripts$' ./internal/cli
ok  \tpolymetrics.ai/internal/cli\t26.419s

$ go test -count=1 -timeout 20m ./internal/cli/...
ok  \tpolymetrics.ai/internal/cli\t560.881s

$ go test -count=1 -timeout 20m ./internal/connectors/commandrunner/...
ok  \tpolymetrics.ai/internal/connectors/commandrunner\t6.977s
```
