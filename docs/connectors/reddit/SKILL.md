---
name: pm-reddit
description: Reddit connector knowledge and safe action guide.
---

# pm-reddit

## Purpose

Reads subreddit posts and comments through the Reddit OAuth API listing endpoints.

## Icon

- id: simple-icons-reddit
- asset: icons/simple-icons/reddit.svg
- title: Reddit
- simple_icon_slug: reddit
- simple_icon_hex: FF4500
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Reddit
- match: exact-name-or-slug
- matched_by: reddit

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- subreddit (required)
- access_token (secret) (required)

## ETL Streams

- posts:
  - primary key: id
  - cursor: created_utc
  - fields: author(string), created_utc(number), id(string), name(string), permalink(string), subreddit(string), title(string)
- comments:
  - primary key: id
  - cursor: created_utc
  - fields: author(string), body(string), created_utc(number), id(string), name(string), permalink(string), subreddit(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Reddit OAuth API read of public subreddit posts and comments
- approval: none; read-only, caller-supplied OAuth token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect reddit
```

### Inspect as structured JSON

```bash
pm connectors inspect reddit --json
```

## Agent Rules

- Run pm connectors inspect reddit before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
