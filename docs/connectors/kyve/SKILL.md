---
name: pm-kyve
description: KYVE connector knowledge and safe action guide.
---

# pm-kyve

## Purpose

Reads public KYVE pools, stakers, funders, and Cosmos validators through the KYVE network's public REST query endpoints. Read-only; no credentials required.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- max_pages
- mode
- page_size

## ETL Streams

- pools:
  - primary key: id
  - fields: id(), name(), runtime()
- stakers:
  - primary key: address
  - fields: address(), amount()
- funders:
  - primary key: address
  - fields: address(), amount()
- validators:
  - primary key: operator_address
  - fields: moniker(), operator_address(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external read of public KYVE network pool/staker/funder/validator data
- approval: none; read-only public Cosmos-style REST API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect kyve
```

### Inspect as structured JSON

```bash
pm connectors inspect kyve --json
```

## Agent Rules

- Run pm connectors inspect kyve before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
