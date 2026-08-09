# TDD Ledger — Zoom AI Services documented-operation parity, R1

## Planned RED contract

Before any AI Services production declaration or shared protocol implementation:

- CRC-complete HEAD has `143` executable / `1,699` Zoom-local rows, `70` conventional direct
  reads, and `69` direct writes. AI Services targets `165` / `1,677` / `82` / `78`, plus one
  fixed WebSocket session operation.
- Every source-defined `ai-services …` command is unknown through the real
  `commandrunner.Preflight` path.
- The test names all twenty-two operation IDs, source method/path pairs, intents, endpoint routes,
  output policies, the three destructive cancellations, and the one `101` WebSocket endpoint.
- No test fixture or failure output includes a credential, token-derived value, signed URL,
  private storage credential, webhook secret, or transcript.

The RED result is appended verbatim below before production code or bundle declarations change.

## RED — captured 2026-08-08

The command ran after the plan/test-only checkpoint and before any AI Services
production declaration or shared protocol change:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run 'TestProviderInventoryLedgerIsComplete|TestCoveredStreamsHaveReachableCommands|TestAIServicesOperationCommandsAreReachable'
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
    command_surface_test.go:171: executable rows = 143, want 165
    command_surface_test.go:174: operations awaiting Zoom-local contracts = 1699, want 1677
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.05s)
    command_surface_test.go:269: reachable direct_read operation commands = 70, want 82
    command_surface_test.go:270: reachable direct_write operation commands = 69, want 78
--- FAIL: TestAIServicesOperationCommandsAreReachable (0.04s)
    command_surface_test.go:814: Preflight("ai-services scribe jobs list") = connector command "ai-services scribe jobs list" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services scribe jobs submit") = connector command "ai-services scribe jobs submit" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services scribe jobs get") = connector command "ai-services scribe jobs get" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services scribe jobs cancel") = connector command "ai-services scribe jobs cancel" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services scribe jobs files list") = connector command "ai-services scribe jobs files list" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services scribe jobs files get") = connector command "ai-services scribe jobs files get" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services scribe live") = connector command "ai-services scribe live" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services scribe transcribe") = connector command "ai-services scribe transcribe" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services summarizer jobs list") = connector command "ai-services summarizer jobs list" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services summarizer jobs submit") = connector command "ai-services summarizer jobs submit" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services summarizer jobs get") = connector command "ai-services summarizer jobs get" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services summarizer jobs cancel") = connector command "ai-services summarizer jobs cancel" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services summarizer jobs files list") = connector command "ai-services summarizer jobs files list" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services summarizer jobs files get") = connector command "ai-services summarizer jobs files get" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services summarizer summarize") = connector command "ai-services summarizer summarize" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services translator jobs list") = connector command "ai-services translator jobs list" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services translator jobs submit") = connector command "ai-services translator jobs submit" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services translator jobs get") = connector command "ai-services translator jobs get" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services translator jobs cancel") = connector command "ai-services translator jobs cancel" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services translator jobs files list") = connector command "ai-services translator jobs files list" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services translator jobs files get") = connector command "ai-services translator jobs files get" is blocked: unknown command, want declared executable AI Services action
    command_surface_test.go:814: Preflight("ai-services translator translate") = connector command "ai-services translator translate" is blocked: unknown command, want declared executable AI Services action
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom    0.919s
FAIL
```

The output contains only route names, counts, and declaration identifiers. No
credential, token-derived value, signed URL, storage credential, webhook secret, or transcript is
present.

## Planned foundation RED/GREEN

The Live Scribe source is an actual bidirectional WebSocket contract, not a REST JSON response.
The foundation must first prove that the runtime currently has no `websocket_session` operation
kind/CLI intent/preflight path. Green tests must establish all of the following with a loopback
server:

- only an exact declaration-owned GET endpoint can negotiate upgrade;
- an admitted `101`, `live-asr` subprotocol, auth, and session-update frame are required;
- frames are client-masked, type/size bounded, and cancellation-aware;
- a fixed PCM16 file is sent as bounded binary frames, not exposed as generic bytes or a caller
  supplied transport request;
- transcript event JSON is bounded and `json_redacted`, while errors never reveal auth material;
- no non-WebSocket response, unrecognized subprotocol, wrong accept hash, cross-origin redirect,
  malformed frame, oversized frame, or undeclared config field is accepted.

The foundation records its individual red/green commits here before AI Services declarations are
authored.

### Foundation RED — captured 2026-08-08

The closed-contract loader test was added and run before any production schema, runtime, CLI, or
generator implementation:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadAcceptsClosedWebSocketSessionContract$'
--- FAIL: TestBundleLoadAcceptsClosedWebSocketSessionContract (0.00s)
    websocket_session_test.go:50: Load closed WebSocket session operation: load bundle acme: operations.json: /operations/0/websocket: additional property not allowed
FAIL
FAIL    polymetrics.ai/internal/connectors/engine    0.744s
FAIL
```

This is an intentional missing-foundation failure. The test declaration contains a fixed relative
path, fixed subprotocol, finite input/output/frame caps, and a closed initial frame schema; it
contains no endpoint chosen by a caller and no secret value.

### Continuation provenance — 2026-08-09

Before resuming this red checkpoint, the official source was fetched again from
`https://developers.zoom.us/docs/api/ai-services.md` at `2026-08-09T20:57:31Z`.
The response was HTTP `200`, `87,750` bytes, with SHA-256
`154631ef97c292468c81a79dc50cd51ea142d18f1f9fab060622215ddf3ba367`, identical to the
pre-RED audit. The source/ledger set therefore remains 22 AI Services operations.

The reusable WebSocket runtime is now split to dedicated foundation issue #3963 and a stacked PR as
required by the connector-lane ownership contract. This is a delivery-boundary change only: no
production connector declaration has changed, and the recorded RED remains the applicable test
contract for the consumer slice.
