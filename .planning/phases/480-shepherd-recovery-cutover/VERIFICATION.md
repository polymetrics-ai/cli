# Verification Checklist: #480

Verdict: **PLANNED — implementation RED pending**.

- [x] Parent #479 and PR #489 are integrated into the non-default parent.
- [x] Plan, TDD ledger, and verification checklist exist before production edits.
- [x] Manual-GSD fallback and required skills/references recorded.
- [ ] Unprotected-parent policy-source RED fails for the intended preflight assertion.
- [ ] Protected/unprotected policy-source GREEN passes and `/pm-shepherd start` reaches durable state.
- [ ] Focused recovery/audit RED executes and fails for intended assertions.
- [ ] Focused GREEN passes.
- [ ] Complete sequential Shepherd suite passes in the child lane.
- [ ] Strict no-emit TypeScript passes against exact Pi 0.80.10 declarations.
- [ ] Exact Pi-family and workflow-engine provenance verifiers pass.
- [ ] Offline isolated Shepherd, isolated workflow-engine, and co-loaded RPC pass.
- [ ] `git diff --check` and exact changed-path scope pass.
- [ ] One bounded exact-head Codex 5.6-sol xhigh review round has no unresolved blocker.
- [ ] GitHub checks are green on the exact child head.
- [ ] Child integrates only into `feat/471-pi-agent-session-shepherd`.
- [ ] Issue remains open pending parent/default-branch completion.

Do not run unrelated connector, certification, runtime-service, or broad Go gates in this child
worktree. Do not activate legacy-shell deprecation in #480.
