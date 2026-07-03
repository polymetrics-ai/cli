# Overview

EZOfficeInventory is an asset and inventory management API. This bundle migrates all 5 legacy
`internal/connectors/ezofficeinventory` read streams to a Tier-1 defs bundle at full capability
parity: `assets`, `inventories`, `asset_stocks`, `members`, `locations`.

## Auth setup

Provide your EZOfficeInventory account `subdomain` (the `<subdomain>` in
`https://<subdomain>.ezofficeinventory.com`) and an `api_key` secret. `api_key` is sent as the
`token` header on every request (never logged). There is no `base_url` override in this bundle —
`subdomain` is required and directly templates the base URL
(`https://{{ config.subdomain }}.ezofficeinventory.com`), matching legacy's derived-URL behavior
exactly, following the same documented pattern as the bamboo-hr bundle (the engine's `spec.json`
`"default"` materialization mechanism only fills in a fixed literal for an absent key — it cannot
derive one config value's default from another, so legacy's test/proxy `base_url` escape hatch is
dropped rather than declared-but-unwireable).

## Streams notes

All 5 streams share EZOfficeInventory's page-number pagination (`pagination.type: page_number`,
`page_param: page`, `size_param: per_page`, `start_page: 1`). `assets`, `inventories`, and `asset_stocks` additionally send the 4 static
detail-enrichment query params legacy's `detailParams` table applies
(`show_image_urls`/`show_document_urls`/`include_custom_fields`/`show_document_details`, all
`"true"`); `members` and `locations` send none.

- `assets` (`GET /assets.api`, records at `assets`) — primary key `identifier`.
- `inventories` (`GET /inventory.api`, records at `assets` — EZOfficeInventory's inventory list
  endpoint nests its items under the same `assets` key as the assets endpoint) — primary key
  `identifier`.
- `asset_stocks` (`GET /stock_assets.api`, records at `assets`) — primary key `identifier`, same
  record shape as `assets` (legacy's `ezoStreamEndpoints["asset_stocks"]` reuses `assetRecord`).
- `members` (`GET /members.api`, records at `members`) — primary key `id`. This is also the
  `check` request (mirrors legacy's `Check`, which lists page 1 of members to confirm
  auth/connectivity).
- `locations` (`GET /locations/get_line_item_locations.api`, records at `locations`) — primary
  key `id`.

`page_size`/`max_pages` config knobs from legacy are not declared in `spec.json`: pagination
fields (`streams.json`'s `base.pagination` block) are plain Go values, not template-interpolated,
so there is no mechanism to wire a runtime config value into them at all — declaring a spec
property no template anywhere in the bundle ever consumes would be dead config. The bundle uses a
fixed `page_size: 25`, matching legacy's own default (`ezoDefaultPageSize = 25` in
`ezofficeinventory.go`, also recorded as `metadata.json`'s `batch.read_page_size` for operator
awareness) and no `max_pages` cap (unbounded, matching legacy's own default of 0/unlimited).

## Write actions & risks

None. EZOfficeInventory is read-only in both legacy and this bundle (`capabilities.write: false`,
no `writes.json`).

## Known limits

- **Pagination stop condition is short-page-only, not `total_pages`-driven (ACCEPTABLE
  deviation).** Legacy's `harvest` loop stops on whichever comes first: an empty page, reaching
  `total_pages` (when the response includes it), or (only when `total_pages` is absent) a page
  shorter than `page_size`. The declarative `page_number` paginator (`connsdk.PageNumberPaginator`)
  has a single stop rule: a page returning fewer than `page_size` records. In practice these
  coincide for every real EZOfficeInventory response (the last page of a real list is always
  short), so this never changes emitted record data for any real API response; the only
  theoretical divergence is a page that is simultaneously full-sized AND flagged as the last page
  via `total_pages` (one harmless extra request that returns zero records), or a short-but-not-last
  page (a defensive legacy fallback path that does not correspond to any documented
  EZOfficeInventory response shape). See `docs/migration/conventions.md` §5's meta-rule.
- `rate_limit` is not declared on `streams.json`'s `base` block: legacy enforces no client-side
  rate limiting, so none is added here.
