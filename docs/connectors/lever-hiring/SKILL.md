---
name: pm-lever-hiring
description: Lever Hiring connector knowledge and safe action guide.
---

# pm-lever-hiring

## Purpose

Reads Lever Hiring opportunities, postings, users, requisitions, and stages through the Lever Data API. Read-only (full-refresh).

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
pm connectors inspect lever-hiring
```

### Inspect as structured JSON

```bash
pm connectors inspect lever-hiring --json
```

## Agent Rules

- Run pm connectors inspect lever-hiring before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

