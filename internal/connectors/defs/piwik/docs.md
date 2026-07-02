# Overview

Piwik / Matomo is a wave2 fan-out migration of `internal/connectors/piwik` (the hand-written
connector it replaces). It reads Piwik/Matomo sites, recent visit details, page-action metrics, and
configured goals through the Matomo Reporting API's single `index.php?module=API` entry point. This
bundle is read-only, matching legacy exactly; the legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Provide `token_auth` as a secret; it is sent as the `token_auth` query parameter
(`api_key_query` mode) on every request and never logged. A `base_url` config value (defaulting to
the same placeholder legacy used, `https://matomo.example.com` — a real Matomo instance origin must
always be supplied) points at the Matomo instance.

## Streams notes

Every stream hits the same `GET /index.php` endpoint with a different `method` query parameter
(Matomo's RPC-over-REST convention), matching legacy's `streamEndpoints` table exactly:

- `sites` — `method=SitesManager.getAllSites`, no site scoping required, records at the top-level
  JSON array. `site_id` renames the raw `idsite` field (schema projection copies by exact key match
  only; the rename is required or the field would silently drop).
- `visits` — `method=Live.getLastVisitsDetails`, requires config `site_id`, records at the top-level
  array, paginated.
- `actions` — `method=Actions.getPageUrls`, requires config `site_id`, records at the top-level
  array, paginated. `hits`/`visits` rename the raw `nb_hits`/`nb_visits` fields.
- `goals` — `method=Goals.getGoals`, requires config `site_id`, records at the top-level array, not
  paginated (matches legacy's `endpoint.paginated: false`, which always returns after the first
  page regardless of the response size).

`period` (default `day`) and `date` (default `today`) are sent as `period`/`date` query params on
every site-scoped stream, matching legacy's `valueOrDefault(cfg.Config["period"], "day")`/
`valueOrDefault(cfg.Config["date"], "today")` fallbacks — both now materialize via `spec.json`'s
`"default"` mechanism rather than a template-level default.

Pagination (`visits`, `actions`) is offset+limit using Matomo's own param names
(`pagination.type: offset_limit`, `limit_param: filter_limit`, `offset_param: filter_offset`,
`page_size: 100`, matching legacy's `defaultPageSize`) — the engine stops when a page returns fewer
records than `page_size`, identical to legacy's own `len(records) < pageSize` stop condition.
`sites`/`goals` always send `filter_limit=100&filter_offset=0` on their single request too (legacy
sends these on every call regardless of `endpoint.paginated`), declared as static per-stream `query`
entries since `pagination.type: none` streams issue exactly one request.

Legacy never sends an incremental filter parameter for any stream (no cursor-based
`request_param`), so no `incremental` block is declared here — every stream is full-refresh only,
matching legacy's actual (non-incremental) read behavior. `last_action_at` is exposed as a plain
field, not a declared cursor.

## Write actions & risks

None. Piwik/Matomo is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- Legacy's `id_site` config-key fallback (`siteID` tries `config.site_id` then `config.id_site`) is
  narrowed to `site_id` only — declaring two spec properties that both resolve to the same query
  param with no coalesce mechanism in the engine dialect would require picking one as canonical
  anyway; `site_id` is legacy's primary/first-checked key, so this narrowing never changes behavior
  for any caller already using `site_id`, only for the undocumented legacy alias.
- The full Matomo Reporting API method surface (hundreds of report methods across `VisitsSummary`,
  `Referrers`, `DevicesDetection`, `Goals` conversion reports, custom segments, multi-site batch
  requests, etc.) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 4
  legacy-parity streams are implemented.
- `page_size`/`max_pages` config overrides from legacy (`intConfig` reading `config.page_size`/
  `config.max_pages`) have no runtime-config-driven equivalent in this engine dialect
  (`PaginationSpec.PageSize`/`MaxPages` are bundle-fixed values, never read from `RuntimeConfig`) —
  they are therefore not declared in `spec.json` at all (a declared-but-unwireable key is worse than
  an absent one, per the F6 dead-config rule) rather than accepted but silently ignored.
