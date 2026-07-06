---
name: pm-freshchat
description: Freshchat connector knowledge and safe action guide.
---

# pm-freshchat

## Purpose

Reads Freshchat account, user, conversation, agent, group, channel, role, outbound, report, metrics, and business-hours data through the Freshchat v2 REST API; writes Freshchat users, conversations, agents, outbound messages, reports, and CSAT ratings.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- agent_id
- base_url
- business_hours_group_id
- conversation_id
- metrics_aggregator
- metrics_end
- metrics_filter_by
- metrics_group_by
- metrics_interval
- metrics_metric
- metrics_start
- metrics_summary
- mode
- outbound_request_id
- report_id
- report_status
- user_id
- api_key (secret)

## ETL Streams

- account_configuration:
  - primary key: account_id
  - fields: account_domain(), account_id(), app_id(), bundle_id(), bundle_type(), datacenter(), organisation_domain(), organisation_id(), plan_type()
- agents:
  - primary key: id
  - cursor: updated_time
  - fields: avatar(), biography(), created_time(), email(), first_name(), groups(), id(), is_deactivated(), is_deleted(), last_name(), role_id(), social_profiles(), updated_time()
- agent_details:
  - primary key: id
  - cursor: updated_time
  - fields: avatar(), biography(), created_time(), email(), first_name(), groups(), id(), is_deactivated(), is_deleted(), last_name(), role_id(), social_profiles(), updated_time()
- agent_statuses:
  - primary key: id
  - fields: enabled(), id(), name(), type()
- users:
  - primary key: id
  - cursor: updated_time
  - fields: avatar(), created_time(), email(), first_name(), id(), last_name(), phone(), properties(), reference_id(), restore_id(), updated_time()
- user_details:
  - primary key: id
  - cursor: updated_time
  - fields: avatar(), created_time(), email(), first_name(), id(), last_name(), phone(), properties(), reference_id(), restore_id(), updated_time()
- user_conversations:
  - primary key: id
  - fields: app_id(), assigned_agent_id(), assigned_group_id(), channel_id(), created_time(), id(), messages(), priority(), properties(), status(), updated_time(), user_id()
- conversation_detail:
  - primary key: id
  - fields: app_id(), assigned_agent_id(), assigned_group_id(), channel_id(), created_time(), id(), messages(), priority(), properties(), status(), updated_time(), user_id()
- conversation_messages:
  - primary key: id
  - fields: actor_id(), actor_type(), app_id(), conversation_id(), created_time(), id(), message_parts(), message_type(), updated_time()
- conversation_fields:
  - primary key: name
  - fields: choices(), label(), name(), required(), type()
- groups:
  - primary key: id
  - fields: created_time(), description(), id(), name(), routing_type(), updated_time()
- channels:
  - primary key: id
  - cursor: updated_time
  - fields: created_time(), enabled(), icon(), id(), locale(), name(), public(), tags(), updated_time(), welcome_message()
- roles:
  - primary key: id
  - fields: description(), id(), name(), role()
- outbound_messages:
  - primary key: id
  - fields: created_time(), from(), id(), provider(), request_id(), status(), template(), to(), updated_time()
- report_status:
  - primary key: id
  - fields: id(), interval(), link(), links(), status()
- historical_metrics:
  - primary key: metric_type
  - fields: aggregator(), data(), end(), filters(), interval(), metric_type(), metrics(), start()
- instant_metrics:
  - primary key: metric_type
  - fields: aggregator(), data(), end(), filters(), interval(), metric_type(), metrics(), start()
- business_hours_status:
  - primary key: group_id
  - fields: business_hours_id(), group_id(), timezone(), within_business_hours()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_user:
  - endpoint: POST /users
  - risk: creates a Freshchat user/contact visible to agents
- update_user:
  - endpoint: PUT /users/{{ record.user_id }}
  - required fields: user_id
  - risk: updates an existing Freshchat user/contact
- delete_user:
  - endpoint: DELETE /users/{{ record.user_id }}
  - required fields: user_id
  - risk: deletes a Freshchat user/contact
- create_conversation:
  - endpoint: POST /conversations
  - risk: creates a Freshchat conversation
- update_conversation:
  - endpoint: PUT /conversations/{{ record.conversation_id }}
  - required fields: conversation_id
  - risk: updates routing, status, or properties on an existing Freshchat conversation
- send_conversation_message:
  - endpoint: POST /conversations/{{ record.conversation_id }}/messages
  - required fields: conversation_id
  - risk: sends a message into an existing Freshchat conversation
- create_agent:
  - endpoint: POST /agents
  - risk: creates a Freshchat agent account
- update_agent:
  - endpoint: PUT /agents/{{ record.agent_id }}
  - required fields: agent_id
  - risk: updates an existing Freshchat agent
- update_agent_status:
  - endpoint: PATCH /agents/{{ record.agent_id }}
  - required fields: agent_id
  - risk: updates an agent's Freshchat availability status
- delete_agent:
  - endpoint: DELETE /agents/{{ record.agent_id }}
  - required fields: agent_id
  - risk: deletes a Freshchat agent
- send_outbound_whatsapp_message:
  - endpoint: POST /outbound-messages/whatsapp
  - risk: sends an outbound WhatsApp message through Freshchat
- extract_report:
  - endpoint: POST /reports/raw
  - risk: requests generation of a Freshchat raw report extract
- create_csat_rating:
  - endpoint: POST /csat/{{ record.conversation_id }}
  - required fields: conversation_id
  - risk: creates a CSAT rating for a Freshchat conversation

## Security

- read risk: external Freshchat API reads of account metadata, users, conversations, messages, agents, groups, channels, roles, outbound messages, reports, metrics, and business-hours status
- write risk: creates, updates, or deletes Freshchat users, conversations, agents, outbound messages, reports, and CSAT ratings
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect freshchat
```

### Inspect as structured JSON

```bash
pm connectors inspect freshchat --json
```

## Agent Rules

- Run pm connectors inspect freshchat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
