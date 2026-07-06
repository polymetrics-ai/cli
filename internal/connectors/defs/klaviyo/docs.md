# Overview

Reads Klaviyo profiles, events, campaigns, lists, metrics, and segments through the Klaviyo REST
(JSON:API) API.

Readable streams: `profiles`, `events`, `campaigns`, `lists`, `metrics`, `segments`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developers.klaviyo.com/en/docs/api_versioning_and_deprecation_policy.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Klaviyo private API key. Sent as the Authorization header
  ("Klaviyo-API-Key <key>"); never logged.
- `base_url` (optional, string); default `https://a.klaviyo.com/api`; format `uri`; Klaviyo API base
  URL override for tests or proxies.
- `mode` (optional, string).
- `revision` (optional, string); default `2024-10-15`; Klaviyo API revision (date-versioned); sent
  as the revision header.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://a.klaviyo.com/api`, `revision=2024-10-15`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Klaviyo-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `links.next`; next URLs
stay on the configured API host.

- `profiles`: GET `/profiles` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  computed output fields `created`, `email`, `external_id`, `first_name`, `last_name`,
  `organization`, `phone_number`, `updated`.
- `events`: GET `/events` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; computed
  output fields `datetime`, `timestamp`, `uuid`.
- `campaigns`: GET `/campaigns` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  computed output fields `archived`, `channel`, `created_at`, `name`, `scheduled_at`, `send_time`,
  `status`, `updated_at`.
- `lists`: GET `/lists` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; computed
  output fields `created`, `name`, `updated`.
- `metrics`: GET `/metrics` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; computed
  output fields `created`, `integration_name`, `name`, `updated`.
- `segments`: GET `/segments` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  computed output fields `created`, `is_active`, `is_processing`, `name`, `updated`.

## Write actions & risks

This connector is read-only. Read behavior: external Klaviyo API read of customer profile, event,
and campaign data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=2.
