# pm connectors inspect feishu

```text
NAME
  pm connectors inspect feishu - Feishu / Lark connector manual

SYNOPSIS
  pm connectors inspect feishu
  pm connectors inspect feishu --json
  pm credentials add <name> --connector feishu [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Feishu/Lark Bitable (Base) records, tables, and field schemas via the Open Platform REST API using a tenant_access_token exchange. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
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
  base_url
  lark_host
  mode
  page_size
  table_id
  app_id (secret)
  app_secret (secret)
  app_token (secret)

ETL STREAMS
  records:
    primary key: record_id
    fields: fields(), record_id()
  tables:
    primary key: table_id
    fields: name(), revision(), table_id()
  fields:
    primary key: field_id
    fields: field_id(), field_name(), is_hidden(), is_primary(), property(), type(), ui_type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Feishu / Lark API reads performed by the legacy connector via a Tier-2 hook
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
