---
name: pm-flexmail
description: Flexmail connector knowledge and safe action guide.
---

# pm-flexmail

## Purpose

Reads Flexmail contacts, custom fields, interests, segments, and sources through the Flexmail REST API.

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

- account_id (required)
- base_url
- mode
- page_size
- personal_access_token (secret) (required)

## ETL Streams

- contacts:
  - primary key: id
  - fields: custom_fields(object), email(string), first_name(string), id(integer), language(string), name(string)
- custom_fields:
  - primary key: id
  - fields: id(string), name(string), placeholder(string), type(string)
- interests:
  - primary key: id
  - fields: description(string), id(string), label(string), name(string), visibility(string)
- segments:
  - primary key: id
  - fields: id(string), name(string), number_of_contacts(integer)
- sources:
  - primary key: id
  - fields: id(integer), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Flexmail API read of contact and marketing-list data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Flexmail's declared streams and reverse-ETL actions.
- Usage: pm flexmail <command> [flags]
- Read streams
- Other Commands
  - contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
  - custom fields list - Run the custom fields ETL stream [intent=etl availability=implemented stream=custom_fields]
  - interests list - Run the interests ETL stream [intent=etl availability=implemented stream=interests]
  - segments list - Run the segments ETL stream [intent=etl availability=implemented stream=segments]
  - sources list - Run the sources ETL stream [intent=etl availability=implemented stream=sources]

## Commands

### Inspect as a manual

```bash
pm connectors inspect flexmail
```

### Inspect as structured JSON

```bash
pm connectors inspect flexmail --json
```

## Agent Rules

- Run pm connectors inspect flexmail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
