---
name: pm-bunny-inc
description: Bunny, Inc. connector knowledge and safe action guide.
---

# pm-bunny-inc

## Purpose

Reads Bunny subscription-billing data (accounts, contacts, invoices, payments, subscriptions) from the per-tenant Bunny GraphQL API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- subdomain
- apikey (secret)

## ETL Streams

- accounts:
  - primary key: id
  - cursor: updatedAt
  - fields: accountTypeId(), annualRevenue(), billingCountry(), code(), createdAt(), currencyId(), employees(), entityId(), id(), name(), netPaymentDays(), ownerUserId(), payingStatus(), phone(), updatedAt(), website()
- contacts:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(), code(), createdAt(), email(), entityId(), firstName(), fullName(), id(), lastName(), mobile(), phone(), portalAccess(), title(), updatedAt()
- invoices:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(), amount(), amountDue(), amountPaid(), createdAt(), credits(), currencyId(), dueAt(), id(), netPaymentDays(), number(), paidAt(), quoteId(), subtotal(), taxAmount(), updatedAt(), url(), uuid()
- payments:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(), amount(), amountUnapplied(), baseCurrencyCash(), baseCurrencyId(), createdAt(), currencyId(), description(), id(), isLegacy(), memo(), receivedAt(), updatedAt()
- subscriptions:
  - primary key: id
  - cursor: updatedAt
  - fields: accountId(), cancelationDate(), createdAt(), currencyId(), endDate(), id(), name(), period(), priceListId(), rampIntervalMonths(), startDate(), trialEndDate(), trialPeriod(), trialStartDate(), updatedAt()

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
