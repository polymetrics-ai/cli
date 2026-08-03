---
name: pm-aws-cloudtrail
description: AWS CloudTrail connector knowledge and safe action guide.
---

# pm-aws-cloudtrail

## Purpose

Reads and safely operates AWS CloudTrail trails, event data stores, channels, dashboards, and Lake queries through fixed AWS JSON-RPC streams, typed direct-read commands, and typed reverse-ETL write/admin actions. Only StartQuery, CreateDashboard, and UpdateDashboard stay blocked: each requires an unrestricted CloudTrail Lake SQL QueryStatement, which this project disables for every connector by policy. Breaking change: the earlier LookupEvents-backed management_events, read_only_events, write_only_events, and console_logins streams are removed and have no replacement here; use the events lookup direct-read command for CloudTrail event lookups instead. See the connector docs migration note.

## Icon

- asset: icons/aws-cloudtrail.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=true query=false
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
pm connectors inspect aws-cloudtrail
```

### Inspect as structured JSON

```bash
pm connectors inspect aws-cloudtrail --json
```

## Agent Rules

- Run pm connectors inspect aws-cloudtrail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
