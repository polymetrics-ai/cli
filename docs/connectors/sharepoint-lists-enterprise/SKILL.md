---
name: pm-sharepoint-lists-enterprise
description: SharePoint Lists Enterprise connector knowledge and safe action guide.
---

# pm-sharepoint-lists-enterprise

## Purpose

Reads and writes SharePoint lists and list items through Microsoft Graph.

## Icon

- asset: icons/microsoft-sharepoint.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- list_id
- login_base_url
- mode
- site_id
- tenant_id
- token_url
- client_id (secret)
- client_secret (secret)

## ETL Streams

- lists:
  - primary key: id
  - cursor: lastModifiedDateTime
  - fields: displayName(), id(), lastModifiedDateTime(), name()
- list_items:
  - primary key: id
  - cursor: lastModifiedDateTime
  - fields: fields(), id(), lastModifiedDateTime()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_list:
  - endpoint: POST /sites/{{ config.site_id }}/lists
  - risk: creates a new SharePoint list (and any custom columns/template declared in the request) on the configured site; low-risk external mutation, no approval required
- update_list:
  - endpoint: PATCH /sites/{{ config.site_id }}/lists/{{ record.id }}
  - required fields: id
  - risk: mutates an existing list's display name/description by id; low-risk external mutation, no approval required
- create_list_item:
  - endpoint: POST /sites/{{ config.site_id }}/lists/{{ config.list_id }}/items
  - risk: creates a new item (row) in the configured list, with the submitted column values; low-risk external mutation, no approval required
- update_list_item:
  - endpoint: PATCH /sites/{{ config.site_id }}/lists/{{ config.list_id }}/items/{{ record.id }}/fields
  - required fields: id
  - risk: mutates an existing list item's column values by id, via the fields sub-resource (Graph's fieldValueSet update); only the submitted column names are changed, matching Graph's own partial-update semantics for this endpoint

## Security

- read risk: external Microsoft Graph API read of SharePoint list and list-item data
- write risk: creates/updates SharePoint lists and list items (rows and their column values) on the configured site via Microsoft Graph
- approval: none for list/list-item create-update (low-risk CRM/CMS-style data, no destructive deletes implemented)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sharepoint-lists-enterprise
```

### Inspect as structured JSON

```bash
pm connectors inspect sharepoint-lists-enterprise --json
```

## Agent Rules

- Run pm connectors inspect sharepoint-lists-enterprise before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
