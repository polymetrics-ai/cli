# pm connectors inspect insightly

```text
NAME
  pm connectors inspect insightly - Insightly connector manual

SYNOPSIS
  pm connectors inspect insightly
  pm connectors inspect insightly --json
  pm credentials add <name> --connector insightly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Insightly CRM contacts, organisations, opportunities, leads, projects, and tasks through the Insightly REST API v3.1.

ICON
  id: insightly
  asset: icons/insightly.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  token (secret) (required)

ETL STREAMS
  contacts:
    primary key: id
    cursor: date_updated_utc
    fields: contact_id(integer), date_created_utc(string), date_updated_utc(string), email_address(string), first_name(string), id(integer), last_name(string), organisation_id(integer), phone(string), title(string)
  organisations:
    primary key: id
    cursor: date_updated_utc
    fields: date_created_utc(string), date_updated_utc(string), id(integer), organisation_id(integer), organisation_name(string), owner_user_id(integer), phone(string), website(string)
  opportunities:
    primary key: id
    cursor: date_updated_utc
    fields: bid_currency(string), date_created_utc(string), date_updated_utc(string), id(integer), opportunity_id(integer), opportunity_name(string), opportunity_state(string), opportunity_value(number), pipeline_id(integer), probability(integer), stage_id(integer)
  leads:
    primary key: id
    cursor: date_updated_utc
    fields: converted(boolean), date_created_utc(string), date_updated_utc(string), email(string), first_name(string), id(integer), last_name(string), lead_id(integer), lead_source_id(integer), lead_status_id(integer), organisation_name(string)
  projects:
    primary key: id
    cursor: date_updated_utc
    fields: date_created_utc(string), date_updated_utc(string), id(integer), owner_user_id(integer), pipeline_id(integer), project_id(integer), project_name(string), stage_id(integer), status(string)
  tasks:
    primary key: id
    cursor: date_updated_utc
    fields: completed(boolean), date_created_utc(string), date_updated_utc(string), due_date(string), id(integer), owner_user_id(integer), priority(integer), status(string), task_id(integer), title(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Insightly API read of contacts, organisations, opportunities, leads, projects, and tasks
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect insightly

  # Inspect as structured JSON
  pm connectors inspect insightly --json

AGENT WORKFLOW
  - Run pm connectors inspect insightly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
