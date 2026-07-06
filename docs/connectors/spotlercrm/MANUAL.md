# pm connectors inspect spotlercrm

```text
NAME
  pm connectors inspect spotlercrm - Spotler CRM connector manual

SYNOPSIS
  pm connectors inspect spotlercrm
  pm connectors inspect spotlercrm --json
  pm credentials add <name> --connector spotlercrm [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Spotler CRM contacts, accounts, opportunities, and tasks, and (via the real CRM API v4) activities, campaigns, and cases.

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
  access_token (secret)
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: id
    fields: email(), firstName(), id(), lastName()
  accounts:
    primary key: id
    fields: id(), name(), status()
  opportunities:
    primary key: id
    fields: id(), name(), status()
  tasks:
    primary key: id
    fields: id(), name(), status()
  activities:
    primary key: id
    fields: createddate(), id(), modifieddate(), ownerid()
  campaigns:
    primary key: id
    fields: createddate(), id(), modifieddate(), name(), ownerid()
  cases:
    primary key: id
    fields: createddate(), id(), modifieddate(), ownerid()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Spotler CRM API read of contact, account, opportunity, task, activity, campaign, and case data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect spotlercrm

  # Inspect as structured JSON
  pm connectors inspect spotlercrm --json

AGENT WORKFLOW
  - Run pm connectors inspect spotlercrm before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
