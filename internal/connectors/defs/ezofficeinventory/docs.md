# Overview

EZOfficeInventory is an asset and inventory management API. This bundle originally migrated the 5
legacy `internal/connectors/ezofficeinventory` read streams to a Tier-1 defs bundle at full
capability parity (`assets`, `inventories`, `asset_stocks`, `members`, `locations`), and has since
been expanded (Pass B full-surface pass) to cover the full practical v1 `.api` surface documented at
https://ezo.io/ezofficeinventory/developers/: 3 new read streams (`groups`, `vendors`,
`purchase_orders`) and 11 new write actions (create/update for assets, members, locations, groups,
vendors; create for purchase orders). See `api_surface.json` for the complete documented-endpoint
disposition (covered vs. excluded-with-reason).

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
- `groups` (`GET /assets/classification_view.api`, records at `groups`) — primary key `id`. The
  real response wraps each list element one level deeper than every other stream here
  (`{"groups": [{"group": {...fields...}}, ...]}` — confirmed against the community Airbyte
  `source-ezofficeinventory` connector's manifest, which documents the identical
  `field_path: [groups, "*", group]` extractor shape), so `records.path: "groups"` alone would
  project a record shaped `{"group": {...}}` with no top-level `id`/`name`/etc. `computed_fields`
  re-projects every field from `record.group.<field>` back to a flat top-level name (`"id": "{{
  record.group.id }}"`, etc.) — this is the sanctioned use of `computed_fields`' raw-record access
  (conventions.md §3) to normalize a one-level-wrapped list item, not a schema/parity deviation.
- `vendors` (`GET /assets/vendors.api`, records at `vendors`) — primary key `id`. Same
  per-item-wrapped shape as `groups` (`{"vendors": [{"vendor": {...}}, ...]}`, also confirmed
  against the community Airbyte manifest), unwrapped the same way via `computed_fields`.
- `purchase_orders` (`GET /purchase_orders.api`, records at `purchase_orders`) — primary key `id`.
  Flat (not per-item-wrapped) response shape, confirmed against the community Airbyte manifest's
  `field_path: [purchase_orders]` extractor (no further nesting) and its `required: [id]` schema.

`groups` and `vendors` do not send the shared `page`/`per_page` pagination params as literally as
the other streams from a "confirmed by official v1 docs" standpoint — the official docs page shows
curl examples for these two endpoints using a bare `?page=<PAGE_NUM>` query with "25 per page"
stated as a fixed fact, not a configurable `per_page`-style parameter; `size_param` is left
declared (matching every other stream's shared `base.pagination` block, which cannot be
overridden per-stream without a full pagination block redeclaration) and is harmless if the live
API silently ignores an unrecognized param, which is the documented behavior for unlisted
parameters on this API.

`page_size`/`max_pages` config knobs from legacy are not declared in `spec.json`: pagination
fields (`streams.json`'s `base.pagination` block) are plain Go values, not template-interpolated,
so there is no mechanism to wire a runtime config value into them at all — declaring a spec
property no template anywhere in the bundle ever consumes would be dead config. The bundle uses a
fixed `page_size: 25`, matching legacy's own default (`ezoDefaultPageSize = 25` in
`ezofficeinventory.go`, also recorded as `metadata.json`'s `batch.read_page_size` for operator
awareness) and no `max_pages` cap (unbounded, matching legacy's own default of 0/unlimited).

## Write actions & risks

`capabilities.write` is now `true` (Pass B expansion added `writes.json`; legacy itself never had a
write path, so there is no legacy write behavior to preserve parity with — every action below is a
genuinely new capability, sourced directly from the documented v1 form-encoded endpoints, not a
migration of existing legacy code). All 11 actions send `body_type: "form"` with bracket-nested
field names exactly matching the documented curl examples' form parameters (e.g.
`fixed_asset[name]`, `user[email]`, `location[name]`, `group[name]`, `vendor[name]`) — the same
bracket-keyed-form-field pattern already established by chargebee's `create_card_payment_source`
(`card[number]`, etc.), since `write.go`'s `buildForm` passes record field names through verbatim
with no nested-object flattening.

- `create_asset` / `update_asset` — `POST`/`PUT /assets(/{{ record.id }}).api`. Required on create:
  `fixed_asset[name]`, `fixed_asset[group_id]`, `fixed_asset[location_id]`.
- `create_member` / `update_member` — `POST`/`PUT /members(/{{ record.id }}).api`. Required on
  create: `user[email]`, `user[first_name]`, `user[last_name]`, `user[role_id]`.
- `create_location` / `update_location` — `POST`/`PUT /locations(/{{ record.id }}).api`. Required
  on create: `location[name]`.
- `create_group` / `update_group` — `POST`/`PUT /groups(/{{ record.id }}).api`. Required on
  create: `group[name]`.
- `create_vendor` / `update_vendor` — `POST`/`PUT /vendors(/{{ record.id }}).api`. Required on
  create: `vendor[name]`.
- `create_purchase_order` — `POST /purchase_orders.api`. Required: `vendor_id` (this one field is
  NOT bracket-nested — the documented curl example sends a bare `vendor_id=<VENDOR_ID>` form
  field, unlike every other create action here).

All 11 actions are approval-gated (`risk` field on each). Delete endpoints
(`DELETE /assets/<id>.api`, `/inventory/<id>.api`, `/stock_assets/<id>.api`, `/groups/<id>.api`,
`/groups/<gid>/sub_groups/<id>.api`, `/purchase_orders/<id>.api`) exist in the documented surface
but are deliberately NOT migrated to writes.json in this pass — see `api_surface.json`'s
`destructive_admin` exclusions; irreversible deletes are held back pending a dedicated
approval/risk review rather than being added under the same blanket "approval required" risk text
as a reversible create/update.

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
- **`groups`/`vendors`'s per-item nested-wrap unwrap is a real, permanent API response shape, not
  a workaround.** See "Streams notes" above; verified independently against the community Airbyte
  `source-ezofficeinventory` connector's own extractor definitions
  (`field_path: [groups, "*", group]` / `[vendors, "*", vendor]`), not merely a project-side guess.
- **Several documented create/update endpoints are NOT migrated because their request-body field
  contract is undocumented beyond a bare curl skeleton** (inventory create/update, asset-stock
  create/update, work order create, project create, bundle create, purchase-order update): the v1
  docs page shows the endpoint existing but not its full field list, unlike assets/members/
  locations/groups/vendors/purchase-order-create, which the docs page shows complete
  `resource[field]`-style form parameter lists for. Guessing an unverified required-fields contract
  for a mutation would risk silently rejecting or mis-shaping real writes; these are left
  unmigrated rather than approximated. See `api_surface.json`'s `out_of_scope` entries for each.
- **`bundles`/`work_orders` are not added as read streams.** `bundles.api`'s real response (per the
  community Airbyte connector's own schema) has no confirmed primary-key `id` field this engine's
  schema (`x-primary-key` required) can honestly declare; `work_orders.api`'s documented GET
  behavior only covers work-order TYPES, not a confirmed paginated work-order list envelope. Both
  are left out rather than inventing an unverified shape. See `api_surface.json`.
- **Compound/lifecycle actions (checkin, checkout, retire, activate, mark_confirm, approve/reject,
  reservations, mass-* bulk actions) are out of scope for this pass.** These are multi-field
  state-transition or bulk-target actions, not the single-resource create/update shape
  `writes.json` models; see `api_surface.json`'s `out_of_scope` entries (one per endpoint) for the
  specific reason each was excluded.
