# pm connectors inspect aws-cloudtrail

```text
NAME
  pm connectors inspect aws-cloudtrail - AWS CloudTrail connector manual

SYNOPSIS
  pm connectors inspect aws-cloudtrail
  pm connectors inspect aws-cloudtrail --json
  pm credentials add <name> --connector aws-cloudtrail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads AWS CloudTrail management events (last 90 days) via the LookupEvents API using AWS Signature V4 authentication. Read-only.

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

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
  pm connectors inspect aws-cloudtrail

  # Inspect as structured JSON
  pm connectors inspect aws-cloudtrail --json

AGENT WORKFLOW
  - Run pm connectors inspect aws-cloudtrail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
