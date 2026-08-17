# pm connectors inspect free-agent-connector

```text
NAME
  pm connectors inspect free-agent-connector - FreeAgent connector manual

SYNOPSIS
  pm connectors inspect free-agent-connector
  pm connectors inspect free-agent-connector --json
  pm credentials add <name> --connector free-agent-connector [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads FreeAgent contacts, invoices, bills, projects, and tasks through the FreeAgent v2 REST API using OAuth2 refresh-token authentication. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  base_url
  mode
  payroll_year
  updated_since
  client_id (secret)
  client_refresh_token_2 (secret)
  client_secret (secret)

ETL STREAMS
  contacts:
    primary key: url
    cursor: updated_at
    fields: account_balance(), created_at(), email(), first_name(), last_name(), organisation_name(), phone_number(), status(), updated_at(), url()
  invoices:
    primary key: url
    cursor: updated_at
    fields: contact(), created_at(), currency(), dated_on(), due_on(), due_value(), net_value(), reference(), status(), total_value(), updated_at(), url()
  bills:
    primary key: url
    cursor: updated_at
    fields: contact(), created_at(), currency(), dated_on(), due_on(), due_value(), reference(), status(), total_value(), updated_at(), url()
  projects:
    primary key: url
    cursor: updated_at
    fields: budget(), budget_units(), contact(), created_at(), currency(), name(), status(), updated_at(), url()
  tasks:
    primary key: url
    cursor: updated_at
    fields: billing_period(), billing_rate(), created_at(), is_billable(), name(), project(), status(), updated_at(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external FreeAgent API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect free-agent-connector

  # Inspect as structured JSON
  pm connectors inspect free-agent-connector --json

AGENT WORKFLOW
  - Run pm connectors inspect free-agent-connector before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
