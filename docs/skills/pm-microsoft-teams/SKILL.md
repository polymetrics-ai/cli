---
name: pm-microsoft-teams
description: Microsoft Teams connector knowledge and safe action guide.
---

# pm-microsoft-teams

## Purpose

Reads Microsoft Teams users, groups, channels, and device-usage reports through the Microsoft Graph REST API using an OAuth2 client-credentials grant. Read-only.

## Icon

- id: microsoft-teams
- asset: icons/microsoft-teams.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- client_id (required)
- login_base_url
- max_pages
- period
- scope
- token_url
- client_secret (secret)
- tenant_id (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: account_enabled(boolean), display_name(string), id(string), job_title(string), mail(string), mobile_phone(string), office_location(string), user_principal_name(string)
- groups:
  - primary key: id
  - fields: created_date_time(string), description(string), display_name(string), id(string), mail(string), mail_enabled(boolean), mail_nickname(string), security_enabled(boolean), visibility(string)
- channels:
  - primary key: id
  - fields: created_date_time(string), description(string), display_name(string), email(string), id(string), membership_type(string), web_url(string)
- team_device_usage_report:
  - primary key: id
  - fields: id(string), is_deleted(boolean), last_activity_date(string), report_period(string), used_android_phone(boolean), used_i_os(boolean), used_mac(boolean), used_web(boolean), used_windows_phone(boolean), user_principal_name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Microsoft Graph API read of tenant users/groups/channels/device-usage data
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect microsoft-teams
```

### Inspect as structured JSON

```bash
pm connectors inspect microsoft-teams --json
```

## Agent Rules

- Run pm connectors inspect microsoft-teams before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
