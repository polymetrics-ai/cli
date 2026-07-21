# TDD Ledger — Issue #475

## Policy

- Mode: `manual_gsd_fallback` because `scripts/gsd prompt programming-loop ...` is absent from the
  healthy 69-command adapter registry.
- Production code is blocked until the RED checkpoint below is captured.
- Tests use a fake injected Pi SDK/session; they must not require model auth, network, secrets, a Pi
  subprocess, tmux, Git/GitHub mutation, or a real workspace.

## Cycle 1 — AgentSession Authority And Lifecycle

### RED

- Status: captured
- Test files: `agent-session-runtime.test.ts`, `tool-policy.test.ts`, optional
  `role-prompts.test.ts`
- Expected coverage: exact route, bounded context, least authority, recursion prevention,
  cancellation/deadline/close/shutdown races, join once, quarantine, and bound/redacted handoff.
- Command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

- Observed failure: exit 1, 0 passed / 2 file-level failures. Node reported
  `ERR_MODULE_NOT_FOUND` for `agent-session-runtime.ts` and `tool-policy.ts`. This is the expected
  pre-production RED state: the fake-SDK authority/lifecycle test contracts exist and the owned
  production adapters do not.

### GREEN

- Status: captured.
- Minimal implementation: exact role router and trusted prompt envelope; opaque workspace and
  typed-capability tool policy; injected Pi 0.80.6 `createAgentSession` lifecycle owner; strict
  bounded/redacted handoff parser.
- Command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

- Observed pass: exit 0, 19 passed / 0 failed. Coverage includes 5.5/fallback/tool/terminal drift,
  injection and read-only authority, path and secret boundaries, abort/timeout/deadline/close/
  parent-shutdown races, late creation, join-once, cleanup quarantine, schema/binding/bounds, and
  mutator concurrency.

### REFACTOR

- Status: blocked on GREEN.
- Refactor notes and broader gates: pending.

## Gate History

| Checkpoint | Command | Result | Evidence |
|---|---|---|---|
| Adapter | `scripts/gsd doctor` | pass | Pi adapter and registry healthy |
| GSD command | `scripts/gsd prompt programming-loop init --phase 475-shepherd-agent-session-runtime --dry-run` | unavailable | `unknown GSD command: programming-loop`; manual fallback activated |
| RED | `node --test .pi/extensions/shepherd/agent-session-runtime.test.ts .pi/extensions/shepherd/tool-policy.test.ts` | expected fail | 0 passed; missing owned production modules |
| GREEN | `node --test .pi/extensions/shepherd/agent-session-runtime.test.ts .pi/extensions/shepherd/tool-policy.test.ts` | pass | 19 passed; 0 failed |
