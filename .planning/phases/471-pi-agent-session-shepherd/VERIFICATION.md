# Verification

Phase: `471-pi-agent-session-shepherd`

| Check | Status | Evidence / next gate |
|---|---|---|
| GitHub parent topology | pass | #471, draft PR #472, and dependency-linked child issues #473-#481 exist. |
| Abandoned Go closure | pass | #372/#389/#470 closed `not_planned`; #390/#456 closed unmerged; history preserved. |
| Role model policy | child-branch pass | #479 tests verify implementation/correction at 5.6 Sol/high and planning/review/orchestration at 5.6 Sol/xhigh; exact parent-head replay remains. |
| Historical read-only foundation | historical pass | Earlier 82-test/full-root gate and #438 read-only canary evidence is preserved; it does not verify autonomous mutation. |
| #473-#477 foundation/ports | integrated; default-branch completion pending | PRs #482-#486 are merged into the parent. #473/#474/#476/#477 were reopened because #472 has not reached `main`; #475 was already open. |
| #478 GitHub orchestration | integrated; default-branch completion pending | PR #487 is merged into the parent; issue remains open. |
| #479 autonomous integration | integrated | PR #489 merged as parent `daaa2263`; 17/17 matrix and hosted complete-suite/security checks were green before integration. |
| #490 workflow-engine/Pi compatibility | integrated | PR #491 merged as parent `c3f4f683`; final child 1718 pass/0 fail/1 skip and Pi-family/provenance/RPC checks were green. Child review is not parent approval. |
| #480 recovery/cutover preparation | worker-ready | Dependency #479 is integrated. A persistent Shepherd worktree must capture RED before recovery/audit/cutover production edits. |
| #481 CLI Architecture canary | dependency-blocked | Requires #480. It may reconcile #397/#438 read-only, must preserve draft/unmerged #438, and may activate deprecation only after the canary passes. |
| Full TypeScript/Pi smoke | pending final | Replay strict pinned-Pi typecheck, family/provenance verification, isolated/co-loaded RPC, and one complete sequential suite on the final integrated parent SHA. |
| Root Go/build/verify | applicability pending | Do not run unrelated broad Go/connector gates. Use GitHub `verify` as repository-hosted evidence unless the final parent diff or concrete required check makes a local Go gate applicable. |
| Automated review coverage | stale after #491 integration | User-selected route is one Codex 5.6-sol xhigh parent review over the unreviewed range plus all cross-child seams. Claude/Copilot are explicitly overridden; Codex is not human approval. |
| Human merge decision | pending by design | One fresh allowlisted `approve-merge` response on #472 exact verified head. |

No credentialed connector or reverse-ETL test is required to validate Shepherd. If a future bounded
child requires live GitHub auth, it must use the host environment/keychain without printing or
passing the token to a child session.
