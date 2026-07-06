# Overview

Reads Pendo Engage visitors, accounts, product objects, guides, reports, metadata, exclusion lists,
servers, and feedback options; exposes safe segment, guide, and feedback mutations.

Readable streams: `visitors`, `accounts`, `pages`, `features`, `page_by_id`, `pages_by_ids`,
`feature_by_id`, `features_by_ids`, `tracktypes`, `tracktype_by_id`, `visitor_by_id`,
`visitor_history`, `account_by_id`, `bulkdelete_requests`, `bulkdelete_request`, `segments`,
`segment_by_id`, `segment_status`, `reports`, `report_results_json`, `guides`, `guide_by_id`,
`guide_history`, `guide_order`, `metadata_schema`, `metadata_dependencies`,
`metadata_field_dependencies`, `blacklist`, `blacklist_by_type`, `servers`, `server_by_name`,
`servers_by_flag`, `feedback_options`.

Write actions: `start_segment_visitor_export`, `create_segment`, `update_segment`, `delete_segment`,
`add_segment_visitor`, `remove_segment_visitor`, `patch_segment_visitors`,
`reset_guide_for_visitor`, `reset_all_guides_for_visitor`, `reset_staged_guide`,
`reset_all_staged_guides`, `change_guide_segment`, `change_guide_state`, `create_feedback`,
`update_feedback`, `delete_feedback`.

Service API documentation: https://engageapi.pendo.io/.

## Auth setup

Connection fields:

- `account_id` (optional, string); Pendo account id used by the corresponding detail stream.
- `base_url` (optional, string); default `https://app.pendo.io/api/v1`; format `uri`; Pendo Engage
  API v1 base URL override for tests, regions, or proxies.
- `blacklist_type` (optional, string); Pendo blacklist type used by the corresponding detail stream.
- `bulkdelete_id` (optional, string); Pendo bulkdelete id used by the corresponding detail stream.
- `feature_id` (optional, string); Pendo feature id used by the corresponding detail stream.
- `feature_ids` (optional, string); Pendo feature ids used by the corresponding detail stream.
- `flag_name` (optional, string); Pendo flag name used by the corresponding detail stream.
- `guide_id` (optional, string); Pendo guide id used by the corresponding detail stream.
- `integration_key` (required, secret, string); Pendo integration key, sent as the
  x-pendo-integration-key header. Never logged.
- `limit` (optional, string); default `100`.
- `max_pages` (optional, string); default `0`.
- `metadata_field_name` (optional, string); Pendo metadata field name used by the corresponding
  detail stream.
- `metadata_group` (optional, string); Pendo metadata group used by the corresponding detail stream.
- `metadata_kind` (optional, string); Pendo metadata kind used by the corresponding detail stream.
- `mode` (optional, string).
- `page_id` (optional, string); Pendo page id used by the corresponding detail stream.
- `page_ids` (optional, string); Pendo page ids used by the corresponding detail stream.
- `report_id` (optional, string); Pendo report id used by the corresponding detail stream.
- `segment_id` (optional, string); Pendo segment id used by the corresponding detail stream.
- `server_name` (optional, string); Pendo server name used by the corresponding detail stream.
- `tracktype_id` (optional, string); Pendo tracktype id used by the corresponding detail stream.
- `visitor_history_starttime` (optional, string); Pendo visitor history starttime used by the
  corresponding detail stream.
- `visitor_id` (optional, string); Pendo visitor id used by the corresponding detail stream.

Secret fields are redacted in logs and write previews: `integration_key`.

Default configuration values: `base_url=https://app.pendo.io/api/v1`, `limit=100`, `max_pages=0`.

Authentication behavior:

- API key authentication in `x-pendo-integration-key` using `secrets.integration_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/visitor` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next`.

Pagination by stream: cursor: `visitors`, `accounts`, `pages`, `features`; none: `page_by_id`,
`pages_by_ids`, `feature_by_id`, `features_by_ids`, `tracktypes`, `tracktype_by_id`,
`visitor_by_id`, `visitor_history`, `account_by_id`, `bulkdelete_requests`, `bulkdelete_request`,
`segments`, `segment_by_id`, `segment_status`, `reports`, `report_results_json`, `guides`,
`guide_by_id`, `guide_history`, `guide_order`, `metadata_schema`, `metadata_dependencies`,
`metadata_field_dependencies`, `blacklist`, `blacklist_by_type`, `servers`, `server_by_name`,
`servers_by_flag`, `feedback_options`.

- `visitors`: GET `/visitor` - records path `data`; query `limit`=`{{ config.limit }}`; `page`=`1`;
  cursor pagination; cursor parameter `page`; next token from `next`.
- `accounts`: GET `/account` - records path `data`; query `limit`=`{{ config.limit }}`; `page`=`1`;
  cursor pagination; cursor parameter `page`; next token from `next`.
- `pages`: GET `/page` - records path `data`; query `limit`=`{{ config.limit }}`; `page`=`1`; cursor
  pagination; cursor parameter `page`; next token from `next`.
- `features`: GET `/feature` - records path `data`; query `limit`=`{{ config.limit }}`; `page`=`1`;
  cursor pagination; cursor parameter `page`; next token from `next`.
- `page_by_id`: GET `/page` - single-object response; records at response root; query `id`=`{{
  config.page_id }}`; emits passthrough records.
- `pages_by_ids`: GET `/page` - single-object response; records at response root; query `id`=`{{
  config.page_ids }}`; emits passthrough records.
- `feature_by_id`: GET `/feature` - single-object response; records at response root; query `id`=`{{
  config.feature_id }}`; emits passthrough records.
- `features_by_ids`: GET `/feature` - single-object response; records at response root; query
  `id`=`{{ config.feature_ids }}`; emits passthrough records.
- `tracktypes`: GET `/tracktype` - single-object response; records at response root; emits
  passthrough records.
- `tracktype_by_id`: GET `/tracktype` - single-object response; records at response root; query
  `id`=`{{ config.tracktype_id }}`; emits passthrough records.
- `visitor_by_id`: GET `/visitor/{{ config.visitor_id }}` - single-object response; records at
  response root; emits passthrough records.
- `visitor_history`: GET `/visitor/{{ config.visitor_id }}/history` - single-object response;
  records at response root; query `starttime`=`{{ config.visitor_history_starttime }}`; emits
  passthrough records.
- `account_by_id`: GET `/account/{{ config.account_id }}` - single-object response; records at
  response root; emits passthrough records.
- `bulkdelete_requests`: GET `/bulkdelete` - single-object response; records at response root; emits
  passthrough records.
- `bulkdelete_request`: GET `/bulkdelete/{{ config.bulkdelete_id }}` - single-object response;
  records at response root; emits passthrough records.
- `segments`: GET `/segment` - single-object response; records at response root; emits passthrough
  records.
- `segment_by_id`: GET `/segment/{{ config.segment_id }}` - single-object response; records at
  response root; emits passthrough records.
- `segment_status`: GET `/segment/{{ config.segment_id }}/status` - single-object response; records
  at response root; emits passthrough records.
- `reports`: GET `/report` - single-object response; records at response root; emits passthrough
  records.
- `report_results_json`: GET `/report/{{ config.report_id }}/results.json` - single-object response;
  records at response root; emits passthrough records.
- `guides`: GET `/guide` - single-object response; records at response root; emits passthrough
  records.
- `guide_by_id`: GET `/guide` - single-object response; records at response root; query `id`=`{{
  config.guide_id }}`; emits passthrough records.
- `guide_history`: GET `/guide/{{ config.guide_id }}/history` - single-object response; records at
  response root; emits passthrough records.
- `guide_order`: GET `/guide/order` - single-object response; records at response root; emits
  passthrough records.
- `metadata_schema`: GET `/metadata/schema/{{ config.metadata_kind }}` - single-object response;
  records at response root; emits passthrough records.
- `metadata_dependencies`: GET `/metadata/dependencies` - single-object response; records at
  response root; emits passthrough records.
- `metadata_field_dependencies`: GET `/metadata/dependencies/{{ config.metadata_kind }}/{{
  config.metadata_group }}/{{ config.metadata_field_name }}` - single-object response; records at
  response root; emits passthrough records.
- `blacklist`: GET `/blacklist` - single-object response; records at response root; emits
  passthrough records.
- `blacklist_by_type`: GET `/blacklist/type/{{ config.blacklist_type }}` - records path `results`;
  emits passthrough records.
- `servers`: GET `/servername` - single-object response; records at response root; emits passthrough
  records.
- `server_by_name`: GET `/servername/{{ config.server_name }}` - single-object response; records at
  response root; emits passthrough records.
- `servers_by_flag`: GET `/servername/flag/{{ config.flag_name }}` - single-object response; records
  at response root; emits passthrough records.
- `feedback_options`: GET `/feedback/options` - single-object response; records at response root;
  emits passthrough records.

## Write actions & risks

Overall write risk: mutates Pendo segments, guide state/seen status, and Listen feedback records.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `start_segment_visitor_export`: POST `/segment/{{ record.segmentId }}/visitors` - kind `create`;
  body type `none`; path fields `segmentId`; required record fields `segmentId`; accepted fields
  `segmentId`; risk: starts an asynchronous Pendo segment visitor export job; approval required.
- `create_segment`: POST `/segment/upload` - kind `create`; body type `json`; required record fields
  `name`, `visitors`; accepted fields `name`, `visitors`; risk: creates a shared Pendo segment from
  visitor ids; approval required.
- `update_segment`: PUT `/segment/{{ record.segmentId }}` - kind `update`; body type `json`; path
  fields `segmentId`; required record fields `segmentId`, `name`, `visitors`; accepted fields
  `name`, `segmentId`, `visitors`; risk: replaces the visitor membership for a Pendo segment;
  approval required.
- `delete_segment`: DELETE `/segment/{{ record.segmentId }}` - kind `delete`; body type `none`; path
  fields `segmentId`; required record fields `segmentId`; accepted fields `segmentId`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes a Pendo
  segment; destructive external mutation.
- `add_segment_visitor`: PUT `/segment/{{ record.segmentId }}/visitor/{{ record.visitorId }}` - kind
  `update`; body type `none`; path fields `segmentId`, `visitorId`; required record fields
  `segmentId`, `visitorId`; accepted fields `segmentId`, `visitorId`; risk: adds a visitor to a
  Pendo segment; approval required.
- `remove_segment_visitor`: DELETE `/segment/{{ record.segmentId }}/visitor/{{ record.visitorId }}`
  - kind `delete`; body type `none`; path fields `segmentId`, `visitorId`; required record fields
  `segmentId`, `visitorId`; accepted fields `segmentId`, `visitorId`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: removes a visitor from a Pendo
  segment; destructive membership mutation.
- `patch_segment_visitors`: PATCH `/segment/{{ record.segmentId }}/visitor` - kind `update`; body
  type `json`; path fields `segmentId`; required record fields `segmentId`, `patch`; accepted fields
  `patch`, `segmentId`; risk: adds/removes a small batch of visitors for a Pendo segment; approval
  required.
- `reset_guide_for_visitor`: POST `/guide/{{ record.guideId }}/visitor/{{ record.visitorId }}/reset`
  - kind `update`; body type `none`; path fields `guideId`, `visitorId`; required record fields
  `guideId`, `visitorId`; accepted fields `guideId`, `visitorId`; risk: resets guide-seen state for
  one visitor; approval required.
- `reset_all_guides_for_visitor`: POST `/guide/all/visitor/{{ record.visitorId }}/reset` - kind
  `update`; body type `none`; path fields `visitorId`; required record fields `visitorId`; accepted
  fields `visitorId`; risk: resets guide-seen state for one visitor across all guides; approval
  required.
- `reset_staged_guide`: POST `/guide/{{ record.guideId }}/reset` - kind `update`; body type `none`;
  path fields `guideId`; required record fields `guideId`; accepted fields `guideId`; risk: resets
  one staged guide; approval required.
- `reset_all_staged_guides`: POST `/guide/staged/reset` - kind `update`; body type `none`; risk:
  resets all staged guides in the subscription; approval required.
- `change_guide_segment`: PUT `/guide/{{ record.guideId }}/segment` - kind `update`; body type
  `json`; path fields `guideId`; required record fields `guideId`, `segmentId`; accepted fields
  `guideId`, `segmentId`; risk: changes the segment assigned to a Pendo guide; approval required.
- `change_guide_state`: PUT `/guide/{{ record.guideId }}/state` - kind `update`; body type `json`;
  path fields `guideId`; required record fields `guideId`, `state`; accepted fields `guideId`,
  `state`; risk: changes a Pendo guide state such as public, staged, disabled, or draft; approval
  required.
- `create_feedback`: POST `/feedback` - kind `create`; body type `json`; required record fields
  `accountId`, `visitorId`, `title`; accepted fields `accountId`, `title`, `visitorId`; risk:
  creates a Pendo Listen feedback item; approval required.
- `update_feedback`: PATCH `/feedback/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `title`; risk:
  updates a Pendo Listen feedback item; approval required.
- `delete_feedback`: DELETE `/feedback` - kind `delete`; body type `json`; body fields `ids`;
  required record fields `ids`; accepted fields `ids`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: deletes Pendo Listen feedback items by id; destructive
  external mutation.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 33 stream-backed endpoint group(s), 16 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=5, destructive_admin=3, non_data_endpoint=2, out_of_scope=9,
  requires_elevated_scope=5.
