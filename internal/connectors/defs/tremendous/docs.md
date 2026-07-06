# Overview

Reads and writes Tremendous campaigns, orders, rewards, funding sources, products, invoices, and
members through the Tremendous API.

Readable streams: `campaigns`, `orders`, `rewards`, `funding_sources`, `products`, `invoices`,
`members`.

Write actions: `create_order`, `approve_order`, `reject_order`, `cancel_reward`, `resend_reward`,
`generate_reward_link`, `create_invoice`, `delete_invoice`, `create_member`, `create_webhook`,
`delete_webhook`.

Service API documentation: https://developers.tremendous.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Tremendous API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://testflight.tremendous.com`; format `uri`;
  Tremendous API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://testflight.tremendous.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v2/orders`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100; maximum 1 page(s).

Pagination by stream: none: `products`, `members`; offset_limit: `invoices`; page_number:
`campaigns`, `orders`, `rewards`, `funding_sources`.

- `campaigns`: GET `/api/v2/campaigns` - records path `campaigns`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed
  output fields `created_at`.
- `orders`: GET `/api/v2/orders` - records path `orders`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed output
  fields `campaign_id`, `created_at`, `payment_status`.
- `rewards`: GET `/api/v2/rewards` - records path `rewards`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed output
  fields `created_at`, `order_id`.
- `funding_sources`: GET `/api/v2/funding_sources` - records path `funding_sources`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1
  page(s); computed output fields `created_at`.
- `products`: GET `/api/v2/products` - records path `products`.
- `invoices`: GET `/api/v2/invoices` - records path `invoices`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `members`: GET `/api/v2/members` - records path `members`.

## Write actions & risks

Overall write risk: external mutation with real financial impact: create_order spends funding-source
balance to issue rewards; approve_order/reject_order/cancel_reward/resend_reward act on
already-issued rewards; create_invoice/delete_invoice/create_member/create_webhook/delete_webhook
are organization-administration mutations.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_order`: POST `/api/v2/orders` - kind `create`; body type `json`; required record fields
  `payment`, `reward`; accepted fields `external_id`, `payment`, `reward`; risk: spends real
  funding-source balance to issue a gift card / prepaid card / donation reward to a recipient;
  external mutation with real financial impact, approval required.
- `approve_order`: POST `/api/v2/order_approvals/{{ record.id }}/approve` - kind `custom`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: approves an
  order pending admin review, releasing its rewards for delivery; real financial impact, approval
  required.
- `reject_order`: POST `/api/v2/order_approvals/{{ record.id }}/reject` - kind `custom`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: rejects an
  order pending admin review; the order's rewards are never delivered.
- `cancel_reward`: POST `/api/v2/rewards/{{ record.id }}/cancel` - kind `custom`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: cancels and refunds a
  reward; only valid for non-expired rewards with a delivery failure per Tremendous's own API
  contract.
- `resend_reward`: POST `/api/v2/rewards/{{ record.id }}/resend` - kind `custom`; body type `json`;
  path fields `id`; body fields `updated_email`, `updated_phone`; required record fields `id`;
  accepted fields `id`, `updated_email`, `updated_phone`; risk: resends a reward to its recipient
  (optionally at a new email/phone); only valid for rewards with a previous delivery failure.
- `generate_reward_link`: POST `/api/v2/rewards/{{ record.id }}/generate_link` - kind `custom`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: generates
  a new redemption link for an existing LINK-delivery reward; low-risk, does not move funds.
- `create_invoice`: POST `/api/v2/invoices` - kind `create`; body type `json`; required record
  fields `amount`; accepted fields `amount`, `currency_code`, `memo`, `po_number`; risk: creates an
  invoice that funds the organization's Tremendous balance once paid; low direct risk (a document,
  not a payment itself), no approval required.
- `delete_invoice`: DELETE `/api/v2/invoices/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: removes an invoice; per Tremendous's own docs this is a cosmetic
  operation with no further financial consequence (an already-paid invoice's funds are unaffected).
- `create_member`: POST `/api/v2/members` - kind `create`; body type `json`; required record fields
  `email`, `role`; accepted fields `email`, `role`; risk: invites a new user to manage the
  Tremendous organization (funding sources, campaigns, orders); grants organization access, approval
  required.
- `create_webhook`: POST `/api/v2/webhooks` - kind `create`; body type `json`; required record
  fields `url`; accepted fields `url`; risk: registers/replaces the organization's single webhook
  endpoint; a changed url redirects all future event deliveries to a different endpoint (Tremendous
  allows exactly one webhook per organization).
- `delete_webhook`: DELETE `/api/v2/webhooks/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently removes the organization's webhook subscription; event
  delivery stops immediately.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=1, duplicate_of=8, non_data_endpoint=2, out_of_scope=12,
  requires_elevated_scope=17.
