# pm connectors inspect aws-cloudtrail

```text
NAME
  pm connectors inspect aws-cloudtrail - AWS CloudTrail connector manual

SYNOPSIS
  pm connectors inspect aws-cloudtrail
  pm connectors inspect aws-cloudtrail --json
  pm credentials add <name> --connector aws-cloudtrail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads AWS CloudTrail configuration and resource metadata through fixed AWS JSON-RPC streams. Provider query/direct-read and write/admin actions remain planned until shared promoted-native forwarding exposes them safely at runtime. Breaking change: the earlier LookupEvents-backed management_events, read_only_events, write_only_events, and console_logins streams are removed and have no replacement here; CloudTrail event and Insights record reads are blocked/planned. See the connector docs migration note.

ICON
  asset: icons/aws-cloudtrail.svg
  source: upstream_registry
  review_status: upstream_seeded

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
