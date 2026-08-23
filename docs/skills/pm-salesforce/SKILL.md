---
name: pm-salesforce
description: Salesforce connector knowledge and safe action guide.
---

# pm-salesforce

## Purpose

Reads Salesforce object metadata and allow-listed Account, Contact, and Lead SOQL queries through the REST API. Read-only.

## Icon

- id: salesforce
- asset: icons/salesforce.svg
- source: official
- review_status: official_verified
- review_url: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_version
- instance_url (required)
- mode
- access_token (secret) (required)

## ETL Streams

- sobjects:
  - primary key: qualified_api_name
  - fields: label(string), qualified_api_name(string)
- accounts:
  - primary key: id
  - fields: email(string), id(string), last_modified_date(string), name(string)
- contacts:
  - primary key: id
  - fields: email(string), id(string), last_modified_date(string), name(string)
- leads:
  - primary key: id
  - fields: email(string), id(string), last_modified_date(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Salesforce API read of object metadata, Account, Contact, and Lead records
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect salesforce
```

### Inspect as structured JSON

```bash
pm connectors inspect salesforce --json
```

## Agent Rules

- Run pm connectors inspect salesforce before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
