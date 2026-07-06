---
name: pm-mailchimp
description: Mailchimp connector knowledge and safe action guide.
---

# pm-mailchimp

## Purpose

Reads Mailchimp Marketing API audiences (lists), campaigns, reports, and automations through the datacenter-scoped REST API.

## Icon

- asset: icons/mailchimp.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://mailchimp.com/developer/release-notes/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- data_center
- mode
- start_date
- access_token (secret)
- api_key (secret)

## ETL Streams

- lists:
  - primary key: id
  - cursor: date_created
  - fields: date_created(), email_type_option(), id(), list_rating(), name(), subscribe_url_short(), visibility(), web_id()
- campaigns:
  - primary key: id
  - cursor: create_time
  - fields: archive_url(), create_time(), emails_sent(), id(), send_time(), status(), type(), web_id()
- reports:
  - primary key: id
  - cursor: send_time
  - fields: abuse_reports(), campaign_title(), emails_sent(), id(), list_id(), send_time(), type(), unsubscribed()
- automations:
  - primary key: id
  - cursor: create_time
  - fields: create_time(), emails_sent(), id(), start_time(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Mailchimp API read of audience, campaign, report, and automation data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mailchimp
```

### Inspect as structured JSON

```bash
pm connectors inspect mailchimp --json
```

## Agent Rules

- Run pm connectors inspect mailchimp before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
