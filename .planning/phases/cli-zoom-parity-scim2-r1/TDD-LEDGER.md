# TDD Ledger — Zoom SCIM2 documented-operation parity, R1

## Planned RED contract

Before any production engine or Zoom bundle change, the RED checkpoint will contain only tests and
planning evidence. It must fail against the current branch because:

- Zoom remains at `27` executable / `1,815` local implementable rows, with `17` direct reads and
  `5` direct writes; the target requires `38` / `1,804` / `21` / `12`.
- All eleven SCIM2 paths are absent from the real commandrunner preflight, so a compiled `pm zoom
  scim2 …` route cannot yet resolve.
- A declared named `json_object` input targeted at exact root `body` is rejected, despite SCIM's
  documented extensible resource and PatchOp bodies. The existing generic `json` type remains
  deliberately unsupported.
- A declared rest-read operation still uses the ordinary bundle origin/auth and therefore cannot
  honor a provider root endpoint such as `/scim2/...` independently of ordinary `/v2` calls.

The RED run occurred before any production change. It contains no provider credential or token
value:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen ./internal/connectors/defs/zoom/...
--- FAIL: TestOperationDirectReadUsesDeclaredOperationOriginAndAuth (0.00s)
    direct_read_test.go:643: Load declared operation-scoped direct-read origin/auth bundle: load bundle acme: operations.json: operation 0 ("acme.list_scim2_groups") rest.base_url/rest.auth are only valid for rest_write operations, got "rest_read"
FAIL
FAIL    polymetrics.ai/internal/connectors/engine
--- FAIL: TestBuildOperationDirectWriteCommandMapsNamedRootJSONObject (0.00s)
    runner_test.go:1905: BuildWriteCommand named root object: connector command "scim groups create" is blocked: intent=direct_write: availability=implemented: flag --resource maps to unsupported target "body"
FAIL
FAIL    polymetrics.ai/internal/connectors/commandrunner
--- FAIL: TestValidate_CLISurfaceImplementedDirectWriteNamedRootJSONObjectPasses (0.00s)
    main_test.go:551: expected zero findings for named root-object direct_write cli surface, got [{Connector:cli-surface File:cli_surface.json Rule:cli_surface_safety Message:implemented direct write command 0 ("widget archive") flag --resource maps to unsupported target "body"} {Connector:cli-surface File:cli_surface.json Rule:cli_surface_safety Message:command 0 ("widget archive") flag --resource json_object type may map only to an operation body}]
FAIL
FAIL    polymetrics.ai/cmd/connectorgen
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:155: executable rows = 27, want 38
    command_surface_test.go:158: operations awaiting Zoom-local contracts = 1815, want 1804
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
    command_surface_test.go:253: reachable direct_read operation commands = 17, want 21
    command_surface_test.go:254: reachable direct_write operation commands = 5, want 12
--- FAIL: TestSCIM2OperationCommandsAreReachable (0.03s)
    command_surface_test.go:325: Preflight("scim2 groups list") = connector command "scim2 groups list" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 groups create") = connector command "scim2 groups create" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 groups get") = connector command "scim2 groups get" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 groups delete") = connector command "scim2 groups delete" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 groups update") = connector command "scim2 groups update" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 users list") = connector command "scim2 users list" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 users create") = connector command "scim2 users create" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 users get") = connector command "scim2 users get" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 users update") = connector command "scim2 users update" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 users delete") = connector command "scim2 users delete" is blocked: unknown command, want declared executable SCIM2 action
    command_surface_test.go:325: Preflight("scim2 users deactivate") = connector command "scim2 users deactivate" is blocked: unknown command, want declared executable SCIM2 action
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom
FAIL
```

The test-only RED checkpoint will be committed and pushed before any production foundation or
connector change.

## Planned GREEN foundation — operation-scoped direct-read origin/auth

The foundation will extend the existing paired `rest.base_url` / `rest.auth` transport contract to
bounded direct reads. A rest-read with the override must issue only to its declared origin, use only
its declared auth, clear unrelated global headers, and retain the fixed relative endpoint and
response cap. It is necessary for SCIM2 because Zoom documents the API server at the host root while
ordinary Zoom operations use `/v2`; it also unblocks similarly documented distinct-origin reads.

## GREEN foundation — operation-scoped direct-read origin/auth

This foundation is now implemented before any SCIM2 bundle declaration. `rest.base_url` and
`rest.auth` remain paired declarations; `rest_read` joins `rest_write` as the only operation kind
that may use them. The direct-read executor selects the operation's origin and auth together,
clears ordinary connector headers before constructing the requester, and resolves the exact fixed
path against the declared base. A programmatically constructed bundle receives the same paired
declaration guard as a loaded one.

The RED test now loads a declared read with an ordinary API server and an independent SCIM2 server.
It proves the latter receives exactly one GET with its own Bearer auth and receives no inherited
ordinary secret header; the ordinary server receives no request.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectReadUsesDeclaredOperationOriginAndAuth$'
ok      polymetrics.ai/internal/connectors/engine

$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok      polymetrics.ai/internal/connectors/engine
```

This standalone foundation unblocks any documented bounded `rest_read` whose provider origin or
base path differs from its connector's ordinary API. It introduces no generic URL input: both the
base template and auth remain operation declarations reviewed in the bundle.

## Planned GREEN foundation — named root JSON-object body

The foundation will permit exact `maps_to: "body"` only for a named `json_object` flag on an
operation-declared object body schema. It will reject path/query mappings, all generic `json` input,
and mixed root/field body mappings. The declared operation remains responsible for schema and byte
limits. This is needed for Zoom SCIM2's documented extensible User/Group/PatchOp object bodies,
including custom extension attributes, without creating raw transport capability.

## GREEN foundation — named root JSON-object body

This foundation is now implemented before any SCIM2 bundle declaration. Exact `maps_to: "body"`
is legal only for one named `json_object` flag on a `direct_write` command. Runtime shaping rejects
root/path/query misuse, non-object types, and any mix of root and dotted body mappings. Static
validation mirrors that closed contract, requires an operation body schema, and treats a required
root object as covering each schema-required member. The existing generic `json` flag remains
rejected.

The RED tests now prove both the real plan command shaper and `connectorgen` accept only the named
root object contract:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/commandrunner ./cmd/connectorgen
ok      polymetrics.ai/internal/connectors/commandrunner
ok      polymetrics.ai/cmd/connectorgen
```

This standalone foundation unblocks declared provider resource objects whose published schema is
extensible (including custom fields) while retaining a fixed operation, fixed endpoint, typed plan
lifecycle, operation body-schema validation, and request cap. It does not create a raw JSON or raw
HTTP command.

## Planned GREEN connector contract

- Four SCIM2 reads use bounded `json_redacted` output with declared PII/account field redaction.
- Seven SCIM2 mutations use declared typed direct writes and the plan lifecycle; both DELETEs use
  destructive confirmation, and 204 status-only actions return `none` rather than an invented body.
- All group/user resource and patch inputs are synthetically fixture-tested for exact method, root
  path, declared Bearer header, JSON body, redaction, and no paging input.
- The ledger changes only the eleven `provider_module=scim2` endpoints; zero Zoom rows become
  `unsafe_or_disallowed`.
