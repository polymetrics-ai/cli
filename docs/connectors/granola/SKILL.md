---
name: pm-granola
description: Granola connector knowledge and safe action guide.
---

# pm-granola

## Purpose

Reads Granola meeting notes metadata and full note detail (summary, owner, attendees, calendar event) through the Granola public API (read-only).

## Icon

- id: source-granola
- asset: icons/source-granola.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.granola.ai/introduction

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- start_date
- api_key (secret)

## ETL Streams

- notes:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), object(string), owner_email(string), owner_name(string), title(string), updated_at(string)
- detailed_notes:
  - primary key: id
  - cursor: created_at
  - fields: attendees(array), calendar_event(object), created_at(string), folders(array), id(string), object(string), owner_email(string), owner_name(string), summary(string), title(string), transcript(array), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Granola API read of meeting notes metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Granola's declared streams and reverse-ETL actions.
- Usage: pm granola <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete v1 webhook-endpoints - Documented DELETE /v1/webhook-endpoints (not implemented) [intent=direct_write availability=not_implemented operation=granola.delete.v1-webhook-endpoints]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 webhook-endpoints webhook-endpoint-id - Documented DELETE /v1/webhook-endpoints/{webhook_endpoint_id} (not implemented) [intent=direct_write availability=not_implemented operation=granola.delete.v1-webhook-endpoints-webhook-endpoint-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get v1 folders - Documented GET /v1/folders (not implemented) [intent=direct_read availability=not_implemented operation=granola.get.v1-folders]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 notes note-id - Documented GET /v1/notes/{note_id} (not implemented) [intent=direct_read availability=not_implemented operation=granola.get.v1-notes-note-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 webhook-endpoints - Documented GET /v1/webhook-endpoints (not implemented) [intent=direct_read availability=not_implemented operation=granola.get.v1-webhook-endpoints]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch v1 webhook-endpoints webhook-endpoint-id - Documented PATCH /v1/webhook-endpoints/{webhook_endpoint_id} (not implemented) [intent=direct_write availability=not_implemented operation=granola.patch.v1-webhook-endpoints-webhook-endpoint-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 webhook-endpoints - Documented POST /v1/webhook-endpoints (not implemented) [intent=direct_write availability=not_implemented operation=granola.post.v1-webhook-endpoints]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - detailed notes list - Run the detailed notes ETL stream [intent=etl availability=implemented stream=detailed_notes]; notes: discrepancy=present-in-surface-absent-from-artifact
  - notes list - Run the notes ETL stream [intent=etl availability=implemented stream=notes]

## Commands

### Inspect as a manual

```bash
pm connectors inspect granola
```

### Inspect as structured JSON

```bash
pm connectors inspect granola --json
```

## Agent Rules

- Run pm connectors inspect granola before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
