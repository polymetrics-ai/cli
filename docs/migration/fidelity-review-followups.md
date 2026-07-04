# Pass B fidelity review — outcomes & follow-ups (2026-07-04)

An adversarial legacy-diff review (244 agents) of the 177 passb-2/3 connectors — all of which
already pass `connectorgen validate` + fixture conformance — found **41 confirmed data-correctness
defects across 37 connectors**. Conformance replays a connector's *own* fixtures, so it cannot see
divergence from the legacy Go implementation's emitted records; this review is that missing layer.

## Fixed (26 connectors) — committed `816b0b6c`, pushed to PR #27

Each re-verified: `connectorgen validate` 0 findings + `TestConformance/<name>` PASS.

- **record_data_drift** (add `projection:"passthrough"` / restore dropped fields & cursors to match
  legacy DATA): posthog, railz, repairshopr, revolut-merchant, ringcentral, shippo,
  retailexpress-by-maropost, square, workday-rest, gong, shortcut, shortio, wikipedia-pageviews
- **invented_incremental** (drop bare `x-cursor-field` legacy never published): nasa, recreation
- **fixture_gaming** (reshape fixtures to the real wire response): help-scout, rentcast, typeform, unleash
- **api_surface_dishonest** (correct false exclusion reasons): lever-hiring, outbrain-amplify, retently
- **misc**: humanitix (restore `since` server filter + incremental), ticketmaster (`start_page` 1→0),
  microsoft-dataverse (remove citation of a nonexistent parity test), productboard

## Deferred (11 connectors) — need an engine-dialect extension

Per conventions §6, faking these in the bundle would introduce a *new* divergence, so they are
tracked here rather than silently shipped. The defect is pre-existing (introduced by the Pass B
expansion, already on PR #27); these connectors keep their current expanded state until the engine
gains the capability.

| Connector | Defect | Required engine feature |
|---|---|---|
| convex | `id = id ?? _id` clobbers a user `id` | **coalesce / fill-when-absent** computed field (type-preserving) |
| netsuite | `entity_id`/`name`/`status` need multi-path fallback | **coalesce / first-non-null** over record paths |
| sigma-computing | `name <- name else displayName`, `updated_at` fallback | **coalesce / first-non-null** |
| simplecast | cursor `updated_at <- updated_at else published_at` | **coalesce / first-non-null** |
| zoom | `name`/`updated_at`/`id` fill-when-absent over alt keys | **coalesce / fill-when-absent** |
| ebay-fulfillment | `line_item_count = len(lineItems)` | **array-length → int** computed field |
| google-forms | `item_count = len(items)` | **array-length → int** |
| adjust | hoist arbitrary `dimensions`/`metrics` sub-keys onto record | **object-flatten / sub-object spread** |
| productive | flatten JSON:API `attributes` + add `raw` copy | **object-flatten / spread** |
| plausible | record key parameterized by `config.property` last segment | **config-parameterized key extraction** (or coalesce over 10 alt keys) |
| openweather | (transient agent error during fix) | none — re-run the fix |

### Proposed engine mini-wave (TDD, additive, backward-compatible)
1. **`coalesce` computed field** — first non-null over N record paths, type-preserving (parallels the
   existing bare `{{ record.path }}` typed-extraction path in `applyComputedFields`). Unblocks 5 (+
   helps plausible). This is the ≥3-recurrence trigger from conventions §6.
2. **`length` filter** → int on an array raw value. Unblocks 2.
3. **object-flatten** (hoist named sub-objects' keys onto the record, drop the container) — larger,
   structural; unblocks adjust + productive. Consider a second increment.
4. plausible's config-parameterized key extraction — niche (1); may stay deferred.

Then re-run the fix agents for the unblocked connectors + re-verify (validate + conformance +
a targeted fidelity re-review).
