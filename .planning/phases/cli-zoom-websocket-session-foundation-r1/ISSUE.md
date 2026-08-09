# Issue #3963 — fixed WebSocket session operation foundation

## Objective

Add the closed, declaration-owned WebSocket-session operation foundation required by Zoom AI
Services Live Scribe, without adding a generic WebSocket, HTTP, shell, SQL, or raw-frame escape
hatch. This is shared engine/command-runner work and is split from connector #3935 under the
connector-lane ownership contract.

## Parent and consumer

- Parent delivery: #3915 (Zoom documented-operation parity).
- Blocking consumer: #3935 (Zoom AI Services, 22 operations).
- Parent branch / stacked-PR base: `fm/cli-zoom-full-parity-r1`.

## Scope

- Extend the engine operation schema, bundle loader, command preflight/run path, and generated
  command metadata support for one fixed `websocket_session` declaration.
- Use only the Go standard library. Do not add a dependency.
- Make every transport-sensitive value declaration-owned: connector-relative endpoint, GET upgrade,
  subprotocol, initial JSON frame schema, client binary-file format, input/output/frame limits,
  and result redaction policy.
- Prove the foundation through loopback tests: handshake status and accept hash, required
  subprotocol, declared auth, fixed frame ordering, per-frame and aggregate bounds, cancellation,
  redaction, and fail-closed rejection of redirects/malformed frames/undeclared controls.
- Do not declare or expose a Zoom command in this foundation PR. The consumer PR owns Zoom's
  `operations.json`, `cli_surface.json`, `api_surface.json`, fixture lifecycle, docs, and binary
  command reachability.

## Non-goals and safety

- No caller-selected URL, origin, headers, subprotocol, initial payload, frame type, or arbitrary
  bytes.
- No credential creation, live credentialed calls, secrets, tokens, transcripts, or provider-host
  recordings in tests, logs, fixtures, or planning records.
- No reverse-ETL execution and no generic write tool.

## Existing RED evidence

The consumer committed the failing loader contract before any production implementation:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestBundleLoadAcceptsClosedWebSocketSessionContract$'
--- FAIL: TestBundleLoadAcceptsClosedWebSocketSessionContract (0.00s)
    websocket_session_test.go:50: Load closed WebSocket session operation: load bundle acme: operations.json: /operations/0/websocket: additional property not allowed
FAIL
```

The test has a fixed relative path, fixed `live-asr` subprotocol, finite bounds, and a closed
initial-frame schema. It contains no secret values or caller transport controls.

## TDD and verification

1. Preserve the RED test above as the first foundation checkpoint.
2. Add focused engine/commandrunner tests before each production behavior.
3. Run targeted Go tests, `gofmt`, `go vet`, the relevant connectorgen schema/validation checks,
   `go build ./cmd/pm`, and a loopback-only binary/preflight check.
4. Record GSD `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` →
   `code-review` evidence and required Go skills in a dedicated `.planning/phases/` directory.
5. Open a stacked PR to `fm/cli-zoom-full-parity-r1`; it must cite this issue, #3935, and #3915,
   and state that it unblocks fixed WebSocket connector operations without making a generic
   transport surface available.

## Sources

- Zoom AI Services Live Scribe artifact: https://developers.zoom.us/docs/api/ai-services.md
  (re-fetched `2026-08-09T20:57:31Z`, HTTP 200, 87,750 bytes, SHA-256
  `154631ef97c292468c81a79dc50cd51ea142d18f1f9fab060622215ddf3ba367`).
- `.planning/phases/cli-zoom-parity-ai-services-r1/{PLAN,TDD-LEDGER,RUN-STATE}.md`
- `AGENTS.md` and `.agents/agentic-delivery/contracts/issue-agent-contract.md`
