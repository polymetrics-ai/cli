# TDD Ledger — Issue #475

## Policy

- Mode: `manual_gsd_fallback` because `scripts/gsd prompt programming-loop ...` is absent from the
  healthy 69-command adapter registry.
- Production code is blocked until the RED checkpoint below is captured.
- Tests use a fake injected Pi SDK/session; they must not require model auth, network, secrets, a Pi
  subprocess, tmux, Git/GitHub mutation, or a real workspace.

## Cycle 1 — AgentSession Authority And Lifecycle

### RED

- Status: planned
- Test files: `agent-session-runtime.test.ts`, `tool-policy.test.ts`, optional
  `role-prompts.test.ts`
- Expected coverage: exact route, bounded context, least authority, recursion prevention,
  cancellation/deadline/close/shutdown races, join once, quarantine, and bound/redacted handoff.
- Command and observed failure: pending.

### GREEN

- Status: blocked on RED evidence.
- Minimal implementation: pending.
- Command and observed pass: pending.

### REFACTOR

- Status: blocked on GREEN.
- Refactor notes and broader gates: pending.

## Gate History

| Checkpoint | Command | Result | Evidence |
|---|---|---|---|
| Adapter | `scripts/gsd doctor` | pass | Pi adapter and registry healthy |
| GSD command | `scripts/gsd prompt programming-loop init --phase 475-shepherd-agent-session-runtime --dry-run` | unavailable | `unknown GSD command: programming-loop`; manual fallback activated |
