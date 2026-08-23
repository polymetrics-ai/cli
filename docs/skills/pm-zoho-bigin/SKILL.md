---
name: pm-zoho-bigin
description: Zoho Bigin connector knowledge and safe action guide.
---

# pm-zoho-bigin

## Purpose

Reads and writes Zoho Bigin pipelines, contacts, companies, products, tasks, events, calls, notes, users, tags, module metadata, and generic module records via the Zoho OAuth 2.0 refresh-token grant.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- module_name
- token_url
- client_id (secret) (required)
- client_refresh_token (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- pipelines:
  - primary key: id
  - fields: display_value(string), id(string), name(string)
- records:
  - primary key: id
  - fields: id(string), name(string)
- fields:
  - primary key: id
  - fields: api_name(string), display_label(string), id(string)
- contacts:
  - primary key: id
  - fields: Account_Name(object), Created_Time(string), Email(string), First_Name(string), Last_Name(string), Mobile(string), Modified_Time(string), Owner(object), Phone(string), Title(string), display_value(string), id(string)
- companies:
  - primary key: id
  - fields: Account_Name(string), Created_Time(string), Modified_Time(string), Owner(object), Phone(string), Website(string), display_value(string), id(string)
- products:
  - primary key: id
  - fields: Created_Time(string), Modified_Time(string), Owner(object), Product_Code(string), Product_Name(string), Unit_Price(number), display_value(string), id(string)
- tasks:
  - primary key: id
  - fields: Created_Time(string), Due_Date(string), Modified_Time(string), Owner(object), Priority(string), Status(string), Subject(string), Who_Id(object), id(string)
- events:
  - primary key: id
  - fields: Created_Time(string), End_DateTime(string), Event_Title(string), Location(string), Modified_Time(string), Owner(object), Start_DateTime(string), Who_Id(object), id(string)
- calls:
  - primary key: id
  - fields: Call_Duration(string), Call_Start_Time(string), Call_Type(string), Created_Time(string), Modified_Time(string), Owner(object), Subject(string), Who_Id(object), id(string)
- notes:
  - primary key: id
  - fields: Created_Time(string), Modified_Time(string), Note_Content(string), Note_Title(string), Owner(object), Parent_Id(object), id(string)
- users:
  - primary key: id
  - fields: email(string), first_name(string), full_name(string), id(string), last_name(string), profile(object), role(object), status(string), time_zone(string)
- tags:
  - primary key: id
  - fields: created_time(string), id(string), modified_time(string), name(string)
- modules:
  - primary key: id
  - fields: api_name(string), api_supported(boolean), creatable(boolean), deletable(boolean), editable(boolean), id(string), module_name(string), plural_label(string), singular_label(string), viewable(boolean)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_record:
  - endpoint: POST /{{ config.module_name }}
  - required fields: data
  - risk: creates one or more new records in config.module_name; external mutation, approval required
- update_record:
  - endpoint: PUT /{{ config.module_name }}
  - required fields: data
  - risk: overwrites the named fields of one or more existing records in config.module_name; external mutation, approval required
- upsert_record:
  - endpoint: POST /{{ config.module_name }}/upsert
  - required fields: data
  - risk: inserts a new record in config.module_name if no match is found on duplicate_check_fields, otherwise overwrites the matched existing record's submitted fields; external mutation, approval required
- delete_record:
  - endpoint: DELETE /{{ config.module_name }}/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a single record from config.module_name; external mutation, approval required
- create_note:
  - endpoint: POST /{{ config.module_name }}/{{ record.parent_id }}/Notes
  - required fields: parent_id, data
  - risk: attaches one or more notes to an existing record in config.module_name; low-risk external mutation, no approval required
- delete_note:
  - endpoint: DELETE /{{ config.module_name }}/{{ record.parent_id }}/Notes/{{ record.id }}
  - required fields: parent_id, id
  - risk: permanently deletes a single note from a record in config.module_name; external mutation, approval required

## Security

- read risk: external Zoho Bigin API read of pipeline, contact, company, product, task, event, call, note, user, tag, module-metadata, and generic module-record data
- write risk: external mutation of Zoho Bigin CRM records (create/update/upsert/delete on the configured module, plus note create/delete); moves real business data, approval required
- approval: required for all write actions; reads require no approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoho-bigin
```

### Inspect as structured JSON

```bash
pm connectors inspect zoho-bigin --json
```

## Agent Rules

- Run pm connectors inspect zoho-bigin before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
