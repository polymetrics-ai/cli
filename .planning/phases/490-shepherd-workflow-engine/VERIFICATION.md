# Verification Checklist: #490

Verdict: **FUNCTIONAL PASS — local child-worktree gates complete; PR publication pending**.

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
- [x] exactly one substantive comprehensive xhigh workflow-engine/Codex exact-head review
- [x] every actionable review finding disposition recorded
- [x] affected tests after fixes, then final full Shepherd gate once
- [x] `git diff --check`
- [x] changed paths are a subset of PLAN.md declared write scope
- [x] no production import or deep import of `pi-workflow-engine`
- [x] all 17 production-matrix guarantees preserved by both complete suites
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

## Exact-head review and dispositions

The one substantive comprehensive review used `openai-codex/gpt-5.6-sol` at xhigh against exact parent `daaa22637e54de87ce5e0c0f5876a19ddf7fb274` and exact child `23345f26cfa78d7648c31fd11cbe595ecf1b9d1d`: workflow run `8e04bba9-012b-422f-89cb-4d845f340e3e`. Run `abaefcd2-90cf-4569-9ac6-b7e58d0428e9` was a failed preflight only: a mistyped nonexistent child SHA caused the agent to stop before inspecting code or assessing 16/17 rows. It is not a second substantive review. Per the bounded single-review policy, fixes were dispositioned with affected tests and the final complete gate rather than another model review.

| Finding | Disposition |
|---|---|
| Pi compatibility was asserted only at first AgentSession dispatch | **Fixed.** `index.ts` now asserts the bounded policy synchronously before command registration, controller construction, filesystem state, Git inspection, or GitHub activity. |
| `.pi/.workflow-runs` dirtied canonical target evidence | **Fixed.** `.gitignore` now classifies the engine's bounded run records as local non-authoritative Pi state; `git check-ignore` verifies the rule. |
| Package provenance and isolated/co-load checks were absent from CI | **Fixed.** CI installs exact `0.12.0` without peer drift, verifies exact settings/version/tarball/integrity/public entry, and exercises isolated Shepherd, isolated workflow-engine, and co-loaded offline Pi RPC surfaces. |
| Built-in advisory workflows expose generic shell and ambient authority | **Fixed for the adopted path.** Built-ins are explicitly outside this partial adoption. Approved inline analysis/review receives host-captured bounded exact-head material with `tools: []`, no tool hints, no skills, no GitHub credentials, and host exact-head revalidation. Production imports and authority remain absent. |
| Telemetry did not count every callback and encoded complete oversized tool names first | **Fixed.** Both collectors charge every callback before filtering, saturate non-authoritatively, pre-bound string length before UTF-8 accounting, and never let telemetry create or revoke success. The focused saturation test proves later callbacks are not inspected. |

Affected post-review gates passed: 207/207 focused tests, strict Pi 0.80.10 typecheck, exact Pi family verifier, workflow-engine provenance verifier, and offline isolated/co-load RPC. The one final complete Shepherd gate passed 1717 tests, failed 0, skipped 1 intentional test, in 149.9 seconds.

## 17-row preservation statement

Both complete sequential suites leave all 17 rows green: production lifecycle; dependency and collision scheduling; persistent workspace leases; stale-parent refresh; bounded durable retries and human waits; all 14 external-effect crash windows; exact human commands; abort/join; CAS/race fences; external-effect reconciliation; stable resource ownership; exact-head review and dispositions; fail-closed evidence; protected default branches and no main merge; hostile-input and secret bounds; command UX and shutdown; and internal-only read-only roles. The modified AgentSession boundary changes only compatibility and completion interpretation: prompt settlement plus a validated bound terminal handoff is authoritative, while raw events are inert bounded telemetry.
