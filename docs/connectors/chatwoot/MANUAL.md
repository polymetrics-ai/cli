# pm connectors inspect chatwoot

```text
NAME
  pm connectors inspect chatwoot - Chatwoot connector manual

SYNOPSIS
  pm connectors inspect chatwoot
  pm connectors inspect chatwoot --json
  pm credentials add <name> --connector chatwoot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Chatwoot Support conversations, contacts, inboxes, agents, teams, labels, and conversation-scoped messages, and writes contact/conversation/message/label mutations through the Chatwoot Application API.

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
    cursor: created_at
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

SECURITY
  read risk: external Chatwoot Application API read of conversation, contact, and message data (account-scoped)
  write risk: external mutation of Chatwoot contacts, conversations, messages, and labels; agent-visible and customer-visible side effects
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
