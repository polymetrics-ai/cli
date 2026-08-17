# pm connectors inspect dynamodb

```text
NAME
  pm connectors inspect dynamodb - DynamoDB connector manual

SYNOPSIS
  pm connectors inspect dynamodb
  pm connectors inspect dynamodb --json
  pm credentials add <name> --connector dynamodb [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads DynamoDB table items through the AWS JSON HTTP API (DynamoDB_20120810.Scan), authenticated with hand-rolled AWS Signature Version 4 request signing. Read-only source; no write support.

ICON
  asset: icons/dynamodb.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: database

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
  pm connectors inspect dynamodb

  # Inspect as structured JSON
  pm connectors inspect dynamodb --json

AGENT WORKFLOW
  - Run pm connectors inspect dynamodb before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
