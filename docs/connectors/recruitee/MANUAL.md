# pm connectors inspect recruitee

```text
NAME
  pm connectors inspect recruitee - Recruitee connector manual

SYNOPSIS
  pm connectors inspect recruitee
  pm connectors inspect recruitee --json
  pm credentials add <name> --connector recruitee [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Recruitee offers, candidates, departments, sources, and tags through the Recruitee REST API.

ICON
  asset: icons/recruitee.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  company_id
  api_key (secret)

ETL STREAMS
  offers:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), status(), title(), updated_at()
  candidates:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), id(), name(), updated_at()
  departments:
    primary key: id
    fields: id(), name()
  sources:
    primary key: id
    fields: id(), name()
  tags:
    primary key: id
    fields: id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Recruitee API read of ATS offer and candidate data
  approval: none; read-only ATS API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect recruitee

  # Inspect as structured JSON
  pm connectors inspect recruitee --json

AGENT WORKFLOW
  - Run pm connectors inspect recruitee before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
