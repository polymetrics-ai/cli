# pm connectors inspect chatwoot

```text
NAME
  pm connectors inspect chatwoot - Chatwoot connector manual

SYNOPSIS
  pm connectors inspect chatwoot
  pm connectors inspect chatwoot --json
  pm credentials add <name> --connector chatwoot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Chatwoot support conversations, contacts, inboxes, agents, teams, labels, messages, reports, public inbox, and platform metadata; writes typed approved Chatwoot mutations through reverse ETL approval gates; and tracks duplicate/disallowed official Swagger rows as blocked-by-default metadata.

ICON
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
  account_id
  base_url
  start_date
  api_access_token (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: updated_at
    fields: account_id(), additional_attributes(), agent_last_seen_at(), assignee_last_seen_at(), can_reply(), contact_last_seen_at(), created_at(), custom_attributes(), first_reply_created_at(), id(), inbox_id(), labels(), last_activity_at(), muted(), priority(), sla_policy_id(), snoozed_until(), status(), timestamp(), unread_count(), updated_at(), uuid(), waiting_since()
  contacts:
    primary key: id
    cursor: last_activity_at
    fields: additional_attributes(), availability_status(), blocked(), contact_inboxes(), created_at(), custom_attributes(), email(), id(), identifier(), last_activity_at(), name(), phone_number(), thumbnail()
  inboxes:
    primary key: id
    fields: allow_messages_after_resolved(), avatar_url(), business_name(), callback_webhook_url(), channel_id(), channel_type(), csat_survey_enabled(), enable_auto_assignment(), enable_email_collect(), greeting_enabled(), greeting_message(), id(), lock_to_single_conversation(), medium(), name(), out_of_office_message(), phone_number(), provider(), timezone(), website_token(), website_url(), welcome_tagline(), welcome_title(), widget_color(), working_hours_enabled()
  agents:
    primary key: id
    fields: account_id(), auto_offline(), availability_status(), available_name(), confirmed(), custom_role_id(), email(), id(), name(), role(), thumbnail()
  teams:
    primary key: id
    fields: account_id(), allow_auto_assign(), description(), id(), is_member(), name()
  labels:
    primary key: id
    fields: color(), description(), id(), show_on_sidebar(), title()
  messages:
    primary key: id
    cursor: id
    fields: account_id(), attachment(), content(), content_attributes(), content_type(), conversation_id(), created_at(), id(), inbox_id(), message_type(), private(), sender(), sender_id(), sender_type(), source_id(), status(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_contact:
    endpoint: POST /contacts
    risk: creates a new Chatwoot contact record; low risk, no customer notification
  update_contact:
    endpoint: PUT /contacts/{{ record.id }}
    required fields: id
    risk: updates an existing Chatwoot contact's profile fields; low risk, no customer notification
  create_conversation:
    endpoint: POST /conversations
    risk: creates a new conversation in the target inbox; customer-visible once the initial message is delivered through a live channel
  send_message:
    endpoint: POST /conversations/{{ record.conversation_id }}/messages
    required fields: conversation_id
    risk: sends a message into a conversation; customer-visible unless private is true and may notify the contact through the inbox channel
  toggle_conversation_status:
    endpoint: POST /conversations/{{ record.conversation_id }}/toggle_status
    required fields: conversation_id
    risk: changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and reporting metrics
  create_label:
    endpoint: POST /labels
    risk: creates a new account-wide label; low risk, visible to all agents in the sidebar when show_on_sidebar is true
  update_account:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}
    optional fields: name, locale, domain, support_email, auto_resolve_after, auto_resolve_message, auto_resolve_ignore_waiting, industry, company_size, timezone
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  merge_contact:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/actions/contact_merge
    optional fields: base_contact_id, mergee_contact_id
    risk: critical: destructive_action Chatwoot operation; reverse ETL approval required
  create_agent_bot:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/agent_bots
    optional fields: name, description, outgoing_url, avatar, avatar_url, bot_type, bot_config
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_agent_bot:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/agent_bots/{{ record.id }}
    required fields: id
    optional fields: name, description, outgoing_url, avatar, avatar_url, bot_type, bot_config
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_agent_bot:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/agent_bots/{{ record.id }}
    required fields: id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_agent:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/agents
    optional fields: name, email, role, availability, auto_offline
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_agent:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/agents/{{ record.id }}
    required fields: id
    optional fields: role, availability, auto_offline
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_agent:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/agents/{{ record.id }}
    required fields: id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_automation_rule:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/automation_rules
    optional fields: name, description, event_name, active, actions, conditions
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_automation_rule:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/automation_rules/{{ record.id }}
    required fields: id
    optional fields: name, description, event_name, active, actions, conditions
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_automation_rule:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/automation_rules/{{ record.id }}
    required fields: id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_canned_response:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/canned_responses
    optional fields: content, short_code
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  update_canned_response:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/canned_responses/{{ record.id }}
    required fields: id
    optional fields: content, short_code
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_canned_response:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/canned_responses/{{ record.id }}
    required fields: id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  filter_contacts:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/contacts/filter
    optional fields: payload
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_contact:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/contacts/{{ record.id }}
    required fields: id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_contact_inbox:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/contacts/{{ record.id }}/contact_inboxes
    required fields: id
    optional fields: inbox_id, source_id
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  set_contact_labels:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/contacts/{{ record.id }}/labels
    required fields: id
    optional fields: labels
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  filter_conversations:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/conversations/filter
    optional fields: payload
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  update_conversation:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/conversations/{{ record.conversation_id }}
    required fields: conversation_id
    optional fields: priority, sla_policy_id
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  assign_conversation:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/conversations/{{ record.conversation_id }}/assignments
    required fields: conversation_id
    optional fields: assignee_id, team_id
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  update_conversation_custom_attributes:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/conversations/{{ record.conversation_id }}/custom_attributes
    required fields: conversation_id
    optional fields: custom_attributes
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  set_conversation_labels:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/conversations/{{ record.conversation_id }}/labels
    required fields: conversation_id
    optional fields: labels
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_conversation_message:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/conversations/{{ record.conversation_id }}/messages/{{ record.message_id }}
    required fields: conversation_id, message_id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  toggle_conversation_priority:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/conversations/{{ record.conversation_id }}/toggle_priority
    required fields: conversation_id
    optional fields: priority
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  create_custom_attribute_definition:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/custom_attribute_definitions
    optional fields: attribute_display_name, attribute_display_type, attribute_description, attribute_key, attribute_values, attribute_model, regex_pattern, regex_cue
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_custom_attribute_definition:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/custom_attribute_definitions/{{ record.id }}
    required fields: id
    optional fields: attribute_display_name, attribute_display_type, attribute_description, attribute_key, attribute_values, attribute_model, regex_pattern, regex_cue
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_custom_attribute_definition:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/custom_attribute_definitions/{{ record.id }}
    required fields: id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_custom_filter:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/custom_filters
    optional fields: name, type, query
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_custom_filter:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/custom_filters/{{ record.custom_filter_id }}
    required fields: custom_filter_id
    optional fields: name, type, query
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_custom_filter:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/custom_filters/{{ record.custom_filter_id }}
    required fields: custom_filter_id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_inbox_member:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/inbox_members
    optional fields: inbox_id, user_ids
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_inbox_member:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/inbox_members
    optional fields: inbox_id, user_ids
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_inbox_member:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/inbox_members
    optional fields: inbox_id, user_ids
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_inbox:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/inboxes
    optional fields: name, avatar, greeting_enabled, greeting_message, enable_email_collect, csat_survey_enabled, csat_config, enable_auto_assignment, working_hours_enabled, out_of_office_message, timezone, allow_messages_after_resolved, lock_to_single_conversation, portal_id, sender_name_type, business_name, channel
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_inbox:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/inboxes/{{ record.id }}
    required fields: id
    optional fields: name, avatar, greeting_enabled, greeting_message, enable_email_collect, csat_survey_enabled, csat_config, enable_auto_assignment, working_hours_enabled, out_of_office_message, timezone, allow_messages_after_resolved, lock_to_single_conversation, portal_id, sender_name_type, business_name, channel
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  set_inbox_agent_bot:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/inboxes/{{ record.id }}/set_agent_bot
    required fields: id
    optional fields: agent_bot
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  create_integration_hook:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/integrations/hooks
    optional fields: app_id, inbox_id, status, settings
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_integration_hook:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/integrations/hooks/{{ record.hook_id }}
    required fields: hook_id
    optional fields: status, settings
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_integration_hook:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/integrations/hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  update_label:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/labels/{{ record.id }}
    required fields: id
    optional fields: title, description, color, show_on_sidebar
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_label:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/labels/{{ record.id }}
    required fields: id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_portal:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/portals
    optional fields: color, custom_domain, header_text, homepage_link, name, page_title, slug, archived, config
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_portal:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/portals/{{ record.id }}
    required fields: id
    optional fields: color, custom_domain, header_text, homepage_link, name, page_title, slug, archived, config
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  create_portal_article:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/portals/{{ record.id }}/articles
    required fields: id
    optional fields: title, slug, position, content, description, category_id, author_id, associated_article_id, status, locale, meta
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  create_portal_category:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/portals/{{ record.id }}/categories
    required fields: id
    optional fields: name, description, position, slug, locale, icon, parent_category_id, associated_category_id
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  create_team:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/teams
    optional fields: name, description, allow_auto_assign
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_team:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/teams/{{ record.team_id }}
    required fields: team_id
    optional fields: name, description, allow_auto_assign
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_team:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/teams/{{ record.team_id }}
    required fields: team_id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_team_member:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/teams/{{ record.team_id }}/team_members
    required fields: team_id
    optional fields: user_ids
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_team_member:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/teams/{{ record.team_id }}/team_members
    required fields: team_id
    optional fields: user_ids
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_team_member:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/teams/{{ record.team_id }}/team_members
    required fields: team_id
    optional fields: user_ids
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  create_webhook:
    endpoint: POST /api/v1/accounts/{{ config.account_id }}/webhooks
    optional fields: url, name, subscriptions
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_webhook:
    endpoint: PATCH /api/v1/accounts/{{ config.account_id }}/webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    optional fields: url, name, subscriptions
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_webhook:
    endpoint: DELETE /api/v1/accounts/{{ config.account_id }}/webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: high: destructive_action Chatwoot operation; reverse ETL approval required
  update_profile:
    endpoint: PUT /api/v1/profile
    optional fields: profile
    risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  create_platform_account:
    endpoint: POST /platform/api/v1/accounts
    optional fields: name, locale, domain, support_email, status, limits, custom_attributes
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_platform_account:
    endpoint: PATCH /platform/api/v1/accounts/{{ record.account_id }}
    required fields: account_id
    optional fields: name, locale, domain, support_email, status, limits, custom_attributes
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_platform_account:
    endpoint: DELETE /platform/api/v1/accounts/{{ record.account_id }}
    required fields: account_id
    risk: critical: destructive_action Chatwoot operation; reverse ETL approval required
  create_platform_account_user:
    endpoint: POST /platform/api/v1/accounts/{{ record.account_id }}/account_users
    required fields: account_id
    optional fields: user_id, role
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_platform_account_user:
    endpoint: DELETE /platform/api/v1/accounts/{{ record.account_id }}/account_users
    required fields: account_id
    risk: critical: destructive_action Chatwoot operation; reverse ETL approval required
  create_platform_agent_bot:
    endpoint: POST /platform/api/v1/agent_bots
    optional fields: name, description, outgoing_url, account_id, avatar, avatar_url
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_platform_agent_bot:
    endpoint: PATCH /platform/api/v1/agent_bots/{{ record.id }}
    required fields: id
    optional fields: name, description, outgoing_url, account_id, avatar, avatar_url
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_platform_agent_bot:
    endpoint: DELETE /platform/api/v1/agent_bots/{{ record.id }}
    required fields: id
    risk: critical: destructive_action Chatwoot operation; reverse ETL approval required
  create_platform_user:
    endpoint: POST /platform/api/v1/users
    optional fields: name, display_name, email, password, custom_attributes
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  update_platform_user:
    endpoint: PATCH /platform/api/v1/users/{{ record.id }}
    required fields: id
    optional fields: name, display_name, email, password, custom_attributes
    risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required
  delete_platform_user:
    endpoint: DELETE /platform/api/v1/users/{{ record.id }}
    required fields: id
    risk: critical: destructive_action Chatwoot operation; reverse ETL approval required
  create_public_inbox_contact:
    endpoint: POST /public/api/v1/inboxes/{{ record.inbox_identifier }}/contacts
    required fields: inbox_identifier
    optional fields: identifier, identifier_hash, email, name, phone_number, avatar, custom_attributes
    risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  update_public_inbox_contact:
    endpoint: PATCH /public/api/v1/inboxes/{{ record.inbox_identifier }}/contacts/{{ record.contact_identifier }}
    required fields: inbox_identifier, contact_identifier
    optional fields: identifier, identifier_hash, email, name, phone_number, avatar, custom_attributes
    risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  create_public_inbox_contact_conversation:
    endpoint: POST /public/api/v1/inboxes/{{ record.inbox_identifier }}/contacts/{{ record.contact_identifier }}/conversations
    required fields: inbox_identifier, contact_identifier
    optional fields: custom_attributes
    risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  create_public_inbox_contact_conversation_message:
    endpoint: POST /public/api/v1/inboxes/{{ record.inbox_identifier }}/contacts/{{ record.contact_identifier }}/conversations/{{ record.conversation_id }}/messages
    required fields: inbox_identifier, contact_identifier, conversation_id
    optional fields: content, echo_id
    risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  update_public_inbox_contact_conversation_message:
    endpoint: PATCH /public/api/v1/inboxes/{{ record.inbox_identifier }}/contacts/{{ record.contact_identifier }}/conversations/{{ record.conversation_id }}/messages/{{ record.message_id }}
    required fields: inbox_identifier, contact_identifier, conversation_id, message_id
    optional fields: submitted_values
    risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
  toggle_public_conversation_status:
    endpoint: POST /public/api/v1/inboxes/{{ record.inbox_identifier }}/contacts/{{ record.contact_identifier }}/conversations/{{ record.conversation_id }}/toggle_status
    required fields: inbox_identifier, contact_identifier, conversation_id
    risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required

SECURITY
  read risk: external Chatwoot API reads of account-scoped support data, reports, public inbox resources, and platform metadata through bounded fixed direct-read commands
  write risk: external mutation of Chatwoot contacts, conversations, messages, labels, admin configuration, public inbox resources, and platform resources through typed reverse ETL actions; destructive actions require approval and typed confirmation metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Chatwoot support-desk data and approved account-scoped mutations.
  Usage: pm chatwoot <command> <subcommand> [flags]
  Source CLI: chatwoot-api (https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Chatwoot connector credential and account scope; maps_to=connection
  Support Desk Commands
    conversation list - List account conversations [intent=etl availability=implemented stream=conversations]; flags: --status
    conversation filter - Sensitive Chatwoot conversation filter [intent=reverse_etl availability=implemented write=filter_conversations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
    conversation update - Sensitive Chatwoot conversation update [intent=reverse_etl availability=implemented write=update_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --conversation-id, --priority
    conversation assign - Sensitive Chatwoot conversation assign [intent=reverse_etl availability=implemented write=assign_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --conversation-id
    conversation custom-attributes update - Sensitive Chatwoot conversation custom attributes update [intent=reverse_etl availability=implemented write=update_conversation_custom_attributes]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --conversation-id, --custom-attributes
    conversation labels update - Sensitive Chatwoot conversation labels update [intent=reverse_etl availability=implemented write=set_conversation_labels]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --conversation-id, --labels
    conversation message delete - Destructive Chatwoot conversation message delete [intent=reverse_etl availability=implemented write=delete_conversation_message]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --conversation-id, --message-id
    conversation priority update - Sensitive Chatwoot conversation priority update [intent=reverse_etl availability=implemented write=toggle_conversation_priority]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --conversation-id, --priority
    conversation view - View one conversation [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --conversation-id
    conversation create - Create a conversation [intent=reverse_etl availability=implemented write=create_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a new Chatwoot conversation in the configured account and may become customer-visible through the inbox channel.; flags: --source-id, --inbox-id, --contact-id, --status
    conversation toggle-status - Change a conversation status [intent=reverse_etl availability=implemented write=toggle_conversation_status]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Changes a conversation's workflow status and can affect routing, reporting, and agent queues.; flags: --conversation-id, --status, --snoozed-until
    conversation meta view - Read conversation meta view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --status, --q, --inbox-id, --team-id, --labels
    conversation label list - Read conversation label list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --conversation-id
    conversation reporting-event list - Read conversation reporting event list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --conversation-id
    contact merge - Destructive Chatwoot contact merge [intent=reverse_etl availability=implemented write=merge_contact]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: critical: destructive_action Chatwoot operation; reverse ETL approval required; flags: --base-contact-id, --mergee-contact-id
    contact filter - Sensitive Chatwoot contact filter [intent=reverse_etl availability=implemented write=filter_contacts]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required
    contact delete - Destructive Chatwoot contact delete [intent=reverse_etl availability=implemented write=delete_contact]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    contact inbox create - Sensitive Chatwoot contact inbox create [intent=reverse_etl availability=implemented write=create_contact_inbox]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --source-id, --inbox-id
    contact labels update - Sensitive Chatwoot contact labels update [intent=reverse_etl availability=implemented write=set_contact_labels]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --labels
    contact list - List contacts [intent=etl availability=implemented stream=contacts]
    contact view - View one contact [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    contact search - Search contacts [intent=direct_read availability=implemented]; notes: Bounded query direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --q
    contact create - Create a contact [intent=reverse_etl availability=implemented write=create_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Chatwoot contact record; visible to agents and usable in conversations.; flags: --inbox-id, --name, --email, --phone-number
    contact update - Update a contact [intent=reverse_etl availability=implemented write=update_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates an existing Chatwoot contact profile; visible to agents and downstream automations.; flags: --id, --name, --email, --phone-number
    contact contactable-inbox list - Read contact contactable inbox list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    contact conversation list - Read contact conversation list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    contact label list - Read contact label list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    message list - List messages for each conversation [intent=etl availability=implemented stream=messages]; notes: The messages stream fans out across conversations; single-conversation message direct reads are planned separately.
    message send - Send a conversation message [intent=reverse_etl availability=implemented write=send_message]; approval: reverse ETL writes require plan, preview, approval, execute. Attachment/multipart variants remain blocked until the binary policy slice.; risk: Sends a message into a Chatwoot conversation; customer-visible unless private is true and may trigger notifications.; flags: --conversation-id, --content, --message-type, --private
  Workspace Metadata Commands
    inbox create - Admin Chatwoot inbox create [intent=reverse_etl availability=implemented write=create_inbox]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --avatar, --greeting-enabled, --greeting-message, --enable-email-collect, --csat-survey-enabled, --enable-auto-assignment, --working-hours-enabled, --out-of-office-message, --timezone, --allow-messages-after-resolved, --lock-to-single-conversation, --portal-id, --sender-name-type, --business-name, --channel
    inbox update - Admin Chatwoot inbox update [intent=reverse_etl availability=implemented write=update_inbox]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --name, --avatar, --greeting-enabled, --greeting-message, --enable-email-collect, --csat-survey-enabled, --enable-auto-assignment, --working-hours-enabled, --out-of-office-message, --timezone, --allow-messages-after-resolved, --lock-to-single-conversation, --portal-id, --sender-name-type, --business-name, --channel
    inbox agent-bot update - Admin Chatwoot inbox agent bot update [intent=reverse_etl availability=implemented write=set_inbox_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --agent-bot
    inbox list - List inboxes [intent=etl availability=implemented stream=inboxes]
    inbox view - Read inbox view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    inbox agent-bot list - Read inbox agent bot list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    agent create - Admin Chatwoot agent create [intent=reverse_etl availability=implemented write=create_agent]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --email, --role, --availability, --auto-offline
    agent update - Admin Chatwoot agent update [intent=reverse_etl availability=implemented write=update_agent]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --role, --availability, --auto-offline
    agent delete - Destructive Chatwoot agent delete [intent=reverse_etl availability=implemented write=delete_agent]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    agent list - List account agents [intent=etl availability=implemented stream=agents]
    team create - Admin Chatwoot team create [intent=reverse_etl availability=implemented write=create_team]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --description, --allow-auto-assign
    team update - Admin Chatwoot team update [intent=reverse_etl availability=implemented write=update_team]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --team-id, --name, --description, --allow-auto-assign
    team delete - Destructive Chatwoot team delete [intent=reverse_etl availability=implemented write=delete_team]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --team-id
    team member create - Admin Chatwoot team member create [intent=reverse_etl availability=implemented write=create_team_member]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --team-id, --user-ids
    team member update - Admin Chatwoot team member update [intent=reverse_etl availability=implemented write=update_team_member]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --team-id, --user-ids
    team member delete - Destructive Chatwoot team member delete [intent=reverse_etl availability=implemented write=delete_team_member]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --team-id, --user-ids
    team list - List teams [intent=etl availability=implemented stream=teams]
    team view - Read team view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --team-id
    team team-member list - Read team team member list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --team-id
    label update - Sensitive Chatwoot label update [intent=reverse_etl availability=implemented write=update_label]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --title, --description, --color, --show-on-sidebar
    label delete - Destructive Chatwoot label delete [intent=reverse_etl availability=implemented write=delete_label]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    label list - List labels [intent=etl availability=implemented stream=labels]
    label create - Create a label [intent=reverse_etl availability=implemented write=create_label]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates an account-wide label visible to Chatwoot agents.; flags: --title, --description, --color, --show-on-sidebar
    label view - Read label view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
  Admin And Configuration Commands
    account update - Admin Chatwoot account update [intent=reverse_etl availability=implemented write=update_account]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --locale, --domain, --support-email, --auto-resolve-after, --auto-resolve-message, --auto-resolve-ignore-waiting, --industry, --company-size, --timezone
    account view - Read account view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    webhook create - Admin Chatwoot webhook create [intent=reverse_etl availability=implemented write=create_webhook]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --url, --name, --subscriptions
    webhook update - Admin Chatwoot webhook update [intent=reverse_etl availability=implemented write=update_webhook]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --webhook-id, --url, --name, --subscriptions
    webhook delete - Destructive Chatwoot webhook delete [intent=reverse_etl availability=implemented write=delete_webhook]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --webhook-id
    webhook list - Read webhook list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    portal create - Admin Chatwoot portal create [intent=reverse_etl availability=implemented write=create_portal]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --color, --custom-domain, --header-text, --homepage-link, --name, --page-title, --slug, --archived
    portal update - Admin Chatwoot portal update [intent=reverse_etl availability=implemented write=update_portal]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --color, --custom-domain, --header-text, --homepage-link, --name, --page-title, --slug, --archived
    portal article create - Admin Chatwoot portal article create [intent=reverse_etl availability=implemented write=create_portal_article]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --title, --slug, --position, --content, --description, --category-id, --author-id, --associated-article-id, --status, --locale
    portal category create - Admin Chatwoot portal category create [intent=reverse_etl availability=implemented write=create_portal_category]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --name, --description, --position, --slug, --locale, --icon, --parent-category-id, --associated-category-id
    portal list - Read portal list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    platform account create - Admin Chatwoot platform account create [intent=reverse_etl availability=implemented write=create_platform_account]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --locale, --domain, --support-email, --status
    platform account update - Admin Chatwoot platform account update [intent=reverse_etl availability=implemented write=update_platform_account]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --account-id, --name, --locale, --domain, --support-email, --status
    platform account delete - Destructive Chatwoot platform account delete [intent=reverse_etl availability=implemented write=delete_platform_account]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: critical: destructive_action Chatwoot operation; reverse ETL approval required; flags: --account-id
    platform account-user create - Admin Chatwoot platform account user create [intent=reverse_etl availability=implemented write=create_platform_account_user]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --account-id, --user-id, --role
    platform account-user delete - Destructive Chatwoot platform account user delete [intent=reverse_etl availability=implemented write=delete_platform_account_user]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: critical: destructive_action Chatwoot operation; reverse ETL approval required; flags: --account-id
    platform agent-bot create - Admin Chatwoot platform agent bot create [intent=reverse_etl availability=implemented write=create_platform_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --description, --outgoing-url, --account-id, --avatar, --avatar-url
    platform agent-bot update - Admin Chatwoot platform agent bot update [intent=reverse_etl availability=implemented write=update_platform_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --name, --description, --outgoing-url, --account-id, --avatar, --avatar-url
    platform agent-bot delete - Destructive Chatwoot platform agent bot delete [intent=reverse_etl availability=implemented write=delete_platform_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: critical: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    platform user create - Admin Chatwoot platform user create [intent=reverse_etl availability=implemented write=create_platform_user]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --display-name, --email
    platform user update - Admin Chatwoot platform user update [intent=reverse_etl availability=implemented write=update_platform_user]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --name, --display-name, --email
    platform user delete - Destructive Chatwoot platform user delete [intent=reverse_etl availability=implemented write=delete_platform_user]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: critical: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    platform account view - Read platform account view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --account-id
    platform account-user list - Read platform account user list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --account-id
    platform agent-bot list - Read platform agent bot list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    platform agent-bot view - Read platform agent bot view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    platform user view - Read platform user view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    platform user login view - Read platform user login view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
  Reporting And Audit Commands
    report account - Read report account [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report conversation list - Read report conversation list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report conversation list-alt - Read report conversation list alt [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report first-response-time-distribution list - Read report first response time distribution list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report inbox-label-matrix list - Read report inbox label matrix list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report outgoing-messages-count list - Read report outgoing messages count list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --group-by
    report summary view - Read report summary view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    audit-log list - Read audit log list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --page
  Other Commands
    agent-bot create - Admin Chatwoot agent bot create [intent=reverse_etl availability=implemented write=create_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --description, --outgoing-url, --avatar, --avatar-url, --bot-type
    agent-bot update - Admin Chatwoot agent bot update [intent=reverse_etl availability=implemented write=update_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --name, --description, --outgoing-url, --avatar, --avatar-url, --bot-type
    agent-bot delete - Destructive Chatwoot agent bot delete [intent=reverse_etl availability=implemented write=delete_agent_bot]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    automation-rule create - Admin Chatwoot automation rule create [intent=reverse_etl availability=implemented write=create_automation_rule]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --description, --event-name, --active
    automation-rule update - Admin Chatwoot automation rule update [intent=reverse_etl availability=implemented write=update_automation_rule]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --name, --description, --event-name, --active
    automation-rule delete - Destructive Chatwoot automation rule delete [intent=reverse_etl availability=implemented write=delete_automation_rule]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    canned-response create - Sensitive Chatwoot canned response create [intent=reverse_etl availability=implemented write=create_canned_response]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --content, --short-code
    canned-response update - Sensitive Chatwoot canned response update [intent=reverse_etl availability=implemented write=update_canned_response]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --content, --short-code
    canned-response delete - Destructive Chatwoot canned response delete [intent=reverse_etl availability=implemented write=delete_canned_response]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    custom-attribute-definition create - Admin Chatwoot custom attribute definition create [intent=reverse_etl availability=implemented write=create_custom_attribute_definition]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --attribute-display-name, --attribute-display-type, --attribute-description, --attribute-values, --attribute-model, --regex-pattern, --regex-cue
    custom-attribute-definition update - Admin Chatwoot custom attribute definition update [intent=reverse_etl availability=implemented write=update_custom_attribute_definition]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --id, --attribute-display-name, --attribute-display-type, --attribute-description, --attribute-values, --attribute-model, --regex-pattern, --regex-cue
    custom-attribute-definition delete - Destructive Chatwoot custom attribute definition delete [intent=reverse_etl availability=implemented write=delete_custom_attribute_definition]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --id
    custom-filter create - Admin Chatwoot custom filter create [intent=reverse_etl availability=implemented write=create_custom_filter]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --name, --type
    custom-filter update - Admin Chatwoot custom filter update [intent=reverse_etl availability=implemented write=update_custom_filter]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --custom-filter-id, --name, --type
    custom-filter delete - Destructive Chatwoot custom filter delete [intent=reverse_etl availability=implemented write=delete_custom_filter]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --custom-filter-id
    inbox-member create - Admin Chatwoot inbox member create [intent=reverse_etl availability=implemented write=create_inbox_member]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-id, --user-ids
    inbox-member update - Admin Chatwoot inbox member update [intent=reverse_etl availability=implemented write=update_inbox_member]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-id, --user-ids
    inbox-member delete - Destructive Chatwoot inbox member delete [intent=reverse_etl availability=implemented write=delete_inbox_member]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --inbox-id, --user-ids
    integration hook create - Admin Chatwoot integration hook create [intent=reverse_etl availability=implemented write=create_integration_hook]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --app-id, --inbox-id, --status
    integration hook update - Admin Chatwoot integration hook update [intent=reverse_etl availability=implemented write=update_integration_hook]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: admin_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --hook-id, --status
    integration hook delete - Destructive Chatwoot integration hook delete [intent=reverse_etl availability=implemented write=delete_integration_hook]; approval: reverse ETL writes require plan, preview, approval, execute. Destructive operations also require typed confirmation.; risk: high: destructive_action Chatwoot operation; reverse ETL approval required; flags: --hook-id
    profile update - Sensitive Chatwoot profile update [intent=reverse_etl availability=implemented write=update_profile]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: medium: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --profile
    public inbox contact create - Sensitive Chatwoot public inbox contact create [intent=reverse_etl availability=implemented write=create_public_inbox_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-identifier, --identifier, --email, --name, --phone-number, --avatar
    public inbox contact update - Sensitive Chatwoot public inbox contact update [intent=reverse_etl availability=implemented write=update_public_inbox_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-identifier, --contact-identifier, --identifier, --email, --name, --phone-number, --avatar
    public inbox contact conversation create - Sensitive Chatwoot public inbox contact conversation create [intent=reverse_etl availability=implemented write=create_public_inbox_contact_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-identifier, --contact-identifier
    public inbox contact conversation message create - Sensitive Chatwoot public inbox contact conversation message create [intent=reverse_etl availability=implemented write=create_public_inbox_contact_conversation_message]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-identifier, --contact-identifier, --conversation-id, --content, --echo-id
    public inbox contact conversation message update - Sensitive Chatwoot public inbox contact conversation message update [intent=reverse_etl availability=implemented write=update_public_inbox_contact_conversation_message]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-identifier, --contact-identifier, --conversation-id, --message-id
    public inbox contact conversation status toggle - Sensitive Chatwoot public inbox contact conversation status toggle [intent=reverse_etl availability=implemented write=toggle_public_conversation_status]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: high: sensitive_reverse_etl Chatwoot operation; reverse ETL approval required; flags: --inbox-identifier, --contact-identifier, --conversation-id
    agent-bot list - Read agent bot list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    agent-bot view - Read agent bot view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    automation-rule list - Read automation rule list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --page
    automation-rule view - Read automation rule view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    canned-response list - Read canned response list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    custom-attribute-definition list - Read custom attribute definition list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --attribute-model
    custom-attribute-definition view - Read custom attribute definition view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    custom-filter list - Read custom filter list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    custom-filter view - Read custom filter view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --custom-filter-id
    inbox-member list - Read inbox member list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --inbox-id
    integration app list - Read integration app list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    reporting-event list - Read reporting event list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --page, --since, --until, --inbox-id, --user-id, --name
    profile view - Read profile view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    summary-report agent list - Read summary report agent list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    summary-report channel list - Read summary report channel list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    summary-report inbox list - Read summary report inbox list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    summary-report team list - Read summary report team list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    public inbox view - Read public inbox view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --inbox-identifier
    public inbox contact view - Read public inbox contact view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --inbox-identifier, --contact-identifier
    public inbox contact conversation list - Read public inbox contact conversation list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --inbox-identifier, --contact-identifier
    public inbox contact conversation view - Read public inbox contact conversation view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --inbox-identifier, --contact-identifier, --conversation-id
    public inbox contact conversation message list - Read public inbox contact conversation message list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --inbox-identifier, --contact-identifier, --conversation-id
  Help topics:
    chatwoot-auth - Use a saved Chatwoot API access token; never pass token values in command text.
    chatwoot-writes - Chatwoot reverse ETL writes require plan, preview, approval, execute; destructive/admin operations use typed fixed commands and confirmation metadata.
    chatwoot-direct-reads - Direct-read commands are implemented for official Chatwoot GET operations with bounded JSON output, except duplicate/disallowed rows; mutation endpoints remain governed by reverse-ETL policy.

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
