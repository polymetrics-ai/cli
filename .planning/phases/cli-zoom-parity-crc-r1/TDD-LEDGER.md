# TDD Ledger — Zoom CRC documented-operation parity, R1

## Planned RED contract

Before any CRC bundle declaration, this test-only checkpoint must prove all of the following:

- Clips-complete HEAD has `123` executable / `1,719` Zoom-local rows, `61` direct reads, and
  `58` direct writes. CRC's target is `143` / `1,699` / `70` / `69`.
- Every documented `crc …` command is unknown through the real `commandrunner.Preflight`.
- The CRC test names all twenty source method/path pairs, correct operation IDs, intents, and
  output policies. It pins the three destructive DELETE routes and all seven status-only actions.
- The source's private-key response has a redacted output-policy expectation; no private value is
  placed in the test or its failure output.

RED output will be captured verbatim below before production changes. No provider credential,
token-derived value, private key, or signed URL may appear in this ledger.

## RED — captured 2026-08-08

The command ran before any CRC production bundle or engine change:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run 'TestProviderInventoryLedgerIsComplete|TestCoveredStreamsHaveReachableCommands|TestCRCOperationCommandsAreReachable'
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:167: executable rows = 123, want 143
    command_surface_test.go:170: operations awaiting Zoom-local contracts = 1719, want 1699
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.05s)
    command_surface_test.go:265: reachable direct_read operation commands = 61, want 70
    command_surface_test.go:266: reachable direct_write operation commands = 58, want 69
--- FAIL: TestCRCOperationCommandsAreReachable (0.04s)
    command_surface_test.go:744: Preflight("crc managed-rooms account-setting get") = connector command "crc managed-rooms account-setting get" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc managed-rooms account-setting update") = connector command "crc managed-rooms account-setting update" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc api-connectors list") = connector command "crc api-connectors list" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc api-connectors create") = connector command "crc api-connectors create" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc api-connectors get") = connector command "crc api-connectors get" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc api-connectors delete") = connector command "crc api-connectors delete" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc api-connectors update") = connector command "crc api-connectors update" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc api-connectors private-key get") = connector command "crc api-connectors private-key get" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc api-connectors private-key update") = connector command "crc api-connectors private-key update" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc managed-rooms list") = connector command "crc managed-rooms list" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc managed-rooms create") = connector command "crc managed-rooms create" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc managed-rooms get") = connector command "crc managed-rooms get" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc managed-rooms delete") = connector command "crc managed-rooms delete" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc managed-rooms update") = connector command "crc managed-rooms update" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc participant-identifier-code get") = connector command "crc participant-identifier-code get" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc room-templates list") = connector command "crc room-templates list" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc room-templates create") = connector command "crc room-templates create" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc room-templates get") = connector command "crc room-templates get" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc room-templates delete") = connector command "crc room-templates delete" is blocked: unknown command, want declared executable CRC action
    command_surface_test.go:744: Preflight("crc room-templates update") = connector command "crc room-templates update" is blocked: unknown command, want declared executable CRC action
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/zoom	0.906s
FAIL
```

The output contains only command paths and declared operation names; it does not contain a
credential, token-derived value, private key, or signed URL.

### Derived camelCase path mapping — additional RED 2026-08-08

The CRC declaration intentionally leaves path `maps_to` values to `surface-sync`. The existing
deriver handled `issue-id` → `{issue_id}` but not the equally conventional
`connector-id` → `{connectorId}` spelling. The focused generator test was added and run before
the shared derivation change:

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSyncBundleDerivesKebabFlagToCamelCasePathVariable$'
--- FAIL: TestSyncBundleDerivesKebabFlagToCamelCasePathVariable (0.00s)
    surfacesync_test.go:179: filled path maps_to = 0, want 1 (stats: {Filled:api_surface=1 output_policy=1 flag_maps_to=0 flag_derived=0 rest.max_bytes=1 Corrected:api_surface=0 output_policy=0 flag_maps_to=0 flag_derived=0 rest.max_bytes=0})
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.727s
FAIL
```

The fixture is local only, declares no credential or request, and proves a reusable metadata
derivation gap rather than a CRC-specific handwritten exception.

### Derived camelCase path mapping — GREEN 2026-08-08

`surface-sync` now tries its established kebab-to-snake spelling first, then derives a lower-camel
candidate and accepts it only if that exact variable occurs in the operation's declared path.
It cannot create a path variable, map a query/body field, or override a non-path author choice.
The foundation is independent of Zoom and allows future provider bundles to keep path metadata
generated instead of copied.

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSyncBundleDerivesKebabFlagToCamelCasePathVariable|TestSyncBundleReportsDivergentFlagMapsTo|TestSyncBundleDirectWriteDerivesOperationContract)$'
ok  	polymetrics.ai/cmd/connectorgen	0.727s

$ go test -count=1 -timeout 20m ./cmd/connectorgen
ok  	polymetrics.ai/cmd/connectorgen	11.417s
```

## GREEN — pending

The passing command and output will be appended after the category declaration and generated
surface reconciliation are complete.
