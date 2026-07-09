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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  agent_id
  base_url
  business_hours_group_id
  conversation_id
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
  api_key (secret)

ETL STREAMS
  account_configuration: 
    primary key: account_id
    fields: account_domain(), account_id(), app_id(), bundle_id(), bundle_type(), datacenter(), organisation_domain(), organisation_id(), plan_type()
  agents: 
    primary key: id
    cursor: updated_time
    fields: avatar(), biography(), created_time(), email(), first_name(), groups(), id(), is_deactivated(), is_deleted(), last_name(), role_id(), social_profiles(), updated_time()
  agent_details: 
    primary key: id
    cursor: updated_time
    fields: avatar(), biography(), created_time(), email(), first_name(), groups(), id(), is_deactivated(), is_deleted(), last_name(), role_id(), social_profiles(), updated_time()
  agent_statuses: 
    primary key: id
    fields: enabled(), id(), name(), type()
  users: 
    primary key: id
    cursor: updated_time
    fields: avatar(), created_time(), email(), first_name(), id(), last_name(), phone(), properties(), reference_id(), restore_id(), updated_time()
  user_details: 
    primary key: id
    cursor: updated_time
    fields: avatar(), created_time(), email(), first_name(), id(), last_name(), phone(), properties(), reference_id(), restore_id(), updated_time()
  user_conversations: 
    primary key: id
    fields: app_id(), assigned_agent_id(), assigned_group_id(), channel_id(), created_time(), id(), messages(), priority(), properties(), status(), updated_time(), user_id()
  conversation_detail: 
    primary key: id
    fields: app_id(), assigned_agent_id(), assigned_group_id(), channel_id(), created_time(), id(), messages(), priority(), properties(), status(), updated_time(), user_id()
  conversation_messages: 
    primary key: id
    fields: actor_id(), actor_type(), app_id(), conversation_id(), created_time(), id(), message_parts(), message_type(), updated_time()
  conversation_fields: 
    primary key: name
    fields: choices(), label(), name(), required(), type()
  groups: 
    primary key: id
    fields: created_time(), description(), id(), name(), routing_type(), updated_time()
  channels: 
    primary key: id
    cursor: updated_time
    fields: created_time(), enabled(), icon(), id(), locale(), name(), public(), tags(), updated_time(), welcome_message()
  roles: 
    primary key: id
    fields: description(), id(), name(), role()
  outbound_messages: 
    primary key: id
    fields: created_time(), from(), id(), provider(), request_id(), status(), template(), to(), updated_time()
  report_status: 
    primary key: id
    fields: id(), interval(), link(), links(), status()
  historical_metrics: 
    primary key: metric_type
    fields: aggregator(), data(), end(), filters(), interval(), metric_type(), metrics(), start()
  instant_metrics: 
    primary key: metric_type
    fields: aggregator(), data(), end(), filters(), interval(), metric_type(), metrics(), start()
  business_hours_status: 
    primary key: group_id
    fields: business_hours_id(), group_id(), timezone(), within_business_hours()

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
    risk: deletes a Freshchat user/contact
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
    risk: creates a Freshchat agent account
  update_agent: 
    endpoint: PUT /agents/{{ record.agent_id }}
    required fields: agent_id
    risk: updates an existing Freshchat agent
  update_agent_status: 
    endpoint: PATCH /agents/{{ record.agent_id }}
    required fields: agent_id
    risk: updates an agent's Freshchat availability status
  delete_agent: 
    endpoint: DELETE /agents/{{ record.agent_id }}
    required fields: agent_id
    risk: deletes a Freshchat agent
  send_outbound_whatsapp_message: 
    endpoint: POST /outbound-messages/whatsapp
    risk: sends an outbound WhatsApp message through Freshchat
  extract_report: 
    endpoint: POST /reports/raw
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
  Work with Freshchat conversations, users, agents, reports, and metrics from the command line.
  Usage: pm freshchat <command> [flags]
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --credential (string): Use a saved Freshchat connector credential.: maps_to=credential
    --connection (string): Alias for --credential.: maps_to=connection
  Account Commands
    account configuration - Read Freshchat account configuration [intent=etl availability=implemented stream=account_configuration]
    business-hours status - Check whether a Freshchat group is within business hours [intent=etl availability=implemented stream=business_hours_status]; notes: Requires business_hours_group_id in connector config; use --config business_hours_group_id=<group_id> with the generic connector command runner.
  User Commands
    user list - List Freshchat users [intent=etl availability=implemented stream=users]
    user fetch - Fetch a request-body-selected subset of Freshchat users [intent=direct_read availability=planned]; notes: POST /users/fetch is read-like but requires request-body criteria; bounded direct-read body support is tracked by issues #185 and #186.
    user view - Read one Freshchat user [intent=etl availability=implemented stream=user_details]; notes: Requires user_id in connector config; use --config user_id=<user_id> until direct-read path flags are implemented.
    user conversations - List conversations for one Freshchat user [intent=etl availability=implemented stream=user_conversations]; notes: Requires user_id in connector config; use --config user_id=<user_id> until direct-read path flags are implemented.
    user create - Create a Freshchat user [intent=reverse_etl availability=implemented write=create_user]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Freshchat user/contact visible to agents.; flags: --email, --first-name, --last-name, --phone, --reference-id
    user update - Update a Freshchat user [intent=reverse_etl availability=implemented write=update_user]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Freshchat user/contact.; flags: --user-id, --email, --first-name, --last-name, --phone, --reference-id
    user delete - Delete a Freshchat user [intent=reverse_etl availability=implemented write=delete_user]; approval: reverse ETL writes require plan, preview, approval, execute and destructive confirmation.; risk: Deletes a Freshchat user/contact.; flags: --user-id
  Conversation Commands
    conversation view - Read one Freshchat conversation [intent=etl availability=implemented stream=conversation_detail]; notes: Requires conversation_id in connector config; use --config conversation_id=<conversation_id> until direct-read path flags are implemented.
    conversation messages - List messages for one Freshchat conversation [intent=etl availability=implemented stream=conversation_messages]; notes: Requires conversation_id in connector config; use --config conversation_id=<conversation_id> until direct-read path flags are implemented.
    conversation fields - List Freshchat conversation fields [intent=etl availability=implemented stream=conversation_fields]
    conversation create - Create a Freshchat conversation [intent=reverse_etl availability=partial write=create_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Freshchat conversation.; notes: The write action supports structured messages/properties records; full command flag mapping for nested payloads is deferred.
    conversation update - Update routing, status, or properties on a Freshchat conversation [intent=reverse_etl availability=implemented write=update_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates routing, status, priority, or properties on an existing Freshchat conversation.; flags: --conversation-id, --status, --assigned-agent-id, --assigned-group-id, --priority
    conversation message-send - Send a message into a Freshchat conversation [intent=reverse_etl availability=partial write=send_conversation_message]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sends a visible message into an existing Freshchat conversation.; notes: The write action supports structured message_parts records; full command flag mapping for nested message payloads is deferred.
    csat create - Create a CSAT rating for a Freshchat conversation [intent=reverse_etl availability=implemented write=create_csat_rating]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a CSAT rating for a Freshchat conversation.; flags: --conversation-id, --rating, --comment
  Agent Commands
    agent list - List Freshchat agents [intent=etl availability=implemented stream=agents]
    agent view - Read one Freshchat agent [intent=etl availability=implemented stream=agent_details]; notes: Requires agent_id in connector config; use --config agent_id=<agent_id> until direct-read path flags are implemented.
    agent status-list - List Freshchat agent statuses [intent=etl availability=implemented stream=agent_statuses]
    agent create - Create a Freshchat agent [intent=reverse_etl availability=implemented write=create_agent]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Freshchat agent account.; flags: --email, --first-name, --last-name, --role-id, --group
    agent update - Update a Freshchat agent [intent=reverse_etl availability=implemented write=update_agent]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Freshchat agent.; flags: --agent-id, --first-name, --last-name, --role-id, --group, --is-deactivated
    agent status-update - Update a Freshchat agent availability status [intent=reverse_etl availability=implemented write=update_agent_status]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an agent's Freshchat availability status.; flags: --agent-id, --status
    agent delete - Delete a Freshchat agent [intent=reverse_etl availability=implemented write=delete_agent]; approval: reverse ETL writes require plan, preview, approval, execute and destructive confirmation.; risk: Deletes a Freshchat agent.; flags: --agent-id
  Directory Commands
    group list - List Freshchat groups [intent=etl availability=implemented stream=groups]
    channel list - List Freshchat channels [intent=etl availability=implemented stream=channels]
    role list - List Freshchat roles [intent=etl availability=implemented stream=roles]
  Outbound Messaging Commands
    outbound whatsapp-send - Send an outbound WhatsApp message through Freshchat [intent=reverse_etl availability=partial write=send_outbound_whatsapp_message]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sends an outbound WhatsApp message through Freshchat.; notes: The write action supports structured from/to/template records; full command flag mapping for nested WhatsApp payloads is deferred.
    outbound message-list - List Freshchat outbound messages [intent=etl availability=implemented stream=outbound_messages]
  Report And Metric Commands
    report extract - Request a Freshchat raw report extract [intent=reverse_etl availability=implemented write=extract_report]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Requests generation of a Freshchat raw report extract.; flags: --start, --end, --event, --format
    report status - Read Freshchat raw report status [intent=etl availability=implemented stream=report_status]; notes: Requires report_id in connector config; optional report_status may also be set in connector config.
    metrics historical - Read Freshchat historical metrics [intent=etl availability=implemented stream=historical_metrics]; notes: Requires metrics_metric, metrics_start, and metrics_end in connector config.
    metrics instant - Read Freshchat instant metrics [intent=etl availability=implemented stream=instant_metrics]; notes: Requires metrics_metric in connector config.
  File Commands
    file upload - Upload a Freshchat file [intent=direct_write availability=excluded]; notes: Requires multipart/binary upload policy and is tracked by issue #186; no raw upload command is exposed.
    image upload - Upload a Freshchat image [intent=direct_write availability=excluded]; notes: Requires multipart/binary upload policy and is tracked by issue #186; no raw upload command is exposed.
  Help topics:
    freshchat - Freshchat command metadata maps account, user, conversation, agent, directory, outbound, report, and metrics APIs to safe connector intents.
    freshchat-writes - Freshchat writes remain reverse ETL plan, preview, approval, execute operations; destructive deletes require confirmation.

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
