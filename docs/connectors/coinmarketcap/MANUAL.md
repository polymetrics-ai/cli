# pm connectors inspect coinmarketcap

```text
NAME
  pm connectors inspect coinmarketcap - CoinMarketCap connector manual

SYNOPSIS
  pm connectors inspect coinmarketcap
  pm connectors inspect coinmarketcap --json
  pm credentials add <name> --connector coinmarketcap [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads CoinMarketCap cryptocurrency map, latest market listings, categories, fiat currencies, and global metrics through the CoinMarketCap Pro API.

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
  pm connectors inspect coinmarketcap

  # Inspect as structured JSON
  pm connectors inspect coinmarketcap --json

AGENT WORKFLOW
  - Run pm connectors inspect coinmarketcap before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
