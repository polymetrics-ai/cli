# Overview

Reads Churnkey cancel-flow sessions and aggregated session counts through the Churnkey Data API, and
sends usage/billing events and customer attribute updates through the Churnkey Event Tracking API.

Readable streams: `sessions`, `session_aggregation`.

Write actions: `create_event`, `update_customer`, `set_billing_users`.

Service API documentation: https://docs.churnkey.co/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Churnkey Data API key. Sent as the x-ck-api-key header;
  never logged.
- `app_id` (required, string); Churnkey application id. Sent as the x-ck-app header on every
  request.
- `base_url` (optional, string); default `https://api.churnkey.co`; format `uri`; Churnkey API host.
  Individual streams/writes append their own /v1/data or /v1/api path prefix (the Data API and the
  Event Tracking API share this host but live at different path prefixes).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.churnkey.co`.

Authentication behavior:

- API key authentication in `x-ck-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/data/sessions`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `session_aggregation`; offset_limit: `sessions`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `sessions`: GET `/v1/data/sessions` - records path `.`; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; incremental cursor `created_at`; formatted as
  `rfc3339`; computed output fields `accepted_offer`, `blueprint_id`, `created_at`,
  `customer_billing_interval`, `customer_email`, `customer_id`, `customer_plan_id`,
  `discount_cooldown_applied`, `offer_type`, `segment_id`, `survey_choice_id`,
  `survey_choice_value`, `survey_id`, `updated_at`.
- `session_aggregation`: GET `/v1/data/session-aggregation` - records path `.`; computed output
  fields `aborted`, `billing_interval`, `canceled`, `count`, `month`, `offer_type`, `plan_id`,
  `save_type`, `trial`.

## Write actions & risks

Overall write risk: external mutation of Churnkey customer event/attribute data used to drive
cancel-flow targeting; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_event`: POST `/v1/api/events/new` - kind `create`; body type `json`; required record
  fields `event`, `customerId`; accepted fields `customerId`, `event`, `eventData`, `eventDate`,
  `uid`, `user`; risk: external mutation; records a usage/billing event against a Churnkey customer,
  influencing cancel-flow offer targeting; approval required.
- `update_customer`: POST `/v1/api/events/customer-update` - kind `update`; body type `json`;
  accepted fields `customerData`, `customerId`, `uid`, `user`; risk: external mutation; overwrites a
  Churnkey customer's tracked attributes used to drive cancel-flow segmentation and offer
  eligibility; approval required.
- `set_billing_users`: POST `/v1/api/events/customer-update/set-users` - kind `update`; body type
  `json`; required record fields `customerId`, `users`; accepted fields `customerId`, `users`; risk:
  external mutation; overwrites which users on a Churnkey customer account receive Payment Recovery
  billing-contact emails; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=2, non_data_endpoint=1, out_of_scope=1.
