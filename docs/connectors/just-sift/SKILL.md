---
name: pm-just-sift
description: JustSift connector knowledge and safe action guide.
---

# pm-just-sift

## Purpose

Reads JustSift people directory profiles and person field definitions through the Sift REST API.

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
- mode
- api_token (secret) (required)

## ETL Streams

- peoples:
  - primary key: id
  - fields: companyName(string), connector(string), department(string), directReportCount(number), directoryId(string), displayName(string), email(string), firstName(string), id(string), isTeamLeader(boolean), lastName(string), officeCity(string), officeState(string), phone(string), pictureUrl(string), title(string)
- fields:
  - primary key: id
  - fields: connector(string), displayName(string), filterable(boolean), id(string), objectKey(string), searchable(boolean), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
