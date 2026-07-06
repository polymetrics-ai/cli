---
name: pm-free-agent-connector
description: FreeAgent connector knowledge and safe action guide.
---

# pm-free-agent-connector

## Purpose

Reads FreeAgent contacts, invoices, bills, projects, and tasks through the FreeAgent v2 REST API using OAuth2 refresh-token authentication. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- payroll_year
- updated_since
- client_id (secret)
- client_refresh_token_2 (secret)
- client_secret (secret)

## ETL Streams

- contacts:
  - primary key: url
  - cursor: updated_at
  - fields: account_balance(), created_at(), email(), first_name(), last_name(), organisation_name(), phone_number(), status(), updated_at(), url()
- invoices:
  - primary key: url
  - cursor: updated_at
  - fields: contact(), created_at(), currency(), dated_on(), due_on(), due_value(), net_value(), reference(), status(), total_value(), updated_at(), url()
- bills:
  - primary key: url
  - cursor: updated_at
  - fields: contact(), created_at(), currency(), dated_on(), due_on(), due_value(), reference(), status(), total_value(), updated_at(), url()
- projects:
  - primary key: url
  - cursor: updated_at
  - fields: budget(), budget_units(), contact(), created_at(), currency(), name(), status(), updated_at(), url()
- tasks:
  - primary key: url
  - cursor: updated_at
  - fields: billing_period(), billing_rate(), created_at(), is_billable(), name(), project(), status(), updated_at(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external FreeAgent API reads performed by the legacy connector via a Tier-2 hook
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
