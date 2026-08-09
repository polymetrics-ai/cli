# Plan — closed WebSocket session operation foundation, R1

## Delivery record

- Foundation issue: [#3963](https://github.com/polymetrics-ai/cli/issues/3963).
- Parent delivery: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); blocking consumer:
  [#3935](https://github.com/polymetrics-ai/cli/issues/3935), Zoom AI Services Live Scribe.
- Branch: `feat/3963-zoom-websocket-session-foundation`, stacked on
  `fm/cli-zoom-full-parity-r1`; the foundation PR targets that parent branch.
- Ownership: shared engine/schema/command-runner work only. This branch must not author Zoom's
  `operations.json`, `cli_surface.json`, `api_surface.json`, generated Zoom docs, or a user-facing
  Zoom command. Those remain #3935's consumer work.

## GSD and skills

- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, and `golang-documentation`.
- Adapter health and all five command sources were resolved before this branch. Generated prompts
  for `discuss-phase --auto`, `plan-phase --tdd --skip-research`, `execute-phase --interactive`,
  `verify-work`, and `code-review` ran on 2026-08-09.
- Inline/manual GSD fallback is required: this issue phase is not registered in the official phase
  runtime and the canonical parent contract forbids role spawning. This plan, the TDD ledger, run
  state, verification checklist, and final review record preserve the same discussion, plan, red,
  green, verification, and review evidence.

## Inline discuss-phase decisions

1. The capability is one closed `websocket_session` operation kind, not a generic WebSocket client.
   Its fixed endpoint, HTTP GET upgrade, subprotocol, initial-frame schema, binary input contract,
   output policy, and all byte bounds are connector declaration data.
2. The operation may never accept a caller-selected URL, origin, header, subprotocol, arbitrary
   initial frame, arbitrary raw frames, redirect target, or unbounded payload.
3. The implementation uses only the Go standard library. No dependency, credential, live provider
   call, or reverse-ETL execution is authorized.
4. The session runner propagates the caller context, checks cancellation between every bounded
   read/write, and owns/cleans up its network connection synchronously—no detached goroutine or
   background session is permitted.
5. Initial adoption must satisfy Zoom Live Scribe's fixed `live-asr` subprotocol and PCM16 input
   needs, but the foundation itself exposes no Zoom command. Its generic tests use only a loopback
   server and synthetic bytes.

## Source and predecessor RED

- Consumer artifact: `https://developers.zoom.us/docs/api/ai-services.md`, fetched
  `2026-08-09T20:57:31Z`, HTTP 200, 87,750 bytes, SHA-256
  `154631ef97c292468c81a79dc50cd51ea142d18f1f9fab060622215ddf3ba367`.
- The test-first predecessor is commit `ae43c153c`:
  `TestBundleLoadAcceptsClosedWebSocketSessionContract` fails because `websocket` is an additional
  forbidden property. This foundation re-runs that test before production edits and records the
  result verbatim in `TDD-LEDGER.md`.

## TDD slices

1. **Schema and loader GREEN.** Extend the closed `operations.json` meta-schema and typed bundle
   model; validate exactly one `websocket` execution block, its fixed GET relative path,
   non-empty subprotocol, positive finite input/output/frame bounds, and a closed compiled
   `session_update_schema`. Add negative tests for every open/invalid transport shape.
2. **Transport GREEN.** Add a standard-library client handshake and masked frame codec behind an
   operation-specific executor. Test acceptance hash, HTTP 101, required subprotocol, frame
   limits, redaction, close/control frames, malformed frames, redirects, and context cancellation
   against loopback only.
3. **Command boundary GREEN.** Add a dedicated command/preflight route that accepts only an
   implemented operation declaration and one matching GET surface endpoint. Reuse typed flag/body
   mappings where appropriate, but reject arbitrary transport controls and preserve the normal
   connector auth/rate-limit boundary. Add runtime/help generator support only if a declared
   operation needs it; no standalone generic command is added.
4. **Consumer handoff.** Verify the foundation's tests and built binary preflight behavior, record
   the exact files/commit that #3935 must consume, push the green foundation slice, and update the
   parent and consumer issues.

## Verification plan

- Focused engine and commandrunner tests, including the inherited loader RED/green contract.
- `gofmt` only on changed Go files; `go vet` for changed packages; targeted `go test -timeout 20m`
  for engine, commandrunner, defs/zoom, and CLI routing where affected.
- `go build ./cmd/pm`, `go run ./cmd/connectorgen validate`, and
  `go run ./cmd/connectorgen surface-sync --check` after a fixture consumer exists.
- Run every evidence command without credentials; no test or output stores tokens, transcript
  content, provider recordings, signed URLs, or raw authorization headers.

## Canonical references

- `AGENTS.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.planning/phases/cli-zoom-parity-ai-services-r1/{PLAN,TDD-LEDGER,RUN-STATE}.md`
