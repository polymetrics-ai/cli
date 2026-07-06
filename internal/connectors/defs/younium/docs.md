# Overview

Reads and writes Younium accounts, subscriptions, invoices, products, payment terms, currencies, and
webhooks through the Younium REST API.

Readable streams: `accounts`, `subscriptions`, `invoices`, `products`, `payment_terms`,
`currencies`, `webhooks`.

Write actions: `create_account`, `update_account`, `cancel_subscription`, `post_invoice`,
`cancel_invoice`.

Service API documentation: https://developer.younium.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.younium.com`; format `uri`; Younium API root.
  Also usable as a base URL override for tests/proxies.
- `legal_entity` (optional, string); Optional legal entity code, sent as the X-Younium-Legal-Entity
  request header when set. Omitted entirely when unset.
- `mode` (optional, string).
- `password` (required, secret, string); Younium API password/account key, used with username for
  Basic auth. Never logged.
- `username` (required, string); Younium API username, used with password for Basic auth.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.younium.com`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/Accounts`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `accounts`, `subscriptions`, `invoices`, `webhooks`; offset_limit:
`products`, `payment_terms`, `currencies`.

- `accounts`: GET `/Accounts` - records path `data`; computed output fields `account_id`, `id`,
  `name`, `updated_at`; emits passthrough records.
- `subscriptions`: GET `/Subscriptions` - records path `data`; computed output fields `account_id`,
  `id`, `name`, `updated_at`; emits passthrough records.
- `invoices`: GET `/Invoices` - records path `data`; computed output fields `account_id`, `id`,
  `name`, `updated_at`; emits passthrough records.
- `products`: GET `/Products` - records path `data`; offset/limit pagination; offset parameter
  `Skip`; limit parameter `Take`; page size 100; computed output fields `id`, `name`, `updated_at`;
  emits passthrough records.
- `payment_terms`: GET `/PaymentTerms` - records path `data`; offset/limit pagination; offset
  parameter `Skip`; limit parameter `Take`; page size 100; computed output fields `id`, `name`;
  emits passthrough records.
- `currencies`: GET `/Currency` - records path `data`; offset/limit pagination; offset parameter
  `Skip`; limit parameter `Take`; page size 100; computed output fields `id`, `name`; emits
  passthrough records.
- `webhooks`: GET `/Webhooks` - records path `data`; computed output fields `id`, `name`; emits
  passthrough records.

## Write actions & risks

Overall write risk: external mutation of billing-critical Younium records: account create/update,
subscription cancellation (ends future billing), and invoice posting/cancellation.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_account`: POST `/Accounts` - kind `create`; body type `json`; required record fields
  `name`, `currency`; accepted fields `accountType`, `currency`, `defaultPaymentTerm`,
  `invoiceEmailAddress`, `invoiceEmailCcAddresses`, `name`, `organizationNumber`, `ourReference`,
  `taxRegistrationNumber`, `yourReference`; risk: creates a new billing account in Younium; external
  mutation, approval required.
- `update_account`: PATCH `/Accounts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `accountType`, `currency`,
  `defaultPaymentTerm`, `id`, `invoiceEmailAddress`, `name`, `organizationNumber`, `ourReference`,
  `taxRegistrationNumber`, `yourReference`; risk: mutates an existing account's billing/contact/tax
  metadata; external mutation, approval required.
- `cancel_subscription`: POST `/Subscriptions/cancel/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `cancellationDate`, `cancellationMode`;
  accepted fields `cancellationDate`, `cancellationMode`, `endDate`, `id`; confirmation
  `destructive`; risk: irreversibly schedules or immediately cancels an active subscription, ending
  future billing; external mutation, approval required.
- `post_invoice`: POST `/Invoices/{{ record.id }}/post` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: finalizes a draft invoice,
  making it official/sendable to the customer; external mutation, approval required.
- `cancel_invoice`: POST `/Invoices/{{ record.id }}/cancel` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  irreversibly cancels a posted invoice; external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 5 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=7, duplicate_of=16, non_data_endpoint=2, out_of_scope=52,
  requires_elevated_scope=6.
