---
name: pm-google-search-console
description: Google Search Console connector knowledge and safe action guide.
---

# pm-google-search-console

## Purpose

Reads Google Search Console sites, sitemaps, and Search Analytics performance reports (by date, query, page, country, and device) through the Search Console v3 REST API. Read-only.

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
pm connectors inspect google-search-console
```

### Inspect as structured JSON

```bash
pm connectors inspect google-search-console --json
```

## Agent Rules

- Run pm connectors inspect google-search-console before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

