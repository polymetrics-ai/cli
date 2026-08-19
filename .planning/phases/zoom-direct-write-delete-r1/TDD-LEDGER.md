# TDD ledger — Zoom direct-write and delete proving cohort

| Slice | Red | Green | Refactor | Status |
| --- | --- | --- | --- | --- |
| Write/delete cohort | Source port validation fails: post-foundation `api_surface` accepts direct write coverage only as a reverse-ETL write action or fixed GraphQL operation, neither of which is honest for typed REST direct writes. | No valid direct-write subset exists. The rejection list records all 61 affected REST writes (18 deletes) as recoverable `foundation-gap`; eight upload operations are schema-incompatible with foundation-gap recovery. | Restore invalid uncommitted declarations; retain only connector-local evidence and rejection accounting. | deferred |

No live write may run until a disposable, run-owned resource ledger proves create, independent readback, cleanup, and no leak.
