# Overview

Reads Wufoo forms, fields, entries, comments, reports, and widgets, and writes entry submissions and
webhook registrations through the Wufoo API.

Readable streams: `forms`, `form_fields`, `entries`, `form_comments`, `reports`, `report_fields`,
`report_entries`, `report_widgets`.

Write actions: `submit_entry`, `add_webhook`, `delete_webhook`.

Service API documentation: https://wufoo.github.io/docs/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Wufoo API key, sent as the HTTP Basic username (password is
  the literal 'pass'). Never logged.
- `base_url` (optional, string); default `https://example.wufoo.com/api/v3`; format `uri`; Wufoo API
  base URL, e.g. https://<subdomain>.wufoo.com/api/v3.
- `form_hash` (optional, string); Wufoo form hash to read/write against; required for the
  form_fields, entries, form_comments streams and the submit_entry/add_webhook/delete_webhook
  writes, substituted into /forms/<form_hash>/... paths.
- `max_pages` (optional, string); default `1`.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).
- `report_hash` (optional, string); Wufoo report hash to read against; required for the
  report_fields, report_entries, report_widgets streams, substituted into /reports/<report_hash>/...
  paths.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://example.wufoo.com/api/v3`, `max_pages=1`,
`page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100; maximum 1 page(s).

Pagination by stream: none: `form_fields`, `report_fields`, `report_widgets`; offset_limit:
`form_comments`, `report_entries`; page_number: `forms`, `entries`, `reports`.

- `forms`: GET `/forms.json` - records path `Forms`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1; page size 100; maximum 1 page(s); emits passthrough
  records.
- `form_fields`: GET `/forms/{{ config.form_hash }}/fields.json` - records path `Fields`; emits
  passthrough records.
- `entries`: GET `/forms/{{ config.form_hash }}/entries.json` - records path `Entries`; page-number
  pagination; page parameter `page`; size parameter `pageSize`; starts at 1; page size 100; maximum
  1 page(s); emits passthrough records.
- `form_comments`: GET `/forms/{{ config.form_hash }}/comments.json` - records path `Comments`;
  offset/limit pagination; offset parameter `pageStart`; limit parameter `pageSize`; page size 25;
  emits passthrough records.
- `reports`: GET `/reports.json` - records path `Reports`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `report_fields`: GET `/reports/{{ config.report_hash }}/fields.json` - records path `Fields`;
  emits passthrough records.
- `report_entries`: GET `/reports/{{ config.report_hash }}/entries.json` - records path `Entries`;
  offset/limit pagination; offset parameter `pageStart`; limit parameter `pageSize`; page size 100;
  emits passthrough records.
- `report_widgets`: GET `/reports/{{ config.report_hash }}/widgets.json` - records path `Widgets`;
  emits passthrough records.

## Write actions & risks

Overall write risk: external mutation: submits live form entries and registers/removes webhook
callback URLs.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `submit_entry`: POST `/forms/{{ config.form_hash }}/entries.json` - kind `create`; body type
  `form`; risk: external mutation; creates a live Wufoo form entry; approval required.
- `add_webhook`: PUT `/forms/{{ config.form_hash }}/webhooks.json` - kind `create`; body type
  `form`; required record fields `url`; accepted fields `handshakeKey`, `metadata`, `url`; risk:
  external mutation; registers a webhook callback URL on the configured form; approval required.
- `delete_webhook`: DELETE `/forms/{{ config.form_hash }}/webhooks/{{ record.hash }}.json` - kind
  `delete`; body type `none`; path fields `hash`; required record fields `hash`; accepted fields
  `hash`; missing records treated as success for status `404`; risk: irreversible external deletion;
  removes a registered webhook from the configured form; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, non_data_endpoint=3, out_of_scope=1, requires_elevated_scope=1.
