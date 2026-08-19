# TDD ledger — issue #4273 connector surface sweep batch 1

| Slice | RED contract | GREEN contract | Status |
| --- | --- | --- | --- |
| Fleet visibility | `2026-08-19` baseline reported only 2 `sync_transport.json` files across 552 bundles; no committed resumable sweep ledger existed. | Progress JSON contains 552 unique rows, all five class/transport fields, totals, named failure/skip reason, timestamp, and a batch-2 resume pointer. | completed |
| Evidence selection | No issue-scoped manifest existed. | `batch plan --size 20` wrote an immutable 20-candidate manifest from the existing provider-artifact ledger; a generator-produced eight-survivor manifest is retained after named drops. | completed |
| Per-candidate runtime admission | Static metadata alone can claim an unavailable command. | Materialization named 8 included and 12 dropped candidates; `batch gate` recorded real commandrunner preflight evidence for all 99 survivor commands. | completed |
| Truthful capability boundaries | Generic direct-read and transport declarations could be invented from provider methods. | The ledger preserves the actual empty parity-class/transport state for all eight survivors; G12--G16 record executor, policy, transport, and bounded-reconcile gaps; no certification was claimed. | completed |
| Generated CLI transcript | The first comprehensive CLI run failed because root command manuals gained four generated connector entries while the nine root golden snapshots retained the old surface. | The targeted generator test refreshed only those nine root snapshots; the targeted golden test then passed before the comprehensive rerun. | completed |

## Manual TDD fallback

This is a JSON/materialization lane, not a shared Go implementation. The RED/GREEN evidence is emitted by the existing generator's source-ledger, materialization, validation, and preflight contracts rather than by a new production test. The raw commands and observable outputs are preserved in `traces/` and the final verification record.
