# Overview

Reads ChargeDesk charges, customers, subscriptions, and products through the ChargeDesk REST API.

Readable streams: `charges`, `customers`, `subscriptions`, `products`, `log_activity`,
`log_cancellations`, `webhook_notifications`.

Write actions: `create_customer`, `update_customer`, `delete_customer`, `update_charge`,
`delete_charge`, `refund_charge`, `capture_charge`, `void_charge`, `cancel_subscription`,
`create_webhook`, `delete_webhook`, `create_agent`, `delete_agent`.

Service API documentation: https://chargedesk.com/api-docs.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.chargedesk.com/v1`; format `uri`; ChargeDesk
  API base URL override for tests or proxies.
- `mode` (optional, string).
- `password` (required, secret, string); ChargeDesk secret API key, used as the HTTP Basic auth
  username (blank password) unless username is also set. Never logged.
- `username` (optional, string); Optional explicit HTTP Basic auth username override; when set,
  password is sent as its Basic auth password instead of a blank password.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.chargedesk.com/v1`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password` when `{{ config.username
  }}`.
- HTTP Basic authentication using `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/charges`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `count`;
page size 100.

Pagination by stream: none: `webhook_notifications`; offset_limit: `charges`, `customers`,
`subscriptions`, `products`, `log_activity`, `log_cancellations`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `charges`: GET `/charges` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; incremental cursor `occurred`; formatted as
  `rfc3339`.
- `customers`: GET `/customers` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; incremental cursor `occurred`; formatted as
  `rfc3339`.
- `subscriptions`: GET `/subscriptions` - records path `data`; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; incremental cursor `occurred`;
  formatted as `rfc3339`.
- `products`: GET `/products` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; incremental cursor `occurred`; formatted as
  `rfc3339`.
- `log_activity`: GET `/log/activity` - records path `data`; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; incremental cursor `occurred`;
  formatted as `rfc3339`.
- `log_cancellations`: GET `/log/cancellations` - records path `data`; offset/limit pagination;
  offset parameter `offset`; limit parameter `count`; page size 100; incremental cursor `occurred`;
  formatted as `rfc3339`.
- `webhook_notifications`: GET `/webhooks/notifications` - records path `.`.

## Write actions & risks

Overall write risk: external mutations creating/updating/deleting customers, charges, webhooks, and
agents, plus live gateway methods (refund/capture/void a charge, cancel a subscription) that mutate
the connected payment gateway; every write action requires approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `form`; accepted fields `country`,
  `customer_id`, `email`, `name`, `phone`, `tax_number`; risk: external mutation creating a new
  ChargeDesk customer record; approval required.
- `update_customer`: POST `/customers/{{ record.customer_id }}` - kind `update`; body type `form`;
  path fields `customer_id`; required record fields `customer_id`; accepted fields `country`,
  `customer_id`, `delinquent`, `email`, `name`, `phone`; risk: external mutation updating an
  existing ChargeDesk customer record; approval required.
- `delete_customer`: DELETE `/customers/{{ record.customer_id }}` - kind `delete`; body type `none`;
  path fields `customer_id`; required record fields `customer_id`; accepted fields `customer_id`;
  missing records treated as success for status `404`; risk: irreversible deletion of a customer
  record (and, by ChargeDesk's own default, all associated charges/tickets); approval required.
- `update_charge`: POST `/charges/{{ record.charge_id }}` - kind `update`; body type `form`; path
  fields `charge_id`; required record fields `charge_id`; accepted fields `amount`, `charge_id`,
  `currency`, `customer_id`, `status`; risk: external mutation updating an existing charge record's
  stored data; approval required.
- `delete_charge`: DELETE `/charges/{{ record.charge_id }}` - kind `delete`; body type `none`; path
  fields `charge_id`; required record fields `charge_id`; accepted fields `charge_id`; missing
  records treated as success for status `404`; risk: irreversible deletion of a charge record;
  approval required.
- `refund_charge`: POST `/gateway/charges/{{ record.charge_id }}/refund` - kind `update`; body type
  `form`; path fields `charge_id`; required record fields `charge_id`; accepted fields `amount`,
  `charge_id`, `log_reason`; risk: gateway method; irreversibly refunds a charge (full or partial)
  on the originating payment gateway as well as ChargeDesk; approval required.
- `capture_charge`: POST `/gateway/charges/{{ record.charge_id }}/capture` - kind `update`; body
  type `form`; path fields `charge_id`; required record fields `charge_id`; accepted fields
  `amount`, `charge_id`; risk: gateway method; captures (settles) a previously authorized charge on
  the originating payment gateway; approval required.
- `void_charge`: POST `/gateway/charges/{{ record.charge_id }}/void` - kind `update`; body type
  `none`; path fields `charge_id`; required record fields `charge_id`; accepted fields `charge_id`;
  risk: gateway method; voids an authorized charge or cancels an outstanding payment request on the
  originating payment gateway; approval required.
- `cancel_subscription`: POST `/gateway/subscriptions/{{ record.subscription_id }}/cancel` - kind
  `update`; body type `form`; path fields `subscription_id`; required record fields
  `subscription_id`; accepted fields `log_reason`, `subscription_id`; risk: gateway method;
  irreversibly cancels future recurring charges for a subscription on the originating payment
  gateway as well as ChargeDesk; approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `form`; required record fields
  `url`; accepted fields `all`, `notifications`, `url`; risk: external mutation creating a new
  outbound webhook subscription that will POST ChargeDesk event data to a third-party URL; approval
  required.
- `delete_webhook`: DELETE `/webhooks/{{ record.webhook_id }}` - kind `delete`; body type `none`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`;
  missing records treated as success for status `404`; risk: irreversible removal of an outbound
  webhook subscription; approval required.
- `create_agent`: POST `/agents` - kind `create`; body type `form`; required record fields `name`,
  `email`, `role`; accepted fields `email`, `name`, `role`; risk: external mutation inviting a new
  support agent (or updating an existing agent's role) with account access to ChargeDesk; approval
  required.
- `delete_agent`: DELETE `/agents/{{ record.email }}` - kind `delete`; body type `none`; path fields
  `email`; required record fields `email`; accepted fields `email`; missing records treated as
  success for status `404`; risk: irreversible removal of a support agent's ChargeDesk account
  access; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=4, non_data_endpoint=1, out_of_scope=10.
