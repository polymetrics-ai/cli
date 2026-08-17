---
name: pm-mixmax
description: Mixmax connector knowledge and safe action guide.
---

# pm-mixmax

## Purpose

Reads Mixmax code snippets, messages, rules, sequences, and meeting types through the Mixmax REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- codesnippets:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(), background(), createdAt(), html(), language(), theme(), title(), userId()
- messages:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(), bcc(), cc(), created(), fileTrackingEnabled(), from(), linkTrackingEnabled(), sent(), sequence(), subject(), to(), trackingEnabled(), updatedAt(), userId()
- rules:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(), createdAt(), isPaused(), modifiedAt(), name(), trigger(), userId()
- sequences:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(), createdAt(), fileTrackingEnabled(), linkTrackingEnabled(), name(), notificationsEnabled(), syncToOrg(), timezone(), userId()
- meetingtypes:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(), createdAt(), durationMin(), link(), name(), type(), updatedAt(), userId()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mixmax API read of code snippet, message, rule, sequence, and meeting-type data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mixmax
```

### Inspect as structured JSON

```bash
pm connectors inspect mixmax --json
```

## Agent Rules

- Run pm connectors inspect mixmax before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
