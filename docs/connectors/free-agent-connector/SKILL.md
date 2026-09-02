---
name: pm-free-agent-connector
description: FreeAgent connector knowledge and safe action guide.
---

# pm-free-agent-connector

## Purpose

Reads FreeAgent contacts, invoices, bills, projects, and tasks through fixed FreeAgent v2 REST routes and OAuth2 refresh-token authentication.

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

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- updated_since
- client_id (secret) (required)
- client_refresh_token_2 (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- contacts:
  - primary key: url
  - cursor: updated_at
  - fields: account_balance(string), created_at(string), email(string), first_name(string), last_name(string), organisation_name(string), phone_number(string), status(string), updated_at(string), url(string)
- invoices:
  - primary key: url
  - cursor: updated_at
  - fields: contact(string), created_at(string), currency(string), dated_on(string), due_on(string), due_value(string), net_value(string), reference(string), status(string), total_value(string), updated_at(string), url(string)
- bills:
  - primary key: url
  - cursor: updated_at
  - fields: contact(string), created_at(string), currency(string), dated_on(string), due_on(string), due_value(string), reference(string), status(string), total_value(string), updated_at(string), url(string)
- projects:
  - primary key: url
  - cursor: updated_at
  - fields: budget(string), budget_units(string), contact(string), created_at(string), currency(string), name(string), status(string), updated_at(string), url(string)
- tasks:
  - primary key: url
  - cursor: updated_at
  - fields: billing_period(string), billing_rate(string), created_at(string), is_billable(boolean), name(string), project(string), status(string), updated_at(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded FreeAgent v2 reads use declared OAuth2 refresh-token authentication and fixed provider routes.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect free-agent-connector
```

### Inspect as structured JSON

```bash
pm connectors inspect free-agent-connector --json
```

## Agent Rules

- Run pm connectors inspect free-agent-connector before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
