# pm connectors inspect kyve

```text
NAME
  pm connectors inspect kyve - KYVE connector manual

SYNOPSIS
  pm connectors inspect kyve
  pm connectors inspect kyve --json
  pm credentials add <name> --connector kyve [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public KYVE pools, stakers, funders, and Cosmos validators through the KYVE network's public REST query endpoints. Read-only; no credentials required.

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
  max_pages
  mode
  page_size

ETL STREAMS
  pools:
    primary key: id
    fields: id(), name(), runtime()
  stakers:
    primary key: address
    fields: address(), amount()
  funders:
    primary key: address
    fields: address(), amount()
  validators:
    primary key: operator_address
    fields: moniker(), operator_address(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external read of public KYVE network pool/staker/funder/validator data
  approval: none; read-only public Cosmos-style REST API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect kyve

  # Inspect as structured JSON
  pm connectors inspect kyve --json

AGENT WORKFLOW
  - Run pm connectors inspect kyve before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
