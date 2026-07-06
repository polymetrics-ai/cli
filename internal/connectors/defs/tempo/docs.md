# Overview

Reads Tempo accounts, customers, worklogs, and workload schemes through the Tempo Cloud REST API v4.

Readable streams: `accounts`, `customers`, `worklogs`, `workload_schemes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://apidocs.tempo.io/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Tempo Cloud API token. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://api.tempo.io/4`; format `uri`; Tempo API base URL
  override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-1000).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.tempo.io/4`, `max_pages=0`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `metadata.next`; next
URLs stay on the configured API host.

- `accounts`: GET `/accounts` - records path `results`; query `limit`=`50`; `offset`=`0`; follows a
  next-page URL from the response body; URL path `metadata.next`; next URLs stay on the configured
  API host; computed output fields `monthly_budget`.
- `customers`: GET `/customers` - records path `results`; query `limit`=`50`; `offset`=`0`; follows
  a next-page URL from the response body; URL path `metadata.next`; next URLs stay on the configured
  API host.
- `worklogs`: GET `/worklogs` - records path `results`; query `limit`=`50`; `offset`=`0`; follows a
  next-page URL from the response body; URL path `metadata.next`; next URLs stay on the configured
  API host; computed output fields `billable_seconds`, `created_at`, `issue_id`, `jira_worklog_id`,
  `start_date`, `start_time`, `tempo_worklog_id`, `time_spent_seconds`, `updated_at`.
- `workload_schemes`: GET `/workload-schemes` - records path `results`; query `limit`=`50`;
  `offset`=`0`; follows a next-page URL from the response body; URL path `metadata.next`; next URLs
  stay on the configured API host; computed output fields `default_scheme`.

## Write actions & risks

This connector is read-only. Read behavior: external Tempo Cloud API read of account, customer, and
worklog data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
