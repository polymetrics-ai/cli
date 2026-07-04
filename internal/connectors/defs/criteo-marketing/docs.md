# Overview

Criteo Marketing is a wave2 fan-out migration. This bundle reads Criteo Marketing Solutions ad
sets, advertisers, campaigns, audiences, ad spend statistics, and (Pass B) Marketplace Performance
Outcomes (MPO) advertisers/sellers/budgets/seller-campaigns through the Criteo REST API, migrating
`internal/connectors/criteo-marketing` (the legacy hand-written connector, which stays registered
and unchanged until wave6's registry flip) at capability parity.

**Pass B full-surface expansion** (2026-07-04): reviewed the full documented Criteo Marketing
Solutions surface via its live OpenAPI spec (`https://api.criteo.com/2026-01/marketingsolutions/
open-api-specifications.json`, 91 operations at the latest published version). Added 4 new
`mpo_*` read streams (Marketplace Performance Outcomes: `mpo_advertisers`, `mpo_sellers`,
`mpo_budgets`, `mpo_seller_campaigns` — real, unfiltered, no-required-param GET list endpoints).
**No write actions were added**: every mutation-shaped endpoint in the real surface requires either
elevated write scopes (`MarketingSolutions_Campaign_Manage`/`Audience_Manage`, beyond this
connector's read-only grant) or a body shape the dialect cannot express (see Known limits). This
review also surfaced a **pre-existing defect in the original 5 streams**, described in full below
and flagged separately (not fixed in this pass, since fixing it needs either an engine change or a
scoped correctness fix, both out of this Pass B review's remit of adding new coverage +
documenting gaps).

## Auth setup

Provide `client_id`/`client_secret` secrets (Criteo Marketing Solutions API app credentials); the
bundle exchanges them for a bearer token via OAuth2 client-credentials
(`auth.mode: oauth2_client_credentials`) against `token_url`, matching legacy's
`connsdk.OAuth2ClientCredentials`. `token_url` defaults to `https://api.criteo.com/oauth2/token`
(legacy's own hardcoded literal-append default: `base + "/oauth2/token"`); `base_url` independently
defaults to `https://api.criteo.com`. See Known limits for the narrowed case where only one of the
two is overridden.

## Streams notes

`ad_sets`, `advertisers`, `campaigns`, and `audiences` share the same JSONAPI shape: `GET` against
a Criteo Marketing Solutions list/search endpoint, records at `data` (each item
`{id, type, attributes: {...}}`). `id`/`type` survive schema projection directly (they are
top-level keys on the raw record, matching the schema's property names exactly); every other field
is lifted out of the nested `attributes` object via a `computed_fields` rename (e.g.
`"name": "{{ record.attributes.name }}"`), matching legacy's `attributes(item)` flattening helper.
`audiences.nbActiveUsers` and `campaigns.spendLimit` use a bare single-reference
`computed_fields` template, so the engine's typed extraction preserves their native JSON types
(integer, object) rather than stringifying them.

`statistics` is shaped differently: it is a flat report row (no JSONAPI `attributes` wrapper), so
every field (`AdvertiserId`/`CampaignId`/`Day`/`Clicks`/`Displays`/`Spend`/`Currency`) survives
plain schema projection with no `computed_fields` at all, matching legacy's
`criteoStatisticsRecord`, which reads fields directly off the row. Its records live at `Rows`
(not `data`), matching Criteo's statistics report envelope. `statistics` also sends three optional
query filters — `startDate`/`endDate`/`currency` — each declared via the optional-query dialect
(`omit_when_absent: true`), so they are sent only when `start_date`/`end_date`/`currency` config
values are set, otherwise omitted entirely; this matches legacy's `statisticsFilters` helper, which
only adds a filter when the corresponding config key is non-empty.

Pagination is `offset_limit` (`limit_param: limit`, `offset_param: offset`) for every stream,
matching legacy's uniform offset/limit `harvest` loop (`base.Set("limit", ...)`, then
`query.Set("offset", ...)` per page) with a short-page stop (no explicit "has more" flag in any
Criteo list response) — exactly the `OffsetPaginator`'s built-in behavior. `streams.json`'s
`pagination.page_size: 100` matches legacy's own `criteoDefaultPageSize` (the value legacy's
`criteoPageSize` config helper falls back to whenever `page_size` is unset — legacy also allows a
per-call config override up to `criteoMaxPageSize = 1000`, which this bundle's static-literal
`base.pagination.page_size` cannot express; see `docs/migration/conventions.md`'s `page_size`/
`max_pages` non-runtime-overridable note). All 6 fixtures (`fixtures/streams/{ad_sets,advertisers,
campaigns,audiences,statistics}/page_1.json`) request `limit=100` accordingly and return their
entire small fixture record set on a single short page (`ad_sets`' fixture, previously split
across two `page_size: 2` pages, is now the single `page_1.json` file with all 3 records). None of
the 5 streams are incremental in legacy (no
persisted-cursor-driven request filtering anywhere in `criteo_marketing.go`): `statistics`'
`startDate`/`endDate` filters are plain config passthroughs, not a state-cursor-driven incremental
mechanism, so no `streams.json` entry declares an `incremental` block. `statistics`' schema still
declares `x-cursor-field: Day` (matching legacy's catalog-only `CursorFields: []string{"Day"}`) —
this is schema metadata only and does not by itself enable an `incremental_append` sync mode
absent an `incremental` block, matching legacy's real behavior of never filtering server-side by
date beyond the plain config passthrough.

`mpo_advertisers`, `mpo_sellers`, `mpo_budgets`, and `mpo_seller_campaigns` (Pass B) target
`/2026-01/marketing-solutions/marketplace-performance-outcomes/{advertisers,sellers,budgets,
seller-campaigns}` — **note the different API version path segment** (`2026-01`, not this bundle's
existing `2024-01`): Marketplace Performance Outcomes did not exist as a resource in `2024-01`
(confirmed by fetching that version's own OpenAPI spec — no `marketplace-performance-outcomes`
path at all), so these 4 new streams are pinned to the only version where the resource is
documented. Each responds with a bare top-level JSON array (`records: {"path": ""}`, the
acuity-scheduling-precedent shape for a root-array response) and documents no `page`/`limit`
pagination parameters at all, so each declares a stream-level `"pagination": {"type": "none"}`
override against the base's `offset_limit` spec. All 4 endpoints accept only optional query
filters (`advertiserId`/`campaignId`/`sellerId`/date-range/status, etc.) — none are sent here, so
every read returns the full unfiltered collection the credentials can access, matching the
"omit every filter" semantics `mpo_advertisers`' own docs describe ("If all parameters are omitted
the entire collection ... is returned").

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).
Every write-shaped endpoint in the real API surface requires either an elevated write scope
(`MarketingSolutions_Campaign_Manage`/`Audience_Manage`, beyond this connector's read-only
`MarketingSolutions_Campaign_Read`/`Audience_Read` grant) or a body shape the dialect cannot
express (a `oneOf`-discriminated body, or the same POST-with-filter-body ENGINE_GAP the read side
hits — see Known limits); no write action was added this pass. `capabilities.write` stays `false`.

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block), for the identical reason as this wave's `sendpulse`
  golden: `oauth2_client_credentials` auth requires a real `token_url` to POST a token request
  against, and conformance's synthetic non-secret config value is not a resolvable URL. Static
  checks (`spec_schema_valid`, `stream_schemas_valid`, `interpolations_resolve`, `docs_present`,
  `fixtures_present`, `secret_redaction`, etc.) are unaffected and still run. This bundle has no
  Tier-2 `AuthHook` (its auth is fully declarative `oauth2_client_credentials`), so there is no
  `paritytest/criteo-marketing` package for this wave; the read/pagination/schema shape is proven
  by structural review against legacy `internal/connectors/criteo-marketing` instead.
- **`max_pages` is left unbounded** (undeclared in `base.pagination`) rather than baked in at
  legacy's own config-overridable default (unbounded when `max_pages` is unset/`0`/`all`/
  `unlimited`) — this already matches legacy's own default behavior exactly (legacy's zero-value
  default IS unbounded), so no scope-narrowing is introduced here.
- `token_url`'s default is a fixed literal (`https://api.criteo.com/oauth2/token`), not a
  `base_url`-derived value. Legacy derives `token_url` from whatever `base_url` resolves to at
  runtime (`base + criteoTokenPath`), so a caller overriding `base_url` alone (leaving `token_url`
  unset) gets a token endpoint under the SAME custom host in legacy, but would still hit the fixed
  default host here. The engine's `spec.json` `"default"` materialization mechanism fills only a
  literal per property, with no cross-property derivation (conventions.md §3, the sentry/chargebee
  derived-default case) — this is a documented, accepted config-surface narrowing: a caller who
  overrides `base_url` for a test/proxy setup must also override `token_url` explicitly to point at
  the same host. Not exercised by any of this bundle's fixtures (both default to the real Criteo
  hosts).
- Full Criteo Marketing Solutions API surface beyond the 9 implemented streams (ad-set/campaign/
  budget mutation, Commerce-Grid audience-segments, catalog/product feeds, creative/coupon
  authoring, MPO per-id detail/stats sub-resources) is out of scope; see `api_surface.json`'s 91
  enumerated endpoints for the complete, real-category-and-reason disposition of every one (no more
  blanket "Pass B capability expansion" placeholders).
- **`ENGINE_GAP`: `ad_sets`, `campaigns`, and `audiences` issue `GET` against endpoints that only
  accept `POST` with a JSON filter body on the real API (discovered during this Pass B review, via
  the live OpenAPI spec).** `POST /marketing-solutions/{ad-sets,campaigns,audiences}/search` (an
  empty/omitted `filters` body returns every accessible record) are the ONLY documented ways to
  list these 3 resources in any spec version reviewed (`2024-01` through `2026-01`) — there is no
  GET alternative. The engine's declarative read path (`internal/connectors/engine/read.go`)
  hard-codes a `nil` request body on every stream request regardless of method;
  `streams.json`'s `StreamSpec.Body` field (`bundle.go`) is declared but never read anywhere in
  `read.go`, so a POST-with-body list read cannot be expressed in this dialect today. This is a
  **pre-existing wave2-migration defect**, not introduced by this Pass B review — the original
  bundle (and the legacy `internal/connectors/criteo-marketing` Go connector it was migrated from)
  both already called `GET` against these paths. It went undetected because `metadata.json`'s
  bundle-level `conformance.skip_dynamic: true` marker (see above) means no dynamic check has ever
  issued a REAL HTTP request against these paths to prove they work; the fixture-replay harness
  only proves the fixture's OWN declared request/response shape round-trips through the engine,
  never that the fixture matches upstream reality. **Flagged as a separate follow-up** (not fixed in
  this pass): the fix requires either (a) an engine mini-wave item wiring `stream.Body` through to
  the read path's request construction (a candidate `ENGINE_GAP` recurrence — worth checking
  whether other quarantined/partial connectors hit the identical "list endpoint is POST-with-body"
  shape before committing to this as a one-off engine change per conventions.md §6's `ENGINE_GAP`
  recurrence-threshold rule), or (b) documenting these 3 streams as blocked and removing them from
  `api_surface.json`'s `covered_by` pending that engine work. `api_surface.json` currently still
  lists them `covered_by` at their (broken) GET path to accurately describe what THIS bundle's
  `streams.json` actually calls today, alongside a companion `excluded` entry at the REAL POST path
  spelling out the gap.
- **`advertisers` targets a plain list path that does not exist in ANY reviewed spec version.**
  `GET /2024-01/marketing-solutions/advertisers` (and the same path under every later version
  checked through `2026-01`) is not a documented endpoint at all — Criteo's Marketing Solutions API
  has no bulk "list every advertiser" endpoint. The closest real single-advertiser-context
  endpoint is `GET /advertisers/me` (singular — returns the CURRENT credential's own advertiser
  identity, not a list of advertisers the credential can access). The closest real BULK list is the
  new `mpo_advertisers` stream (`GET .../marketplace-performance-outcomes/advertisers`), but it is
  a genuinely different, MPO-scoped resource with a different shape
  (`{advertiserName, currencyName, id, timeZoneId}` vs. the existing `advertisers` stream's
  JSONAPI-flattened `{name, country, currency, timezone}`) — swapping one for the other would be an
  accepted-input emitted-DATA change (a real deviation, not a drop-in replacement), so it was added
  as an ADDITIONAL stream (`mpo_advertisers`) rather than a silent substitution for the existing
  (already-broken) `advertisers` stream. Also flagged as part of the same follow-up as the
  `ad_sets`/`campaigns`/`audiences` `ENGINE_GAP` above.
- **`statistics` has the identical GET-instead-of-POST defect.** `POST /statistics/report` (a JSON
  body carrying report dimensions/metrics/date range) is the real, only documented method; this
  bundle's `statistics` stream sends `startDate`/`endDate`/`currency` as plain GET query parameters
  against the same path instead. Same root cause and same follow-up as above.
