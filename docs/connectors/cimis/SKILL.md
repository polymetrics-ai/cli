---
name: pm-cimis
description: CIMIS connector knowledge and safe action guide.
---

# pm-cimis

## Purpose

Reads California Irrigation Management Information System (CIMIS) daily and hourly weather/ET observations and station metadata through the CIMIS Web API. Read-only.

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
pm connectors inspect cimis
```

### Inspect as structured JSON

```bash
pm connectors inspect cimis --json
```

## Agent Rules

- Run pm connectors inspect cimis before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

