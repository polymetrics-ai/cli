# Overview

Reads Braintree transactions, customers, subscriptions, reference data, payment methods, disputes,
merchant accounts, and Apple Pay domains through the gateway HTTP API.

Readable streams: `transactions`, `customers`, `subscriptions`, `add_ons`, `discounts`, `plans`,
`merchant_accounts`, `payment_methods`, `disputes`, `apple_pay_domains`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.paypal.com/braintree/docs.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.sandbox.braintreegateway.com`; format `uri`;
  Braintree API base URL. Set to https://api.braintreegateway.com for production.
- `merchant_id` (required, string); Braintree merchant ID; substituted into every request path
  (merchants/{{ config.merchant_id }}/...).
- `mode` (optional, string).
- `page_size` (optional, string); default `100`.
- `private_key` (required, secret, string); Braintree private key, sent as the HTTP Basic auth
  password. Never logged.
- `public_key` (required, secret, string); Braintree public key, sent as the HTTP Basic auth
  username. Never logged.

Secret fields are redacted in logs and write previews: `private_key`, `public_key`.

Default configuration values: `base_url=https://api.sandbox.braintreegateway.com`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.public_key`, `secrets.private_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/merchants/{{ config.merchant_id }}/transactions` with query
`page_size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from
`pagination.next_page`.

Pagination by stream: cursor: `transactions`, `customers`, `subscriptions`; none: `add_ons`,
`discounts`, `plans`, `merchant_accounts`, `payment_methods`, `disputes`, `apple_pay_domains`.

- `transactions`: GET `/merchants/{{ config.merchant_id }}/transactions` - records path
  `transactions`; query `page_size`=`{{ config.page_size }}`; cursor pagination; cursor parameter
  `page`; next token from `pagination.next_page`; emits passthrough records.
- `customers`: GET `/merchants/{{ config.merchant_id }}/customers` - records path `customers`; query
  `page_size`=`{{ config.page_size }}`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next_page`; emits passthrough records.
- `subscriptions`: GET `/merchants/{{ config.merchant_id }}/subscriptions` - records path
  `subscriptions`; query `page_size`=`{{ config.page_size }}`; cursor pagination; cursor parameter
  `page`; next token from `pagination.next_page`; emits passthrough records.
- `add_ons`: GET `/merchants/{{ config.merchant_id }}/add_ons` - records path `add_ons`; emits
  passthrough records.
- `discounts`: GET `/merchants/{{ config.merchant_id }}/discounts` - records path `discounts`; emits
  passthrough records.
- `plans`: GET `/merchants/{{ config.merchant_id }}/plans` - records path `plans`; emits passthrough
  records.
- `merchant_accounts`: GET `/merchants/{{ config.merchant_id }}/merchant_accounts` - records path
  `merchant_accounts`; emits passthrough records.
- `payment_methods`: GET `/merchants/{{ config.merchant_id }}/payment_methods` - records path
  `payment_methods`; emits passthrough records.
- `disputes`: GET `/merchants/{{ config.merchant_id }}/disputes` - records path `disputes`; emits
  passthrough records.
- `apple_pay_domains`: GET `/merchants/{{ config.merchant_id }}/apple_pay/registered_domains` -
  records path `domains`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Braintree API read of transaction, customer,
subscription, reference, dispute, payment method, and merchant account data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=17, duplicate_of=11, non_data_endpoint=3, out_of_scope=16,
  requires_elevated_scope=14.
