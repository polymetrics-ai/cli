# Overview

Criteo Marketing is a wave2 fan-out migration. This bundle reads Criteo Marketing Solutions ad
sets, advertisers, campaigns, audiences, and ad spend statistics through the Criteo REST API,
migrating `internal/connectors/criteo-marketing` (the legacy hand-written connector, which stays
registered and unchanged until wave6's registry flip) at capability parity.

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

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

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
- Full Criteo Marketing Solutions API surface (ad-set/campaign/budget mutation, catalog/product
  feeds, audience creation, creative management) is out of scope for wave2; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Only the 5 legacy-parity read streams are implemented.
