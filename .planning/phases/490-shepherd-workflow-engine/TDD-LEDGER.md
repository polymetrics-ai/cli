# TDD Ledger: #490

## Workflow state

- Manual-GSD fallback active: adapter health passed; `programming-loop` is absent from the 69-command registry and failed once as `unknown GSD command`.
- Required execution decision, planning cycle: `read_only_spawned` — workflow run `43dc4d81-66bf-4642-bed2-207a00d5fec0` used four read-only analysis lanes and one typed synthesis lane. The issue worker retains the only mutating path in this dedicated worktree.
- Required skill used: `gsd-core`; required Pi/runtime and issue-first references loaded.

## Bounded merge-readiness continuation

- GSD path: `scripts/gsd doctor` passed; `scripts/gsd prompt code-review 490` generated the repo-local `/gsd-code-review 490` contract.
- Starting Git/GitHub evidence: clean local worktree; local, remote branch, and PR head all equal `dceef7ffc4193b9251e2b2159abd03c876174b39`; PR #491 is open, non-draft, correctly based, mergeable/CLEAN, and all current checks pass.
- Review route override: independent Codex `openai-codex/gpt-5.6-sol`/`xhigh`, explicitly selected instead of Claude or Copilot. It is not human approval.
- This round starts documentation-only. No new production behavior is planned, so no new RED is required for the known artifact and authority-wording corrections. Existing exact-code-head evidence is retained.
- If either read-only review lane identifies an actionable production/CI blocker, add the smallest focused failing regression before the one allowed correction pass, then run the full conditional gate set from PLAN.md. Otherwise run only `git diff --check` and delivery/scope checks.
- Parent orchestration decision for preflight: `local_critical_path` — existing child PR #491 needs only bounded artifact correction and exact-diff review; no mutating implementation worker is ready or required. The required review lanes will be recorded as `read_only_spawned` after dispatch.

## Baseline and provenance

- Issue: #490, parent #471, parent PR #472.
- Branch: `refactor/490-shepherd-workflow-engine`; base/PR target `feat/471-pi-agent-session-shepherd`.
- Exact local engine: `pi-workflow-engine@0.12.0`.
- Registry integrity: `sha512-DX+e2U03raK8o8YbwnDUcAQSKNZm0v1J6jWS+bk2j2kEFihLmZCf0sUlrHWou1kWC3Zw+CA4HCgqpjLWlmtcRg==`.
- Adoption: partial; production `ProductionAgentSessionPort` retained.
- Existing `.pi/settings.json` modification was present at task start and registers the user-approved exact package. `.pi/.workflow-runs/**` is now explicitly ignored and `.pi/npm/**` remains ignored; both are local/non-authoritative and excluded from delivery.

## RED plan

| ID | Required failing contract before production edit | Expected RED |
|---|---|---|
| R1 | stable Pi `0.80.10` is accepted | current exact `0.80.6` gate rejects |
| R2 | bounded policy rejects earlier, later, prerelease, malformed, and mixed family versions | no shared policy exists |
| R3 | harmless unknown AgentSession event is ignored | current parser throws `invalid, unbounded, or terminal-sequence event` |
| R4 | unknown non-authoritative event cannot invalidate a valid typed terminal result | current event state machine rejects before result validation |
| R5 | prompt fulfillment without a typed terminal result fails closed | retained terminal-result validation |
| R6 | duplicate/malformed/stale binding terminal result fails closed | retained typed binding authority |
| R7 | claimed cwd and exact scoped host tools remain the only session authority | retained construction contract |
| R8 | cancellation and abort/join settle accepted resources before release | retained lifecycle contract |
| R9 | production source has no workflow-engine import/deep import | partial-adoption contract test/source assertion |
| R10 | all 17 production-matrix rows remain green | full focused/matrix suite |

Production files stay unchanged until R1–R4 execute and fail for their intended assertions. Missing-module/file-load failures do not count as RED.

## Evidence

| Stage | Command/evidence | Result |
|---|---|---|
| GSD doctor | `scripts/gsd doctor` | PASS |
| programming loop | `scripts/gsd prompt programming-loop init --phase 490 --dry-run` | unavailable; manual fallback recorded |
| parallel analysis | workflow run `43dc4d81-66bf-4642-bed2-207a00d5fec0` | PASS; typed partial-adoption synthesis |
| baseline focused | `node --test .pi/extensions/shepherd/agent-session-runtime.test.ts .pi/extensions/shepherd/sdk-runner.test.ts` | PASS — 158/158 before #490 RED |
| RED | focused `--test-name-pattern` over the five #490 contracts | EXPECTED FAIL — 0/5 passed; Pi 0.80.10 rejected in both runtimes, event-free typed results rejected in both runtimes, harmless unknown event rejected |
| GREEN focused | compatibility + runtime + SDK + tool-policy suites with `SHEPHERD_PI_PACKAGE_ROOT` bound to exact Pi 0.80.10 | PASS — 170/170 |
| full Shepherd before review | `node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts` with exact Pi root | PASS — 1717 pass, 0 fail, 1 intentional skip |
| strict Pi 0.80.10 typecheck | TypeScript 5.9.3 `--noEmit --strict --target ES2024 --module NodeNext`, production Shepherd sources, exact Pi 0.80.10 base URL | PASS |
| exact Pi family verifier | `SHEPHERD_PI_PACKAGE_ROOT=...pi-coding-agent node .github/scripts/verify-shepherd-pi-runtime.mjs` | PASS — coding-agent/core/ai/tui all exact 0.80.10 |
| offline RPC | isolated AgentSession command load, workflow-engine package load, and co-load through Pi 0.80.10 RPC | PASS — `pm-shepherd` plus `workflow`, `workflow:dynamax`, `workflow:inspector`, `workflow:models`, `workflow:results`, and `workflow:runs` |
| exact-head review | workflow run `8e04bba9-012b-422f-89cb-4d845f340e3e`, exact `daaa2263...23345f26`, Codex gpt-5.6-sol xhigh | FINDINGS — five actionable items; all dispositioned without a second substantive model review |
| failed review preflight | workflow run `abaefcd2-90cf-4569-9ac6-b7e58d0428e9` | stopped before code review because the requested child SHA was mistyped/nonexistent; no substantive assessment |
| post-review affected | 207 focused tests + strict typecheck + Pi family + package provenance + offline RPC | PASS |
| final full Shepherd | one post-disposition sequential run | PASS — 1717 pass, 0 fail, 1 intentional skip in 149.9s |

## RED checkpoint detail

The focused command executed five top-level behavior tests and failed exactly all five intended assertions. No file-load, compile, missing-module, timeout, skip, cancellation, or todo contributed. Three failures show the exact Pi `0.80.6` pin rejecting `0.80.10`; two show lifecycle-event authority rejecting an otherwise valid typed handoff, including the harmless unknown-event reproduction. Production files remained unchanged.

## Review-fix checkpoint

The single substantive exact-head review found delayed compatibility assertion, dirty workflow-run state, missing reproducible engine CI, generic advisory-tool authority, and incomplete telemetry callback/pre-materialization bounds. Fixes moved compatibility rejection to extension entry, ignored non-authoritative run records, added exact package/RPC CI, prohibited built-ins and all workflow-agent tools for the adopted path, and charged every telemetry callback before bounded tool-name inspection. The affected 207-test slice and all non-full gates passed before the one final complete suite.
