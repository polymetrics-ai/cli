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

## RED numeric-range foundation — captured before numeric-bound production change

The live staffing-import schema additionally publishes
`forecast_duration_weeks` as a **number** in the inclusive range `1`–`4`. During connector
validation, the closed schema compiler rejected the standard Draft-07 `minimum` and `maximum`
keywords. Replacing the numeric contract with an integer enum would change the provider surface,
so the reusable declaration-owned numeric-range foundation is required before this operation can
ship. The following test-only checkpoint was staged and committed before engine production code;
the uncommitted Workforce Management declaration was deliberately not included in that test
commit.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestCompileSchemaNumericRangeKeywords|TestSchemaValidateNumericRange|TestSchemaNumericRangeIgnoresNonNumbers|TestCompileSchemaRejectsInvalidNumericRange'
--- FAIL: TestCompileSchemaNumericRangeKeywords (0.00s)
    schema_test.go:496: CompileSchema({"type":"number","minimum":1}): unexpected error: compile schema: unknown keyword "minimum"
--- FAIL: TestSchemaValidateNumericRange (0.00s)
    schema_test.go:504: CompileSchema: compile schema: unknown keyword "minimum"
--- FAIL: TestSchemaNumericRangeIgnoresNonNumbers (0.00s)
    schema_test.go:538: CompileSchema: compile schema: unknown keyword "minimum"
--- FAIL: TestCompileSchemaRejectsInvalidNumericRange (0.00s)
    --- FAIL: TestCompileSchemaRejectsInvalidNumericRange/maximum_below_minimum (0.00s)
        schema_test.go:564: error should mention "maximum", got compile schema: unknown keyword "minimum"
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.757s
FAIL
```

The command exited `1`, exactly as expected. The required green foundation must accept only
finite JSON numbers, reject an unsatisfiable maximum below a declared minimum at compile time,
and apply the bounds only to numeric instances just as Draft-07 specifies.

## GREEN numeric-range foundation — captured before connector authoring

The engine now accepts Draft-07 `minimum` and `maximum` only as finite JSON numbers, rejects an
unsatisfiable declared maximum below minimum at bundle-load time, and applies bounds to numeric
instances only. This preserves provider decimals and the normal Draft-07 applicability rule; it
does not create a generic validation tool or weaken the closed bundle dialect.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestCompileSchemaNumericRangeKeywords|TestSchemaValidateNumericRange|TestSchemaNumericRangeIgnoresNonNumbers|TestCompileSchemaRejectsInvalidNumericRange'
ok      polymetrics.ai/internal/connectors/engine    0.738s
```

The test covers lower/upper source boundaries, an in-range fraction, both out-of-range failures,
nonnumeric applicability, malformed bounds, and a contradictory declared range. This foundation
unblocks any declarative connector that must preserve provider-published numeric request bounds;
its implementation commit is separate from the Workforce Management authoring commit.

## GREEN connector — captured

All eighteen audited Workforce Management operations are now declarative, typed, and executable:
eleven fixed-path reads and seven approval-gated writes. The two source-declared `204 No Content`
DELETE actions require destructive confirmation and assert status-only success; the two CSV writes
use the closed CSV policy. No request paging flag was introduced.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/... -run 'TestWorkforceManagement(OperationCommandsAreReachable|DirectReadCommandsExecuteWithFixtures|JSONDirectWriteCommandsExecuteWithFixtures|CSVDirectWritesExecuteWithFixtures)'
ok      polymetrics.ai/internal/connectors/defs/zoom    3.762s

$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
ok      polymetrics.ai/internal/connectors/defs/zoom    15.171s

$ go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/connsdk ./internal/connectors/commandrunner ./cmd/connectorgen
ok      polymetrics.ai/internal/connectors/engine           5.313s
ok      polymetrics.ai/internal/connectors/connsdk          0.952s
ok      polymetrics.ai/internal/connectors/commandrunner    8.033s
ok      polymetrics.ai/cmd/connectorgen                    13.080s
```

The fixture lifecycle reaches all eleven GET paths via the real command runner and all seven
writes via plan → no-network preview → single-use approval → execute. It proves ordinary bearer
delivery, exact paths, absent undeclared query/paging input, JSON or multipart bodies, CSV
snapshot/header validation, output redaction, and both destructive confirmation gates. The
staffing fixture also proves that `4.01` is rejected by the declared source maximum before any
network dispatch.

```text
$ go run ./cmd/connectorgen surface-reconcile --notes-contains provider_module=workforce-management
zoom: reconciled covered=18 blocked=0 unchanged=0 refused=0
connectorgen surface-reconcile: 551 connector(s) scanned; covered=18 blocked=0 unchanged=0 refused=0

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

Reconciliation converts exactly the audited Zoom rows to `11` direct reads and `7` direct writes;
the generated endpoint-ledger delta consists only of Zoom's eleven `rest_read` Workforce
Management paths. The provider has zero `unsafe_or_disallowed` rows.

## Verification/review — captured

A fresh `go build -o .tmp/pm ./cmd/pm` binary returned success for `pm help zoom`, bare `pm zoom`,
bare `pm zoom workforce-management`, and every one of the eighteen exact Workforce Management
command paths with `--help`. Generated connector docs were validated; whole-tree docs generation
retained only Zoom manuals, and whole-tree website generation retained only the Zoom records after
non-Zoom JSON equivalence checks. Scoped vet, lint, CLI tests, contract, smoke, documentation,
surface, boundary, and release-workflow checks are recorded in `VERIFICATION.md`.
