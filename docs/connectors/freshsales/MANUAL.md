# pm connectors inspect freshsales

```text
NAME
  pm connectors inspect freshsales - Freshsales connector manual

SYNOPSIS
  pm connectors inspect freshsales
  pm connectors inspect freshsales --json
  pm credentials add <name> --connector freshsales [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freshsales (Freshworks CRM) contacts, sales accounts, deals, and leads through the Freshsales REST API.

ICON
  id: freshsales
  asset: icons/freshsales.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.freshworks.com/crm/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  domain_name (required)
  max_pages
  mode
  view_id
  api_key (secret) (required)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: city(string), country(string), created_at(string), display_name(string), email(string), first_name(string), id(integer), job_title(string), last_name(string), mobile_number(string), owner_id(integer), updated_at(string), work_number(string)
  sales_accounts:
    primary key: id
    cursor: updated_at
    fields: annual_revenue(number), city(string), country(string), created_at(string), id(integer), industry_type_id(integer), name(string), number_of_employees(integer), owner_id(integer), phone(string), updated_at(string), website(string)
  deals:
    primary key: id
    cursor: updated_at
    fields: amount(number), created_at(string), currency_id(integer), deal_pipeline_id(integer), deal_stage_id(integer), expected_close(string), id(integer), name(string), owner_id(integer), probability(integer), sales_account_id(integer), updated_at(string)
  leads:
    primary key: id
    cursor: updated_at
    fields: city(string), company_name(string), country(string), created_at(string), display_name(string), email(string), first_name(string), id(integer), job_title(string), last_name(string), lead_stage_id(integer), owner_id(integer), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Freshsales API read of CRM contact, account, deal, and lead data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect freshsales

  # Inspect as structured JSON
  pm connectors inspect freshsales --json

AGENT WORKFLOW
  - Run pm connectors inspect freshsales before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
