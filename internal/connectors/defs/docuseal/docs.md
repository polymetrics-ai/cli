# Overview

Reads DocuSeal templates, submissions, and submitters, and writes submission/submitter/template
mutations through the DocuSeal REST API.

Readable streams: `templates`, `submissions`, `submitters`, `template_detail`.

Write actions: `create_submission`, `archive_submission`, `update_submitter`, `update_template`,
`archive_template`, `clone_template`.

Service API documentation: https://www.docuseal.co/docs/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); DocuSeal API key, sent as the X-Auth-Token header. Never
  logged.
- `base_url` (optional, string); default `https://api.docuseal.com`; format `uri`; DocuSeal API base
  URL override for self-hosted instances, tests, or proxies.
- `page_size` (optional, integer); default `10`; Page size (1-100) sent as the 'limit' query
  parameter on each list request.
- `template_id` (optional, string); Template ID the 'template_detail' stream reads a single template
  record for. Required only when reading the 'template_detail' stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.docuseal.com`, `page_size=10`.

Authentication behavior:

- API key authentication in `X-Auth-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/templates`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`; next token from `pagination.next`.

Pagination by stream: cursor: `templates`, `submissions`, `submitters`; none: `template_detail`.

- `templates`: GET `/templates` - records path `data`; query `limit`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `after`; next token from `pagination.next`.
- `submissions`: GET `/submissions` - records path `data`; query `limit`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `after`; next token from `pagination.next`; computed output
  fields `template_id`, `template_name`.
- `submitters`: GET `/submitters` - records path `data`; query `limit`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `after`; next token from `pagination.next`.
- `template_detail`: GET `/templates/{{ config.template_id }}` - single-object response; records at
  response root.

## Write actions & risks

Overall write risk: external mutation; sends live signature requests, archives
submissions/templates, and edits submitter/template records in DocuSeal.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_submission`: POST `/submissions` - kind `create`; body type `json`; required record fields
  `template_id`, `submitters`; accepted fields `bcc_completed`, `completed_redirect_url`,
  `expire_at`, `order`, `reply_to`, `send_email`, `send_sms`, `submitters`, `template_id`,
  `variables`; risk: external mutation; dispatches a live signature-request email/SMS to every
  listed submitter unless send_email/send_sms are explicitly set false; approval required.
- `archive_submission`: DELETE `/submissions/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  archives a live DocuSeal submission (soft-delete, still recoverable via the DocuSeal UI); approval
  required.
- `update_submitter`: PUT `/submitters/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `completed`, `completed_redirect_url`,
  `email`, `external_id`, `id`, `metadata`, `name`, `phone`, `reply_to`, `send_email`, `send_sms`,
  `values`; risk: external mutation; overwrites a live DocuSeal submitter's pre-filled
  values/contact info, can re-send signature request notifications, and can force-mark the submitter
  completed/auto-signed; approval required.
- `update_template`: PUT `/templates/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `archived`, `folder_name`, `id`, `name`,
  `roles`; risk: external mutation; renames/moves/relabels a live DocuSeal template and can
  unarchive it; approval required.
- `archive_template`: DELETE `/templates/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation; archives
  a live DocuSeal template (soft-delete, recoverable by unarchiving via update_template); approval
  required.
- `clone_template`: POST `/templates/{{ record.id }}/clone` - kind `create`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `external_id`, `folder_name`, `id`,
  `name`; risk: external mutation; creates a new live DocuSeal template by cloning an existing one;
  approval required.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 4 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=6, out_of_scope=6.
