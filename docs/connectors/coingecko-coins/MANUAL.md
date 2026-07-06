# pm connectors inspect coingecko-coins

```text
NAME
  pm connectors inspect coingecko-coins - CoinGecko Coins connector manual

SYNOPSIS
  pm connectors inspect coingecko-coins
  pm connectors inspect coingecko-coins --json
  pm credentials add <name> --connector coingecko-coins [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads a coin's current metadata/market snapshot and exchange tickers from the CoinGecko REST API (GET /coins/{id}, GET /coins/{id}/tickers). Read-only; unauthenticated by default, an optional pro api_key unlocks the pro base URL and higher limits.

ICON
  asset: icons/coingeckocoins.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.coingecko.com/en/api/documentation

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  coin_id
  mode
  api_key (secret)

ETL STREAMS
  coin:
    primary key: id
    fields: categories(), hashing_algorithm(), id(), last_updated(), market_cap_rank(), market_data(), name(), symbol()
  tickers:
    primary key: coin_id, target_coin_id, market_identifier
    fields: base(), bid_ask_spread_percentage(), coin_id(), converted_last(), converted_volume(), is_anomaly(), is_stale(), last(), last_fetch_at(), last_traded_at(), market(), market_identifier(), target(), target_coin_id(), timestamp(), trade_url(), trust_score(), volume()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external CoinGecko public API read of a single coin's metadata/market snapshot
  approval: none; read-only public market-data API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect coingecko-coins

  # Inspect as structured JSON
  pm connectors inspect coingecko-coins --json

AGENT WORKFLOW
  - Run pm connectors inspect coingecko-coins before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
