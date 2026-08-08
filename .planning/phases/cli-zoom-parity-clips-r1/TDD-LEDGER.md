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

## GREEN foundations

### Root JSON-array direct writes — green 2026-08-08

Implemented the closed `json_array` flag type only for declared operation-body mappings. The
typed value now survives command shaping, plan persistence, preview hashing, typed approval, and
execution; it cannot map to a path, query, or arbitrary generic request body. Existing persisted
object-body plans continue to fall back to `connector_command_record`.

```text
$ go test -count=1 -timeout 20m ./internal/app -run 'TestDirectWriteCommandPlanPreviewApprovalAndExecute|TestDirectWriteCommandRootJSONArrayPlanPreviewApprovalAndExecute'
ok  polymetrics.ai/internal/app

$ go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run '^TestCoerceFlagValueAcceptsJSONArray$'
ok  polymetrics.ai/internal/connectors/commandrunner
```

The root-array lifecycle test proves a declared `json_array` reaches one POST only after
plan/preview/typed confirmation, stores the real body separately from the legacy record shape,
and exposes only a redacted plan sample. Object-body lifecycle regression remains green.

Remaining foundations: declared bearer redirect; operation-level bounded base64 path upload.

### Declared bearer binary redirect — green 2026-08-08

Added a separate declaration-owned `binary.bearer_redirect` policy. It validates a finite
provider suffix boundary, requires an unchanged scheme and an admitted original host, strips all
credential/default headers, then restores only the original non-empty `Bearer` authorization. It
is mutually exclusive with generic `allow_cross_host` / `allowed_hosts`; the binary executor also
requires exactly one declared bearer authenticator before enabling it.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/connsdk -run 'TestDoStreamDeclaredBearerRedirectRetainsAuthorization|TestDoStreamDeclaredBearerRedirectStripsCustomCredentials'
ok  polymetrics.ai/internal/connectors/connsdk

$ go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestBinaryDownloadDeclaredBearerRedirectRetainsOnlyBearer|TestBinaryDownloadDeclaredBearerRedirectRequiresBearerAuth'
ok  polymetrics.ai/internal/connectors/engine
```

The transport test proves the retained bearer reaches only an allowed suffix, while custom auth is
refused before the target receives a request. The executor test proves that the declarative binary
operation, rather than an ad-hoc caller option, selects the narrow policy.

Remaining foundation: operation-level bounded base64 path upload.

### Operation-level base64 path upload — additional RED 2026-08-08

The original RED fixture was strengthened to declare the exact closed operation-level path
contract: a required source field, bounded PNG/JPEG/GIF source policy, a plan-provided payload
digest, no local path/base64 bytes in preview, and a changed-file rejection before network
dispatch. Before the missing canonical preview path was implemented, the focused command failed
verbatim:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteBase64PathUploadIsPreviewBound$'
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/direct_write.go:455:30: undefined: prepareCanonicalOperationBase64Upload
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

The fixture uses a synthetic one-pixel image and contains no credential, token-derived value, or
signed URL.

### Base64 mutation redirect — additional RED 2026-08-08

The source artifact requires the temporary image endpoint to retain bearer authentication on a
provider-owned 30x hop. The existing declared-mutation redirect transport can safely replay its
bounded JSON body, but bundle validation admitted only multipart contracts. The focused semantic
test failed verbatim before that narrow admission was extended:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationBase64UploadAdmitsDeclaredMutationRedirect$'
--- FAIL: TestOperationBase64UploadAdmitsDeclaredMutationRedirect (0.00s)
    direct_write_test.go:482: base64 operation declared redirect = operation 0 ("zoom.clips.files.temporary_upload") rest.redirect is only valid for a multipart rest_write, want admitted provider-owned boundary
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.750s
FAIL
```

The declaration contains only a synthetic secret template and performs no request.

## GREEN connector — pending

Record the real runner fixture lifecycle, source reconciliation, and command reachability here
after all 21 endpoint rows are covered.

## Verification/review — pending

Record scoped verification and inline manual review findings here after green.
