# Overview

Reads affiliates, programs, coupons, referrals, and payments from the UpPromote API, and writes
affiliate/referral/coupon/payment/webhook-subscription lifecycle mutations.

Readable streams: `affiliates`, `programs`, `coupons`, `referrals`, `unpaid_payments`,
`paid_payments`.

Write actions: `create_affiliate`, `approve_deny_affiliate`, `set_upline_affiliate`,
`move_affiliate_to_program`, `connect_customer_to_affiliate`, `assign_coupon_to_affiliate`,
`create_referral`, `approve_deny_referral`, `add_referral_adjustment`,
`mark_as_paid_manual_payment`, `subscribe_webhook_event`, `update_webhook_subscription`,
`delete_webhook_subscription`.

Service API documentation: https://uppromote.com/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); UpPromote API key, sent as a Bearer token on every request.
- `base_url` (optional, string); default `https://api.uppromote.com`; format `uri`; UpPromote API
  base URL override for tests or proxies.
- `start_date` (optional, string); format `date-time`; Optional RFC3339 lower bound sent as the
  start_date query parameter, when set.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.uppromote.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `api/affiliates`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `affiliates`; page_number: `programs`, `coupons`, `referrals`,
`unpaid_payments`, `paid_payments`.

- `affiliates`: GET `api/affiliates` - records path `affiliates`; query `start_date` from template
  `{{ config.start_date }}`, omitted when absent.
- `programs`: GET `api/v2/programs` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 10.
- `coupons`: GET `api/v2/coupons` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 10.
- `referrals`: GET `api/v2/referrals` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 10.
- `unpaid_payments`: GET `api/v2/payments/unpaid` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 10.
- `paid_payments`: GET `api/v2/payments/paid` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 10.

## Write actions & risks

Overall write risk: external mutation of UpPromote affiliates, referrals, coupons, payments, and
webhook subscriptions; no destructive deletes are modeled.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_affiliate`: POST `api/v2/affiliates` - kind `create`; body type `json`; required record
  fields `email`; accepted fields `address`, `city`, `company`, `email`, `first_name`, `last_name`,
  `password`, `phone`, `program_id`, `send_email`, `state`, `status`; risk: creates a new affiliate
  account; low-risk, no approval required (UpPromote caps this at 150 affiliates/day per its own
  API).
- `approve_deny_affiliate`: POST `api/v2/affiliate/active` - kind `update`; body type `json`;
  required record fields `affiliate_email`, `status`; accepted fields `affiliate_email`, `status`;
  risk: approves or denies a pending affiliate application; low-risk, no approval required.
- `set_upline_affiliate`: POST `api/v2/affiliate/set-upline` - kind `update`; body type `json`;
  required record fields `affiliate_email`, `upline_affiliate_email`; accepted fields
  `affiliate_email`, `upline_affiliate_email`; risk: sets the referring (upline) affiliate for a
  downline affiliate, affecting multi-tier commission attribution; no approval required.
- `move_affiliate_to_program`: POST `api/v2/affiliate/move-affiliate-to-program` - kind `update`;
  body type `json`; required record fields `affiliate_email`, `program_id`; accepted fields
  `affiliate_email`, `program_id`; risk: reassigns an affiliate to a different commission program,
  changing future commission rules; no approval required.
- `connect_customer_to_affiliate`: POST `api/v2/affiliate/create-connect-customer` - kind `create`;
  body type `json`; required record fields `affiliate_email`, `customer_email`; accepted fields
  `affiliate_email`, `customer_email`; risk: links a Shopify customer email to an affiliate for
  future referral attribution; low-risk, no approval required.
- `assign_coupon_to_affiliate`: POST `api/v2/coupons/assign` - kind `create`; body type `json`;
  required record fields `affiliate_email`; accepted fields `affiliate_email`, `coupon`,
  `description`; risk: assigns a discount coupon code to an affiliate for referral tracking;
  low-risk, no approval required.
- `create_referral`: POST `api/v2/referrals` - kind `create`; body type `json`; required record
  fields `type`, `affiliate_email`; accepted fields `affiliate_email`, `comment`, `commission`,
  `is_replace`, `order_id`, `total_sale`, `type`; risk: creates a manual commission-bearing referral
  for an affiliate, either tied to a Shopify order or as a fixed amount; affects payout totals, no
  approval required.
- `approve_deny_referral`: POST `api/v2/referral/{{ record.id }}/status` - kind `update`; body type
  `json`; path fields `id`; body fields `status`; required record fields `id`, `status`; accepted
  fields `id`, `status`; risk: approves or denies a pending referral, affecting affiliate payout
  eligibility; no approval required.
- `add_referral_adjustment`: POST `api/v2/referral/{{ record.id }}/adjustment` - kind `update`; body
  type `json`; path fields `id`; body fields `adjustment`; required record fields `id`,
  `adjustment`; accepted fields `adjustment`, `id`; risk: adds a positive or negative commission
  adjustment to an existing referral, directly changing affiliate payout amounts; no approval
  required.
- `mark_as_paid_manual_payment`: POST `api/v2/payments/mark-as-paid` - kind `update`; body type
  `json`; required record fields `affiliate_email`; accepted fields `affiliate_email`, `message`,
  `note`, `referral_ids`; risk: marks approved referrals as manually paid outside UpPromote's own
  payout processing; affects financial records, no approval required.
- `subscribe_webhook_event`: POST `api/v2/webhook-subscriptions` - kind `create`; body type `json`;
  required record fields `target_url`, `event`; accepted fields `event`, `target_url`; risk:
  registers a new outbound webhook subscription that will deliver event payloads to an external URL;
  low-risk, no approval required.
- `update_webhook_subscription`: PUT `api/v2/webhook-subscriptions` - kind `update`; body type
  `json`; required record fields `target_url`, `event`; accepted fields `event`, `target_url`; risk:
  updates an existing webhook subscription's target URL; low-risk, no approval required.
- `delete_webhook_subscription`: DELETE `api/v2/webhook-subscriptions` - kind `delete`; body type
  `json`; body fields `event`; required record fields `event`; accepted fields `event`; missing
  records treated as success for status `404`; risk: removes a webhook subscription; the external
  endpoint stops receiving that event type, no approval required.

## Known limits

- API coverage includes 6 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=6, out_of_scope=3.
