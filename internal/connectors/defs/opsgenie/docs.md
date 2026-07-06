# Overview

Reads Opsgenie alerts, incidents, users, teams, and services through the Opsgenie REST API.

Readable streams: `alerts`, `incidents`, `users`, `teams`, `services`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.opsgenie.com/docs/api-overview.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Opsgenie API integration token, sent as 'Authorization:
  GenieKey <api_token>'. Never logged.
- `base_url` (optional, string); default `https://api.opsgenie.com/v2`; format `uri`; Opsgenie API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.opsgenie.com/v2`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `GenieKey` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/alerts`.

## Streams notes

Default pagination: single request; no pagination.

- `alerts`: GET `/alerts` - records path `data`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `paging.next`; next URLs stay on the configured API host; computed
  output fields `created_at`, `last_occurred_at`, `tiny_id`, `updated_at`.
- `incidents`: GET `/incidents` - records path `data`; query `limit`=`100`; follows a next-page URL
  from the response body; URL path `paging.next`; next URLs stay on the configured API host;
  computed output fields `created_at`, `impacted_services`, `owner_team`, `tiny_id`, `updated_at`.
- `users`: GET `/users` - records path `data`; query `limit`=`100`; follows a next-page URL from the
  response body; URL path `paging.next`; next URLs stay on the configured API host; computed output
  fields `full_name`, `time_zone`.
- `teams`: GET `/teams` - records path `data`; query `limit`=`100`; follows a next-page URL from the
  response body; URL path `paging.next`; next URLs stay on the configured API host; computed output
  fields `created_at`, `updated_at`.
- `services`: GET `/services` - records path `data`; query `limit`=`100`; follows a next-page URL
  from the response body; URL path `paging.next`; next URLs stay on the configured API host;
  computed output fields `created_at`, `team_id`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Opsgenie API read of alerting/incident/team
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=7.
