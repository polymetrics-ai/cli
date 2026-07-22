# Cycle 12 Pi Lifecycle Map Trace

## Delegation result

- Agent: `cycle12-pi-lifecycle-map`
- Mode: read-only; no files or external state changed.
- Frozen start: `7882cd70c25971e889ec04f63b98c936d605003e`.
- Scope: exact installed Pi 0.80.6 lifecycle, settlement, whole-session offline seam, creation
  result, and diagnostic DTOs.

## Exact normal event order

No-tool:

```text
agent_start
turn_start
message_start(user)
message_end(user)
message_start(assistant partial)
message_update(assistant content event)*
message_end(final assistant)
turn_end(final assistant, [])
agent_end([user, final assistant], willRetry=false)
agent_settled
```

One-tool:

```text
agent_start
turn_start
message_start(user)
message_end(user)
message_start(intermediate assistant)
message_update(toolcall_start/delta*/end)
message_end(intermediate assistant, stopReason=toolUse)
tool_execution_start
tool_execution_update*
tool_execution_end
message_start(toolResult)
message_end(toolResult)
turn_end(intermediate assistant, [toolResult])
turn_start
message_start(final assistant)
message_update*
message_end(final assistant)
turn_end(final assistant, [])
agent_end([user, intermediate assistant, toolResult, final assistant], willRetry=false)
agent_settled
```

All-terminating tool results may omit the second turn. Tool-execution start is
`{type,toolCallId,toolName,args}`, update additionally carries `partialResult`, and end is
`{type,toolCallId,toolName,result,isError}`. Tool-result messages carry role, tool-call ID/name,
content, optional details, error flag, and timestamp.

Pi forwards inner text/thinking/tool-call start/delta/end as outer `message_update` events. Inner
`done` is not forwarded; it produces `message_end`. `agent_end` is attempt-terminal and may precede
retry/continuation; only the later `agent_settled` means the AgentSession run is quiescent.
`prompt()` returns only after `_emitAgentSettled()` has synchronously called public listeners and
resolved idle waiters. Public session listeners return `void` and their returned promises are not
awaited. Therefore Shepherd must keep its listener through settled, freeze there, and verify after
prompt returns.

## Whole-session no-network seam

Use the actual pinned `createAgentSession` with public in-memory `AuthStorage`, `ModelRegistry`,
`SettingsManager`, `SessionManager`, and a resource loader with extensions, skills, templates,
themes, and context files disabled. Register a unique programmatic provider through public
`ModelRegistry.registerProvider` with a custom API and inert/scripted `streamSimple`. Its
`createAssistantMessageEventStream` emits start/content/done and can script two responses for the
tool case. Preserve Shepherd's required provider/model route and use the real policy custom tool.

Pi's public `AgentSession.prompt()` always requires configured auth. A non-secret offline sentinel
marker in the in-memory provider config is required for preflight and is never transmitted because
the custom stream is inert. Dynamic API registration is process-global, so serialize this test and
unregister/dispose afterward. Return the complete actual factory result to Shepherd; do not graft
its extensions onto a fake session. Shepherd owns subscribe, prompt, cancellation/wait,
unsubscribe, and dispose.

## Exact creation and diagnostic shapes

The actual result has three own enumerable/writable/configurable data properties:
`session`, `extensionsResult`, and `modelFallbackMessage`; the fallback property exists even when
its value is `undefined`. `extensionsResult` has exact own data fields `extensions`, `errors`, and
required `runtime`. Runtime is compatibility evidence only and must never be read as authority.

Installed assistant diagnostics contain required string `type`, required finite numeric
`timestamp`, optional `error`, and optional `details`. The error projection has required string
`message` and optional string `name`, string `stack`, and string-or-number `code`. Optional own
`undefined` fields are legitimate and should be omitted consistently from the captured DTO;
required undefined and arbitrary fields reject. Codex transport fallback uses type
`provider_transport_failure` and details `{configuredTransport, fallbackTransport?, eventsEmitted,
phase, requestBytes}`; `fallbackTransport` is legitimately own-undefined when streaming already
started. A successful final assistant may retain this diagnostic.

## Source evidence

- `pi-agent-core/dist/agent-loop.js`: run loop at lines 42+, streaming at 177+, tool lifecycle at
  295+ and 449+.
- `pi-agent-core/dist/types.d.ts`: AgentEvent shapes at line 344+.
- `pi-coding-agent/dist/core/agent-session.js`: settled emission at 288+, prompt loop at 728+,
  prompt auth preflight at 827+.
- `pi-coding-agent/dist/core/sdk.js`: exact result construction at 268+.
- `pi-coding-agent/dist/core/model-registry.js`: public provider validation/stream registration at
  688+.
- `pi-ai/dist/utils/diagnostics.js` and `pi-ai/dist/api/openai-codex-responses.js`: diagnostic and
  Codex fallback DTO construction.

## Implementation consequence

The frozen runtime is assistant-only: it rejects the initial user events, treats the first
assistant end as the only terminal, rejects the tool-result/multiple-assistant transcript, requires
exactly one assistant in `agent_end.messages`, and rejects the required `agent_settled`. Cycle 12
must add per-role message/tool lifecycles, select the last assistant from the final non-retrying
`agent_end`, and complete/freeze exactly at `agent_settled`.

## Cycle 13 Read-Only Trace

- Agent: `/root/issue_475_agent_session_runtime/cycle13_seam_map`.
- Input: seven exact-head blockers from both complete Cycle 12 reports; frozen candidate
  `5dafc572`; same issue-owned source/tests and existing artifact paths only.
- Actions: inspected runtime normalization/SDK callbacks/terminal capture, policy public-array and
  classifier seams, prompt public-array seams, existing focused helpers, and installed Pi tool
  event/result fields. No command mutated files or external state.
- Key standards result: arbitrary hidden/symbol absence cannot be proven with bounded JavaScript
  reflection. Parent/root disposition selects canonical influence capture: bounded length/index
  descriptors only; all non-index peers untouched and discarded.
- Tool lifecycle result: assistant `{id,name,arguments}` must originate each authorized record;
  execution start/end, tool-result message, `turn_end.toolResults`, and subsequent turn/final
  handoff must close the same one-to-one identity.
- Ownership: advisory map only. The issue worker retains every artifact edit, behavior RED,
  production edit, gate, commit, and handoff.

## Cycle 14 Read-Only Trace

- Agent: `/root/issue_475_agent_session_runtime/cycle14_boundary_map`.
- Input: three unique correction families from both complete Cycle 13 reports; frozen candidate
  `67050a4a`; same issue-owned source/tests and existing artifact paths only.
- Lifecycle result: map every post-create getter/method callback through claim and subscription;
  acquire cleanup ownership before optional validation, capture a synchronous subscription's
  unsubscriber, then enforce closure/scope/capture-failure barriers before any later callback or
  prompt. The existing close race is specifically visible to the combined execution barrier, not
  the scope-only assertion.
- Capability result: the exact closed host domain is `host_inspect` and `host_verify`; workspace
  tools remain separate and #479 controller/Git/GitHub/review/integration ports remain outside the
  AgentSession surface. Unknown host strings require rejection, not semantic synonym matching.
- Redaction result: every production consumer reaches `redactSensitiveText`; replace fuzzy
  ancestor subsequences with a bounded canonical path grammar, exact secret schema/terminal rules,
  narrow public-metadata controls, and fail-closed unknown assignments.
- Ownership: advisory map only; no file or external state changed. The issue worker owns all plan,
  test, production, commit, and verification actions.
