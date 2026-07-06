---
name: pm-tally-prime
description: TallyPrime connector knowledge and safe action guide.
---

# pm-tally-prime

## Purpose

Reads TallyPrime accounting data (companies, ledgers, groups, stock items, vouchers) via TDL Export/Collection envelope requests POSTed to a locally-running TallyPrime Gateway Server. Read-only source; schema is discovered dynamically since TallyPrime has no static REST resource surface.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: accounting

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
pm connectors inspect tally-prime
```

### Inspect as structured JSON

```bash
pm connectors inspect tally-prime --json
```

## Agent Rules

- Run pm connectors inspect tally-prime before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
