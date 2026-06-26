---
name: pm-google-tasks
description: Google Tasks connector knowledge and safe action guide.
---

# pm-google-tasks

## Purpose

Reads Google task lists and tasks through the Google Tasks REST API.

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
pm connectors inspect google-tasks
```

### Inspect as structured JSON

```bash
pm connectors inspect google-tasks --json
```

## Agent Rules

- Run pm connectors inspect google-tasks before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

