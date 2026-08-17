---
name: pm-wufoo
description: Wufoo connector knowledge and safe action guide.
---

# pm-wufoo

## Purpose

Reads Wufoo forms, fields, entries, comments, reports, and widgets, and writes entry submissions and webhook registrations through the Wufoo API.

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
- form_hash
- max_pages
- mode
- page_size
- report_hash
- api_key (secret)

## ETL Streams

- forms:
  - primary key: Hash
  - cursor: DateUpdated
  - fields: DateUpdated(), Hash(), Name()
- form_fields:
  - primary key: ID
  - fields: ClassNames(), ID(), Instructions(), IsRequired(), Title(), Type()
- entries:
  - primary key: Hash
  - cursor: DateUpdated
  - fields: DateCreated(), DateUpdated(), EntryId(), Hash()
- form_comments:
  - primary key: CommentId
  - cursor: DateCreated
  - fields: CommentId(), CommentedBy(), DateCreated(), EntryId(), Text()
- reports:
  - primary key: Hash
  - cursor: DateUpdated
  - fields: DateUpdated(), Hash(), Name()
- report_fields:
  - primary key: ID
  - fields: ClassNames(), ID(), Instructions(), Title(), Type()
- report_entries:
  - primary key: EntryId
  - cursor: DateUpdated
  - fields: DateCreated(), DateUpdated(), EntryId()
- report_widgets:
  - primary key: Hash
  - fields: Hash(), Name(), Size(), Type(), TypeDesc()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- submit_entry:
  - endpoint: POST /forms/{{ config.form_hash }}/entries.json
  - risk: external mutation; creates a live Wufoo form entry; approval required
- add_webhook:
  - endpoint: PUT /forms/{{ config.form_hash }}/webhooks.json
  - risk: external mutation; registers a webhook callback URL on the configured form; approval required
- delete_webhook:
  - endpoint: DELETE /forms/{{ config.form_hash }}/webhooks/{{ record.hash }}.json
  - required fields: hash
  - risk: irreversible external deletion; removes a registered webhook from the configured form; approval required

## Security

- read risk: external Wufoo API read of form, field, entry, comment, report, and widget data
- write risk: external mutation: submits live form entries and registers/removes webhook callback URLs
- approval: required for all write actions (submit_entry, add_webhook, delete_webhook); reads require none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect wufoo
```

### Inspect as structured JSON

```bash
pm connectors inspect wufoo --json
```

## Agent Rules

- Run pm connectors inspect wufoo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
