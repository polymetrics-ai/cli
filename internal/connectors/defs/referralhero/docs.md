# Overview

Reads ReferralHero lists, subscribers, referrals, rewards, coupon groups, and campaign-scoped
subscriber resources, and performs approved ReferralHero API v2 mutations.

Readable streams: `lists`, `subscribers`, `referrals`, `rewards`, `list_leaderboard`,
`list_bonuses`, `subscribers_search_by_name`, `campaign_subscribers`, `subscriber_detail`,
`subscriber_by_email`, `subscriber_by_mwr`, `subscriber_referrals`,
`subscriber_level_2_all_referrals`, `subscriber_level_3_all_referrals`,
`subscriber_level_1_referrals`, `subscriber_level_2_referrals`, `subscriber_level_3_referrals`,
`campaign_rewards`, `subscriber_rewards`, `coupon_groups`, `coupon_group_coupons`.

Write actions: `create_list`, `add_subscriber`, `track_referral_conversion_event`,
`confirm_subscriber_by_id`, `confirm_subscriber_by_identifier`, `update_subscriber`, `add_points`,
`add_transaction`, `add_bulk_transactions`, `promote_subscriber`, `unlock_promoted_reward`,
`delete_subscriber`, `update_reward_status`, `create_coupon_group`, `create_coupons`,
`unqualify_referral`, `qualify_referral`.

Service API documentation: https://support.referralhero.com/integrate/rest-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); ReferralHero API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://app.referralhero.com/api/v2`; format `uri`;
  ReferralHero API base URL override for tests or proxies.
- `coupon_group_id` (optional, string); Coupon group ID used by the coupon_group_coupons stream.
- `list_uuid` (optional, string).
- `subscriber_email` (optional, string); Subscriber email used by the subscriber_by_email stream.
- `subscriber_id` (optional, string); Subscriber ID used by subscriber detail/referral/reward
  streams.
- `subscriber_mwr` (optional, string); Referral code/MWR used by the subscriber_by_mwr stream.
- `subscriber_name` (optional, string); Name search term used by subscribers_search_by_name.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.referralhero.com/api/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/lists` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from
`pagination.next_page`.

Pagination by stream: cursor: `lists`, `subscribers`, `referrals`, `rewards`; none:
`list_leaderboard`, `list_bonuses`, `subscriber_detail`, `subscriber_by_email`, `subscriber_by_mwr`,
`coupon_group_coupons`; page_number: `subscribers_search_by_name`, `campaign_subscribers`,
`subscriber_referrals`, `subscriber_level_2_all_referrals`, `subscriber_level_3_all_referrals`,
`subscriber_level_1_referrals`, `subscriber_level_2_referrals`, `subscriber_level_3_referrals`,
`campaign_rewards`, `subscriber_rewards`, `coupon_groups`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `lists`: GET `/lists` - records path `data`; query `page`=`1`; `per_page`=`100`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next_page`.
- `subscribers`: GET `/subscribers` - records path `data`; query `page`=`1`; `per_page`=`100`;
  cursor pagination; cursor parameter `page`; next token from `pagination.next_page`; incremental
  cursor `updated_at`; formatted as `rfc3339`.
- `referrals`: GET `/referrals` - records path `data`; query `page`=`1`; `per_page`=`100`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next_page`; incremental cursor
  `created_at`; formatted as `rfc3339`.
- `rewards`: GET `/rewards` - records path `data`; query `page`=`1`; `per_page`=`100`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next_page`; incremental cursor
  `updated_at`; formatted as `rfc3339`.
- `list_leaderboard`: GET `/lists/{{ config.list_uuid }}/leaderboard` - records path `data.ranking`.
- `list_bonuses`: GET `/lists/{{ config.list_uuid }}/bonuses` - records path `data`.
- `subscribers_search_by_name`: GET `/subscribers/search_by_name` - records path `data.subscribers`;
  query `name`=`{{ config.subscriber_name }}`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 50.
- `campaign_subscribers`: GET `/lists/{{ config.list_uuid }}/subscribers` - records path
  `data.subscribers`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 50.
- `subscriber_detail`: GET `/lists/{{ config.list_uuid }}/subscribers/{{ config.subscriber_id }}` -
  single-object response; records path `data`.
- `subscriber_by_email`: GET `/lists/{{ config.list_uuid }}/subscribers/retrieve_by_email` -
  single-object response; records path `data`; query `email`=`{{ config.subscriber_email }}`.
- `subscriber_by_mwr`: GET `/lists/{{ config.list_uuid }}/subscribers/retrieve_by_mwr` -
  single-object response; records path `data`; query `mwr`=`{{ config.subscriber_mwr }}`.
- `subscriber_referrals`: GET `/lists/{{ config.list_uuid }}/subscribers/{{ config.subscriber_id
  }}/referred` - records path `data.subscribers`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 50.
- `subscriber_level_2_all_referrals`: GET `/lists/{{ config.list_uuid }}/subscribers/{{
  config.subscriber_id }}/level_2_all_referrals` - records path `data.subscribers`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `subscriber_level_3_all_referrals`: GET `/lists/{{ config.list_uuid }}/subscribers/{{
  config.subscriber_id }}/level_3_all_referrals` - records path `data.subscribers`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `subscriber_level_1_referrals`: GET `/lists/{{ config.list_uuid }}/subscribers/{{
  config.subscriber_id }}/level_1_referrals` - records path `data.subscribers`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `subscriber_level_2_referrals`: GET `/lists/{{ config.list_uuid }}/subscribers/{{
  config.subscriber_id }}/level_2_referrals` - records path `data.subscribers`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `subscriber_level_3_referrals`: GET `/lists/{{ config.list_uuid }}/subscribers/{{
  config.subscriber_id }}/level_3_referrals` - records path `data.subscribers`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `campaign_rewards`: GET `/lists/{{ config.list_uuid }}/rewards` - records path `data.rewards`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  50.
- `subscriber_rewards`: GET `/lists/{{ config.list_uuid }}/subscribers/{{ config.subscriber_id
  }}/rewards` - records path `data.rewards`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 50.
- `coupon_groups`: GET `/lists/{{ config.list_uuid }}/coupon_groups` - records path
  `data.coupon_groups`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 50.
- `coupon_group_coupons`: GET `/lists/{{ config.list_uuid }}/coupon_groups/{{ config.coupon_group_id
  }}` - records path `data.coupons`.

## Write actions & risks

Overall write risk: creates and mutates live ReferralHero campaign, subscriber, transaction, reward,
coupon, and qualification state; approval required before execution.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_list`: POST `/lists` - kind `create`; body type `json`; required record fields `website`,
  `name`; accepted fields `name`, `website`; risk: creates a live ReferralHero campaign/list in the
  account; external mutation, approval required.
- `add_subscriber`: POST `/lists/{{ record.uuid }}/subscribers` - kind `create`; body type `json`;
  path fields `uuid`; required record fields `uuid`, `email`; accepted fields `advocate_name`,
  `conversion_category`, `conversion_value`, `crypto_wallet_address`, `device`, `domain`,
  `double_optin`, `email`, `extra_field`, `extra_field_2`, `name`, `other_identifier_value`,
  `phone_number`, `points`, `referrer`, `source`, `status`, `stripe_customer_id`, and 3 more; risk:
  creates or registers a live subscriber in a ReferralHero campaign and may trigger campaign
  email/referral workflows; approval required.
- `track_referral_conversion_event`: POST `/lists/{{ record.uuid
  }}/subscribers/track_referral_conversion_event` - kind `update`; body type `json`; path fields
  `uuid`; required record fields `uuid`, `email`; accepted fields `conversion_value`,
  `crypto_wallet_address`, `email`, `other_identifier_value`, `phone_number`, `product_id`,
  `referrer`, `stripe_customer_id`, `tags`, `transaction_id`, `uuid`; risk: confirms/unconfirms
  referral conversion state and may create a referral when a referrer is provided; external
  mutation, approval required.
- `confirm_subscriber_by_id`: POST `/lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id
  }}/confirm` - kind `update`; body type `none`; path fields `uuid`, `subscriber_id`; required
  record fields `uuid`, `subscriber_id`; accepted fields `subscriber_id`, `uuid`; risk: confirms a
  verified referral/subscriber conversion in the campaign; external mutation, approval required.
- `confirm_subscriber_by_identifier`: POST `/lists/{{ record.uuid }}/subscribers/confirm` - kind
  `update`; body type `json`; path fields `uuid`; required record fields `uuid`, `email`; accepted
  fields `crypto_wallet_address`, `email`, `other_identifier_value`, `phone_number`, `uuid`; risk:
  confirms a verified referral/subscriber conversion by unique identifier; external mutation,
  approval required.
- `update_subscriber`: POST `/lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}` - kind
  `update`; body type `json`; path fields `uuid`, `subscriber_id`; required record fields `uuid`,
  `subscriber_id`; accepted fields `address`, `city`, `country`, `crypto_wallet_address`, `email`,
  `extra_field`, `extra_field_2`, `name`, `other_identifier_value`, `phone_number`, `points`,
  `stripe_customer_id`, `subscriber_id`, `tags`, `uuid`; risk: updates profile, identifier, points,
  address, or tag fields for a verified subscriber; external mutation, approval required.
- `add_points`: POST `/lists/{{ record.uuid }}/subscribers/add_points` - kind `update`; body type
  `json`; path fields `uuid`; required record fields `uuid`, `email`, `points`; accepted fields
  `crypto_wallet_address`, `email`, `other_identifier_value`, `phone_number`, `points`, `uuid`;
  risk: adds points to a subscriber, changing contest/reward standings; external mutation, approval
  required.
- `add_transaction`: POST `/lists/{{ record.uuid }}/subscribers/add_transactions` - kind `create`;
  body type `json`; path fields `uuid`; required record fields `uuid`, `email`, `amount`; accepted
  fields `amount`, `crypto_wallet_address`, `email`, `lifetime_spend`, `other_identifier_value`,
  `phone_number`, `product_id`, `reward_value`, `transaction_id`, `uuid`; risk: records a
  transaction against a subscriber and may affect conversion/reward calculations; external mutation,
  approval required.
- `add_bulk_transactions`: POST `/lists/{{ record.uuid }}/subscribers/add_bulk_transactions` - kind
  `create`; body type `json`; path fields `uuid`; required record fields `uuid`, `transactions`;
  accepted fields `transactions`, `uuid`; risk: records up to 500 transactions in one call and
  emails an admin CSV result; high-blast-radius external mutation, approval required.
- `promote_subscriber`: POST `/lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id
  }}/promote` - kind `update`; body type `none`; path fields `uuid`, `subscriber_id`; required
  record fields `uuid`, `subscriber_id`; accepted fields `subscriber_id`, `uuid`; risk: promotes a
  subscriber into the campaign winners/promoted state; external mutation, approval required.
- `unlock_promoted_reward`: POST `/lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id
  }}/unlock_promoted_reward` - kind `update`; body type `none`; path fields `uuid`, `subscriber_id`;
  required record fields `uuid`, `subscriber_id`; accepted fields `subscriber_id`, `uuid`; risk:
  unlocks a promoted reward for a subscriber, changing reward fulfillment state; external mutation,
  approval required.
- `delete_subscriber`: DELETE `/lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}` -
  kind `delete`; body type `none`; path fields `uuid`, `subscriber_id`; required record fields
  `uuid`, `subscriber_id`; accepted fields `subscriber_id`, `uuid`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: permanently deletes a subscriber from
  a live campaign; destructive external mutation, approval required.
- `update_reward_status`: POST `/lists/{{ record.uuid }}/subscribers/update_reward_status` - kind
  `update`; body type `json`; path fields `uuid`; required record fields `uuid`, `reward_id`,
  `status`; accepted fields `reward_id`, `status`, `uuid`; risk: changes fulfillment status for an
  unlocked reward; external mutation, approval required.
- `create_coupon_group`: POST `/lists/{{ record.uuid }}/coupon_groups` - kind `create`; body type
  `json`; path fields `uuid`; required record fields `uuid`, `name`, `coupons`, `active`; accepted
  fields `active`, `coupons`, `name`, `uuid`; risk: creates a campaign coupon group and coupon
  inventory; external mutation, approval required.
- `create_coupons`: POST `/lists/{{ record.uuid }}/coupons` - kind `create`; body type `json`; path
  fields `uuid`; required record fields `uuid`, `coupon_group_id`, `coupons`; accepted fields
  `coupon_group_id`, `coupons`, `uuid`; risk: adds redeemable coupon codes to an existing campaign
  coupon group; external mutation, approval required.
- `unqualify_referral`: POST `/lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id
  }}/unqualify` - kind `update`; body type `none`; path fields `uuid`, `subscriber_id`; required
  record fields `uuid`, `subscriber_id`; accepted fields `subscriber_id`, `uuid`; risk: marks a
  referral/subscriber as unqualified, changing campaign qualification and reward state; external
  mutation, approval required.
- `qualify_referral`: POST `/lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/qualify`
  - kind `update`; body type `none`; path fields `uuid`, `subscriber_id`; required record fields
  `uuid`, `subscriber_id`; accepted fields `subscriber_id`, `uuid`; risk: marks a
  referral/subscriber as qualified, changing campaign qualification and reward state; external
  mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 21 stream-backed endpoint group(s), 17 write-backed endpoint group(s).
