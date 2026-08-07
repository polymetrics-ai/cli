# pm connectors inspect zoom

```text
NAME
  pm connectors inspect zoom - Zoom connector manual

SYNOPSIS
  pm connectors inspect zoom
  pm connectors inspect zoom --json
  pm credentials add <name> --connector zoom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoom users, meetings, and webinars through the Zoom REST API, plus bounded Quality of Service summary reads.

ICON
  id: zoom
  asset: icons/zoom.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.zoom.us/docs/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
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

SECURITY
  read risk: external Zoom API read of user, meeting, and webinar data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read the currently stream-backed Zoom users, meetings, and webinars API routes, plus bounded Quality of Service summary reads.
  Usage: pm zoom <users|meetings|webinars|qss> <list> [flags]
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

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
