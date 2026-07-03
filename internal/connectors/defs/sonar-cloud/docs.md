# Overview

SonarCloud is a wave2 fan-out migration from `internal/connectors/sonar-cloud` (the legacy
hand-written connector this bundle replaces at capability parity). It reads SonarCloud issues,
components, quality gates, and measures through the SonarCloud Web API. Read-only; the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a SonarCloud user token via the `user_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <user_token>`) and is never logged.

## Streams notes

All 4 streams (`issues`, `components`, `quality_gates`, `measures`) share legacy's single-page,
non-paginated read shape: legacy's `readRecords` issues exactly one request per `Read` call (no
page-advance loop at all), so this bundle declares no `pagination` block (`type: none`, the
default) for any stream — porting a loop where legacy has none would be a behavior change, not a
migration.

Every stream sends `p=1` (Sonar's page-number param, always fixed at page 1 by legacy) and `ps`
(page size, `config.page_size`, defaulting to `100` exactly like legacy's `defaultPageSize`).
`organization` is sent when `config.organization` is set (legacy's `copyConfig` no-ops when the
config value is empty) via `omit_when_absent`. `component_keys` narrows results: legacy sends the
first comma-separated key as `component` on the `components` stream and the whole
comma-separated value as `componentKeys` on the other three streams — this bundle reproduces that
exact per-stream difference (`components`' `query.component` vs. the other three streams'
`query.componentKeys`, both templated from the same `config.component_keys`). `start_date`/
`end_date` map to `createdAfter`/`createdBefore` on every stream, matching legacy's
`copyConfig(q, cfg, "start_date", "createdAfter")`/`"end_date"→"createdBefore"`.

Legacy has no incremental/state-cursor read mode (no persisted cursor is ever read or written) —
`start_date`/`end_date` are static per-read filters, not an `incremental` block, so no stream
declares `incremental` or `x-cursor-field` here.

All 4 streams declare `"projection": "passthrough"`. Legacy's `Read` emits the raw API record
verbatim (`emit(connectors.Record(rec))`, `sonar_cloud.go:126`, inside `readRecords`) with no
field-building/filtering step — `streams()`'s four-field `Fields` list (`sonar_cloud.go:108`) is
consumed only by `Catalog`, never by `Read`. Every real SonarCloud field beyond each schema's
declared properties (e.g. `issues`' `effort`/`debt`/`hash`/`textRange`, `components`'
`qualifier`/`path`) survives to the emitted record exactly as legacy would emit it. Declaring the
default `"schema"` projection mode here would silently narrow every emitted record to the schema's
declared properties — an undocumented parity deviation from legacy's verbatim passthrough — so
`passthrough` is required, matching conventions.md §8 rule 1 (legacy's raw `emit(record)` with no
`mapRecord` field-building is the mechanical signal to use `passthrough`).

## Write actions & risks

None. SonarCloud is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- Only the 4 legacy-parity streams (`issues`, `components`, `quality_gates`, `measures`) are
  implemented; the wider SonarCloud Web API (security hotspots, projects, webhooks) is out of
  scope for this wave — see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `page_size`'s legacy bound (1-500, `maxPageSize`) is not separately re-validated by this bundle
  (the engine has no per-config-value numeric-range validation primitive) — an out-of-range value
  is passed through to SonarCloud, which will itself reject or clamp it. This is a config-surface
  narrowing, not an emitted-record-data change.
- Legacy performs no request pagination beyond the single fixed `p=1` page — matching that
  behavior exactly means this bundle's fixtures ship a single page per stream (no 2-page fixture
  required; `pagination_terminates` is exercised against a stream elsewhere in this wave's sibling
  bundles that declare real pagination).
