---
name: pm-microsoft-dataverse
description: Microsoft Dataverse connector knowledge and safe action guide.
---

# pm-microsoft-dataverse

## Purpose

Reads Microsoft Dataverse accounts, contacts, leads, opportunities, and users through the Web API.

## Icon

- id: microsoftdataverse
- asset: icons/microsoftdataverse.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/overview

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- login_base_url
- max_pages
- mode
- page_size
- scope (required)
- token_url
- client_id (secret)
- client_secret (secret)
- tenant_id (secret)

## ETL Streams

- accounts:
  - primary key: id
  - fields: created_on(string), email(string), id(string), modified_on(string), name(string)
- contacts:
  - primary key: id
  - fields: created_on(string), email(string), id(string), modified_on(string), name(string)
- leads:
  - primary key: id
  - fields: created_on(string), email(string), id(string), modified_on(string), name(string)
- opportunities:
  - primary key: id
  - fields: created_on(string), email(string), id(string), modified_on(string), name(string)
- systemusers:
  - primary key: id
  - fields: created_on(string), email(string), id(string), modified_on(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Microsoft Dataverse Web API read of CRM records
- approval: none; read-only OAuth2 client-credentials API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect microsoft-dataverse
```

### Inspect as structured JSON

```bash
pm connectors inspect microsoft-dataverse --json
```

## Agent Rules

- Run pm connectors inspect microsoft-dataverse before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
