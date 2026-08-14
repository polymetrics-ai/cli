---
name: pm-jamf-pro
description: Jamf Pro connector knowledge and safe action guide.
---

# pm-jamf-pro

## Purpose

Reads Jamf Pro buildings, departments, categories, and scripts through the Jamf Pro REST API using Basic-credential token-exchange authentication.

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

- base_url (required)
- max_pages
- mode
- page_size
- username (required)
- password (secret) (required)

## ETL Streams

- buildings:
  - primary key: id
  - fields: city(string), country(string), id(string), name(string), stateProvince(string), streetAddress1(string), streetAddress2(string), zipPostalCode(string)
- departments:
  - primary key: id
  - fields: id(string), name(string)
- categories:
  - primary key: id
  - fields: id(string), name(string), priority(integer)
- scripts:
  - primary key: id
  - fields: categoryId(string), categoryName(string), id(string), info(string), name(string), notes(string), osRequirements(string), priority(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Jamf Pro API read of MDM configuration data (buildings, departments, categories, scripts)
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect jamf-pro
```

### Inspect as structured JSON

```bash
pm connectors inspect jamf-pro --json
```

## Agent Rules

- Run pm connectors inspect jamf-pro before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
