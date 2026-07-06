---
name: pm-zoho-analytics-metadata-api
description: Zoho Analytics Metadata API connector knowledge and safe action guide.
---

# pm-zoho-analytics-metadata-api

## Purpose

Reads Zoho Analytics workspace/view/table/organization/folder/query-table/datasource metadata and triggers datasource/view data syncs, via the Zoho OAuth 2.0 refresh-token grant.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
- client_id (secret)
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- workspaces:
  - primary key: id
  - fields: created_time(), id(), name()
- views:
  - primary key: id
  - fields: id(), name()
- tables:
  - primary key: id
  - fields: id(), name()
- organizations:
  - primary key: orgId
  - fields: createdBy(), createdByZuId(), isDefault(), numberOfworkspaces(), orgDesc(), orgId(), orgName(), planName(), role()
- recent_views:
  - primary key: viewId
  - fields: viewId(), viewLastAccessedTime(), viewName(), viewType(), workspaceId(), workspaceName()
- shared_workspaces:
  - primary key: workspaceId
  - fields: createdBy(), createdTime(), isDefault(), orgId(), workspaceDesc(), workspaceId(), workspaceName()
- shared_dashboards:
  - primary key: viewId
  - fields: createdBy(), createdTime(), folderId(), isFavorite(), lastModifiedBy(), lastModifiedTime(), orgId(), parentViewId(), sharedBy(), viewDesc(), viewId(), viewName(), viewType(), workspaceId()
- folders:
  - primary key: folderId
  - fields: folderDesc(), folderId(), folderIndex(), folderName(), isDefault(), parentFolderId()
- query_tables:
  - primary key: viewId
  - fields: createdBy(), createdTime(), description(), folderId(), isFavorite(), lastModifiedBy(), lastModifiedTime(), orgId(), parentViewId(), sharedBy(), type(), viewId(), viewName(), workspaceId()
- datasources:
  - primary key: datasourceId
  - fields: datasourceId(), datasourceName(), lastDataSyncStatus(), lastDataSyncTime(), nextScheduleTime(), schedule(), source(), syncIntervalId(), syncUsed(), tableDetails(), totalSyncAllowed()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
