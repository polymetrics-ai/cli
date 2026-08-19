---
name: pm-mode
description: Mode connector knowledge and safe action guide.
---

# pm-mode

## Purpose

Reads Mode collections (spaces), reports, data sources, groups, and memberships through the Mode REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

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
pm connectors inspect mode
```

### Inspect as structured JSON

```bash
pm connectors inspect mode --json
```

## Agent Rules

- Run pm connectors inspect mode before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
