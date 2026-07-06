# Overview

Reads Postmark server-token API resources including messages, bounces, templates, message streams,
stats, webhooks, suppressions, and inbound rules; exposes server-token write actions for sends and
resource mutations.

Readable streams: `outbound_messages`, `inbound_messages`, `current_server`, `bulk_email_status`,
`delivery_stats`, `bounces`, `bounce`, `bounce_dump`, `templates`, `template`, `message_streams`,
`message_stream`, `outbound_message_details`, `outbound_message_dump`, `inbound_message_details`,
`outbound_message_opens`, `outbound_message_opens_by_message`, `outbound_message_clicks`,
`outbound_message_clicks_by_message`, `stats_outbound`, `stats_outbound_sends`,
`stats_outbound_bounces`, `stats_outbound_spam`, `stats_outbound_tracked`, `stats_outbound_opens`,
`stats_outbound_open_platforms`, `stats_outbound_email_clients`, `stats_outbound_clicks`,
`stats_outbound_click_browser_families`, `stats_outbound_click_platforms`,
`stats_outbound_click_location`, `inbound_rule_triggers`, `webhooks`, `webhook`, `suppressions`.

Write actions: `send_email`, `send_bulk_email`, `send_email_with_template`, `edit_current_server`,
`activate_bounce`, `create_template`, `edit_template`, `delete_template`, `validate_template`,
`create_message_stream`, `edit_message_stream`, `archive_message_stream`,
`unarchive_message_stream`, `bypass_inbound_message`, `retry_inbound_message`,
`create_inbound_rule_trigger`, `delete_inbound_rule_trigger`, `create_webhook`, `edit_webhook`,
`delete_webhook`, `create_suppression`, `delete_suppression`.

Service API documentation: https://postmarkapp.com/developer.

## Auth setup

Connection fields:

- `X-Postmark-Server-Token` (required, secret, string); Postmark server token, sent as the
  X-Postmark-Server-Token header for server-level endpoints. Never logged.
- `base_url` (optional, string); default `https://api.postmarkapp.com`; format `uri`; Postmark API
  base URL override for tests or proxies.
- `bounce_id` (optional, string); Postmark bounce id used by the corresponding detail stream.
- `bulk_request_id` (optional, string); Postmark bulk request id used by the corresponding detail
  stream.
- `message_id` (optional, string); Postmark message id used by the corresponding detail stream.
- `message_stream_id` (optional, string); Postmark message stream id used by the corresponding
  detail stream.
- `mode` (optional, string).
- `template_id_or_alias` (optional, string); Postmark template id or alias used by the corresponding
  detail stream.
- `webhook_id` (optional, string); Postmark webhook id used by the corresponding detail stream.

Secret fields are redacted in logs and write previews: `X-Postmark-Server-Token`.

Default configuration values: `base_url=https://api.postmarkapp.com`.

Authentication behavior:

- API key authentication in `X-Postmark-Server-Token` using `secrets.X-Postmark-Server-Token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/messages/outbound` with query `count`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `count`;
page size 100.

Pagination by stream: none: `current_server`, `bulk_email_status`, `delivery_stats`, `bounce`,
`bounce_dump`, `template`, `message_stream`, `outbound_message_details`, `outbound_message_dump`,
`inbound_message_details`, `stats_outbound`, `stats_outbound_sends`, `stats_outbound_bounces`,
`stats_outbound_spam`, `stats_outbound_tracked`, `stats_outbound_opens`,
`stats_outbound_open_platforms`, `stats_outbound_email_clients`, `stats_outbound_clicks`,
`stats_outbound_click_browser_families`, `stats_outbound_click_platforms`,
`stats_outbound_click_location`, `webhook`; offset_limit: `outbound_messages`, `inbound_messages`,
`bounces`, `templates`, `message_streams`, `outbound_message_opens`,
`outbound_message_opens_by_message`, `outbound_message_clicks`,
`outbound_message_clicks_by_message`, `inbound_rule_triggers`, `webhooks`, `suppressions`.

- `outbound_messages`: GET `/messages/outbound` - records path `Messages`; offset/limit pagination;
  offset parameter `offset`; limit parameter `count`; page size 100; computed output fields `from`,
  `id`, `received_at`, `status`, `subject`, `to`.
- `inbound_messages`: GET `/messages/inbound` - records path `InboundMessages`; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `from`, `id`, `status`, `subject`, `to`.
- `current_server`: GET `/server` - single-object response; records at response root; emits
  passthrough records.
- `bulk_email_status`: GET `/email/bulk/{{ config.bulk_request_id }}` - single-object response;
  records at response root; emits passthrough records.
- `delivery_stats`: GET `/deliverystats` - single-object response; records at response root; emits
  passthrough records.
- `bounces`: GET `/bounces` - records path `Bounces`; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; emits passthrough records.
- `bounce`: GET `/bounces/{{ config.bounce_id }}` - single-object response; records at response
  root; emits passthrough records.
- `bounce_dump`: GET `/bounces/{{ config.bounce_id }}/dump` - single-object response; records at
  response root; emits passthrough records.
- `templates`: GET `/templates` - records path `Templates`; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; emits passthrough records.
- `template`: GET `/templates/{{ config.template_id_or_alias }}` - single-object response; records
  at response root; emits passthrough records.
- `message_streams`: GET `/message-streams` - records path `MessageStreams`; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; emits passthrough
  records.
- `message_stream`: GET `/message-streams/{{ config.message_stream_id }}` - single-object response;
  records at response root; emits passthrough records.
- `outbound_message_details`: GET `/messages/outbound/{{ config.message_id }}/details` -
  single-object response; records at response root; emits passthrough records.
- `outbound_message_dump`: GET `/messages/outbound/{{ config.message_id }}/dump` - single-object
  response; records at response root; emits passthrough records.
- `inbound_message_details`: GET `/messages/inbound/{{ config.message_id }}/details` - single-object
  response; records at response root; emits passthrough records.
- `outbound_message_opens`: GET `/messages/outbound/opens` - records path `Opens`; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; emits passthrough
  records.
- `outbound_message_opens_by_message`: GET `/messages/outbound/opens/{{ config.message_id }}` -
  records path `Opens`; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; emits passthrough records.
- `outbound_message_clicks`: GET `/messages/outbound/clicks` - records path `Clicks`; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; emits passthrough
  records.
- `outbound_message_clicks_by_message`: GET `/messages/outbound/clicks/{{ config.message_id }}` -
  records path `Clicks`; offset/limit pagination; offset parameter `offset`; limit parameter
  `count`; page size 100; emits passthrough records.
- `stats_outbound`: GET `/stats/outbound` - single-object response; records at response root; emits
  passthrough records.
- `stats_outbound_sends`: GET `/stats/outbound/sends` - single-object response; records at response
  root; emits passthrough records.
- `stats_outbound_bounces`: GET `/stats/outbound/bounces` - single-object response; records at
  response root; emits passthrough records.
- `stats_outbound_spam`: GET `/stats/outbound/spam` - single-object response; records at response
  root; emits passthrough records.
- `stats_outbound_tracked`: GET `/stats/outbound/tracked` - single-object response; records at
  response root; emits passthrough records.
- `stats_outbound_opens`: GET `/stats/outbound/opens` - single-object response; records at response
  root; emits passthrough records.
- `stats_outbound_open_platforms`: GET `/stats/outbound/opens/platforms` - single-object response;
  records at response root; emits passthrough records.
- `stats_outbound_email_clients`: GET `/stats/outbound/opens/emailclients` - single-object response;
  records at response root; emits passthrough records.
- `stats_outbound_clicks`: GET `/stats/outbound/clicks` - single-object response; records at
  response root; emits passthrough records.
- `stats_outbound_click_browser_families`: GET `/stats/outbound/clicks/browserfamilies` -
  single-object response; records at response root; emits passthrough records.
- `stats_outbound_click_platforms`: GET `/stats/outbound/clicks/platforms` - single-object response;
  records at response root; emits passthrough records.
- `stats_outbound_click_location`: GET `/stats/outbound/clicks/location` - single-object response;
  records at response root; emits passthrough records.
- `inbound_rule_triggers`: GET `/triggers/inboundrules` - records path `InboundRules`; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; emits passthrough
  records.
- `webhooks`: GET `/webhooks` - records path `Webhooks`; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; emits passthrough records.
- `webhook`: GET `/webhooks/{{ config.webhook_id }}` - single-object response; records at response
  root; emits passthrough records.
- `suppressions`: GET `/message-streams/{{ config.message_stream_id }}/suppressions/dump` - records
  path `Suppressions`; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; emits passthrough records.

## Write actions & risks

Overall write risk: sends emails and mutates Postmark templates, message streams, server settings,
webhooks, inbound rules, suppressions, and inbound message processing state.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `send_email`: POST `/email` - kind `create`; body type `json`; required record fields `From`,
  `To`, `Subject`; accepted fields `From`, `HtmlBody`, `MessageStream`, `Subject`, `TextBody`, `To`;
  risk: sends a live Postmark email; approval required.
- `send_bulk_email`: POST `/email/bulk` - kind `create`; body type `json`; required record fields
  `From`, `Subject`, `Messages`; accepted fields `From`, `MessageStream`, `Messages`, `Subject`;
  risk: submits a live Postmark bulk email request; approval required.
- `send_email_with_template`: POST `/email/withTemplate` - kind `create`; body type `json`; required
  record fields `From`, `To`; accepted fields `From`, `TemplateAlias`, `TemplateId`,
  `TemplateModel`, `To`; risk: sends a live Postmark template email; approval required.
- `edit_current_server`: PUT `/server` - kind `update`; body type `json`; accepted fields `Color`,
  `Name`; risk: mutates the current Postmark server settings; approval required.
- `activate_bounce`: PUT `/bounces/{{ record.bounce_id }}/activate` - kind `update`; body type
  `none`; path fields `bounce_id`; required record fields `bounce_id`; accepted fields `bounce_id`;
  risk: reactivates a bounced email address in Postmark; approval required.
- `create_template`: POST `/templates` - kind `create`; body type `json`; required record fields
  `Name`; accepted fields `Alias`, `HtmlBody`, `Name`, `Subject`, `TextBody`; risk: creates a
  Postmark template; approval required.
- `edit_template`: PUT `/templates/{{ record.template_id_or_alias }}` - kind `update`; body type
  `json`; path fields `template_id_or_alias`; required record fields `template_id_or_alias`;
  accepted fields `Alias`, `HtmlBody`, `Name`, `Subject`, `TextBody`, `template_id_or_alias`; risk:
  updates a Postmark template; approval required.
- `delete_template`: DELETE `/templates/{{ record.template_id_or_alias }}` - kind `delete`; body
  type `none`; path fields `template_id_or_alias`; required record fields `template_id_or_alias`;
  accepted fields `template_id_or_alias`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes a Postmark template; destructive external mutation.
- `validate_template`: POST `/templates/validate` - kind `custom`; body type `json`; accepted fields
  `HtmlBody`, `Subject`, `TextBody`; risk: validates Postmark template content; no persistent
  mutation expected but still invokes the external API.
- `create_message_stream`: POST `/message-streams` - kind `create`; body type `json`; required
  record fields `ID`, `Name`, `MessageStreamType`; accepted fields `ID`, `MessageStreamType`,
  `Name`; risk: creates a Postmark message stream; approval required.
- `edit_message_stream`: PATCH `/message-streams/{{ record.message_stream_id }}` - kind `update`;
  body type `json`; path fields `message_stream_id`; required record fields `message_stream_id`;
  accepted fields `Name`, `message_stream_id`; risk: updates a Postmark message stream; approval
  required.
- `archive_message_stream`: POST `/message-streams/{{ record.message_stream_id }}/archive` - kind
  `update`; body type `none`; path fields `message_stream_id`; required record fields
  `message_stream_id`; accepted fields `message_stream_id`; risk: archives a Postmark message
  stream; approval required.
- `unarchive_message_stream`: POST `/message-streams/{{ record.message_stream_id }}/unarchive` -
  kind `update`; body type `none`; path fields `message_stream_id`; required record fields
  `message_stream_id`; accepted fields `message_stream_id`; risk: unarchives a Postmark message
  stream; approval required.
- `bypass_inbound_message`: PUT `/messages/inbound/{{ record.message_id }}/bypass` - kind `update`;
  body type `none`; path fields `message_id`; required record fields `message_id`; accepted fields
  `message_id`; risk: bypasses inbound blocking for one Postmark message; approval required.
- `retry_inbound_message`: PUT `/messages/inbound/{{ record.message_id }}/retry` - kind `update`;
  body type `none`; path fields `message_id`; required record fields `message_id`; accepted fields
  `message_id`; risk: retries processing for one inbound Postmark message; approval required.
- `create_inbound_rule_trigger`: POST `/triggers/inboundrules` - kind `create`; body type `json`;
  required record fields `Rule`; accepted fields `Rule`; risk: creates an inbound rule trigger;
  approval required.
- `delete_inbound_rule_trigger`: DELETE `/triggers/inboundrules/{{ record.trigger_id }}` - kind
  `delete`; body type `none`; path fields `trigger_id`; required record fields `trigger_id`;
  accepted fields `trigger_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes an inbound rule trigger; destructive external mutation.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `Url`; accepted fields `MessageStream`, `Triggers`, `Url`; risk: creates a Postmark webhook
  endpoint; approval required.
- `edit_webhook`: PUT `/webhooks/{{ record.webhook_id }}` - kind `update`; body type `json`; path
  fields `webhook_id`; required record fields `webhook_id`; accepted fields `Triggers`, `Url`,
  `webhook_id`; risk: updates a Postmark webhook endpoint; approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.webhook_id }}` - kind `delete`; body type `none`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: deletes a
  Postmark webhook endpoint; destructive external mutation.
- `create_suppression`: POST `/message-streams/{{ record.message_stream_id }}/suppressions` - kind
  `create`; body type `json`; path fields `message_stream_id`; required record fields
  `message_stream_id`, `Suppressions`; accepted fields `Suppressions`, `message_stream_id`; risk:
  adds one or more suppressions to a Postmark message stream; approval required.
- `delete_suppression`: POST `/message-streams/{{ record.message_stream_id }}/suppressions/delete` -
  kind `delete`; body type `json`; path fields `message_stream_id`; required record fields
  `message_stream_id`, `Suppressions`; accepted fields `Suppressions`, `message_stream_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: removes one or more
  suppressions from a Postmark message stream; approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 35 stream-backed endpoint group(s), 22 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=27.
