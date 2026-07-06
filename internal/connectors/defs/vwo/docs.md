# Overview

Reads and writes VWO (Visual Website Optimizer) A/B testing campaigns.

Readable streams: `campaigns`, `campaign_variations`.

Write actions: `create_campaign`, `update_campaign`.

Service API documentation: https://developers.wingify.com/reference/get-the-campaigns-of-an-account.

## Auth setup

Connection fields:

- `account_id` (optional, string); default `current`; Set to a specific numeric workspace id to
  target a sub-account instead.
- `api_key` (required, secret, string); VWO personal API token, generated at
  https://app.wingify.com/#/developers/tokens. Never logged.
- `base_url` (optional, string); default `https://app.vwo.com/api/v2`; format `uri`; VWO API base
  URL. Defaults to the public VWO API v2 host (confirmed current and correct against the live
  OpenAPI spec published at developers.wingify.com, which documents the identical
  https://app.wingify.com/api/v2 host family).
- `campaign_type` (optional, string); allowed values `ab`, `multivariate`, `split`, `targeting`,
  `heatmap`, `conversion`, `usability`, `survey`, `analysis`, `analyze-heatmap`, `recording`,
  `form-analysis`, and 5 more.
- `page_size` (optional, integer); default `25`; VWO's documented default is 25.
- `platform` (optional, string); allowed values `website`, `full-stack`, `mobile-app`.
- `start_date` (optional, string); format `date-time`; Optional start_date query parameter passed
  through to the campaigns endpoint verbatim (RFC3339).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `account_id=current`, `base_url=https://app.vwo.com/api/v2`,
`page_size=25`.

Authentication behavior:

- API key authentication in `token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts/{{ config.account_id }}/campaigns`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `campaigns`: GET `/accounts/{{ config.account_id }}/campaigns` - records path `_data`; query
  `platform` from template `{{ config.platform }}`, omitted when absent; `start_date` from template
  `{{ config.start_date }}`, omitted when absent; `type` from template `{{ config.campaign_type }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
  page size 100; computed output fields `created_at`, `id`.
- `campaign_variations`: GET `/accounts/{{ config.account_id }}/campaigns/{{ fanout.id
  }}/variations` - records path `_data`; offset/limit pagination; offset parameter `offset`; limit
  parameter `limit`; page size 100; computed output fields `id`, `is_control`, `is_disabled`,
  `percent_split`; fan-out; ids from request `/accounts/{{ config.account_id }}/campaigns`; id-list
  records path `_data`; id field `id`; id inserted into the request path; stamps `campaign_id`.

## Write actions & risks

Overall write risk: external mutation of VWO campaign configuration, including
starting/pausing/stopping live A/B experiments; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_campaign`: POST `/accounts/{{ config.account_id }}/campaigns` - kind `create`; body type
  `json`; required record fields `type`, `urls`, `primaryUrl`, `goals`; accepted fields `goals`,
  `primaryUrl`, `type`, `urls`; risk: creates a new A/B testing campaign visible to the workspace;
  external mutation, approval required.
- `update_campaign`: PATCH `/accounts/{{ config.account_id }}/campaigns/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`,
  `name`, `percentTraffic`, `status`; risk: updates an existing campaign's configuration (name,
  status, traffic allocation, etc.); can start/pause/stop a live experiment; external mutation,
  approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, non_data_endpoint=2, out_of_scope=9.
