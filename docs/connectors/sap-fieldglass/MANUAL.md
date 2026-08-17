# pm connectors inspect sap-fieldglass

```text
NAME
  pm connectors inspect sap-fieldglass - SAP Fieldglass connector manual

SYNOPSIS
  pm connectors inspect sap-fieldglass
  pm connectors inspect sap-fieldglass --json
  pm credentials add <name> --connector sap-fieldglass [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SAP Fieldglass workers, job postings, and time sheets through the SAP Fieldglass API. Read-only.

ICON
  asset: icons/sapfieldglass.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.sap.com/package/SAPFieldglass/rest

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  access_token (secret)

ETL STREAMS
  workers:
    primary key: id
    fields: id(), name(), status(), stream()
  job_postings:
    primary key: id
    fields: id(), name(), status(), stream()
  time_sheets:
    primary key: id
    fields: id(), name(), status(), stream()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external SAP Fieldglass API read of worker, job posting, and time sheet data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sap-fieldglass

  # Inspect as structured JSON
  pm connectors inspect sap-fieldglass --json

AGENT WORKFLOW
  - Run pm connectors inspect sap-fieldglass before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
