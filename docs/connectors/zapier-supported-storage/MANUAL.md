# pm connectors inspect zapier-supported-storage

```text
NAME
  pm connectors inspect zapier-supported-storage - Zapier Supported Storage connector manual

SYNOPSIS
  pm connectors inspect zapier-supported-storage
  pm connectors inspect zapier-supported-storage --json
  pm credentials add <name> --connector zapier-supported-storage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Zapier Storage key/value records.

ICON
  asset: icons/zapiersupportedstorage.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://help.zapier.com/hc/en-us/articles/8496293271053-Save-and-retrieve-data-from-Zaps

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  secret (secret)

ETL STREAMS
  records:
    primary key: id
    cursor: updated_at
    fields: id(), key(), updated_at(), value()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  set_record:
    endpoint: PATCH /api/records
    risk: creates or overwrites a single key/value pair in the caller's Zapier Storage bucket (optionally only when the existing value matches only_if_value); external mutation, no approval required
  increment_record:
    endpoint: PATCH /api/records
    risk: atomically increments a numeric-valued key by amount (creating it at amount if absent); external mutation, no approval required
  delete_record:
    endpoint: DELETE /api/records?key={{ record.key }}
    required fields: key
    risk: irreversibly deletes a single key from the caller's Zapier Storage bucket
  delete_all_records:
    endpoint: DELETE /api/records
    risk: irreversibly deletes EVERY key in the caller's Zapier Storage bucket (whole-bucket wipe); destructive, requires explicit confirmation

SECURITY
  read risk: external Zapier Storage API read of stored key/value records
  write risk: external mutation of a shared per-Zap/per-app key/value store: set/increment a single key, delete a single key, or wipe the entire bucket (delete_all_records, destructive)
  approval: required for write actions; delete_all_records requires explicit destructive confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zapier-supported-storage

  # Inspect as structured JSON
  pm connectors inspect zapier-supported-storage --json

AGENT WORKFLOW
  - Run pm connectors inspect zapier-supported-storage before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
