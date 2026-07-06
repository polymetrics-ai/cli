# Overview

Reads and writes Brex transactions, users, expenses, vendors, budgets, cards, accounts, statements,
transfers, and webhooks through the Brex platform REST API.

Readable streams: `transactions`, `users`, `expenses`, `vendors`, `budgets`, `departments`,
`locations`, `titles`, `legal_entities`, `cards`, `accounts_card`, `accounts_cash`,
`card_statements`, `linked_accounts`, `transfers`, `webhooks`.

Write actions: `update_vendor`, `delete_vendor`, `create_department`, `create_location`,
`create_title`, `create_user`, `update_user`, `update_card`, `lock_card`, `unlock_card`,
`terminate_card`, `update_expense`, `update_webhook`, `delete_webhook`.

Service API documentation: https://developer.brex.com/openapi/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://platform.brexapis.com`; format `uri`; Brex API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for
  transactions/expenses; only records posted/purchased at or after this time are read.
- `user_token` (required, secret, string); Brex API user access token. Sent as Authorization: Bearer
  <user_token>; never logged.

Secret fields are redacted in logs and write previews: `user_token`.

Default configuration values: `base_url=https://platform.brexapis.com`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.user_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/users` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `next_cursor`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `transactions`: GET `/v2/transactions/card/primary` - records path `items`; query `limit`=`100`;
  cursor pagination; cursor parameter `cursor`; next token from `next_cursor`; incremental cursor
  `posted_at_date`; sent as `posted_at_start`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `users`: GET `/v2/users` - records path `items`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`.
- `expenses`: GET `/v1/expenses/card` - records path `items`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next_cursor`; incremental cursor
  `purchased_at`; sent as `purchased_at_start`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `vendors`: GET `/v1/vendors` - records path `items`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`.
- `budgets`: GET `/v2/budgets` - records path `items`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`.
- `departments`: GET `/v2/departments` - records path `items`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next_cursor`.
- `locations`: GET `/v2/locations` - records path `items`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`.
- `titles`: GET `/v2/titles` - records path `items`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`.
- `legal_entities`: GET `/v2/legal_entities` - records path `items`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next_cursor`.
- `cards`: GET `/v2/cards` - records path `items`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`.
- `accounts_card`: GET `/v2/accounts/card` - records at response root; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`.
- `accounts_cash`: GET `/v2/accounts/cash` - records path `items`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next_cursor`.
- `card_statements`: GET `/v2/accounts/card/primary/statements` - records path `items`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `next_cursor`.
- `linked_accounts`: GET `/v1/linked_accounts` - records path `items`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next_cursor`.
- `transfers`: GET `/v1/transfers` - records path `items`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`.
- `webhooks`: GET `/v1/webhooks` - records path `items`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`.

## Write actions & risks

Overall write risk: external mutation of vendors, org directory
(departments/locations/titles/users), card controls/lifecycle, expenses, and webhooks; card
lock/unlock/terminate take effect on real payment instruments immediately.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_vendor`: PUT `/v1/vendors/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `beneficiary_name`, `company_name`, `email`,
  `id`, `phone`; risk: mutates an existing vendor's name, contact details, or payment accounts;
  affects future transfer counterparty resolution.
- `delete_vendor`: DELETE `/v1/vendors/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a vendor record; any transfer still referencing it as
  counterparty will fail to resolve.
- `create_department`: POST `/v2/departments` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `description`, `name`; risk: creates a new organizational
  department; low-risk external mutation, no approval required.
- `create_location`: POST `/v2/locations` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `description`, `name`; risk: creates a new organizational location;
  low-risk external mutation, no approval required.
- `create_title`: POST `/v2/titles` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: creates a new job title; low-risk external mutation, no
  approval required.
- `create_user`: POST `/v2/users` - kind `create`; body type `json`; required record fields `email`,
  `first_name`, `last_name`; accepted fields `department_id`, `email`, `first_name`, `last_name`,
  `location_id`, `manager_id`, `title_id`; risk: invites a new user to the Brex account; sends a
  real invitation email to the target address.
- `update_user`: PUT `/v2/users/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `department_id`, `id`, `location_id`,
  `manager_id`, `status`, `title_id`; risk: mutates an existing user's status, manager, department,
  location, or title; setting status to a terminated/suspended state revokes account access.
- `update_card`: PUT `/v2/cards/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `spend_controls`; risk: mutates an
  existing card's spend controls (limit amount/category/merchant restrictions); takes effect on the
  physical/virtual card immediately.
- `lock_card`: POST `/v2/cards/{{ record.id }}/lock` - kind `custom`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: immediately blocks all new
  transactions on the card until unlocked; does not affect already-authorized/pending transactions.
- `unlock_card`: POST `/v2/cards/{{ record.id }}/unlock` - kind `custom`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: immediately re-enables new
  transactions on a previously locked card.
- `terminate_card`: POST `/v2/cards/{{ record.id }}/terminate` - kind `custom`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: permanently deactivates
  a card; irreversible, the card can never be unlocked or reused after termination.
- `update_expense`: PUT `/v1/expenses/card/{{ record.expense_id }}` - kind `update`; body type
  `json`; path fields `expense_id`; required record fields `expense_id`; accepted fields
  `expense_id`, `memo`; risk: mutates an existing card expense's memo; low-risk metadata-only
  external mutation.
- `update_webhook`: PUT `/v1/webhooks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `url`, `event_types`, `status`; accepted fields
  `event_types`, `group_id`, `id`, `status`, `url`; risk: re-points an already-registered webhook's
  delivery URL, event set, or active status; redirects live event delivery immediately.
- `delete_webhook`: DELETE `/v1/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a webhook subscription; irreversible.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 16 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, deprecated=7, destructive_admin=2, duplicate_of=19, non_data_endpoint=2,
  out_of_scope=35, requires_elevated_scope=12.
