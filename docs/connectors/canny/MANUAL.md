# pm connectors inspect canny

```text
NAME
  pm connectors inspect canny - Canny connector manual

SYNOPSIS
  pm connectors inspect canny
  pm connectors inspect canny --json
  pm credentials add <name> --connector canny [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Canny boards, posts, comments, categories, and companies through fixed Canny REST form requests.

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
  api_key (secret) (required)

ETL STREAMS
  boards:
    primary key: id
    cursor: created
    fields: created(string), id(string)
  posts:
    primary key: id
    cursor: created
    fields: created(string), id(string)
  comments:
    primary key: id
    cursor: created
    fields: created(string), id(string)
  categories:
    primary key: id
    cursor: created
    fields: created(string), id(string)
  companies:
    primary key: id
    cursor: created
    fields: created(string), id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: Bounded Canny list requests carry the declared API key only in typed form bodies.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect canny

  # Inspect as structured JSON
  pm connectors inspect canny --json

AGENT WORKFLOW
  - Run pm connectors inspect canny before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
