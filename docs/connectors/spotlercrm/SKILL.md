---
name: pm-spotlercrm
description: Spotler CRM connector knowledge and safe action guide.
---

# pm-spotlercrm

## Purpose

Reads Spotler CRM contacts, accounts, opportunities, and tasks, and (via the real CRM API v4) activities, campaigns, and cases.

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
- access_token (secret)
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: id
  - fields: email(), firstName(), id(), lastName()
- accounts:
  - primary key: id
  - fields: id(), name(), status()
- opportunities:
  - primary key: id
  - fields: id(), name(), status()
- tasks:
  - primary key: id
  - fields: id(), name(), status()
- activities:
  - primary key: id
  - fields: createddate(), id(), modifieddate(), ownerid()
- campaigns:
  - primary key: id
  - fields: createddate(), id(), modifieddate(), name(), ownerid()
- cases:
  - primary key: id
  - fields: createddate(), id(), modifieddate(), ownerid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Spotler CRM API read of contact, account, opportunity, task, activity, campaign, and case data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect spotlercrm
```

### Inspect as structured JSON

```bash
pm connectors inspect spotlercrm --json
```

## Agent Rules

- Run pm connectors inspect spotlercrm before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
