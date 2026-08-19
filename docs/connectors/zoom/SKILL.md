---
name: pm-zoom
description: Zoom connector knowledge and safe action guide.
---

# pm-zoom

## Purpose

Reads Zoom users, meetings, and webinars through the Zoom REST API.

## Icon

- id: zoom
- asset: icons/zoom.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.zoom.us/docs/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- user_id
- access_token (secret) (required)

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

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Zoom API read of user, meeting, and webinar data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run existing Zoom stream reads and 70 bounded module-specific direct reads.
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
  - qss meeting-participants list - List a past meeting's participants' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_meeting_participants_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required meeting-id path parameter.; flags: --meeting-id (required), --page, --page-cursor
  - qss webinar-participants list - List a live or past webinar's participants' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_webinar_participants_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required webinar-id path parameter.; flags: --webinar-id (required), --page, --page-cursor
  - qss session-users list - List a past Video SDK session's users' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_session_users_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required session-id path parameter.; flags: --session-id (required), --page, --page-cursor
- AI Companion
  - ai-companion conversation-archive get - Get a user's AI Companion conversation archive. [intent=direct_read availability=implemented operation=zoom.get_ai_companion_conversation_archives]; notes: Bounded Zoom read; fixed method and path with a typed required user-id path parameter. Response download-URL fields are redacted by the json_redacted output policy.; flags: --user-id (required), --page, --page-cursor
- My Notes
  - my-notes list - List the authenticated user's My Notes. [intent=direct_read availability=implemented operation=zoom.list_my_notes]; notes: Bounded Zoom read; fixed method and path with no request parameters (the live artifact documents none for this operation).; flags: --page, --page-cursor
  - my-notes content get - Get a My Notes note's content, and optionally its meeting transcript. [intent=direct_read availability=implemented operation=zoom.get_my_notes_content]; notes: Bounded Zoom read; fixed method and path with a typed required note-id path parameter and an optional include=transcript query flag (explicitly documented in provider prose, unlike qss's response-only pagination fields).; flags: --note-id (required), --include, --page, --page-cursor
- Healthcare
  - healthcare clinical-notes list - List clinical notes, optionally filtered by owner and/or meeting. [intent=direct_read availability=implemented operation=zoom.list_clinical_notes]; notes: Bounded sensitive Zoom read. Only note-owner-user-id and meeting-id are explicit provider request inputs; response-only dates and paging fields are not CLI flags. The clinical_json_redacted policy redacts clinical-note content and identifiers.; flags: --note-owner-user-id, --meeting-id, --page, --page-cursor
  - healthcare clinical-notes get - Get a single clinical note by ID. [intent=direct_read availability=implemented operation=zoom.get_clinical_note]; notes: Bounded sensitive Zoom read with a typed required note-id path parameter. The clinical_json_redacted policy redacts clinical-note content and identifiers.; flags: --note-id (required), --page, --page-cursor
- Quality Management
  - quality-management automated-evaluations list - List Quality Management automated evaluations. [intent=direct_read availability=implemented operation=zoom.list_quality_management_automated_evaluations]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination fields are not CLI flags; identifiers and personal-contact fields are redacted before output.; flags: --page, --page-cursor
  - quality-management evaluations list - List completed Quality Management evaluations. [intent=direct_read availability=implemented operation=zoom.list_quality_management_evaluations]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination fields are not CLI flags; identifiers and personal-contact fields are redacted before output.; flags: --page, --page-cursor
  - quality-management evaluations get - View one Quality Management evaluation. [intent=direct_read availability=implemented operation=zoom.get_quality_management_evaluation]; notes: Bounded sensitive Zoom read with a typed required evaluation-id path parameter; identifiers and personal-contact fields are redacted before output.; flags: --evaluation-id (required), --page, --page-cursor
  - quality-management interactions list - List Quality Management interactions. [intent=direct_read availability=implemented operation=zoom.list_quality_management_interactions]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination/date fields are not CLI flags; identifiers and personal-contact fields are redacted before output.; flags: --page, --page-cursor
  - quality-management interactions get - View one Quality Management interaction. [intent=direct_read availability=implemented operation=zoom.get_quality_management_interaction]; notes: Bounded sensitive Zoom read with a typed required interaction-id path parameter; identifiers and personal-contact fields are redacted before output.; flags: --interaction-id (required), --page, --page-cursor
- Cobrowse SDK
  - cobrowse-sdk live-sessions list - List live Cobrowse SDK sessions for an optional monthly date range. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_live_sessions]; notes: Bounded sensitive Zoom read. The provider explicitly permits an optional monthly from/to date range; page_size and next_page_token are response-only fields and are not CLI flags.; flags: --from, --to, --page, --page-cursor
  - cobrowse-sdk past-sessions list - List past Cobrowse SDK sessions for an optional monthly date range. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_past_sessions]; notes: Bounded sensitive Zoom read. The provider explicitly permits an optional monthly from/to date range; page_size and next_page_token are response-only fields and are not CLI flags.; flags: --from, --to, --page, --page-cursor
  - cobrowse-sdk sessions get - Get details for one Cobrowse SDK session. [intent=direct_read availability=implemented operation=zoom.get_cobrowse_session]; notes: Bounded sensitive Zoom read with a typed required session-id path parameter; session pins, user/session identifiers, display names, connection IDs, and IP addresses are redacted before output.; flags: --session-id (required), --page, --page-cursor
  - cobrowse-sdk sessions users list - List users from one Cobrowse SDK session. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_session_users]; notes: Bounded sensitive Zoom read with a typed required session-id path parameter; page_size and next_page_token are response-only fields and session/user connection data is redacted before output.; flags: --session-id (required), --page, --page-cursor
- SCIM2
  - scim2 groups list - List Zoom SCIM2 groups. [intent=direct_read availability=implemented operation=zoom.list_scim2_groups]; notes: Bounded sensitive Zoom SCIM2 read with a fixed provider root path. The live artifact does not declare a standalone paging input, so no paging flag is exposed.; flags: --page, --page-cursor
  - scim2 groups get - Get one Zoom SCIM2 group. [intent=direct_read availability=implemented operation=zoom.get_scim2_group]; notes: Bounded sensitive Zoom SCIM2 read with a typed required group-id path parameter. The live artifact documents no standalone paging input.; flags: --group-id (required), --page, --page-cursor
  - scim2 users list - List Zoom SCIM2 users. [intent=direct_read availability=implemented operation=zoom.list_scim2_users]; notes: Bounded sensitive Zoom SCIM2 read with a fixed provider root path. The live artifact does not declare a standalone paging input, so no paging flag is exposed.; flags: --page, --page-cursor
  - scim2 users get - Get one Zoom SCIM2 user. [intent=direct_read availability=implemented operation=zoom.get_scim2_user]; notes: Bounded sensitive Zoom SCIM2 read with a typed required user-id path parameter. The live artifact documents no standalone paging input.; flags: --user-id (required), --page, --page-cursor
- Virtual Agent
  - virtual-agent knowledge-bases articles list - List Zoom Virtual Agent knowledge base articles. [intent=direct_read availability=implemented operation=zoom.list_virtual_agent_articles]; notes: Bounded sensitive Virtual Agent read with a fixed knowledge-base path. The live artifact declares paging values only in response schemas, so no paging, token, date, or guessed query flag is exposed.; flags: --kb-id (required), --page, --page-cursor
  - virtual-agent knowledge-bases articles get - Get one Zoom Virtual Agent knowledge base article. [intent=direct_read availability=implemented operation=zoom.get_virtual_agent_article]; notes: Bounded sensitive Virtual Agent read with fixed knowledge-base and article path parameters. The live artifact declares no standalone request paging input.; flags: --kb-id (required), --article-id (required), --page, --page-cursor
  - virtual-agent knowledge-bases sync get - Get one Zoom Virtual Agent knowledge base sync. [intent=direct_read availability=implemented operation=zoom.get_virtual_agent_sync]; notes: Bounded sensitive Virtual Agent read with fixed knowledge-base and sync path parameters. The live artifact declares no standalone request paging input.; flags: --kb-id (required), --sync-id (required), --page, --page-cursor
  - virtual-agent reports engagements list - List Zoom Virtual Agent engagements. [intent=direct_read availability=implemented operation=zoom.list_virtual_agent_engagements]; notes: Bounded sensitive Virtual Agent report read. The live artifact declares paging/date values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
  - virtual-agent reports engagements query-details list - List Zoom Virtual Agent engagement query details. [intent=direct_read availability=implemented operation=zoom.list_virtual_agent_engagement_query_details]; notes: Bounded sensitive Virtual Agent report read. The live artifact declares paging/date values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
  - virtual-agent reports engagements variable-details list - List Zoom Virtual Agent engagement variable details. [intent=direct_read availability=implemented operation=zoom.list_virtual_agent_engagement_variable_details]; notes: Bounded sensitive Virtual Agent report read. The live artifact declares paging/date values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
  - virtual-agent reports surveys list - List Zoom Virtual Agent surveys. [intent=direct_read availability=implemented operation=zoom.list_virtual_agent_surveys]; notes: Bounded sensitive Virtual Agent report read. The live artifact declares paging/date values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
  - virtual-agent reports transcripts list - List Zoom Virtual Agent transcripts. [intent=direct_read availability=implemented operation=zoom.list_virtual_agent_transcripts]; notes: Bounded sensitive Virtual Agent report read. The live artifact declares paging/date values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
  - virtual-agent reports operation-logs list - List Zoom AI Management operation logs. [intent=direct_read availability=implemented operation=zoom.list_virtual_agent_operation_logs]; notes: Bounded sensitive AI Management report read under Zoom's Virtual Agent artifact. The live artifact declares paging/date values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
- Auto Dialer
  - auto-dialer call-histories get - Get one Zoom Auto Dialer call-history record. [intent=direct_read availability=implemented operation=zoom.get_auto_dialer_call_history]; notes: Bounded sensitive Auto Dialer read at a fixed path. The provider artifact declares no request paging, token, date, or query input.; flags: --call-history-id (required), --page, --page-cursor
  - auto-dialer call-history list - List Zoom Auto Dialer call history. [intent=direct_read availability=implemented operation=zoom.list_auto_dialer_call_history]; notes: Bounded sensitive Auto Dialer read. The provider artifact declares paging values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
  - auto-dialer reports call-history list - List Zoom Auto Dialer report call history. [intent=direct_read availability=implemented operation=zoom.list_auto_dialer_report_call_history]; notes: Bounded sensitive Auto Dialer report read. The provider artifact declares no request paging, token, date, or query input.; flags: --page, --page-cursor
  - auto-dialer reports seller-productivity get - Get the Zoom Auto Dialer seller-productivity report. [intent=direct_read availability=implemented operation=zoom.get_auto_dialer_seller_productivity_report]; notes: Bounded sensitive Auto Dialer report read. The provider artifact declares no request paging, token, date, or query input.; flags: --page, --page-cursor
  - auto-dialer call-lists list - List Zoom Auto Dialer call lists. [intent=direct_read availability=implemented operation=zoom.list_auto_dialer_call_lists]; notes: Bounded sensitive Auto Dialer call-list read. The provider artifact declares paging values only in the response schema, so no request query flag is exposed.; flags: --page, --page-cursor
  - auto-dialer call-lists get - Get one Zoom Auto Dialer call list. [intent=direct_read availability=implemented operation=zoom.get_auto_dialer_call_list]; notes: Bounded sensitive Auto Dialer call-list read at a fixed path. The provider artifact declares no request query input.; flags: --call-list-id (required), --page, --page-cursor
  - auto-dialer call-lists prospects list - List prospects in one Zoom Auto Dialer call list. [intent=direct_read availability=implemented operation=zoom.list_auto_dialer_call_list_prospects]; notes: Bounded sensitive Auto Dialer prospect read at a fixed call-list path. The provider artifact declares response-only paging values, so no request query flag is exposed.; flags: --call-list-id (required), --page, --page-cursor
  - auto-dialer prospects get - Get one Zoom Auto Dialer prospect. [intent=direct_read availability=implemented operation=zoom.get_auto_dialer_prospect]; notes: Bounded sensitive Auto Dialer prospect read at a fixed path. The provider artifact declares no request query input.; flags: --prospect-id (required), --page, --page-cursor
- Tasks
  - tasks assignees list - List assignees of one Zoom Task. [intent=direct_read availability=implemented operation=zoom.list_task_assignees]; notes: Bounded sensitive Task assignee read at a fixed path. The provider artifact declares no request query or paging input.; flags: --task-id (required), --page, --page-cursor
  - tasks collaborators list - List collaborators of one Zoom Task. [intent=direct_read availability=implemented operation=zoom.list_task_collaborators]; notes: Bounded sensitive Task collaborator read at a fixed path. The provider artifact declares no request query or paging input.; flags: --task-id (required), --page, --page-cursor
  - tasks comments list - List comments on one Zoom Task. [intent=direct_read availability=implemented operation=zoom.list_task_comments]; notes: Bounded sensitive Task comment read at a fixed path. The provider artifact declares response-only paging values, so no request query flag is exposed.; flags: --task-id (required), --page, --page-cursor
  - tasks imports get - Get one Zoom Tasks import job. [intent=direct_read availability=implemented operation=zoom.get_task_import]; notes: Bounded sensitive Task import-job read at a fixed path. The provider artifact declares no request query or paging input.; flags: --import-id (required), --page, --page-cursor
  - tasks items list - List Zoom Tasks items. [intent=direct_read availability=implemented operation=zoom.list_tasks]; notes: Bounded sensitive Task item read. The provider artifact declares response-only paging values but no request query, filter, sort, token, or date input, so no such flag is exposed.; flags: --page, --page-cursor
  - tasks items get - Get one Zoom Task. [intent=direct_read availability=implemented operation=zoom.get_task]; notes: Bounded sensitive Task detail read at a fixed path. The provider artifact declares no request query or paging input.; flags: --task-id (required), --page, --page-cursor
- Workforce Management
  - workforce-management filter-groups list - List Zoom Workforce Management filter groups. [intent=direct_read availability=implemented operation=zoom.list_workforce_filter_groups]; notes: Bounded sensitive Workforce Management filter-group read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
  - workforce-management forecasts list - List Zoom Workforce Management forecasts. [intent=direct_read availability=implemented operation=zoom.list_workforce_forecasts]; notes: Bounded sensitive Workforce Management forecast read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
  - workforce-management forecasts scheduling-groups get - Get a Workforce Management forecast for one scheduling group. [intent=direct_read availability=implemented operation=zoom.get_workforce_forecast_scheduling_group]; notes: Bounded sensitive forecast scheduling-group read at the provider's fixed path. The artifact declares no request query or paging input.; flags: --forecast-id (required), --scheduling-group-id (required), --page, --page-cursor
  - workforce-management imports historical-queue-metrics get - Get Workforce Management historical queue-metrics import metadata. [intent=direct_read availability=implemented operation=zoom.get_workforce_historical_queue_metrics_import]; notes: Bounded sensitive historical queue-metrics import-metadata read at the provider's fixed path. The artifact declares no request query or paging input.; flags: --import-id (required), --page, --page-cursor
  - workforce-management organizational-groups list - List Zoom Workforce Management organizational groups. [intent=direct_read availability=implemented operation=zoom.list_workforce_organizational_groups]; notes: Bounded sensitive Workforce Management organizational-group read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
  - workforce-management organizational-groups get - Get one Zoom Workforce Management organizational group. [intent=direct_read availability=implemented operation=zoom.get_workforce_organizational_group]; notes: Bounded sensitive organizational-group detail read at the provider's fixed path. The artifact declares no request query or paging input.; flags: --organizational-group-id (required), --page, --page-cursor
  - workforce-management reports adherence agents list - List Zoom Workforce Management adherence agents. [intent=direct_read availability=implemented operation=zoom.list_workforce_adherence_agents]; notes: Bounded sensitive Workforce Management adherence-agent report read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
  - workforce-management reports schedules agents list - List Zoom Workforce Management report schedule agents. [intent=direct_read availability=implemented operation=zoom.list_workforce_report_schedule_agents]; notes: Bounded sensitive Workforce Management report schedule-agent read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
  - workforce-management schedules agents list - List Zoom Workforce Management schedule agents. [intent=direct_read availability=implemented operation=zoom.list_workforce_schedule_agents]; notes: Bounded sensitive Workforce Management schedule-agent read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
  - workforce-management scheduling-groups list - List Zoom Workforce Management scheduling groups. [intent=direct_read availability=implemented operation=zoom.list_workforce_scheduling_groups]; notes: Bounded sensitive Workforce Management scheduling-group read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
  - workforce-management users list - List Zoom Workforce Management users. [intent=direct_read availability=implemented operation=zoom.list_workforce_users]; notes: Bounded sensitive Workforce Management user read. The provider artifact declares no request query or paging input.; flags: --page, --page-cursor
- Clips
  - clips list - List all Zoom Clips. [intent=direct_read availability=implemented operation=zoom.list_clips]; notes: Bounded sensitive Clips list read. The provider artifact documents optional user_id but exposes paging values only in the response, so no request paging flag is exposed.; flags: --user-id, --page, --page-cursor
  - clips collaborators list - List collaborators of one Zoom Clip. [intent=direct_read availability=implemented operation=zoom.list_clip_collaborators]; notes: Bounded sensitive Clip collaborator read. The provider artifact exposes paging values only in the response, so no request paging flag is exposed.; flags: --clip-id (required), --page, --page-cursor
  - clips comments list - List comments on one Zoom Clip. [intent=direct_read availability=implemented operation=zoom.list_clip_comments]; notes: Bounded sensitive Clip comment read. The provider artifact exposes paging values only in the response, so no request paging flag is exposed.; flags: --clip-id (required), --page, --page-cursor
  - clips get - Get one Zoom Clip's metadata and sharing state. [intent=direct_read availability=implemented operation=zoom.get_clip]; notes: Bounded sensitive Clip metadata read at the provider's fixed path.; flags: --clip-id (required), --page, --page-cursor
  - clips chapters get - Get the complete chapter list for one Zoom Clip. [intent=direct_read availability=implemented operation=zoom.get_clip_chapters]; notes: Bounded sensitive Clip chapter read at the provider's fixed path.; flags: --clip-id (required), --page, --page-cursor
  - clips transfers get - Get one Zoom Clip ownership-transfer task. [intent=direct_read availability=implemented operation=zoom.get_clip_transfer]; notes: Bounded sensitive asynchronous Clip transfer-task read at the provider's fixed path.; flags: --task-id (required), --page, --page-cursor
- Conference Room Connector (CRC)
  - crc managed-rooms account-setting get - Get the CRC managed-room account setting. [intent=direct_read availability=implemented operation=zoom.get_crc_managed_room_account_setting]; approval: Read-only; no write approval is required.; risk: Reads sensitive CRC account settings.; notes: The provider artifact exposes response-only paging values, so no paging flag is exposed.; flags: --page, --page-cursor
  - crc api-connectors list - List CRC API Connectors. [intent=direct_read availability=implemented operation=zoom.list_crc_api_connectors]; approval: Read-only; no write approval is required.; risk: Reads sensitive CRC API Connector details.; notes: The provider artifact exposes response-only paging values, so no paging flag is exposed.; flags: --page, --page-cursor
  - crc api-connectors get - Get one CRC API Connector. [intent=direct_read availability=implemented operation=zoom.get_crc_api_connector]; approval: Read-only; no write approval is required.; risk: Reads sensitive CRC API Connector details.; notes: The connector ID is a fixed provider path parameter and is redacted in output.; flags: --connector-id (required), --page, --page-cursor
  - crc api-connectors private-key get - Get one CRC API Connector private key. [intent=direct_read availability=implemented operation=zoom.get_crc_api_connector_private_key]; approval: Read-only; no write approval is required.; risk: Reads a private key only through redacted output.; notes: The declared json_redacted policy prevents the private key from reaching CLI output.; flags: --connector-id (required), --page, --page-cursor
  - crc managed-rooms list - List CRC managed rooms. [intent=direct_read availability=implemented operation=zoom.list_crc_managed_rooms]; approval: Read-only; no write approval is required.; risk: Reads sensitive CRC managed-room details.; notes: The provider artifact exposes response-only paging values, so no paging flag is exposed.; flags: --page, --page-cursor
  - crc managed-rooms get - Get one CRC managed room. [intent=direct_read availability=implemented operation=zoom.get_crc_managed_room]; approval: Read-only; no write approval is required.; risk: Reads sensitive CRC managed-room details.; notes: The device ID is a fixed provider path parameter and is redacted in output.; flags: --device-id (required), --page, --page-cursor
  - crc participant-identifier-code get - Get the CRC participant identifier code. [intent=direct_read availability=implemented operation=zoom.get_crc_participant_identifier_code]; approval: Read-only; no write approval is required.; risk: Reads a CRC meeting authentication identifier.; notes: The participant identifier code is redacted in output.; flags: --page, --page-cursor
  - crc room-templates list - List CRC room templates. [intent=direct_read availability=implemented operation=zoom.list_crc_room_templates]; approval: Read-only; no write approval is required.; risk: Reads sensitive CRC room-template details.; notes: The provider artifact exposes response-only paging values, so no paging flag is exposed.; flags: --page, --page-cursor
  - crc room-templates get - Get one CRC room template. [intent=direct_read availability=implemented operation=zoom.get_crc_room_template]; approval: Read-only; no write approval is required.; risk: Reads sensitive CRC room-template details.; notes: The template ID is a fixed provider path parameter and is redacted in output.; flags: --template-id (required), --page, --page-cursor
- Help topics:
  - provider-inventory - The Zoom provider ledger tracks 1,913 documented REST operations; Wave 1 executes only these three existing stream-backed reads.

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
