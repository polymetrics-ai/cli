# TDD Ledger — Zoom Auto Dialer documented-operation parity, R1

## Planned RED contract

Before any Auto Dialer production declaration changes, the RED checkpoint contains only the
command-surface test and phase evidence. Against Virtual-Agent-complete HEAD it must fail because:

- Zoom is at `51` executable / `1,791` locally implementable rows, with `30` direct reads and
  `16` direct writes; the Auto Dialer target is `67` / `1,775` / `38` / `24`.
- All sixteen provider paths are absent from the real commandrunner preflight, so a compiled
  `pm zoom auto-dialer …` route remains an `unknown command` before its declaration exists.
- The existing named-root-object support is fixed-operation and schema-bound; RED does not add a
  generic body, HTTP, or paging capability.

The RED output will be pasted verbatim below before any production JSON, metadata, fixture, or
generated-file edit.

## RED — captured before connector declaration changes

The test-only RED checkpoint was run against the Virtual-Agent-complete bundle before any Auto
Dialer production declaration, fixture, metadata, or generated-file change. It contains no provider
credential or token value:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:156: executable rows = 51, want 67
    command_surface_test.go:159: operations awaiting Zoom-local contracts = 1791, want 1775
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.04s)
    command_surface_test.go:254: reachable direct_read operation commands = 30, want 38
    command_surface_test.go:255: reachable direct_write operation commands = 16, want 24
--- FAIL: TestAutoDialerOperationCommandsAreReachable (0.04s)
    command_surface_test.go:437: Preflight("auto-dialer call-histories get") = connector command "auto-dialer call-histories get" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-history list") = connector command "auto-dialer call-history list" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer reports call-history list") = connector command "auto-dialer reports call-history list" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer reports seller-productivity get") = connector command "auto-dialer reports seller-productivity get" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists list") = connector command "auto-dialer call-lists list" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists create") = connector command "auto-dialer call-lists create" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists get") = connector command "auto-dialer call-lists get" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists delete") = connector command "auto-dialer call-lists delete" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists update") = connector command "auto-dialer call-lists update" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists prospects list") = connector command "auto-dialer call-lists prospects list" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists prospects create") = connector command "auto-dialer call-lists prospects create" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists prospects update-batch") = connector command "auto-dialer call-lists prospects update-batch" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists prospects create-batch") = connector command "auto-dialer call-lists prospects create-batch" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists prospects delete") = connector command "auto-dialer call-lists prospects delete" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer call-lists prospects update") = connector command "auto-dialer call-lists prospects update" is blocked: unknown command, want declared executable Auto Dialer action
    command_surface_test.go:437: Preflight("auto-dialer prospects get") = connector command "auto-dialer prospects get" is blocked: unknown command, want declared executable Auto Dialer action
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom
FAIL
```

This is the committed red state. Connector declaration work begins only after it is pushed.

## GREEN connector — pending

## Verification/review — pending
