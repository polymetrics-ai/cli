---
name: pm-bunny-inc
description: Bunny, Inc. connector knowledge and safe action guide.
---

# pm-bunny-inc

## Purpose

Reads Bunny subscription-billing data (accounts, contacts, invoices, payments, subscriptions) from the per-tenant Bunny GraphQL API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- subdomain (required)
- apikey (secret) (required)

## ETL Streams

- accounts:
  - primary key: id
  - cursor: updatedAt
  - fields: accountTypeId(string), annualRevenue(number), billingCountry(string), code(string), createdAt(string), currencyId(string), employees(integer), entityId(string), id(string), name(string), netPaymentDays(integer), ownerUserId(string), payingStatus(string), phone(string), updatedAt(string), website(string)
- contacts:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(string), code(string), createdAt(string), email(string), entityId(string), firstName(string), fullName(string), id(string), lastName(string), mobile(string), phone(string), portalAccess(boolean), title(string), updatedAt(string)
- invoices:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(string), amount(number), amountDue(number), amountPaid(number), createdAt(string), credits(number), currencyId(string), dueAt(string), id(string), netPaymentDays(integer), number(string), paidAt(string), quoteId(string), subtotal(number), taxAmount(number), updatedAt(string), url(string), uuid(string)
- payments:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(string), amount(number), amountUnapplied(number), baseCurrencyCash(number), baseCurrencyId(string), createdAt(string), currencyId(string), description(string), id(string), isLegacy(boolean), memo(string), receivedAt(string), updatedAt(string)
- subscriptions:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(string), cancelationDate(string), createdAt(string), currencyId(string), endDate(string), id(string), name(string), period(string), priceListId(string), rampIntervalMonths(integer), startDate(string), trialEndDate(string), trialPeriod(integer), trialStartDate(string), updatedAt(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Bunny, Inc. API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect bunny-inc
```

### Inspect as structured JSON

```bash
pm connectors inspect bunny-inc --json
```

## Agent Rules

- Run pm connectors inspect bunny-inc before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
