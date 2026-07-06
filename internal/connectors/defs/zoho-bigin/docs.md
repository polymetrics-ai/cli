# Overview

Reads and writes Zoho Bigin pipelines, contacts, companies, products, tasks, events, calls, notes,
users, tags, module metadata, and generic module records via the Zoho OAuth 2.0 refresh-token grant.

Readable streams: `pipelines`, `records`, `fields`, `contacts`, `companies`, `products`, `tasks`,
`events`, `calls`, `notes`, `users`, `tags`, `modules`.

Write actions: `create_record`, `update_record`, `upsert_record`, `delete_record`, `create_note`,
`delete_note`.

Service API documentation: https://www.bigin.com/developer/docs/apis/v2/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://www.zohoapis.com/bigin/v2`; format `uri`; Zoho
  Bigin API base URL override for tests or region-specific data centers (e.g.
  https://www.zohoapis.eu/bigin/v2).
- `client_id` (required, secret, string); Zoho OAuth 2.0 client ID for the refresh-token grant. Used
  only in the token-request form; never logged.
- `client_refresh_token` (required, secret, string); Long-lived Zoho OAuth 2.0 refresh token.
  Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope (credentials layer already owns it).
- `client_secret` (required, secret, string); Zoho OAuth 2.0 client secret. Used only in the
  token-request form; never logged.
- `mode` (optional, string).
- `module_name` (optional, string); default `Deals`; Zoho Bigin module API name used by the
  'records' stream (e.g. Deals, Contacts).
- `token_url` (optional, string); default `https://accounts.zoho.com/oauth/v2/token`; format `uri`;
  Zoho OAuth 2.0 token endpoint override. MUST be https in production; the hook fails closed on a
  non-https or unparseable value to prevent exfiltrating the refresh token to an attacker-chosen
  endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token`,
`client_secret`.

Default configuration values: `base_url=https://www.zohoapis.com/bigin/v2`, `module_name=Deals`,
`token_url=https://accounts.zoho.com/oauth/v2/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.client_refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/Pipelines`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 200.

Pagination by stream: none: `fields`, `tags`, `modules`; page_number: `pipelines`, `records`,
`contacts`, `companies`, `products`, `tasks`, `events`, `calls`, `notes`, `users`.

- `pipelines`: GET `/Pipelines` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200.
- `records`: GET `/{{ config.module_name }}` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `name`.
- `fields`: GET `/settings/fields` - records path `fields`; computed output fields `id`.
- `contacts`: GET `/Contacts` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `companies`: GET `/Accounts` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `products`: GET `/Products` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `tasks`: GET `/Tasks` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `events`: GET `/Events` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `calls`: GET `/Calls` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `notes`: GET `/Notes` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `users`: GET `/users` - records path `users`; query `type`=`AllUsers`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 200; emits passthrough
  records.
- `tags`: GET `/settings/tags` - records path `tags`; query `module`=`{{ config.module_name }}`;
  emits passthrough records.
- `modules`: GET `/settings/modules` - records path `modules`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of Zoho Bigin CRM records (create/update/upsert/delete on the
configured module, plus note create/delete); moves real business data, approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_record`: POST `/{{ config.module_name }}` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`, `trigger`; risk: creates one or more new records in
  config.module_name; external mutation, approval required.
- `update_record`: PUT `/{{ config.module_name }}` - kind `update`; body type `json`; required
  record fields `data`; accepted fields `data`, `trigger`; risk: overwrites the named fields of one
  or more existing records in config.module_name; external mutation, approval required.
- `upsert_record`: POST `/{{ config.module_name }}/upsert` - kind `upsert`; body type `json`;
  required record fields `data`; accepted fields `data`, `duplicate_check_fields`, `trigger`; risk:
  inserts a new record in config.module_name if no match is found on duplicate_check_fields,
  otherwise overwrites the matched existing record's submitted fields; external mutation, approval
  required.
- `delete_record`: DELETE `/{{ config.module_name }}/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: permanently deletes a
  single record from config.module_name; external mutation, approval required.
- `create_note`: POST `/{{ config.module_name }}/{{ record.parent_id }}/Notes` - kind `create`; body
  type `json`; path fields `parent_id`; required record fields `parent_id`, `data`; accepted fields
  `data`, `parent_id`; risk: attaches one or more notes to an existing record in config.module_name;
  low-risk external mutation, no approval required.
- `delete_note`: DELETE `/{{ config.module_name }}/{{ record.parent_id }}/Notes/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `parent_id`, `id`; required record fields
  `parent_id`, `id`; accepted fields `id`, `parent_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: permanently deletes a single note from a record in
  config.module_name; external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 13 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, destructive_admin=1, duplicate_of=3, non_data_endpoint=4, out_of_scope=16,
  requires_elevated_scope=2.
