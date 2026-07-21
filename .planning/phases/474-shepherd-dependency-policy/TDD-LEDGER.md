# Issue #474 TDD Ledger

GSD mode: `manual_gsd_fallback` because the repository adapter does not register
`programming-loop`.

| Slice | RED evidence | GREEN evidence | Refactor/broader evidence | Status |
|---|---|---|---|---|
| Lifecycle and retry policy | pending | pending | pending | planned |
| DAG/scopes/maximum ready queue | pending | pending | pending | planned |
| Pure idempotent reconciler | pending | pending | pending | planned |

Rules:

- Production code is not written until the matching focused tests fail for the expected missing
  module/export behavior.
- Tests are not weakened to fit implementation.
- Every command, exit result, and exact failure/pass count is recorded after it runs.
