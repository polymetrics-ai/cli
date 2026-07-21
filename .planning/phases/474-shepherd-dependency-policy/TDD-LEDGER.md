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
