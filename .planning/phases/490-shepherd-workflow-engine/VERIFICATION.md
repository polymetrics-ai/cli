# Verification Checklist: #490

Verdict: **IN PROGRESS — not yet verified**.

## Required gates

- [x] `scripts/gsd doctor`
- [x] `programming-loop` availability checked once; manual fallback recorded
- [x] exact local `pi-workflow-engine@0.12.0` version, tarball, and npm integrity recorded
- [x] bounded parallel read-only analysis plus typed synthesis completed
- [ ] focused Pi compatibility and AgentSession terminal-result tests
- [ ] focused adapter/runtime tests preserving cwd, capability, binding, cancellation, abort/join, terminal validation
- [ ] complete sequential Shepherd suite: `node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts`
- [ ] strict no-emit TypeScript against exact coherent Pi `0.80.10` declarations
- [ ] exact Pi family verifier passes for coding-agent/core/ai/tui `0.80.10`
- [ ] offline Shepherd RPC exposes `pm-shepherd`
- [ ] offline workflow-engine RPC exposes workflow commands
- [ ] offline co-load preserves both command surfaces without duplicate/route drift
- [ ] exactly one comprehensive xhigh workflow-engine/Codex exact-head review
- [ ] every actionable review finding disposition recorded
- [ ] affected tests after fixes, then final full Shepherd gate once
- [ ] `git diff --check`
- [ ] changed paths are a subset of PLAN.md declared write scope
- [ ] no production import or deep import of `pi-workflow-engine`
- [ ] all 17 production-matrix guarantees preserved
- [ ] branch pushed and PR targets `feat/471-pi-agent-session-shepherd`

## Excluded child-worktree gates

Per the #471 Shepherd boundary, do not run Go, connector, certification, runtime-service, or root `make verify` gates here. The parent orchestrator owns repository-wide verification on the exact integrated parent head.

## 17-row preservation statement

Pending final evidence. The implementation must leave Shepherd authoritative for lifecycle, scheduling/collisions, persistent workspace leases, stale-parent refresh, bounded durable retries/human waits, all 14 external-effect crash windows, exact human commands, abort/join, race fences, external-effect reconciliation, stable ownership, exact-head review/dispositions, fail-closed evidence, protected default branches/no main merge, hostile input bounds, command UX/shutdown, and internal-only read-only roles.
