# Issue #474 TDD Ledger

GSD mode: `manual_gsd_fallback` because the repository adapter does not register
`programming-loop`.

| Slice | RED evidence | GREEN evidence | Refactor/broader evidence | Status |
|---|---|---|---|---|
| Lifecycle and retry policy | `node --test ...` failed because `autonomy-policy.ts` did not exist | focused suite 23/23 pass | strict production typecheck pass | green |
| DAG/scopes/maximum ready queue | same run failed because `dependency-graph.ts` did not exist | focused suite 23/23 pass | strict production typecheck pass | green |
| Pure idempotent reconciler | same run failed because `reconciler.ts`/its imports did not exist | focused suite 23/23 pass | strict production typecheck pass | green |

Rules:

- Production code is not written until the matching focused tests fail for the expected missing
  module/export behavior.
- Tests are not weakened to fit implementation.
- Every command, exit result, and exact failure/pass count is recorded after it runs.

## RED checkpoint

Command:

```bash
node --test .pi/extensions/shepherd/autonomy-policy.test.ts \
  .pi/extensions/shepherd/dependency-graph.test.ts \
  .pi/extensions/shepherd/reconciler.test.ts
```

Observed: 3 file-level tests, 0 pass, 3 fail. Each failed with `ERR_MODULE_NOT_FOUND` for the
intentionally absent production modules. The surrounding evidence wrapper then also reported
`zsh: read-only variable: status`; this shell-wrapper mistake did not cause or conceal the three
expected Node failures and will not be reused.

## GREEN checkpoint

After the minimum three pure modules were added, the first focused run exposed one Node strip-mode
syntax incompatibility (a TypeScript constructor parameter property): 6 tests passed and 2 test
files failed before loading. Replacing that syntax with an explicit readonly field produced:

```text
tests 23
pass 23
fail 0
```

The existing TypeScript 5.9.3 compiler available in the environment then ran with `--noEmit
--strict --target ES2022 --module NodeNext --moduleResolution NodeNext
--allowImportingTsExtensions --skipLibCheck` over the three production modules. It first found one
implicit-any callback introduced by runtime `Array.isArray` narrowing; after the minimal annotation,
the same strict command exited 0. The focused 23/23 suite and `git diff --check` also exited 0.

## Refactor gap loop

An adversarial pass added four fail-closed expectations before editing production code:

- runtime-invalid lifecycle/failure vocabulary;
- empty DAG completion;
- terminal completion independent of capabilities used only for future spawns; and
- invalid concurrency data returning a typed repository blocker instead of throwing.

Gap RED: 26 tests, 22 pass, 4 fail with the expected uncaught/incorrect decisions. Gap GREEN after
the minimal validation and decision-order changes: 26/26 pass. Strict production TypeScript and
`git diff --check` pass. The compact full Shepherd suite then passed 163/163.

## Exact-head correction loop

Status: RED tests pending. Production edits are prohibited until the new adversarial expectations
fail for the reviewed `28f165412de4c8165ba7717a1690c36b00af8857` behavior.

| Correction slice | Expected RED signal | Status |
|---|---|---|
| Authenticated resumable human decision and terminal abort | missing stage/result/guards | pending |
| Bounded conflict-component scheduling | hostile 64-item case exceeds deadline or lacks typed rejection | pending |
| Code-unit ordering and Darwin/Git aliases | locale/case/NFD expectations differ | pending |
| Dependency/status coherence and exact DTO validation | incoherent/hostile values are accepted or throw | pending |
| Reconciliation precedence and selected-only isolation | wrong blocker/result or safe readers suppressed | pending |
| Correction-required and ordinary evidence handling | unsafe advancement or human-gate misclassification | pending |
