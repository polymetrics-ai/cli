# Verification Checklist: #490

Verdict: **IN PROGRESS — not yet verified**.

## Required gates

- [x] `scripts/gsd doctor`
- [x] `programming-loop` availability checked once; manual fallback recorded
- [x] exact local `pi-workflow-engine@0.12.0` version, tarball, and npm integrity recorded
- [x] bounded parallel read-only analysis plus typed synthesis completed
- [x] focused Pi compatibility and AgentSession terminal-result tests
- [x] focused adapter/runtime tests preserving cwd, capability, binding, cancellation, abort/join, terminal validation
- [x] complete sequential Shepherd suite before review: 1717 pass, 0 fail, 1 intentional skip
- [x] strict no-emit TypeScript against exact coherent Pi `0.80.10` declarations
- [x] exact Pi family verifier passes for coding-agent/core/ai/tui `0.80.10`
- [x] offline Shepherd RPC exposes `pm-shepherd`
- [x] offline workflow-engine RPC exposes workflow commands
- [x] offline co-load preserves both command surfaces without duplicate/route drift
- [ ] exactly one comprehensive xhigh workflow-engine/Codex exact-head review
- [ ] every actionable review finding disposition recorded
- [ ] affected tests after fixes, then final full Shepherd gate once
- [x] `git diff --check`
- [x] changed paths are a subset of PLAN.md declared write scope
- [x] no production import or deep import of `pi-workflow-engine`
- [x] all 17 production-matrix guarantees preserved by the complete pre-review suite
- [ ] branch pushed and PR targets `feat/471-pi-agent-session-shepherd`

## Excluded child-worktree gates

Per the #471 Shepherd boundary, do not run Go, connector, certification, runtime-service, or root `make verify` gates here. The parent orchestrator owns repository-wide verification on the exact integrated parent head.

## Pre-review evidence

- Focused GREEN: 170/170 across Pi compatibility, production AgentSession runtime, retained SDK canary runner, and tool-policy integration.
- Sequential Shepherd suite: 1717 pass, 0 fail, 1 intentional skip in 166.3 seconds.
- Typecheck: TypeScript 5.9.3 strict no-emit over production Shepherd sources against the installed coherent Pi 0.80.10 declaration tree.
- Runtime family: exact coding-agent/core/ai/tui 0.80.10 verified.
- Offline RPC: isolated Shepherd, isolated workflow-engine, and co-loaded Pi 0.80.10 command surfaces passed. Pi 0.80.10 does not expose the old `--list-extensions` flag; machine-readable RPC command enumeration supplied the equivalent load evidence.
- Source/scope: no `pi-workflow-engine` production import or deep import; local `.pi/.workflow-runs/**` and `.pi/npm/**` remain excluded.

## 17-row preservation statement

The complete pre-review suite leaves all 17 rows green: production lifecycle; dependency and collision scheduling; persistent workspace leases; stale-parent refresh; bounded durable retries and human waits; all 14 external-effect crash windows; exact human commands; abort/join; CAS/race fences; external-effect reconciliation; stable resource ownership; exact-head review and dispositions; fail-closed evidence; protected default branches and no main merge; hostile-input and secret bounds; command UX and shutdown; and internal-only read-only roles. The modified AgentSession boundary changes only compatibility and completion interpretation: prompt settlement plus a validated bound terminal handoff is authoritative, while raw events are inert bounded telemetry.
