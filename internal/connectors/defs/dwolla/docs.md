# Overview

Reads Dwolla customers, events, exchange partners, and business classifications, and writes
customer/funding-source/transfer/webhook-subscription/beneficial-owner lifecycle mutations, via the
Dwolla HAL+JSON REST API using OAuth2 client-credentials.

Readable streams: `customers`, `events`, `exchange_partners`, `business_classifications`.

Write actions: `create_customer`, `update_customer`, `create_funding_source`,
`update_funding_source`, `remove_funding_source`, `initiate_micro_deposits`,
`verify_micro_deposits`, `cancel_transfer`, `create_webhook_subscription`,
`update_webhook_subscription`, `delete_webhook_subscription`, `create_beneficial_owner`,
`update_beneficial_owner`, `remove_beneficial_owner`, `certify_beneficial_ownership`.

Service API documentation: https://developers.dwolla.com/api-reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.dwolla.com`; format `uri`; Dwolla API base URL
  override; defaults to production. Set to https://api-sandbox.dwolla.com for sandbox, or a test
  proxy URL.
- `client_id` (required, secret, string); Dwolla OAuth2 client-credentials client ID. Used only for
  the token exchange; never logged.
- `client_secret` (required, secret, string); Dwolla OAuth2 client-credentials client secret. Used
  only for the token exchange; never logged.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.dwolla.com`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.base_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `_links.next.href`;
next URLs stay on the configured API host.

- `customers`: GET `/customers` - records path `_embedded.customers`; query `limit`=`25`; follows a
  next-page URL from the response body; URL path `_links.next.href`; next URLs stay on the
  configured API host.
- `events`: GET `/events` - records path `_embedded.events`; query `limit`=`25`; follows a next-page
  URL from the response body; URL path `_links.next.href`; next URLs stay on the configured API
  host.
- `exchange_partners`: GET `/exchange-partners` - records path `_embedded.exchange-partners`; query
  `limit`=`25`; follows a next-page URL from the response body; URL path `_links.next.href`; next
  URLs stay on the configured API host.
- `business_classifications`: GET `/business-classifications` - records path
  `_embedded.business-classifications`; query `limit`=`25`; follows a next-page URL from the
  response body; URL path `_links.next.href`; next URLs stay on the configured API host.

## Write actions & risks

Overall write risk: external mutation of Dwolla customers, funding sources, transfers (cancel only,
never create/send), webhook subscriptions, and beneficial owners; several actions are
destructive/not reversible (remove_funding_source, cancel_transfer, delete_webhook_subscription,
remove_beneficial_owner) and beneficial-owner actions carry SSN/PII; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `json`; required record fields
  `firstName`, `lastName`, `email`; accepted fields `address1`, `address2`,
  `businessClassification`, `businessName`, `businessType`, `city`, `correlationId`, `dateOfBirth`,
  `doingBusinessAs`, `ein`, `email`, `firstName`, `ipAddress`, `lastName`, `phone`, `postalCode`,
  `ssn`, `state`, and 2 more; risk: external mutation; creates a new Dwolla customer (personal,
  business, receive-only, or unverified), subject to Dwolla's identity-verification rules for the
  requested type.
- `update_customer`: POST `/customers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `address1`, `address2`, `businessName`,
  `city`, `doingBusinessAs`, `email`, `firstName`, `id`, `ipAddress`, `lastName`, `phone`,
  `postalCode`, `state`, `status`, `website`; risk: external mutation; updates a customer's profile
  fields, or its status (e.g. deactivating/reactivating the customer, which blocks/restores its
  ability to transact).
- `create_funding_source`: POST `/customers/{{ record.customer_id }}/funding-sources` - kind
  `create`; body type `json`; path fields `customer_id`; required record fields `customer_id`,
  `name`; accepted fields `accountNumber`, `bankAccountType`, `customer_id`, `name`,
  `onDemandAuthorizationId`, `plaidToken`, `routingNumber`; risk: external mutation; attaches a new
  bank-account funding source to a customer, either as unverified (routingNumber/accountNumber,
  requiring later micro-deposit verification) or pre-verified via an open-banking
  plaidToken/onDemandAuthorizationId.
- `update_funding_source`: POST `/funding-sources/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `accountNumber`,
  `bankAccountType`, `id`, `name`, `routingNumber`; risk: external mutation; renames a funding
  source or replaces its unverified bank-account routing/account numbers.
- `remove_funding_source`: POST `/funding-sources/{{ record.id }}` - kind `delete`; body type
  `json`; path fields `id`; body fields `removed`; required record fields `id`, `removed`; accepted
  fields `id`, `removed`; confirmation `destructive`; risk: destructive external mutation; Dwolla
  has no hard-delete for funding sources, this soft-removes it (POST {removed:true}) so it can no
  longer send/receive transfers; not reversible via the API.
- `initiate_micro_deposits`: POST `/funding-sources/{{ record.id }}/initiate-micro-deposits` - kind
  `custom`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: external mutation; sends two small trial-deposit ACH transactions to an unverified
  bank-account funding source, the first step of micro-deposit verification.
- `verify_micro_deposits`: POST `/funding-sources/{{ record.id }}/verify-micro-deposits` - kind
  `custom`; body type `json`; path fields `id`; body fields `amount1`, `amount2`; required record
  fields `id`, `amount1`, `amount2`; accepted fields `amount1`, `amount2`, `id`; risk: external
  mutation; verifies a funding source's micro-deposit amounts, completing bank-account verification;
  Dwolla locks the funding source after repeated failed attempts.
- `cancel_transfer`: POST `/transfers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `status`; required record fields `id`, `status`; accepted fields `id`,
  `status`; confirmation `destructive`; risk: external mutation; cancels a still-pending transfer
  before it clears, this action is not reversible and only succeeds while the transfer's status is
  pending.
- `create_webhook_subscription`: POST `/webhook-subscriptions` - kind `create`; body type `json`;
  required record fields `url`, `secret`; accepted fields `secret`, `url`; risk: external mutation;
  registers a new webhook subscription; Dwolla enforces a maximum of 10 active subscriptions in
  Sandbox and 5 in Production.
- `update_webhook_subscription`: POST `/webhook-subscriptions/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; body fields `paused`; required record fields `id`, `paused`;
  accepted fields `id`, `paused`; risk: external mutation; pauses or resumes webhook delivery for a
  subscription (Dwolla still generates the events while paused, it just withholds delivery).
- `delete_webhook_subscription`: DELETE `/webhook-subscriptions/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  confirmation `destructive`; risk: destructive external mutation; permanently deletes a webhook
  subscription and stops all future event delivery to it; not reversible.
- `create_beneficial_owner`: POST `/customers/{{ record.customer_id }}/beneficial-owners` - kind
  `create`; body type `json`; path fields `customer_id`; required record fields `customer_id`,
  `firstName`, `lastName`, `dateOfBirth`, `ssn`, `address`; accepted fields `address`,
  `customer_id`, `dateOfBirth`, `firstName`, `lastName`, `ssn`; risk: external mutation; registers a
  beneficial owner (25%+ equity holder) for a business verified customer; submits sensitive PII
  (SSN, date of birth, address) to Dwolla for identity verification.
- `update_beneficial_owner`: POST `/beneficial-owners/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `address`, `dateOfBirth`,
  `firstName`, `id`, `lastName`, `ssn`; risk: external mutation; updates a beneficial owner's
  identity/PII fields, which resets its verification status pending re-review.
- `remove_beneficial_owner`: DELETE `/beneficial-owners/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; confirmation
  `destructive`; risk: destructive external mutation; permanently removes a beneficial owner from a
  business verified customer; not reversible and may affect the customer's beneficial-ownership
  certification status.
- `certify_beneficial_ownership`: POST `/customers/{{ record.customer_id }}/beneficial-ownership` -
  kind `custom`; body type `json`; path fields `customer_id`; body fields `status`; required record
  fields `customer_id`, `status`; accepted fields `customer_id`, `status`; risk: external mutation;
  the Account Admin attests that a business customer's beneficial-owner information is accurate and
  complete, which is required before the customer can transact.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 4 stream-backed endpoint group(s), 15 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=2, duplicate_of=12, non_data_endpoint=7, out_of_scope=16,
  requires_elevated_scope=2.
