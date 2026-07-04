# Overview

ReferralHero is a declarative-HTTP API v2 connector. The bundle keeps the four legacy-parity
top-level streams (`lists`, `subscribers`, `referrals`, `rewards`) unchanged, then expands Pass B
coverage with documented campaign-scoped streams for leaderboards, bonuses, subscribers, referral
trees, rewards, coupon groups, and coupon inventories. It also exposes approved write actions for
the documented JSON/no-body mutation endpoints.

The legacy package under `internal/connectors/referralhero` remains registered until the wave 6
registry flip. Its emitted record fields are not changed by this bundle.

## Auth setup

Provide a ReferralHero API key via the `api_key` secret. It is sent as
`Authorization: Bearer <api_key>` and is never logged. `base_url` defaults to
`https://app.referralhero.com/api/v2` and can be overridden for tests or proxies.

## Streams notes

The original legacy-parity streams still read:

- `GET /lists` at `data`
- `GET /subscribers` at `data`
- `GET /referrals` at `data`
- `GET /rewards` at `data`

Those streams intentionally retain the old record projections: `lists` (`id`, `name`, `status`,
`created_at`), `subscribers` (`id`, `email`, `name`, `status`, `referral_code`, `updated_at`),
`referrals` (`id`, `subscriber_id`, `email`, `status`, `created_at`), and `rewards` (`id`, `name`,
`status`, `updated_at`). They also retain legacy pagination semantics: the first request sends
`page=1&per_page=100`, and subsequent requests follow `pagination.next_page` from the response
body rather than inferring continuation from page length.

Pass B streams follow ReferralHero's documented v2 envelopes:

- `list_leaderboard`: `GET /lists/{uuid}/leaderboard`, records at `data.ranking`.
- `list_bonuses`: `GET /lists/{uuid}/bonuses`, records at `data`.
- `subscribers_search_by_name`: `GET /subscribers/search_by_name?name=...`, records at
  `data.subscribers`.
- `campaign_subscribers`, `subscriber_detail`, `subscriber_by_email`, and `subscriber_by_mwr`:
  subscriber list/detail endpoints under `/lists/{uuid}/subscribers`.
- `subscriber_referrals`, `subscriber_level_2_all_referrals`,
  `subscriber_level_3_all_referrals`, `subscriber_level_1_referrals`,
  `subscriber_level_2_referrals`, and `subscriber_level_3_referrals`: documented referral-tree
  reads under a configured subscriber.
- `campaign_rewards` and `subscriber_rewards`: reward records at `data.rewards`.
- `coupon_groups` and `coupon_group_coupons`: coupon group and coupon inventory reads.

Campaign-scoped streams use config values for required path/query identifiers:
`list_uuid`, `subscriber_id`, `subscriber_email`, `subscriber_mwr`, `subscriber_name`, and
`coupon_group_id`. Fixture requests use the synthetic literal `synthetic-conformance-value` for
those templated config values, per migration convention.

## Write actions & risks

`capabilities.write` is `true`. The bundle includes declarative write actions for creating lists,
adding/updating/confirming/promoting/deleting subscribers, tracking conversion events, adding
points and transactions, updating reward status, creating coupon groups/coupons, and qualifying or
unqualifying referrals.

All write actions mutate live ReferralHero campaign state and require reverse-ETL plan approval.
`delete_subscriber` is marked destructive. Bulk transaction writes can affect up to 500 transaction
records in one API call and have higher blast radius than single-record actions.

## Known limits

- The legacy-parity streams remain as-is even though ReferralHero's public v2 docs now emphasize
  campaign-scoped envelopes such as `data.subscribers` and `data.rewards`. This preserves legacy
  emitted data until the registry cutover.
- Legacy accepted runtime `page_size`/`max_pages` config values for the four top-level streams.
  This bundle keeps the legacy default `per_page=100` and next-page-token behavior, but the engine
  pagination block is static and does not expose those runtime overrides.
- The dialect models the documented GET endpoints with explicit config identifiers. It does not
  perform nested discovery from every list to every subscriber/coupon group because the current
  declarative fan-out supports one parent source at a time.
- ReferralHero documents many request fields as "Path Parameters" even when the URL contains no
  slot for them. The write actions model those values as JSON request body fields, matching the
  documented JSON bulk-transaction shape and the connector engine's declarative write dialect.
