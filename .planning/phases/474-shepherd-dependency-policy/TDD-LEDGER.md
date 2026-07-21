# Issue #474 TDD Ledger

GSD mode: `manual_gsd_fallback` because the repository adapter does not register
`programming-loop`.

| Slice | RED evidence | GREEN evidence | Refactor/broader evidence | Status |
|---|---|---|---|---|
| Lifecycle and retry policy | `node --test ...` failed because `autonomy-policy.ts` did not exist | pending | pending | red |
| DAG/scopes/maximum ready queue | same run failed because `dependency-graph.ts` did not exist | pending | pending | red |
| Pure idempotent reconciler | same run failed because `reconciler.ts`/its imports did not exist | pending | pending | red |

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
