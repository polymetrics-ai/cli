---
name: pm-zoho-campaign
description: Zoho Campaign connector knowledge and safe action guide.
---

# pm-zoho-campaign

## Purpose

Reads Zoho Campaigns lists, campaigns, and contacts through the Zoho Campaigns REST API.

## Icon

- id: simple-icons-zoho-campaign
- asset: icons/simple-icons/zoho-campaign.svg
- title: Zoho
- simple_icon_slug: zoho
- simple_icon_hex: E42527
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Zoho
- match: curated-alias
- matched_by: zoho

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
- access_token (secret) (required)

## ETL Streams

- lists:
  - primary key: id
  - cursor: updated_at
  - fields: createdtime(string), id(string), list_name(string), listkey(string), listname(string), listtype(string), modified_time(string), name(string), tag(string), updated_at(string)
- campaigns:
  - primary key: id
  - cursor: updated_at
  - fields: campaign_key(string), campaign_name(string), campaignkey(string), campaignname(string), from_email(string), id(string), modified_time(string), name(string), sent_time(string), status(string), subject(string), updated_at(string)
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: contact_id(string), contact_key(string), email(string), first_name(string), id(string), last_name(string), modified_time(string), name(string), status(string), updated_at(string)

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
