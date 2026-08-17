# pm connectors inspect flowlu

```text
NAME
  pm connectors inspect flowlu - Flowlu connector manual

SYNOPSIS
  pm connectors inspect flowlu
  pm connectors inspect flowlu --json
  pm credentials add <name> --connector flowlu [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Flowlu CRM accounts, leads, tasks, projects, invoices, and agile issues through the Flowlu REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  company
  api_key (secret)

ETL STREAMS
  accounts:
    primary key: id
    cursor: updated_date
    fields: active(), created_date(), email(), first_name(), id(), last_name(), name(), owner_id(), phone(), type(), updated_date()
  leads:
    primary key: id
    cursor: updated_date
    fields: active(), budget(), created_date(), id(), name(), owner_id(), pipeline_id(), stage_id(), title(), updated_date()
  tasks:
    primary key: id
    cursor: updated_date
    fields: created_date(), deadline(), description(), id(), name(), owner_id(), priority(), responsible_id(), updated_date(), workflow_stage_id()
  projects:
    primary key: id
    cursor: updated_date
    fields: active(), created_date(), description(), id(), manager_id(), name(), owner_id(), stage_id(), updated_date()
  invoices:
    primary key: id
    cursor: updated_date
    fields: created_date(), currency_id(), customer_id(), id(), invoice_date(), invoice_number(), invoice_status(), name(), total_amount(), updated_date()
  agile_issues:
    primary key: id
    cursor: updated_date
    fields: created_date(), description(), id(), name(), owner_id(), priority(), project_id(), sprint_id(), type(), updated_date()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Flowlu CRM read of accounts/leads/tasks/projects/invoices/agile issues
  approval: none; read-only API-key access
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect flowlu

  # Inspect as structured JSON
  pm connectors inspect flowlu --json

AGENT WORKFLOW
  - Run pm connectors inspect flowlu before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
