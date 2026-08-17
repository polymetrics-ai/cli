---
name: pm-zoho-bigin
description: Zoho Bigin connector knowledge and safe action guide.
---

# pm-zoho-bigin

## Purpose

Reads and writes Zoho Bigin pipelines, contacts, companies, products, tasks, events, calls, notes, users, tags, module metadata, and generic module records via the Zoho OAuth 2.0 refresh-token grant.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- pipelines:
  - primary key: id
  - fields: display_value(), id(), name()
- records:
  - primary key: id
  - fields: id(), name()
- fields:
  - primary key: id
  - fields: api_name(), display_label(), id()
- contacts:
  - primary key: id
  - fields: Account_Name(), Created_Time(), Email(), First_Name(), Last_Name(), Mobile(), Modified_Time(), Owner(), Phone(), Title(), display_value(), id()
- companies:
  - primary key: id
  - fields: Account_Name(), Created_Time(), Modified_Time(), Owner(), Phone(), Website(), display_value(), id()
- products:
  - primary key: id
  - fields: Created_Time(), Modified_Time(), Owner(), Product_Code(), Product_Name(), Unit_Price(), display_value(), id()
- tasks:
  - primary key: id
  - fields: Created_Time(), Due_Date(), Modified_Time(), Owner(), Priority(), Status(), Subject(), Who_Id(), id()
- events:
  - primary key: id
  - fields: Created_Time(), End_DateTime(), Event_Title(), Location(), Modified_Time(), Owner(), Start_DateTime(), Who_Id(), id()
- calls:
  - primary key: id
  - fields: Call_Duration(), Call_Start_Time(), Call_Type(), Created_Time(), Modified_Time(), Owner(), Subject(), Who_Id(), id()
- notes:
  - primary key: id
  - fields: Created_Time(), Modified_Time(), Note_Content(), Note_Title(), Owner(), Parent_Id(), id()
- users:
  - primary key: id
  - fields: email(), first_name(), full_name(), id(), last_name(), profile(), role(), status(), time_zone()
- tags:
  - primary key: id
  - fields: created_time(), id(), modified_time(), name()
- modules:
  - primary key: id
  - fields: api_name(), api_supported(), creatable(), deletable(), editable(), id(), module_name(), plural_label(), singular_label(), viewable()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_record:
  - endpoint: POST /{{ config.module_name }}
  - risk: creates one or more new records in config.module_name; external mutation, approval required
- update_record:
  - endpoint: PUT /{{ config.module_name }}
  - risk: overwrites the named fields of one or more existing records in config.module_name; external mutation, approval required
- upsert_record:
  - endpoint: POST /{{ config.module_name }}/upsert
  - risk: inserts a new record in config.module_name if no match is found on duplicate_check_fields, otherwise overwrites the matched existing record's submitted fields; external mutation, approval required
- delete_record:
  - endpoint: DELETE /{{ config.module_name }}/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a single record from config.module_name; external mutation, approval required
- create_note:
  - endpoint: POST /{{ config.module_name }}/{{ record.parent_id }}/Notes
  - required fields: parent_id
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
