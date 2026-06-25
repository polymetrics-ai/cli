---
name: pm-apple-search-ads
description: Apple Ads connector knowledge and safe action guide.
---

# pm-apple-search-ads

## Purpose

Reads Apple Search Ads campaigns, ad groups, targeting keywords, and ads via the Apple Search Ads Campaign Management API using an OAuth2 client-credentials grant scoped to an organization. Read-only.

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
pm connectors inspect apple-search-ads
```

### Inspect as structured JSON

```bash
pm connectors inspect apple-search-ads --json
```

## Agent Rules

- Run pm connectors inspect apple-search-ads before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

