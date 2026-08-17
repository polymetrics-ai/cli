---
name: pm-pypi
description: PyPI connector knowledge and safe action guide.
---

# pm-pypi

## Purpose

Reads PyPI project metadata through the PyPI JSON API. Read-only and credential-free.

## Icon

- asset: icons/pypi.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://warehouse.pypa.io/api-reference/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- project_name

## ETL Streams

- project:
  - primary key: name
  - fields: author(), author_email(), classifiers(), description(), home_page(), keywords(), license(), name(), project_url(), project_urls(), requires_python(), summary(), version(), yanked()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external PyPI JSON API read of public package metadata
- approval: none; read-only public package registry API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pypi
```

### Inspect as structured JSON

```bash
pm connectors inspect pypi --json
```

## Agent Rules

- Run pm connectors inspect pypi before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
