---
name: pm-easypromos
description: Easypromos connector knowledge and safe action guide.
---

# pm-easypromos

## Purpose

Reads Easypromos promotions, organizing brands, stages, users, participations, and prizes through the Easypromos REST API.

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
- promotion_id
- bearer_token (secret)

## ETL Streams

- promotions:
  - primary key: id
  - fields: created(), default_language(), description(), end_date(), id(), organizing_brand_id(), organizing_brand_name(), promotion_type(), start_date(), status(), timezone(), title(), url()
- organizing_brands:
  - primary key: id
  - fields: id(), name()
- stages:
  - primary key: id
  - fields: end_date(), id(), name(), start_date(), type(), visible()
- users:
  - primary key: id
  - fields: country(), created(), email(), external_id(), first_name(), id(), language(), last_name(), login_type(), nickname(), promotion_id(), status()
- participations:
  - primary key: id
  - fields: created(), id(), ip(), promotion_id(), stage_id(), user_agent(), user_id()
- prizes:
  - primary key: id
  - fields: code(), created(), download_url(), id(), participation_id(), prize_type_id(), prize_type_name(), redeem_url(), stage_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Easypromos API read of promotion, user, participation, and prize data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect easypromos
```

### Inspect as structured JSON

```bash
pm connectors inspect easypromos --json
```

## Agent Rules

- Run pm connectors inspect easypromos before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
