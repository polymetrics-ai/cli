---
name: pm-mention
description: Mention connector knowledge and safe action guide.
---

# pm-mention

## Purpose

Reads Mention app metadata, accounts, alerts, mentions, alert tags, alert shares, alert preferences, and alert tasks from the Mention social listening REST API.

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

- account_id
- alert_id
- base_url
- mode
- api_key (secret)

## ETL Streams

- app_data:
  - fields: actions(), alert_languages(), countries(), days(), folders(), integrations(), languages(), sources(), tones()
- account_me:
  - primary key: id
  - fields: created_at(), id(), language(), name(), permission(), timezone()
- account:
  - primary key: id
  - fields: created_at(), id(), language(), name(), permission(), timezone()
- alert:
  - primary key: id
  - fields: countries(), created_at(), description(), id(), languages(), name(), query(), sources(), updated_at()
- mention:
  - primary key: id
  - fields: created_at(), description(), favorite(), id(), language(), published_at(), source_name(), source_type(), title(), tone(), url()
- alert_tag:
  - primary key: id
  - fields: color(), id(), name()
- alert_share:
  - primary key: id
  - fields: created_at(), email(), id(), permission(), updated_at()
- alert_preferences:
  - fields: frequency(), notification_frequency(), send_email(), send_push(), shared()
- alert_task:
  - primary key: id
  - fields: created_at(), description(), id(), mention(), state(), title(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mention API read of app metadata, account, alert, mention, tag, share, preference, and task data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mention
```

### Inspect as structured JSON

```bash
pm connectors inspect mention --json
```

## Agent Rules

- Run pm connectors inspect mention before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
