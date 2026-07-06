# Overview

Judge.me Reviews reads 23 stream(s), and writes through 13 action(s).

Readable streams: `reviews`, `products`, `widgets`, `product_review_widget`, `preview_badge_widget`,
`featured_carousel_widget`, `reviews_tab_widget`, `all_reviews_page_widget`,
`verified_badge_widget`, `all_reviews_count_widget`, `all_reviews_rating_widget`,
`shop_reviews_count_widget`, `shop_reviews_rating_widget`, `widget_settings`, `html_miracle_widget`,
`checkout_comments_widget`, `reviews_count`, `review`, `reviewer`, `webhooks`, `webhook`,
`shop_info`, `settings`.

Write actions: `create_review`, `update_review`, `update_reviewer`, `request_reviewer_data`,
`delete_webhook`, `create_webhook`, `update_webhook`, `bulk_create_webhooks`, `update_shop`,
`uninstall_shop`, `create_checkout_comment`, `create_reply`, `create_private_reply`.

Service API documentation: https://judge.me/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Judge.me private, public, or OAuth API token. Sent as the
  X-Api-Token header.
- `base_url` (optional, string); default `https://api.judge.me/api/v1`; format `uri`; Judge.me API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`.
- `product_external_id` (optional, string); Optional external platform product id for widget
  lookups.
- `product_handle` (optional, string); Optional product handle for widget lookups.
- `product_id` (optional, string); Optional Judge.me product id for review/widget filters.
- `rating` (optional, string); Optional review rating filter (1-5).
- `review_id` (optional, string); Judge.me review id for the review detail stream.
- `reviewer_email` (optional, string); Optional reviewer email for reviewer detail lookups.
- `reviewer_external_id` (optional, string); Optional external platform reviewer id for reviewer
  detail lookups.
- `reviewer_id` (optional, string); Judge.me reviewer id for reviewer detail and filters.
- `setting_keys` (optional, string); Optional setting key value for settings stream; sent as
  setting_keys[].
- `shop_domain` (required, string); The Shopify shop domain (for example example.myshopify.com).
  Sent as the shop_domain query parameter on every request.
- `webhook_id` (optional, string); Judge.me webhook id for the webhook detail stream.
- `widget_page` (optional, string); Optional page value for widget HTML endpoints that accept page.
- `widget_per_page` (optional, string); Optional per_page value for widget HTML endpoints that
  accept per_page.
- `widget_review_type` (optional, string); Optional review_type for reviews_tab and all_reviews_page
  widgets.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.judge.me/api/v1`, `page_size=100`.

Authentication behavior:

- API key authentication in query parameter `shop_domain` using `config.shop_domain`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/reviews`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
1; page size 100; maximum 1000 page(s).

Pagination by stream: none: `product_review_widget`, `preview_badge_widget`,
`featured_carousel_widget`, `reviews_tab_widget`, `all_reviews_page_widget`,
`verified_badge_widget`, `all_reviews_count_widget`, `all_reviews_rating_widget`,
`shop_reviews_count_widget`, `shop_reviews_rating_widget`, `widget_settings`, `html_miracle_widget`,
`checkout_comments_widget`, `reviews_count`, `review`, `reviewer`, `webhooks`, `webhook`,
`shop_info`, `settings`; page_number: `reviews`, `products`, `widgets`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `reviews`: GET `/reviews` - records path `reviews`; query `per_page` from template `{{
  config.page_size }}`, default `100`; `product_id` from template `{{ config.product_id }}`, omitted
  when absent; `rating` from template `{{ config.rating }}`, omitted when absent; `reviewer_id` from
  template `{{ config.reviewer_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 100; maximum 1000 page(s); incremental
  cursor `created_at`; formatted as `rfc3339`; computed output fields `reviewer_email`,
  `reviewer_id`, `reviewer_name`.
- `products`: GET `/products` - records path `products`; query `per_page` from template `{{
  config.page_size }}`, default `100`; page-number pagination; page parameter `page`; no page-size
  parameter; starts at 1; page size 100; maximum 1000 page(s); incremental cursor `created_at`;
  formatted as `rfc3339`.
- `widgets`: GET `/widgets` - records path `widgets`; query `per_page` from template `{{
  config.page_size }}`, default `100`; page-number pagination; page parameter `page`; no page-size
  parameter; starts at 1; page size 100; maximum 1000 page(s); incremental cursor `created_at`;
  formatted as `rfc3339`.
- `product_review_widget`: GET `/widgets/product_review` - records path `.`; query `external_id`
  from template `{{ config.product_external_id }}`, omitted when absent; `handle` from template `{{
  config.product_handle }}`, omitted when absent; `id` from template `{{ config.product_id }}`,
  omitted when absent; `page` from template `{{ config.widget_page }}`, omitted when absent;
  `per_page` from template `{{ config.widget_per_page }}`, omitted when absent; computed output
  fields `id`; emits passthrough records.
- `preview_badge_widget`: GET `/widgets/preview_badge` - records path `.`; query `external_id` from
  template `{{ config.product_external_id }}`, omitted when absent; `handle` from template `{{
  config.product_handle }}`, omitted when absent; `id` from template `{{ config.product_id }}`,
  omitted when absent; computed output fields `id`; emits passthrough records.
- `featured_carousel_widget`: GET `/widgets/featured_carousel` - records path `.`; computed output
  fields `id`; emits passthrough records.
- `reviews_tab_widget`: GET `/widgets/reviews_tab` - records path `.`; query `page` from template
  `{{ config.widget_page }}`, omitted when absent; `per_page` from template `{{
  config.widget_per_page }}`, omitted when absent; `review_type` from template `{{
  config.widget_review_type }}`, omitted when absent; computed output fields `id`; emits passthrough
  records.
- `all_reviews_page_widget`: GET `/widgets/all_reviews_page` - records path `.`; query `page` from
  template `{{ config.widget_page }}`, omitted when absent; `review_type` from template `{{
  config.widget_review_type }}`, omitted when absent; computed output fields `id`; emits passthrough
  records.
- `verified_badge_widget`: GET `/widgets/verified_badge` - records path `.`; computed output fields
  `id`; emits passthrough records.
- `all_reviews_count_widget`: GET `/widgets/all_reviews_count` - records path `.`; computed output
  fields `id`; emits passthrough records.
- `all_reviews_rating_widget`: GET `/widgets/all_reviews_rating` - records path `.`; computed output
  fields `id`; emits passthrough records.
- `shop_reviews_count_widget`: GET `/widgets/shop_reviews_count` - records path `.`; computed output
  fields `id`; emits passthrough records.
- `shop_reviews_rating_widget`: GET `/widgets/shop_reviews_rating` - records path `.`; computed
  output fields `id`; emits passthrough records.
- `widget_settings`: GET `/widgets/settings` - records path `.`; computed output fields `id`; emits
  passthrough records.
- `html_miracle_widget`: GET `/widgets/html_miracle` - records path `.`; computed output fields
  `id`; emits passthrough records.
- `checkout_comments_widget`: GET `/widgets/checkout_comments_widget` - records path `.`; query
  `external_id` from template `{{ config.product_external_id }}`, omitted when absent; `handle` from
  template `{{ config.product_handle }}`, omitted when absent; `id` from template `{{
  config.product_id }}`, omitted when absent; computed output fields `id`; emits passthrough
  records.
- `reviews_count`: GET `/reviews/count` - records path `.`; query `product_id` from template `{{
  config.product_id }}`, omitted when absent; `rating` from template `{{ config.rating }}`, omitted
  when absent; `reviewer_id` from template `{{ config.reviewer_id }}`, omitted when absent; computed
  output fields `id`; emits passthrough records.
- `review`: GET `/reviews/{{ config.review_id }}` - records path `review`; computed output fields
  `reviewer_email`, `reviewer_id`, `reviewer_name`; emits passthrough records.
- `reviewer`: GET `/reviewers/{{ config.reviewer_id }}` - records path `reviewer`; query `email`
  from template `{{ config.reviewer_email }}`, omitted when absent; `external_id` from template `{{
  config.reviewer_external_id }}`, omitted when absent; emits passthrough records.
- `webhooks`: GET `/webhooks` - records path `webhooks`; emits passthrough records.
- `webhook`: GET `/webhooks/{{ config.webhook_id }}` - records path `webhook`; emits passthrough
  records.
- `shop_info`: GET `/shops/info` - records path `shop`; emits passthrough records.
- `settings`: GET `/settings` - records path `settings`; query `setting_keys[]` from template `{{
  config.setting_keys }}`, omitted when absent; computed output fields `id`; emits passthrough
  records.

## Write actions & risks

Overall write risk: external Judge.me API mutations can create reviews, update
moderation/reviewer/shop/webhook state, create replies/comments, and uninstall a shop.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_review`: POST `/reviews` - kind `create`; body type `json`; required record fields
  `shop_domain`, `platform`, `name`, `email`, `rating`, `body`; accepted fields `body`,
  `cf_answers`, `email`, `id`, `ip_addr`, `name`, `picture_urls`, `platform`, `rating`,
  `reviewer_name_format`, `shop_domain`, `title`; risk: creates a public web review in Judge.me;
  approval required.
- `update_review`: PUT `/reviews/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `curated`; accepted fields `curated`, `id`; risk: publishes or
  hides a Judge.me review by changing curated status; approval required.
- `update_reviewer`: PUT `/reviewers/{{ record.id }}` - kind `upsert`; body type `json`; path fields
  `id`; required record fields `id`, `reviewer`; accepted fields `id`, `reviewer`; risk: creates or
  updates reviewer identity fields in Judge.me; approval required.
- `request_reviewer_data`: POST `/reviewers/data_request` - kind `custom`; body type `json`;
  required record fields `customer`; accepted fields `customer`, `orders_requested`; risk: submits a
  Judge.me reviewer data request; approval required.
- `delete_webhook`: DELETE `/webhooks` - kind `delete`; body type `json`; body fields `key`, `url`;
  required record fields `key`, `url`; accepted fields `key`, `url`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes a Judge.me webhook
  subscription; approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `webhook`; accepted fields `webhook`; risk: creates a Judge.me webhook subscription; approval
  required.
- `update_webhook`: PUT `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `webhook`; accepted fields `id`, `webhook`; risk: updates a
  Judge.me webhook subscription; approval required.
- `bulk_create_webhooks`: POST `/webhooks/bulk_create` - kind `create`; body type `json`; required
  record fields `webhooks`; accepted fields `webhooks`; risk: creates multiple Judge.me webhook
  subscriptions; approval required.
- `update_shop`: PUT `/shops` - kind `update`; body type `json`; accepted fields `country`,
  `custom_domain`, `domain`, `email`, `name`, `owner`, `phone`, `plan`, `timezone`; risk: updates
  Judge.me shop profile fields; approval required.
- `uninstall_shop`: DELETE `/shops` - kind `delete`; body type `none`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: uninstalls the shop from Judge.me;
  destructive approval required.
- `create_checkout_comment`: POST `/shops` - kind `create`; body type `json`; required record fields
  `content`, `external_product_id`, `create_from`, `customer`; accepted fields `content`,
  `create_from`, `customer`, `external_product_id`; risk: creates a checkout comment in Judge.me
  Checkout Comments; approval required.
- `create_reply`: POST `/replies` - kind `create`; body type `json`; required record fields
  `review_id`, `reply`; accepted fields `reply`, `review_id`, `send_reply_email`; risk: creates a
  public reply on a Judge.me review; approval required.
- `create_private_reply`: POST `/private_replies` - kind `create`; body type `json`; required record
  fields `review_id`, `private_reply`; accepted fields `private_reply`, `review_id`,
  `send_private_email`; risk: creates a private email reply for a Judge.me review; approval
  required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=100.
- API coverage includes 23 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
