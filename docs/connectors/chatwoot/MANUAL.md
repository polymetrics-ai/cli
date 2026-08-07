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
  create_contact:
    endpoint: POST /contacts
    required fields: inbox_id
    risk: creates a new Chatwoot contact record; low risk, no customer notification
  update_contact:
    endpoint: PUT /contacts/{{ record.id }}
    required fields: id
    risk: updates an existing Chatwoot contact's profile fields; low risk, no customer notification
  create_conversation:
    endpoint: POST /conversations
    required fields: source_id
    risk: creates a new conversation in the target inbox; customer-visible once the initial message is delivered through a live channel
  send_message:
    endpoint: POST /conversations/{{ record.conversation_id }}/messages
    required fields: conversation_id, content
    risk: sends a message into a conversation; customer-visible unless private is true and may notify the contact through the inbox channel
  toggle_conversation_status:
    endpoint: POST /conversations/{{ record.conversation_id }}/toggle_status
    required fields: conversation_id, status
    risk: changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and reporting metrics
  create_label:
    endpoint: POST /labels
    required fields: title
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
