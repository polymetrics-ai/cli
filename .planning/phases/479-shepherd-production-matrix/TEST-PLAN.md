# Correction test plan

Each production edit follows behavior RED, minimal GREEN, then focused refactor/regression.

| Slice | RED behavior | GREEN gate |
|---|---|---|
| Issue bootstrap | Missing plan starts a planning session; issue-less proposal materializes returned issue numbers; existing/invalid plans are reused/fail closed. | `production-plan-bootstrap.test.ts`, `gh-planning-issue-source.test.ts`, intake and Pi-host tests |
| Agent verification | Session can request only the next plan ID; omitted/out-of-order/fabricated results fail; a real small Go fixture passes. | `agent-session-verification.test.ts`, workspace lifecycle/runtime recovery tests |
| Trusted-local boundary | No model can supply argv/env/shell. The independent verification role receives no workspace write tool; implementation/correction roles retain only their scoped workspace edit/write tools plus ID-only host verification. | tool-policy and verification tests |
| Bounded review correction | Arbitrary Node/Git/Go/Make recipes, escaped descendants, and post-failure AgentSession exceptions fail closed. | contract/bootstrap, bounded process-tree, and AgentSession verification tests |
| Cross-layer plan parity | A proposal accepted at bootstrap is executable by host verification and consumable by real GitHub orchestration/scheduling; the actual `host_inspect` schema compiles through the closed tool policy; unsafe IDs/bounds, duplicate IDs/slugs, empty skills, inline-unsafe fields, cycles, and ambiguous scopes fail before publication. | contract, real production orchestration, bootstrap/tool-policy, and dependency-graph tests |
| Same-second ready | Exact `draft:false` at equal revision succeeds; equal draft, foreign head, restart, timeout reconciliation, and duplicate calls stay safe. | parent-ready host/lifecycle tests |
| Merge-readiness CI | No workflow runs the complete Shepherd test inventory, so the 64 sandbox-blocked cases have no ordinary-host release gate. | A least-privilege workflow runs `node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts`; workflow structure, complete inventory, strict TypeScript, offline Pi RPC, and diff hygiene are checked before handoff. |

Focused gates: affected tests, all `.pi/extensions/shepherd/*.test.ts`, strict production TypeScript,
offline Pi RPC, and `git diff --check`. Broad Go/connector gates remain parent-head gates; only the small
real Go verification fixture runs in this TypeScript correction.
