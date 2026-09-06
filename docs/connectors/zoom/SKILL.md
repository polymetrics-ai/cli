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

- Read the currently stream-backed Zoom users, meetings, and webinars API routes.
- Usage: pm zoom <users|meetings|webinars> list [flags]
- Global flags:
  - --credential (string): Credential name to use for the Zoom request.
  - --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum records to emit from a stream command.
- Users
  - users list - Read Zoom users through the declared ETL stream. [intent=etl availability=implemented stream=users]
- Meetings and webinars
  - meetings list - Read meetings for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=meetings]; flags: --user-id (non-empty)
  - webinars list - Read webinars for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=webinars]; flags: --user-id (non-empty)
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
