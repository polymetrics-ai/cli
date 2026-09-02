---
name: pm-lokalise
description: Lokalise connector knowledge and safe action guide.
---

# pm-lokalise

## Purpose

Reads Lokalise project keys, languages, translations, contributors, and comments through the Lokalise REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: lokalise
- asset: icons/lokalise.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.lokalise.com/reference/api-introduction

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- project_id (required)
- api_key (secret) (required)

## ETL Streams

- keys:
  - primary key: key_id
  - cursor: modified_at_timestamp
  - fields: created_at(string), created_at_timestamp(integer), description(string), is_archived(boolean), is_hidden(boolean), is_plural(boolean), key_id(integer), key_name(object), modified_at(string), modified_at_timestamp(integer), platforms(array), tags(array)
- languages:
  - primary key: lang_id
  - fields: is_rtl(boolean), lang_id(integer), lang_iso(string), lang_name(string), plural_forms(array)
- translations:
  - primary key: translation_id
  - cursor: modified_at_timestamp
  - fields: is_reviewed(boolean), is_unverified(boolean), key_id(integer), language_iso(string), modified_at(string), modified_at_timestamp(integer), modified_by(integer), modified_by_email(string), reviewed_by(integer), translation(string), translation_id(integer)
- contributors:
  - primary key: user_id
  - fields: created_at(string), created_at_timestamp(integer), email(string), fullname(string), is_admin(boolean), is_reviewer(boolean), languages(array), role_id(integer), user_id(integer)
- comments:
  - primary key: comment_id
  - fields: added_at(string), added_at_timestamp(integer), added_by(integer), added_by_email(string), comment(string), comment_id(integer), key_id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Lokalise API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect lokalise
```

### Inspect as structured JSON

```bash
pm connectors inspect lokalise --json
```

## Agent Rules

- Run pm connectors inspect lokalise before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
