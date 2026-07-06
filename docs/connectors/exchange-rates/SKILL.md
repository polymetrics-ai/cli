---
name: pm-exchange-rates
description: Exchange Rates API connector knowledge and safe action guide.
---

# pm-exchange-rates

## Purpose

Reads latest, currency-conversion, time-series, and fluctuation foreign-exchange rate data from the exchangeratesapi.io REST API. The legacy exchange_rates daily-historical stream (a date-by-date iteration over a start_date..end_date window) and the symbols stream are not ported; see docs.md Known limits.

## Icon

- asset: icons/exchangeratesapi.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://exchangeratesapi.io/documentation/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base
- base_url
- convert_amount
- convert_date
- convert_from
- convert_to
- fluctuation_end_date
- fluctuation_start_date
- mode
- timeseries_end_date
- timeseries_start_date
- access_key (secret)

## ETL Streams

- latest:
  - primary key: date
  - fields: base(), date(), historical(), rates(), success(), timestamp()
- convert:
  - primary key: date
  - fields: date(), historical(), info(), query(), result(), success()
- timeseries:
  - primary key: start_date, end_date
  - fields: base(), end_date(), rates(), start_date(), success(), timeseries()
- fluctuation:
  - primary key: start_date, end_date
  - fields: base(), end_date(), fluctuation(), rates(), start_date(), success()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external exchangeratesapi.io read of public foreign-exchange rate data
- approval: none; read-only public data API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect exchange-rates
```

### Inspect as structured JSON

```bash
pm connectors inspect exchange-rates --json
```

## Agent Rules

- Run pm connectors inspect exchange-rates before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
