# Connector surface sweep progress

Generated from `internal/connectors/defs/` at the issue #4273 baseline. This is a declaration inventory, not certification or live-provider proof.

| Measure | Total |
| --- | ---: |
| Connector rows | 552 |
| Pending | 552 |
| Materialized / validated / gated | 0 / 0 / 0 |
| Failed / skipped | 0 / 0 |
| Direct-read declarations | 23 |
| Direct-write declarations | 6 |
| ETL transport declarations | 2 |
| Reverse-ETL transport declarations | 2 |
| Binary declarations | 14 |
| `sync_transport.json` files | 2 |
| Legacy stream bundles | 548 |
| Legacy write bundles | 240 |

## Resume pointer

`batch 1` is ready to plan from the external provider-artifact ledger recorded in [`progress.json`](cli-surface-sweep-batches-r1-progress.json). The next worker must not infer capabilities from a stream or write file: update each selected row only from the materialize/gate report and record any absence as a foundation gap.
