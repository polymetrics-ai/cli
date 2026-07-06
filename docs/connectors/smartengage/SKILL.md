---
name: pm-smartengage
description: SmartEngage connector knowledge and safe action guide.
---

# pm-smartengage

## Purpose

Reads SmartEngage avatars, tags, custom fields, sequences, and subscribers; creates/updates subscribers, tags, custom fields, and sequence enrollments.

## Icon

- asset: icons/smartengage.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://smartengage.com/docs/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- avatar_id
- base_url
- api_key (secret)

## ETL Streams

- avatars:
  - primary key: id
  - fields: avatar_id(), id(), name()
- tags:
  - primary key: id
  - fields: avatar_id(), id(), name()
- custom_fields:
  - primary key: id
  - fields: avatar_id(), id(), name()
- sequences:
  - primary key: id
  - fields: avatar_id(), id(), name()
- subscribers:
  - primary key: id
  - fields: avatar_id(), id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_subscriber:
  - endpoint: POST subscribers/add
  - risk: external mutation; creates a new subscriber on the connected SmartEngage account; approval required
- update_subscriber:
  - endpoint: POST subscribers/update
  - risk: external mutation; overwrites subscriber fields on the connected SmartEngage account (fields omitted from the record remain unchanged); approval required
- create_tag:
  - endpoint: POST tags/create
  - risk: external mutation; creates a new tag on the connected SmartEngage account; approval required
- add_tag_to_subscriber:
  - endpoint: POST tags/add
  - risk: external mutation; attaches an existing tag to a subscriber; approval required
- remove_tag_from_subscriber:
  - endpoint: POST tags/delete
  - risk: external mutation; detaches a tag from a subscriber; approval required
- create_custom_field:
  - endpoint: POST customfields/create
  - risk: external mutation; creates a new custom field definition on the connected SmartEngage account; approval required
- set_custom_field_value:
  - endpoint: POST customfields/update
  - risk: external mutation; sets a custom field value on a subscriber; approval required
- add_subscriber_to_sequence:
  - endpoint: POST sequences/add
  - risk: external mutation; enrolls a subscriber into an automation sequence, triggering scheduled messages; approval required
- remove_subscriber_from_sequence:
  - endpoint: POST sequences/remove
  - risk: external mutation; unenrolls a subscriber from an automation sequence, stopping scheduled messages; approval required

## Security

- read risk: read-only avatar/tag/custom-field/sequence/subscriber data from a connected SmartEngage account
- write risk: creates/updates subscribers and custom-field values, creates tags and attaches/detaches them from subscribers, and enrolls/unenrolls subscribers in automation sequences (which triggers or stops scheduled outbound messages)
- approval: required for all 9 write actions; read is unapproved
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect smartengage
```

### Inspect as structured JSON

```bash
pm connectors inspect smartengage --json
```

## Agent Rules

- Run pm connectors inspect smartengage before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
