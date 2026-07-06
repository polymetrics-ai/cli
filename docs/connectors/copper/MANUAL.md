# pm connectors inspect copper

```text
NAME
  pm connectors inspect copper - Copper connector manual

SYNOPSIS
  pm connectors inspect copper
  pm connectors inspect copper --json
  pm credentials add <name> --connector copper [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Copper CRM people, companies, opportunities, leads, and tasks through the Copper REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/copper.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.copper.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  user_email
  api_key (secret)

ETL STREAMS
  people:
    primary key: id
    cursor: date_modified
    fields: assignee_id(), company_id(), company_name(), contact_type_id(), date_created(), date_modified(), emails(), first_name(), id(), last_name(), name(), phone_numbers(), prefix(), title()
  companies:
    primary key: id
    cursor: date_modified
    fields: address(), assignee_id(), date_created(), date_modified(), details(), email_domain(), id(), name(), phone_numbers(), websites()
  opportunities:
    primary key: id
    cursor: date_modified
    fields: assignee_id(), close_date(), company_id(), company_name(), date_created(), date_modified(), id(), monetary_value(), name(), pipeline_id(), pipeline_stage_id(), primary_contact_id(), status(), win_probability()
  leads:
    primary key: id
    cursor: date_modified
    fields: assignee_id(), company_name(), date_created(), date_modified(), email(), id(), monetary_value(), name(), phone_numbers(), status(), status_id(), title()
  tasks:
    primary key: id
    cursor: date_modified
    fields: assignee_id(), completed_date(), date_created(), date_modified(), details(), due_date(), id(), name(), priority(), related_resource(), reminder_date(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Copper API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect copper

  # Inspect as structured JSON
  pm connectors inspect copper --json

AGENT WORKFLOW
  - Run pm connectors inspect copper before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
