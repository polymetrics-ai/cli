# TDD Ledger — Zoom Virtual Agent documented-operation parity, R1

## Planned RED contract

Before any Virtual Agent production declaration changes, the RED checkpoint contains only the
command-surface test and phase evidence. Against SCIM2-complete HEAD it must fail because:

- Zoom is at `38` executable / `1,804` locally implementable rows, with `21` direct reads and
  `12` direct writes; the Virtual Agent target is `51` / `1,791` / `30` / `16`.
- All thirteen provider paths are absent from the real commandrunner preflight, so a compiled
  `pm zoom virtual-agent …` route remains an `unknown command` before its declaration exists.
- No declaration can turn response-only `page_size`, `next_page_token`, `from`, `to`, or other
  response schema values into an invented request flag.

The RED output will be pasted verbatim below before any production JSON, metadata, fixture, or
generated-file edit.

## RED — captured before connector declaration changes

The test-only RED checkpoint was run against the SCIM2-complete bundle before any Virtual Agent
production declaration, fixture, metadata, or generated-file change. It contains no provider
credential or token value:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
--- FAIL: TestProviderInventoryLedgerIsComplete (0.03s)
    command_surface_test.go:156: executable rows = 38, want 51
    command_surface_test.go:159: operations awaiting Zoom-local contracts = 1804, want 1791
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
    command_surface_test.go:254: reachable direct_read operation commands = 21, want 30
    command_surface_test.go:255: reachable direct_write operation commands = 12, want 16
--- FAIL: TestVirtualAgentOperationCommandsAreReachable (0.03s)
    command_surface_test.go:380: Preflight("virtual-agent knowledge-bases articles list") = connector command "virtual-agent knowledge-bases articles list" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent knowledge-bases articles create") = connector command "virtual-agent knowledge-bases articles create" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent knowledge-bases articles get") = connector command "virtual-agent knowledge-bases articles get" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent knowledge-bases articles update") = connector command "virtual-agent knowledge-bases articles update" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent knowledge-bases articles delete") = connector command "virtual-agent knowledge-bases articles delete" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent knowledge-bases sync create") = connector command "virtual-agent knowledge-bases sync create" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent knowledge-bases sync get") = connector command "virtual-agent knowledge-bases sync get" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent reports engagements list") = connector command "virtual-agent reports engagements list" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent reports engagements query-details list") = connector command "virtual-agent reports engagements query-details list" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent reports engagements variable-details list") = connector command "virtual-agent reports engagements variable-details list" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent reports surveys list") = connector command "virtual-agent reports surveys list" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent reports transcripts list") = connector command "virtual-agent reports transcripts list" is blocked: unknown command, want declared executable Virtual Agent action
    command_surface_test.go:380: Preflight("virtual-agent reports operation-logs list") = connector command "virtual-agent reports operation-logs list" is blocked: unknown command, want declared executable Virtual Agent action
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom
FAIL
```

This is the committed red state. Connector declaration work begins only after it is pushed.

## GREEN connector

The Virtual Agent contract is now declared without a new engine foundation:

- nine `rest_read` actions and four `rest_write` actions bind exactly one published method/path each;
- article create/update expose all and only the documented typed request fields (`content`, `exclude`,
  `title`, plus optional `category`, `external_id`, `language`, and `url`);
- create-sync has no request body; article delete retains destructive typed confirmation and the
  documented `204 No Content` status-only result;
- all response-bearing actions use redacted output and declared provider-sensitive fields;
- `surface-sync` generated `api_surface` / output-policy metadata, then `surface-reconcile` changed
  exactly the thirteen `provider_module=virtual-agent` rows to executable coverage.

The original RED command is now green after the declarations and reconciliation:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
ok      polymetrics.ai/internal/connectors/defs/zoom
```

Focused fixture lifecycle proof also passes:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/... -run '^TestVirtualAgent'
ok      polymetrics.ai/internal/connectors/defs/zoom
```

## Verification/review

No reusable foundation was needed: the normal Zoom `/v2` bearer transport, ordinary typed
path/body mappings, status-only executor, and redacting output policy cover every published
contract. The manually reviewed diff found no raw transport or body escape hatch, no hand-authored
paging input, no uncovered artifact endpoint, and no unscoped endpoint-ledger change.
