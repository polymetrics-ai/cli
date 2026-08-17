---
name: pm-lokalise
description: Lokalise connector knowledge and safe action guide.
---

# pm-lokalise

## Purpose

Reads Lokalise project keys, languages, translations, contributors, and comments through the Lokalise REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

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
- project_id
- api_key (secret)

## ETL Streams

- keys:
  - primary key: key_id
  - cursor: modified_at_timestamp
  - fields: created_at(), created_at_timestamp(), description(), is_archived(), is_hidden(), is_plural(), key_id(), key_name(), modified_at(), modified_at_timestamp(), platforms(), tags()
- languages:
  - primary key: lang_id
  - fields: is_rtl(), lang_id(), lang_iso(), lang_name(), plural_forms()
- translations:
  - primary key: translation_id
  - cursor: modified_at_timestamp
  - fields: is_reviewed(), is_unverified(), key_id(), language_iso(), modified_at(), modified_at_timestamp(), modified_by(), modified_by_email(), reviewed_by(), translation(), translation_id()
- contributors:
  - primary key: user_id
  - fields: created_at(), created_at_timestamp(), email(), fullname(), is_admin(), is_reviewer(), languages(), role_id(), user_id()
- comments:
  - primary key: comment_id
  - fields: added_at(), added_at_timestamp(), added_by(), added_by_email(), comment(), comment_id(), key_id()

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
