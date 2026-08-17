# pm connectors inspect whisky-hunter

```text
NAME
  pm connectors inspect whisky-hunter - Whisky Hunter connector manual

SYNOPSIS
  pm connectors inspect whisky-hunter
  pm connectors inspect whisky-hunter --json
  pm credentials add <name> --connector whisky-hunter [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public Whisky Hunter auction and distillery data. Read-only, no credentials required.

ICON
  asset: icons/whiskyhunter.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://whiskyhunter.net/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url

ETL STREAMS
  auctions:
    primary key: id
    fields: dt(), id(), winning_bid()
  distilleries:
    primary key: id
    fields: country(), id(), name()
  auctions_data:
    primary key: auction_slug, dt
    fields: all_auctions_lots_count(), auction_lots_count(), auction_name(), auction_slug(), auction_trading_volume(), dt(), winning_bid_mean()
  auctions_info:
    primary key: slug
    fields: base_currency(), buyers_fee(), listing_fee(), name(), reserve_fee(), sellers_fee(), slug(), url()
  distilleries_info:
    primary key: slug
    fields: country(), name(), slug()
  auction_data:
    primary key: auction_slug, dt
    fields: all_auctions_lots_count(), auction_lots_count(), auction_name(), auction_slug(), auction_trading_volume(), dt(), winning_bid_mean()
  distillery_data:
    primary key: slug, dt
    fields: dt(), lots_count(), name(), slug(), trading_volume(), winning_bid_max(), winning_bid_mean(), winning_bid_min()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Whisky Hunter API read of public auction and distillery data
  approval: none; read-only public API, no credentials
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect whisky-hunter

  # Inspect as structured JSON
  pm connectors inspect whisky-hunter --json

AGENT WORKFLOW
  - Run pm connectors inspect whisky-hunter before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
