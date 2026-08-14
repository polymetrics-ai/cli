# pm connectors inspect countercyclical

```text
NAME
  pm connectors inspect countercyclical - Countercyclical connector manual

SYNOPSIS
  pm connectors inspect countercyclical
  pm connectors inspect countercyclical --json
  pm credentials add <name> --connector countercyclical [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Countercyclical investments, valuations, research memos, teams, assumptions, and pipelines, and creates investments, through the Countercyclical REST API.

ICON
  id: pm-sample
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
  mode
  api_key (secret) (required)

ETL STREAMS
  investments:
    primary key: id
    fields: cik(string), country(string), createdAt(string), description(string), editedName(string), employees(integer), exchange(string), figi(string), financingType(string), id(string), industry(string), isArchived(boolean), isFavorite(boolean), isLocked(boolean), lei(string), marketType(string), name(string), sector(string), tickerSymbol(string), updatedAt(string), visibility(string), website(string)
  valuations:
    primary key: id
    fields: createdAt(string), delineation(string), description(string), discountRate(number), endingQuarter(integer), endingYear(integer), growthMetric(string), growthRate(number), id(string), isFavorite(boolean), name(string), shareToken(string), startingQuarter(integer), startingYear(integer), status(string), terminalPeriod(string), terminalRate(number), updatedAt(string)
  memos:
    primary key: id
    fields: archived(boolean), backgroundColor(string), bannerVisible(boolean), body(string), createdAt(string), documentType(string), emoji(string), favorited(boolean), foregroundColor(string), id(string), locked(boolean), publiclyVisible(boolean), sourcesVisible(boolean), title(string), tocVisible(boolean), updatedAt(string), views(integer)
  teams:
    primary key: id
    fields: id(string), title(string)
  assumptions:
    primary key: id
    fields: discountRate(string), id(string), name(string)
  pipelines:
    primary key: id
    fields: id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_investment:
    endpoint: POST /integrations/make/actions/investments
    required fields: tickerSymbol
    risk: creates a new Investment in the caller's Countercyclical workspace via the Make-integration action endpoint (the only documented general-purpose creation endpoint; the functionally-identical Zapier-integration endpoint is not separately exposed, see api_surface.json); external mutation, no approval required

SECURITY
  read risk: external Countercyclical API read of investment and valuation data
  write risk: external mutation: creates a new Investment record in the caller's workspace; no update/delete actions are exposed
  approval: required for the create_investment write action; read-only otherwise
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect countercyclical

  # Inspect as structured JSON
  pm connectors inspect countercyclical --json

AGENT WORKFLOW
  - Run pm connectors inspect countercyclical before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
