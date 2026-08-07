# pm connectors inspect sigma-computing

```text
NAME
  pm connectors inspect sigma-computing - Sigma Computing connector manual

SYNOPSIS
  pm connectors inspect sigma-computing
  pm connectors inspect sigma-computing --json
  pm credentials add <name> --connector sigma-computing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Sigma workbooks, datasets, teams, and members through the Sigma REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  page_size
  access_token (secret)

ETL STREAMS
  workbooks:
    primary key: id
    cursor: updated_at
    fields: email(string), id(string), name(string), updated_at(string)
  datasets:
    primary key: id
    cursor: updated_at
    fields: email(string), id(string), name(string), updated_at(string)
  teams:
    primary key: id
    cursor: updated_at
    fields: email(string), id(string), name(string), updated_at(string)
  members:
    primary key: id
    cursor: updated_at
    fields: email(string), id(string), name(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Sigma Computing API read of workbook, dataset, team, and member data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sigma-computing

  # Inspect as structured JSON
  pm connectors inspect sigma-computing --json

AGENT WORKFLOW
  - Run pm connectors inspect sigma-computing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
