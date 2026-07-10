---
name: pm-chatwoot
description: Chatwoot connector knowledge and safe action guide.
---

# pm-chatwoot

## Purpose

Reads Chatwoot support conversations, contacts, inboxes, agents, teams, labels, and conversation-scoped messages; writes approved contact, conversation, message, and label mutations; and tracks the full official Swagger surface as blocked-by-default operation metadata for future direct-read, binary, and sensitive/admin slices.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- start_date
- api_access_token (secret)

## ETL Streams

- conversations:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(), additional_attributes(), agent_last_seen_at(), assignee_last_seen_at(), can_reply(), contact_last_seen_at(), created_at(), custom_attributes(), first_reply_created_at(), id(), inbox_id(), labels(), last_activity_at(), muted(), priority(), sla_policy_id(), snoozed_until(), status(), timestamp(), unread_count(), updated_at(), uuid(), waiting_since()
- contacts:
  - primary key: id
  - cursor: last_activity_at
  - fields: additional_attributes(), availability_status(), blocked(), contact_inboxes(), created_at(), custom_attributes(), email(), id(), identifier(), last_activity_at(), name(), phone_number(), thumbnail()
- inboxes:
  - primary key: id
  - fields: allow_messages_after_resolved(), avatar_url(), business_name(), callback_webhook_url(), channel_id(), channel_type(), csat_survey_enabled(), enable_auto_assignment(), enable_email_collect(), greeting_enabled(), greeting_message(), id(), lock_to_single_conversation(), medium(), name(), out_of_office_message(), phone_number(), provider(), timezone(), website_token(), website_url(), welcome_tagline(), welcome_title(), widget_color(), working_hours_enabled()
- agents:
  - primary key: id
  - fields: account_id(), auto_offline(), availability_status(), available_name(), confirmed(), custom_role_id(), email(), id(), name(), role(), thumbnail()
- teams:
  - primary key: id
  - fields: account_id(), allow_auto_assign(), description(), id(), is_member(), name()
- labels:
  - primary key: id
  - fields: color(), description(), id(), show_on_sidebar(), title()
- messages:
  - primary key: id
  - cursor: id
  - fields: account_id(), attachment(), content(), content_attributes(), content_type(), conversation_id(), created_at(), id(), inbox_id(), message_type(), private(), sender(), sender_id(), sender_type(), source_id(), status(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_contact:
  - endpoint: POST /contacts
  - risk: creates a new Chatwoot contact record; low risk, no customer notification
- update_contact:
  - endpoint: PUT /contacts/{{ record.id }}
  - required fields: id
  - risk: updates an existing Chatwoot contact's profile fields; low risk, no customer notification
- create_conversation:
  - endpoint: POST /conversations
  - risk: creates a new conversation in the target inbox; customer-visible once the initial message is delivered through a live channel
- send_message:
  - endpoint: POST /conversations/{{ record.conversation_id }}/messages
  - required fields: conversation_id
  - risk: sends a message into a conversation; customer-visible unless private is true and may notify the contact through the inbox channel
- toggle_conversation_status:
  - endpoint: POST /conversations/{{ record.conversation_id }}/toggle_status
  - required fields: conversation_id
  - risk: changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and reporting metrics
- create_label:
  - endpoint: POST /labels
  - risk: creates a new account-wide label; low risk, visible to all agents in the sidebar when show_on_sidebar is true

## Security

- read risk: external Chatwoot Application API reads of account-scoped conversation, contact, inbox, agent, team, label, and message data; additional official GET endpoints are direct-read candidates blocked by default until typed bounds exist
- write risk: external mutation of Chatwoot contacts, conversations, messages, and labels; additional official write/admin/destructive endpoints are blocked by default until named reverse-ETL actions, approval text, and typed confirmation policies exist
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Work with Chatwoot support-desk data and approved account-scoped mutations.
- Usage: pm chatwoot <command> <subcommand> [flags]
- Source CLI: chatwoot-api (https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json)
- Global flags:
  - --json (boolean): Write machine-readable JSON output.
  - --connection (string): Use a saved Chatwoot connector credential and account scope.; maps_to=connection
- Support Desk Commands
  - conversation list - List account conversations [intent=etl availability=implemented stream=conversations]; flags: --status
  - conversation view - View one conversation [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --conversation-id
  - conversation create - Create a conversation [intent=reverse_etl availability=implemented write=create_conversation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a new Chatwoot conversation in the configured account and may become customer-visible through the inbox channel.; flags: --source-id, --inbox-id, --contact-id, --status
  - conversation toggle-status - Change a conversation status [intent=reverse_etl availability=implemented write=toggle_conversation_status]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Changes a conversation's workflow status and can affect routing, reporting, and agent queues.; flags: --conversation-id, --status, --snoozed-until
  - contact list - List contacts [intent=etl availability=implemented stream=contacts]
  - contact view - View one contact [intent=direct_read availability=implemented]; notes: Bounded direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --id
  - contact search - Search contacts [intent=direct_read availability=implemented]; notes: Bounded query direct read; response JSON is capped and secret-shaped keys are redacted.; flags: --q
  - contact create - Create a contact [intent=reverse_etl availability=implemented write=create_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Chatwoot contact record; visible to agents and usable in conversations.; flags: --inbox-id, --name, --email, --phone-number
  - contact update - Update a contact [intent=reverse_etl availability=implemented write=update_contact]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates an existing Chatwoot contact profile; visible to agents and downstream automations.; flags: --id, --name, --email, --phone-number
  - contact delete - Delete a contact [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked until a destructive typed reverse-ETL action with confirmation is approved.; risk: Deletes a contact record and can remove conversation context.; notes: No generic direct-write fallback is exposed.
  - message list - List messages for each conversation [intent=etl availability=implemented stream=messages]; notes: The messages stream fans out across conversations; single-conversation message direct reads are planned separately.
  - message send - Send a conversation message [intent=reverse_etl availability=implemented write=send_message]; approval: reverse ETL writes require plan, preview, approval, execute. Attachment/multipart variants remain blocked until the binary policy slice.; risk: Sends a message into a Chatwoot conversation; customer-visible unless private is true and may trigger notifications.; flags: --conversation-id, --content, --message-type, --private
- Workspace Metadata Commands
  - inbox list - List inboxes [intent=etl availability=implemented stream=inboxes]
  - agent list - List account agents [intent=etl availability=implemented stream=agents]
  - team list - List teams [intent=etl availability=implemented stream=teams]
  - label list - List labels [intent=etl availability=implemented stream=labels]
  - label create - Create a label [intent=reverse_etl availability=implemented write=create_label]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates an account-wide label visible to Chatwoot agents.; flags: --title, --description, --color, --show-on-sidebar
- Blocked Admin And Configuration Commands
  - account update - Update account settings [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy and typed confirmation are implemented.; risk: Mutates account configuration.
  - automation rule create - Create an automation rule [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy and typed confirmation are implemented.; risk: Changes automated Chatwoot behavior.
  - webhook create - Create an outbound webhook [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy, endpoint validation, and typed confirmation are implemented.; risk: Creates an outbound integration that can exfiltrate event payloads.
  - portal create - Create a help-center portal [intent=reverse_etl availability=planned]; approval: Blocked until admin reverse-ETL policy and typed confirmation are implemented.; risk: Creates help-center content configuration.
  - platform account create - Create a platform account [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; no generic platform write command is exposed.; risk: Requires elevated Platform API credentials and provisions a new account.
- Planned Reporting And Audit Commands
  - report account - Read account report metrics [intent=direct_read availability=planned]; notes: Blocked until report direct reads add required bounds and redacted output policy.
  - audit-log list - List audit logs [intent=direct_read availability=planned]; notes: Blocked until direct reads support bounded audit-log pagination and redaction.
- Help topics:
  - chatwoot-auth - Use a saved Chatwoot API access token; never pass token values in command text.
  - chatwoot-writes - Chatwoot reverse ETL writes require plan, preview, approval, execute; destructive/admin operations remain blocked until typed policies exist.
  - chatwoot-direct-reads - Direct-read commands are implemented for selected conversation/contact lookups with bounded JSON output; reports, audit logs, public inbox reads, and binary/file endpoints remain planned.

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
