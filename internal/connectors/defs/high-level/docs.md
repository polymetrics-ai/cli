# Overview

Reads HighLevel (Go HighLevel / LeadConnector) contacts, opportunities, pipelines, custom fields,
and form submissions for a location through the HighLevel REST API.

Readable streams: `pipelines`, `contacts`, `opportunities`, `custom_fields`, `form_submissions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://highlevel.stoplight.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); HighLevel API key, sent as the x-api-key header on every
  request. Never logged.
- `api_version` (optional, string); default `2021-07-28`; Value sent as the Version header on every
  request (LeadConnector API version).
- `base_url` (optional, string); default `https://api.leadconnectorpro.co`; format `uri`; HighLevel
  (LeadConnector proxy) API base URL override for tests or proxies.
- `location_id` (required, string); HighLevel location id every read is scoped to; sent as the
  locationId query param on every request.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `api_version=2021-07-28`, `base_url=https://api.leadconnectorpro.co`.

Authentication behavior:

- API key authentication in `x-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/upstream/pipelines` with query `locationId`=`{{ config.location_id }}`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `contacts`, `opportunities`, `form_submissions`; none: `pipelines`,
`custom_fields`.

- `pipelines`: GET `/upstream/pipelines` - records path `pipelines`; query `locationId`=`{{
  config.location_id }}`.
- `contacts`: GET `/upstream/contacts` - records path `contacts`; query `limit`=`100`;
  `locationId`=`{{ config.location_id }}`; follows a next-page URL from the response body; URL path
  `meta.nextPageUrl`; next URLs stay on the configured API host.
- `opportunities`: GET `/upstream/opportunities` - records path `opportunities`; query
  `limit`=`100`; `locationId`=`{{ config.location_id }}`; follows a next-page URL from the response
  body; URL path `meta.nextPageUrl`; next URLs stay on the configured API host.
- `custom_fields`: GET `/upstream/customfields` - records path `customFields`; query
  `locationId`=`{{ config.location_id }}`.
- `form_submissions`: GET `/upstream/form-submissions` - records path `submissions`; query
  `limit`=`100`; `locationId`=`{{ config.location_id }}`; follows a next-page URL from the response
  body; URL path `meta.nextPageUrl`; next URLs stay on the configured API host.

## Write actions & risks

This connector is read-only. Read behavior: external HighLevel (LeadConnector) API read of contact,
opportunity, pipeline, custom field, and form submission data for a configured location.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
