# pm connectors inspect amazon-sqs

```text
NAME
  pm connectors inspect amazon-sqs - Amazon SQS connector manual

SYNOPSIS
  pm connectors inspect amazon-sqs
  pm connectors inspect amazon-sqs --json
  pm credentials add <name> --connector amazon-sqs [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads messages from Amazon SQS via signed ReceiveMessage calls. Read-only; messages are not deleted.

ICON
  asset: icons/amazon-sqs.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: queue

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  No connector-specific config fields.

SECURITY
  read risk: connector-specific
  write risk: connector-specific
  approval: external mutations require preview and approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect amazon-sqs

  # Inspect as structured JSON
  pm connectors inspect amazon-sqs --json

AGENT WORKFLOW
  - Run pm connectors inspect amazon-sqs before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
