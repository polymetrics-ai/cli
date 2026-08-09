# TDD Ledger — closed WebSocket session operation foundation, R1

## RED contract

The inherited consumer test in `internal/connectors/engine/websocket_session_test.go` was committed
at `ae43c153c` before any production WebSocket schema/runtime implementation. It declares only a
fixed connector-relative GET endpoint, a fixed `live-asr` subprotocol, finite byte bounds, and a
closed session-update schema.

The red command is re-run from this foundation branch before any production edit. Its output is
appended verbatim below. Follow-on slices add their focused negative test before corresponding
production code and preserve the same red/green ordering.

## Expected RED

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadAcceptsClosedWebSocketSessionContract$'
--- FAIL: TestBundleLoadAcceptsClosedWebSocketSessionContract (0.00s)
    websocket_session_test.go:50: Load closed WebSocket session operation: load bundle acme: operations.json: /operations/0/websocket: additional property not allowed
FAIL
```

## RED — captured 2026-08-09

The command ran from `feat/3963-zoom-websocket-session-foundation` after the foundation plan and
before any production engine, schema, commandrunner, or generator edit:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadAcceptsClosedWebSocketSessionContract$'
--- FAIL: TestBundleLoadAcceptsClosedWebSocketSessionContract (0.00s)
    websocket_session_test.go:50: Load closed WebSocket session operation: load bundle acme: operations.json: /operations/0/websocket: additional property not allowed
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.695s
FAIL
```

The failure confirms that `websocket` is rejected by the closed operations schema. The output
contains only the synthetic connector name and a schema property; no credential, provider response,
or transcript was read or printed.

## Schema/loader RED — captured 2026-08-09

Before a production schema or loader edit, the foundation added specific unsafe-declaration cases
and ran:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestBundle(LoadAcceptsClosedWebSocketSessionContract|RejectsUnsafeWebSocketSessionContracts)$'
--- FAIL: TestBundleLoadAcceptsClosedWebSocketSessionContract (0.00s)
    websocket_session_test.go:51: Load closed WebSocket session operation: load bundle acme: operations.json: /operations/0/websocket: additional property not allowed
--- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts (0.00s)
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/non_get_upgrade (0.00s)
        websocket_session_test.go:103: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session method must be GET"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/absolute_endpoint (0.00s)
        websocket_session_test.go:103: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session path must be connector-relative"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/empty_subprotocol (0.00s)
        websocket_session_test.go:103: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session requires subprotocol"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/unbounded_frame (0.00s)
        websocket_session_test.go:103: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session max_frame_bytes must be positive"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/open_session_update (0.00s)
        websocket_session_test.go:103: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session session_update_schema must declare additionalProperties false"
FAIL
```

This is deliberately red because the current closed schema rejects every `websocket` block before
the loader can assess its method, path, subprotocol, bounds, or nested schema. The next GREEN
change must make the valid declaration load while preserving each specific rejection.

## Schema/loader RED expansion — captured 2026-08-10

Before production edits, the same focused test was extended to cover a frame bound larger than a
declared session bound and a non-redacted output policy, then rerun:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestBundle(LoadAcceptsClosedWebSocketSessionContract|RejectsUnsafeWebSocketSessionContracts)$'
--- FAIL: TestBundleLoadAcceptsClosedWebSocketSessionContract (0.00s)
    websocket_session_test.go:51: Load closed WebSocket session operation: load bundle acme: operations.json: /operations/0/websocket: additional property not allowed
--- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts (0.00s)
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/non_get_upgrade (0.00s)
        websocket_session_test.go:117: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session method must be GET"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/absolute_endpoint (0.00s)
        websocket_session_test.go:117: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session path must be connector-relative"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/empty_subprotocol (0.00s)
        websocket_session_test.go:117: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session requires subprotocol"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/unbounded_frame (0.00s)
        websocket_session_test.go:117: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session max_frame_bytes must be positive"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/frame_larger_than_session_bound (0.00s)
        websocket_session_test.go:117: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session max_frame_bytes must not exceed max_input_bytes or max_output_bytes"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/open_session_update (0.00s)
        websocket_session_test.go:117: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session session_update_schema must declare additionalProperties false"
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/unredacted_output (0.00s)
        websocket_session_test.go:117: Load unsafe websocket session contract = load bundle acme: operations.json: /operations/0/websocket: additional property not allowed, want error containing "websocket_session requires json_redacted output_policy"
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.701s
FAIL
```

The schema still rejects the declaration at its closed boundary; that is the expected red state.
The added cases lock the output-redaction and finite-bound invariants before the schema recognizes
the operation kind.

## Schema/loader GREEN — captured 2026-08-10

The production slice adds only the closed `websocket_session` kind: its meta-schema, typed bundle
model, one-block discriminator, API-surface coverage annotation, and load-time semantic validator.
It has no transport or command route yet. The focused red suite is green:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestBundle(LoadAcceptsClosedWebSocketSessionContract|RejectsUnsafeWebSocketSessionContracts)$'
ok  	polymetrics.ai/internal/connectors/engine	0.698s
```

The package regression suite is also green:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok  	polymetrics.ai/internal/connectors/engine	4.642s
```

The green validator enforces a rooted connector-relative GET path, exactly one valid protocol token,
positive input/output/frame limits with the frame limit no larger than either session limit,
`json_redacted` output, no mutation or approval escalation, and a compiled recursively closed,
bounded session-update schema. The next slice must begin with a separate transport RED test.

## Upgrade transport RED — captured 2026-08-10

Before adding any client upgrade code, a loopback-only connsdk contract test was added. It requires
one fixed relative GET request, standard WebSocket upgrade headers and acceptance hash, the declared
subprotocol, an authentication marker applied through the existing authenticator seam, a writable
post-101 connection, and refusal of an ordinary or redirect HTTP response.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/connsdk -run '^TestOpenWebSocket'
# polymetrics.ai/internal/connectors/connsdk [polymetrics.ai/internal/connectors/connsdk.test]
internal/connectors/connsdk/websocket_test.go:45:32: undefined: websocketAcceptGUID
internal/connectors/connsdk/websocket_test.go:80:28: requester.OpenWebSocket undefined (type *Requester has no field or method OpenWebSocket)
internal/connectors/connsdk/websocket_test.go:125:45: (&Requester{…}).OpenWebSocket undefined (type *Requester has no field or method OpenWebSocket)
FAIL	polymetrics.ai/internal/connectors/connsdk [build failed]
FAIL
```

The test contains only loopback traffic and a non-secret `X-Connector-Auth: present` marker. It does
not read a provider endpoint, credential, token, transcript, or signed URL. The missing method is
the intended transport-foundation red state.

## Upgrade transport GREEN — captured 2026-08-10

The green slice adds `connsdk.Requester.OpenWebSocket`, a deliberately narrow client bridge. It
uses only the Requester's existing configured base URL, headers, authenticator, and declared
rate-limit admission. It requires a rooted connector-relative path, `GET`, one token-shaped
subprotocol, RFC 6455 key/accept validation, exact selected subprotocol, and `101 Switching
Protocols`; redirects return as terminal non-101 responses and are never followed. It exposes only
the verified read-write upgraded connection to the operation executor.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/connsdk -run '^TestOpenWebSocket'
ok  	polymetrics.ai/internal/connectors/connsdk	0.396s

$ go test -count=1 -timeout 20m ./internal/connectors/connsdk
ok  	polymetrics.ai/internal/connectors/connsdk	0.726s
```

The implementation returns status/protocol failures without provider response content and redacts
transport error text before returning it. It does not add a CLI command, operation executor, raw
frame API, or caller-selected transport control. The next slice starts with an engine frame/session
RED test.

## Frame/session executor RED — captured 2026-08-10

Before adding an engine session runner, a loopback frame test was added. It requires the runner to
validate and send a schema-bound `session.update` text frame, split PCM16 into finite masked binary
frames, send a masked close, answer a server ping, bound server frames, decode only JSON text event
frames, redact protected response fields, and account for the exact bounded input/output. A second
case sends one server frame larger than the declared cap.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationWebSocketSession'
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/websocket_session_runtime_test.go:63:17: undefined: OperationWebSocketSession
internal/connectors/engine/websocket_session_runtime_test.go:63:110: undefined: WebSocketSessionRequest
internal/connectors/engine/websocket_session_runtime_test.go:95:12: undefined: OperationWebSocketSession
internal/connectors/engine/websocket_session_runtime_test.go:95:105: undefined: WebSocketSessionRequest
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

The fixture speaks only loopback WebSocket framing and synthetic PCM16 bytes. It never emits a
provider transcript, credential, authorization value, token-derived value, signed URL, or a raw
transport control.

## Frame/session executor GREEN — captured 2026-08-10

The green runner is an operation-specific `OperationWebSocketSession`, not an exported generic
WebSocket command. It validates the closed session-update schema before opening a connection,
requires non-empty complete PCM16 samples, binds total input to the declaration, uses the declared
surface endpoint only, writes masked finite text/binary/close frames, answers ping with a masked
pong, refuses masked, fragmented, binary, malformed, or oversized server payloads, and accepts
only bounded JSON text events. Events run through `json_redacted`; transport and frame errors run
through the redacted error policy. A context cancellation closes the connection through a joined
cancellation callback, so no session goroutine survives the call.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationWebSocketSession'
ok  	polymetrics.ai/internal/connectors/engine	0.699s

$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok  	polymetrics.ai/internal/connectors/engine	4.640s
```

The next slice begins with command/preflight RED coverage. The runner currently has no connector
interface, command intent, file-path adapter, generated help route, or provider bundle consumer.

## Command boundary and generated-surface RED — captured 2026-08-10

Before adding any production command, connector interface, schema, generator, or validation code,
the foundation added three focused contracts:

1. `TestRunImplementedWebSocketSessionCommand` asks the real command runner to pass only a required
   JSON `session-update` and a project-confined PCM16 file to a typed operation. It also asserts that
   a caller-selected `--subprotocol` remains an unknown flag.
2. `TestCheckAPISurfaceAndCLISurface_AcceptsClosedWebSocketSessionCoverage` requires the checker to
   recognize `covered_by.websocket_session` by its operation ID and an implemented closed session
   command.
3. `TestSyncBundleDerivesClosedWebSocketSessionCommand` requires `surface-sync` to derive the
   operation endpoint, exact redacted policy, and the only two session inputs, rather than relying
   on hand-authored metadata.

The frame-boundary test was also extended before code: `max_frame_bytes: 1` must be rejected because
the fixed runner sends the two-byte normal-close code.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run '^TestRunImplementedWebSocketSessionCommand$'
# polymetrics.ai/internal/connectors/commandrunner [polymetrics.ai/internal/connectors/commandrunner.test]
internal/connectors/commandrunner/runner_test.go:36:42: undefined: connectors.OperationWebSocketSessionRequest
internal/connectors/commandrunner/runner_test.go:38:42: undefined: connectors.OperationWebSocketSessionMetadata
internal/connectors/commandrunner/runner_test.go:193:85: undefined: connectors.OperationWebSocketSessionRequest
internal/connectors/commandrunner/runner_test.go:193:131: undefined: connectors.OperationWebSocketSessionResult
internal/connectors/commandrunner/runner_test.go:196:21: undefined: connectors.OperationWebSocketSessionResult
internal/connectors/commandrunner/runner_test.go:198:20: undefined: connectors.OperationWebSocketSessionResult
internal/connectors/commandrunner/runner_test.go:216:89: undefined: connectors.OperationWebSocketSessionMetadata
internal/connectors/commandrunner/runner_test.go:218:21: undefined: connectors.OperationWebSocketSessionMetadata
internal/connectors/commandrunner/runner_test.go:1933:12: result.WebSocketSession undefined (type Result has no field or method WebSocketSession)
FAIL	polymetrics.ai/internal/connectors/commandrunner [build failed]
FAIL

$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleRejectsUnsafeWebSocketSessionContracts/frame_cannot_hold_required_close_code$'
--- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts (0.00s)
    --- FAIL: TestBundleRejectsUnsafeWebSocketSessionContracts/frame_cannot_hold_required_close_code (0.00s)
        websocket_session_test.go:124: Load unsafe websocket session contract = <nil>, want error containing "websocket_session max_frame_bytes must allow the required close code"
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.702s
FAIL

$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestCheckAPISurfaceAndCLISurface_AcceptsClosedWebSocketSessionCoverage|TestSyncBundleDerivesClosedWebSocketSessionCommand)$'
--- FAIL: TestSyncBundleDerivesClosedWebSocketSessionCommand (0.00s)
    surfacesync_test.go:424: sync stats = {Filled:api_surface=0 output_policy=0 flag_maps_to=0 flag_derived=0 rest.max_bytes=0 Corrected:api_surface=0 output_policy=0 flag_maps_to=0 flag_derived=0 rest.max_bytes=0}, want endpoint/policy/two closed flags
--- FAIL: TestCheckAPISurfaceAndCLISurface_AcceptsClosedWebSocketSessionCoverage (0.00s)
    validate_surface_test.go:87: checkAPISurface rejected closed websocket-session coverage: [{Connector:acme File:api_surface.json Rule:surface_coverage Message:endpoint 0 (GET /live) has no classifier}]
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.700s
FAIL
```

All fixtures are local and synthetic. The command test writes four PCM16 bytes under `t.TempDir()`;
it does not open a provider connection or reveal a credential, token-derived value, transcript,
signed URL, raw authorization header, or audio recording.

## Safety assertions

- No test fixture carries a credential, authorization value, token-derived value, signed URL, or
  live transcript.
- The foundation's test server is loopback-only; it never contacts Zoom or another provider.
- A green result is insufficient unless negative tests still reject caller-controlled transport and
  unbounded inputs.
