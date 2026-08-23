---
name: pm-yahoo-finance-price
description: Yahoo Finance Price connector knowledge and safe action guide.
---

# pm-yahoo-finance-price

## Purpose

Reads public Yahoo Finance chart prices and flattens them into OHLCV records. Read-only.

## Icon

- id: yahoo-finance-price
- asset: icons/yahoo-finance-price.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.yahoofinanceapi.com/

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
pm connectors inspect yahoo-finance-price
```

### Inspect as structured JSON

```bash
pm connectors inspect yahoo-finance-price --json
```

## Agent Rules

- Run pm connectors inspect yahoo-finance-price before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
