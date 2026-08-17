---
name: pm-microsoft-entra-id
description: Microsoft Entra ID connector knowledge and safe action guide.
---

# pm-microsoft-entra-id

## Purpose

Reads Microsoft Entra ID (Azure AD) directory objects — users, groups, applications, service principals, and directory roles — from the Microsoft Graph API using an OAuth2 client-credentials grant. Read-only.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
  - fields: account_enabled(), department(), display_name(), given_name(), id(), job_title(), mail(), mobile_phone(), office_location(), surname(), user_principal_name()
- groups:
  - primary key: id
  - fields: created_date_time(), description(), display_name(), id(), mail(), mail_enabled(), mail_nickname(), security_enabled(), visibility()
- applications:
  - primary key: id
  - fields: app_id(), created_date_time(), description(), display_name(), id(), publisher_domain(), sign_in_audience()
- serviceprincipals:
  - primary key: id
  - fields: account_enabled(), app_id(), app_owner_organization_id(), display_name(), id(), service_principal_type(), sign_in_audience()
- directoryroles:
  - primary key: id
  - fields: description(), display_name(), id(), role_template_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
