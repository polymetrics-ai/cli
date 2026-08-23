---
name: pm-mendeley
description: Mendeley connector knowledge and safe action guide.
---

# pm-mendeley

## Purpose

Reads documents, folders, groups, and annotations from the Mendeley reference manager REST API.

## Icon

- id: simple-icons-mendeley
- asset: icons/simple-icons/mendeley.svg
- title: Mendeley
- simple_icon_slug: mendeley
- simple_icon_hex: 9D1620
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Mendeley
- match: exact-name-or-slug
- matched_by: mendeley

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- No connector-specific config fields.

## Security

- read risk: connector-specific
- write risk: connector-specific
- approval: external mutations require preview and approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mendeley
```

### Inspect as structured JSON

```bash
pm connectors inspect mendeley --json
```

## Agent Rules

- Run pm connectors inspect mendeley before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
