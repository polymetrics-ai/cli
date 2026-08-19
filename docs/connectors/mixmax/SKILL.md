---
name: pm-mixmax
description: Mixmax connector knowledge and safe action guide.
---

# pm-mixmax

## Purpose

Reads Mixmax code snippets, messages, rules, sequences, and meeting types through the Mixmax REST API.

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

- base_url
- api_key (secret) (required)

## ETL Streams

- codesnippets:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(string), background(string), createdAt(string), html(string), language(string), theme(string), title(string), userId(string)
- messages:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(string), bcc(string), cc(string), created(string), fileTrackingEnabled(boolean), from(string), linkTrackingEnabled(boolean), sent(string), sequence(string), subject(string), to(string), trackingEnabled(boolean), updatedAt(string), userId(string)
- rules:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(string), createdAt(string), isPaused(boolean), modifiedAt(string), name(string), trigger(string), userId(string)
- sequences:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(string), createdAt(string), fileTrackingEnabled(boolean), linkTrackingEnabled(boolean), name(string), notificationsEnabled(boolean), syncToOrg(boolean), timezone(string), userId(string)
- meetingtypes:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(string), createdAt(string), durationMin(integer), link(string), name(string), type(string), updatedAt(string), userId(string)

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
