# Cycle 11 Pi 0.80.6 Contract Map — Read-Only Trace

## Scope and actions

The explorer inspected the explicit installation at
`/Users/karthiksivadas/.nvm/versions/node/v24.13.1/lib/node_modules/@earendil-works/pi-coding-agent`
and the issue-owned runtime/test seam. It made no edit, commit, ref, network, credential, model, or
verification-gate action.

## Binding findings

- Public Pi 0.80.6 `createAgentSession` returns own enumerable data fields `session`,
  `extensionsResult`, and `modelFallbackMessage`. Its `extensionsResult` is the exact loader object
  with own data fields `extensions`, `errors`, and `runtime`; the current two-field fake/gate is not
  compatible with the installed factory.
- The returned extension runtime is already bound to session actions. Shepherd may require and
  descriptor-check the field for compatibility, but must never retain, enumerate deeply, invoke,
  or derive authority from its value.
- A verified no-model route uses `AuthStorage.inMemory`, `ModelRegistry.inMemory`,
  `SettingsManager.inMemory`, `SessionManager.inMemory`, and `DefaultResourceLoader` with every
  resource family disabled; call `reload`, call public `createAgentSession`, never call `prompt`,
  inspect the result, and dispose. An inert required-route model descriptor exercises Shepherd's
  routing shape without auth, network, credentials, or a model call.
- Outer `message_start` carries inner `start`; nine content start/delta/end variants become
  `message_update`; inner done/error become `message_end`. Stateful capture must begin at
  `message_start`, compare outer message with inner partial on each update, and compare the final
  assistant `message_end` with the assistant selected from `agent_end`.
- Complete assistant evidence includes required `api` and `usage`, response model/id, diagnostics,
  error message, all usage/cost fields, text and thinking bodies/signatures/redaction, and tool-call
  ID/name/arguments/thought signature. Tool-call deltas may temporarily carry the runtime-only
  `partialJson`; it disappears at end and must be bounded as actual state growth.
- Pi may emit `message_end` for non-assistant user/tool-result messages and `agent_end.messages` may
  include multiple message roles. The terminal state machine must select exactly one final
  assistant pair without overwriting it with non-assistant evidence.

## Primary installed sources

- package public export and factory: `dist/index.js`, `dist/core/sdk.{d.ts,js}`
- split services: `dist/core/agent-session-services.js`
- extension result/runtime: `dist/core/extensions/types.d.ts`, `dist/core/agent-session.js`,
  `dist/core/extensions/runner.js`
- outer events: `dist/core/agent-session.d.ts` and nested `pi-agent-core/dist/{types.d.ts,agent-loop.js}`
- assistant/event DTOs: nested `pi-ai/dist/types.d.ts` and `pi-ai/dist/utils/diagnostics.d.ts`
- OpenAI-Codex tool partials: nested `pi-ai/dist/api/openai-responses-shared.js`

## Handoff

RED should use the actual factory path without prompting, update only test fixtures after RED to the
real three-field result and mandatory assistant fields, and cover runtime-only tool `partialJson` as
a bounded transient projection. Production remains frozen until the comprehensive test-only RED is
committed.
