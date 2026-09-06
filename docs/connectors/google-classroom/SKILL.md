---
name: pm-google-classroom
description: Google Classroom connector knowledge and safe action guide.
---

# pm-google-classroom

## Purpose

Reads Classroom courses and course-scoped resources through fixed REST routes and OAuth2 refresh-token authentication.

## Icon

- id: simple-icons-googleclassroom
- asset: icons/simple-icons/googleclassroom.svg
- title: Google Classroom
- simple_icon_slug: googleclassroom
- simple_icon_hex: 0F9D58
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Google%20Classroom
- match: exact-name-or-slug
- matched_by: google-classroom

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- client_id (secret) (required)
- client_refresh_token (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- courses:
  - primary key: id
  - cursor: updateTime
  - fields: alternateLink(string), courseState(string), creationTime(string), description(string), id(string), name(string), ownerId(string), section(string), updateTime(string)
- teachers:
  - primary key: courseId, userId
  - fields: courseId(string), emailAddress(string), fullName(string), photoUrl(string), userId(string)
- students:
  - primary key: courseId, userId
  - fields: courseId(string), emailAddress(string), fullName(string), photoUrl(string), userId(string)
- course_work:
  - primary key: id
  - cursor: updateTime
  - fields: alternateLink(string), courseId(string), creationTime(string), description(string), id(string), maxPoints(number), state(string), title(string), updateTime(string), workType(string)
- announcements:
  - primary key: id
  - cursor: updateTime
  - fields: alternateLink(string), courseId(string), creationTime(string), creatorUserId(string), id(string), state(string), text(string), updateTime(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: Bounded Classroom reads use fixed OAuth2 and REST routes.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-classroom
```

### Inspect as structured JSON

```bash
pm connectors inspect google-classroom --json
```

## Agent Rules

- Run pm connectors inspect google-classroom before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
