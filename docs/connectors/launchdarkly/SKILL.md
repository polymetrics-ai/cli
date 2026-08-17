---
name: pm-launchdarkly
description: LaunchDarkly connector knowledge and safe action guide.
---

# pm-launchdarkly

## Purpose

Reads LaunchDarkly projects, members, audit log entries, feature flags, and environments through the LaunchDarkly REST API.

## Icon

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
- access_token (secret)

## ETL Streams

- projects:
  - primary key: _id
  - fields: _id(), key(), name(), tags()
- members:
  - primary key: _id
  - fields: _id(), _pendingInvite(), email(), firstName(), lastName(), role()
- auditlog:
  - primary key: _id
  - cursor: date
  - fields: _id(), date(), description(), kind(), name(), shortDescription()
- flags:
  - primary key: key
  - fields: creationDate(), description(), key(), kind(), name(), tags(), temporary()
- environments:
  - primary key: _id
  - fields: _id(), color(), defaultTtl(), key(), name(), tags()

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
