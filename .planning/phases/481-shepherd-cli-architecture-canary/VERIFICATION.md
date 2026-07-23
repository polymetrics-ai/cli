# Verification Checklist: #481

Verdict: **DEPENDENCY-BLOCKED — waiting for #480 integration**.

- [x] Plan, TDD ledger, and verification checklist exist before production edits.
- [x] Current #397/#438 state recorded without mutation.
- [ ] #480 exact-head handoff is integrated and clean.
- [ ] Focused deterministic canary RED fails for intended assertions.
- [ ] Focused GREEN passes.
- [x] Generation 1 attempted supported read-only live #397/#438 reconciliation at exact head `21d195aff`; #438 stayed unchanged and scout evidence completed.
- [ ] A later generation completes after the parent-owned embedded OAuth runtime preflight is green.
- [ ] #438 remains open, draft, and unmerged; no branch or issue mutation is attributed to the canary.
- [ ] Complete sequential Shepherd suite passes in the child lane.
- [ ] Strict no-emit TypeScript passes against exact Pi 0.80.10 declarations.
- [ ] Exact Pi-family and workflow-engine provenance verifiers pass.
- [ ] Offline isolated/co-loaded RPC passes.
- [ ] `git diff --check` and changed-path scope pass.
- [ ] One bounded exact-head Codex 5.6-sol xhigh review has no unresolved blocker.
- [ ] GitHub checks are green on the exact child head.
- [ ] Post-pass deprecation receipt is bound to the passing canary head and remains reversible.
- [ ] Child integrates only into `feat/471-pi-agent-session-shepherd` and issue stays open pending `main`.

Do not run credentialed connector checks, reverse ETL execution, runtime-service gates, or unrelated
connector/certification work.
