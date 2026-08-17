---
name: pm-just-sift
description: JustSift connector knowledge and safe action guide.
---

# pm-just-sift

## Purpose

Reads JustSift people directory profiles and person field definitions through the Sift REST API.

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
- api_token (secret)

## ETL Streams

- peoples:
  - primary key: id
  - fields: companyName(), connector(), department(), directReportCount(), directoryId(), displayName(), email(), firstName(), id(), isTeamLeader(), lastName(), officeCity(), officeState(), phone(), pictureUrl(), title()
- fields:
  - primary key: id
  - fields: connector(), displayName(), filterable(), id(), objectKey(), searchable(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external JustSift API read of people directory profiles and field definitions
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect just-sift
```

### Inspect as structured JSON

```bash
pm connectors inspect just-sift --json
```

## Agent Rules

- Run pm connectors inspect just-sift before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
