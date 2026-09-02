---
name: pm-copper
description: Copper connector knowledge and safe action guide.
---

# pm-copper

## Purpose

Reads Copper CRM people, companies, opportunities, leads, and tasks through the Copper REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: copper
- asset: icons/copper.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.copper.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- user_email (required)
- api_key (secret) (required)

## ETL Streams

- people:
  - primary key: id
  - cursor: date_modified
  - fields: assignee_id(integer), company_id(integer), company_name(string), contact_type_id(integer), date_created(integer), date_modified(integer), emails(array), first_name(string), id(integer), last_name(string), name(string), phone_numbers(array), prefix(string), title(string)
- companies:
  - primary key: id
  - cursor: date_modified
  - fields: address(object), assignee_id(integer), date_created(integer), date_modified(integer), details(string), email_domain(string), id(integer), name(string), phone_numbers(array), websites(array)
- opportunities:
  - primary key: id
  - cursor: date_modified
  - fields: assignee_id(integer), close_date(string), company_id(integer), company_name(string), date_created(integer), date_modified(integer), id(integer), monetary_value(number), name(string), pipeline_id(integer), pipeline_stage_id(integer), primary_contact_id(integer), status(string), win_probability(number)
- leads:
  - primary key: id
  - cursor: date_modified
  - fields: assignee_id(integer), company_name(string), date_created(integer), date_modified(integer), email(object), id(integer), monetary_value(number), name(string), phone_numbers(array), status(string), status_id(integer), title(string)
- tasks:
  - primary key: id
  - cursor: date_modified
  - fields: assignee_id(integer), completed_date(integer), date_created(integer), date_modified(integer), details(string), due_date(integer), id(integer), name(string), priority(string), related_resource(object), reminder_date(integer), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Copper API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect copper
```

### Inspect as structured JSON

```bash
pm connectors inspect copper --json
```

## Agent Rules

- Run pm connectors inspect copper before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
