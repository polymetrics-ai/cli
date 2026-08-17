---
name: pm-pinterest
description: Pinterest connector knowledge and safe action guide.
---

# pm-pinterest

## Purpose

Reads Pinterest ad accounts, boards, campaigns, ad groups, and audiences through the Pinterest API v5 (OAuth2 refresh-token auth). Read-only.

## Icon

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
- client_id (secret)
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- ad_accounts:
  - primary key: id
  - fields: country(), currency(), id(), name(), owner()
- boards:
  - primary key: id
  - fields: created_at(), description(), follower_count(), id(), name(), owner(), pin_count(), privacy()
- campaigns:
  - primary key: id
  - fields: ad_account_id(), created_time(), id(), name(), objective_type(), status(), updated_time()
- ad_groups:
  - primary key: id
  - fields: ad_account_id(), campaign_id(), created_time(), id(), name(), status(), updated_time()
- audiences:
  - primary key: id
  - fields: ad_account_id(), audience_type(), created_timestamp(), id(), name(), size(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
