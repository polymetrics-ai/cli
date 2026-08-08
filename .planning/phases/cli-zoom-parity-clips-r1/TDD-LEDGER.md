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

### Connectorgen root-array declaration gate — additional RED 2026-08-08

The runtime root-array foundation already admits a named `json_array` command input for a closed
operation body. While authoring the real Clips collaborator command, the declarative validator
still rejected the same safe shape solely because it assumed every root body was an object. The
test was added before the validator change and failed verbatim:

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestValidate_CLISurfaceImplementedDirectWriteNamedRootJSONArrayPasses$'
--- FAIL: TestValidate_CLISurfaceImplementedDirectWriteNamedRootJSONArrayPasses (0.00s)
    main_test.go:599: expected zero findings for named root-array direct_write cli surface, got [{Connector:cli-surface File:operations.json Rule:cli_surface_safety Message:implemented direct write command 0 ("widget archive") operation "cli-surface.widgets.archive" root body mapping requires an object body_schema}]
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.745s
FAIL
```

The fixture contains only an invalid-domain synthetic recipient address and does not load a
credential, issue a request, or permit a generic raw-body route.

### Binary-download surface reconciliation — additional RED 2026-08-08

The source ledger's valid `binary_read` model was still refused by
`surface-reconcile`, even though the commandrunner already admits the bounded,
declared binary executor. The test copies the existing YouTube Analytics
download declaration into a temporary fixture, changes only its endpoint row
to blocked `binary_read`, and requires real runtime preflight before promotion.
It failed before the reconciler change:

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestRunSurfaceReconcileCoversBinaryDownloadWithRuntimePreflight$'
--- FAIL: TestRunSurfaceReconcileCoversBinaryDownloadWithRuntimePreflight (0.01s)
    surfacereconcile_test.go:85: stats = {Scanned:1 Covered:0 Blocked:0 Unchanged:0 Refused:1}, want one runtime-covered binary download
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.740s
FAIL
```

The fixture invokes no network or filesystem download and contains no credential,
token-derived value, or signed URL.

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

### Connectorgen root-array declaration gate — green 2026-08-08

The validator now checks a root body mapping against the declared CLI flag type: a
`json_object` requires a closed object schema and a `json_array` requires a closed
array schema. This preserves the existing no-generic-body boundary while admitting
the documented Clips collaborator array operation.

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run 'TestValidate_CLISurfaceImplementedDirectWrite(NamedRootJSONArrayPasses|RejectsRootArrayBodySchema|.*RootBody)'
ok  	polymetrics.ai/cmd/connectorgen	0.765s
```

### Binary-download surface reconciliation — green 2026-08-08

`surface-reconcile` now handles the valid `binary_read` ledger model. It finds
only matching `binary_download` commands, invokes the real runtime preflight,
and records the admitted command with the existing direct-read coverage form.
It never invokes the downloader or supplies a destination, so reconciliation
cannot perform a network request or write a file.

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestRunSurfaceReconcileCoversBinaryDownloadWithRuntimePreflight|TestSurfaceReconcileKeepsUnreachableRowsBlockedAndRefusesUnknownModel|TestRunSurfaceReconcileHelp)$'
ok  	polymetrics.ai/cmd/connectorgen	0.777s

$ go test -count=1 -timeout 20m ./cmd/connectorgen
ok  	polymetrics.ai/cmd/connectorgen	11.004s
```

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

### Operation-level base64 path upload — green 2026-08-08

Added the closed `rest.base64_upload` contract for an operation body schema. Its required local
source field is exposed to the planner as a file identity, bounded/read under the project root,
checked against declaration-owned filename and sniffed-media policies, verified against the
planner-provided SHA-256, omitted from preview and wire JSON, then converted to canonical base64
only at the typed execution boundary. A file changed after preview fails before any network call.

The existing declaration-owned mutation redirect may now admit this snapshot-bound JSON body as
well as multipart, but still requires one fixed bearer-authenticated base URL, a literal
same-provider suffix boundary, and a finite 307/308-only hop cap. It does not introduce a generic
redirect or raw body option.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestOperationDirectWriteBase64PathUploadIsPreviewBound|TestOperationBase64UploadAdmitsDeclaredMutationRedirect|TestOperationDirectWriteBase64UploadFollowsDeclaredRedirect|TestWriteBase64UploadRejectsDeclaredFilePolicy)$'
ok  	polymetrics.ai/internal/connectors/engine

$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok  	polymetrics.ai/internal/connectors/engine

$ go test -count=1 -timeout 20m ./internal/connectors/connsdk
ok  	polymetrics.ai/internal/connectors/connsdk

$ go test -count=1 -timeout 20m ./cmd/connectorgen
ok  	polymetrics.ai/cmd/connectorgen
```

All fixtures are local and synthetic; no credential, token-derived value, or signed redirect URL
was emitted.

### JSON multipart-event mutation redirect — additional RED 2026-08-08

The live Clips artifact says the JSON `CreateMultipartUpload` and
`CompleteMultipartUpload` event calls must follow an HTTP 30x and retain bearer
authorization at the admitted provider host. The existing transport already
rebuilds a preview-bound JSON body for a declared mutation redirect, but the
bundle semantic gate admitted only multipart or base64 upload declarations.
The focused test failed verbatim before that narrow declaration rule was
extended:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationJSONWriteAdmitsDeclaredMutationRedirect$'
--- FAIL: TestOperationJSONWriteAdmitsDeclaredMutationRedirect (0.00s)
    direct_write_test.go:505: closed JSON operation declared redirect = operation 0 ("zoom.clips.files.multipart_upload_events") rest.redirect is only valid for a multipart or base64_upload rest_write, want admitted provider-owned boundary
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.759s
FAIL

$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestOperationJSONWriteAdmitsDeclaredMutationRedirect|TestOperationDirectWriteJSONFollowsDeclaredRedirect)$'
--- FAIL: TestOperationJSONWriteAdmitsDeclaredMutationRedirect (0.00s)
    direct_write_test.go:505: closed JSON operation declared redirect = operation 0 ("zoom.clips.files.multipart_upload_events") rest.redirect is only valid for a multipart or base64_upload rest_write, want admitted provider-owned boundary
--- FAIL: TestOperationDirectWriteJSONFollowsDeclaredRedirect (0.00s)
    direct_write_test.go:594: PreviewOperationDirectWrite declared JSON redirect: operation 0 ("acme.clips.multipart_upload_events.create") rest.redirect is only valid for a multipart or base64_upload rest_write
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.732s
FAIL
```

The fixture names Zoom's documented operation but uses only a synthetic secret
template and issues no request. The planned green constraint remains strict:
fixed literal base URL, exactly one declared bearer auth, finite same-provider
suffix boundary, 307/308 only, preview-bound closed JSON body, and no
caller-supplied redirect target or credential header.

### JSON multipart-event closed-body guard — additional RED after rebase 2026-08-08

The rebase checkpoint retained the known JSON-event redirect gap. Before the
semantic admission changed, a new negative test proved that the narrow
exception must recursively close every object in the JSON body: an open nested
`params` object could otherwise become a caller-shaped generic mutation. The
focused test failed verbatim because the prior rule did not yet distinguish a
closed JSON event from an open one:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestOperationJSONWriteAdmitsDeclaredMutationRedirect|TestOperationJSONMutationRedirectRequiresClosedBody|TestOperationDirectWriteJSONFollowsDeclaredRedirect)$'
--- FAIL: TestOperationJSONWriteAdmitsDeclaredMutationRedirect (0.00s)
    direct_write_test.go:505: closed JSON operation declared redirect = operation 0 ("zoom.clips.files.multipart_upload_events") rest.redirect is only valid for a multipart or base64_upload rest_write, want admitted provider-owned boundary
--- FAIL: TestOperationJSONMutationRedirectRequiresClosedBody (0.00s)
    direct_write_test.go:528: open JSON mutation redirect = operation 0 ("zoom.clips.files.multipart_upload_events.open_body") rest.redirect is only valid for a multipart or base64_upload rest_write, want nested closed-body rejection
--- FAIL: TestOperationDirectWriteJSONFollowsDeclaredRedirect (0.00s)
    direct_write_test.go:617: PreviewOperationDirectWrite declared JSON redirect: operation 0 ("acme.clips.multipart_upload_events.create") rest.redirect is only valid for a multipart or base64_upload rest_write
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.777s
FAIL
```

The tests use synthetic local endpoints and secret templates only. No request
was sent in this RED state.

### JSON multipart-event mutation redirect — green after rebase 2026-08-08

The mutation redirect declaration now accepts only a `rest_write` that is
already a multipart/base64 upload or declares `application/json` with a
recursively closed object body schema. The JSON path requires the same fixed
literal operation base URL, one declared bearer authenticator, finite provider
suffix, and 307/308-only policy as the existing upload path. Its target URL
and authorization header remain declaration-owned. An open nested object (and
any open `allOf`/`anyOf`/`oneOf` branch) is refused before a command can reach
the transport.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestOperationJSONWriteAdmitsDeclaredMutationRedirect|TestOperationJSONMutationRedirectRequiresClosedBody|TestOperationDirectWriteJSONFollowsDeclaredRedirect)$'
ok  polymetrics.ai/internal/connectors/engine

$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok  polymetrics.ai/internal/connectors/engine

$ go test -count=1 -timeout 20m ./internal/connectors/connsdk
ok  polymetrics.ai/internal/connectors/connsdk

$ go test -count=1 -timeout 20m ./internal/connectors/commandrunner
ok  polymetrics.ai/internal/connectors/commandrunner

$ go test -count=1 -timeout 20m ./cmd/connectorgen
ok  polymetrics.ai/cmd/connectorgen

$ go run ./cmd/connectorgen validate internal/connectors/defs/zoom
connectorgen validate: 1 connector(s) checked, 0 findings
```

During the rebase check, two pre-existing `noReplayClient` test callers were
updated to pass `false` for the helper's current declared-redirect argument.
They exercise the no-declared-redirect branch, so this preserves the original
non-replay behavior while keeping the compatibility check buildable.

## GREEN connector — pending

Record the real runner fixture lifecycle, source reconciliation, and command reachability here
after all 21 endpoint rows are covered.

### Clip transfer cardinality correction — RED 2026-08-08

The live artifact's partial-transfer payload requires a non-empty `clip_id_list` as well as its
maximum of fifty. The category surface test was extended before correcting either the CLI flag or
the closed operation schema, and failed verbatim:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^TestClipsOperationCommandsAreReachable$'
--- FAIL: TestClipsOperationCommandsAreReachable (0.13s)
    command_surface_test.go:676: Clips command "clips transfers partial" clip-id-list min_items = 0, want 1
    command_surface_test.go:698: Clips operation "zoom.transfer_clips_partial" clip_id_list minItems = 0, want 1
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/zoom	0.900s
FAIL
```

This is a test-only checkpoint; the artifact and fixture use no credential or token-derived value.

### Legacy reverse-ETL record compatibility — RED 2026-08-09

The root-array foundation widened the operation-body helper, but initially treated a legacy
`connectors.Record` dynamic value as if it were not an object. The existing record-only
connector-plan regression was strengthened to assert that legacy reverse-ETL plans still leave
the new operation-body field empty, then failed before the compatibility repair:

```text
$ go test -count=1 -timeout 20m ./internal/app -run '^TestPlanConnectorCommandPersistsCompleteDeclaredContent$'
--- FAIL: TestPlanConnectorCommandPersistsCompleteDeclaredContent (1.36s)
    connector_command_content_test.go:48: PlanConnectorCommand: connector command record must be a JSON object
FAIL
FAIL	polymetrics.ai/internal/app	2.154s
FAIL
```

The regression fixture contains synthetic content only and no credential or token-derived value.

## Verification/review — pending

Record scoped verification and inline manual review findings here after green.
