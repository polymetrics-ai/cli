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

## GREEN connector

The declarative connector now contains all sixteen fixed Auto Dialer contracts: eight bounded
`rest_read` commands and eight approval-gated `rest_write` commands. Every command maps to exactly
one live-artifact method/path and carries normal Zoom bearer authentication, source-derived typed
path/body input, and a redacted or status-only output policy. No generic transport, raw JSON,
or hand-authored paging flag was introduced.

- Both DELETE commands, call-list update, and single-prospect update assert documented `204 No
  Content` by status only; DELETE additionally requires destructive typed confirmation.
- Typed named-object bodies remain restricted to their declared operation schemas. Call-list and
  single-prospect updates require a nonempty object; batch prospect update requires a prospect ID
  plus at least one mutable member, and batch request limits match the documented 1,000-item cap.
- `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...` passes after declarations,
  fixture lifecycle checks, generated metadata, and source-reconciled endpoint coverage.
- `go run ./cmd/connectorgen surface-sync --check` and
  `go run ./cmd/connectorgen surface-reconcile --check --notes-contains provider_module=auto-dialer`
  both pass with zero pending changes.

No missing engine foundation was discovered or added in this slice.

## Verification/review

Inline manual `verify-work` and `code-review` were completed because this provider-category phase
is not registered by the official runtime and the parent contract prohibits role spawning. The
fixture tests exercise all eight reads and all eight writes against the real commandrunner path,
including exact bearer/method/path/body/status assertions, no invented query/paging input,
redaction, approval flow, and destructive confirmation. A fresh compiled binary reaches Zoom,
the bare Auto Dialer namespace, and all sixteen command help routes. Scoped tests, vet, generated
docs/site validation, Make gates, and website typecheck/lint pass; website lint has only the
repository's 13 pre-existing warnings and no errors.

Manual review found and corrected two contract-hardening details before the final checks: a batch
prospect update can no longer submit an ID-only no-op object, and the output policies now redact
the provider's call, contact, prospect, company, report, and transcript-sensitive fields. No
remaining blocking review finding exists.
