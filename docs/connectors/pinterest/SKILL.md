---
name: pm-pinterest
description: Pinterest connector knowledge and safe action guide.
---

# pm-pinterest

## Purpose

Reads Pinterest ad accounts, boards, campaigns, ad groups, and audiences through the Pinterest API v5 (OAuth2 refresh-token auth). Read-only.

## Icon

- id: pinterest
- asset: icons/pinterest.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.pinterest.com/docs/changelog/changelog/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- mode
- page_size
- token_url
- client_id (secret) (required)
- client_secret (secret) (required)
- refresh_token (secret) (required)

## ETL Streams

- ad_accounts:
  - primary key: id
  - fields: country(string), currency(string), id(string), name(string), owner(object)
- boards:
  - primary key: id
  - fields: created_at(string), description(string), follower_count(integer), id(string), name(string), owner(object), pin_count(integer), privacy(string)
- campaigns:
  - primary key: id
  - fields: ad_account_id(string), created_time(integer), id(string), name(string), objective_type(string), status(string), updated_time(integer)
- ad_groups:
  - primary key: id
  - fields: ad_account_id(string), campaign_id(string), created_time(integer), id(string), name(string), status(string), updated_time(integer)
- audiences:
  - primary key: id
  - fields: ad_account_id(string), audience_type(string), created_timestamp(integer), id(string), name(string), size(integer), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Pinterest API read of ad account, board, campaign, ad group, and audience data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pinterest
```

### Inspect as structured JSON

```bash
pm connectors inspect pinterest --json
```

## Agent Rules

- Run pm connectors inspect pinterest before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
