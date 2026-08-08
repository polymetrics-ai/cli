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

## RED JSON-file foundation — captured before JSON validation production changes

The published upload supports only `.json` files. A typed declaration must not pretend that MIME
sniffing verifies JSON: Go reports ordinary JSON as text/plain. The new test therefore declared a
bounded `content_validation: "json"` and `.json` extension list, then proved the existing schema
cannot express either constraint before production code changed.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteMultipartValidatesDeclaredJSONFile$'
--- FAIL: TestOperationDirectWriteMultipartValidatesDeclaredJSONFile (0.00s)
    operation_multipart_test.go:113: Load declared JSON multipart contract: load bundle acme: operations.json: /operations/0/rest/multipart/parts/1/allowed_file_extensions: additional property not allowed
FAIL
FAIL    polymetrics.ai/internal/connectors/engine  0.748s
FAIL
```

The command exited `1`, as required. This red checkpoint contains only its test and evidence; no
multipart JSON-file implementation has been added yet.

## GREEN JSON-file foundation — captured before Tasks connector authoring

The separate foundation adds only declaration-owned `content_validation: "json"` and
`allowed_file_extensions` file-part constraints. It validates a single JSON document from the
approved bounded snapshot and rejects an extension mismatch before a request. JSON validation
requires a positive file or aggregate upload cap; no caller can select a file type, validator, or
header. The exact green run was:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/connsdk ./internal/connectors/commandrunner ./cmd/connectorgen
ok      polymetrics.ai/internal/connectors/engine           6.769s
ok      polymetrics.ai/internal/connectors/connsdk          0.957s
ok      polymetrics.ai/internal/connectors/commandrunner    7.461s
ok      polymetrics.ai/cmd/connectorgen                    13.539s
```

The JSON-file test proves a valid `.json` source reaches a loopback endpoint with an
`application/json` part header, while a `.txt` name or malformed JSON produces a local error and
no second request. Constraint tests cover extension-policy syntax, file-only use, JSON header
compatibility, and bounded-snapshot enforcement.

## GREEN connector — captured

All seventeen source operations are now declared through the real command runner: six bounded
`rest_read` / `direct_read` operations and eleven approval-gated `rest_write` / `direct_write`
operations. The declaration uses the source's exact methods and paths, contains no derived paging
flags, and maps all 17 Tasks endpoint rows directly. The four DELETEs and the task PATCH assert
their documented `204 No Content` status-only response. The update body is deliberately limited to
the seven live-artifact fields: `description`, `due_date`, `is_public`, `priority`, `starred`,
`status`, and `title`.

The ordinary connector GREEN run used six isolated direct-read fixtures and eleven isolated
direct-write fixtures. The upload fixture exercises its declared JSON-only file, initial
`fileapi.zoom.us` request, admitted `307` redirect to a Zoom-owned HTTPS host, rebuilt multipart
body, re-applied bearer authentication, and redacted response. Fixture values use `.invalid`
identifiers and contain no credentials.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
ok      polymetrics.ai/internal/connectors/defs/zoom  15.163s

$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/... ./internal/connectors/engine ./internal/connectors/connsdk ./internal/connectors/commandrunner ./cmd/connectorgen
ok      polymetrics.ai/internal/connectors/defs/zoom          15.163s
ok      polymetrics.ai/internal/connectors/engine               6.464s
ok      polymetrics.ai/internal/connectors/connsdk              0.808s
ok      polymetrics.ai/internal/connectors/commandrunner       10.354s
ok      polymetrics.ai/cmd/connectorgen                        13.898s

$ go test -count=1 -timeout 20m ./internal/cli
ok      polymetrics.ai/internal/cli
```

The compiled binary, not just generated metadata, accepted the provider base, bare namespace, and
every Tasks help route with exit status `0`:

```text
$ go build -o .tmp/pm ./cmd/pm
$ .tmp/pm help zoom
$ .tmp/pm zoom
$ .tmp/pm zoom tasks
$ .tmp/pm zoom tasks assignees list --help
$ .tmp/pm zoom tasks assignees add --help
$ .tmp/pm zoom tasks assignees remove --help
$ .tmp/pm zoom tasks collaborators list --help
$ .tmp/pm zoom tasks collaborators add --help
$ .tmp/pm zoom tasks collaborators remove --help
$ .tmp/pm zoom tasks comments list --help
$ .tmp/pm zoom tasks comments add --help
$ .tmp/pm zoom tasks comments delete --help
$ .tmp/pm zoom tasks files upload --help
$ .tmp/pm zoom tasks imports submit --help
$ .tmp/pm zoom tasks imports get --help
$ .tmp/pm zoom tasks items list --help
$ .tmp/pm zoom tasks items create --help
$ .tmp/pm zoom tasks items get --help
$ .tmp/pm zoom tasks items delete --help
$ .tmp/pm zoom tasks items update --help
```

Each command exited `0`; no credential was loaded or printed in this reachability check.

## Verification/review — captured

The generated derived surface was refreshed rather than hand-edited. Reconciliation reported
`zoom: reconciled covered=17 blocked=0 unchanged=0 refused=0`; the only endpoint-ledger additions
are six Zoom Tasks direct-read entries. The regenerated documentation and website data were
checked to retain Zoom-only output; non-Zoom website portions remained byte-identical to HEAD.

```text
$ go run ./cmd/connectorgen validate
connectorgen validate: 551 connector(s) checked, 0 finding(s)

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled, 0 finding(s)

$ go run ./cmd/connectorgen surface-reconcile --check --notes-contains provider_module=tasks
zoom: reconciled covered=17 blocked=0 unchanged=0 refused=0

$ go vet ./internal/connectors/defs/zoom ./internal/connectors/engine ./internal/connectors/connsdk ./internal/connectors/commandrunner ./cmd/connectorgen
$ make lint
golangci-lint ...
0 issues.
$ make tidy-check && make docs-check && make smoke-no-build && make agent-contract-check
$ make connectorgen-validate && make connectorgen-surface-sync
$ make connector-boundary && make release-workflow-check
$ git diff --check
```

All commands above exited `0`. Inline manual `verify-work` and `code-review` examined the source
method/path/body contracts, output redaction/status-only behavior, destructive approval routes,
multipart redirect boundary, generated-only files, fixture execution, and endpoint scope. No
actionable finding remained.
