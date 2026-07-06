# pm connectors inspect interzoid

```text
NAME
  pm connectors inspect interzoid - Interzoid connector manual

SYNOPSIS
  pm connectors inspect interzoid
  pm connectors inspect interzoid --json
  pm credentials add <name> --connector interzoid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Interzoid data-matching lookups: company-name, individual-name, and street-address similarity keys, plus organization-name standardization, via the Interzoid REST API.

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
  address
  address_match_algorithm
  base_url
  company
  company_match_algorithm
  fullname
  org
  api_key (secret)

ETL STREAMS
  company_name_matching:
    primary key: SimKey
    fields: Code(), Credits(), SimKey(), query_company()
  individual_name_matching:
    primary key: SimKey
    fields: Code(), Credits(), SimKey(), query_fullname()
  street_address_matching:
    primary key: SimKey
    fields: Code(), Credits(), SimKey(), query_address()
  standardize_company_names:
    primary key: Standard
    fields: Code(), Credits(), Standard(), query_org()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Interzoid API single-lookup read; each read spends an API credit
  approval: none; read-only data-matching lookup API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect interzoid

  # Inspect as structured JSON
  pm connectors inspect interzoid --json

AGENT WORKFLOW
  - Run pm connectors inspect interzoid before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
