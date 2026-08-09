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

## Consumer resume confirmation — 2026-08-10

The consumer branch starts from the verified #3963 foundation without adding Zoom declarations to
that shared-runtime PR. The formerly missing WebSocket schema property is now accepted by the closed
operation loader; the remaining consumer RED is the 22 exact AI Services command paths:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadAcceptsClosedWebSocketSessionContract$'
ok  polymetrics.ai/internal/connectors/engine  0.728s

$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^TestAIServicesOperationCommandsAreReachable$'
--- FAIL: TestAIServicesOperationCommandsAreReachable (0.06s)
    command_surface_test.go:814: Preflight("ai-services scribe jobs list") = connector command "ai-services scribe jobs list" is blocked: unknown command, want declared executable AI Services action
    ... all 22 documented AI Services command paths remain unknown ...
FAIL
FAIL  polymetrics.ai/internal/connectors/defs/zoom
```

The current artifact was also re-fetched immediately before this continuation: HTTP `200`, `87,750`
bytes, SHA-256 `154631ef97c292468c81a79dc50cd51ea142d18f1f9fab060622215ddf3ba367`, retrieved
2026-08-09T23:06:39Z / 2026-08-10 IST. No source or ledger delta exists.

## WebSocket reconciliation bootstrap RED — captured 2026-08-10

After the closed schema property was accepted, the first AI Services reconciliation exposed a
separate foundation defect: preflight required the generated `covered_by.websocket_session` marker
before `surface-reconcile` could derive it. The regression test was added and committed as
`2b79911aa` before the engine change:

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestRunSurfaceReconcileCoversWebSocketSessionWithRuntimePreflight$'
--- FAIL: TestRunSurfaceReconcileCoversWebSocketSessionWithRuntimePreflight (0.87s)
    surfacereconcile_test.go:85: stats = {Scanned:1 Covered:0 Blocked:1 Unchanged:0 Refused:0}, want one runtime-covered websocket session
FAIL
FAIL    polymetrics.ai/cmd/connectorgen    1.688s
FAIL
```

No credential, request payload, provider token, signed URL, or transcript appears in this failure.

## GREEN — captured 2026-08-10

The separate engine commit `29e4e64c1` permits only an exact source-ledger
`operation.model=websocket_session` row as the bootstrap state. It does not weaken the fixed
endpoint/subprotocol/schema checks; reconciliation immediately replaces the row with generated
coverage. The lint-only loopback test cleanup is `8518509c3`.

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestRunSurfaceReconcileCoversWebSocketSessionWithRuntimePreflight$'
ok      polymetrics.ai/cmd/connectorgen    0.851s

$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadAcceptsClosedWebSocketSessionContract$'
ok      polymetrics.ai/internal/connectors/engine    0.715s

$ go run ./cmd/connectorgen surface-reconcile internal/connectors/defs/zoom --notes-contains provider_module=ai-services --json
{
  "total": {"scanned": 1, "covered": 1, "blocked": 0, "unchanged": 0, "refused": 0}
}
```

## AI Services GREEN — captured 2026-08-10

The 22 command declarations, imported parameter metadata, generated surface metadata, and exactly
22 reconciled provider rows now pass the real command runner and the category's loopback tests.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestAIServicesOperationCommandsAreReachable|TestProviderInventoryLedgerIsComplete|TestCoveredStreamsHaveReachableCommands)$'
ok      polymetrics.ai/internal/connectors/defs/zoom    1.057s

$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^TestAIServices(DirectRead|DirectWrite)CommandsExecuteWithFixtures$'
ok      polymetrics.ai/internal/connectors/defs/zoom    3.404s

$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom
ok      polymetrics.ai/internal/connectors/defs/zoom    22.824s

$ go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json
{"findings":null,"warnings":null,"connectors_checked":1}

$ go run ./cmd/connectorgen surface-sync internal/connectors/defs --check
connectorgen surface-sync: 552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

The loopback read test covers the twelve GETs and asserts that only imported `state`, `jobId`, and
`fileId` inputs reach the fixed request. It refuses raw `page_size`, `next_page_token`, `page`,
`per_page`, and `limit`. The write test covers six JSON submissions/synchronous actions and all
three `DELETE` cancellations through plan → preview → single-use approval → execute; each delete
requires destructive confirmation and returns HTTP 204 with no invented body. All test values are
synthetic, and response token fields are asserted redacted.
