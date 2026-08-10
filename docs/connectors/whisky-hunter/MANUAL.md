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
  id: whiskyhunter
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
    fields: dt(string), id(integer), winning_bid(number)
  distilleries:
    primary key: id
    fields: country(string), id(integer), name(string)
  auctions_data:
    primary key: auction_slug, dt
    fields: all_auctions_lots_count(integer), auction_lots_count(integer), auction_name(string), auction_slug(string), auction_trading_volume(number), dt(string), winning_bid_mean(number)
  auctions_info:
    primary key: slug
    fields: base_currency(string), buyers_fee(number), listing_fee(number), name(string), reserve_fee(number), sellers_fee(number), slug(string), url(string)
  distilleries_info:
    primary key: slug
    fields: country(string), name(string), slug(string)
  auction_data:
    primary key: auction_slug, dt
    fields: all_auctions_lots_count(integer), auction_lots_count(integer), auction_name(string), auction_slug(string), auction_trading_volume(number), dt(string), winning_bid_mean(number)
  distillery_data:
    primary key: slug, dt
    fields: dt(string), lots_count(integer), name(string), slug(string), trading_volume(number), winning_bid_max(number), winning_bid_mean(number), winning_bid_min(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Whisky Hunter API read of public auction and distillery data
  approval: none; read-only public API, no credentials
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Whisky Hunter's declared streams and reverse-ETL actions.
  Usage: pm whisky-hunter <command> [flags]
  Read streams
  Other Commands
    auction data list - Run the auction data ETL stream [intent=etl availability=implemented stream=auction_data]
    auctions data list - Run the auctions data ETL stream [intent=etl availability=implemented stream=auctions_data]
    auctions info list - Run the auctions info ETL stream [intent=etl availability=implemented stream=auctions_info]
    auctions list - Run the auctions ETL stream [intent=etl availability=implemented stream=auctions]
    distilleries info list - Run the distilleries info ETL stream [intent=etl availability=implemented stream=distilleries_info]
    distilleries list - Run the distilleries ETL stream [intent=etl availability=implemented stream=distilleries]
    distillery data list - Run the distillery data ETL stream [intent=etl availability=implemented stream=distillery_data]

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
