# Overview

Reads Appfigures app-store reviews, products, analytics reports
(sales/ratings/revenue/subscriptions/ads/estimates), reference data
(categories/countries/languages/currencies/stores/SDKs), release events, connected external
accounts, account users, and account info through the Appfigures v2 REST API, and manages release
events and review responses.

Readable streams: `reviews`, `products`, `sales`, `ratings`, `categories`, `revenue`,
`subscriptions`, `ads`, `estimates`, `events`, `external_accounts`, `users`, `account_info`,
`data_countries`, `data_languages`, `data_stores`, `data_currencies`, `data_sdks`.

Write actions: `reply_to_review`, `create_event`, `update_event`, `delete_event`.

Service API documentation: https://docs.appfigures.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Appfigures Personal Access Token, sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.appfigures.com/v2`; format `uri`; Appfigures
  API base URL override for tests or proxies.
- `end_date` (optional, string); Optional end query parameter (Appfigures date filter upper bound).
- `group_by` (optional, string); Optional group_by query parameter passed through to Appfigures list
  endpoints.
- `search_store` (optional, string); Optional store filter (e.g. apple, google_play) sent as the
  store query parameter.
- `start_date` (optional, string); Optional start query parameter (Appfigures date filter lower
  bound).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.appfigures.com/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/products/mine`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `count`; starts at
1; page size 100.

Pagination by stream: none: `products`, `sales`, `ratings`, `categories`, `revenue`,
`subscriptions`, `ads`, `estimates`, `events`, `external_accounts`, `account_info`,
`data_countries`, `data_languages`, `data_stores`, `data_currencies`, `data_sdks`; page_number:
`reviews`, `users`.

- `reviews`: GET `/reviews` - records path `reviews`; query `end` from template `{{ config.end_date
  }}`, omitted when absent; `group_by` from template `{{ config.group_by }}`, omitted when absent;
  `start` from template `{{ config.start_date }}`, omitted when absent; `store` from template `{{
  config.search_store }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `count`; starts at 1; page size 100.
- `products`: GET `/products/mine` - records at response root; flattens keyed objects.
- `sales`: GET `/reports/sales` - records at response root; flattens keyed objects; query `end` from
  template `{{ config.end_date }}`, omitted when absent; `group_by` from template `{{
  config.group_by }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `store` from template `{{ config.search_store }}`, omitted when absent.
- `ratings`: GET `/reports/ratings` - records at response root; flattens keyed objects; query `end`
  from template `{{ config.end_date }}`, omitted when absent; `group_by` from template `{{
  config.group_by }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `store` from template `{{ config.search_store }}`, omitted when absent.
- `categories`: GET `/data/categories` - records at response root; flattens keyed objects.
- `revenue`: GET `/reports/revenue` - records at response root; query `end` from template `{{
  config.end_date }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `store` from template `{{ config.search_store }}`, omitted when absent; computed
  output fields `report`.
- `subscriptions`: GET `/reports/subscriptions` - records at response root; query `end` from
  template `{{ config.end_date }}`, omitted when absent; `start` from template `{{ config.start_date
  }}`, omitted when absent; `store` from template `{{ config.search_store }}`, omitted when absent;
  computed output fields `report`.
- `ads`: GET `/reports/ads` - records at response root; query `end` from template `{{
  config.end_date }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `store` from template `{{ config.search_store }}`, omitted when absent; computed
  output fields `report`.
- `estimates`: GET `/reports/estimates` - records at response root; query `end` from template `{{
  config.end_date }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `store` from template `{{ config.search_store }}`, omitted when absent; computed
  output fields `report`.
- `events`: GET `/events/` - records at response root; flattens keyed objects.
- `external_accounts`: GET `/external_accounts` - records at response root; flattens keyed objects.
- `users`: GET `/users` - records path `results`; page-number pagination; page parameter `page`;
  size parameter `count`; starts at 1; page size 100.
- `account_info`: GET `/` - records at response root; computed output fields `daily_limit`,
  `daily_used`, `user_email`, `user_id`, `user_name`.
- `data_countries`: GET `/data/countries` - records at response root; flattens keyed objects; key
  field `iso`.
- `data_languages`: GET `/data/languages` - records at response root; flattens keyed objects.
- `data_stores`: GET `/data/stores` - records at response root; flattens keyed objects; key field
  `store_key`.
- `data_currencies`: GET `/data/currencies` - records at response root.
- `data_sdks`: GET `/data/sdks` - records at response root.

## Write actions & risks

Overall write risk: external Appfigures API mutation - publishes a public review response, and
creates/edits/deletes release-event markers overlaid on analytics charts.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `reply_to_review`: POST `/reviews/{{ record.id }}/response` - kind `update`; body type `json`;
  path fields `id`; body fields `content`; required record fields `id`, `content`; accepted fields
  `content`, `id`; risk: publishes a developer response to a customer review, visible on the public
  app store listing.
- `create_event`: POST `/events/` - kind `create`; body type `json`; required record fields
  `caption`, `date`; accepted fields `caption`, `date`, `details`, `products`; risk: creates a
  release/marketing event marker overlaid on every Appfigures analytics chart.
- `update_event`: PUT `/events/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `caption`, `date`, `id`, `products`; risk: mutates an
  existing release/marketing event marker overlaid on every Appfigures analytics chart.
- `delete_event`: DELETE `/events/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently deletes an event marker from every Appfigures analytics chart.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 18 stream-backed endpoint group(s), 4 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=6, out_of_scope=13.
