---
name: pm-outbox
description: Local Outbox connector knowledge and safe action guide.
---

# pm-outbox

## Purpose

Local JSONL destination that records reverse ETL writes and receipts.

## Icon

- asset: icons/pm-outbox.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/karthik-sivadas/polymetrics-cli

## Capabilities

- check=true catalog=true read=false write=true query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- path: Local outbox directory.

## ETL Streams

- records: Reverse ETL outbox records.

## Security

- read risk: unsupported
- write risk: local file write
- mutation risk: reverse ETL receipt writes
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect outbox
```

### Inspect as structured JSON

```bash
pm connectors inspect outbox --json
```

### Outbox reverse ETL

```bash
pm credentials add outbox-local --connector outbox --config path=$ROOT/.polymetrics/outbox
pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map email:email
pm reverse run <plan-id> --approve <approval-token> --json
```

## Agent Rules

- Run pm connectors inspect outbox before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
