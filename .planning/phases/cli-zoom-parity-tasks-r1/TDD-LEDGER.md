# TDD Ledger — Zoom Tasks documented-operation parity, R1

## Planned RED contract

Before any Tasks connector declaration or redirect-foundation production change, the test-only RED
checkpoint must prove both gaps:

- Auto-Dialer-complete HEAD has `67` executable / `1,775` Zoom-local rows, `38` direct reads, and
  `24` direct writes; this category's target is `84` / `1,758` / `44` / `35`.
- All 17 real `tasks …` paths are unknown to the command runner before their declarations exist.
- The published multipart file upload cannot be executed correctly because the current strict
  direct-write transport rejects every 30x redirect even when a closed declaration would bound it
  to a Zoom-owned HTTPS host and reapply declared bearer authentication.

RED output will be recorded verbatim below before production changes.

## RED — captured before connector declaration or foundation changes

The test-only RED checkpoint ran against Tasks-unimplemented HEAD. It contains no provider
credential or token value:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/... -run 'TestProviderInventoryLedgerIsComplete|TestCoveredStreamsHaveReachableCommands|TestTasksOperationCommandsAreReachable'
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:157: executable rows = 67, want 84
    command_surface_test.go:160: operations awaiting Zoom-local contracts = 1775, want 1758
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.04s)
    command_surface_test.go:255: reachable direct_read operation commands = 38, want 44
    command_surface_test.go:256: reachable direct_write operation commands = 24, want 35
--- FAIL: TestTasksOperationCommandsAreReachable (0.03s)
    command_surface_test.go:496: Preflight("tasks assignees list") = connector command "tasks assignees list" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks assignees add") = connector command "tasks assignees add" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks assignees remove") = connector command "tasks assignees remove" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks collaborators list") = connector command "tasks collaborators list" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks collaborators add") = connector command "tasks collaborators add" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks collaborators remove") = connector command "tasks collaborators remove" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks comments list") = connector command "tasks comments list" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks comments add") = connector command "tasks comments add" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks comments delete") = connector command "tasks comments delete" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks files upload") = connector command "tasks files upload" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks imports submit") = connector command "tasks imports submit" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks imports get") = connector command "tasks imports get" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks items list") = connector command "tasks items list" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks items create") = connector command "tasks items create" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks items get") = connector command "tasks items get" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks items delete") = connector command "tasks items delete" is blocked: unknown command, want declared executable Tasks action
    command_surface_test.go:496: Preflight("tasks items update") = connector command "tasks items update" is blocked: unknown command, want declared executable Tasks action
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom
FAIL
```

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteMultipartFollowsDeclaredRedirect$'
--- FAIL: TestOperationDirectWriteMultipartFollowsDeclaredRedirect (0.00s)
    operation_multipart_test.go:141: Load declared multipart redirect contract: load bundle acme: operations.json: /operations/0/rest/redirect: additional property not allowed
FAIL
FAIL    polymetrics.ai/internal/connectors/engine
FAIL
```

Both commands exited `1`, exactly as expected. This is the committed red state; only the red tests
and phase evidence changed before it was captured.

## GREEN foundation — captured before Tasks connector authoring

The redirect foundation was implemented separately from any Zoom Tasks bundle declaration. It is
closed to operation-level `rest_write` multipart uploads with one literal base URL, one declared
bearer authenticator, a non-wildcard provider DNS-suffix boundary that includes the base host, and
one through three same-scheme `307`/`308` hops. The execution path rebuilds the preview-bound
multipart snapshot and reapplies the declared bearer for each admitted hop; ordinary strict direct
writes continue to reject redirects. Redirect target query/fragment/userinfo cannot appear in a
returned error or response URL.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectWriteMultipart|TestBundleLoad.*Multipart'
ok      polymetrics.ai/internal/connectors/engine  0.755s

$ go test -count=1 -timeout 20m ./internal/connectors/connsdk ./internal/connectors/commandrunner
ok      polymetrics.ai/internal/connectors/connsdk          0.724s
ok      polymetrics.ai/internal/connectors/commandrunner    7.053s
```

These GREEN tests cover bearer retention, multipart field/file replay, fixed-base and bearer-only
load validation, suffix/wildcard/IP rejection, source-base containment, status/method refusal,
hop caps, and signed redirect-value redaction. They also exercise the existing command runner so
the added requester field does not change ordinary strict-write preflight behavior.

## GREEN connector — pending

## Verification/review — pending
