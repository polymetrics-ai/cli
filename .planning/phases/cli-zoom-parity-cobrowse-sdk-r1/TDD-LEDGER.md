# TDD Ledger — Zoom Cobrowse SDK documented-operation parity, R1

## RED — captured before production declarations

The red checkpoint changes only `internal/connectors/defs/zoom/command_surface_test.go` and
synthetic Cobrowse fixtures. It asserts the documented four-operation target before any Zoom
production declaration changes:

- covered operations: `18 → 22`
- locally blocked: `1824 → 1820`
- direct reads: `13 → 17`
- writes: unchanged at `2`
- two exact monthly report GET routes with only explicit `from`/`to` input; two exact session-ID
  routes; response redaction; and proof that response-only pagination fields are never sent.

The test and synthetic fixture-only state was run against the pre-Cobrowse Zoom bundle:

```text
$ go test -count=1 ./internal/connectors/defs/zoom/...
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:147: executable rows = 18, want 22
    command_surface_test.go:150: operations awaiting Zoom-local contracts = 1824, want 1820
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
    command_surface_test.go:245: reachable direct_read operation commands = 13, want 17
--- FAIL: TestCobrowseSDKCommandsExecuteWithFixtures (0.12s)
    --- FAIL: TestCobrowseSDKCommandsExecuteWithFixtures/list_live_sessions (0.03s)
        command_surface_test.go:1059: Run("cobrowse-sdk live-sessions list") = connector command "cobrowse-sdk live-sessions list" is blocked: unknown command
    --- FAIL: TestCobrowseSDKCommandsExecuteWithFixtures/list_past_sessions (0.03s)
        command_surface_test.go:1059: Run("cobrowse-sdk past-sessions list") = connector command "cobrowse-sdk past-sessions list" is blocked: unknown command
    --- FAIL: TestCobrowseSDKCommandsExecuteWithFixtures/get_session (0.03s)
        command_surface_test.go:1059: Run("cobrowse-sdk sessions get") = connector command "cobrowse-sdk sessions get" is blocked: unknown command
    --- FAIL: TestCobrowseSDKCommandsExecuteWithFixtures/list_session_users (0.03s)
        command_surface_test.go:1059: Run("cobrowse-sdk sessions users list") = connector command "cobrowse-sdk sessions users list" is blocked: unknown command
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom    1.139s
FAIL
```

This red state contains only the command-surface test, synthetic fixtures, and delivery evidence.
`operations.json`, `writes.json`, `cli_surface.json`, `api_surface.json`, metadata, generated
ledger, connector docs, and website data remain untouched. Commit and push it before creating any
of those declarations.

## GREEN — pending

Record focused test, conformance, surface/validator, binary, docs/website, and inline review
evidence after the declarations exist and the red test becomes green.
