---
name: pm-sentry
description: Sentry connector knowledge and safe action guide.
---

# pm-sentry

## Purpose

Reads Sentry projects, issues, error events, and releases through the Sentry REST API (read-only).

## Icon

- id: sentry
- asset: icons/sentry.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.sentry.io/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- organization
- page_size
- project
- auth_token (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), id(string), isBookmarked(boolean), isPublic(boolean), name(string), platform(string), slug(string), status(string)
- issues:
  - primary key: id
  - cursor: lastSeen
  - fields: count(string), culprit(string), firstSeen(string), id(string), lastSeen(string), level(string), shortId(string), status(string), title(string), type(string), userCount(integer)
- events:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), eventID(string), groupID(string), id(string), message(string), platform(string), title(string), type(string)
- releases:
  - primary key: version
  - cursor: dateCreated
  - fields: dateCreated(string), dateReleased(string), ref(string), shortVersion(string), status(string), url(string), version(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Sentry API read of project, issue, event, and release data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Sentry command surface
- Usage: pm sentry <command>
- Seer
- Other Commands
  - seer list-models - List the declared Sentry Seer models. [intent=direct_read availability=implemented operation=sentry.seer_models_list]; flags: --page, --page-cursor

## Commands

### Inspect as a manual

```bash
pm connectors inspect sentry
```

### Inspect as structured JSON

```bash
pm connectors inspect sentry --json
```

## Agent Rules

- Run pm connectors inspect sentry before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
