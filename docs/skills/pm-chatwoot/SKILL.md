---
name: pm-chatwoot
description: Chatwoot connector knowledge and safe action guide.
---

# pm-chatwoot

## Purpose

Reads and writes the full account-scoped Chatwoot Application API support-desk surface: conversations, contacts, inboxes, agents, agent bots, teams, labels, conversation-scoped messages, canned responses, custom attribute definitions, custom filters, webhooks, integration hooks, automation rules, help-center portals, inbox membership, and account settings.

## Icon

- id: simple-icons-chatwoot
- asset: icons/simple-icons/chatwoot.svg
- title: Chatwoot
- simple_icon_slug: chatwoot
- simple_icon_hex: 1F93FF
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Chatwoot
- match: exact-name-or-slug
- matched_by: chatwoot

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id (required)
- base_url (required)
- start_date
- api_access_token (secret)

## ETL Streams

- conversations:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(integer), additional_attributes(object), agent_last_seen_at(integer), assignee_last_seen_at(integer), can_reply(boolean), contact_last_seen_at(integer), created_at(integer), custom_attributes(object), first_reply_created_at(integer), id(integer), inbox_id(integer), labels(array), last_activity_at(integer), muted(boolean), priority(string), sla_policy_id(integer), snoozed_until(integer), status(string), timestamp(integer), unread_count(integer), updated_at(integer), uuid(string), waiting_since(integer)
- contacts:
  - primary key: id
  - cursor: last_activity_at
  - fields: additional_attributes(object), availability_status(string), blocked(boolean), contact_inboxes(array), created_at(integer), custom_attributes(object), email(string), id(integer), identifier(string), last_activity_at(integer), name(string), phone_number(string), thumbnail(string)
- inboxes:
  - primary key: id
  - fields: allow_messages_after_resolved(boolean), avatar_url(string), business_name(string), callback_webhook_url(string), channel_id(integer), channel_type(string), csat_survey_enabled(boolean), enable_auto_assignment(boolean), enable_email_collect(boolean), greeting_enabled(boolean), greeting_message(string), id(integer), lock_to_single_conversation(boolean), medium(string), name(string), out_of_office_message(string), phone_number(string), provider(string), timezone(string), website_token(string), website_url(string), welcome_tagline(string), welcome_title(string), widget_color(string), working_hours_enabled(boolean)
- agents:
  - primary key: id
  - fields: account_id(integer), auto_offline(boolean), availability_status(string), available_name(string), confirmed(boolean), custom_role_id(integer), email(string), id(integer), name(string), role(string), thumbnail(string)
- teams:
  - primary key: id
  - fields: account_id(integer), allow_auto_assign(boolean), description(string), id(integer), is_member(boolean), name(string)
- labels:
  - primary key: id
  - fields: color(string), description(string), id(integer), show_on_sidebar(boolean), title(string)
- messages:
  - primary key: id
  - cursor: created_at
  - fields: account_id(integer), attachment(object), content(string), content_attributes(object), content_type(string), conversation_id(string), created_at(integer), id(integer), inbox_id(integer), message_type(integer), private(boolean), sender(object), sender_id(integer), sender_type(string), source_id(string), status(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- add_inbox_members:
  - endpoint: POST /inbox_members
  - required fields: inbox_id, user_ids
  - risk: medium: add a new agent via the Chatwoot Application API
- add_team_members:
  - endpoint: POST /teams/{{ record.team_id }}/team_members
  - required fields: team_id, user_ids
  - risk: medium: add a new agent via the Chatwoot Application API
- assign_conversation:
  - endpoint: POST /conversations/{{ record.conversation_id }}/assignments
  - required fields: conversation_id
  - risk: medium: assign conversation via the Chatwoot Application API
- create_agent:
  - endpoint: POST /agents
  - required fields: name, email, role
  - risk: medium: add a new agent via the Chatwoot Application API
- create_agent_bot:
  - endpoint: POST /agent_bots
  - risk: medium: create an agent bot via the Chatwoot Application API
- create_automation_rule:
  - endpoint: POST /automation_rules
  - risk: medium: add a new automation rule via the Chatwoot Application API
- create_canned_response:
  - endpoint: POST /canned_responses
  - risk: medium: add a new canned response via the Chatwoot Application API
- create_contact:
  - endpoint: POST /contacts
  - required fields: inbox_id
  - risk: creates a new Chatwoot contact record; low risk, no customer notification
- create_contact_inbox:
  - endpoint: POST /contacts/{{ record.id }}/contact_inboxes
  - required fields: id, inbox_id
  - risk: medium: create contact inbox via the Chatwoot Application API
- create_conversation:
  - endpoint: POST /conversations
  - required fields: source_id
  - risk: creates a new conversation in the target inbox; customer-visible once the initial message is delivered through a live channel
- create_custom_attribute_definition:
  - endpoint: POST /custom_attribute_definitions
  - risk: medium: add a new custom attribute via the Chatwoot Application API
- create_custom_filter:
  - endpoint: POST /custom_filters
  - risk: medium: create a custom filter via the Chatwoot Application API
- create_integration_hook:
  - endpoint: POST /integrations/hooks
  - risk: medium: create an integration hook via the Chatwoot Application API
- create_label:
  - endpoint: POST /labels
  - required fields: title
  - risk: creates a new account-wide label; low risk, visible to all agents in the sidebar when show_on_sidebar is true
- create_portal:
  - endpoint: POST /portals
  - risk: medium: add a new portal via the Chatwoot Application API
- create_portal_article:
  - endpoint: POST /portals/{{ record.id }}/articles
  - required fields: id
  - risk: medium: add a new article via the Chatwoot Application API
- create_portal_category:
  - endpoint: POST /portals/{{ record.id }}/categories
  - required fields: id
  - risk: medium: add a new category via the Chatwoot Application API
- create_team:
  - endpoint: POST /teams
  - risk: medium: create a team via the Chatwoot Application API
- create_webhook:
  - endpoint: POST /webhooks
  - risk: medium: add a webhook via the Chatwoot Application API
- delete_agent:
  - endpoint: DELETE /agents/{{ record.id }}
  - required fields: id
  - risk: high: remove an agent from account via the Chatwoot Application API
- delete_agent_bot:
  - endpoint: DELETE /agent_bots/{{ record.id }}
  - required fields: id
  - risk: high: delete an agentbot via the Chatwoot Application API
- delete_automation_rule:
  - endpoint: DELETE /automation_rules/{{ record.id }}
  - required fields: id
  - risk: high: remove a automation rule from account via the Chatwoot Application API
- delete_canned_response:
  - endpoint: DELETE /canned_responses/{{ record.id }}
  - required fields: id
  - risk: high: remove a canned response from account via the Chatwoot Application API
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: high: delete contact via the Chatwoot Application API
- delete_custom_attribute_definition:
  - endpoint: DELETE /custom_attribute_definitions/{{ record.id }}
  - required fields: id
  - risk: high: remove a custom attribute from account via the Chatwoot Application API
- delete_custom_filter:
  - endpoint: DELETE /custom_filters/{{ record.custom_filter_id }}
  - required fields: custom_filter_id
  - risk: high: delete a custom filter via the Chatwoot Application API
- delete_integration_hook:
  - endpoint: DELETE /integrations/hooks/{{ record.hook_id }}
  - required fields: hook_id
  - risk: high: delete an integration hook via the Chatwoot Application API
- delete_label:
  - endpoint: DELETE /labels/{{ record.id }}
  - required fields: id
  - risk: high: delete a label via the Chatwoot Application API
- delete_message:
  - endpoint: DELETE /conversations/{{ record.conversation_id }}/messages/{{ record.message_id }}
  - required fields: conversation_id, message_id
  - risk: high: delete a message via the Chatwoot Application API
- delete_team:
  - endpoint: DELETE /teams/{{ record.team_id }}
  - required fields: team_id
  - risk: high: delete a team via the Chatwoot Application API
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.webhook_id }}
  - required fields: webhook_id
  - risk: high: delete a webhook via the Chatwoot Application API
- destroy_conversation_custom_attributes:
  - endpoint: POST /conversations/{{ record.conversation_id }}/destroy_custom_attributes
  - required fields: conversation_id, custom_attributes
  - risk: medium: destroy custom attributes via the Chatwoot Application API
- merge_contacts:
  - endpoint: POST /actions/contact_merge
  - required fields: base_contact_id, mergee_contact_id
  - risk: high: merge contacts via the Chatwoot Application API
- remove_inbox_members:
  - endpoint: DELETE /inbox_members
  - required fields: inbox_id, user_ids
  - risk: medium: remove an agent from inbox via the Chatwoot Application API
- remove_team_members:
  - endpoint: DELETE /teams/{{ record.team_id }}/team_members
  - required fields: team_id, user_ids
  - risk: medium: remove an agent from team via the Chatwoot Application API
- send_message:
  - endpoint: POST /conversations/{{ record.conversation_id }}/messages
  - required fields: conversation_id, content
  - risk: sends a message into a conversation; customer-visible unless private is true and may notify the contact through the inbox channel
- set_inbox_agent_bot:
  - endpoint: POST /inboxes/{{ record.id }}/set_agent_bot
  - required fields: id, agent_bot
  - risk: medium: add or remove agent bot via the Chatwoot Application API
- set_inbox_members:
  - endpoint: PATCH /inbox_members
  - required fields: inbox_id, user_ids
  - risk: medium: update agents in inbox via the Chatwoot Application API
- set_team_members:
  - endpoint: PATCH /teams/{{ record.team_id }}/team_members
  - required fields: team_id, user_ids
  - risk: medium: update agents in team via the Chatwoot Application API
- toggle_conversation_priority:
  - endpoint: POST /conversations/{{ record.conversation_id }}/toggle_priority
  - required fields: conversation_id, priority
  - risk: medium: toggle priority via the Chatwoot Application API
- toggle_conversation_status:
  - endpoint: POST /conversations/{{ record.conversation_id }}/toggle_status
  - required fields: conversation_id, status
  - risk: changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and reporting metrics
- toggle_conversation_typing_status:
  - endpoint: POST /conversations/{{ record.conversation_id }}/toggle_typing_status
  - required fields: conversation_id, typing_status
  - risk: medium: toggle typing status via the Chatwoot Application API
- update_account:
  - endpoint: PATCH /
  - risk: medium: update account via the Chatwoot Application API
- update_agent:
  - endpoint: PATCH /agents/{{ record.id }}
  - required fields: id, role
  - risk: medium: update agent in account via the Chatwoot Application API
- update_agent_bot:
  - endpoint: PATCH /agent_bots/{{ record.id }}
  - required fields: id
  - risk: medium: update an agent bot via the Chatwoot Application API
- update_automation_rule:
  - endpoint: PATCH /automation_rules/{{ record.id }}
  - required fields: id
  - risk: medium: update automation rule in account via the Chatwoot Application API
- update_branded_email_layout:
  - endpoint: PATCH /branded_email_layout
  - risk: medium: update account branded email layout via the Chatwoot Application API
- update_canned_response:
  - endpoint: PATCH /canned_responses/{{ record.id }}
  - required fields: id
  - risk: medium: update canned response in account via the Chatwoot Application API
- update_contact:
  - endpoint: PUT /contacts/{{ record.id }}
  - required fields: id
  - risk: updates an existing Chatwoot contact's profile fields; low risk, no customer notification
- update_contact_labels:
  - endpoint: POST /contacts/{{ record.id }}/labels
  - required fields: id, labels
  - risk: medium: add labels via the Chatwoot Application API
- update_conversation:
  - endpoint: PATCH /conversations/{{ record.conversation_id }}
  - required fields: conversation_id
  - risk: medium: update conversation via the Chatwoot Application API
- update_conversation_custom_attributes:
  - endpoint: POST /conversations/{{ record.conversation_id }}/custom_attributes
  - required fields: conversation_id, custom_attributes
  - risk: medium: update custom attributes via the Chatwoot Application API
- update_conversation_labels:
  - endpoint: POST /conversations/{{ record.conversation_id }}/labels
  - required fields: conversation_id, labels
  - risk: medium: add labels via the Chatwoot Application API
- update_custom_attribute_definition:
  - endpoint: PATCH /custom_attribute_definitions/{{ record.id }}
  - required fields: id
  - risk: medium: update custom attribute in account via the Chatwoot Application API
- update_custom_filter:
  - endpoint: PATCH /custom_filters/{{ record.custom_filter_id }}
  - required fields: custom_filter_id
  - risk: medium: update a custom filter via the Chatwoot Application API
- update_integration_hook:
  - endpoint: PATCH /integrations/hooks/{{ record.hook_id }}
  - required fields: hook_id
  - risk: medium: update an integration hook via the Chatwoot Application API
- update_label:
  - endpoint: PATCH /labels/{{ record.id }}
  - required fields: id
  - risk: medium: update a label via the Chatwoot Application API
- update_portal:
  - endpoint: PATCH /portals/{{ record.id }}
  - required fields: id
  - risk: medium: update a portal via the Chatwoot Application API
- update_team:
  - endpoint: PATCH /teams/{{ record.team_id }}
  - required fields: team_id
  - risk: medium: update a team via the Chatwoot Application API
- update_webhook:
  - endpoint: PATCH /webhooks/{{ record.webhook_id }}
  - required fields: webhook_id
  - risk: medium: update a webhook object via the Chatwoot Application API

## Security

- read risk: external Chatwoot Application API read of conversation, contact, agent, and account-configuration data (account-scoped)
- write risk: external mutation of Chatwoot contacts, conversations, messages, labels, agents, agent bots, teams, automation rules, webhooks, and other account-scoped configuration; agent-visible and customer-visible side effects, including destructive deletes behind typed confirmation
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Sync Chatwoot conversations, contacts, inboxes, agents, teams and labels, and manage the full Chatwoot Application API support-desk surface from the command line.
- Usage: pm chatwoot <command> [flags]
- Source CLI: Chatwoot API (https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json)
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Account
  - account get - Retrieve account settings [intent=direct_read availability=implemented operation=chatwoot.account_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - account get-branded-email-layout - Retrieve the account's branded email layout [intent=direct_read availability=implemented operation=chatwoot.account_get_branded_email_layout]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - account update - Update account [intent=reverse_etl availability=implemented write=update_account]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update account via the Chatwoot Application API; flags: --name (string): Name of the account: maps_to=record.name, --locale (string): The locale of the account: maps_to=record.locale, --domain (string): The domain of the account: maps_to=record.domain, --support-email (string): The support email of the account: maps_to=record.support_email, --auto-resolve-after (integer): Auto resolve conversations after specified minutes: maps_to=record.auto_resolve_after, --auto-resolve-message (string): Message to send when auto resolving: maps_to=record.auto_resolve_message, --auto-resolve-ignore-waiting (boolean): Whether to ignore waiting conversations for auto resolve: maps_to=record.auto_resolve_ignore_waiting, --industry (string): Industry type: maps_to=record.industry, --company-size (string): Company size: maps_to=record.company_size, --timezone (string): Account timezone: maps_to=record.timezone
  - account update-branded-email-layout - Update account branded email layout [intent=reverse_etl availability=implemented write=update_branded_email_layout]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update account branded email layout via the Chatwoot Application API; flags: --branded-email-layout (string): Account-scoped Liquid HTML layout for branded email replies. Maximum length is 262,144 characters. Blank or null removes the account override.: maps_to=record.branded_email_layout
- Agent Bots
  - agent-bots create - Create an Agent Bot [intent=reverse_etl availability=implemented write=create_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: create an agent bot via the Chatwoot Application API; flags: --name (string): The name of the agent bot: maps_to=record.name, --description (string): The description of the agent bot: maps_to=record.description, --outgoing-url (string): The webhook URL for the bot: maps_to=record.outgoing_url, --avatar (string): Send the form data with the avatar image binary or use the avatar_url: maps_to=record.avatar, --avatar-url (string): The url to a jpeg, png file for the agent bot avatar: maps_to=record.avatar_url, --bot-type (integer): The type of the bot (0 for webhook): maps_to=record.bot_type
  - agent-bots delete - Delete an AgentBot [intent=reverse_etl availability=implemented write=delete_agent_bot]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete an agentbot via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id
  - agent-bots get - Retrieve a single agent bot [intent=direct_read availability=implemented operation=chatwoot.agent_bots_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - agent-bots list - List agent bots [intent=direct_read availability=implemented operation=chatwoot.agent_bots_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - agent-bots update - Update an agent bot [intent=reverse_etl availability=implemented write=update_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update an agent bot via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --name (string): The name of the agent bot: maps_to=record.name, --description (string): The description of the agent bot: maps_to=record.description, --outgoing-url (string): The webhook URL for the bot: maps_to=record.outgoing_url, --avatar (string): Send the form data with the avatar image binary or use the avatar_url: maps_to=record.avatar, --avatar-url (string): The url to a jpeg, png file for the agent bot avatar: maps_to=record.avatar_url, --bot-type (integer): The type of the bot (0 for webhook): maps_to=record.bot_type
- Agents
  - agents create - Add a New Agent [intent=reverse_etl availability=implemented write=create_agent]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new agent via the Chatwoot Application API; flags: --name (required) (string): Full Name of the agent: maps_to=record.name, --email (required) (string): Email of the Agent: maps_to=record.email, --role (required) (string): Whether its administrator or agent: maps_to=record.role, --availability (string): The configured availability of the agent.: maps_to=record.availability, --auto-offline (boolean): Whether the agent is automatically marked offline when they are away.: maps_to=record.auto_offline
  - agents delete - Remove an Agent from Account [intent=reverse_etl availability=implemented write=delete_agent]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: remove an agent from account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id
  - agents list - List agents as ETL records. [intent=etl availability=implemented stream=agents]
  - agents update - Update Agent in Account [intent=reverse_etl availability=implemented write=update_agent]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update agent in account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --role (required) (string): Whether its administrator or agent: maps_to=record.role, --availability (string): The configured availability of the agent.: maps_to=record.availability, --auto-offline (boolean): Whether the agent is automatically marked offline when they are away.: maps_to=record.auto_offline
- Audit Logs
  - audit-logs list - List account audit log entries [intent=direct_read availability=implemented operation=chatwoot.audit_logs_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
- Automation Rules
  - automation-rules create - Add a new automation rule [intent=reverse_etl availability=implemented write=create_automation_rule]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new automation rule via the Chatwoot Application API; flags: --name (string): Rule name: maps_to=record.name, --description (string): The description about the automation and actions: maps_to=record.description, --event-name (string): The event when you want to execute the automation actions: maps_to=record.event_name, --active (boolean): Enable/disable automation rule: maps_to=record.active
  - automation-rules delete - Remove a automation rule from account [intent=reverse_etl availability=implemented write=delete_automation_rule]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: remove a automation rule from account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id
  - automation-rules get - Retrieve a single automation rule [intent=direct_read availability=implemented operation=chatwoot.automation_rules_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - automation-rules list - List automation rules [intent=direct_read availability=implemented operation=chatwoot.automation_rules_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - automation-rules update - Update automation rule in Account [intent=reverse_etl availability=implemented write=update_automation_rule]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update automation rule in account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --name (string): Rule name: maps_to=record.name, --description (string): The description about the automation and actions: maps_to=record.description, --event-name (string): The event when you want to execute the automation actions: maps_to=record.event_name, --active (boolean): Enable/disable automation rule: maps_to=record.active
- Canned Responses
  - canned-responses create - Add a New Canned Response [intent=reverse_etl availability=implemented write=create_canned_response]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new canned response via the Chatwoot Application API; flags: --content (string): Message content for canned response: maps_to=record.content, --short-code (string): Short Code for quick access of the canned response: maps_to=record.short_code
  - canned-responses delete - Remove a Canned Response from Account [intent=reverse_etl availability=implemented write=delete_canned_response]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: remove a canned response from account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id
  - canned-responses list - List canned responses [intent=direct_read availability=implemented operation=chatwoot.canned_responses_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - canned-responses update - Update Canned Response in Account [intent=reverse_etl availability=implemented write=update_canned_response]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update canned response in account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --content (string): Message content for canned response: maps_to=record.content, --short-code (string): Short Code for quick access of the canned response: maps_to=record.short_code
- Contacts
  - contacts create - Create a new contact [intent=reverse_etl availability=implemented write=create_contact]; approval: reverse ETL writes require plan, preview, approval, execute; risk: creates a new Chatwoot contact record; low risk, no customer notification; flags: --inbox-id (required) (integer): Field `inbox_id`.: maps_to=record.inbox_id, --name (string): Field `name`.: maps_to=record.name, --email (string): Field `email`.: maps_to=record.email, --phone-number (string): Field `phone_number`.: maps_to=record.phone_number, --identifier (string): Field `identifier`.: maps_to=record.identifier, --blocked (boolean): Field `blocked`.: maps_to=record.blocked, --avatar-url (string): Field `avatar_url`.: maps_to=record.avatar_url
  - contacts create-inbox - Create contact inbox [intent=reverse_etl availability=implemented write=create_contact_inbox]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: create contact inbox via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --inbox-id (required) (number): The ID of the inbox: maps_to=record.inbox_id, --source-id (string): Contact Inbox Source Id: maps_to=record.source_id
  - contacts delete - Delete Contact [intent=reverse_etl availability=implemented write=delete_contact]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete contact via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id
  - contacts filter - Advanced filtered contact search [intent=direct_read availability=implemented operation=chatwoot.contacts_filter]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - contacts get - Retrieve a single contact [intent=direct_read availability=implemented operation=chatwoot.contacts_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - contacts list - List contacts as ETL records. [intent=etl availability=implemented stream=contacts]
  - contacts list-contactable-inboxes - List inboxes a contact can be contacted through [intent=direct_read availability=implemented operation=chatwoot.contacts_list_contactable_inboxes]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - contacts list-conversations - List conversations for a contact [intent=direct_read availability=implemented operation=chatwoot.contacts_list_conversations]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - contacts list-labels - List labels attached to a contact [intent=direct_read availability=implemented operation=chatwoot.contacts_list_labels]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - contacts merge - Merge Contacts [intent=reverse_etl availability=implemented write=merge_contacts]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: merge contacts via the Chatwoot Application API; flags: --base-contact-id (required) (integer): ID of the contact that will remain after the merge and receive all data: maps_to=record.base_contact_id, --mergee-contact-id (required) (integer): ID of the contact that will be merged into the base contact and deleted: maps_to=record.mergee_contact_id
  - contacts search - Search contacts by name/email/phone [intent=direct_read availability=implemented operation=chatwoot.contacts_search]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - contacts update - Update an existing contact [intent=reverse_etl availability=implemented write=update_contact]; approval: reverse ETL writes require plan, preview, approval, execute; risk: updates an existing Chatwoot contact's profile fields; low risk, no customer notification; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --name (string): Field `name`.: maps_to=record.name, --email (string): Field `email`.: maps_to=record.email, --phone-number (string): Field `phone_number`.: maps_to=record.phone_number, --identifier (string): Field `identifier`.: maps_to=record.identifier, --blocked (boolean): Field `blocked`.: maps_to=record.blocked, --avatar-url (string): Field `avatar_url`.: maps_to=record.avatar_url
  - contacts update-labels - Add Labels [intent=reverse_etl availability=implemented write=update_contact_labels]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add labels via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --labels (required) (string_array): Array of labels (comma-separated strings): maps_to=record.labels
- Conversations
  - conversations assign - Assign Conversation [intent=reverse_etl availability=implemented write=assign_conversation]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: assign conversation via the Chatwoot Application API; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --assignee-id (number): Id of the assignee user: maps_to=record.assignee_id, --team-id (number): Id of the team. If the assignee_id is present, this param would be ignored: maps_to=record.team_id
  - conversations create - Create a new conversation [intent=reverse_etl availability=implemented write=create_conversation]; approval: reverse ETL writes require plan, preview, approval, execute; risk: creates a new conversation in the target inbox; customer-visible once the initial message is delivered through a live channel; flags: --source-id (required) (string): Field `source_id`.: maps_to=record.source_id, --inbox-id (integer): Field `inbox_id`.: maps_to=record.inbox_id, --contact-id (integer): Field `contact_id`.: maps_to=record.contact_id, --status (string): Field `status`.: maps_to=record.status, --assignee-id (integer): Field `assignee_id`.: maps_to=record.assignee_id, --team-id (integer): Field `team_id`.: maps_to=record.team_id
  - conversations destroy-custom-attributes - Destroy Custom Attributes [intent=reverse_etl availability=implemented write=destroy_conversation_custom_attributes]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: destroy custom attributes via the Chatwoot Application API; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --custom-attributes (required) (string_array): The list of custom attribute keys to remove from the conversation: maps_to=record.custom_attributes
  - conversations filter - Advanced filtered conversation search [intent=direct_read availability=implemented operation=chatwoot.conversations_filter]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - conversations get - Retrieve a single conversation [intent=direct_read availability=implemented operation=chatwoot.conversations_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - conversations list - List conversations as ETL records. [intent=etl availability=implemented stream=conversations]
  - conversations list-labels - List labels attached to a conversation [intent=direct_read availability=implemented operation=chatwoot.conversations_list_labels]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - conversations list-reporting-events - List reporting events for a conversation [intent=direct_read availability=implemented operation=chatwoot.conversations_list_reporting_events]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - conversations meta - Retrieve conversation count summary (mine/unassigned/all) [intent=direct_read availability=implemented operation=chatwoot.conversations_meta]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - conversations toggle-priority - Toggle Priority [intent=reverse_etl availability=implemented write=toggle_conversation_priority]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: toggle priority via the Chatwoot Application API; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --priority (required) (string): The priority of the conversation: maps_to=record.priority
  - conversations toggle-status - Toggle a conversation's status [intent=reverse_etl availability=implemented write=toggle_conversation_status]; approval: reverse ETL writes require plan, preview, approval, execute; risk: changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and reporting metrics; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --status (required) (string): Field `status`.: maps_to=record.status, --snoozed-until (integer): Field `snoozed_until`.: maps_to=record.snoozed_until
  - conversations toggle-typing-status - Toggle Typing Status [intent=reverse_etl availability=implemented write=toggle_conversation_typing_status]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: toggle typing status via the Chatwoot Application API; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --typing-status (required) (string): Typing status to set.: maps_to=record.typing_status, --is-private (boolean): Whether the typing event is for private notes.: maps_to=record.is_private
  - conversations update - Update Conversation [intent=reverse_etl availability=implemented write=update_conversation]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update conversation via the Chatwoot Application API; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --priority (string): The priority of the conversation: maps_to=record.priority, --sla-policy-id (number): The ID of the SLA policy (Available only in Enterprise edition): maps_to=record.sla_policy_id
  - conversations update-custom-attributes - Update Custom Attributes [intent=reverse_etl availability=partial write=update_conversation_custom_attributes]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update custom attributes via the Chatwoot Application API; notes: partial: `custom_attributes` is a required, arbitrary-keyed JSON object (the API's own record_schema has no fixed property set); flat CLI flags cannot express it. Reverse ETL plan JSON (--config/plan-file record override) is the supported path.; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --merge (boolean): When true, only the keys sent in `custom_attributes` are updated and the remaining keys are kept intact. When false (default), the whole custom attributes hash is replaced. Use the destroy_custom_attributes endpoint to remove keys.: maps_to=record.merge
  - conversations update-labels - Add Labels [intent=reverse_etl availability=implemented write=update_conversation_labels]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add labels via the Chatwoot Application API; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --labels (required) (string_array): Array of labels (comma-separated strings): maps_to=record.labels
- Custom Attribute Definitions
  - custom-attributes create - Add a new custom attribute [intent=reverse_etl availability=implemented write=create_custom_attribute_definition]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new custom attribute via the Chatwoot Application API; flags: --attribute-display-name (string): Attribute display name: maps_to=record.attribute_display_name, --attribute-display-type (integer): Attribute display type (text- 0, number- 1, currency- 2, percent- 3, link- 4, date- 5, list- 6, checkbox- 7): maps_to=record.attribute_display_type, --attribute-description (string): Attribute description: maps_to=record.attribute_description, --attribute-key (string): Attribute unique key value: maps_to=record.attribute_key, --attribute-values (string_array): Attribute values: maps_to=record.attribute_values, --attribute-model (integer): Attribute type(conversation_attribute- 0, contact_attribute- 1): maps_to=record.attribute_model, --regex-pattern (string): Regex pattern (Only applicable for type- text). The regex pattern is used to validate the attribute value(s).: maps_to=record.regex_pattern, --regex-cue (string): Regex cue message (Only applicable for type- text). The cue message is shown when the regex pattern is not matched.: maps_to=record.regex_cue
  - custom-attributes delete - Remove a custom attribute from account [intent=reverse_etl availability=implemented write=delete_custom_attribute_definition]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: remove a custom attribute from account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id
  - custom-attributes get - Retrieve a single custom attribute definition [intent=direct_read availability=implemented operation=chatwoot.custom_attributes_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - custom-attributes list - List custom attribute definitions [intent=direct_read availability=implemented operation=chatwoot.custom_attributes_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - custom-attributes update - Update custom attribute in Account [intent=reverse_etl availability=implemented write=update_custom_attribute_definition]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update custom attribute in account via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --attribute-display-name (string): Attribute display name: maps_to=record.attribute_display_name, --attribute-display-type (integer): Attribute display type (text- 0, number- 1, currency- 2, percent- 3, link- 4, date- 5, list- 6, checkbox- 7): maps_to=record.attribute_display_type, --attribute-description (string): Attribute description: maps_to=record.attribute_description, --attribute-key (string): Attribute unique key value: maps_to=record.attribute_key, --attribute-values (string_array): Attribute values: maps_to=record.attribute_values, --attribute-model (integer): Attribute type(conversation_attribute- 0, contact_attribute- 1): maps_to=record.attribute_model, --regex-pattern (string): Regex pattern (Only applicable for type- text). The regex pattern is used to validate the attribute value(s).: maps_to=record.regex_pattern, --regex-cue (string): Regex cue message (Only applicable for type- text). The cue message is shown when the regex pattern is not matched.: maps_to=record.regex_cue
- Custom Filters
  - custom-filters create - Create a custom filter [intent=reverse_etl availability=implemented write=create_custom_filter]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: create a custom filter via the Chatwoot Application API; flags: --name (string): The name of the custom filter: maps_to=record.name, --type (string): The description about the custom filter: maps_to=record.type
  - custom-filters delete - Delete a custom filter [intent=reverse_etl availability=implemented write=delete_custom_filter]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete a custom filter via the Chatwoot Application API; flags: --custom-filter-id (required) (integer): Path parameter `custom_filter_id`.: maps_to=record.custom_filter_id
  - custom-filters get - Retrieve a single custom filter [intent=direct_read availability=implemented operation=chatwoot.custom_filters_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - custom-filters list - List saved custom filters [intent=direct_read availability=implemented operation=chatwoot.custom_filters_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - custom-filters update - Update a custom filter [intent=reverse_etl availability=implemented write=update_custom_filter]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update a custom filter via the Chatwoot Application API; flags: --custom-filter-id (required) (integer): Path parameter `custom_filter_id`.: maps_to=record.custom_filter_id, --name (string): The name of the custom filter: maps_to=record.name, --type (string): The description about the custom filter: maps_to=record.type
- Inbox Members
  - inbox-members add - Add a New Agent [intent=reverse_etl availability=implemented write=add_inbox_members]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new agent via the Chatwoot Application API; flags: --inbox-id (required) (integer): The ID of the inbox: maps_to=record.inbox_id, --user-ids (required) (string_array): IDs of users to be added to the inbox: maps_to=record.user_ids
  - inbox-members list - List agents who are members of an inbox [intent=direct_read availability=implemented operation=chatwoot.inbox_members_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - inbox-members remove - Remove an Agent from Inbox [intent=reverse_etl availability=implemented write=remove_inbox_members]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: remove an agent from inbox via the Chatwoot Application API; flags: --inbox-id (required) (string): The ID of the inbox: maps_to=record.inbox_id, --user-ids (required) (string_array): IDs of users to be deleted from the inbox: maps_to=record.user_ids
  - inbox-members set - Update Agents in Inbox [intent=reverse_etl availability=implemented write=set_inbox_members]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update agents in inbox via the Chatwoot Application API; flags: --inbox-id (required) (string): The ID of the inbox: maps_to=record.inbox_id, --user-ids (required) (string_array): IDs of users to be added to the inbox: maps_to=record.user_ids
- Inboxes
  - inboxes get - Retrieve a single inbox [intent=direct_read availability=implemented operation=chatwoot.inboxes_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - inboxes get-agent-bot - Retrieve the agent bot attached to an inbox [intent=direct_read availability=implemented operation=chatwoot.inboxes_get_agent_bot]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - inboxes list - List inboxes as ETL records. [intent=etl availability=implemented stream=inboxes]
  - inboxes list-message-templates - List WhatsApp message templates for an inbox [intent=direct_read availability=implemented operation=chatwoot.inboxes_list_message_templates]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - inboxes set-agent-bot - Add or remove agent bot [intent=reverse_etl availability=implemented write=set_inbox_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add or remove agent bot via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --agent-bot (required) (number): Agent bot ID: maps_to=record.agent_bot
- Integrations
  - integrations create-hook - Create an integration hook [intent=reverse_etl availability=implemented write=create_integration_hook]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: create an integration hook via the Chatwoot Application API; flags: --app-id (integer): The ID of app for which integration hook is being created: maps_to=record.app_id, --inbox-id (integer): The inbox ID, if the hook is an inbox hook: maps_to=record.inbox_id, --status (integer): The status of the integration (0 for inactive, 1 for active): maps_to=record.status
  - integrations delete-hook - Delete an Integration Hook [intent=reverse_etl availability=implemented write=delete_integration_hook]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete an integration hook via the Chatwoot Application API; flags: --hook-id (required) (integer): Path parameter `hook_id`.: maps_to=record.hook_id
  - integrations list-apps - List installed integrations [intent=direct_read availability=implemented operation=chatwoot.integrations_list_apps]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - integrations update-hook - Update an Integration Hook [intent=reverse_etl availability=implemented write=update_integration_hook]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update an integration hook via the Chatwoot Application API; flags: --hook-id (required) (integer): Path parameter `hook_id`.: maps_to=record.hook_id, --status (integer): The status of the integration (0 for inactive, 1 for active): maps_to=record.status
- Labels
  - labels create - Create a new label [intent=reverse_etl availability=implemented write=create_label]; approval: reverse ETL writes require plan, preview, approval, execute; risk: creates a new account-wide label; low risk, visible to all agents in the sidebar when show_on_sidebar is true; flags: --title (required) (string): Field `title`.: maps_to=record.title, --description (string): Field `description`.: maps_to=record.description, --color (string): Field `color`.: maps_to=record.color, --show-on-sidebar (boolean): Field `show_on_sidebar`.: maps_to=record.show_on_sidebar
  - labels delete - Delete a label [intent=reverse_etl availability=implemented write=delete_label]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete a label via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id
  - labels get - Retrieve a single label [intent=direct_read availability=implemented operation=chatwoot.labels_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - labels list - List labels as ETL records. [intent=etl availability=implemented stream=labels]
  - labels update - Update a label [intent=reverse_etl availability=implemented write=update_label]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update a label via the Chatwoot Application API; flags: --id (required) (integer): Path parameter `id`.: maps_to=record.id, --title (string): The label title: maps_to=record.title, --description (string): A short description for the label: maps_to=record.description, --color (string): Hex color code for the label: maps_to=record.color, --show-on-sidebar (boolean): Whether the label should appear in the sidebar: maps_to=record.show_on_sidebar
- Messages
  - messages delete - Delete a message [intent=reverse_etl availability=implemented write=delete_message]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete a message via the Chatwoot Application API; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --message-id (required) (integer): Path parameter `message_id`.: maps_to=record.message_id
  - messages list - List conversation-scoped messages as ETL records, fanned out across every conversation. [intent=etl availability=implemented stream=messages]
  - messages send - Send a message in a conversation [intent=reverse_etl availability=implemented write=send_message]; approval: reverse ETL writes require plan, preview, approval, execute; risk: sends a message into a conversation; customer-visible unless private is true and may notify the contact through the inbox channel; flags: --conversation-id (required) (integer): Path parameter `conversation_id`.: maps_to=record.conversation_id, --content (required) (string): Field `content`.: maps_to=record.content, --message-type (string): Field `message_type`.: maps_to=record.message_type, --private (boolean): Field `private`.: maps_to=record.private, --content-type (string): Field `content_type`.: maps_to=record.content_type
- Help Center Portals
  - portals create - Add a new portal [intent=reverse_etl availability=implemented write=create_portal]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new portal via the Chatwoot Application API; flags: --color (string): Header color for help-center in hex format: maps_to=record.color, --custom-domain (string): Custom domain to display help center.: maps_to=record.custom_domain, --header-text (string): Help center header: maps_to=record.header_text, --homepage-link (string): link to main dashboard: maps_to=record.homepage_link, --name (string): Name for the portal: maps_to=record.name, --page-title (string): Page title for the portal: maps_to=record.page_title, --slug (string): Slug for the portal to display in link: maps_to=record.slug, --archived (boolean): Status to check if portal is live: maps_to=record.archived
  - portals create-article - Add a new article [intent=reverse_etl availability=implemented write=create_portal_article]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new article via the Chatwoot Application API; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --title (string): The title of the article: maps_to=record.title, --slug (string): The slug of the article: maps_to=record.slug, --position (integer): article position in category: maps_to=record.position, --content (string): The text content.: maps_to=record.content, --description (string): The description of the article: maps_to=record.description, --category-id (integer): The category id of the article: maps_to=record.category_id, --author-id (integer): The author agent id of the article: maps_to=record.author_id, --associated-article-id (integer): To associate similar articles to each other, e.g to provide the link for the reference.: maps_to=record.associated_article_id, --status (integer): The status of the article. 0 for draft, 1 for published, 2 for archived: maps_to=record.status, --locale (string): The locale of the article: maps_to=record.locale
  - portals create-category - Add a new category [intent=reverse_etl availability=implemented write=create_portal_category]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new category via the Chatwoot Application API; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --name (string): The name of the category: maps_to=record.name, --description (string): A description for the category: maps_to=record.description, --position (integer): Category position in the portal list to sort: maps_to=record.position, --slug (string): The category slug used in the URL: maps_to=record.slug, --locale (string): The locale of the category: maps_to=record.locale, --icon (string): The icon of the category as a string (emoji): maps_to=record.icon, --parent-category-id (integer): To define parent category, e.g product documentation has multiple level features in sales category or in engineering category.: maps_to=record.parent_category_id, --associated-category-id (integer): To associate similar categories to each other, e.g same category of product documentation in different languages: maps_to=record.associated_category_id
  - portals list - List help-center portals [intent=direct_read availability=implemented operation=chatwoot.portals_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - portals update - Update a portal [intent=reverse_etl availability=implemented write=update_portal]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update a portal via the Chatwoot Application API; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --color (string): Header color for help-center in hex format: maps_to=record.color, --custom-domain (string): Custom domain to display help center.: maps_to=record.custom_domain, --header-text (string): Help center header: maps_to=record.header_text, --homepage-link (string): link to main dashboard: maps_to=record.homepage_link, --name (string): Name for the portal: maps_to=record.name, --page-title (string): Page title for the portal: maps_to=record.page_title, --slug (string): Slug for the portal to display in link: maps_to=record.slug, --archived (boolean): Status to check if portal is live: maps_to=record.archived
- Reporting Events
  - reporting-events list - List account-wide reporting events [intent=direct_read availability=implemented operation=chatwoot.reporting_events_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
- Teams
  - teams add-members - Add a New Agent [intent=reverse_etl availability=implemented write=add_team_members]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a new agent via the Chatwoot Application API; flags: --team-id (required) (integer): Path parameter `team_id`.: maps_to=record.team_id, --user-ids (required) (string_array): IDs of users to be added to the team: maps_to=record.user_ids
  - teams create - Create a team [intent=reverse_etl availability=implemented write=create_team]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: create a team via the Chatwoot Application API; flags: --name (string): The name of the team: maps_to=record.name, --description (string): The description of the team: maps_to=record.description, --allow-auto-assign (boolean): If this setting is turned on, the system would automatically assign the conversation to an agent in the team while assigning the conversation to a team: maps_to=record.allow_auto_assign
  - teams delete - Delete a team [intent=reverse_etl availability=implemented write=delete_team]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete a team via the Chatwoot Application API; flags: --team-id (required) (integer): Path parameter `team_id`.: maps_to=record.team_id
  - teams get - Retrieve a single team [intent=direct_read availability=implemented operation=chatwoot.teams_get]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - teams list - List teams as ETL records. [intent=etl availability=implemented stream=teams]
  - teams list-members - List a team's member agents [intent=direct_read availability=implemented operation=chatwoot.teams_list_members]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - teams remove-members - Remove an Agent from Team [intent=reverse_etl availability=implemented write=remove_team_members]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: remove an agent from team via the Chatwoot Application API; flags: --team-id (required) (integer): Path parameter `team_id`.: maps_to=record.team_id, --user-ids (required) (string_array): IDs of users to be deleted from the team: maps_to=record.user_ids
  - teams set-members - Update Agents in Team [intent=reverse_etl availability=implemented write=set_team_members]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update agents in team via the Chatwoot Application API; flags: --team-id (required) (integer): Path parameter `team_id`.: maps_to=record.team_id, --user-ids (required) (string_array): IDs of users to be added to the team: maps_to=record.user_ids
  - teams update - Update a team [intent=reverse_etl availability=implemented write=update_team]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update a team via the Chatwoot Application API; flags: --team-id (required) (integer): Path parameter `team_id`.: maps_to=record.team_id, --name (string): The name of the team: maps_to=record.name, --description (string): The description of the team: maps_to=record.description, --allow-auto-assign (boolean): If this setting is turned on, the system would automatically assign the conversation to an agent in the team while assigning the conversation to a team: maps_to=record.allow_auto_assign
- Webhooks
  - webhooks create - Add a webhook [intent=reverse_etl availability=implemented write=create_webhook]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: add a webhook via the Chatwoot Application API; flags: --url (string): The url where the events should be sent: maps_to=record.url, --name (string): The name of the webhook: maps_to=record.name, --subscriptions (string_array): The events you want to subscribe to.: maps_to=record.subscriptions
  - webhooks delete - Delete a webhook [intent=reverse_etl availability=implemented write=delete_webhook]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete a webhook via the Chatwoot Application API; flags: --webhook-id (required) (integer): Path parameter `webhook_id`.: maps_to=record.webhook_id
  - webhooks list - List webhook subscriptions [intent=direct_read availability=implemented operation=chatwoot.webhooks_list]; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - webhooks update - Update a webhook object [intent=reverse_etl availability=implemented write=update_webhook]; approval: reverse ETL writes require plan, preview, approval, execute; risk: medium: update a webhook object via the Chatwoot Application API; flags: --webhook-id (required) (integer): Path parameter `webhook_id`.: maps_to=record.webhook_id, --url (string): The url where the events should be sent: maps_to=record.url, --name (string): The name of the webhook: maps_to=record.name, --subscriptions (string_array): The events you want to subscribe to.: maps_to=record.subscriptions
- Help topics:
  - auth - Add a Chatwoot API access token with `pm credentials add chatwoot`.

## Commands

### Inspect as a manual

```bash
pm connectors inspect chatwoot
```

### Inspect as structured JSON

```bash
pm connectors inspect chatwoot --json
```

## Agent Rules

- Run pm connectors inspect chatwoot before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
