# Overview

ReferralHero is a wave2 fan-out declarative-HTTP migration. It reads ReferralHero lists,
subscribers, referrals, and rewards through the ReferralHero API v2 list endpoints
(`GET https://app.referralhero.com/api/v2/...`). This bundle targets capability parity with
`internal/connectors/referralhero` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a ReferralHero API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(key)`, and is never logged.
`base_url` defaults to `https://app.referralhero.com/api/v2` (legacy's `referralHeroDefaultBaseURL`)
and can be overridden for tests or proxies.

## Streams notes

All four streams (`lists`, `subscribers`, `referrals`, `rewards`) share the identical ReferralHero
envelope: `GET /api/v2/<resource>` returns `{"data":[...],"pagination":{"next_page":...}}`, records
live at `data`. Pagination is `page_number` (`page`/`per_page`, `start_page: 1`, static
`page_size: 100` matching legacy's `referralHeroDefaultPageSize`). Legacy's own `harvest` loop stops
when the response body's `pagination.next_page` is blank or does not advance past the current page
number, never on record count alone; the engine's `page_number` paginator instead stops on a short
page (`recordCount < page_size`). These two stop conditions are equivalent for every dataset except
one whose total record count is an exact multiple of 100, where legacy stops immediately via the
blank `next_page` check and the engine would issue one additional request returning an empty page
before stopping — no different records are ever emitted either way (the same documented
non-data-affecting divergence as this wave's aha/adobe-commerce-magento bundles).

Each stream's schema is a field-for-field projection of legacy's own `mapRecord` functions
(`listRecord`/`subscriberRecord`/`referralRecord`/`rewardRecord`): `lists` (`id`, `name`, `status`,
`created_at`), `subscribers` (`id`, `email`, `name`, `status`, `referral_code`, `updated_at`),
`referrals` (`id`, `subscriber_id`, `email`, `status`, `created_at`), `rewards` (`id`, `name`,
`status`, `updated_at`). `subscribers`/`rewards` publish `updated_at` as `x-cursor-field`,
`referrals` publishes `created_at`, matching legacy's own `CursorFields` declarations exactly —
but ReferralHero's list endpoints expose no server-side incremental filter parameter and legacy's
own `harvest` never applies one, so no `request_param`/`start_config_key`/`client_filtered` is
declared; every read is a full paginated sweep, matching legacy's true read behavior. `lists` has no
cursor field, also matching legacy (no `CursorFields` on that stream).

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (1-250, default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides (`pageSize`/`maxPages`). The engine's `page_number` paginator has no config-driven
  page-size or request-count-cap knob (mirrors this wave's aha/adobe-commerce-magento precedent);
  `page_size`/`max_pages` are therefore not declared in `spec.json`, and this bundle sends
  ReferralHero's own default (`per_page=100`) as a static pagination-block value.
- **`next_page`-based early stop is approximated by short-page stop only** — see Streams notes
  above; the only observable difference is one extra empty-page request when a stream's total
  record count is an exact multiple of 100, never a difference in which records are emitted.
- Full ReferralHero API surface (campaign/list management writes, webhooks, custom fields, exports)
  is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
