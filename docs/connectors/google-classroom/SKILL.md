---
name: pm-google-classroom
description: Google Classroom connector knowledge and safe action guide.
---

# pm-google-classroom

## Purpose

Reads Google Classroom courses, teachers, students, course work, and announcements through the Classroom REST API using an OAuth2 refresh token. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
- mode
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- courses:
  - primary key: id
  - cursor: updateTime
  - fields: alternateLink(), calendarId(), courseGroupEmail(), courseState(), creationTime(), description(), descriptionHeading(), enrollmentCode(), guardiansEnabled(), id(), name(), ownerId(), room(), section(), teacherGroupEmail(), updateTime()
- teachers:
  - primary key: courseId, userId
  - fields: courseId(), emailAddress(), fullName(), photoUrl(), userId()
- students:
  - primary key: courseId, userId
  - fields: courseId(), emailAddress(), fullName(), photoUrl(), userId()
- course_work:
  - primary key: id
  - cursor: updateTime
  - fields: alternateLink(), courseId(), creationTime(), description(), dueDate(), id(), maxPoints(), state(), title(), updateTime(), workType()
- announcements:
  - primary key: id
  - cursor: updateTime
  - fields: alternateLink(), courseId(), creationTime(), creatorUserId(), id(), state(), text(), updateTime()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Google Classroom API reads performed by the legacy connector via a Tier-2 hook
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
