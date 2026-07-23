# TDD Ledger: #490

## Workflow state

- Manual-GSD fallback active: adapter health passed; `programming-loop` is absent from the 69-command registry and failed once as `unknown GSD command`.
- Required execution decision, planning cycle: `read_only_spawned` — workflow run `43dc4d81-66bf-4642-bed2-207a00d5fec0` used four read-only analysis lanes and one typed synthesis lane. The issue worker retains the only mutating path in this dedicated worktree.
- Required skill used: `gsd-core`; required Pi/runtime and issue-first references loaded.

## Baseline and provenance

- Issue: #490, parent #471, parent PR #472.
- Branch: `refactor/490-shepherd-workflow-engine`; base/PR target `feat/471-pi-agent-session-shepherd`.
- Exact local engine: `pi-workflow-engine@0.12.0`.
- Registry integrity: `sha512-DX+e2U03raK8o8YbwnDUcAQSKNZm0v1J6jWS+bk2j2kEFihLmZCf0sUlrHWou1kWC3Zw+CA4HCgqpjLWlmtcRg==`.
- Adoption: partial; production `ProductionAgentSessionPort` retained.
- Existing `.pi/settings.json` modification was present at task start and registers the user-approved exact package. `.pi/.workflow-runs/**` and `.pi/npm/**` are local/non-authoritative and excluded from delivery.

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
| GREEN focused | pending | pending |
| full Shepherd | pending | pending |
| strict Pi 0.80.10 typecheck | pending | pending |
| offline RPC/canary | pending | pending |
| exact-head review | pending | exactly one round required |
| final full Shepherd | pending | run once after review disposition |

## RED checkpoint detail

The focused command executed five top-level behavior tests and failed exactly all five intended assertions. No file-load, compile, missing-module, timeout, skip, cancellation, or todo contributed. Three failures show the exact Pi `0.80.6` pin rejecting `0.80.10`; two show lifecycle-event authority rejecting an otherwise valid typed handoff, including the harmless unknown-event reproduction. Production files remained unchanged.
