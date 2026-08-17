# pm connectors inspect plaid

```text
NAME
  pm connectors inspect plaid - Plaid connector manual

SYNOPSIS
  pm connectors inspect plaid
  pm connectors inspect plaid --json
  pm credentials add <name> --connector plaid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Plaid institutions and category metadata through read-only POST endpoints. All credentials and pagination/filter state travel in the JSON request body (Plaid's own convention), driven by a StreamHook.

ICON
  asset: icons/plaid.svg
  source: official
  review_status: official_verified
  review_url: https://plaid.com/docs/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  country_codes
  max_pages
  mode
  page_size
  client_id (secret)
  secret (secret)

ETL STREAMS
  institutions:
    primary key: institution_id
    fields: country_codes(), institution_id(), name()
  categories:
    primary key: category_id
    fields: category_id(), group(), hierarchy()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Plaid API read of institution/category metadata
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect plaid

  # Inspect as structured JSON
  pm connectors inspect plaid --json

AGENT WORKFLOW
  - Run pm connectors inspect plaid before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
