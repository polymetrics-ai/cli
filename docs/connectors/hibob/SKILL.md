---
name: pm-hibob
description: HiBob connector knowledge and safe action guide.
---

# pm-hibob

## Purpose

Reads HiBob HR data: employee profiles, company named lists, and people field definitions via the HiBob REST API (read-only).

## Icon

- asset: icons/hibob.svg
- source: official
- review_status: official_verified
- review_url: https://apidocs.hibob.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- username
- password (secret)

## ETL Streams

- profiles:
  - primary key: id
  - fields: displayName(), email(), firstName(), fullName(), id(), personal_pronouns(), surname(), work_department(), work_isManager(), work_site(), work_startDate(), work_title()
- named_lists:
  - primary key: id
  - fields: archived(), children(), id(), name(), parentId(), value()
- company_lists:
  - primary key: id
  - fields: category(), description(), id(), name(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external HiBob API read of employee profile and HR metadata
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect hibob
```

### Inspect as structured JSON

```bash
pm connectors inspect hibob --json
```

## Agent Rules

- Run pm connectors inspect hibob before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
