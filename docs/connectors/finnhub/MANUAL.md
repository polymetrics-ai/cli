# pm connectors inspect finnhub

```text
NAME
  pm connectors inspect finnhub - Finnhub connector manual

SYNOPSIS
  pm connectors inspect finnhub
  pm connectors inspect finnhub --json
  pm credentials add <name> --connector finnhub [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Finnhub stock symbols, market news, per-symbol company profiles, and per-symbol analyst recommendation trends through the Finnhub REST API.

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
  exchange
  market_news_category
  mode
  symbols
  api_key (secret) (required)

ETL STREAMS
  stock_symbols:
    primary key: symbol
    fields: currency(string), description(string), displaySymbol(string), figi(string), mic(string), symbol(string), type(string)
  market_news:
    primary key: id
    cursor: datetime
    fields: category(string), datetime(integer), headline(string), id(integer), image(string), related(string), source(string), summary(string), symbol(string), url(string)
  company_profile:
    primary key: ticker
    fields: country(string), currency(string), exchange(string), finnhubIndustry(string), ipo(string), logo(string), marketCapitalization(number), name(string), phone(string), shareOutstanding(number), ticker(string), weburl(string)
  stock_recommendations:
    primary key: symbol, period
    fields: buy(integer), hold(integer), period(string), sell(integer), strongBuy(integer), strongSell(integer), symbol(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Finnhub API read of market data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect finnhub

  # Inspect as structured JSON
  pm connectors inspect finnhub --json

AGENT WORKFLOW
  - Run pm connectors inspect finnhub before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
