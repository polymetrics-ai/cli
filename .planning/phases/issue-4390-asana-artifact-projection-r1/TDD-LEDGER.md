# TDD ledger — Issue #4390 Asana artifact projection

## Red

- Hide one loader-recognized artifact from a copied filesystem view. The focused projection validator must reject the lane and name the unavailable artifact.
- Remove the direct-read backlink from one copied applicable matrix cell. The focused projection validator must reject the source ID/lane rather than accepting aggregate counts.
- Treat a `mapped_unproven` ETL row as a stream binding. The focused projection validator must reject that promotion because it has descriptor-only evidence and no selected stream/schema/API route.

## Green

- All seven enabled-contract lane artifacts are present in the actual Asana definition directory and every lane has the Track A matrix link.
- Direct read resolves 119 source-backed GET cells; direct write and reverse ETL each resolve 130 mutation cells, including 23 DELETE source rows; binary download remains source-evidenced not applicable; binary upload resolves exactly one attachment source row.
- ETL resolves exactly 12 stream/schema/API cells and exactly 52 `mapped_unproven` descriptor-only cells; none of the latter claims a `streams.json` route.
- Sync resolves exactly the event, hydration, and snapshot source IDs selected by `event_source_contract.json` and `sync_transport.json`; its source evidence retains `event_total_order=not_documented`.
- The connector-local ledger retains all seven existing foundation records with their original IDs/states/reasons, projects their exact source/lane evidence, and adds one new 52-cell ETL mapping-deficit entry. This remains overlapping evidence, never a second 249-row denominator.

## Retained baseline regression evidence

- `TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation` remains unchanged and currently fails on pre-existing source-projection coverage findings. This task does not change its source mutation dispositions, descriptor, writes, operations, API surface, or CLI surface; `missing-foundation.json` is not read by that test path. The failure is recorded rather than suppressed or made green by altering an out-of-scope projection input.

## Refactor

- Keep matrix ordering and all source identities stable. Use existing contract and connector artifact conventions; do not introduce a second source denominator, a runtime executor, or a provider-specific shared schema.
