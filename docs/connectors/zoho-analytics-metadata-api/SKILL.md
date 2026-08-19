---
name: pm-zoho-analytics-metadata-api
description: Zoho Analytics Metadata API connector knowledge and safe action guide.
---

# pm-zoho-analytics-metadata-api

## Purpose

Reads Zoho Analytics workspace/view/table/organization/folder/query-table/datasource metadata and triggers datasource/view data syncs, via the Zoho OAuth 2.0 refresh-token grant.

## Icon

- id: simple-icons-zoho-analytics-metadata-api
- asset: icons/simple-icons/zoho-analytics-metadata-api.svg
- title: Zoho
- simple_icon_slug: zoho
- simple_icon_hex: E42527
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Zoho
- match: curated-alias
- matched_by: zoho

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- org_id
- token_url
- workspace_id
- client_id (secret) (required)
- client_secret (secret) (required)
- refresh_token (secret) (required)

## ETL Streams

- workspaces:
  - primary key: id
  - fields: created_time(string), id(string), name(string)
- views:
  - primary key: id
  - fields: id(string), name(string)
- tables:
  - primary key: id
  - fields: id(string), name(string)
- organizations:
  - primary key: orgId
  - fields: createdBy(string), createdByZuId(string), isDefault(boolean), numberOfworkspaces(integer), orgDesc(string), orgId(string), orgName(string), planName(string), role(string)
- recent_views:
  - primary key: viewId
  - fields: viewId(string), viewLastAccessedTime(string), viewName(string), viewType(string), workspaceId(string), workspaceName(string)
- shared_workspaces:
  - primary key: workspaceId
  - fields: createdBy(string), createdTime(string), isDefault(boolean), orgId(string), workspaceDesc(string), workspaceId(string), workspaceName(string)
- shared_dashboards:
  - primary key: viewId
  - fields: createdBy(string), createdTime(string), folderId(string), isFavorite(boolean), lastModifiedBy(string), lastModifiedTime(string), orgId(string), parentViewId(string), sharedBy(string), viewDesc(string), viewId(string), viewName(string), viewType(string), workspaceId(string)
- folders:
  - primary key: folderId
  - fields: folderDesc(string), folderId(string), folderIndex(integer), folderName(string), isDefault(boolean), parentFolderId(string)
- query_tables:
  - primary key: viewId
  - fields: createdBy(string), createdTime(string), description(string), folderId(string), isFavorite(boolean), lastModifiedBy(string), lastModifiedTime(string), orgId(string), parentViewId(string), sharedBy(string), type(string), viewId(string), viewName(string), workspaceId(string)
- datasources:
  - primary key: datasourceId
  - fields: datasourceId(string), datasourceName(string), lastDataSyncStatus(string), lastDataSyncTime(string), nextScheduleTime(string), schedule(string), source(string), syncIntervalId(string), syncUsed(string), tableDetails(array), totalSyncAllowed(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- sync_datasource:
  - endpoint: POST /workspaces/{{ record.workspace_id }}/datasources/{{ record.datasource_id }}/sync
  - required fields: workspace_id, datasource_id
  - risk: triggers an asynchronous data sync for one datasource in a workspace; low-risk (re-fetches data from the connected source, does not itself mutate any Zoho Analytics record). The documented optional CONFIG query parameter, which can carry a datasource's own username/password credential for the sync, is NOT supported by this action (see docs.md Known limits) -- only the no-CONFIG invocation shown in Zoho's own sample request is modeled
- refetch_view_data:
  - endpoint: POST /workspaces/{{ record.workspace_id }}/views/{{ record.view_id }}/sync
  - required fields: workspace_id, view_id
  - risk: triggers an asynchronous data refetch for one view from its available datasource; low-risk (re-fetches, does not itself mutate any Zoho Analytics record). Same CONFIG-credential limitation as sync_datasource -- see docs.md Known limits

## Security

- read risk: external Zoho Analytics API read of workspace/view/table metadata plus organizations, recently-accessed views, shared workspaces/dashboards, folders, query tables, and datasources
- write risk: triggers an asynchronous datasource or view data sync/refetch in Zoho Analytics; does not create, modify, or delete any Zoho Analytics workspace/view/table/data record itself -- only re-pulls from an already-configured external datasource
- approval: required for write actions (sync_datasource/refetch_view_data); read access uses the same OAuth refresh-token grant with no separate approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoho-analytics-metadata-api
```

### Inspect as structured JSON

```bash
pm connectors inspect zoho-analytics-metadata-api --json
```

## Agent Rules

- Run pm connectors inspect zoho-analytics-metadata-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
