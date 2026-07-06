# Overview

Reads Trustpilot business-unit reviews, invitations, and business-unit profile metadata.

Readable streams: `reviews`, `invitations`, `business_units`, `categories`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.trustpilot.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Trustpilot API key, sent as the apikey query parameter on
  every request. Never logged.
- `base_url` (optional, string); default `https://api.trustpilot.com`; format `uri`; Trustpilot API
  base URL override for tests or proxies.
- `business_unit_id` (required, string); Trustpilot business unit id the
  reviews/invitations/business_units streams are scoped to; substituted into each stream's
  business-unit-scoped path.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.trustpilot.com`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/business-units/{{ config.business_unit_id }}/reviews` with query
`page`=`1`; `perPage`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `perPage`; starts
at 1; page size 100; maximum 1 page(s).

Pagination by stream: none: `categories`; page_number: `reviews`, `invitations`, `business_units`.

- `reviews`: GET `/v1/business-units/{{ config.business_unit_id }}/reviews` - records path
  `reviews`; page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1;
  page size 100; maximum 1 page(s); computed output fields `created_at`.
- `invitations`: GET `/v1/private/business-units/{{ config.business_unit_id }}/invitations` -
  records path `invitations`; page-number pagination; page parameter `page`; size parameter
  `perPage`; starts at 1; page size 100; maximum 1 page(s); computed output fields `created_at`.
- `business_units`: GET `/v1/business-units/{{ config.business_unit_id }}` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1; page size
  100; maximum 1 page(s); computed output fields `display_name`.
- `categories`: GET `/v1/business-units/{{ config.business_unit_id }}/categories` - records path
  `categories`; computed output fields `category_id`, `display_name`, `is_primary`.

## Write actions & risks

This connector is read-only. Read behavior: external Trustpilot API read of business-unit reviews,
invitations, and profile metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=8, duplicate_of=5, non_data_endpoint=6, out_of_scope=11,
  requires_elevated_scope=25.
