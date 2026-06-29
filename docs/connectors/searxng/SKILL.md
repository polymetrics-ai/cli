---
name: pm-searxng
description: SearXNG connector knowledge and safe action guide.
---

# pm-searxng

## Purpose

Reads web and Reddit search results from a SearXNG metasearch instance's JSON API (format=json). Read-only. Requires base_url; no credentials by default.

## Icon

- asset: icons/searxng.svg
- source: official_site
- review_status: manual_override
- review_url: https://docs.searxng.org/

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
pm connectors inspect searxng
```

### Inspect as structured JSON

```bash
pm connectors inspect searxng --json
```

## Agent Rules

- Run pm connectors inspect searxng before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
