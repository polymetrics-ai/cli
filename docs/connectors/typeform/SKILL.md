---
name: pm-typeform
description: Typeform connector knowledge and safe action guide.
---

# pm-typeform

## Purpose

Reads Typeform forms, workspaces, themes, and images through the Typeform REST API.

## Icon

- asset: icons/typeform.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.typeform.com/developers/changelog/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- form_ids
- mode
- page_size
- access_token (secret)

## ETL Streams

- forms:
  - primary key: id
  - cursor: last_updated_at
  - fields: created_at(), id(), is_public(), last_updated_at(), self_href(), theme_href(), title(), type()
- responses:
  - primary key: response_id
  - cursor: submitted_at
  - fields: answers(), calculated(), form_id(), hidden(), landed_at(), landing_id(), metadata(), response_id(), submitted_at(), token()
- workspaces:
  - primary key: id
  - fields: account_id(), default(), id(), name(), self_href(), shared()
- themes:
  - primary key: id
  - fields: background(), colors(), font(), id(), name(), visibility()
- images:
  - primary key: id
  - fields: file_name(), has_alpha(), height(), id(), media_type(), src(), width()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Typeform API read of form, workspace, theme, and image metadata
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect typeform
```

### Inspect as structured JSON

```bash
pm connectors inspect typeform --json
```

## Agent Rules

- Run pm connectors inspect typeform before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
