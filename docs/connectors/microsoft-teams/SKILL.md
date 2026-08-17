---
name: pm-microsoft-teams
description: Microsoft Teams connector knowledge and safe action guide.
---

# pm-microsoft-teams

## Purpose

Reads Microsoft Teams users, groups, channels, and device-usage reports through the Microsoft Graph REST API using an OAuth2 client-credentials grant. Read-only.

## Icon

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
- client_id
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
  - fields: account_enabled(), display_name(), id(), job_title(), mail(), mobile_phone(), office_location(), user_principal_name()
- groups:
  - primary key: id
  - fields: created_date_time(), description(), display_name(), id(), mail(), mail_enabled(), mail_nickname(), security_enabled(), visibility()
- channels:
  - primary key: id
  - fields: created_date_time(), description(), display_name(), email(), id(), membership_type(), web_url()
- team_device_usage_report:
  - primary key: id
  - fields: id(), is_deleted(), last_activity_date(), report_period(), used_android_phone(), used_i_os(), used_mac(), used_web(), used_windows_phone(), user_principal_name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
