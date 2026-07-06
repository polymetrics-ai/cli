# Overview

Reads Spotler CRM contacts, accounts, opportunities, and tasks, and (via the real CRM API v4)
activities, campaigns, and cases.

Readable streams: `contacts`, `accounts`, `opportunities`, `tasks`, `activities`, `campaigns`,
`cases`.

This connector is read-only; no write actions are declared.

Service API documentation: https://support.reallysimplesystems.com/api-v4/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Spotler CRM (Really Simple Systems) API v4 OAuth2
  access token, sent as a Bearer token (Authorization: Bearer <access_token>). Never logged.
  Preferred over api_key when both are present; required by every stream added in this pass
  (activities, campaigns, cases, documents, opportunity_lines) since those target the real, live CRM
  API v4 host directly.
- `api_key` (optional, secret, string); Spotler CRM API key, sent as the X-API-Key header value.
  Never logged.
- `base_url` (optional, string); default `https://api.spotlercrm.com/api/v1`; format `uri`; Spotler
  CRM API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Default configuration values: `base_url=https://api.spotlercrm.com/api/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- API key authentication in `X-API-Key` using `secrets.api_key` when `{{ secrets.api_key }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `contacts`: GET `/contacts` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.
- `accounts`: GET `/accounts` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.
- `opportunities`: GET `/opportunities` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `tasks`: GET `/tasks` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `activities`: GET `https://apiv4.reallysimplesystems.com/activities` - records path `list`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 10;
  computed output fields `createddate`, `id`, `modifieddate`, `ownerid`.
- `campaigns`: GET `https://apiv4.reallysimplesystems.com/campaigns` - records path `list`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 10;
  computed output fields `createddate`, `id`, `modifieddate`, `name`, `ownerid`.
- `cases`: GET `https://apiv4.reallysimplesystems.com/cases` - records path `list`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 10; computed
  output fields `createddate`, `id`, `modifieddate`, `ownerid`.

## Write actions & risks

This connector is read-only. Read behavior: external Spotler CRM API read of contact, account,
opportunity, task, activity, campaign, and case data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, non_data_endpoint=2, out_of_scope=14.
