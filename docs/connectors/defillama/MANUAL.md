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
  mode

ETL STREAMS
  protocols:
    primary key: id
    fields: category(), chain(), chains(), change_1d(), change_7d(), id(), mcap(), name(), slug(), symbol(), tvl(), url()
  chains:
    primary key: name
    fields: chainId(), cmcId(), gecko_id(), name(), tokenSymbol(), tvl()
  stablecoins:
    primary key: id
    fields: circulating(), gecko_id(), id(), name(), pegMechanism(), pegType(), price(), symbol()
  dexs:
    primary key: defillamaId
    fields: category(), chains(), change_1d(), defillamaId(), displayName(), name(), total24h(), total30d(), total7d(), totalAllTime()
  fees:
    primary key: defillamaId
    fields: category(), chains(), change_1d(), defillamaId(), displayName(), name(), total24h(), total30d(), total7d(), totalAllTime()
  options:
    primary key: defillamaId
    fields: category(), chains(), change_1d(), defillamaId(), displayName(), name(), total24h(), total30d(), total7d(), totalAllTime()
  open_interest:
    primary key: defillamaId
    fields: category(), chains(), change_1d(), defillamaId(), displayName(), name(), total24h(), total30d(), total7d(), totalAllTime()
  pools:
    primary key: pool
    fields: apy(), apyBase(), apyPct1D(), apyPct30D(), apyPct7D(), apyReward(), chain(), exposure(), ilRisk(), pool(), poolMeta(), project(), rewardTokens(), stablecoin(), symbol(), tvlUsd(), underlyingTokens()
  stablecoin_chains:
    primary key: name
    fields: name(), totalCirculatingUSD()
  historical_chain_tvl:
    primary key: date
    fields: date(), tvl()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external DefiLlama API read of public DeFi analytics data
  approval: none; read-only public analytics API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
