# pm connectors inspect chatwoot

```text
NAME
  pm connectors inspect chatwoot - Chatwoot connector manual

SYNOPSIS
  pm connectors inspect chatwoot
  pm connectors inspect chatwoot --json
  pm credentials add <name> --connector chatwoot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Chatwoot support conversations, contacts, inboxes, agents, teams, labels, and conversation-scoped messages; writes approved contact, conversation, message, and label mutations; and tracks the full official Swagger surface as blocked-by-default operation metadata for future direct-read, binary, and sensitive/admin slices.

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

SECURITY
  read risk: external Chatwoot Application API reads of account-scoped conversation, contact, inbox, agent, team, label, and message data; additional official GET endpoints are direct-read candidates blocked by default until typed bounds exist
  write risk: external mutation of Chatwoot contacts, conversations, messages, and labels; additional official write/admin/destructive endpoints are blocked by default until named reverse-ETL actions, approval text, and typed confirmation policies exist
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Chatwoot support-desk data and approved account-scoped mutations.
  Usage: pm chatwoot <command> <subcommand> [flags]
  Source CLI: chatwoot-api (https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Chatwoot connector credential and account scope.; maps_to=connection
  Support Desk Commands
    conversation list - List account conversations [intent=etl availability=implemented stream=conversations]; flags: --status
    conversation view - View one conversation [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --conversation-id
    conversation create - Create a conversation [intent=reverse_etl availability=implemented write=create_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a new Chatwoot conversation in the configured account and may become customer-visible through the inbox channel.; flags: --source-id, --inbox-id, --contact-id, --status
    conversation toggle-status - Change a conversation status [intent=reverse_etl availability=implemented write=toggle_conversation_status]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Changes a conversation's workflow status and can affect routing, reporting, and agent queues.; flags: --conversation-id, --status, --snoozed-until
    conversation meta view - Read conversation meta view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --status, --q, --inbox-id, --team-id, --labels
    conversation label list - Read conversation label list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --conversation-id
    conversation reporting-event list - Read conversation reporting event list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --conversation-id
    contact list - List contacts [intent=etl availability=implemented stream=contacts]
    contact view - View one contact [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    contact search - Search contacts [intent=direct_read availability=implemented]; notes: Bounded query direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --q
    contact create - Create a contact [intent=reverse_etl availability=implemented write=create_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Chatwoot contact record; visible to agents and usable in conversations.; flags: --inbox-id, --name, --email, --phone-number
    contact update - Update a contact [intent=reverse_etl availability=implemented write=update_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates an existing Chatwoot contact profile; visible to agents and downstream automations.; flags: --id, --name, --email, --phone-number
    contact delete - Delete a contact [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked until a destructive typed reverse-ETL action with confirmation is approved.; risk: Deletes a contact record and can remove conversation context.; notes: No generic direct-write fallback is exposed.
    contact contactable-inbox list - Read contact contactable inbox list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    contact conversation list - Read contact conversation list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    contact label list - Read contact label list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    message list - List messages for each conversation [intent=etl availability=implemented stream=messages]; notes: The messages stream fans out across conversations; single-conversation message direct reads are planned separately.
    message send - Send a conversation message [intent=reverse_etl availability=implemented write=send_message]; approval: reverse ETL writes require plan, preview, approval, execute. Attachment/multipart variants remain blocked until the binary policy slice.; risk: Sends a message into a Chatwoot conversation; customer-visible unless private is true and may trigger notifications.; flags: --conversation-id, --content, --message-type, --private
  Workspace Metadata Commands
    inbox list - List inboxes [intent=etl availability=implemented stream=inboxes]
    inbox view - Read inbox view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    inbox agent-bot list - Read inbox agent bot list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    agent list - List account agents [intent=etl availability=implemented stream=agents]
    team list - List teams [intent=etl availability=implemented stream=teams]
    team view - Read team view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --team-id
    team team-member list - Read team team member list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --team-id
    label list - List labels [intent=etl availability=implemented stream=labels]
    label create - Create a label [intent=reverse_etl availability=implemented write=create_label]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates an account-wide label visible to Chatwoot agents.; flags: --title, --description, --color, --show-on-sidebar
    label view - Read label view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
  Blocked Admin And Configuration Commands
    account view - Read account view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    account update - Update account settings [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy and typed confirmation are implemented.; risk: Mutates account configuration.
    automation rule create - Create an automation rule [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy and typed confirmation are implemented.; risk: Changes automated Chatwoot behavior.
    webhook list - Read webhook list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    webhook create - Create an outbound webhook [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy, endpoint validation, and typed confirmation are implemented.; risk: Creates an outbound integration that can exfiltrate event payloads.
    portal list - Read portal list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    portal create - Create a help-center portal [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy and typed confirmation are implemented.; risk: Creates help-center content configuration.
    platform account view - Read platform account view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --account-id
    platform account-user list - Read platform account user list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --account-id
    platform agent-bot list - Read platform agent bot list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    platform agent-bot view - Read platform agent bot view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    platform user view - Read platform user view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    platform user login view - Read platform user login view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
    platform account create - Create a platform account [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; no generic platform write command is exposed.; risk: Requires elevated Platform API credentials and provisions a new account.
  Planned Reporting And Audit Commands
    report account - Read report account [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report conversation list - Read report conversation list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report conversation list-alt - Read report conversation list alt [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report first-response-time-distribution list - Read report first response time distribution list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report inbox-label-matrix list - Read report inbox label matrix list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    report outgoing-messages-count list - Read report outgoing messages count list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --group-by
    report summary view - Read report summary view [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.
    audit-log list - Read audit log list [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --page
  Other Commands
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
    chatwoot-writes - Chatwoot reverse ETL writes require plan, preview, approval, execute; destructive/admin operations remain blocked until typed policies exist.
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
