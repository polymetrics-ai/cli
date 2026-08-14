# pm connectors inspect alpaca-broker-api

```text
NAME
  pm connectors inspect alpaca-broker-api - Alpaca Broker API connector manual

SYNOPSIS
  pm connectors inspect alpaca-broker-api
  pm connectors inspect alpaca-broker-api --json
  pm credentials add <name> --connector alpaca-broker-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Alpaca Broker API accounts, assets, market calendar, clock, account activities, journals, and per-account positions/watchlists/orders/documents over the Broker REST API (read-only).

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
  limit
  username (required)
  password (secret) (required)

ETL STREAMS
  accounts:
    primary key: id
    fields: account_number(string), account_type(string), created_at(string), crypto_status(string), currency(string), enabled_assets(string), id(string), kyc_results(string), last_equity(string), status(string)
  assets:
    primary key: id
    fields: class(string), easy_to_borrow(boolean), exchange(string), fractionable(boolean), id(string), marginable(boolean), name(string), shortable(boolean), status(string), symbol(string), tradable(boolean)
  calendar:
    primary key: date
    fields: close(string), date(string), open(string), session_close(string), session_open(string)
  clock:
    primary key: timestamp
    fields: is_open(boolean), next_close(string), next_open(string), timestamp(string)
  account_activities:
    primary key: id
    fields: account_id(string), activity_sub_type(string), activity_type(string), cum_qty(string), cusip(string), date(string), description(string), id(string), leaves_qty(string), net_amount(string), order_id(string), per_share_amount(string), price(string), qty(string), side(string), status(string), symbol(string), transaction_time(string), type(string)
  journals:
    primary key: id
    fields: created_at(string), description(string), entry_type(string), from_account(string), id(string), net_amount(string), price(string), qty(string), settle_date(string), status(string), symbol(string), system_date(string), to_account(string)
  positions:
    primary key: id, account_id
    fields: account_id(string), asset_class(string), asset_id(string), avg_entry_price(string), change_today(string), cost_basis(string), current_price(string), exchange(string), id(string), lastday_price(string), market_value(string), qty(string), qty_available(string), side(string), symbol(string), unrealized_pl(string), unrealized_plpc(string)
  watchlists:
    primary key: id
    fields: account_id(string), created_at(string), id(string), name(string), updated_at(string)
  orders:
    primary key: id
    fields: account_id(string), asset_class(string), canceled_at(string), created_at(string), filled_at(string), filled_avg_price(string), filled_qty(string), id(string), limit_price(string), notional(string), order_class(string), order_type(string), qty(string), side(string), status(string), stop_price(string), submitted_at(string), symbol(string), time_in_force(string), type(string), updated_at(string)
  documents:
    primary key: id, account_id
    fields: account_id(string), date(string), id(string), name(string), sub_type(string), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Alpaca Broker API read of account/asset/market metadata, plus per-account trading positions, orders, watchlists, and document metadata (financial PII adjacent; no document content is downloaded, only listing metadata)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Alpaca Broker API's declared streams and reverse-ETL actions.
  Usage: pm alpaca-broker-api <command> [flags]
  Read streams
  Other Commands
    account activities list - Run the account activities ETL stream [intent=etl availability=implemented stream=account_activities]
    accounts list - Run the accounts ETL stream [intent=etl availability=implemented stream=accounts]
    assets list - Run the assets ETL stream [intent=etl availability=implemented stream=assets]
    calendar list - Run the calendar ETL stream [intent=etl availability=implemented stream=calendar]
    clock list - Run the clock ETL stream [intent=etl availability=implemented stream=clock]
    documents list - Run the documents ETL stream [intent=etl availability=implemented stream=documents]
    journals list - Run the journals ETL stream [intent=etl availability=implemented stream=journals]
    orders list - Run the orders ETL stream [intent=etl availability=implemented stream=orders]
    positions list - Run the positions ETL stream [intent=etl availability=implemented stream=positions]
    watchlists list - Run the watchlists ETL stream [intent=etl availability=implemented stream=watchlists]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect alpaca-broker-api

  # Inspect as structured JSON
  pm connectors inspect alpaca-broker-api --json

AGENT WORKFLOW
  - Run pm connectors inspect alpaca-broker-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
