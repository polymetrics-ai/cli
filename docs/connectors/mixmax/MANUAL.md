# pm connectors inspect mixmax

```text
NAME
  pm connectors inspect mixmax - Mixmax connector manual

SYNOPSIS
  pm connectors inspect mixmax
  pm connectors inspect mixmax --json
  pm credentials add <name> --connector mixmax [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mixmax code snippets, messages, rules, sequences, and meeting types through the Mixmax REST API.

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
  base_url
  api_key (secret) (required)

ETL STREAMS
  codesnippets:
    primary key: _id
    cursor: createdAt
    fields: _id(string), background(string), createdAt(string), html(string), language(string), theme(string), title(string), userId(string)
  messages:
    primary key: _id
    cursor: updatedAt
    fields: _id(string), bcc(string), cc(string), created(string), fileTrackingEnabled(boolean), from(string), linkTrackingEnabled(boolean), sent(string), sequence(string), subject(string), to(string), trackingEnabled(boolean), updatedAt(string), userId(string)
  rules:
    primary key: _id
    cursor: createdAt
    fields: _id(string), createdAt(string), isPaused(boolean), modifiedAt(string), name(string), trigger(string), userId(string)
  sequences:
    primary key: _id
    cursor: createdAt
    fields: _id(string), createdAt(string), fileTrackingEnabled(boolean), linkTrackingEnabled(boolean), name(string), notificationsEnabled(boolean), syncToOrg(boolean), timezone(string), userId(string)
  meetingtypes:
    primary key: _id
    cursor: createdAt
    fields: _id(string), createdAt(string), durationMin(integer), link(string), name(string), type(string), updatedAt(string), userId(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Mixmax API read of code snippet, message, rule, sequence, and meeting-type data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mixmax

  # Inspect as structured JSON
  pm connectors inspect mixmax --json

AGENT WORKFLOW
  - Run pm connectors inspect mixmax before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
