---
name: pm-fullstory
description: Fullstory connector knowledge and safe action guide.
---

# pm-fullstory

## Purpose

Reads FullStory segments, users, events, and user-scoped sessions; writes server-side user and custom event data through the FullStory Server API.

## Icon

- asset: icons/fullstory.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.fullstory.com/reference

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- session_email
- session_uid
- api_key (secret)
- uid (secret)

## ETL Streams

- segments:
  - primary key: id
  - cursor: created
  - fields: created(), creator(), description(), id(), is_public(), name(), type()
- users:
  - primary key: id
  - cursor: created
  - fields: created(), display_name(), email(), id(), is_being_processed(), uid(), updated()
- events:
  - primary key: id
  - cursor: event_time
  - fields: device_id(), event_time(), id(), name(), session_id(), type(), user_id()
- sessions:
  - primary key: id
  - fields: app_url(), duration_ms(), email(), id(), start_time(), uid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_user:
  - endpoint: POST /v2/users
  - risk: creates or upserts a FullStory user profile and associated custom user properties
- update_user:
  - endpoint: POST /v2/users/{{ record.id }}
  - required fields: id
  - risk: updates a FullStory user profile's display fields or custom properties
- create_event:
  - endpoint: POST /v2/events
  - risk: creates a custom FullStory event that becomes part of analytics/session context

## Security

- read risk: external FullStory API read of session-analytics segment, user, event, and user-scoped session data
- write risk: creates or updates FullStory server-side user attributes and custom events used for analytics segmentation
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect fullstory
```

### Inspect as structured JSON

```bash
pm connectors inspect fullstory --json
```

## Agent Rules

- Run pm connectors inspect fullstory before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
