---
name: pm-posthog
description: PostHog connector knowledge and safe action guide.
---

# pm-posthog

## Purpose

Reads PostHog events and persons for a project via the PostHog REST API. Read-only.

## Icon

- asset: icons/posthog.svg
- source: official
- review_status: official_verified
- review_url: https://posthog.com/docs/api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- project_id
- start_date
- api_key (secret)

## ETL Streams

- events:
  - primary key: id
  - cursor: timestamp
  - fields: distinct_id(), event(), id(), properties(), timestamp()
- persons:
  - primary key: id
  - fields: created_at(), distinct_id(), id(), properties()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external PostHog API read of project event and person data
- approval: none; read-only analytics API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect posthog
```

### Inspect as structured JSON

```bash
pm connectors inspect posthog --json
```

## Agent Rules

- Run pm connectors inspect posthog before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
