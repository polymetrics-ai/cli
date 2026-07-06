# pm connectors inspect pypi

```text
NAME
  pm connectors inspect pypi - PyPI connector manual

SYNOPSIS
  pm connectors inspect pypi
  pm connectors inspect pypi --json
  pm credentials add <name> --connector pypi [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PyPI project metadata through the PyPI JSON API. Read-only and credential-free.

ICON
  asset: icons/pypi.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://warehouse.pypa.io/api-reference/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  project_name

ETL STREAMS
  project:
    primary key: name
    fields: author(), author_email(), classifiers(), description(), home_page(), keywords(), license(), name(), project_url(), project_urls(), requires_python(), summary(), version(), yanked()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external PyPI JSON API read of public package metadata
  approval: none; read-only public package registry API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pypi

  # Inspect as structured JSON
  pm connectors inspect pypi --json

AGENT WORKFLOW
  - Run pm connectors inspect pypi before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
