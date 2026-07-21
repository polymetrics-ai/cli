# TDD Ledger: #478

| Slice | RED evidence | GREEN evidence | Refactor/broad evidence | State |
| --- | --- | --- | --- | --- |
| bounded child planning | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `github-orchestrator.ts` | bounded plan/DAG tests pass | strict owned TypeScript pass | green |
| reconcile-before-mutate idempotency | same test-only run; timeout/restart and collision contracts are present before production | timeout/restart/collision tests pass | strict owned TypeScript pass | green |
| stacked PR topology and linkage | same test-only run; parent `Closes` versus child `Refs` assertions are present | topology/linkage test passes | strict owned TypeScript pass | green |
| authoritative CI/thread/disposition evidence | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `github-evidence.ts` | all evidence cases pass | strict owned TypeScript pass | green |
| exact-head independent Codex review | focused run fails with `ERR_MODULE_NOT_FOUND` for absent `review-router.ts` | all route/rejection cases pass | strict owned TypeScript pass | green |
| scoped provisional integration | same absent-orchestrator failure; handoff/scope/head recheck cases are present | capture/scope/re-read/integration cases pass | strict owned TypeScript pass | green |
| exact-generation/head parent human gate | same absent-orchestrator failure; incomplete/pending/reject/head-move cases are present | parent readiness/broker cases pass | strict owned TypeScript pass | green |

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

## Minimal GREEN checkpoint

- Focused command: 21 pass, 0 fail.
- `review-router`: 5 pass; declarative work records only.
- `github-evidence`: 6 pass; authoritative checks, requested changes, threads, dispositions, and
  exact-range review policy.
- `github-orchestrator`: 10 pass; bounded plans, reconcile-before-mutate publication, stacked
  topology, handoff capture, roster updates, exact-head child integration, and broker-gated parent
  readiness.
- Strict no-emit TypeScript over all three owned modules and matching tests passes using TypeScript
  5.9.3 with the cached Pi 0.80.6 Node type root.

## Adversarial correction RED

The same focused command against unchanged GREEN production reports 27 total: 17 pass and 10
expected failures. Failures cover proxy-trap execution, duplicate finding ambiguity,
`not_actionable` bypass, missing parent handoff capture, unbound receipt schema, plan/review binding,
merged-PR restart reuse, and the downstream parent completeness cases that now require bound
receipts. Production remains exactly at pushed GREEN checkpoint `90321ffb` for this test-only RED.

## Adversarial correction GREEN and refactor

- Correction implementation checkpoint:
  `40ce66d4b5010b92089895a05709687143d15a05`.
- Focused command: 27 pass, 0 fail in 230.914417 ms.
- Strict owned TypeScript: pass with TypeScript 5.9.3 and the cached Pi 0.80.6 Node type root.
- Receipts now bind child ID, PR, generation, stable marker, base SHA, exact head SHA, and parent
  branch; a merged PR with an exact receipt is reused safely after restart.
- The eligible review returned by authoritative evidence must also bind the planned repository,
  work item, generation, changed paths, and exact allowed scopes before integration/readiness.
- Parent handoff setup uses the upstream `captureHandoff` boundary. Transport arrays and caller
  DTOs are descriptor-validated without invoking Proxy/accessor code. Cross-review finding IDs are
  unique, and a blocking finding requires an exact-head `fixed` disposition.
- The fake-only PR number hint was removed from the production transport contract.

## Final authorized verification

- Focused #478: 27/27 pass.
- Complete serialized Shepherd: 291 total, 290 pass, 0 fail, 1 intentional sandbox skip;
  127120.23075 ms.
- Strict all-production TypeScript: 20 production modules pass with TypeScript 5.9.3 using the
  cached Pi 0.80.6 package resolver and Node type root.
- Pinned Pi 0.80.6 offline RPC `get_commands`: `true` for `pm-shepherd` from `extension`.
- Immutable merge base equals `3addb1f48be1afe8b1e2b59b54247679d7293805`; full-range
  `git diff --check` and coordinator-owned path validation pass.
- No Go, connector, certification, runtime-service, `make`, live orchestration transport,
  reviewer, or merge command ran.
