# Overview

Whisky Hunter is a Pass B full-surface declarative-HTTP connector. It reads the complete public
Whisky Hunter API (`GET https://whiskyhunter.net/api/...`) — all 5 documented GET endpoints, both
aggregate list resources and both per-slug detail resources. This bundle originated as a wave2
fan-out migration parity-tested against `internal/connectors/whisky-hunter` (the hand-written
connector it migrates, which implements only 2 of these 5 endpoints); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

No credentials are required: Whisky Hunter's API is fully public and unauthenticated. `base_url`
defaults to `https://whiskyhunter.net` and may be overridden for tests/proxies, matching legacy's
`defaultBaseURL`/`baseURL` validation (scheme+host required, trailing slash trimmed). The live
OpenAPI 2.0 document (`https://whiskyhunter.net/api/?format=openapi`) declares a `Basic`
`securityDefinition`, but this is a drf-yasg (Django REST Framework) template artifact describing
the underlying Django admin site's own login, not a real credentialed path for API callers — every
documented endpoint is verified reachable with no `Authorization` header at all.

## Streams notes

5 streams cover the API's full documented surface (`api_surface.json`):

- `auctions_data` (`GET /api/auctions_data/`) — monthly aggregated stats across ALL online
  auctions. Primary key `[auction_slug, dt]` (verified duplicate-free against the live response: no
  `(auction_slug, dt)` pair repeats).
- `auctions_info` (`GET /api/auctions_info`) — one row per known online auction house (fees,
  currency, URL). Primary key `[slug]`.
- `distilleries_info` (`GET /api/distilleries_info/`) — one row per known distillery (name,
  country). Primary key `[slug]`.
- `auction_data` (`GET /api/auction_data/{slug}/`) — the same monthly-aggregate shape as
  `auctions_data`, scoped to one auction house. Modeled with `fan_out`: the id list comes from a
  preliminary, fully-paginated `GET /api/auctions_info` request (`ids_from.request`, `id_field:
  slug`), and each resolved slug is substituted into the stream's own `path` via
  `into.path_var: id` / `{{ fanout.id }}`. This reaches every auction house's per-auction detail
  without any caller-supplied slug list, unlike a `config_key`-driven fan_out.
- `distillery_data` (`GET /api/distillery_data/{slug}/`) — the same per-month aggregate shape as
  `distilleries_info`'s companion resource, scoped to one distillery; fan_out sources its slugs from
  `GET /api/distilleries_info/` the identical way.

None of the 5 streams declares pagination or an `incremental` block: every endpoint (aggregate,
info, and per-slug detail alike) returns its full result set in one unpaginated JSON array response
with no cursor/page envelope of any kind — confirmed against every endpoint's live response, not
only the 2 legacy covered. All 5 declare `"projection": "passthrough"`, since every response element
is emitted verbatim with no field-building.

**The legacy connector's field-shape assumption for `auctions_data`/`distilleries_info` (an integer
`id` field, and a `winning_bid` field on auctions) does not exist anywhere in the real, live wire
shape** — verified by direct request against the production API. The real `auctions_data`/
`auction_data` shape is `{dt, auction_slug, auction_name, winning_bid_mean,
auction_trading_volume, auction_lots_count, all_auctions_lots_count}`; the real `distilleries_info`
shape is `{slug, name, country}` (no `id` at all); the real `distillery_data` shape is `{dt, slug,
name, winning_bid_max, winning_bid_min, winning_bid_mean, trading_volume, lots_count}`. This
bundle's schemas/fixtures were rewritten from scratch against the real recorded shape (§4's
recorded-real-shape rule) rather than preserving the legacy connector's incorrect field
assumptions — legacy itself was never a faithful implementation of this API to begin with (see
Known limits).

## Write actions & risks

None. The live OpenAPI document publishes zero non-GET operations anywhere in its `paths` map;
Whisky Hunter's public API has no write capability of any kind. `capabilities.write` is `false` and
this bundle ships no `writes.json`.

## Known limits

- **Legacy's fixture-mode-only `stream` marker field is not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`, a credential-free conformance-harness affordance)
  stamps an extra `stream` field onto every fixture-mode record (`whisky_hunter.go:134`). This is
  not part of the live record shape; this bundle's schemas and fixtures target the live path only.
  The engine's own conformance/fixture-replay harness provides the credential-free test affordance
  this bundle needs, so no fixture-mode equivalent is needed here.
- No pagination or incremental sync is modeled for any of the 5 streams, matching every endpoint's
  real behavior — Whisky Hunter's public API returns each resource as a single flat array with no
  cursor field and no page envelope, confirmed live for all 5 endpoints (not only the 2 legacy
  covered).
- **Legacy's own field mapping for `auctions`/`distilleries` never matched the real API.** Legacy
  declared an `id` (integer) field on both streams and a `winning_bid` field on auctions; the live
  API has never returned an `id` field on either resource, and the real auction-aggregate field is
  `winning_bid_mean`, not `winning_bid`. This bundle's `auctions_data`/`distilleries_info` schemas
  (the direct successors of legacy's `auctions`/`distilleries` streams) reflect the real wire shape;
  this is a bundle-vs-legacy field-shape correction, not a parity deviation, since matching legacy's
  incorrect assumption would mean the bundle also never emitted real API fields correctly.
- `auction_data`/`distillery_data` fan out over every slug returned by `auctions_info`/
  `distilleries_info` respectively (31 auctions, 159 distilleries at time of writing) — a full sync
  of either stream issues one HTTP request per known slug, with no way to narrow the fan-out to a
  caller-chosen subset (the `ids_from.request` fan_out shape has no config-driven override; see
  `docs/migration/conventions.md`'s fan_out dialect reference).
