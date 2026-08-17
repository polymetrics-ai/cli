---
name: pm-amazon-sqs
description: Amazon SQS connector knowledge and safe action guide.
---

# pm-amazon-sqs

## Purpose

Reads messages from Amazon SQS via signed ReceiveMessage calls. Read-only; messages are not deleted.

## Icon

- asset: icons/amazon-sqs.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: queue

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
pm connectors inspect amazon-sqs
```

### Inspect as structured JSON

```bash
pm connectors inspect amazon-sqs --json
```

## Agent Rules

- Run pm connectors inspect amazon-sqs before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
