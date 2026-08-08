# pm connectors inspect zoom

```text
NAME
  pm connectors inspect zoom - Zoom connector manual

SYNOPSIS
  pm connectors inspect zoom
  pm connectors inspect zoom --json
  pm credentials add <name> --connector zoom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoom users, meetings, webinars, and bounded module-specific data through the Zoom REST API; includes sensitive Cobrowse SDK session reads and approval-gated clinical-note and Quality Management interaction actions.

ICON
  id: zoom
  asset: icons/zoom.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.zoom.us/docs/api/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  user_id
  access_token (secret)

ETL STREAMS
  users:
    primary key: id
    fields: email(string), id(string), name(string), updated_at(string)
  meetings:
    primary key: id
    fields: email(string), id(string), name(string), updated_at(string)
  webinars:
    primary key: id
    fields: email(string), id(string), name(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_clinical_note:
    endpoint: PATCH /clinical_notes/notes/{{ record.note_id }}
    required fields: note_id, is_note_completed
    risk: high: mutates a patient's clinical note completion status; requires reverse ETL approval
  create_quality_management_interaction:
    endpoint: POST /qm/interactions
    required fields: download_url
    risk: high: imports a third-party interaction into Zoom Quality Management; requires reverse ETL approval

SECURITY
  read risk: external Zoom API read of user, meeting, webinar, Quality of Service, AI Companion, My Notes, healthcare clinical-note, Quality Management, and Cobrowse SDK session data
  write risk: typed Zoom reverse ETL mutation of a healthcare clinical-note completion status or Quality Management interaction creation
  approval: reverse ETL writes require plan, preview, explicit approval, and execute; read-only commands require none
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run declared Zoom stream reads, bounded module-specific direct reads including sensitive Cobrowse SDK sessions, and approval-gated clinical-note and Quality Management interaction actions.
  Usage: pm zoom <group> <command> [flags]
  Source CLI: Zoom API reference (OpenAPI 3.1.1; docs static build 2026-08-03T14-58-19-06-00; retrieved 2026-08-05)
  Global flags:
    --credential (string): Credential name to use for the Zoom request.
    --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
    --config (string_array): Connector config override as key=value; never pass secret values here.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum records to emit from a stream command.
  Users
    users list - Read Zoom users through the declared ETL stream. [intent=etl availability=implemented stream=users]
  Meetings and webinars
    meetings list - Read meetings for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=meetings]; flags: --user-id
    webinars list - Read webinars for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=webinars]; flags: --user-id
  Quality of Service Summaries
    qss meeting-participants list - List a past meeting's participants' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_meeting_participants_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required meeting-id path parameter.; flags: --meeting-id (required)
    qss webinar-participants list - List a live or past webinar's participants' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_webinar_participants_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required webinar-id path parameter.; flags: --webinar-id (required)
    qss session-users list - List a past Video SDK session's users' Quality of Service summary. [intent=direct_read availability=implemented operation=zoom.list_session_users_qos_summary]; notes: Bounded Zoom read; fixed method and path with a typed required session-id path parameter.; flags: --session-id (required)
  AI Companion
    ai-companion conversation-archive get - Get a user's AI Companion conversation archive. [intent=direct_read availability=implemented operation=zoom.get_ai_companion_conversation_archives]; notes: Bounded Zoom read; fixed method and path with a typed required user-id path parameter. Response download-URL fields are redacted by the json_redacted output policy.; flags: --user-id (required)
  My Notes
    my-notes list - List the authenticated user's My Notes. [intent=direct_read availability=implemented operation=zoom.list_my_notes]; notes: Bounded Zoom read; fixed method and path with no request parameters (the live artifact documents none for this operation).
    my-notes content get - Get a My Notes note's content, and optionally its meeting transcript. [intent=direct_read availability=implemented operation=zoom.get_my_notes_content]; notes: Bounded Zoom read; fixed method and path with a typed required note-id path parameter and an optional include=transcript query flag (explicitly documented in provider prose, unlike qss's response-only pagination fields).; flags: --note-id (required), --include
  Healthcare
    healthcare clinical-notes list - List clinical notes, optionally filtered by owner and/or meeting. [intent=direct_read availability=implemented operation=zoom.list_clinical_notes]; notes: Bounded sensitive Zoom read. Only note-owner-user-id and meeting-id are explicit provider request inputs; response-only dates and paging fields are not CLI flags. The clinical_json_redacted policy redacts clinical-note content and identifiers.; flags: --note-owner-user-id, --meeting-id
    healthcare clinical-notes get - Get a single clinical note by ID. [intent=direct_read availability=implemented operation=zoom.get_clinical_note]; notes: Bounded sensitive Zoom read with a typed required note-id path parameter. The clinical_json_redacted policy redacts clinical-note content and identifiers.; flags: --note-id (required)
    healthcare clinical-notes update - Plan an update to a clinical note's completion status. [intent=reverse_etl availability=implemented write=update_clinical_note]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: high: changes a patient's clinical note completion status through an approval-gated reverse ETL action; notes: Typed high-risk mutation; preview and explicit approval are required before execute. The clinical note ID is redacted in write errors.; flags: --note-id (required), --is-note-completed (required)
  Quality Management
    quality-management automated-evaluations list - List Quality Management automated evaluations. [intent=direct_read availability=implemented operation=zoom.list_quality_management_automated_evaluations]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination fields are not CLI flags; identifiers and personal-contact fields are redacted before output.
    quality-management evaluations list - List completed Quality Management evaluations. [intent=direct_read availability=implemented operation=zoom.list_quality_management_evaluations]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination fields are not CLI flags; identifiers and personal-contact fields are redacted before output.
    quality-management evaluations get - View one Quality Management evaluation. [intent=direct_read availability=implemented operation=zoom.get_quality_management_evaluation]; notes: Bounded sensitive Zoom read with a typed required evaluation-id path parameter; identifiers and personal-contact fields are redacted before output.; flags: --evaluation-id (required)
    quality-management interactions list - List Quality Management interactions. [intent=direct_read availability=implemented operation=zoom.list_quality_management_interactions]; notes: Bounded sensitive Zoom read with no provider-declared request parameters. Response-only pagination/date fields are not CLI flags; identifiers and personal-contact fields are redacted before output.
    quality-management interactions get - View one Quality Management interaction. [intent=direct_read availability=implemented operation=zoom.get_quality_management_interaction]; notes: Bounded sensitive Zoom read with a typed required interaction-id path parameter; identifiers and personal-contact fields are redacted before output.; flags: --interaction-id (required)
    quality-management interactions create - Plan creation of a Quality Management interaction from a third-party download URL. [intent=reverse_etl availability=implemented write=create_quality_management_interaction]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: high: imports a third-party interaction into Zoom Quality Management through an approval-gated reverse ETL action; notes: Typed high-risk mutation. Download URL and interaction-info fields are redacted in generic write errors; preview and explicit approval are required before execute. If any interaction-info field is supplied, Zoom requires interaction-channel-type.; flags: --download-url (required), --direction, --disposition, --interaction-channel-type, --interaction-agent-email, --interaction-agent-id, --interaction-consumer-name, --interaction-from, --interaction-to, --primary-language, --queue-id, --start-time
  Cobrowse SDK
    cobrowse-sdk live-sessions list - List live Cobrowse SDK sessions for an optional monthly date range. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_live_sessions]; notes: Bounded sensitive Zoom read. The provider explicitly permits an optional monthly from/to date range; page_size and next_page_token are response-only fields and are not CLI flags.; flags: --from, --to
    cobrowse-sdk past-sessions list - List past Cobrowse SDK sessions for an optional monthly date range. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_past_sessions]; notes: Bounded sensitive Zoom read. The provider explicitly permits an optional monthly from/to date range; page_size and next_page_token are response-only fields and are not CLI flags.; flags: --from, --to
    cobrowse-sdk sessions get - Get details for one Cobrowse SDK session. [intent=direct_read availability=implemented operation=zoom.get_cobrowse_session]; notes: Bounded sensitive Zoom read with a typed required session-id path parameter; session pins, user/session identifiers, display names, connection IDs, and IP addresses are redacted before output.; flags: --session-id (required)
    cobrowse-sdk sessions users list - List users from one Cobrowse SDK session. [intent=direct_read availability=implemented operation=zoom.list_cobrowse_session_users]; notes: Bounded sensitive Zoom read with a typed required session-id path parameter; page_size and next_page_token are response-only fields and session/user connection data is redacted before output.; flags: --session-id (required)
  Help topics:
    provider-inventory - The Zoom provider ledger tracks 1,913 documented REST operations; Wave 1 executes three stream-backed reads; Wave 2+ adds bounded direct-read/write operations module by module (see #3915).

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoom

  # Inspect as structured JSON
  pm connectors inspect zoom --json

AGENT WORKFLOW
  - Run pm connectors inspect zoom before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
