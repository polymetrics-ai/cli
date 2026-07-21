# Verification

Phase: `471-pi-agent-session-shepherd`

| Check | Status | Evidence / next gate |
|---|---|---|
| GitHub parent topology | pass | #471, draft PR #472, and dependency-linked child issues #473-#481 exist. |
| Abandoned Go closure | pass | #372/#389/#470 closed `not_planned`; #390/#456 closed unmerged; history preserved. |
| Role model policy | pending commit/test | Configuration now routes implementation to 5.6 Sol/high and all other roles to 5.6 Sol/xhigh; grep/smoke required. |
| Historical read-only foundation | historical pass | Earlier 82-test/full-root gate and #438 read-only canary evidence is preserved; it does not verify autonomous mutation. |
| #473 control-plane hardening | in progress | New adversarial code/tests exist; known lease/root/lifecycle/invariant blockers remain to close. |
| #474-#477 parallel ports | pending dependency | Start only after #473 integration. |
| #478 GitHub orchestration | pending dependency | Requires #474/#476/#477. |
| #479 autonomous integration | pending dependency | Requires #474-#478. |
| #480 recovery/cutover | pending dependency | Requires #479. |
| #481 CLI Architecture canary | pending dependency | Requires #480; must not bypass #397/#438 gates. |
| Full TypeScript/Pi smoke | pending final | All Shepherd tests, strict typecheck, and offline extension discovery on exact parent head. |
| Root Go/build/verify | pending final | `go vet`, `go test`, build, and `make verify` on exact parent head. |
| Automated review coverage | pending | Per-child exact ranges plus final parent coverage/dispositions. |
| Human merge decision | pending by design | One fresh allowlisted `approve-merge` response on #472 exact verified head. |

No credentialed connector or reverse-ETL test is required to validate Shepherd. If a future bounded
child requires live GitHub auth, it must use the host environment/keychain without printing or
passing the token to a child session.
