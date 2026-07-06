# Overview

Reads Insightly CRM contacts, organisations, opportunities, leads, projects, and tasks through the
Insightly REST API v3.1.

Readable streams: `contacts`, `organisations`, `opportunities`, `leads`, `projects`, `tasks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.na1.insightly.com/v3.1/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.na1.insightly.com/v3.1`; format `uri`;
  Insightly API base URL, pod-specific (e.g. https://api.<pod>.insightly.com/v3.1).
- `mode` (optional, string).
- `token` (required, secret, string); Insightly API token, sent as the HTTP Basic auth username with
  a blank password. Never logged.

Secret fields are redacted in logs and write previews: `token`.

Default configuration values: `base_url=https://api.na1.insightly.com/v3.1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/Contacts` with query `top`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `skip`; limit parameter `top`.

- `contacts`: GET `/Contacts` - records at response root; offset/limit pagination; offset parameter
  `skip`; limit parameter `top`; page size 100; computed output fields `contact_id`,
  `date_created_utc`, `date_updated_utc`, `email_address`, `first_name`, `id`, `last_name`,
  `organisation_id`, `phone`, `title`.
- `organisations`: GET `/Organisations` - records at response root; offset/limit pagination; offset
  parameter `skip`; limit parameter `top`; page size 100; computed output fields `date_created_utc`,
  `date_updated_utc`, `id`, `organisation_id`, `organisation_name`, `owner_user_id`, `phone`,
  `website`.
- `opportunities`: GET `/Opportunities` - records at response root; offset/limit pagination; offset
  parameter `skip`; limit parameter `top`; page size 100; computed output fields `bid_currency`,
  `date_created_utc`, `date_updated_utc`, `id`, `opportunity_id`, `opportunity_name`,
  `opportunity_state`, `opportunity_value`, `pipeline_id`, `probability`, `stage_id`.
- `leads`: GET `/Leads` - records at response root; offset/limit pagination; offset parameter
  `skip`; limit parameter `top`; page size 100; computed output fields `converted`,
  `date_created_utc`, `date_updated_utc`, `email`, `first_name`, `id`, `last_name`, `lead_id`,
  `lead_source_id`, `lead_status_id`, `organisation_name`.
- `projects`: GET `/Projects` - records at response root; offset/limit pagination; offset parameter
  `skip`; limit parameter `top`; page size 100; computed output fields `date_created_utc`,
  `date_updated_utc`, `id`, `owner_user_id`, `pipeline_id`, `project_id`, `project_name`,
  `stage_id`, `status`.
- `tasks`: GET `/Tasks` - records at response root; offset/limit pagination; offset parameter
  `skip`; limit parameter `top`; page size 100; computed output fields `completed`,
  `date_created_utc`, `date_updated_utc`, `due_date`, `id`, `owner_user_id`, `priority`, `status`,
  `task_id`, `title`.

## Write actions & risks

This connector is read-only. Read behavior: external Insightly API read of contacts, organisations,
opportunities, leads, projects, and tasks.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
