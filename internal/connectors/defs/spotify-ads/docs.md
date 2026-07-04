# Overview

Spotify Ads is a wave2 fan-out migration from `internal/connectors/spotify-ads` (the legacy
hand-written connector this bundle replaces at capability parity). It reads Spotify Ads ad
accounts, campaigns, ad sets, and ads through the Spotify Ads API. Read-only; the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Spotify Ads OAuth access token via the `access_token` secret; it is used only for
Bearer auth (`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)`.

## Streams notes

`ad_accounts` needs no path parameter. `campaigns`, `ad_sets`, and `ads` each require
`config.ad_account_id` to substitute the `{ad_account_id}` path segment (`path:
"/ad_accounts/{{ config.ad_account_id }}/campaigns"`, etc.) — legacy's `resolveResource`
hard-errors with `"stream requires config ad_account_id"` when unset; an absent
`config.ad_account_id` surfaces as a hard path-interpolation error on those three streams' reads,
reproducing the same requirement through the engine's own error path.

All 4 streams share the identical pagination shape: `offset_limit` (`limit_param: limit`,
`offset_param: offset`, `page_size: 100`) — matches legacy's `connsdk.OffsetPaginator{LimitParam:
"limit", OffsetParam: "offset", PageSize: pageSize}` with `defaultPageSize = 100`. Each stream's
`records.path` matches its own resource name (`ad_accounts`, `campaigns`, `ad_sets`, `ads`),
identical to legacy's per-endpoint `recordsPath`.

Legacy performs no incremental/state-cursor filtering during `Read` — no stream declares an
`incremental` block.

### Pass B additions

Three new read streams, added against the real, live v3.0 OpenAPI reference (fetched directly from
the `__NEXT_DATA__` JSON embedded in every `developer.spotify.com/documentation/ads-api/
reference/v3.0/<operationId>` page — see `api_surface.json`'s `scope` note for the full picture,
including this pass's discovery that the wave2 `ad_accounts` stream's flat `/ad_accounts` path does
not exist on the real API at all):

- **`businesses`** — `GET /businesses`, records at `businesses`; a business is the top-level
  account container that owns one or more ad accounts. No pagination parameters are documented for
  this endpoint (`pagination: {"type": "none"}`).
- **`business_ad_accounts`** — `GET /businesses/{business_id}/ad_accounts`, records at
  `ad_accounts`, the REAL way to list ad accounts (the wave2 `ad_accounts` stream's flat path does
  not exist — see Known limits). Implemented with the `fan_out` dialect: `ids_from.request` lists
  every business id from `/businesses` (the same endpoint the `businesses` stream reads), `into.
  path_var` substitutes each id into `{{ fanout.id }}` in this stream's own `path`, and
  `stamp_field: "business_id"` stamps the current business id onto every emitted ad-account record.
  No pagination parameters are documented for the per-business ad-accounts endpoint either.
- **`assets`** — `GET /ad_accounts/{{ config.ad_account_id }}/assets`, records at `assets`,
  `offset_limit` pagination identical to the other `ad_account_id`-scoped streams. The real API's
  per-asset shape is polymorphic (`oneOf` image/video/audio variants per the OpenAPI spec), but
  `id`/`name`/`asset_type`/`status`/`created_at`/`updated_at` are common `requiredProperties`
  across every variant — this bundle's schema declares exactly that common subset, the same
  "reconstructed from the aggregate documentation footprint" approach `bitly`'s `docs.md` uses for
  its own polymorphic/partially-documented objects.

## Write actions & risks

One write action, added against the real, live OpenAPI spec. Legacy shipped none (`spotify_ads.go`'s
`Write` returned `connectors.ErrUnsupportedOperation` unconditionally), so there is no parity
baseline to diverge from.

- **`update_campaign`** (`PATCH /ad_accounts/{{ record.ad_account_id }}/campaigns/{{ record.id
  }}`) — mutates an existing campaign's `name`, `purchase_order` reference, or `status` (one of
  `UNSET`/`ACTIVE`/`PAUSED`/`ARCHIVED`/`AGENT_CONTROLLED`/`ACTIVE_RESTRICTED`/
  `PENDING_ADVERTISER_REVIEW`/`UNRECOGNIZED`, the real API's own enum). `minProperties: 3` on the
  record schema requires the two path-identifying fields (`ad_account_id`, `id`) plus at least one
  actual mutation field, matching the real API's own `minProperties: 1` requirement on the body
  (2 path fields + 1 body field = 3). Setting `status` to `PAUSED`/`ARCHIVED` stops that campaign's
  ad delivery and spend immediately; approval required.

Every other mutation endpoint in the real v3 API (campaign/ad-set/ad creation, ad-set/ad updates,
account/business administration) was evaluated and excluded — see `api_surface.json` for the
endpoint-by-endpoint reasoning. The unifying theme for the excluded create/update-beyond-campaign
endpoints: their real bodies are deeply nested campaign-graph objects (`targets`, `frequency_caps`,
`cost_modifiers`, `budget`, `pacing`, `assets`) with no flat-field shape the engine's default JSON
write body (record fields minus `path_fields`, copied verbatim) can construct without a Tier-2
`WriteHook`.

## Known limits

- **`page_size`/`max_pages` are not exposed as config properties.** Legacy accepts
  `config.page_size` (bounded 1-500, default 100) and `config.max_pages` (default unbounded) at
  read time. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are static values baked into
  `streams.json`'s pagination block, with no `{{ }}` templating support from `config.*` (matching
  the split-io/searxng precedent, F6 REVIEW.md). This bundle hard-codes `page_size: 100` (legacy's
  own default) and declares no `max_pages` (unbounded, matching legacy's own default). A caller
  that previously overrode either value per-run loses that capability; every default-config
  caller sees byte-identical behavior.
- **The wave2 `base_url` default and `ad_accounts` stream do not match the real, live API, and this
  bundle does not correct either.** This is a pre-existing wave2 discrepancy discovered during this
  Pass B research pass, not introduced by it — left untouched per the meta-rule against altering
  already-migrated behavior:
  - `base_url`'s default (`https://api-partner.spotify.com/ads/v2`) targets Spotify Ads API **v2**,
    which Spotify's own migration notice states was **sunset on 2025-08-05** — as of this pass's
    research date, that endpoint version no longer serves live traffic at all. The real, current
    API is v3 (`https://api-partner.spotify.com/ads/v3`), confirmed directly from the OpenAPI
    spec's `servers` block.
  - Independent of the version-sunset issue, the wave2 `ad_accounts` stream's flat `GET
    /ad_accounts` path **does not exist on the real API at any version** — every ad account
    genuinely belongs to a business, and the only real way to list them is `GET /businesses/
    {business_id}/ad_accounts` (per-business, requiring a business id first). This pass adds the
    CORRECT endpoint as a new, separately-named stream (`business_ad_accounts`, via `fan_out` over
    `businesses`) rather than silently repointing the existing `ad_accounts` stream, since doing so
    would change an already-migrated stream's resolved request shape outside this task's scope.
  - The three `config.ad_account_id`-scoped streams this pass leaves untouched (`campaigns`,
    `ad_sets`, `ads`) DO match the real v3 API's documented paths exactly (`/ad_accounts/
    {ad_account_id}/campaigns`, etc.) — only the version segment in `base_url` and the standalone
    `ad_accounts` stream are affected by this discrepancy.
  - The three NEW streams this pass adds (`businesses`, `business_ad_accounts`, `assets`) and the
    one new write action (`update_campaign`) were authored directly against the live v3 spec and
    inherit only the shared `base_url` version-segment issue (fixing it is out of this task's
    scope; a future capability-expansion or bug-fix pass should flip `base_url`'s default to
    `.../ads/v3` and confirm every existing stream/write still resolves correctly against it,
    since a version bump is a base-level, bundle-wide change, not a single-stream one).
- Full endpoint-by-endpoint accounting, including every excluded endpoint and the reasoning for
  each (billing, CAPI/dataset/pixel/mobile-app integrations, business/account administration,
  async reporting, targeting-lookup reference data), is in `api_surface.json`.
