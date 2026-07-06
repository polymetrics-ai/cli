---
name: pm-zoho-campaign
description: Zoho Campaign connector knowledge and safe action guide.
---

# pm-zoho-campaign

## Purpose

Reads Zoho Campaigns lists, campaigns, and contacts through the Zoho Campaigns REST API.

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
- max_pages
- mode
- page_size
- access_token (secret)

## ETL Streams

- lists:
  - primary key: id
  - cursor: updated_at
  - fields: createdtime(), id(), list_name(), listkey(), listname(), listtype(), modified_time(), name(), tag(), updated_at()
- campaigns:
  - primary key: id
  - cursor: updated_at
  - fields: campaign_key(), campaign_name(), campaignkey(), campaignname(), from_email(), id(), modified_time(), name(), sent_time(), status(), subject(), updated_at()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: contact_id(), contact_key(), email(), first_name(), id(), last_name(), modified_time(), name(), status(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zoho Campaigns API read of email marketing data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoho-campaign
```

### Inspect as structured JSON

```bash
pm connectors inspect zoho-campaign --json
```

## Agent Rules

- Run pm connectors inspect zoho-campaign before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
