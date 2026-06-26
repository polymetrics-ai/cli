---
name: pm-finnhub
description: Finnhub connector knowledge and safe action guide.
---

# pm-finnhub

## Purpose

Reads Finnhub stock symbols, company and market news, company profiles, and analyst recommendations through the Finnhub REST API.

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
pm connectors inspect finnhub
```

### Inspect as structured JSON

```bash
pm connectors inspect finnhub --json
```

## Agent Rules

- Run pm connectors inspect finnhub before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

