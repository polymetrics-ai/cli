# Verification Checklist: #490

Verdict: **BOUNDED MERGE-READINESS PASS IN PROGRESS — PR #491 is published and green; exact final-head review remains**.

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
- [x] branch pushed and PR #491 targets `feat/471-pi-agent-session-shepherd`
- [x] PR #491 is open, non-draft, mergeable/CLEAN, and references #490 and #471
- [x] workflow-engine boundary distinguishes Shepherd production from explicitly invoked developer tooling
- [x] bounded two-lane Codex 5.6-sol xhigh review covers the complete current base...HEAD diff
- [x] two accepted P1 terminal-snapshot findings are corrected with focused RED/GREEN and conditional local gates
- [ ] dispositions are verified on the exact final pushed SHA without another broad review
- [ ] final PR checks are green after the bounded artifact/correction commit

## Excluded child-worktree gates

Per the #471 Shepherd boundary, do not run Go, connector, certification, runtime-service, or root `make verify` gates here. The parent orchestrator owns repository-wide verification on the exact integrated parent head.

## PR and pre-review evidence

- GitHub verification at the start of this bounded pass: PR #491 is open and non-draft from `refactor/490-shepherd-workflow-engine` to `feat/471-pi-agent-session-shepherd`; local, remote, and GitHub head equal `dceef7ffc4193b9251e2b2159abd03c876174b39`; GitHub reports `MERGEABLE`/`CLEAN`; all reported checks pass.
- PR body verification: contains `Refs #490` and `Refs #471`, exact package/test/review evidence, and the 17-row preservation table.
- Review-route override: the user selected independent `openai-codex/gpt-5.6-sol` at xhigh instead of Claude or Copilot. This evidence is not human approval and does not authorize integration or merge to `main`.
- Focused GREEN: 170/170 across Pi compatibility, production AgentSession runtime, retained SDK canary runner, and tool-policy integration.
- Sequential Shepherd suite: 1717 pass, 0 fail, 1 intentional skip in 166.3 seconds.
- Typecheck: TypeScript 5.9.3 strict no-emit over production Shepherd sources against the installed coherent Pi 0.80.10 declaration tree.
- Runtime family: exact coding-agent/core/ai/tui 0.80.10 verified.
- Offline RPC: isolated Shepherd, isolated workflow-engine, and co-loaded Pi 0.80.10 command surfaces passed. Pi 0.80.10 does not expose the old `--list-extensions` flag; machine-readable RPC command enumeration supplied the equivalent load evidence.
- Source/scope: no `pi-workflow-engine` production import or deep import; local `.pi/.workflow-runs/**` and `.pi/npm/**` remain excluded.

## Bounded merge-readiness review

Workflow run `29165c4c-c62c-405d-a6b0-78010dcc5a82` reviewed exact `daaa22637e54de87ce5e0c0f5876a19ddf7fb274...7c5f927f0d970f1b38536f94b61e857bf7e51a68` using two parallel read-only `openai-codex/gpt-5.6-sol` xhigh lanes and one xhigh synthesis. The host-captured material digest was `297306d7f3f0afe9da886de3dfdb62ad9d3b43e619df24df39f2675b74fff38b`; agents had `tools: []`, no skills, and no GitHub mutation authority. The CI/delivery lane was clean. The runtime/security lane found two P1 blockers:

| Finding | Disposition |
|---|---|
| `P1-terminal-status-erased`: text-only completion accepted error/aborted/length/toolUse final messages | **Accepted with modification.** Preserve event-agnostic completion, but require the final public Pi session-message snapshot to have `stopReason: "stop"` before parsing its typed handoff. |
| `P1-terminal-route-unbound`: configured session route substituted for actual producing route | **Accepted with modification.** Validate provider/model on the final public Pi session-message snapshot and reconcile its text with `getLastAssistantText()`. |

The one allowed correction pass is complete. The selected RED command produced the six intended status/route failures with the unchanged `stop` control passing. GREEN passed 9/9 selected tests and 163/163 complete affected runtime tests. Production now reads the final assistant message from Pi's documented public `session.messages` state after `prompt()` settles, requires `stopReason: "stop"` and exact producing provider/model, reconciles its text with `getLastAssistantText()`, and only then validates the typed handoff. Unknown/reordered events remain inert bounded telemetry.

Post-correction gates passed: strict TypeScript 5.9.3 source typecheck against Pi 0.80.10; exact Pi-family verifier; workflow-engine provenance verifier; offline isolated/co-loaded RPC verifier; `git diff --check`; and exactly one final sequential complete Shepherd suite with 1718 pass, 0 fail, and 1 intentional skip (1719 total) in 159.6 seconds. Independent Codex review is not human approval and authorizes no merge.

## Prior exact-head review and dispositions

The one substantive comprehensive review used `openai-codex/gpt-5.6-sol` at xhigh against exact parent `daaa22637e54de87ce5e0c0f5876a19ddf7fb274` and exact child `23345f26cfa78d7648c31fd11cbe595ecf1b9d1d`: workflow run `8e04bba9-012b-422f-89cb-4d845f340e3e`. Run `abaefcd2-90cf-4569-9ac6-b7e58d0428e9` was a failed preflight only: a mistyped nonexistent child SHA caused the agent to stop before inspecting code or assessing 16/17 rows. It is not a second substantive review. Per the bounded single-review policy, fixes were dispositioned with affected tests and the final complete gate rather than another model review.

| Finding | Disposition |
|---|---|
| Pi compatibility was asserted only at first AgentSession dispatch | **Fixed.** `index.ts` now asserts the bounded policy synchronously before command registration, controller construction, filesystem state, Git inspection, or GitHub activity. |
| `.pi/.workflow-runs` dirtied canonical target evidence | **Fixed.** `.gitignore` now classifies the engine's bounded run records as local non-authoritative Pi state; `git check-ignore` verifies the rule. |
| Package provenance and isolated/co-load checks were absent from CI | **Fixed.** CI installs exact `0.12.0` without peer drift, verifies exact settings/version/tarball/integrity/public entry, and exercises isolated Shepherd, isolated workflow-engine, and co-loaded offline Pi RPC surfaces. |
| Built-in advisory workflows expose generic shell and ambient authority | **Boundary confirmed and wording corrected; no Shepherd production defect was found.** Production Shepherd imports only public Pi APIs, registers only `pm-shepherd`, and has no route to workflow-engine commands or built-ins; isolated/co-loaded RPC tests keep the command surfaces distinct. A developer may explicitly invoke package built-ins with the normal authority of that Pi coding-agent session, but that execution is outside Shepherd production and never Shepherd evidence or durable authority. The bounded review path receives host-captured exact material with `tools: []` and no GitHub mutation authority. |
| Telemetry did not count every callback and encoded complete oversized tool names first | **Fixed.** Both collectors charge every callback before filtering, saturate non-authoritatively, pre-bound string length before UTF-8 accounting, and never let telemetry create or revoke success. The focused saturation test proves later callbacks are not inspected. |

Affected post-review gates passed: 207/207 focused tests, strict Pi 0.80.10 typecheck, exact Pi family verifier, workflow-engine provenance verifier, and offline isolated/co-load RPC. The one final complete Shepherd gate passed 1717 tests, failed 0, skipped 1 intentional test, in 149.9 seconds.

## 17-row preservation statement

Both complete sequential suites leave all 17 rows green: production lifecycle; dependency and collision scheduling; persistent workspace leases; stale-parent refresh; bounded durable retries and human waits; all 14 external-effect crash windows; exact human commands; abort/join; CAS/race fences; external-effect reconciliation; stable resource ownership; exact-head review and dispositions; fail-closed evidence; protected default branches and no main merge; hostile-input and secret bounds; command UX and shutdown; and internal-only read-only roles. The modified AgentSession boundary changes only compatibility and completion interpretation: prompt settlement plus a validated bound terminal handoff is authoritative, while raw events are inert bounded telemetry.
