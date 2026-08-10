# pm connectors inspect chatwoot

```text
NAME
  pm connectors inspect chatwoot - Chatwoot connector manual

SYNOPSIS
  pm connectors inspect chatwoot
  pm connectors inspect chatwoot --json
  pm credentials add <name> --connector chatwoot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes the full account-scoped Chatwoot Application API support-desk surface: conversations, contacts, inboxes, agents, agent bots, teams, labels, conversation-scoped messages, canned responses, custom attribute definitions, custom filters, webhooks, integration hooks, automation rules, help-center portals, inbox membership, and account settings.

ICON
  id: simple-icons-chatwoot
  asset: icons/simple-icons/chatwoot.svg
  title: Chatwoot
  simple_icon_slug: chatwoot
  simple_icon_hex: 1F93FF
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Chatwoot
  match: exact-name-or-slug
  matched_by: chatwoot

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  start_date
  api_access_token (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: updated_at
    fields: account_id(integer), additional_attributes(object), agent_last_seen_at(integer), assignee_last_seen_at(integer), can_reply(boolean), contact_last_seen_at(integer), created_at(integer), custom_attributes(object), first_reply_created_at(integer), id(integer), inbox_id(integer), labels(array), last_activity_at(integer), muted(boolean), priority(string), sla_policy_id(integer), snoozed_until(integer), status(string), timestamp(integer), unread_count(integer), updated_at(integer), uuid(string), waiting_since(integer)
  contacts:
    primary key: id
    cursor: last_activity_at
    fields: additional_attributes(object), availability_status(string), blocked(boolean), contact_inboxes(array), created_at(integer), custom_attributes(object), email(string), id(integer), identifier(string), last_activity_at(integer), name(string), phone_number(string), thumbnail(string)
  inboxes:
    primary key: id
    fields: allow_messages_after_resolved(boolean), avatar_url(string), business_name(string), callback_webhook_url(string), channel_id(integer), channel_type(string), csat_survey_enabled(boolean), enable_auto_assignment(boolean), enable_email_collect(boolean), greeting_enabled(boolean), greeting_message(string), id(integer), lock_to_single_conversation(boolean), medium(string), name(string), out_of_office_message(string), phone_number(string), provider(string), timezone(string), website_token(string), website_url(string), welcome_tagline(string), welcome_title(string), widget_color(string), working_hours_enabled(boolean)
  agents:
    primary key: id
    fields: account_id(integer), auto_offline(boolean), availability_status(string), available_name(string), confirmed(boolean), custom_role_id(integer), email(string), id(integer), name(string), role(string), thumbnail(string)
  teams:
    primary key: id
    fields: account_id(integer), allow_auto_assign(boolean), description(string), id(integer), is_member(boolean), name(string)
  labels:
    primary key: id
    fields: color(string), description(string), id(integer), show_on_sidebar(boolean), title(string)
  messages:
    primary key: id
    cursor: created_at
    fields: account_id(integer), attachment(object), content(string), content_attributes(object), content_type(string), conversation_id(string), created_at(integer), id(integer), inbox_id(integer), message_type(integer), private(boolean), sender(object), sender_id(integer), sender_type(string), source_id(string), status(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  add_inbox_members:
    endpoint: POST /inbox_members
    required fields: inbox_id, user_ids
    risk: medium: add a new agent via the Chatwoot Application API
  add_team_members:
    endpoint: POST /teams/{{ record.team_id }}/team_members
    required fields: team_id, user_ids
    risk: medium: add a new agent via the Chatwoot Application API
  assign_conversation:
    endpoint: POST /conversations/{{ record.conversation_id }}/assignments
    required fields: conversation_id
    risk: medium: assign conversation via the Chatwoot Application API
  create_agent:
    endpoint: POST /agents
    required fields: name, email, role
    risk: medium: add a new agent via the Chatwoot Application API
  create_agent_bot:
    endpoint: POST /agent_bots
    risk: medium: create an agent bot via the Chatwoot Application API
  create_automation_rule:
    endpoint: POST /automation_rules
    risk: medium: add a new automation rule via the Chatwoot Application API
  create_canned_response:
    endpoint: POST /canned_responses
    risk: medium: add a new canned response via the Chatwoot Application API
  create_contact:
    endpoint: POST /contacts
    required fields: inbox_id
    risk: creates a new Chatwoot contact record; low risk, no customer notification
  create_contact_inbox:
    endpoint: POST /contacts/{{ record.id }}/contact_inboxes
    required fields: id, inbox_id
    risk: medium: create contact inbox via the Chatwoot Application API
  create_conversation:
    endpoint: POST /conversations
    required fields: source_id
    risk: creates a new conversation in the target inbox; customer-visible once the initial message is delivered through a live channel
  create_custom_attribute_definition:
    endpoint: POST /custom_attribute_definitions
    risk: medium: add a new custom attribute via the Chatwoot Application API
  create_custom_filter:
    endpoint: POST /custom_filters
    risk: medium: create a custom filter via the Chatwoot Application API
  create_integration_hook:
    endpoint: POST /integrations/hooks
    risk: medium: create an integration hook via the Chatwoot Application API
  create_label:
    endpoint: POST /labels
    required fields: title
    risk: creates a new account-wide label; low risk, visible to all agents in the sidebar when show_on_sidebar is true
  create_portal:
    endpoint: POST /portals
    risk: medium: add a new portal via the Chatwoot Application API
  create_portal_article:
    endpoint: POST /portals/{{ record.id }}/articles
    required fields: id
    risk: medium: add a new article via the Chatwoot Application API
  create_portal_category:
    endpoint: POST /portals/{{ record.id }}/categories
    required fields: id
    risk: medium: add a new category via the Chatwoot Application API
  create_team:
    endpoint: POST /teams
    risk: medium: create a team via the Chatwoot Application API
  create_webhook:
    endpoint: POST /webhooks
    risk: medium: add a webhook via the Chatwoot Application API
  delete_agent:
    endpoint: DELETE /agents/{{ record.id }}
    required fields: id
    risk: high: remove an agent from account via the Chatwoot Application API
  delete_agent_bot:
    endpoint: DELETE /agent_bots/{{ record.id }}
    required fields: id
    risk: high: delete an agentbot via the Chatwoot Application API
  delete_automation_rule:
    endpoint: DELETE /automation_rules/{{ record.id }}
    required fields: id
    risk: high: remove a automation rule from account via the Chatwoot Application API
  delete_canned_response:
    endpoint: DELETE /canned_responses/{{ record.id }}
    required fields: id
    risk: high: remove a canned response from account via the Chatwoot Application API
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: high: delete contact via the Chatwoot Application API
  delete_custom_attribute_definition:
    endpoint: DELETE /custom_attribute_definitions/{{ record.id }}
    required fields: id
    risk: high: remove a custom attribute from account via the Chatwoot Application API
  delete_custom_filter:
    endpoint: DELETE /custom_filters/{{ record.custom_filter_id }}
    required fields: custom_filter_id
    risk: high: delete a custom filter via the Chatwoot Application API
  delete_integration_hook:
    endpoint: DELETE /integrations/hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: high: delete an integration hook via the Chatwoot Application API
  delete_label:
    endpoint: DELETE /labels/{{ record.id }}
    required fields: id
    risk: high: delete a label via the Chatwoot Application API
  delete_message:
    endpoint: DELETE /conversations/{{ record.conversation_id }}/messages/{{ record.message_id }}
    required fields: conversation_id, message_id
    risk: high: delete a message via the Chatwoot Application API
  delete_team:
    endpoint: DELETE /teams/{{ record.team_id }}
    required fields: team_id
    risk: high: delete a team via the Chatwoot Application API
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: high: delete a webhook via the Chatwoot Application API
  destroy_conversation_custom_attributes:
    endpoint: POST /conversations/{{ record.conversation_id }}/destroy_custom_attributes
    required fields: conversation_id, custom_attributes
    risk: medium: destroy custom attributes via the Chatwoot Application API
  merge_contacts:
    endpoint: POST /actions/contact_merge
    required fields: base_contact_id, mergee_contact_id
    risk: high: merge contacts via the Chatwoot Application API
  remove_inbox_members:
    endpoint: DELETE /inbox_members
    required fields: inbox_id, user_ids
    risk: medium: remove an agent from inbox via the Chatwoot Application API
  remove_team_members:
    endpoint: DELETE /teams/{{ record.team_id }}/team_members
    required fields: team_id, user_ids
    risk: medium: remove an agent from team via the Chatwoot Application API
  send_message:
    endpoint: POST /conversations/{{ record.conversation_id }}/messages
    required fields: conversation_id, content
    risk: sends a message into a conversation; customer-visible unless private is true and may notify the contact through the inbox channel
  set_inbox_agent_bot:
    endpoint: POST /inboxes/{{ record.id }}/set_agent_bot
    required fields: id, agent_bot
    risk: medium: add or remove agent bot via the Chatwoot Application API
  set_inbox_members:
    endpoint: PATCH /inbox_members
    required fields: inbox_id, user_ids
    risk: medium: update agents in inbox via the Chatwoot Application API
  set_team_members:
    endpoint: PATCH /teams/{{ record.team_id }}/team_members
    required fields: team_id, user_ids
    risk: medium: update agents in team via the Chatwoot Application API
  toggle_conversation_priority:
    endpoint: POST /conversations/{{ record.conversation_id }}/toggle_priority
    required fields: conversation_id, priority
    risk: medium: toggle priority via the Chatwoot Application API
  toggle_conversation_status:
    endpoint: POST /conversations/{{ record.conversation_id }}/toggle_status
    required fields: conversation_id, status
    risk: changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and reporting metrics
  toggle_conversation_typing_status:
    endpoint: POST /conversations/{{ record.conversation_id }}/toggle_typing_status
    required fields: conversation_id, typing_status
    risk: medium: toggle typing status via the Chatwoot Application API
  update_account:
    endpoint: PATCH /
    risk: medium: update account via the Chatwoot Application API
  update_agent:
    endpoint: PATCH /agents/{{ record.id }}
    required fields: id, role
    risk: medium: update agent in account via the Chatwoot Application API
  update_agent_bot:
    endpoint: PATCH /agent_bots/{{ record.id }}
    required fields: id
    risk: medium: update an agent bot via the Chatwoot Application API
  update_automation_rule:
    endpoint: PATCH /automation_rules/{{ record.id }}
    required fields: id
    risk: medium: update automation rule in account via the Chatwoot Application API
  update_branded_email_layout:
    endpoint: PATCH /branded_email_layout
    risk: medium: update account branded email layout via the Chatwoot Application API
  update_canned_response:
    endpoint: PATCH /canned_responses/{{ record.id }}
    required fields: id
    risk: medium: update canned response in account via the Chatwoot Application API
  update_contact:
    endpoint: PUT /contacts/{{ record.id }}
    required fields: id
    risk: updates an existing Chatwoot contact's profile fields; low risk, no customer notification
  update_contact_labels:
    endpoint: POST /contacts/{{ record.id }}/labels
    required fields: id, labels
    risk: medium: add labels via the Chatwoot Application API
  update_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: medium: update conversation via the Chatwoot Application API
  update_conversation_custom_attributes:
    endpoint: POST /conversations/{{ record.conversation_id }}/custom_attributes
    required fields: conversation_id, custom_attributes
    risk: medium: update custom attributes via the Chatwoot Application API
  update_conversation_labels:
    endpoint: POST /conversations/{{ record.conversation_id }}/labels
    required fields: conversation_id, labels
    risk: medium: add labels via the Chatwoot Application API
  update_custom_attribute_definition:
    endpoint: PATCH /custom_attribute_definitions/{{ record.id }}
    required fields: id
    risk: medium: update custom attribute in account via the Chatwoot Application API
  update_custom_filter:
    endpoint: PATCH /custom_filters/{{ record.custom_filter_id }}
    required fields: custom_filter_id
    risk: medium: update a custom filter via the Chatwoot Application API
  update_integration_hook:
    endpoint: PATCH /integrations/hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: medium: update an integration hook via the Chatwoot Application API
  update_label:
    endpoint: PATCH /labels/{{ record.id }}
    required fields: id
    risk: medium: update a label via the Chatwoot Application API
  update_portal:
    endpoint: PATCH /portals/{{ record.id }}
    required fields: id
    risk: medium: update a portal via the Chatwoot Application API
  update_team:
    endpoint: PATCH /teams/{{ record.team_id }}
    required fields: team_id
    risk: medium: update a team via the Chatwoot Application API
  update_webhook:
    endpoint: PATCH /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: medium: update a webhook object via the Chatwoot Application API

SECURITY
  read risk: external Chatwoot Application API read of conversation, contact, agent, and account-configuration data (account-scoped)
  write risk: external mutation of Chatwoot contacts, conversations, messages, labels, agents, agent bots, teams, automation rules, webhooks, and other account-scoped configuration; agent-visible and customer-visible side effects, including destructive deletes behind typed confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Chatwoot's declared streams and reverse-ETL actions.
  Usage: pm chatwoot <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    account get - Retrieve account settings [intent=direct_read availability=implemented operation=chatwoot.account_get]; flags: --page, --page-cursor
    account get-branded-email-layout - Retrieve the account's branded email layout [intent=direct_read availability=implemented operation=chatwoot.account_get_branded_email_layout]; flags: --page, --page-cursor
    account update - Update account [intent=reverse_etl availability=implemented write=update_account]; approval: requires plan, preview, approval, and execute; risk: medium: update account via the Chatwoot Application API
    account update-branded-email-layout - Update account branded email layout [intent=reverse_etl availability=implemented write=update_branded_email_layout]; approval: requires plan, preview, approval, and execute; risk: medium: update account branded email layout via the Chatwoot Application API
    agent-bots create - Create an Agent Bot [intent=reverse_etl availability=implemented write=create_agent_bot]; approval: requires plan, preview, approval, and execute; risk: medium: create an agent bot via the Chatwoot Application API
    agent-bots delete - Delete an AgentBot [intent=reverse_etl availability=implemented write=delete_agent_bot]; approval: requires plan, preview, approval, and execute; risk: high: delete an agentbot via the Chatwoot Application API; flags: --id (required)
    agent-bots get - Retrieve a single agent bot [intent=direct_read availability=implemented operation=chatwoot.agent_bots_get]; flags: --id (required), --page, --page-cursor
    agent-bots list - List agent bots [intent=direct_read availability=implemented operation=chatwoot.agent_bots_list]; flags: --page, --page-cursor
    agent-bots update - Update an agent bot [intent=reverse_etl availability=implemented write=update_agent_bot]; approval: requires plan, preview, approval, and execute; risk: medium: update an agent bot via the Chatwoot Application API; flags: --id (required)
    agents create - Add a New Agent [intent=reverse_etl availability=implemented write=create_agent]; approval: requires plan, preview, approval, and execute; risk: medium: add a new agent via the Chatwoot Application API; flags: --email (required), --name (required), --role (required)
    agents delete - Remove an Agent from Account [intent=reverse_etl availability=implemented write=delete_agent]; approval: requires plan, preview, approval, and execute; risk: high: remove an agent from account via the Chatwoot Application API; flags: --id (required)
    agents list - List agents as ETL records. [intent=etl availability=implemented stream=agents]
    agents update - Update Agent in Account [intent=reverse_etl availability=implemented write=update_agent]; approval: requires plan, preview, approval, and execute; risk: medium: update agent in account via the Chatwoot Application API; flags: --id (required), --role (required)
    api delete platform api v1 accounts account-id - Documented DELETE /platform/api/v1/accounts/{account_id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.delete.platform-api-v1-accounts-account-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete platform api v1 accounts account-id account-users - Documented DELETE /platform/api/v1/accounts/{account_id}/account_users (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.delete.platform-api-v1-accounts-account-id-account-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete platform api v1 agent-bots id - Documented DELETE /platform/api/v1/agent_bots/{id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.delete.platform-api-v1-agent-bots-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete platform api v1 users id - Documented DELETE /platform/api/v1/users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.delete.platform-api-v1-users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get accounts account-id conversations conversation-id messages - Documented GET /accounts/{account_id}/conversations/{conversation_id}/messages (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.accounts-account-id-conversations-conversation-id-messages]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 profile - Documented GET /api/v1/profile (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v1-profile]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id reports - Documented GET /api/v2/accounts/{account_id}/reports (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-reports]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id reports conversations - Documented GET /api/v2/accounts/{account_id}/reports/conversations (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-reports-conversations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id reports conversations operation - Documented GET /api/v2/accounts/{account_id}/reports/conversations/ (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-reports-conversations-2]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id reports first-response-time-distribution - Documented GET /api/v2/accounts/{account_id}/reports/first_response_time_distribution (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-reports-first-response-time-distribution]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id reports inbox-label-matrix - Documented GET /api/v2/accounts/{account_id}/reports/inbox_label_matrix (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-reports-inbox-label-matrix]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id reports outgoing-messages-count - Documented GET /api/v2/accounts/{account_id}/reports/outgoing_messages_count (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-reports-outgoing-messages-count]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id reports summary - Documented GET /api/v2/accounts/{account_id}/reports/summary (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-reports-summary]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id summary-reports agent - Documented GET /api/v2/accounts/{account_id}/summary_reports/agent (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-summary-reports-agent]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id summary-reports channel - Documented GET /api/v2/accounts/{account_id}/summary_reports/channel (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-summary-reports-channel]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id summary-reports inbox - Documented GET /api/v2/accounts/{account_id}/summary_reports/inbox (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-summary-reports-inbox]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 accounts account-id summary-reports team - Documented GET /api/v2/accounts/{account_id}/summary_reports/team (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.api-v2-accounts-account-id-summary-reports-team]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get platform api v1 accounts account-id - Documented GET /platform/api/v1/accounts/{account_id} (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.platform-api-v1-accounts-account-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get platform api v1 accounts account-id account-users - Documented GET /platform/api/v1/accounts/{account_id}/account_users (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.platform-api-v1-accounts-account-id-account-users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get platform api v1 agent-bots - Documented GET /platform/api/v1/agent_bots (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.platform-api-v1-agent-bots]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get platform api v1 agent-bots id - Documented GET /platform/api/v1/agent_bots/{id} (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.platform-api-v1-agent-bots-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get platform api v1 users id - Documented GET /platform/api/v1/users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.platform-api-v1-users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get platform api v1 users id login - Documented GET /platform/api/v1/users/{id}/login (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.platform-api-v1-users-id-login]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public api v1 inboxes inbox-identifier - Documented GET /public/api/v1/inboxes/{inbox_identifier} (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.public-api-v1-inboxes-inbox-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public api v1 inboxes inbox-identifier contacts contact-identifier - Documented GET /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier} (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public api v1 inboxes inbox-identifier contacts contact-identifier conversations - Documented GET /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public api v1 inboxes inbox-identifier contacts contact-identifier conversations conversation-id - Documented GET /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations/{conversation_id} (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations-conversation-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public api v1 inboxes inbox-identifier contacts contact-identifier conversations conversation-id messages - Documented GET /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations/{conversation_id}/messages (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations-conversation-id-messages]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get survey responses conversation-uuid - Documented GET /survey/responses/{conversation_uuid} (not implemented) [intent=direct_read availability=not_implemented operation=chatwoot.get.survey-responses-conversation-uuid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch api v1 accounts account-id inboxes id - Documented PATCH /api/v1/accounts/{account_id}/inboxes/{id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.patch.api-v1-accounts-account-id-inboxes-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch platform api v1 accounts account-id - Documented PATCH /platform/api/v1/accounts/{account_id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.patch.platform-api-v1-accounts-account-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch platform api v1 agent-bots id - Documented PATCH /platform/api/v1/agent_bots/{id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.patch.platform-api-v1-agent-bots-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch platform api v1 users id - Documented PATCH /platform/api/v1/users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.patch.platform-api-v1-users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch public api v1 inboxes inbox-identifier contacts contact-identifier - Documented PATCH /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.patch.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch public api v1 inboxes inbox-identifier contacts contact-identifier conversations conversation-id messages message-id - Documented PATCH /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations/{conversation_id}/messages/{message_id} (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.patch.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations-conversation-id-messages-message-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 accounts account-id inboxes - Documented POST /api/v1/accounts/{account_id}/inboxes (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.api-v1-accounts-account-id-inboxes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post platform api v1 accounts - Documented POST /platform/api/v1/accounts (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.platform-api-v1-accounts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post platform api v1 accounts account-id account-users - Documented POST /platform/api/v1/accounts/{account_id}/account_users (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.platform-api-v1-accounts-account-id-account-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post platform api v1 agent-bots - Documented POST /platform/api/v1/agent_bots (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.platform-api-v1-agent-bots]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post platform api v1 users - Documented POST /platform/api/v1/users (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.platform-api-v1-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public api v1 inboxes inbox-identifier contacts - Documented POST /public/api/v1/inboxes/{inbox_identifier}/contacts (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.public-api-v1-inboxes-inbox-identifier-contacts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public api v1 inboxes inbox-identifier contacts contact-identifier conversations - Documented POST /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public api v1 inboxes inbox-identifier contacts contact-identifier conversations conversation-id messages - Documented POST /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations/{conversation_id}/messages (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations-conversation-id-messages]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public api v1 inboxes inbox-identifier contacts contact-identifier conversations conversation-id toggle-status - Documented POST /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations/{conversation_id}/toggle_status (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations-conversation-id-toggle-status]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public api v1 inboxes inbox-identifier contacts contact-identifier conversations conversation-id toggle-typing - Documented POST /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations/{conversation_id}/toggle_typing (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations-conversation-id-toggle-typing]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public api v1 inboxes inbox-identifier contacts contact-identifier conversations conversation-id update-last-seen - Documented POST /public/api/v1/inboxes/{inbox_identifier}/contacts/{contact_identifier}/conversations/{conversation_id}/update_last_seen (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.post.public-api-v1-inboxes-inbox-identifier-contacts-contact-identifier-conversations-conversation-id-update-last-seen]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 profile - Documented PUT /api/v1/profile (not implemented) [intent=direct_write availability=not_implemented operation=chatwoot.put.api-v1-profile]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: medium; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    audit-logs list - List account audit log entries [intent=direct_read availability=implemented operation=chatwoot.audit_logs_list]; flags: --page, --page-cursor
    automation-rules create - Add a new automation rule [intent=reverse_etl availability=implemented write=create_automation_rule]; approval: requires plan, preview, approval, and execute; risk: medium: add a new automation rule via the Chatwoot Application API
    automation-rules delete - Remove a automation rule from account [intent=reverse_etl availability=implemented write=delete_automation_rule]; approval: requires plan, preview, approval, and execute; risk: high: remove a automation rule from account via the Chatwoot Application API; flags: --id (required)
    automation-rules get - Retrieve a single automation rule [intent=direct_read availability=implemented operation=chatwoot.automation_rules_get]; flags: --id (required), --page, --page-cursor
    automation-rules list - List automation rules [intent=direct_read availability=implemented operation=chatwoot.automation_rules_list]; flags: --page, --page-cursor
    automation-rules update - Update automation rule in Account [intent=reverse_etl availability=implemented write=update_automation_rule]; approval: requires plan, preview, approval, and execute; risk: medium: update automation rule in account via the Chatwoot Application API; flags: --id (required)
    canned-responses create - Add a New Canned Response [intent=reverse_etl availability=implemented write=create_canned_response]; approval: requires plan, preview, approval, and execute; risk: medium: add a new canned response via the Chatwoot Application API
    canned-responses delete - Remove a Canned Response from Account [intent=reverse_etl availability=implemented write=delete_canned_response]; approval: requires plan, preview, approval, and execute; risk: high: remove a canned response from account via the Chatwoot Application API; flags: --id (required)
    canned-responses list - List canned responses [intent=direct_read availability=implemented operation=chatwoot.canned_responses_list]; flags: --page, --page-cursor
    canned-responses update - Update Canned Response in Account [intent=reverse_etl availability=implemented write=update_canned_response]; approval: requires plan, preview, approval, and execute; risk: medium: update canned response in account via the Chatwoot Application API; flags: --id (required)
    contacts create - Create a new contact [intent=reverse_etl availability=implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: creates a new Chatwoot contact record; low risk, no customer notification; flags: --inbox_id (required)
    contacts create-inbox - Create contact inbox [intent=reverse_etl availability=implemented write=create_contact_inbox]; approval: requires plan, preview, approval, and execute; risk: medium: create contact inbox via the Chatwoot Application API; flags: --id (required), --inbox_id (required)
    contacts delete - Delete Contact [intent=reverse_etl availability=implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: high: delete contact via the Chatwoot Application API; flags: --id (required)
    contacts filter - Advanced filtered contact search [intent=direct_read availability=implemented operation=chatwoot.contacts_filter]; flags: --page, --page-cursor
    contacts get - Retrieve a single contact [intent=direct_read availability=implemented operation=chatwoot.contacts_get]; flags: --id (required), --page, --page-cursor
    contacts list - List contacts as ETL records. [intent=etl availability=implemented stream=contacts]
    contacts list-contactable-inboxes - List inboxes a contact can be contacted through [intent=direct_read availability=implemented operation=chatwoot.contacts_list_contactable_inboxes]; flags: --id (required), --page, --page-cursor
    contacts list-conversations - List conversations for a contact [intent=direct_read availability=implemented operation=chatwoot.contacts_list_conversations]; flags: --id (required), --page, --page-cursor
    contacts list-labels - List labels attached to a contact [intent=direct_read availability=implemented operation=chatwoot.contacts_list_labels]; flags: --id (required), --page, --page-cursor
    contacts merge - Merge Contacts [intent=reverse_etl availability=implemented write=merge_contacts]; approval: requires plan, preview, approval, and execute; risk: high: merge contacts via the Chatwoot Application API; flags: --base_contact_id (required), --mergee_contact_id (required)
    contacts search - Search contacts by name/email/phone [intent=direct_read availability=implemented operation=chatwoot.contacts_search]; flags: --q, --sort, --page, --page-cursor
    contacts update - Update an existing contact [intent=reverse_etl availability=implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: updates an existing Chatwoot contact's profile fields; low risk, no customer notification; flags: --id (required)
    contacts update-labels - Add Labels [intent=reverse_etl availability=implemented write=update_contact_labels]; approval: requires plan, preview, approval, and execute; risk: medium: add labels via the Chatwoot Application API; flags: --id (required), --labels (required)
    conversations assign - Assign Conversation [intent=reverse_etl availability=implemented write=assign_conversation]; approval: requires plan, preview, approval, and execute; risk: medium: assign conversation via the Chatwoot Application API; flags: --conversation_id (required)
    conversations create - Create a new conversation [intent=reverse_etl availability=implemented write=create_conversation]; approval: requires plan, preview, approval, and execute; risk: creates a new conversation in the target inbox; customer-visible once the initial message is delivered through a live channel; flags: --source_id (required)
    conversations destroy-custom-attributes - Destroy Custom Attributes [intent=reverse_etl availability=implemented write=destroy_conversation_custom_attributes]; approval: requires plan, preview, approval, and execute; risk: medium: destroy custom attributes via the Chatwoot Application API; flags: --conversation_id (required), --custom_attributes (required)
    conversations filter - Advanced filtered conversation search [intent=direct_read availability=implemented operation=chatwoot.conversations_filter]; flags: --page, --page-cursor
    conversations get - Retrieve a single conversation [intent=direct_read availability=implemented operation=chatwoot.conversations_get]; flags: --conversation-id (required), --page, --page-cursor
    conversations list - List conversations as ETL records. [intent=etl availability=implemented stream=conversations]
    conversations list-labels - List labels attached to a conversation [intent=direct_read availability=implemented operation=chatwoot.conversations_list_labels]; flags: --conversation-id (required), --page, --page-cursor
    conversations list-reporting-events - List reporting events for a conversation [intent=direct_read availability=implemented operation=chatwoot.conversations_list_reporting_events]; flags: --conversation-id (required), --page, --page-cursor
    conversations meta - Retrieve conversation count summary (mine/unassigned/all) [intent=direct_read availability=implemented operation=chatwoot.conversations_meta]; flags: --status, --q, --inbox-id, --team-id, --labels, --page, --page-cursor
    conversations toggle-priority - Toggle Priority [intent=reverse_etl availability=implemented write=toggle_conversation_priority]; approval: requires plan, preview, approval, and execute; risk: medium: toggle priority via the Chatwoot Application API; flags: --conversation_id (required), --priority (required)
    conversations toggle-status - Toggle a conversation's status [intent=reverse_etl availability=implemented write=toggle_conversation_status]; approval: requires plan, preview, approval, and execute; risk: changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and reporting metrics; flags: --conversation_id (required), --status (required)
    conversations toggle-typing-status - Toggle Typing Status [intent=reverse_etl availability=implemented write=toggle_conversation_typing_status]; approval: requires plan, preview, approval, and execute; risk: medium: toggle typing status via the Chatwoot Application API; flags: --conversation_id (required), --typing_status (required)
    conversations update - Update Conversation [intent=reverse_etl availability=implemented write=update_conversation]; approval: requires plan, preview, approval, and execute; risk: medium: update conversation via the Chatwoot Application API; flags: --conversation_id (required)
    conversations update-custom-attributes - Update Custom Attributes [intent=reverse_etl availability=partial write=update_conversation_custom_attributes]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update custom attributes via the Chatwoot Application API; notes: partial: `custom_attributes` is a required, arbitrary-keyed JSON object (the API's own record_schema has no fixed property set); flat CLI flags cannot express it. Reverse ETL plan JSON (--config/plan-file record override) is the supported path.; flags: --conversation-id (required), --merge
    conversations update-labels - Add Labels [intent=reverse_etl availability=implemented write=update_conversation_labels]; approval: requires plan, preview, approval, and execute; risk: medium: add labels via the Chatwoot Application API; flags: --conversation_id (required), --labels (required)
    custom-attributes create - Add a new custom attribute [intent=reverse_etl availability=implemented write=create_custom_attribute_definition]; approval: requires plan, preview, approval, and execute; risk: medium: add a new custom attribute via the Chatwoot Application API
    custom-attributes delete - Remove a custom attribute from account [intent=reverse_etl availability=implemented write=delete_custom_attribute_definition]; approval: requires plan, preview, approval, and execute; risk: high: remove a custom attribute from account via the Chatwoot Application API; flags: --id (required)
    custom-attributes get - Retrieve a single custom attribute definition [intent=direct_read availability=implemented operation=chatwoot.custom_attributes_get]; flags: --id (required), --page, --page-cursor
    custom-attributes list - List custom attribute definitions [intent=direct_read availability=implemented operation=chatwoot.custom_attributes_list]; flags: --attribute-model, --page, --page-cursor
    custom-attributes update - Update custom attribute in Account [intent=reverse_etl availability=implemented write=update_custom_attribute_definition]; approval: requires plan, preview, approval, and execute; risk: medium: update custom attribute in account via the Chatwoot Application API; flags: --id (required)
    custom-filters create - Create a custom filter [intent=reverse_etl availability=implemented write=create_custom_filter]; approval: requires plan, preview, approval, and execute; risk: medium: create a custom filter via the Chatwoot Application API
    custom-filters delete - Delete a custom filter [intent=reverse_etl availability=implemented write=delete_custom_filter]; approval: requires plan, preview, approval, and execute; risk: high: delete a custom filter via the Chatwoot Application API; flags: --custom_filter_id (required)
    custom-filters get - Retrieve a single custom filter [intent=direct_read availability=implemented operation=chatwoot.custom_filters_get]; flags: --custom-filter-id (required), --page, --page-cursor
    custom-filters list - List saved custom filters [intent=direct_read availability=implemented operation=chatwoot.custom_filters_list]; flags: --filter-type, --page, --page-cursor
    custom-filters update - Update a custom filter [intent=reverse_etl availability=implemented write=update_custom_filter]; approval: requires plan, preview, approval, and execute; risk: medium: update a custom filter via the Chatwoot Application API; flags: --custom_filter_id (required)
    inbox-members add - Add a New Agent [intent=reverse_etl availability=not_implemented write=add_inbox_members]; approval: requires plan, preview, approval, and execute; risk: medium: add a new agent via the Chatwoot Application API; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    inbox-members list - List agents who are members of an inbox [intent=direct_read availability=implemented operation=chatwoot.inbox_members_list]; flags: --inbox-id (required), --page, --page-cursor
    inbox-members remove - Remove an Agent from Inbox [intent=reverse_etl availability=not_implemented write=remove_inbox_members]; approval: requires plan, preview, approval, and execute; risk: medium: remove an agent from inbox via the Chatwoot Application API; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    inbox-members set - Update Agents in Inbox [intent=reverse_etl availability=not_implemented write=set_inbox_members]; approval: requires plan, preview, approval, and execute; risk: medium: update agents in inbox via the Chatwoot Application API; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    inboxes get - Retrieve a single inbox [intent=direct_read availability=implemented operation=chatwoot.inboxes_get]; flags: --id (required), --page, --page-cursor
    inboxes get-agent-bot - Retrieve the agent bot attached to an inbox [intent=direct_read availability=implemented operation=chatwoot.inboxes_get_agent_bot]; flags: --id (required), --page, --page-cursor
    inboxes list - List inboxes as ETL records. [intent=etl availability=implemented stream=inboxes]
    inboxes list-message-templates - List WhatsApp message templates for an inbox [intent=direct_read availability=implemented operation=chatwoot.inboxes_list_message_templates]; flags: --id (required), --name, --page, --page-cursor
    inboxes set-agent-bot - Add or remove agent bot [intent=reverse_etl availability=implemented write=set_inbox_agent_bot]; approval: requires plan, preview, approval, and execute; risk: medium: add or remove agent bot via the Chatwoot Application API; flags: --agent_bot (required), --id (required)
    integrations create-hook - Create an integration hook [intent=reverse_etl availability=implemented write=create_integration_hook]; approval: requires plan, preview, approval, and execute; risk: medium: create an integration hook via the Chatwoot Application API
    integrations delete-hook - Delete an Integration Hook [intent=reverse_etl availability=implemented write=delete_integration_hook]; approval: requires plan, preview, approval, and execute; risk: high: delete an integration hook via the Chatwoot Application API; flags: --hook_id (required)
    integrations list-apps - List installed integrations [intent=direct_read availability=implemented operation=chatwoot.integrations_list_apps]; flags: --page, --page-cursor
    integrations update-hook - Update an Integration Hook [intent=reverse_etl availability=implemented write=update_integration_hook]; approval: requires plan, preview, approval, and execute; risk: medium: update an integration hook via the Chatwoot Application API; flags: --hook_id (required)
    labels create - Create a new label [intent=reverse_etl availability=implemented write=create_label]; approval: requires plan, preview, approval, and execute; risk: creates a new account-wide label; low risk, visible to all agents in the sidebar when show_on_sidebar is true; flags: --title (required)
    labels delete - Delete a label [intent=reverse_etl availability=implemented write=delete_label]; approval: requires plan, preview, approval, and execute; risk: high: delete a label via the Chatwoot Application API; flags: --id (required)
    labels get - Retrieve a single label [intent=direct_read availability=implemented operation=chatwoot.labels_get]; flags: --id (required), --page, --page-cursor
    labels list - List labels as ETL records. [intent=etl availability=implemented stream=labels]
    labels update - Update a label [intent=reverse_etl availability=implemented write=update_label]; approval: requires plan, preview, approval, and execute; risk: medium: update a label via the Chatwoot Application API; flags: --id (required)
    messages delete - Delete a message [intent=reverse_etl availability=implemented write=delete_message]; approval: requires plan, preview, approval, and execute; risk: high: delete a message via the Chatwoot Application API; flags: --conversation_id (required), --message_id (required)
    messages list - List conversation-scoped messages as ETL records, fanned out across every conversation. [intent=etl availability=implemented stream=messages]
    messages send - Send a message in a conversation [intent=reverse_etl availability=implemented write=send_message]; approval: requires plan, preview, approval, and execute; risk: sends a message into a conversation; customer-visible unless private is true and may notify the contact through the inbox channel; flags: --content (required), --conversation_id (required)
    portals create - Add a new portal [intent=reverse_etl availability=implemented write=create_portal]; approval: requires plan, preview, approval, and execute; risk: medium: add a new portal via the Chatwoot Application API
    portals create-article - Add a new article [intent=reverse_etl availability=implemented write=create_portal_article]; approval: requires plan, preview, approval, and execute; risk: medium: add a new article via the Chatwoot Application API; flags: --id (required)
    portals create-category - Add a new category [intent=reverse_etl availability=implemented write=create_portal_category]; approval: requires plan, preview, approval, and execute; risk: medium: add a new category via the Chatwoot Application API; flags: --id (required)
    portals list - List help-center portals [intent=direct_read availability=implemented operation=chatwoot.portals_list]; flags: --page, --page-cursor
    portals update - Update a portal [intent=reverse_etl availability=implemented write=update_portal]; approval: requires plan, preview, approval, and execute; risk: medium: update a portal via the Chatwoot Application API; flags: --id (required)
    reporting-events list - List account-wide reporting events [intent=direct_read availability=implemented operation=chatwoot.reporting_events_list]; flags: --since, --until, --inbox-id, --user-id, --name, --page, --page-cursor
    teams add-members - Add a New Agent [intent=reverse_etl availability=not_implemented write=add_team_members]; approval: requires plan, preview, approval, and execute; risk: medium: add a new agent via the Chatwoot Application API; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    teams create - Create a team [intent=reverse_etl availability=implemented write=create_team]; approval: requires plan, preview, approval, and execute; risk: medium: create a team via the Chatwoot Application API
    teams delete - Delete a team [intent=reverse_etl availability=implemented write=delete_team]; approval: requires plan, preview, approval, and execute; risk: high: delete a team via the Chatwoot Application API; flags: --team_id (required)
    teams get - Retrieve a single team [intent=direct_read availability=implemented operation=chatwoot.teams_get]; flags: --team-id (required), --page, --page-cursor
    teams list - List teams as ETL records. [intent=etl availability=implemented stream=teams]
    teams list-members - List a team's member agents [intent=direct_read availability=implemented operation=chatwoot.teams_list_members]; flags: --team-id (required), --page, --page-cursor
    teams remove-members - Remove an Agent from Team [intent=reverse_etl availability=not_implemented write=remove_team_members]; approval: requires plan, preview, approval, and execute; risk: medium: remove an agent from team via the Chatwoot Application API; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    teams set-members - Update Agents in Team [intent=reverse_etl availability=not_implemented write=set_team_members]; approval: requires plan, preview, approval, and execute; risk: medium: update agents in team via the Chatwoot Application API; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    teams update - Update a team [intent=reverse_etl availability=implemented write=update_team]; approval: requires plan, preview, approval, and execute; risk: medium: update a team via the Chatwoot Application API; flags: --team_id (required)
    webhooks create - Add a webhook [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: medium: add a webhook via the Chatwoot Application API
    webhooks delete - Delete a webhook [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: high: delete a webhook via the Chatwoot Application API; flags: --webhook_id (required)
    webhooks list - List webhook subscriptions [intent=direct_read availability=implemented operation=chatwoot.webhooks_list]; flags: --page, --page-cursor
    webhooks update - Update a webhook object [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: medium: update a webhook object via the Chatwoot Application API; flags: --webhook_id (required)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect chatwoot

  # Inspect as structured JSON
  pm connectors inspect chatwoot --json

AGENT WORKFLOW
  - Run pm connectors inspect chatwoot before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
