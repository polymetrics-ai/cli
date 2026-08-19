# TDD ledger — issue #4273 connector surface sweep batch 1

| Slice | RED contract | GREEN contract | Status |
| --- | --- | --- | --- |
| Fleet visibility | Baseline scan reports only 2 `sync_transport.json` files across 552 bundles; no committed resumable sweep ledger exists. | Progress JSON contains one row per bundle, all class/transport fields, totals, failure reason, timestamp, and resume state. | planned |
| Evidence selection | No new manifest exists for this issue. | `batch plan --size 20` admits only completed public/versioned OpenAPI/Swagger survey rows and writes an immutable manifest. | planned |
| Per-candidate runtime admission | Static metadata alone can claim an unavailable command. | `batch gate` records a real commandrunner preflight count for every accepted candidate, or a named drop. | planned |
| Truthful capability boundaries | Generic direct-read and transport declarations could be invented from provider methods. | The ledger reports only actual declaration fields; unsupported executor/policy constraints enter the foundation log and no connector becomes certified. | planned |

## Manual TDD fallback

This is a JSON/materialization lane, not a shared Go implementation. The RED/GREEN evidence is emitted by the existing generator's source-ledger, materialization, validation, and preflight contracts rather than by a new production test. The raw commands and observable outputs are preserved in `traces/` and the final verification record.
