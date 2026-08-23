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
  base_url (required)
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
  api_key (secret) (required)

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
  Read Freshchat data and safely plan Freshchat write actions without raw provider escape hatches.
  Usage: pm freshchat <account|agents|users|conversations|groups|channels|roles|outbound|reports|metrics|business-hours|csat|direct|binary> <command> [flags]
  Source CLI: Freshchat API (Freshchat API docs ETag W/"26e4fd8b1fe01578eae1dbaff6b69224")
  Global flags:
    --credential (non-empty) (string): Credential profile name; never pass secret values as flags.: maps_to=config.credential
    --config (string_array): Connector config override as key=value; never pass secret values here.
    --json (boolean): Render machine-readable JSON output.
    --limit (integer): Maximum records to emit from stream-backed commands.
    --preview (boolean): Preview a reverse-ETL write without making a network mutation.
    --plan (string): Execute an approved reverse-ETL plan by id.
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
    --confirm (string): Typed confirmation challenge for destructive/admin reverse-ETL writes.
  Freshchat account read commands
    account configuration - Read Freshchat account configuration. [intent=etl availability=implemented stream=account_configuration]
  Freshchat agent read/write/admin commands
    agents list - List Freshchat agents with optional documented filters. [intent=etl availability=implemented stream=agents]; flags: --is-deactivated, --groups (non-empty), --availability-status, --sort-order, --sort-by (non-empty)
    agents view - Read one Freshchat agent by configured agent id. [intent=etl availability=implemented stream=agent_details]; flags: --agent-id (non-empty)
    agents statuses - List Freshchat agent statuses. [intent=etl availability=implemented stream=agent_statuses]
    agents create - Plan creation of a Freshchat admin/agent account. [intent=reverse_etl availability=implemented write=create_agent]; approval: Destructive/admin reverse ETL writes require typed destructive confirmation plus plan -> preview -> explicit approval -> execute.; risk: Creates a Freshchat admin/agent account.; flags: --email (non-empty), --first-name (non-empty), --last-name (non-empty), --role-id (non-empty), --groups
    agents update - Plan update of a Freshchat admin/agent account. [intent=reverse_etl availability=implemented write=update_agent]; approval: Destructive/admin reverse ETL writes require typed destructive confirmation plus plan -> preview -> explicit approval -> execute.; risk: Updates a Freshchat admin/agent account.; flags: --agent-id (non-empty), --first-name (non-empty), --last-name (non-empty), --role-id (non-empty), --groups, --is-deactivated
    agents status update - Plan update of a Freshchat agent availability status. [intent=reverse_etl availability=implemented write=update_agent_status]; approval: Destructive/admin reverse ETL writes require typed destructive confirmation plus plan -> preview -> explicit approval -> execute.; risk: Updates a Freshchat agent availability status.; flags: --agent-id (non-empty), --status (non-empty)
    agents delete - Plan deletion of a Freshchat admin/agent account. [intent=reverse_etl availability=implemented write=delete_agent]; approval: Destructive/admin reverse ETL writes require typed destructive confirmation plus plan -> preview -> explicit approval -> execute.; risk: Deletes a Freshchat admin/agent account.; flags: --agent-id (non-empty)
  Freshchat user read/write commands
    users list - List Freshchat users with documented search filters. [intent=etl availability=implemented stream=users]; notes: Freshchat requires at least one user search filter for successful live /users responses; connector config and command metadata expose the documented filter set without a raw query escape hatch.; flags: --first-name (non-empty), --last-name (non-empty), --email (non-empty), --reference-id (non-empty), --phone-no (non-empty), --created-from (non-empty, format=date-time), --created-to (non-empty, format=date-time), --updated-from (non-empty, format=date-time), --updated-to (non-empty, format=date-time)
    users view - Read one Freshchat user by configured user id. [intent=etl availability=implemented stream=user_details]; flags: --user-id (non-empty)
    users conversations - List conversations for a Freshchat user. [intent=etl availability=implemented stream=user_conversations]; flags: --user-id (non-empty)
    users create - Plan creation of a Freshchat user/contact. [intent=reverse_etl availability=implemented write=create_user]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Creates a Freshchat user/contact visible to agents.; flags: --email (non-empty), --first-name (non-empty), --last-name (non-empty), --phone (non-empty), --reference-id (non-empty)
    users update - Plan update of a Freshchat user/contact. [intent=reverse_etl availability=implemented write=update_user]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Updates an existing Freshchat user/contact.; flags: --user-id (non-empty), --email (non-empty), --first-name (non-empty), --last-name (non-empty), --phone (non-empty), --reference-id (non-empty)
    users delete - Plan deletion of a Freshchat user/contact. [intent=reverse_etl availability=implemented write=delete_user]; approval: Destructive/admin reverse ETL writes require typed destructive confirmation plus plan -> preview -> explicit approval -> execute.; risk: Deletes a Freshchat user/contact.; flags: --user-id (non-empty)
  Freshchat conversation read/write commands
    conversations view - Read one Freshchat conversation. [intent=etl availability=implemented stream=conversation_detail]; flags: --conversation-id (non-empty)
    conversations messages - List messages in a Freshchat conversation. [intent=etl availability=implemented stream=conversation_messages]; flags: --conversation-id (non-empty), --from-time (non-empty, format=date-time)
    conversations fields - List Freshchat conversation property fields. [intent=etl availability=implemented stream=conversation_fields]
    conversations create - Plan creation of a Freshchat conversation. [intent=reverse_etl availability=implemented write=create_conversation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Creates a Freshchat conversation.; flags: --user-id (non-empty), --channel-id (non-empty)
    conversations update - Plan update of a Freshchat conversation. [intent=reverse_etl availability=implemented write=update_conversation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Updates status, routing, priority, or properties on a Freshchat conversation.; flags: --conversation-id (non-empty), --status (non-empty), --assigned-agent-id (non-empty), --assigned-group-id (non-empty), --priority (non-empty)
    conversations message send - Plan sending a message into a Freshchat conversation. [intent=reverse_etl availability=implemented write=send_conversation_message]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Sends a visible message into an existing Freshchat conversation.; flags: --conversation-id (non-empty), --actor-type (non-empty), --actor-id (non-empty)
  Freshchat group read commands
    groups list - List Freshchat groups. [intent=etl availability=implemented stream=groups]; flags: --sort-order, --sort-by (non-empty)
  Freshchat channel read commands
    channels list - List Freshchat channels/topics. [intent=etl availability=implemented stream=channels]; flags: --locale (non-empty)
  Freshchat role read commands
    roles list - List Freshchat roles. [intent=etl availability=implemented stream=roles]
  Freshchat outbound read/write commands
    outbound messages list - List Freshchat outbound messages. [intent=etl availability=implemented stream=outbound_messages]; flags: --request-id (non-empty)
    outbound whatsapp send - Plan sending an outbound WhatsApp message through Freshchat. [intent=reverse_etl availability=implemented write=send_outbound_whatsapp_message]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Sends an outbound WhatsApp message through Freshchat.; flags: --provider (non-empty)
  Freshchat report read/write commands
    reports status - Read Freshchat raw report status. [intent=etl availability=implemented stream=report_status]; flags: --report-id (non-empty), --status (non-empty)
    reports extract - Plan generation of a Freshchat raw report extract. [intent=reverse_etl availability=implemented write=extract_report]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Requests generation of a Freshchat raw report extract.; flags: --start (non-empty, format=date-time), --end (non-empty, format=date-time), --event (non-empty), --format
  Freshchat metrics read commands
    metrics historical - Read Freshchat historical metrics. [intent=etl availability=implemented stream=historical_metrics]; flags: --metric (non-empty), --start (non-empty, format=date-time), --end (non-empty, format=date-time), --group-by (non-empty), --filter-by (non-empty), --aggregator (non-empty), --interval (non-empty)
    metrics instant - Read Freshchat instant metrics. [intent=etl availability=implemented stream=instant_metrics]; flags: --metric (non-empty), --group-by (non-empty), --filter-by (non-empty), --summary (non-empty)
  Freshchat business-hours read commands
    business-hours status - Check whether a Freshchat group is within business hours. [intent=etl availability=implemented stream=business_hours_status]; flags: --group-id (non-empty)
  Freshchat CSAT write commands
    csat create - Plan creation of a Freshchat CSAT rating. [intent=reverse_etl availability=implemented write=create_csat_rating]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute.; risk: Creates a CSAT rating for a Freshchat conversation.; flags: --conversation-id (non-empty), --rating, --comment (non-empty)
  Planned bounded provider-search commands
    direct users fetch - Planned bounded subset fetch for up to 100 Freshchat user ids. [intent=direct_read availability=planned]; approval: none until #2985 provides executable provider_search/provider_query safety; no live call is available today.; risk: medium; notes: Blocked on #2985 typed provider-search/query foundation. The official request body is ids[] with a provider-documented maximum of 100 users; no raw body escape hatch is exposed.; flags: --ids, --page, --page-cursor
  Planned binary/multipart commands
    binary files upload - Planned Freshchat file upload. [intent=direct_write availability=planned]; approval: blocked until a typed binary/multipart plan -> preview -> approval -> execute contract exists.; risk: high; notes: Blocked binary/file operation. Official API accepts one multipart file with a documented 25 MB cap; this connector currently accepts no filesystem path or binary payload.; flags: --file (non-empty)
    binary images upload - Planned Freshchat image upload. [intent=direct_write availability=planned]; approval: blocked until a typed binary/multipart plan -> preview -> approval -> execute contract exists.; risk: high; notes: Blocked binary/file operation. Official API accepts multipart image input; this connector currently accepts no filesystem path or binary payload.; flags: --image (non-empty)
  Help topics:
    destructive-confirmation - Freshchat DELETE and admin/agent-management operations require typed destructive confirmation plus reverse ETL plan -> preview -> approval -> execute.
    provider-search - Freshchat POST /users/fetch remains planned/blocked on #2985 and is not exposed as a raw query/body command.
    binary-uploads - Freshchat file/image uploads remain planned/blocked until a typed binary/multipart safety contract exists.

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
