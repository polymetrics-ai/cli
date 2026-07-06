# pm connectors inspect tinyemail

```text
NAME
  pm connectors inspect tinyemail - TinyEmail connector manual

SYNOPSIS
  pm connectors inspect tinyemail
  pm connectors inspect tinyemail --json
  pm credentials add <name> --connector tinyemail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads subscribers, lists, and campaigns, and writes subscriber create/upsert actions, through the tinyEmail API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  subscribers:
    primary key: id
    fields: created_at(), email(), first_name(), id(), last_name(), status()
  lists:
    primary key: id
    fields: created_at(), id(), name(), subscriber_count()
  campaigns:
    primary key: id
    fields: created_at(), id(), name(), status(), subject()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_subscriber:
    endpoint: POST /segment/customer
    risk: creates or upserts a subscriber (customer) into the caller's tinyEmail account, optionally into a named audience segment; low-risk external mutation, no approval required

SECURITY
  read risk: external tinyEmail API read of subscriber, list, and campaign data
  write risk: external tinyEmail API mutation: creates or upserts a subscriber (customer) record, optionally assigning it to a named audience segment
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tinyemail

  # Inspect as structured JSON
  pm connectors inspect tinyemail --json

AGENT WORKFLOW
  - Run pm connectors inspect tinyemail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
