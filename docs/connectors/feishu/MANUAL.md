# pm connectors inspect feishu

```text
NAME
  pm connectors inspect feishu - Feishu / Lark connector manual

SYNOPSIS
  pm connectors inspect feishu
  pm connectors inspect feishu --json
  pm credentials add <name> --connector feishu [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Feishu/Lark Bitable records, tables, and field schemas through declared Bitable REST routes and a bounded tenant token exchange.

ICON
  id: feishu
  asset: icons/feishu.svg
  source: official
  review_status: official_verified
  review_url: https://open.feishu.cn/document/server-docs/docs/bitable-v1/bitable-overview

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  region
  table_id (required)
  app_id (secret) (required)
  app_secret (secret) (required)
  app_token (secret) (required)

ETL STREAMS
  records:
    primary key: record_id
    fields: fields(object), record_id(string)
  tables:
    primary key: table_id
    fields: name(string), revision(integer), table_id(string)
  fields:
    primary key: field_id
    fields: field_id(string), field_name(string), is_hidden(boolean), is_primary(boolean), property(object), type(integer), ui_type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: Bounded Feishu/Lark Bitable reads use a source-declared provider host and tenant token exchange.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect feishu

  # Inspect as structured JSON
  pm connectors inspect feishu --json

AGENT WORKFLOW
  - Run pm connectors inspect feishu before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
