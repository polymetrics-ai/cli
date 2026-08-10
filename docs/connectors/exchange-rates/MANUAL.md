# pm connectors inspect exchange-rates

```text
NAME
  pm connectors inspect exchange-rates - Exchange Rates API connector manual

SYNOPSIS
  pm connectors inspect exchange-rates
  pm connectors inspect exchange-rates --json
  pm credentials add <name> --connector exchange-rates [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads latest, currency-conversion, time-series, and fluctuation foreign-exchange rate data from the exchangeratesapi.io REST API. The legacy exchange_rates daily-historical stream (a date-by-date iteration over a start_date..end_date window) and the symbols stream are not ported; see docs.md Known limits.

ICON
  id: exchangeratesapi
  asset: icons/exchangeratesapi.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://exchangeratesapi.io/documentation/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base
  base_url
  convert_amount
  convert_date
  convert_from
  convert_to
  fluctuation_end_date
  fluctuation_start_date
  mode
  timeseries_end_date
  timeseries_start_date
  access_key (secret)

ETL STREAMS
  latest:
    primary key: date
    fields: base(string), date(string), historical(boolean), rates(object), success(boolean), timestamp(integer)
  convert:
    primary key: date
    fields: date(string), historical(string), info(object), query(object), result(number), success(boolean)
  timeseries:
    primary key: start_date, end_date
    fields: base(string), end_date(string), rates(object), start_date(string), success(boolean), timeseries(boolean)
  fluctuation:
    primary key: start_date, end_date
    fields: base(string), end_date(string), fluctuation(boolean), rates(object), start_date(string), success(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external exchangeratesapi.io read of public foreign-exchange rate data
  approval: none; read-only public data API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Exchange Rates API's declared streams and reverse-ETL actions.
  Usage: pm exchange-rates <command> [flags]
  Read streams
  Other Commands
    api get date - Documented GET /{date} (not implemented) [intent=direct_read availability=not_implemented operation=exchange-rates.get.date]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get symbols - Documented GET /symbols (not implemented) [intent=direct_read availability=not_implemented operation=exchange-rates.get.symbols]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    convert list - Run the convert ETL stream [intent=etl availability=implemented stream=convert]
    fluctuation list - Run the fluctuation ETL stream [intent=etl availability=implemented stream=fluctuation]
    latest list - Run the latest ETL stream [intent=etl availability=implemented stream=latest]
    timeseries list - Run the timeseries ETL stream [intent=etl availability=implemented stream=timeseries]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect exchange-rates

  # Inspect as structured JSON
  pm connectors inspect exchange-rates --json

AGENT WORKFLOW
  - Run pm connectors inspect exchange-rates before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
