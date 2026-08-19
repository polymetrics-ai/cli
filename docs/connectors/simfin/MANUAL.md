# pm connectors inspect simfin

```text
NAME
  pm connectors inspect simfin - SimFin connector manual

SYNOPSIS
  pm connectors inspect simfin
  pm connectors inspect simfin --json
  pm credentials add <name> --connector simfin [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SimFin company, financial statement, price, share, filing, and database-change data through the SimFin REST API.

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
  as_reported
  base_url
  company_ids
  end_date
  filing_company_id
  filing_company_ticker
  fiscal_years
  include_details
  include_ratios
  include_ttm
  periods
  start_date
  statements
  tickers
  api_key (secret) (required)

ETL STREAMS
  companies:
    primary key: id
    fields: id(string), name(string), ticker(string), updated_at(string)
  statements:
    primary key: id
    fields: id(string), name(string), ticker(string), updated_at(string)
  markets:
    primary key: id
    fields: id(string), name(string), ticker(string), updated_at(string)
  company_general_compact:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  company_general_verbose:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  company_statements_compact:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  company_statements_verbose:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  company_prices_compact:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  company_prices_verbose:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  common_shares_outstanding:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  weighted_shares_outstanding:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  filings_by_company:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  filings:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  changed_companies:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)
  data_change_log:
    fields: changes(array), columns(array), companyId(string), companyName(string), data(array), date(string), filingIdentifier(string), filingType(string), fiscalPeriod(string), fiscalYear(string), id(string), name(string), prices(array), shares(array), simId(string), statements(array), ticker(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external SimFin API read of company, statement, price, share, filing, and change-log data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect simfin

  # Inspect as structured JSON
  pm connectors inspect simfin --json

AGENT WORKFLOW
  - Run pm connectors inspect simfin before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
