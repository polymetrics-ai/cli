---
name: pm-smartengage
description: SmartEngage connector knowledge and safe action guide.
---

# pm-smartengage

## Purpose

Reads SmartEngage avatars, tags, custom fields, sequences, and subscribers; creates/updates subscribers, tags, custom fields, and sequence enrollments.

## Icon

- id: smartengage
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
- api_key (secret) (required)

## ETL Streams

- avatars:
  - primary key: id
  - fields: avatar_id(string), id(string), name(string)
- tags:
  - primary key: id
  - fields: avatar_id(string), id(string), name(string)
- custom_fields:
  - primary key: id
  - fields: avatar_id(string), id(string), name(string)
- sequences:
  - primary key: id
  - fields: avatar_id(string), id(string), name(string)
- subscribers:
  - primary key: id
  - fields: avatar_id(string), id(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- add_subscriber:
  - endpoint: POST subscribers/add
  - required fields: avatar_id
  - risk: external mutation; creates a new subscriber on the connected SmartEngage account; approval required
- update_subscriber:
  - endpoint: POST subscribers/update
  - required fields: avatar_id, subscriber_id
  - risk: external mutation; overwrites subscriber fields on the connected SmartEngage account (fields omitted from the record remain unchanged); approval required
- create_tag:
  - endpoint: POST tags/create
  - required fields: avatar_id, name
  - risk: external mutation; creates a new tag on the connected SmartEngage account; approval required
- add_tag_to_subscriber:
  - endpoint: POST tags/add
  - required fields: avatar_id, subscriber_id, tag
  - risk: external mutation; attaches an existing tag to a subscriber; approval required
- remove_tag_from_subscriber:
  - endpoint: POST tags/delete
  - required fields: avatar_id, subscriber_id, tag
  - risk: external mutation; detaches a tag from a subscriber; approval required
- create_custom_field:
  - endpoint: POST customfields/create
  - required fields: avatar_id, custom_field_name, custom_field_type
  - risk: external mutation; creates a new custom field definition on the connected SmartEngage account; approval required
- set_custom_field_value:
  - endpoint: POST customfields/update
  - required fields: avatar_id, subscriber_id, field, value
  - risk: external mutation; sets a custom field value on a subscriber; approval required
- add_subscriber_to_sequence:
  - endpoint: POST sequences/add
  - required fields: avatar_id, subscriber_id, sequence
  - risk: external mutation; enrolls a subscriber into an automation sequence, triggering scheduled messages; approval required
- remove_subscriber_from_sequence:
  - endpoint: POST sequences/remove
  - required fields: avatar_id, subscriber_id, sequence
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
