# TDD Ledger — Zoom Workforce Management documented-operation parity, R1

## Planned RED contract

Before any Workforce Management connector declaration or CSV-foundation production change, the
test-only RED checkpoint must prove both gaps:

- Tasks-complete HEAD has `84` executable / `1,758` Zoom-local rows, `44` direct reads, and `35`
  direct writes; this category's target is `102` / `1,740` / `55` / `42`.
- All 18 real `workforce-management …` paths are unknown to the command runner before their
  declarations exist.
- A closed multipart `content_validation: "csv"` declaration is rejected because the existing
  policy supports only JSON validation; a `.csv` filename alone is not adequate source validation.

RED output will be recorded verbatim below before production changes.

## RED — captured before connector declaration or foundation changes

The test-only RED checkpoint ran against Workforce-Management-unimplemented HEAD. It contains no
provider credential or token value:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/... -run 'TestProviderInventoryLedgerIsComplete|TestCoveredStreamsHaveReachableCommands|TestWorkforceManagementOperationCommandsAreReachable'
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:160: executable rows = 84, want 102
    command_surface_test.go:163: operations awaiting Zoom-local contracts = 1758, want 1740
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.04s)
    command_surface_test.go:258: reachable direct_read operation commands = 44, want 55
    command_surface_test.go:259: reachable direct_write operation commands = 35, want 42
--- FAIL: TestWorkforceManagementOperationCommandsAreReachable (0.04s)
    command_surface_test.go:558: Preflight("workforce-management filter-groups list") = connector command "workforce-management filter-groups list" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management forecasts list") = connector command "workforce-management forecasts list" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management forecasts scheduling-groups get") = connector command "workforce-management forecasts scheduling-groups get" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management imports historical-agent-status upload") = connector command "workforce-management imports historical-agent-status upload" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management imports historical-agent-status delete") = connector command "workforce-management imports historical-agent-status delete" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management imports historical-queue-metrics upload") = connector command "workforce-management imports historical-queue-metrics upload" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management imports staffing upload") = connector command "workforce-management imports staffing upload" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management imports historical-queue-metrics get") = connector command "workforce-management imports historical-queue-metrics get" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management organizational-groups list") = connector command "workforce-management organizational-groups list" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management organizational-groups create") = connector command "workforce-management organizational-groups create" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management organizational-groups get") = connector command "workforce-management organizational-groups get" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management organizational-groups delete") = connector command "workforce-management organizational-groups delete" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management organizational-groups update") = connector command "workforce-management organizational-groups update" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management reports adherence agents list") = connector command "workforce-management reports adherence agents list" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management reports schedules agents list") = connector command "workforce-management reports schedules agents list" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management schedules agents list") = connector command "workforce-management schedules agents list" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management scheduling-groups list") = connector command "workforce-management scheduling-groups list" is blocked: unknown command, want declared executable Workforce Management action
    command_surface_test.go:558: Preflight("workforce-management users list") = connector command "workforce-management users list" is blocked: unknown command, want declared executable Workforce Management action
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom
FAIL
```

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteMultipartAcceptsDeclaredCSVFile$'
--- FAIL: TestOperationDirectWriteMultipartAcceptsDeclaredCSVFile (0.00s)
    operation_multipart_test.go:194: Load declared CSV multipart contract: load bundle acme: operations.json: /operations/0/rest/multipart/parts/1/content_validation: value not in enum [json]
FAIL
FAIL    polymetrics.ai/internal/connectors/engine
FAIL
```

Both commands exited `1`, exactly as expected. This is the committed red state; only tests and
phase evidence changed before it was captured.

## GREEN CSV foundation — captured before Workforce Management connector authoring

The separate foundation adds only the declaration-owned `content_validation: "csv"` policy plus
the existing closed file-extension list. It parses the bounded approved snapshot with the standard
CSV grammar before any wire send, preserves the declared `text/csv` part header, and leaves
provider-specific column semantics in the provider's own operation schema. A malformed CSV or a
non-`.csv` source fails locally; no caller can choose a parser, MIME policy, header, or upload
target. Existing JSON validation remains unchanged.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/connsdk ./internal/connectors/commandrunner ./cmd/connectorgen
ok      polymetrics.ai/internal/connectors/engine           5.368s
ok      polymetrics.ai/internal/connectors/connsdk          0.941s
ok      polymetrics.ai/internal/connectors/commandrunner    8.961s
ok      polymetrics.ai/cmd/connectorgen                    13.084s
```

The `TestOperationDirectWriteMultipartAcceptsDeclaredCSVFile` loopback test proves that a valid
`.csv` snapshot reaches the endpoint once with a `text/csv` part, while a malformed CSV and a
wrong extension fail before a second request. Constraint tests cover file-only use, canonical
extension lists, `text/csv` header compatibility, positive max-byte requirements, and the
unchanged JSON path.

## GREEN connector — pending

## Verification/review — pending
