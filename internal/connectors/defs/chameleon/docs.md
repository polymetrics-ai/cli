# Overview

Reads Chameleon surveys, tours, launchers, tooltips, and segments through the Chameleon v3 REST API.

Readable streams: `surveys`, `tours`, `launchers`, `tooltips`, `segments`, `embeds`, `event_names`,
`tags`, `deliveries`, `webhooks`, `companies`.

Write actions: `publish_survey`, `publish_tour`, `publish_launcher`, `publish_tooltip`,
`publish_embed`, `create_delivery`, `delete_delivery`, `create_webhook`, `delete_webhook`.

Service API documentation: https://developers.chameleon.io/reference/introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Chameleon account secret, sent as the X-Account-Secret
  header. Never logged.
- `base_url` (optional, string); default `https://api.chameleon.io/v3`; format `uri`; Chameleon API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.chameleon.io/v3`.

Authentication behavior:

- API key authentication in `X-Account-Secret` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/edit/segments`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `before`; next token from `cursor.before`.

Pagination by stream: cursor: `surveys`, `tours`, `launchers`, `tooltips`, `segments`, `embeds`,
`event_names`, `tags`, `deliveries`, `companies`; none: `webhooks`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `surveys`: GET `/edit/surveys` - records path `surveys`; query `limit`=`50`; cursor pagination;
  cursor parameter `before`; next token from `cursor.before`; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `tours`: GET `/edit/tours` - records path `tours`; query `limit`=`50`; cursor pagination; cursor
  parameter `before`; next token from `cursor.before`; incremental cursor `updated_at`; formatted as
  `rfc3339`.
- `launchers`: GET `/edit/launchers` - records path `launchers`; query `limit`=`50`; cursor
  pagination; cursor parameter `before`; next token from `cursor.before`; incremental cursor
  `updated_at`; formatted as `rfc3339`.
- `tooltips`: GET `/edit/tooltips` - records path `tooltips`; query `limit`=`50`; cursor pagination;
  cursor parameter `before`; next token from `cursor.before`; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `segments`: GET `/edit/segments` - records path `segments`; query `limit`=`50`; cursor pagination;
  cursor parameter `before`; next token from `cursor.before`; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `embeds`: GET `/edit/embeds` - records path `embeds`; query `limit`=`50`; cursor pagination;
  cursor parameter `before`; next token from `cursor.before`; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `event_names`: GET `/edit/event_names` - records path `event_names`; query `limit`=`50`; cursor
  pagination; cursor parameter `before`; next token from `cursor.before`; incremental cursor
  `updated_at`; formatted as `rfc3339`.
- `tags`: GET `/edit/tags` - records path `tags`; query `limit`=`50`; cursor pagination; cursor
  parameter `before`; next token from `cursor.before`; incremental cursor `updated_at`; formatted as
  `rfc3339`.
- `deliveries`: GET `/edit/deliveries` - records path `deliveries`; query `limit`=`50`; cursor
  pagination; cursor parameter `before`; next token from `cursor.before`; incremental cursor
  `updated_at`; formatted as `rfc3339`.
- `webhooks`: GET `/edit/webhooks` - records path `webhooks`; query `kind`=`webhook`.
- `companies`: GET `/analyze/companies` - records path `companies`; query `limit`=`50`; cursor
  pagination; cursor parameter `before`; next token from `cursor.before`.

## Write actions & risks

Overall write risk: external mutations publishing/unpublishing in-product experiences,
triggering/cancelling user-targeted Deliveries, and creating/deleting outbound Webhook
subscriptions; every write action requires approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `publish_survey`: PATCH `/edit/surveys/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `published_at`; accepted fields `id`, `published_at`;
  risk: external mutation publishing/unpublishing a live in-product Microsurvey to end-users;
  approval required.
- `publish_tour`: PATCH `/edit/tours/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `published_at`; accepted fields `id`, `published_at`; risk:
  external mutation publishing/unpublishing a live in-product Tour to end-users; approval required.
- `publish_launcher`: PATCH `/edit/launchers/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `published_at`; accepted fields `id`,
  `published_at`; risk: external mutation publishing/unpublishing a live in-product Launcher to
  end-users; approval required.
- `publish_tooltip`: PATCH `/edit/tooltips/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `published_at`; accepted fields `id`, `published_at`;
  risk: external mutation publishing/unpublishing a live in-product Tooltip to end-users; approval
  required.
- `publish_embed`: PATCH `/edit/embeds/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `published_at`; accepted fields `id`, `published_at`;
  risk: external mutation publishing/unpublishing a live in-product Embeddable to end-users;
  approval required.
- `create_delivery`: POST `/edit/deliveries` - kind `create`; body type `json`; required record
  fields `model_kind`, `model_id`; accepted fields `email`, `from`, `idempotency_key`, `model_id`,
  `model_kind`, `once`, `options`, `profile_id`, `uid`, `until`, `use_segmentation`; risk: external
  mutation directly triggering a Tour or Microsurvey experience for one specific end-user; approval
  required.
- `delete_delivery`: DELETE `/edit/deliveries/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: cancels a not-yet-triggered Delivery; irreversible once the target
  has already been shown, approval required.
- `create_webhook`: POST `/edit/webhooks` - kind `create`; body type `json`; required record fields
  `url`, `topics`; accepted fields `experience_id`, `experience_ids`, `kind`, `topics`, `url`; risk:
  external mutation creating a new outbound webhook subscription that will POST Chameleon event data
  to a third-party URL; approval required.
- `delete_webhook`: DELETE `/edit/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible removal of an outbound webhook subscription; approval
  required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 11 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=6, duplicate_of=13, non_data_endpoint=5, out_of_scope=5.
