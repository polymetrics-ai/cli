# Overview

Plausible is a wave2 fan-out migration of `internal/connectors/plausible` (the hand-written
connector it replaces). It reads Plausible Analytics sites and stats reports (aggregate, timeseries,
and property breakdowns) through the Plausible Stats API. This bundle is read-only, matching legacy
exactly; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `api_token` as a secret; it is sent as a Bearer token (`Authorization: Bearer <api_token>`)
and never logged.

## Streams notes

- `sites` (`GET /sites`, records at `sites`) — lists sites available to the token. Real Plausible
  site objects have only a `domain` field (no separate `site_id`); `site_id` is a `computed_fields`
  rename from `domain`, matching legacy's `first(item, "site_id", "domain")`/
  `first(item, "domain", "site_id")` mutual fallback, which in practice always resolves both sides
  from the same `domain` key since Plausible's wire shape never emits a distinct `site_id`. Not
  paginated (matches legacy's `endpoint.paginated: false` for this stream — a single request only,
  even though Plausible's real `/sites` response includes `meta.after`/`before` cursor fields
  legacy's own `sites` stream never reads).
- `aggregate` (`GET /stats/aggregate`, records at `results`) — requires config `site_id`. Real
  Plausible wire shape wraps every metric in a `{"value": N}` object (confirmed against Plausible's
  published example response), so each metric's `computed_fields` entry reaches one level deeper
  (`{{ record.visitors.value }}`) than `timeseries`/`breakdown` — matching legacy's `metricValue`
  helper, which unwraps a `{value: ...}` object when present. `results` is a single JSON object, not
  an array; the engine's `RecordsAt` wraps a lone object as one record, identical to legacy's own
  `emitRecords`/`connsdk.RecordsAt` call.
- `timeseries` (`GET /stats/timeseries`, records at `results`) — requires config `site_id`. Metrics
  are bare (unwrapped) scalars on this endpoint (confirmed against Plausible's published example and
  legacy's own unit test fixture), so each metric's `computed_fields` entry is a plain
  `{{ record.<metric> }}` reference with no `.value` unwrap.
- `breakdown` (`GET /stats/breakdown`, records at `results`) — requires config `site_id` and
  `property`; the only paginated stream (matches legacy's `endpoint.paginated: true`).

All three stats streams send `site_id`, `period` (default `30d`, matching legacy's
`valueOrDefault(cfg.Config["period"], "30d")`), and the optional `date`/`metrics`/`filters`/
`compare` query params (each declared with `omit_when_absent: true` — sent only when configured,
matching legacy's `for _, key := range [...] { if v := cfg.Config[key]; v != "" { q.Set(key, v) } }`
loop). `breakdown` additionally sends `property` (default `event:page`, matching legacy's
`valueOrDefault(cfg.Config["property"], "event:page")`).

Pagination (`breakdown` only) is page-number-based (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`, `start_page: 1`, `page_size: 100`, matching legacy's
`defaultPageSize`) — the engine stops when a page returns fewer records than `page_size`, identical
to legacy's own `count < pageSize` stop condition.

Legacy never sends an incremental filter parameter for any stream (there is no cursor-based
`request_param`; Plausible's `date`/`period` are report-scoping inputs, not sync cursors), so no
`incremental` block is declared here — every stream is full-refresh only, matching legacy's actual
(non-incremental) read behavior.

## Write actions & risks

None. Plausible is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`breakdown`'s `property_value` field only supports the default `property=event:page` dimension**
  (projected from the raw `page` field). Plausible's real API names the breakdown dimension field
  after whichever `property` was requested (`event:page` -> `page`, `visit:source` -> `source`,
  `visit:country` -> `country`, etc. — confirmed against Plausible's published API docs); legacy's
  own `breakdownRecord` defensively tries ten possible dimension field names
  (`first(item, "page", "source", "referrer", "utm_campaign", "country", "region", "city",
  "browser", "os", "device")`) to cover whichever one the caller configured. The engine's
  `computed_fields` dialect has no mechanism to select a template BASED ON another config value's
  runtime value (a template is a fixed string chosen at bundle-authoring time, not itself
  parameterized by `config.property`), so this bundle implements only the single, most common
  dimension (`event:page`, legacy's own default) and treats every other `property` value as out of
  scope rather than silently mis-projecting. This is the identical shape to searxng's documented
  subreddit-narrowing deviation (`docs/migration/conventions.md` §5 item 7) — the base/default case
  is implemented at full parity, a config-driven variant case is not, and is documented here rather
  than silently wrong.
- The full Plausible API surface (goals-scoped breakdowns, realtime visitor counts, funnels, CSV
  export, site provisioning/management) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- `page_size`/`max_pages` config overrides from legacy (`intConfig` reading `config.page_size`/
  `config.max_pages`) have no runtime-config-driven equivalent in this engine dialect
  (`PaginationSpec.PageSize`/`MaxPages` are bundle-fixed values, never read from `RuntimeConfig`) —
  they are therefore not declared in `spec.json` at all (a declared-but-unwireable key is worse than
  an absent one, per the F6 dead-config rule) rather than accepted but silently ignored.
