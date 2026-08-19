# pm connectors inspect defillama

```text
NAME
  pm connectors inspect defillama - DefiLlama connector manual

SYNOPSIS
  pm connectors inspect defillama
  pm connectors inspect defillama --json
  pm credentials add <name> --connector defillama [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads DefiLlama DeFi analytics: protocols, chains, stablecoins, DEX volumes, and fees/revenue from the public DefiLlama REST API. Read-only; no authentication required.

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
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  mode

ETL STREAMS
  protocols:
    primary key: id
    fields: category(string), chain(string), chains(array), change_1d(number), change_7d(number), id(string), mcap(number), name(string), slug(string), symbol(string), tvl(number), url(string)
  chains:
    primary key: name
    fields: chainId(number), cmcId(string), gecko_id(string), name(string), tokenSymbol(string), tvl(number)
  stablecoins:
    primary key: id
    fields: circulating(object), gecko_id(string), id(string), name(string), pegMechanism(string), pegType(string), price(number), symbol(string)
  dexs:
    primary key: defillamaId
    fields: category(string), chains(array), change_1d(number), defillamaId(string), displayName(string), name(string), total24h(number), total30d(number), total7d(number), totalAllTime(number)
  fees:
    primary key: defillamaId
    fields: category(string), chains(array), change_1d(number), defillamaId(string), displayName(string), name(string), total24h(number), total30d(number), total7d(number), totalAllTime(number)
  options:
    primary key: defillamaId
    fields: category(string), chains(array), change_1d(number), defillamaId(string), displayName(string), name(string), total24h(number), total30d(number), total7d(number), totalAllTime(number)
  open_interest:
    primary key: defillamaId
    fields: category(string), chains(array), change_1d(number), defillamaId(string), displayName(string), name(string), total24h(number), total30d(number), total7d(number), totalAllTime(number)
  pools:
    primary key: pool
    fields: apy(number), apyBase(number), apyPct1D(number), apyPct30D(number), apyPct7D(number), apyReward(number), chain(string), exposure(string), ilRisk(string), pool(string), poolMeta(string), project(string), rewardTokens(array), stablecoin(boolean), symbol(string), tvlUsd(number), underlyingTokens(array)
  stablecoin_chains:
    primary key: name
    fields: name(string), totalCirculatingUSD(object)
  historical_chain_tvl:
    primary key: date
    fields: date(integer), tvl(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external DefiLlama API read of public DeFi analytics data
  approval: none; read-only public analytics API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run DefiLlama's declared streams and reverse-ETL actions.
  Usage: pm defillama <command> [flags]
  Read streams
  Other Commands
    chains list - Run the chains ETL stream [intent=etl availability=implemented stream=chains]
    dexs list - Run the dexs ETL stream [intent=etl availability=implemented stream=dexs]
    fees list - Run the fees ETL stream [intent=etl availability=implemented stream=fees]
    historical chain tvl list - Run the historical chain tvl ETL stream [intent=etl availability=implemented stream=historical_chain_tvl]
    open interest list - Run the open interest ETL stream [intent=etl availability=implemented stream=open_interest]
    options list - Run the options ETL stream [intent=etl availability=implemented stream=options]
    pools list - Run the pools ETL stream [intent=etl availability=implemented stream=pools]
    protocols list - Run the protocols ETL stream [intent=etl availability=implemented stream=protocols]
    stablecoin chains list - Run the stablecoin chains ETL stream [intent=etl availability=implemented stream=stablecoin_chains]
    stablecoins list - Run the stablecoins ETL stream [intent=etl availability=implemented stream=stablecoins]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect defillama

  # Inspect as structured JSON
  pm connectors inspect defillama --json

AGENT WORKFLOW
  - Run pm connectors inspect defillama before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
