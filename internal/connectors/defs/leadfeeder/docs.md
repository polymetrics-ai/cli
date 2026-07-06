# Overview

Reads Leadfeeder accounts and their leads, visits, and custom feeds through the Leadfeeder JSON:API.

Readable streams: `accounts`, `leads`, `visits`, `custom_feeds`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.leadfeeder.com/api/.

## Auth setup

Connection fields:

- `account_id` (optional, string); Leadfeeder account id. Required only for the leads, visits, and
  custom_feeds streams, which are nested under an account.
- `api_token` (required, secret, string); Leadfeeder API token. Sent as the Authorization header
  ("Token token=<api_token>"); never logged.
- `base_url` (optional, string); default `https://api.leadfeeder.com`; format `uri`; Leadfeeder API
  base URL override for tests or proxies.
- `end_date` (optional, string); RFC3339 or yyyy-mm-dd upper bound for the leads/visits date window.
- `mode` (optional, string).
- `start_date` (optional, string); RFC3339 or yyyy-mm-dd lower bound for the leads/visits date
  window. When set, end_date (or today) is also required by the Leadfeeder API.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.leadfeeder.com`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts` with query `page[size]`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `links.next`; next URLs
stay on the configured API host.

- `accounts`: GET `/accounts` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  computed output fields `currency`, `industry`, `name`, `status`, `time_zone`.
- `leads`: GET `/accounts/{{ config.account_id }}/leads` - records path `data`; query `end_date`
  from template `{{ config.end_date }}`, omitted when absent; `page[size]`=`100`; `start_date` from
  template `{{ config.start_date }}`, omitted when absent; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host; computed output fields
  `city`, `country`, `employee_count`, `first_visit_date`, `industry`, `last_visit_date`, `name`,
  `quality`, `visits`, `website`.
- `visits`: GET `/accounts/{{ config.account_id }}/visits` - records path `data`; query `end_date`
  from template `{{ config.end_date }}`, omitted when absent; `page[size]`=`100`; `start_date` from
  template `{{ config.start_date }}`, omitted when absent; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host; computed output fields
  `ended_at`, `hostname`, `pageviews`, `referring_url`, `source`, `started_at`, `visit_date`,
  `visit_length`.
- `custom_feeds`: GET `/accounts/{{ config.account_id }}/custom-feeds` - records path `data`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host; computed output fields `name`.

## Write actions & risks

This connector is read-only. Read behavior: external Leadfeeder API read of account, lead, and visit
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
