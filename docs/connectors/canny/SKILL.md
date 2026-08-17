---
name: pm-canny
description: Canny connector knowledge and safe action guide.
---

# pm-canny

## Purpose

Reads Canny boards, posts, comments, categories, and companies through the Canny REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
- api_key (secret)

## ETL Streams

- boards:
  - primary key: id
  - cursor: created
  - fields: created(), id(), isPrivate(), name(), postCount(), url()
- posts:
  - primary key: id
  - cursor: created
  - fields: commentCount(), created(), details(), eta(), id(), score(), status(), statusChangedAt(), title(), url()
- comments:
  - primary key: id
  - cursor: created
  - fields: created(), id(), internal(), likeCount(), parentID(), private(), value()
- categories:
  - primary key: id
  - cursor: created
  - fields: created(), id(), name(), parentID(), postCount(), url()
- companies:
  - primary key: id
  - cursor: created
  - fields: created(), domain(), id(), memberCount(), monthlySpend(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Canny API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect canny
```

### Inspect as structured JSON

```bash
pm connectors inspect canny --json
```

## Agent Rules

- Run pm connectors inspect canny before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
