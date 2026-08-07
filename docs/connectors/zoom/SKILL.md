---
name: pm-zoom
description: Zoom connector knowledge and safe action guide.
---

# pm-zoom

## Purpose

Reads Zoom users, meetings, and webinars through the Zoom REST API, plus bounded Quality of Service summary reads.

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
- access_token (secret)

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

## Security

- read risk: external Zoom API read of user, meeting, and webinar data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read the currently stream-backed Zoom users, meetings, and webinars API routes, plus bounded Quality of Service summary reads.
- Usage: pm zoom <users|meetings|webinars|qss> <list> [flags]
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
- Help topics:
  - provider-inventory - The Zoom provider ledger tracks 1,913 documented REST operations; Wave 1 executes three stream-backed reads and Wave 2 (qss) adds three bounded direct-read operations.

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
