---
name: pm-dynamodb
description: DynamoDB connector knowledge and safe action guide.
---

# pm-dynamodb

## Purpose

Reads DynamoDB table items through the AWS JSON HTTP API (DynamoDB_20120810.Scan), authenticated with hand-rolled AWS Signature Version 4 request signing. Read-only source; no write support.

## Icon

- asset: icons/dynamodb.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: database

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
pm connectors inspect dynamodb
```

### Inspect as structured JSON

```bash
pm connectors inspect dynamodb --json
```

## Agent Rules

- Run pm connectors inspect dynamodb before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
