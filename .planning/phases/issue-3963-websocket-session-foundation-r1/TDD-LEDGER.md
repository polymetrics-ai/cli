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

## Safety assertions

- No test fixture carries a credential, authorization value, token-derived value, signed URL, or
  live transcript.
- The foundation's test server is loopback-only; it never contacts Zoom or another provider.
- A green result is insufficient unless negative tests still reject caller-controlled transport and
  unbounded inputs.
