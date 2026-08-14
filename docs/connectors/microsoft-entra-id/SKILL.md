---
name: pm-microsoft-entra-id
description: Microsoft Entra ID connector knowledge and safe action guide.
---

# pm-microsoft-entra-id

## Purpose

Reads Microsoft Entra ID (Azure AD) directory objects — users, groups, applications, service principals, and directory roles — from the Microsoft Graph API using an OAuth2 client-credentials grant. Read-only.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- login_base_url
- max_pages
- mode
- page_size
- scope
- token_url
- client_id (secret)
- client_secret (secret)
- tenant_id (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: account_enabled(boolean), department(string), display_name(string), given_name(string), id(string), job_title(string), mail(string), mobile_phone(string), office_location(string), surname(string), user_principal_name(string)
- groups:
  - primary key: id
  - fields: created_date_time(string), description(string), display_name(string), id(string), mail(string), mail_enabled(boolean), mail_nickname(string), security_enabled(boolean), visibility(string)
- applications:
  - primary key: id
  - fields: app_id(string), created_date_time(string), description(string), display_name(string), id(string), publisher_domain(string), sign_in_audience(string)
- serviceprincipals:
  - primary key: id
  - fields: account_enabled(boolean), app_id(string), app_owner_organization_id(string), display_name(string), id(string), service_principal_type(string), sign_in_audience(string)
- directoryroles:
  - primary key: id
  - fields: description(string), display_name(string), id(string), role_template_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Microsoft Graph API read of tenant directory (users/groups/applications/service principals/directory roles) data
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect microsoft-entra-id
```

### Inspect as structured JSON

```bash
pm connectors inspect microsoft-entra-id --json
```

## Agent Rules

- Run pm connectors inspect microsoft-entra-id before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
