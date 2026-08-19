---
name: pm-zoom
description: Zoom connector knowledge and safe action guide.
---

# pm-zoom

## Purpose

Reads Zoom users, meetings, and webinars, and plans two source-backed approval-gated Zoom warehouse destination actions through the Zoom REST API.

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

- read risk: external Zoom API read of user, meeting, and webinar data
- approval: reverse ETL actions require plan, preview, explicit approval, and execute; all other provider writes remain disabled
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read existing Zoom streams and plan the two source-backed, approval-gated warehouse destination actions.
- Usage: pm zoom <users|meetings|webinars|healthcare|quality-management> <action> [flags]
- Source CLI: Zoom API reference (Pinned public source lock at sources/zoom-operation-source-lock.json (captured 2026-08-19; provider documents report OpenAPI 3.0.0); api_surface retains the 2026-08-05 ledger snapshot for inventory continuity.)
- Global flags:
  - --credential (string): Credential name to use for the Zoom request.
  - --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum records to emit from a stream command.
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Users
  - users list - Read Zoom users through the declared ETL stream. [intent=etl availability=implemented stream=users]
- Meetings and webinars
  - meetings list - Read meetings for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=meetings]; flags: --user-id
  - webinars list - Read webinars for one Zoom user through the declared ETL stream. [intent=etl availability=implemented stream=webinars]; flags: --user-id
- Approval-gated warehouse destination actions
  - healthcare clinical-notes update - Plan an update to a clinical note's completion status. [intent=reverse_etl availability=implemented write=update_clinical_note]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: high: changes a patient's clinical note completion status through an approval-gated reverse ETL action; notes: Typed high-risk mutation; preview and explicit approval are required before execute. The clinical note ID is redacted in write errors.; flags: --note-id (required), --is-note-completed (required)
  - quality-management interactions create - Plan creation of a Quality Management interaction from a third-party download URL. [intent=reverse_etl availability=implemented write=create_quality_management_interaction]; approval: reverse ETL plan -> preview -> explicit approval -> execute; risk: high: imports a third-party interaction into Zoom Quality Management through an approval-gated reverse ETL action; notes: Typed high-risk mutation. Download URL and interaction-info fields are redacted in generic write errors; preview and explicit approval are required before execute. If any interaction-info field is supplied, Zoom requires interaction-channel-type.; flags: --download-url (required), --direction, --disposition, --interaction-channel-type, --interaction-agent-email, --interaction-agent-id, --interaction-consumer-name, --interaction-from, --interaction-to, --primary-language, --queue-id, --start-time
- Help topics:
  - provider-inventory - The Zoom source lock and disposition ledger account for all 1,913 tracked REST operations. This connector exposes three preserved ETL streams and two fixture-proven, approval-gated destination actions; the remaining provider contracts are explicitly disabled pending their recorded evidence and foundations.

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
