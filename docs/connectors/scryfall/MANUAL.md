# pm connectors inspect scryfall

```text
NAME
  pm connectors inspect scryfall - Scryfall connector manual

SYNOPSIS
  pm connectors inspect scryfall
  pm connectors inspect scryfall --json
  pm credentials add <name> --connector scryfall [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads cards and sets from the public Scryfall API. Read-only and credential-free.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  q

ETL STREAMS
  cards:
    primary key: id
    fields: id(), name(), set()
  sets:
    primary key: id
    fields: id(), name(), set()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: public, credential-free Scryfall API read of card and set data
  approval: none; read-only, no reverse-ETL writes implemented by legacy
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect scryfall

  # Inspect as structured JSON
  pm connectors inspect scryfall --json

AGENT WORKFLOW
  - Run pm connectors inspect scryfall before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
