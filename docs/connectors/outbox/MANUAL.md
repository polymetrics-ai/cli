# pm connectors inspect outbox

```text
NAME
  pm connectors inspect outbox - Local Outbox connector manual

SYNOPSIS
  pm connectors inspect outbox
  pm connectors inspect outbox --json
  pm credentials add <name> --connector outbox [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Local JSONL destination that records reverse ETL writes and receipts.

CAPABILITIES
  check=true catalog=true read=false write=true query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  path: Local outbox directory.

ETL STREAMS
  records: Reverse ETL outbox records.

SECURITY
  read risk: unsupported
  write risk: local file write
  mutation risk: reverse ETL receipt writes
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect outbox

  # Inspect as structured JSON
  pm connectors inspect outbox --json

  # Outbox reverse ETL
  pm credentials add outbox-local --connector outbox --config path=$ROOT/.polymetrics/outbox
  pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map email:email
  pm reverse run <plan-id> --approve <approval-token> --json

AGENT WORKFLOW
  - Run pm connectors inspect outbox before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
