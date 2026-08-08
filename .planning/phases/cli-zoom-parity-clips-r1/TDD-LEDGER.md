# TDD Ledger — Zoom Clips documented-operation parity, R1

## Planned RED contract

Before any Clips connector declaration or foundation production change, the test-only checkpoint
must prove all of the following:

- Workforce-Management-complete HEAD has `102` executable / `1,740` Zoom-local rows, `55` direct
  reads, and `42` direct writes. This category's endpoint target is `123` / `1,719` / `61` / `58`.
- All twenty-one real `clips …` source paths are unknown through `commandrunner.Preflight`; the
  source's concrete multipart-event and transfer variants are also absent as commands.
- A root JSON array cannot currently be used as a closed direct-write body.
- A permitted binary cross-host redirect strips bearer authorization, and no declared provider
  suffix exception exists.
- A direct rest write cannot transform a preview-bound local image path into the source-required,
  bounded base64 JSON field.

RED output will be recorded verbatim below before production changes. No provider credential,
token, or raw signed URL may appear in this ledger.

## RED — captured 2026-08-08

Commands run before any production foundation or Clips-bundle change:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/... -run 'TestProviderInventoryLedgerIsComplete|TestCoveredStreamsHaveReachableCommands|TestClipsOperationCommandsAreReachable'
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:165: executable rows = 102, want 123
    command_surface_test.go:168: operations awaiting Zoom-local contracts = 1740, want 1719
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.04s)
    command_surface_test.go:263: reachable direct_read operation commands = 55, want 61
    command_surface_test.go:264: reachable direct_write operation commands = 42, want 58
--- FAIL: TestClipsOperationCommandsAreReachable (0.04s)
    command_surface_test.go:626: Preflight("clips list") = connector command "clips list" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips starred set") = connector command "clips starred set" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips collaborators list") = connector command "clips collaborators list" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips collaborators share") = connector command "clips collaborators share" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips collaborators remove") = connector command "clips collaborators remove" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips comments list") = connector command "clips comments list" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips comments delete") = connector command "clips comments delete" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips download") = connector command "clips download" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips get") = connector command "clips get" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips delete") = connector command "clips delete" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips update") = connector command "clips update" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips chapters get") = connector command "clips chapters get" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips chapters create") = connector command "clips chapters create" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips duplicate") = connector command "clips duplicate" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips share-settings update") = connector command "clips share-settings update" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips transfers partial") = connector command "clips transfers partial" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips transfers full") = connector command "clips transfers full" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips transfers get") = connector command "clips transfers get" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips files upload") = connector command "clips files upload" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips files multipart upload") = connector command "clips files multipart upload" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips files multipart initiate") = connector command "clips files multipart initiate" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips files multipart complete") = connector command "clips files multipart complete" is blocked: unknown command, want declared executable Clips action
    command_surface_test.go:626: Preflight("clips files temporary upload") = connector command "clips files temporary upload" is blocked: unknown command, want declared executable Clips action
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/zoom	0.880s
FAIL

$ go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run '^TestCoerceFlagValueAcceptsJSONArray$'
--- FAIL: TestCoerceFlagValueAcceptsJSONArray (0.00s)
    runner_test.go:1285: coerce declared json_array: connector command "unknown" is blocked: flag --collaborators has unsupported type "json_array"
FAIL
FAIL	polymetrics.ai/internal/connectors/commandrunner	0.757s
FAIL

$ go test -count=1 -timeout 20m ./internal/connectors/connsdk -run '^TestDoStreamDeclaredBearerRedirectRetainsAuthorization$'
--- FAIL: TestDoStreamDeclaredBearerRedirectRetainsAuthorization (0.00s)
    stream_test.go:277: declared provider redirect Authorization = "", want retained bearer
FAIL
FAIL	polymetrics.ai/internal/connectors/connsdk	0.192s
FAIL

$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteBase64PathUploadIsPreviewBound$'
--- FAIL: TestOperationDirectWriteBase64PathUploadIsPreviewBound (0.00s)
    direct_write_test.go:419: temporary-file base64 = "", want canonical local image payload
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.730s
FAIL
```

The fixtures use synthetic values only; no credential, token-derived value, or signed URL was
emitted.

## GREEN foundations — pending

Record each independent green foundation command and its safety regressions here before connector
authoring.

## GREEN connector — pending

Record the real runner fixture lifecycle, source reconciliation, and command reachability here
after all 21 endpoint rows are covered.

## Verification/review — pending

Record scoped verification and inline manual review findings here after green.
