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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  api_key (secret)

ETL STREAMS
  companies:
    primary key: id
    fields: id(), name(), ticker(), updated_at()
  statements:
    primary key: id
    fields: id(), name(), ticker(), updated_at()
  markets:
    primary key: id
    fields: id(), name(), ticker(), updated_at()
  company_general_compact:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  company_general_verbose:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  company_statements_compact:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  company_statements_verbose:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  company_prices_compact:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  company_prices_verbose:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  common_shares_outstanding:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  weighted_shares_outstanding:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  filings_by_company:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  filings:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  changed_companies:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()
  data_change_log:
    fields: changes(), columns(), companyId(), companyName(), data(), date(), filingIdentifier(), filingType(), fiscalPeriod(), fiscalYear(), id(), name(), prices(), shares(), simId(), statements(), ticker()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
