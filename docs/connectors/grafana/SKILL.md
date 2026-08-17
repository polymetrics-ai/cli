---
name: pm-grafana
description: Grafana connector knowledge and safe action guide.
---

# pm-grafana

## Purpose

Reads Grafana dashboards, folders, data sources, organization users, and provisioned alert rules through the Grafana REST API (read-only).

## Icon

- asset: icons/grafana.svg
- source: official
- review_status: official_verified
- review_url: https://grafana.com/docs/grafana/latest/developer-resources/api-reference/http-api/api-legacy/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- dashboards:
  - primary key: uid
  - fields: folderId(), folderTitle(), folderUid(), id(), isStarred(), orgId(), tags(), title(), type(), uid(), url()
- folders:
  - primary key: uid
  - fields: id(), orgId(), tags(), title(), type(), uid(), url()
- datasources:
  - primary key: uid
  - fields: access(), id(), isDefault(), name(), orgId(), readOnly(), type(), uid(), url()
- org_users:
  - primary key: userId
  - fields: email(), lastSeenAt(), login(), orgId(), role(), userId()
- alert_rules:
  - primary key: uid
  - fields: condition(), execErrState(), folderUID(), for(), id(), noDataState(), orgID(), ruleGroup(), title(), uid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Grafana instance API read of dashboards, folders, data sources, org users, and alert rules
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect grafana
```

### Inspect as structured JSON

```bash
pm connectors inspect grafana --json
```

## Agent Rules

- Run pm connectors inspect grafana before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
