# pm connectors inspect twenty

```text
NAME
  pm connectors inspect twenty - Twenty CRM connector manual

SYNOPSIS
  pm connectors inspect twenty
  pm connectors inspect twenty --json
  pm credentials add <name> --connector twenty [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Twenty CRM companies, people, opportunities, notes, tasks, messages, calendar events, workflows, workspace members, and the rest of the 28-object workspace surface, and writes create/update/batch/delete mutations through the Twenty REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  replication_start_date
  api_key (secret)

SECURITY
  read risk: external Twenty CRM API read of CRM, messaging, calendar, workflow, and workspace-member data
  write risk: creates, updates, batch-writes, and deletes records across all 28 Twenty CRM REST objects
  approval: required for every update_<object>, batch_<object>, and delete_<object> action across the Twenty REST object surface; create_<object> actions require no approval (low-risk, non-destructive)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect twenty

  # Inspect as structured JSON
  pm connectors inspect twenty --json

AGENT WORKFLOW
  - Run pm connectors inspect twenty before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
