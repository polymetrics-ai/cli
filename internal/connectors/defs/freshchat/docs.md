# Overview

Freshchat (`freshchat`) reads account, user, conversation, message, agent, group, channel/topic,
role, outbound-message, report, metrics, and business-hours data through the official Freshchat v2
REST API. It writes Freshchat users, conversations, messages, agents, outbound WhatsApp messages,
raw report extract requests, and CSAT ratings through named reverse-ETL actions.

Official source reviewed for this bundle: https://developers.freshchat.com/api/ with ETag
`W/"26e4fd8b1fe01578eae1dbaff6b69224"`.

Operation ledger: 34 documented endpoint rows, 31 executable rows, 3 blocked/planned rows, 0 exclusions.

Readable streams: `account_configuration`, `agents`, `agent_details`, `agent_statuses`, `users`, `user_details`, `user_conversations`, `conversation_detail`, `conversation_messages`, `conversation_fields`, `groups`, `channels`, `roles`, `outbound_messages`, `report_status`, `historical_metrics`, `instant_metrics`, `business_hours_status`.

Write actions: `create_user`, `update_user`, `delete_user`, `create_conversation`, `update_conversation`, `send_conversation_message`, `create_agent`, `update_agent`, `update_agent_status`, `delete_agent`, `send_outbound_whatsapp_message`, `extract_report`, `create_csat_rating`.

## Auth setup

Freshchat uses Bearer token authentication. Store the Freshchat API key as the secret field
`api_key`; never pass or record the secret in prompt text, CLI flags, fixtures, logs, or docs.

Connection fields:

- `api_key` (required, secret, string) — Freshchat API key. Sent as Authorization: Bearer <api_key>; never logged.
- `base_url` (required, string; format `uri`) — Freshchat API base URL, e.g. https://<account_name>.freshchat.com/v2. This bundle requires the full base URL directly: legacy derives it from an account_name config value (https://<account_name>.freshchat.com/v2), but that cross-key derivation has no engine mechanism today (see docs.md Known limits) — spec.json's default materialization only supports a fixed literal, never one derived from another config value.
- `user_id` (optional, string) — Freshchat user id for user_details and user_conversations streams.
- `conversation_id` (optional, string) — Freshchat conversation id for conversation_detail and conversation_messages streams.
- `agent_id` (optional, string) — Freshchat agent id for agent_details.
- `report_id` (optional, string) — Freshchat report id for report_status.
- `report_status` (optional, string) — Optional status filter for report_status.
- `business_hours_group_id` (optional, string) — Freshchat group id for business_hours_status.
- `outbound_request_id` (optional, string) — Optional request id filter for outbound_messages.
- `metrics_metric` (optional, string) — Metric name for Freshchat historical_metrics and instant_metrics streams.
- `metrics_start` (optional, string) — Historical metrics start timestamp.
- `metrics_end` (optional, string) — Historical metrics end timestamp.
- `metrics_group_by` (optional, string) — Optional metrics group_by query value.
- `metrics_filter_by` (optional, string) — Optional metrics filter_by query value.
- `metrics_aggregator` (optional, string) — Optional historical metrics aggregator query value.
- `metrics_interval` (optional, string) — Optional historical metrics interval query value.
- `metrics_summary` (optional, string) — Optional instant metrics summary query value.
- `agents_is_deactivated` (optional, string; values `true`, `false`) — Optional /agents filter for is_deactivated; values true or false.
- `agents_groups` (optional, string) — Optional /agents filter for a Freshchat group id.
- `agents_availability_status` (optional, string; values `AVAILABLE`, `UNAVAILABLE`) — Optional /agents filter for availability_status.
- `agents_sort_order` (optional, string; values `asc`, `desc`) — Optional /agents pagination sort_order value.
- `agents_sort_by` (optional, string) — Optional /agents sort_by value.
- `users_first_name` (optional, string) — Optional /users first_name filter. Freshchat requires at least one user filter for live list responses.
- `users_last_name` (optional, string) — Optional /users last_name filter.
- `users_email` (optional, string) — Optional /users email filter.
- `users_reference_id` (optional, string) — Optional /users reference_id filter.
- `users_phone_no` (optional, string) — Optional /users phone_no filter.
- `users_created_from` (optional, string; format `date-time`) — Optional /users created_from UTC timestamp filter.
- `users_created_to` (optional, string; format `date-time`) — Optional /users created_to UTC timestamp filter.
- `users_updated_from` (optional, string; format `date-time`) — Optional /users updated_from UTC timestamp filter.
- `users_updated_to` (optional, string; format `date-time`) — Optional /users updated_to UTC timestamp filter.
- `conversation_messages_from_time` (optional, string; format `date-time`) — Optional /conversations/{conversation_id}/messages from_time UTC timestamp filter.
- `groups_sort_order` (optional, string; values `asc`, `desc`) — Optional /groups sort_order value.
- `groups_sort_by` (optional, string) — Optional /groups sort_by value.
- `channels_locale` (optional, string) — Optional /channels locale filter in ISO-639 format.
- `mode` (optional, string) — Runtime mode: live (default) or fixture for credential-free conformance.

Secret fields redacted by the connector runtime: `api_key`.

Connection checks call `GET /agents?page=1&items_per_page=1` against the configured `base_url`.

`base_url` is required because Freshchat account subdomains are tenant-specific; the current engine
can materialize literal defaults but cannot derive `https://<account>.freshchat.com/v2` from another
config key.

## Streams notes

Default list pagination is page-number pagination with `page`, `items_per_page`, start page `1`, and
page size `50`. Freshchat documents filter query parameters for several list operations; this bundle
models them as optional, named config/CLI metadata fields rather than a raw query escape hatch.

Freshchat's `GET /users` documentation requires at least one user search filter for successful live
responses. Set one of the `users_*` filter config fields for live user-list reads; fixtures use only
sanitized replay data and do not certify unfiltered provider behavior.

- `account_configuration`: GET `/accounts/configuration`; records path `.`; single object; pagination `none`.
- `agents`: GET `/agents`; records path `agents`; default page-number pagination; incremental cursor `updated_time`; query keys `is_deactivated`, `groups`, `availability_status`, `sort_order`, `sort_by`.
- `agent_details`: GET `/agents/{{ config.agent_id }}`; records path `.`; single object; pagination `none`.
- `agent_statuses`: GET `/agents/status`; records path `statuses`; pagination `none`.
- `users`: GET `/users`; records path `users`; default page-number pagination; incremental cursor `updated_time`; query keys `first_name`, `last_name`, `email`, `reference_id`, `phone_no`, `created_from`, `created_to`, `updated_from`, `updated_to`.
- `user_details`: GET `/users/{{ config.user_id }}`; records path `.`; single object; pagination `none`.
- `user_conversations`: GET `/users/{{ config.user_id }}/conversations`; records path `conversations`; default page-number pagination; computed fields `user_id`.
- `conversation_detail`: GET `/conversations/{{ config.conversation_id }}`; records path `.`; single object; pagination `none`.
- `conversation_messages`: GET `/conversations/{{ config.conversation_id }}/messages`; records path `messages`; default page-number pagination; query keys `from_time`; computed fields `conversation_id`.
- `conversation_fields`: GET `/conversations/fields`; records path `.`; pagination `none`.
- `groups`: GET `/groups`; records path `groups`; default page-number pagination; query keys `sort_order`, `sort_by`.
- `channels`: GET `/channels`; records path `channels`; default page-number pagination; incremental cursor `updated_time`; query keys `locale`.
- `roles`: GET `/roles`; records path `roles`; default page-number pagination.
- `outbound_messages`: GET `/outbound-messages`; records path `outbound_messages`; default page-number pagination; query keys `request_id`.
- `report_status`: GET `/reports/raw/{{ config.report_id }}`; records path `.`; single object; pagination `none`; query keys `status`.
- `historical_metrics`: GET `/metrics/historical`; records path `.`; single object; pagination `none`; query keys `metric`, `start`, `end`, `group_by`, `filter_by`, `aggregator`, `interval`; computed fields `metric_type`; passthrough projection.
- `instant_metrics`: GET `/metrics/instant`; records path `.`; single object; pagination `none`; query keys `metric`, `group_by`, `filter_by`, `summary`; computed fields `metric_type`; passthrough projection.
- `business_hours_status`: GET `/business-hours/within-bh`; records path `.`; single object; pagination `none`; query keys `group_id`; computed fields `group_id`; passthrough projection.

## Write actions & risks

All Freshchat writes remain behind reverse ETL plan -> preview -> explicit approval -> execute. The
bundle does not authorize live provider writes, and fixtures do not certify live provider behavior.
DELETE and administrative agent-management operations additionally declare `confirm: "destructive"`
so execution requires typed destructive confirmation through the existing safety path.

- `create_user`: POST `/users`; kind `create`; required none; accepted fields `email`, `first_name`, `last_name`, `phone`, `properties`, `reference_id`; risk: creates a Freshchat user/contact visible to agents.
- `update_user`: PUT `/users/{{ record.user_id }}`; kind `update`; required `user_id`; accepted fields `email`, `first_name`, `last_name`, `phone`, `properties`, `reference_id`, `user_id`; risk: updates an existing Freshchat user/contact.
- `delete_user`: DELETE `/users/{{ record.user_id }}`; kind `delete`; required `user_id`; accepted fields `user_id`; risk: deletes a Freshchat user/contact; destructive and idempotent for configured missing statuses; typed destructive confirmation required; idempotent delete missing statuses 404; redacts `user_id`.
- `create_conversation`: POST `/conversations`; kind `create`; required none; accepted fields `channel_id`, `messages`, `properties`, `user_id`; risk: creates a Freshchat conversation.
- `update_conversation`: PUT `/conversations/{{ record.conversation_id }}`; kind `update`; required `conversation_id`; accepted fields `assigned_agent_id`, `assigned_group_id`, `conversation_id`, `priority`, `properties`, `status`; risk: updates routing, status, or properties on an existing Freshchat conversation.
- `send_conversation_message`: POST `/conversations/{{ record.conversation_id }}/messages`; kind `create`; required `conversation_id`; accepted fields `actor_id`, `actor_type`, `conversation_id`, `message_parts`; risk: sends a message into an existing Freshchat conversation.
- `create_agent`: POST `/agents`; kind `create`; required `email`; accepted fields `email`, `first_name`, `groups`, `last_name`, `role_id`; risk: creates a Freshchat admin/agent account; requires typed destructive confirmation because this is an administrative user-management action; typed destructive confirmation required; redacts `email`.
- `update_agent`: PUT `/agents/{{ record.agent_id }}`; kind `update`; required `agent_id`; accepted fields `agent_id`, `first_name`, `groups`, `is_deactivated`, `last_name`, `role_id`; risk: updates a Freshchat admin/agent account; requires typed destructive confirmation because this mutates administrative access metadata; typed destructive confirmation required; redacts `agent_id`.
- `update_agent_status`: PATCH `/agents/{{ record.agent_id }}`; kind `update`; required `agent_id`, `status`; accepted fields `agent_id`, `status`; risk: updates a Freshchat agent availability status; requires typed destructive confirmation for administrative agent-state mutation; typed destructive confirmation required; redacts `agent_id`.
- `delete_agent`: DELETE `/agents/{{ record.agent_id }}`; kind `delete`; required `agent_id`; accepted fields `agent_id`; risk: deletes a Freshchat admin/agent account; destructive and idempotent for configured missing statuses; typed destructive confirmation required; idempotent delete missing statuses 404; redacts `agent_id`.
- `send_outbound_whatsapp_message`: POST `/outbound-messages/whatsapp`; kind `create`; required none; accepted fields `from`, `provider`, `template`, `to`; risk: sends an outbound WhatsApp message through Freshchat.
- `extract_report`: POST `/reports/raw`; kind `create`; required `start`, `end`, `event`, `format`; accepted fields `end`, `event`, `format`, `start`; risk: requests generation of a Freshchat raw report extract.
- `create_csat_rating`: POST `/csat/{{ record.conversation_id }}`; kind `create`; required `conversation_id`; accepted fields `comment`, `conversation_id`, `rating`; risk: creates a CSAT rating for a Freshchat conversation.

Blocked/planned direct, provider-search, and binary/file operations:

- `POST /users/fetch` — blocked `direct_read` risk `medium`: Freshchat subset user fetch is a fixed POST-body provider-search operation with ids[] bounded by the provider to 100 users, but executable provider_search/provider_query support is blocked on shared foundation #2985; no raw request body/query escape hatch is exposed.
- `POST /files/upload` — blocked `sensitive_reverse_etl` risk `high`: Freshchat file upload is multipart/form-data with local file input and a documented 25 MB single-file cap; the current Freshchat bundle has no connector-local binary/multipart execution contract without shared binary/file foundation or an approved hook.
- `POST /images/upload` — blocked `sensitive_reverse_etl` risk `high`: Freshchat image upload is multipart/form-data with local image input; the current Freshchat bundle has no connector-local binary/multipart execution contract without shared binary/file foundation or an approved hook.

## Known limits

- Fixture-only status: this bundle is not live-certified. `certification.json` supplies source
  defaults and live-unavailable classifiers only; it deliberately declares no Freshchat write
  pairing so certification write stages use the built-in outbox self-test unless a future approved
  live-safe pairing is added.
- `POST /users/fetch` is a fixed provider-search operation with an `ids[]` request body and an
  official maximum of 100 users. It remains blocked on shared provider-search/query foundation #2985
  and is not exposed as a raw body/query command.
- `POST /files/upload` and `POST /images/upload` are multipart binary/file uploads. They remain
  blocked until a typed binary/multipart contract or approved connector hook provides file-path
  validation, size caps, redaction, preview, approval, cleanup, and conformance evidence.
- Freshchat has no documented CDC/changefeed operations in the reviewed official inventory; this
  connector keeps `capabilities.cdc=false` and makes no CDC certification claim.
- Shared foundation #2986/#2988 CDC work is tracked outside this connector and does not block
  Freshchat because no Freshchat CDC executor is advertised.
- No generic HTTP, SQL, shell, file, GraphQL, method/path/body, or raw query escape hatch is exposed
  by this connector.
