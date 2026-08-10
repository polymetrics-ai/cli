---
name: pm-pypi
description: PyPI connector knowledge and safe action guide.
---

# pm-pypi

## Purpose

Reads PyPI project metadata through the PyPI JSON API. Read-only and credential-free.

## Icon

- id: pypi
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
  - fields: author(string), author_email(string), classifiers(array), description(string), home_page(string), keywords(string), license(string), name(string), project_url(string), project_urls(object), requires_python(string), summary(string), version(string), yanked(boolean)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external PyPI JSON API read of public package metadata
- approval: none; read-only public package registry API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run PyPI's declared streams and reverse-ETL actions.
- Usage: pm pypi <command> [flags]
- Read streams
- Other Commands
  - api get pypi project version json - Documented GET /pypi/{project}/{version}/json (not implemented) [intent=direct_read availability=not_implemented operation=pypi.get.pypi-project-version-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get simple project - Documented GET /simple/{project}/ (not implemented) [intent=direct_read availability=not_implemented operation=pypi.get.simple-project]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stats - Documented GET /stats (not implemented) [intent=direct_read availability=not_implemented operation=pypi.get.stats]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - project list - Run the project ETL stream [intent=etl availability=implemented stream=project]

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
