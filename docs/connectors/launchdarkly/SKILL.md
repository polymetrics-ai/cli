---
name: pm-launchdarkly
description: LaunchDarkly connector knowledge and safe action guide.
---

# pm-launchdarkly

## Purpose

Reads LaunchDarkly projects, members, audit log entries, feature flags, and environments through the LaunchDarkly REST API.

## Icon

- id: launchdarkly
- asset: icons/launchdarkly.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://apidocs.launchdarkly.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- project_key
- access_token (secret) (required)

## ETL Streams

- projects:
  - primary key: _id
  - fields: _id(string), key(string), name(string), tags(array)
- members:
  - primary key: _id
  - fields: _id(string), _pendingInvite(boolean), email(string), firstName(string), lastName(string), role(string)
- auditlog:
  - primary key: _id
  - cursor: date
  - fields: _id(string), date(integer), description(string), kind(string), name(string), shortDescription(string)
- flags:
  - primary key: key
  - fields: creationDate(integer), description(string), key(string), kind(string), name(string), tags(array), temporary(boolean)
- environments:
  - primary key: _id
  - fields: _id(string), color(string), defaultTtl(integer), key(string), name(string), tags(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external LaunchDarkly API read of project, membership, audit, and feature-flag configuration data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect launchdarkly
```

### Inspect as structured JSON

```bash
pm connectors inspect launchdarkly --json
```

## Agent Rules

- Run pm connectors inspect launchdarkly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
