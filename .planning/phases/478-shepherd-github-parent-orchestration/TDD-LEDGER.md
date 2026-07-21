# TDD Ledger: #478

| Slice | RED evidence | GREEN evidence | Refactor/broad evidence | State |
| --- | --- | --- | --- | --- |
| bounded child planning | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `github-orchestrator.ts` | pending | pending | red |
| reconcile-before-mutate idempotency | same test-only run; timeout/restart and collision contracts are present before production | pending | pending | red |
| stacked PR topology and linkage | same test-only run; parent `Closes` versus child `Refs` assertions are present | pending | pending | red |
| authoritative CI/thread/disposition evidence | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `github-evidence.ts` | pending | pending | red |
| exact-head independent Codex review | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `review-router.ts` | pending | pending | red |
| scoped provisional integration | same absent-orchestrator failure; handoff/scope/head recheck cases are present | pending | pending | red |
| exact-generation/head parent human gate | same absent-orchestrator failure; incomplete/pending/reject/head-move cases are present | pending | pending | red |

## Initial state

- Exact base: `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Production files and matching tests do not exist at plan time.
- RED will be committed with tests/fixtures only before any production file is added.
- The missing GSD adapter command is recorded as `manual_gsd_fallback`; strict test-first behavior
  remains mandatory.

## RED checkpoint

Command:

```bash
node --test .pi/extensions/shepherd/github-orchestrator.test.ts \
  .pi/extensions/shepherd/review-router.test.ts \
  .pi/extensions/shepherd/github-evidence.test.ts
```

Result: exit 1 with 0 pass / 3 test-file failures. Each matching test file fails deterministically
with `ERR_MODULE_NOT_FOUND` for its intentionally absent production module. The first attempted run
also exposed and corrected one test-only illegal nested `await`; the recorded RED run contains no
test syntax error. `scripts/tdd-gate.mjs` is not present, so the command output plus unchanged
production-file absence is the manual strict-TDD gate.
