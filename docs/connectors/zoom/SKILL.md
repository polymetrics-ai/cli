---
name: pm-zoom
description: Zoom connector knowledge and safe action guide.
---

# pm-zoom

## Purpose

Reads Zoom users, meetings, webinars, bounded module-specific data, and SCIM2 groups/users through the Zoom REST API; includes sensitive Cobrowse SDK and SCIM2 reads, approval-gated Chatbot, SCIM2, clinical-note, and Quality Management actions, and a redacted Customer Managed Keys Hybrid archival key action.

## Icon

- id: zoom
- asset: icons/zoom.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.zoom.us/docs/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- chatbot_token_url
- key_connector_base_url
- max_pages
- mode
- page_size
- scim2_base_url
- user_id
- access_token (secret)
- chatbot_client_id (secret)
- chatbot_client_secret (secret)
- key_connector_jwt (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: email(string), id(string), name(string), updated_at(string)
- meetings:
  - primary key: id
  - fields: email(string), id(string), name(string), updated_at(string)
- webinars:
  - primary key: id
  - fields: email(string), id(string), name(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- update_clinical_note:
  - endpoint: PATCH /clinical_notes/notes/{{ record.note_id }}
  - required fields: note_id, is_note_completed
  - risk: high: mutates a patient's clinical note completion status; requires reverse ETL approval
- create_quality_management_interaction:
  - endpoint: POST /qm/interactions
  - required fields: download_url
  - risk: high: imports a third-party interaction into Zoom Quality Management; requires reverse ETL approval

## Security

- read risk: external Zoom API read of user, meeting, webinar, Quality of Service, AI Companion, My Notes, healthcare clinical-note, Quality Management, Cobrowse SDK session, and SCIM2 user/group data
- write risk: typed Zoom Chatbot message, Link Unfurls, SCIM2 group/user, healthcare clinical-note, and Quality Management actions, plus Customer Managed Keys Hybrid archival key decryption with redacted output
- approval: mutating commands require plan, preview, explicit approval, and execute; Chatbot and SCIM2 deletion plus Customer Managed Keys Hybrid archival key decryption additionally require typed confirmation; read-only commands require none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run declared Zoom stream reads, bounded module-specific direct reads, approval-gated Chatbot, SCIM2, clinical-note, and Quality Management actions, and a redacted Customer Managed Keys Hybrid archival key action.
- Usage: pm zoom <group> <command> [flags]
- Source CLI: Zoom API reference (OpenAPI 3.1.1; docs static build 2026-08-03T14-58-19-06-00; retrieved 2026-08-05)
- Global flags:
  - --credential (string): Credential name to use for the Zoom request.
  - --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum records to emit from a stream command.
- Users
  - users list - Read Zoom users through the declared ETL stream. [intent=etl availability=implemented stream=users]
- Meetings and webinars
  - meetings list - Read meetings for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=meetings]; flags: --user-id
  - webinars list - Read webinars for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=webinars]; flags: --user-id
- Quality of Service Summaries
  - qss meeting-participants list - List a past meeting's participants' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_meeting_participants_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required meeting-id path parameter.; flags: --meeting-id (required)
  - qss webinar-participants list - List a live or past webinar's participants' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_webinar_participants_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required webinar-id path parameter.; flags: --webinar-id (required)
  - qss session-users list - List a past Video SDK session's users' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_session_users_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required session-id path parameter.; flags: --session-id (required)
- AI Companion
  - ai-companion conversation-archive get - Get a user's AI Companion conversation archive. [intent=direct_read availability=implemented operation=zoom.get_ai_companion_conversation_archives]; notes: Bounded Zoom read; fixed method and path with a typed required user-id path parameter. Response download-URL fields are redacted by the json_redacted output policy.; flags: --user-id (required)
- My Notes
  - my-notes list - List the authenticated user's My Notes. [intent=direct_read availability=implemented operation=zoom.list_my_notes]; notes: Bounded Zoom read; fixed method and path with no request parameters (the live artifact documents none for this operation).
  - my-notes content get - Get a My Notes note's content, and optionally its meeting transcript. [intent=direct_read availability=implemented operation=zoom.get_my_notes_content]; notes: Bounded Zoom read; fixed method and path with a typed required note-id path parameter and an optional include=transcript query flag (explicitly documented in provider prose, unlike qss's response-only pagination fields).; flags: --note-id (required), --include
- Healthcare
  - healthcare clinical-notes list - List clinical notes, optionally filtered by owner and/or meeting. [intent=direct_read availability=implemented operation=zoom.list_clinical_notes]; notes: Bounded sensitive Zoom read. Only note-owner-user-id and meeting-id are explicit provider request inputs; response-only dates and paging fields are not CLI flags. The clinical_json_redacted policy redacts clinical-note content and identifiers.; flags: --note-owner-user-id, --meeting-id
  - healthcare clinical-notes get - Get a single clinical note by ID. [intent=direct_read availability=implemented operation=zoom.get_clinical_note]; notes: Bounded sensitive Zoom read with a typed required note-id path parameter. The clinical_json_redacted policy redacts clinical-note content and identifiers.; flags: --note-id (required)
  - healthcare clinical-notes update - Plan an update to a clinical note's completion status. [intent=reverse_etl availability=implemented write=update_clinical_note]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: high: changes a patient's clinical note completion status through an approval-gated reverse ETL action; notes: Typed high-risk mutation; preview and explicit approval are required before execute. The clinical note ID is redacted in write errors.; flags: --note-id (required), --is-note-completed (required)
- Quality Management
  - quality-management automated-evaluations list - List Quality Management automated evaluations. [intent=direct_read availability=implemented operation=zoom.list_quality_management_automated_evaluations]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination fields are not CLI flags; identifiers and personal-contact fields are redacted before output.
  - quality-management evaluations list - List completed Quality Management evaluations. [intent=direct_read availability=implemented operation=zoom.list_quality_management_evaluations]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination fields are not CLI flags; identifiers and personal-contact fields are redacted before output.
  - quality-management evaluations get - View one Quality Management evaluation. [intent=direct_read availability=implemented operation=zoom.get_quality_management_evaluation]; notes: Bounded sensitive Zoom read with a typed required evaluation-id path parameter; identifiers and personal-contact fields are redacted before output.; flags: --evaluation-id (required)
  - quality-management interactions list - List Quality Management interactions. [intent=direct_read availability=implemented operation=zoom.list_quality_management_interactions]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination/date fields are not CLI flags; identifiers and personal-contact fields are redacted before output.
  - quality-management interactions get - View one Quality Management interaction. [intent=direct_read availability=implemented operation=zoom.get_quality_management_interaction]; notes: Bounded sensitive Zoom read with a typed required interaction-id path parameter; identifiers and personal-contact fields are redacted before output.; flags: --interaction-id (required)
  - quality-management interactions create - Plan creation of a Quality Management interaction from a third-party download URL. [intent=reverse_etl availability=implemented write=create_quality_management_interaction]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: high: imports a third-party interaction into Zoom Quality Management through an approval-gated reverse ETL action; notes: Typed high-risk mutation. Download URL and interaction-info fields are redacted in generic write errors; preview and explicit approval are required before execute. If any interaction-info field is supplied, Zoom requires interaction-channel-type.; flags: --download-url (required), --direction, --disposition, --interaction-channel-type, --interaction-agent-email, --interaction-agent-id, --interaction-consumer-name, --interaction-from, --interaction-to, --primary-language, --queue-id, --start-time
- Cobrowse SDK
  - cobrowse-sdk live-sessions list - List live Cobrowse SDK sessions for an optional monthly date range. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_live_sessions]; notes: Bounded sensitive Zoom read. The provider explicitly permits an optional monthly from/to date range; page_size and next_page_token are response-only fields and are not CLI flags.; flags: --from, --to
  - cobrowse-sdk past-sessions list - List past Cobrowse SDK sessions for an optional monthly date range. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_past_sessions]; notes: Bounded sensitive Zoom read. The provider explicitly permits an optional monthly from/to date range; page_size and next_page_token are response-only fields and are not CLI flags.; flags: --from, --to
  - cobrowse-sdk sessions get - Get details for one Cobrowse SDK session. [intent=direct_read availability=implemented operation=zoom.get_cobrowse_session]; notes: Bounded sensitive Zoom read with a typed required session-id path parameter; session pins, user/session identifiers, display names, connection IDs, and IP addresses are redacted before output.; flags: --session-id (required)
  - cobrowse-sdk sessions users list - List users from one Cobrowse SDK session. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_session_users]; notes: Bounded sensitive Zoom read with a typed required session-id path parameter; page_size and next_page_token are response-only fields and session/user connection data is redacted before output.; flags: --session-id (required)
- Chatbot
  - chatbot messages send - Send one Zoom Chatbot message. [intent=direct_write availability=implemented operation=zoom.send_chatbot_message]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Sends provider-defined Chatbot content and JID/account identifiers through Zoom.; notes: Uses the declared Chatbot-only client-credentials token exchange with HTTP Basic client authentication; it never reuses the ordinary Zoom access token. Message and JID values are redacted in plans, errors, and JSON output.; flags: --account-id (required), --content (required), --robot-jid (required), --to-jid (required), --user-jid (required), --is-markdown-support, --reply-to, --visible-to-user
  - chatbot messages edit - Edit one Zoom Chatbot message. [intent=direct_write availability=implemented operation=zoom.edit_chatbot_message]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Changes provider-defined Chatbot content and carries account/JID identifiers.; notes: Uses the declared Chatbot-only client-credentials token exchange with HTTP Basic client authentication; it never reuses the ordinary Zoom access token. Message and JID values are redacted in plans, errors, and JSON output.; flags: --message-id (required), --account-id (required), --content (required), --robot-jid (required), --is-markdown-support, --user-jid
  - chatbot messages delete - Delete one Zoom Chatbot message. [intent=direct_write availability=implemented operation=zoom.delete_chatbot_message]; approval: Requires plan, no-network preview, explicit single-use approval, and destructive typed confirmation before execute.; risk: Irreversibly deletes a Zoom Chatbot message.; notes: Uses the declared Chatbot-only client-credentials token exchange with HTTP Basic client authentication; it never reuses the ordinary Zoom access token. The message identifier and JSON response are redacted.; flags: --message-id (required)
  - chatbot link-unfurls create - Create one Zoom Chatbot Link Unfurls action. [intent=direct_write availability=implemented operation=zoom.create_chatbot_link_unfurl]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Posts provider-defined Link Unfurls content for a Chatbot trigger.; notes: Uses the declared Chatbot-only client-credentials token exchange with HTTP Basic client authentication; it never reuses the ordinary Zoom access token. Zoom returns 204 No Content, which is recorded as a successful status-only action.; flags: --user-id (required), --trigger-id (required), --content (required)
- SCIM2
  - scim2 groups list - List Zoom SCIM2 groups. [intent=direct_read availability=implemented operation=zoom.list_scim2_groups]; notes: Bounded sensitive Zoom SCIM2 read with a fixed provider root path. The live artifact does not declare a standalone paging input, so no paging flag is exposed.
  - scim2 groups create - Create one Zoom SCIM2 group. [intent=direct_write availability=implemented operation=zoom.create_scim2_group]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Creates a Zoom SCIM2 group with provider-defined membership and alias data.; notes: The named resource flag accepts exactly one documented SCIM Group object for this fixed endpoint; it is not a generic JSON or HTTP escape hatch. Group values are redacted in previews, errors, and JSON output.; flags: --resource (required)
  - scim2 groups get - Get one Zoom SCIM2 group. [intent=direct_read availability=implemented operation=zoom.get_scim2_group]; notes: Bounded sensitive Zoom SCIM2 read with a typed required group-id path parameter. The live artifact documents no standalone paging input.; flags: --group-id (required)
  - scim2 groups delete - Delete one Zoom SCIM2 group. [intent=direct_write availability=implemented operation=zoom.delete_scim2_group]; approval: Requires plan, no-network preview, explicit single-use approval, and destructive typed confirmation before execute.; risk: Irreversibly deletes a Zoom SCIM2 group.; notes: The fixed SCIM2 delete endpoint returns documented 204 No Content. Success is asserted by status without inventing a response body.; flags: --group-id (required)
  - scim2 groups update - Update one Zoom SCIM2 group. [intent=direct_write availability=implemented operation=zoom.update_scim2_group]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Changes Zoom SCIM2 group membership or metadata through a provider-defined PatchOp.; notes: The named patch flag accepts exactly one documented SCIM PatchOp object for this fixed endpoint; it is not a generic JSON or HTTP escape hatch. The documented 204 response is recorded as a status-only success.; flags: --group-id (required), --patch (required)
  - scim2 users list - List Zoom SCIM2 users. [intent=direct_read availability=implemented operation=zoom.list_scim2_users]; notes: Bounded sensitive Zoom SCIM2 read with a fixed provider root path. The live artifact does not declare a standalone paging input, so no paging flag is exposed.
  - scim2 users create - Create one Zoom SCIM2 user. [intent=direct_write availability=implemented operation=zoom.create_scim2_user]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Creates a Zoom SCIM2 user with provider-defined profile, contact, organization, and extension data.; notes: The named resource flag accepts exactly one documented SCIM User object for this fixed endpoint; it is not a generic JSON or HTTP escape hatch. User and account values are redacted in previews, errors, and JSON output.; flags: --resource (required)
  - scim2 users get - Get one Zoom SCIM2 user. [intent=direct_read availability=implemented operation=zoom.get_scim2_user]; notes: Bounded sensitive Zoom SCIM2 read with a typed required user-id path parameter. The live artifact documents no standalone paging input.; flags: --user-id (required)
  - scim2 users update - Update one Zoom SCIM2 user. [intent=direct_write availability=implemented operation=zoom.update_scim2_user]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Replaces Zoom SCIM2 user profile, contact, organization, or extension data.; notes: The named resource flag accepts exactly one documented SCIM User object for this fixed endpoint; it is not a generic JSON or HTTP escape hatch. User and account values are redacted in previews, errors, and JSON output.; flags: --user-id (required), --resource (required)
  - scim2 users delete - Delete one Zoom SCIM2 user. [intent=direct_write availability=implemented operation=zoom.delete_scim2_user]; approval: Requires plan, no-network preview, explicit single-use approval, and destructive typed confirmation before execute.; risk: Irreversibly deletes a Zoom SCIM2 user.; notes: The fixed SCIM2 delete endpoint returns documented 204 No Content. Success is asserted by status without inventing a response body.; flags: --user-id (required)
  - scim2 users deactivate - Deactivate one Zoom SCIM2 user. [intent=direct_write availability=implemented operation=zoom.deactivate_scim2_user]; approval: Requires plan, no-network preview, and explicit single-use approval before execute.; risk: Changes a Zoom SCIM2 user's active state through a provider-defined PatchOp.; notes: The named patch flag accepts exactly one documented SCIM activation-state PatchOp object for this fixed endpoint; it is not a generic JSON or HTTP escape hatch. User and account values are redacted in previews, errors, and JSON output.; flags: --user-id (required), --patch (required)
- Customer Managed Keys Hybrid
  - customer-managed-keys-hybrid archival-key decrypt - Decrypt one Customer Managed Keys Hybrid archival data key. [intent=direct_write availability=implemented operation=zoom.decrypt_customer_managed_key_archival]; approval: Requires plan, no-network preview, explicit single-use approval, typed confirmation, then execute.; risk: Returns sensitive key material only through a redacted output policy.; notes: Requires an operator-provisioned Customer Managed Keys Hybrid credential with key_connector_base_url ending in /api/v2 and an environment- or stdin-supplied key_connector_jwt. The operation declares its own customer-hosted origin and bearer; it never reuses the ordinary Zoom OAuth bearer.; flags: --encrypt-context (required), --key-id (required)
- Help topics:
  - provider-inventory - The Zoom provider ledger tracks 1,913 documented REST operations; Wave 1 executes three stream-backed reads; Wave 2+ adds bounded direct-read/write operations module by module (see #3915).

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoom
```

### Inspect as structured JSON

```bash
pm connectors inspect zoom --json
```

## Agent Rules

- Run pm connectors inspect zoom before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
