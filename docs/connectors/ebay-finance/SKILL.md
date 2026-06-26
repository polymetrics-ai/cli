---
name: pm-ebay-finance
description: eBay Finance connector knowledge and safe action guide.
---

# pm-ebay-finance

## Purpose

Reads eBay seller financial data — transactions, payouts, transfers, and the seller funds summary — through the eBay Sell Finances REST API.

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
pm connectors inspect ebay-finance
```

### Inspect as structured JSON

```bash
pm connectors inspect ebay-finance --json
```

## Agent Rules

- Run pm connectors inspect ebay-finance before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

