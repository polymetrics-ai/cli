# Verification

Phase: `471-pi-agent-session-shepherd`

| Check | Status | Evidence / next gate |
|---|---|---|
| GitHub parent topology | pass | #471, draft PR #472, and dependency-linked child issues #473-#481 exist. |
| Abandoned Go closure | pass | #372/#389/#470 closed `not_planned`; #390/#456 closed unmerged; history preserved. |
| Role model policy | child-branch pass | #479 tests verify implementation/correction at 5.6 Sol/high and planning/review/orchestration at 5.6 Sol/xhigh; exact parent-head replay remains. |
| Historical read-only foundation | historical pass | Earlier 82-test/full-root gate and #438 read-only canary evidence is preserved; it does not verify autonomous mutation. |
| #473 control-plane hardening | capability present; lifecycle pending | Branch-local functional evidence exists; reconcile the independent issue/PR record during parent integration. |
| #474-#477 parallel ports | capabilities present; lifecycle pending | Policy, AgentSession, Git/worktree, and decision capabilities are exercised by #479; independent issue/PR lifecycle records remain. |
| #478 GitHub orchestration | capability present; lifecycle pending | Exact-effect orchestration is exercised by #479; independent issue/PR lifecycle records remain. |
| #479 autonomous integration | functional matrix pass; external gates pending | Production `78708cbe`; CI correction `a594be98`; child evidence `d895dc38`; focused 767/767; complete local 1,647 pass/64 sandbox-blocked/1 skip. Parent publication, PR, remote CI, exact-head internal review, policy coverage, and parent integration remain. |
| #480 recovery/cutover preparation | waiting #479 integration | #479 implementation is verified but not yet integrated; do not activate deprecation before the canary. |
| #481 CLI Architecture canary | pending dependency | Requires #480; must not bypass #397/#438 gates; successful canary precedes parent-owned deprecation activation. |
| Full TypeScript/Pi smoke | child-branch pass; parent final pending | #479 strict production TypeScript and offline extension discovery pass; replay on the exact integrated parent head. |
| Root Go/build/verify | pending final, parent-only | `go vet`, `go test`, build, and `make verify` run once on the exact integrated parent head, not in child lanes. |
| Automated review coverage | blocked | Codex xhigh at `ca3f6c6f` returned two blockers; the Pi-family blocker is corrected. Exact-head follow-up and source-policy `claude_auto` or allowed fallback remain pending. |
| Human merge decision | pending by design | One fresh allowlisted `approve-merge` response on #472 exact verified head. |

No credentialed connector or reverse-ETL test is required to validate Shepherd. If a future bounded
child requires live GitHub auth, it must use the host environment/keychain without printing or
passing the token to a child session.
