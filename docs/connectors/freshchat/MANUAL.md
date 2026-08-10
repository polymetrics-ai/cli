# pm connectors inspect freshchat

```text
NAME
  pm connectors inspect freshchat - Freshchat connector manual

SYNOPSIS
  pm connectors inspect freshchat
  pm connectors inspect freshchat --json
  pm credentials add <name> --connector freshchat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freshchat account, user, conversation, agent, group, channel, role, outbound, report, metrics, and business-hours data through the Freshchat v2 REST API; writes Freshchat users, conversations, agents, outbound messages, reports, and CSAT ratings.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  agent_id
  agents_availability_status
  agents_groups
  agents_is_deactivated
  agents_sort_by
  agents_sort_order
  base_url
  business_hours_group_id
  channels_locale
  conversation_id
  conversation_messages_from_time
  groups_sort_by
  groups_sort_order
  metrics_aggregator
  metrics_end
  metrics_filter_by
  metrics_group_by
  metrics_interval
  metrics_metric
  metrics_start
  metrics_summary
  mode
  outbound_request_id
  report_id
  report_status
  user_id
  users_created_from
  users_created_to
  users_email
  users_first_name
  users_last_name
  users_phone_no
  users_reference_id
  users_updated_from
  users_updated_to
  api_key (secret)

ETL STREAMS
  account_configuration:
    primary key: account_id
    fields: account_domain(string), account_id(integer), app_id(string), bundle_id(integer), bundle_type(string), datacenter(string), organisation_domain(string), organisation_id(integer), plan_type(string)
  agents:
    primary key: id
    cursor: updated_time
    fields: avatar(object), biography(string), created_time(string), email(string), first_name(string), groups(array), id(string), is_deactivated(boolean), is_deleted(boolean), last_name(string), role_id(string), social_profiles(array), updated_time(string)
  agent_details:
    primary key: id
    cursor: updated_time
    fields: avatar(object), biography(string), created_time(string), email(string), first_name(string), groups(array), id(string), is_deactivated(boolean), is_deleted(boolean), last_name(string), role_id(string), social_profiles(array), updated_time(string)
  agent_statuses:
    primary key: id
    fields: enabled(boolean), id(string), name(string), type(string)
  users:
    primary key: id
    cursor: updated_time
    fields: avatar(object), created_time(string), email(string), first_name(string), id(string), last_name(string), phone(string), properties(array), reference_id(string), restore_id(string), updated_time(string)
  user_details:
    primary key: id
    cursor: updated_time
    fields: avatar(object), created_time(string), email(string), first_name(string), id(string), last_name(string), phone(string), properties(array), reference_id(string), restore_id(string), updated_time(string)
  user_conversations:
    primary key: id
    fields: app_id(string), assigned_agent_id(string), assigned_group_id(string), channel_id(string), created_time(string), id(string), messages(array), priority(string), properties(object), status(string), updated_time(string), user_id(string)
  conversation_detail:
    primary key: id
    fields: app_id(string), assigned_agent_id(string), assigned_group_id(string), channel_id(string), created_time(string), id(string), messages(array), priority(string), properties(object), status(string), updated_time(string), user_id(string)
  conversation_messages:
    primary key: id
    fields: actor_id(string), actor_type(string), app_id(string), conversation_id(string), created_time(string), id(string), message_parts(array), message_type(string), updated_time(string)
  conversation_fields:
    primary key: name
    fields: choices(array), label(string), name(string), required(boolean), type(string)
  groups:
    primary key: id
    fields: created_time(string), description(string), id(string), name(string), routing_type(string), updated_time(string)
  channels:
    primary key: id
    cursor: updated_time
    fields: created_time(string), enabled(boolean), icon(object), id(string), locale(string), name(string), public(boolean), tags(array), updated_time(string), welcome_message(object)
  roles:
    primary key: id
    fields: description(string), id(string), name(string), role(string)
  outbound_messages:
    primary key: id
    fields: created_time(string), from(object), id(string), provider(string), request_id(string), status(string), template(object), to(object), updated_time(string)
  report_status:
    primary key: id
    fields: id(string), interval(string), link(object), links(array), status(string)
  historical_metrics:
    primary key: metric_type
    fields: aggregator(string), data(array), end(string), filters(object), interval(string), metric_type(string), metrics(array), start(string)
  instant_metrics:
    primary key: metric_type
    fields: aggregator(string), data(array), end(string), filters(object), interval(string), metric_type(string), metrics(array), start(string)
  business_hours_status:
    primary key: group_id
    fields: business_hours_id(string), group_id(string), timezone(string), within_business_hours(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /users
    risk: creates a Freshchat user/contact visible to agents
  update_user:
    endpoint: PUT /users/{{ record.user_id }}
    required fields: user_id
    risk: updates an existing Freshchat user/contact
  delete_user:
    endpoint: DELETE /users/{{ record.user_id }}
    required fields: user_id
    risk: deletes a Freshchat user/contact; destructive and idempotent for configured missing statuses
  create_conversation:
    endpoint: POST /conversations
    risk: creates a Freshchat conversation
  update_conversation:
    endpoint: PUT /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: updates routing, status, or properties on an existing Freshchat conversation
  send_conversation_message:
    endpoint: POST /conversations/{{ record.conversation_id }}/messages
    required fields: conversation_id
    risk: sends a message into an existing Freshchat conversation
  create_agent:
    endpoint: POST /agents
    required fields: email
    risk: creates a Freshchat admin/agent account; requires typed destructive confirmation because this is an administrative user-management action
  update_agent:
    endpoint: PUT /agents/{{ record.agent_id }}
    required fields: agent_id
    risk: updates a Freshchat admin/agent account; requires typed destructive confirmation because this mutates administrative access metadata
  update_agent_status:
    endpoint: PATCH /agents/{{ record.agent_id }}
    required fields: agent_id, status
    risk: updates a Freshchat agent availability status; requires typed destructive confirmation for administrative agent-state mutation
  delete_agent:
    endpoint: DELETE /agents/{{ record.agent_id }}
    required fields: agent_id
    risk: deletes a Freshchat admin/agent account; destructive and idempotent for configured missing statuses
  send_outbound_whatsapp_message:
    endpoint: POST /outbound-messages/whatsapp
    risk: sends an outbound WhatsApp message through Freshchat
  extract_report:
    endpoint: POST /reports/raw
    required fields: start, end, event, format
    risk: requests generation of a Freshchat raw report extract
  create_csat_rating:
    endpoint: POST /csat/{{ record.conversation_id }}
    required fields: conversation_id
    risk: creates a CSAT rating for a Freshchat conversation

SECURITY
  read risk: external Freshchat API reads of account metadata, users, conversations, messages, agents, groups, channels, roles, outbound messages, reports, metrics, and business-hours status
  write risk: creates, updates, or deletes Freshchat users, conversations, agents, outbound messages, reports, and CSAT ratings
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Freshchat's declared streams and reverse-ETL actions.
  Usage: pm freshchat <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    account configuration - Read Freshchat account configuration. [intent=etl availability=implemented stream=account_configuration]
    agents create - Plan creation of a Freshchat admin/agent account. [intent=reverse_etl availability=implemented write=create_agent]; approval: requires plan, preview, approval, and execute; risk: creates a Freshchat admin/agent account; requires typed destructive confirmation because this is an administrative user-management action; flags: --email (required)
    agents delete - Plan deletion of a Freshchat admin/agent account. [intent=reverse_etl availability=implemented write=delete_agent]; approval: requires plan, preview, approval, and execute; risk: deletes a Freshchat admin/agent account; destructive and idempotent for configured missing statuses; flags: --agent_id (required)
    agents list - List Freshchat agents with optional documented filters. [intent=etl availability=implemented stream=agents]; flags: --is-deactivated, --groups, --availability-status, --sort-order, --sort-by
    agents status update - Plan update of a Freshchat agent availability status. [intent=reverse_etl availability=implemented write=update_agent_status]; approval: requires plan, preview, approval, and execute; risk: updates a Freshchat agent availability status; requires typed destructive confirmation for administrative agent-state mutation; flags: --agent_id (required), --status (required)
    agents statuses - List Freshchat agent statuses. [intent=etl availability=implemented stream=agent_statuses]
    agents update - Plan update of a Freshchat admin/agent account. [intent=reverse_etl availability=implemented write=update_agent]; approval: requires plan, preview, approval, and execute; risk: updates a Freshchat admin/agent account; requires typed destructive confirmation because this mutates administrative access metadata; flags: --agent_id (required)
    agents view - Read one Freshchat agent by configured agent id. [intent=etl availability=implemented stream=agent_details]; flags: --agent-id
    api post files upload - Documented POST /files/upload (not implemented) [intent=direct_write availability=not_implemented operation=freshchat.post.files-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post images upload - Documented POST /images/upload (not implemented) [intent=direct_write availability=not_implemented operation=freshchat.post.images-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post users fetch - Documented POST /users/fetch (not implemented) [intent=direct_read availability=not_implemented operation=freshchat.post.users-fetch]; approval: not implemented: the direct-read executor lacks a reviewed operation-specific request, output-policy, and CLI contract; risk: medium; notes: named_dependency=engine.direct_read_operation_contract: the direct-read executor lacks a reviewed operation-specific request, output-policy, and CLI contract; flags: --page, --page-cursor
    business-hours status - Check whether a Freshchat group is within business hours. [intent=etl availability=implemented stream=business_hours_status]; flags: --group-id
    channels list - List Freshchat channels/topics. [intent=etl availability=implemented stream=channels]; flags: --locale
    conversations create - Plan creation of a Freshchat conversation. [intent=reverse_etl availability=implemented write=create_conversation]; approval: requires plan, preview, approval, and execute; risk: creates a Freshchat conversation
    conversations fields - List Freshchat conversation property fields. [intent=etl availability=implemented stream=conversation_fields]
    conversations message send - Plan sending a message into a Freshchat conversation. [intent=reverse_etl availability=implemented write=send_conversation_message]; approval: requires plan, preview, approval, and execute; risk: sends a message into an existing Freshchat conversation; flags: --conversation_id (required)
    conversations messages - List messages in a Freshchat conversation. [intent=etl availability=implemented stream=conversation_messages]; flags: --conversation-id, --from-time
    conversations update - Plan update of a Freshchat conversation. [intent=reverse_etl availability=implemented write=update_conversation]; approval: requires plan, preview, approval, and execute; risk: updates routing, status, or properties on an existing Freshchat conversation; flags: --conversation_id (required)
    conversations view - Read one Freshchat conversation. [intent=etl availability=implemented stream=conversation_detail]; flags: --conversation-id
    csat create - Plan creation of a Freshchat CSAT rating. [intent=reverse_etl availability=implemented write=create_csat_rating]; approval: requires plan, preview, approval, and execute; risk: creates a CSAT rating for a Freshchat conversation; flags: --conversation_id (required)
    groups list - List Freshchat groups. [intent=etl availability=implemented stream=groups]; flags: --sort-order, --sort-by
    metrics historical - Read Freshchat historical metrics. [intent=etl availability=implemented stream=historical_metrics]; flags: --metric, --start, --end, --group-by, --filter-by, --aggregator, --interval
    metrics instant - Read Freshchat instant metrics. [intent=etl availability=implemented stream=instant_metrics]; flags: --metric, --group-by, --filter-by, --summary
    outbound messages list - List Freshchat outbound messages. [intent=etl availability=implemented stream=outbound_messages]; flags: --request-id
    outbound whatsapp send - Plan sending an outbound WhatsApp message through Freshchat. [intent=reverse_etl availability=implemented write=send_outbound_whatsapp_message]; approval: requires plan, preview, approval, and execute; risk: sends an outbound WhatsApp message through Freshchat
    reports extract - Plan generation of a Freshchat raw report extract. [intent=reverse_etl availability=implemented write=extract_report]; approval: requires plan, preview, approval, and execute; risk: requests generation of a Freshchat raw report extract; flags: --end (required), --event (required), --format (required), --start (required)
    reports status - Read Freshchat raw report status. [intent=etl availability=implemented stream=report_status]; flags: --report-id, --status
    roles list - List Freshchat roles. [intent=etl availability=implemented stream=roles]
    users conversations - List conversations for a Freshchat user. [intent=etl availability=implemented stream=user_conversations]; flags: --user-id
    users create - Plan creation of a Freshchat user/contact. [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: creates a Freshchat user/contact visible to agents
    users delete - Plan deletion of a Freshchat user/contact. [intent=reverse_etl availability=implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: deletes a Freshchat user/contact; destructive and idempotent for configured missing statuses; flags: --user_id (required)
    users list - List Freshchat users with documented search filters. [intent=etl availability=implemented stream=users]; notes: Freshchat requires at least one user search filter for successful live /users responses; connector config and command metadata expose the documented filter set without a raw query escape hatch.; flags: --first-name, --last-name, --email, --reference-id, --phone-no, --created-from, --created-to, --updated-from, --updated-to
    users update - Plan update of a Freshchat user/contact. [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: updates an existing Freshchat user/contact; flags: --user_id (required)
    users view - Read one Freshchat user by configured user id. [intent=etl availability=implemented stream=user_details]; flags: --user-id

EXAMPLES
  # Inspect as a manual
  pm connectors inspect freshchat

  # Inspect as structured JSON
  pm connectors inspect freshchat --json

AGENT WORKFLOW
  - Run pm connectors inspect freshchat before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
