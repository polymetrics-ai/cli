# Overview

Reads Freshchat account, user, conversation, agent, group, channel, role, outbound, report, metrics,
and business-hours data through the Freshchat v2 REST API; writes Freshchat users, conversations,
agents, outbound messages, reports, and CSAT ratings.

Readable streams: `account_configuration`, `agents`, `agent_details`, `agent_statuses`, `users`,
`user_details`, `user_conversations`, `conversation_detail`, `conversation_messages`,
`conversation_fields`, `groups`, `channels`, `roles`, `outbound_messages`, `report_status`,
`historical_metrics`, `instant_metrics`, `business_hours_status`.

Write actions: `create_user`, `update_user`, `delete_user`, `create_conversation`,
`update_conversation`, `send_conversation_message`, `create_agent`, `update_agent`,
`update_agent_status`, `delete_agent`, `send_outbound_whatsapp_message`, `extract_report`,
`create_csat_rating`.

Service API documentation: https://developers.freshchat.com/api/.

## Auth setup

Connection fields:

- `agent_id` (optional, string); Freshchat agent id for agent_details.
- `api_key` (required, secret, string); Freshchat API key. Sent as Authorization: Bearer <api_key>;
  never logged.
- `base_url` (required, string); format `uri`; Freshchat API base URL, e.g.
  https://<account_name>.freshchat.com/v2.
- `business_hours_group_id` (optional, string); Freshchat group id for business_hours_status.
- `conversation_id` (optional, string); Freshchat conversation id for conversation_detail and
  conversation_messages streams.
- `metrics_aggregator` (optional, string); Optional historical metrics aggregator query value.
- `metrics_end` (optional, string); Historical metrics end timestamp.
- `metrics_filter_by` (optional, string); Optional metrics filter_by query value.
- `metrics_group_by` (optional, string); Optional metrics group_by query value.
- `metrics_interval` (optional, string); Optional historical metrics interval query value.
- `metrics_metric` (optional, string); Metric name for Freshchat historical_metrics and
  instant_metrics streams.
- `metrics_start` (optional, string); Historical metrics start timestamp.
- `metrics_summary` (optional, string); Optional instant metrics summary query value.
- `mode` (optional, string).
- `outbound_request_id` (optional, string); Optional request id filter for outbound_messages.
- `report_id` (optional, string); Freshchat report id for report_status.
- `report_status` (optional, string); Optional status filter for report_status.
- `user_id` (optional, string); Freshchat user id for user_details and user_conversations streams.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/agents` with query `items_per_page`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `items_per_page`;
starts at 1; page size 50.

Pagination by stream: none: `account_configuration`, `agent_details`, `agent_statuses`,
`user_details`, `conversation_detail`, `conversation_fields`, `report_status`, `historical_metrics`,
`instant_metrics`, `business_hours_status`; page_number: `agents`, `users`, `user_conversations`,
`conversation_messages`, `groups`, `channels`, `roles`, `outbound_messages`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `account_configuration`: GET `/accounts/configuration` - single-object response; records path `.`.
- `agents`: GET `/agents` - records path `agents`; page-number pagination; page parameter `page`;
  size parameter `items_per_page`; starts at 1; page size 50; incremental cursor `updated_time`;
  formatted as `rfc3339`.
- `agent_details`: GET `/agents/{{ config.agent_id }}` - single-object response; records path `.`.
- `agent_statuses`: GET `/agents/status` - records path `statuses`.
- `users`: GET `/users` - records path `users`; page-number pagination; page parameter `page`; size
  parameter `items_per_page`; starts at 1; page size 50; incremental cursor `updated_time`;
  formatted as `rfc3339`.
- `user_details`: GET `/users/{{ config.user_id }}` - single-object response; records path `.`.
- `user_conversations`: GET `/users/{{ config.user_id }}/conversations` - records path
  `conversations`; page-number pagination; page parameter `page`; size parameter `items_per_page`;
  starts at 1; page size 50; computed output fields `user_id`.
- `conversation_detail`: GET `/conversations/{{ config.conversation_id }}` - single-object response;
  records path `.`.
- `conversation_messages`: GET `/conversations/{{ config.conversation_id }}/messages` - records path
  `messages`; page-number pagination; page parameter `page`; size parameter `items_per_page`; starts
  at 1; page size 50; computed output fields `conversation_id`.
- `conversation_fields`: GET `/conversations/fields` - records path `.`.
- `groups`: GET `/groups` - records path `groups`; page-number pagination; page parameter `page`;
  size parameter `items_per_page`; starts at 1; page size 50.
- `channels`: GET `/channels` - records path `channels`; page-number pagination; page parameter
  `page`; size parameter `items_per_page`; starts at 1; page size 50; incremental cursor
  `updated_time`; formatted as `rfc3339`.
- `roles`: GET `/roles` - records path `roles`; page-number pagination; page parameter `page`; size
  parameter `items_per_page`; starts at 1; page size 50.
- `outbound_messages`: GET `/outbound-messages` - records path `outbound_messages`; query
  `request_id` from template `{{ config.outbound_request_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `items_per_page`; starts at 1; page size 50.
- `report_status`: GET `/reports/raw/{{ config.report_id }}` - single-object response; records path
  `.`; query `status` from template `{{ config.report_status }}`, omitted when absent.
- `historical_metrics`: GET `/metrics/historical` - single-object response; records path `.`; query
  `aggregator` from template `{{ config.metrics_aggregator }}`, omitted when absent; `end`=`{{
  config.metrics_end }}`; `filter_by` from template `{{ config.metrics_filter_by }}`, omitted when
  absent; `group_by` from template `{{ config.metrics_group_by }}`, omitted when absent; `interval`
  from template `{{ config.metrics_interval }}`, omitted when absent; `metric`=`{{
  config.metrics_metric }}`; `start`=`{{ config.metrics_start }}`; computed output fields
  `metric_type`; emits passthrough records.
- `instant_metrics`: GET `/metrics/instant` - single-object response; records path `.`; query
  `filter_by` from template `{{ config.metrics_filter_by }}`, omitted when absent; `group_by` from
  template `{{ config.metrics_group_by }}`, omitted when absent; `metric`=`{{ config.metrics_metric
  }}`; `summary` from template `{{ config.metrics_summary }}`, omitted when absent; computed output
  fields `metric_type`; emits passthrough records.
- `business_hours_status`: GET `/business-hours/within-bh` - single-object response; records path
  `.`; query `group_id`=`{{ config.business_hours_group_id }}`; computed output fields `group_id`;
  emits passthrough records.

## Write actions & risks

Overall write risk: creates, updates, or deletes Freshchat users, conversations, agents, outbound
messages, reports, and CSAT ratings.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/users` - kind `create`; body type `json`; accepted fields `email`,
  `first_name`, `last_name`, `phone`, `properties`, `reference_id`; risk: creates a Freshchat
  user/contact visible to agents.
- `update_user`: PUT `/users/{{ record.user_id }}` - kind `update`; body type `json`; path fields
  `user_id`; required record fields `user_id`; accepted fields `email`, `first_name`, `last_name`,
  `phone`, `properties`, `reference_id`, `user_id`; risk: updates an existing Freshchat
  user/contact.
- `delete_user`: DELETE `/users/{{ record.user_id }}` - kind `delete`; body type `none`; path fields
  `user_id`; required record fields `user_id`; accepted fields `user_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes a Freshchat user/contact.
- `create_conversation`: POST `/conversations` - kind `create`; body type `json`; accepted fields
  `channel_id`, `messages`, `properties`, `user_id`; risk: creates a Freshchat conversation.
- `update_conversation`: PUT `/conversations/{{ record.conversation_id }}` - kind `update`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `assigned_agent_id`, `assigned_group_id`, `conversation_id`, `priority`, `properties`,
  `status`; risk: updates routing, status, or properties on an existing Freshchat conversation.
- `send_conversation_message`: POST `/conversations/{{ record.conversation_id }}/messages` - kind
  `create`; body type `json`; path fields `conversation_id`; required record fields
  `conversation_id`; accepted fields `actor_id`, `actor_type`, `conversation_id`, `message_parts`;
  risk: sends a message into an existing Freshchat conversation.
- `create_agent`: POST `/agents` - kind `create`; body type `json`; required record fields `email`;
  accepted fields `email`, `first_name`, `groups`, `last_name`, `role_id`; risk: creates a Freshchat
  agent account.
- `update_agent`: PUT `/agents/{{ record.agent_id }}` - kind `update`; body type `json`; path fields
  `agent_id`; required record fields `agent_id`; accepted fields `agent_id`, `first_name`, `groups`,
  `is_deactivated`, `last_name`, `role_id`; risk: updates an existing Freshchat agent.
- `update_agent_status`: PATCH `/agents/{{ record.agent_id }}` - kind `update`; body type `json`;
  path fields `agent_id`; required record fields `agent_id`, `status`; accepted fields `agent_id`,
  `status`; risk: updates an agent's Freshchat availability status.
- `delete_agent`: DELETE `/agents/{{ record.agent_id }}` - kind `delete`; body type `none`; path
  fields `agent_id`; required record fields `agent_id`; accepted fields `agent_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a Freshchat agent.
- `send_outbound_whatsapp_message`: POST `/outbound-messages/whatsapp` - kind `create`; body type
  `json`; accepted fields `from`, `provider`, `template`, `to`; risk: sends an outbound WhatsApp
  message through Freshchat.
- `extract_report`: POST `/reports/raw` - kind `create`; body type `json`; required record fields
  `start`, `end`, `event`, `format`; accepted fields `end`, `event`, `format`, `start`; risk:
  requests generation of a Freshchat raw report extract.
- `create_csat_rating`: POST `/csat/{{ record.conversation_id }}` - kind `create`; body type `json`;
  path fields `conversation_id`; required record fields `conversation_id`; accepted fields
  `comment`, `conversation_id`, `rating`; risk: creates a CSAT rating for a Freshchat conversation.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 18 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, out_of_scope=1.
