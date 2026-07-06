# pm connectors inspect bunny-inc

```text
NAME
  pm connectors inspect bunny-inc - Bunny, Inc. connector manual

SYNOPSIS
  pm connectors inspect bunny-inc
  pm connectors inspect bunny-inc --json
  pm credentials add <name> --connector bunny-inc [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Bunny subscription-billing data (accounts, contacts, invoices, payments, subscriptions) from the per-tenant Bunny GraphQL API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  base_url
  mode
  start_date
  subdomain
  apikey (secret)

ETL STREAMS
  accounts:
    primary key: id
    cursor: updatedAt
    fields: accountTypeId(), annualRevenue(), billingCountry(), code(), createdAt(), currencyId(), employees(), entityId(), id(), name(), netPaymentDays(), ownerUserId(), payingStatus(), phone(), updatedAt(), website()
  contacts:
    primary key: id
    cursor: updatedAt
    fields: accountId(), code(), createdAt(), email(), entityId(), firstName(), fullName(), id(), lastName(), mobile(), phone(), portalAccess(), title(), updatedAt()
  invoices:
    primary key: id
    cursor: updatedAt
    fields: accountId(), amount(), amountDue(), amountPaid(), createdAt(), credits(), currencyId(), dueAt(), id(), netPaymentDays(), number(), paidAt(), quoteId(), subtotal(), taxAmount(), updatedAt(), url(), uuid()
  payments:
    primary key: id
    cursor: updatedAt
    fields: accountId(), amount(), amountUnapplied(), baseCurrencyCash(), baseCurrencyId(), createdAt(), currencyId(), description(), id(), isLegacy(), memo(), receivedAt(), updatedAt()
  subscriptions:
    primary key: id
    cursor: updatedAt
    fields: accountId(), cancelationDate(), createdAt(), currencyId(), endDate(), id(), name(), period(), priceListId(), rampIntervalMonths(), startDate(), trialEndDate(), trialPeriod(), trialStartDate(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Bunny, Inc. API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bunny-inc

  # Inspect as structured JSON
  pm connectors inspect bunny-inc --json

AGENT WORKFLOW
  - Run pm connectors inspect bunny-inc before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
