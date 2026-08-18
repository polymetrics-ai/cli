---
name: pm-grafana
description: Grafana connector knowledge and safe action guide.
---

# pm-grafana

## Purpose

Reads Grafana dashboards, folders, data sources, organization users, and provisioned alert rules through the Grafana REST API (read-only).

## Icon

- id: grafana
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

- base_url (required)
- mode
- api_key (secret) (required)

## ETL Streams

- dashboards:
  - primary key: uid
  - fields: folderId(integer), folderTitle(string), folderUid(string), id(integer), isStarred(boolean), orgId(integer), tags(array), title(string), type(string), uid(string), url(string)
- folders:
  - primary key: uid
  - fields: id(integer), orgId(integer), tags(array), title(string), type(string), uid(string), url(string)
- datasources:
  - primary key: uid
  - fields: access(string), id(integer), isDefault(boolean), name(string), orgId(integer), readOnly(boolean), type(string), uid(string), url(string)
- org_users:
  - primary key: userId
  - fields: email(string), lastSeenAt(string), login(string), orgId(integer), role(string), userId(integer)
- alert_rules:
  - primary key: uid
  - fields: condition(string), execErrState(string), folderUID(string), for(string), id(integer), noDataState(string), orgID(integer), ruleGroup(string), title(string), uid(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
