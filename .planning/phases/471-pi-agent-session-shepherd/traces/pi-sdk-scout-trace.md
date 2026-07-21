# Pi SDK Scout Trace

## Prompt reference

Read-only inspection of installed Pi 0.80.6 public APIs for embedded AgentSession construction,
resource isolation, model/tool verification, cancellation, cleanup, and fake-model testing.

## Files inspected

- installed package root exports and TypeScript declarations for Pi 0.80.6;
- Pi SDK and extension documentation;
- official full-control SDK and subagent examples.

## Actions and commands

Read-only source/type inspection plus an isolated in-memory faux-provider smoke. No repository,
GitHub, auth, or provider state changed and no real model/network call was made.

## Findings

- Use package-root exports only: `createAgentSession`, `SessionManager`, `SettingsManager`,
  `DefaultResourceLoader`, `ModelRegistry`, `VERSION`, and public extension/session types.
- Fail closed unless `VERSION` is exactly `0.80.6` and required public methods exist.
- Supply `noExtensions`, `noSkills`, `noPromptTemplates`, `noThemes`, and `noContextFiles`; call
  `resourceLoader.reload()` before session construction.
- Reuse the current locked model registry without extracting auth values. Verify exact
  `openai-codex/gpt-5.6-sol`, thinking level, active tool set, zero nested extensions, and an
  in-memory session with no session file.
- `prompt()` has no AbortSignal. Timeout/stop calls `abort()`, then waits idle, unsubscribes, and
  disposes in an idempotent `finally` path.
- Fake session factories cover most tests without model calls. The faux provider can cover a full
  in-memory prompt loop serially with test-only auth and no network.

## Handoff

Keep Pi SDK imports in `sdk-runner.ts`; inject the factory and managers so the domain/controller
tests remain dependency-free and deterministic.

## Unresolved risks

All embedded children share one process, event loop, heap, environment, and crash domain. Abort is
cooperative. These are inherent product limits and must remain visible in help/status/docs.
